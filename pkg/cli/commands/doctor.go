package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/specledger/specledger/pkg/cli/prerequisites"
	"github.com/specledger/specledger/pkg/cli/tui"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/specledger/specledger/pkg/templates"
	"github.com/specledger/specledger/pkg/version"
	"github.com/spf13/cobra"
)

var (
	doctorJSONOutput   bool
	doctorUpdateFlag   bool
	doctorTemplateFlag bool
)

// VarDoctorCmd represents the doctor command
var VarDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check installation status of required and optional tools",
	Long: `Check the installation status of all tools required by SpecLedger.

This command verifies that:
- Core tools (mise) are installed and accessible
- CLI version is up to date (prompts to update if not)
- Project templates match the CLI version (prompts to apply if not)

Use --update to update the CLI binary without prompting.
Use --template to apply embedded templates without prompting.
Use --json flag for machine-readable output suitable for CI/CD pipelines.`,
	Example: `  sl doctor              # Full check, prompt to update CLI and templates if needed
  sl doctor --update     # Update CLI to latest version (non-interactive)
  sl doctor --template   # Apply embedded templates (non-interactive)
  sl doctor --json       # JSON output for CI/CD`,
	RunE:          runDoctor,
	SilenceUsage:  true, // Don't print usage on error
	SilenceErrors: true, // Don't print error message from return (we handle it in UI)
}

func init() {
	VarDoctorCmd.Flags().BoolVar(&doctorJSONOutput, "json", false, "Output results in JSON format")
	VarDoctorCmd.Flags().BoolVar(&doctorUpdateFlag, "update", false, "Update CLI to latest version (non-interactive)")
	VarDoctorCmd.Flags().BoolVar(&doctorTemplateFlag, "template", false, "Apply embedded templates to project (non-interactive)")
}

// DoctorOutput represents the JSON output structure for doctor command
type DoctorOutput struct {
	Status              string             `json:"status"`
	Tools               []DoctorToolStatus `json:"tools"`
	Missing             []string           `json:"missing,omitempty"`
	InstallInstructions string             `json:"install_instructions,omitempty"`

	// CLI version info
	CLIVersion            string `json:"cli_version"`
	CLILatestVersion      string `json:"cli_latest_version,omitempty"`
	CLIUpdateAvailable    bool   `json:"cli_update_available"`
	CLIUpdateInstructions string `json:"cli_update_instructions,omitempty"`
	CLICheckError         string `json:"cli_check_error,omitempty"`

	// Template version info
	TemplateVersion         string   `json:"template_version,omitempty"`
	TemplateUpdateAvailable bool     `json:"template_update_available"`
	TemplateCustomizedFiles []string `json:"template_customized_files,omitempty"`
}

// DoctorToolStatus represents a tool's status in JSON output
type DoctorToolStatus struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
	Path      string `json:"path,omitempty"`
	Category  string `json:"category"`
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Flag-only mode: skip full doctor output, just do the requested action(s)
	if doctorUpdateFlag || doctorTemplateFlag {
		var errs []error
		if doctorUpdateFlag {
			if err := performCLIUpdate(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				errs = append(errs, err)
			}
		}
		if doctorTemplateFlag {
			if err := performTemplateUpdate(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	}

	// Full doctor mode
	check := prerequisites.CheckPrerequisites()

	if doctorJSONOutput {
		return outputDoctorJSON(check)
	}

	return outputDoctorHuman(check)
}

// performCLIUpdate checks for a newer CLI version and updates without prompting.
func performCLIUpdate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("  Checking for CLI updates...\n")
	versionInfo := version.CheckLatestVersion(ctx)
	if versionInfo.Error != "" {
		return fmt.Errorf("version check failed: %s", versionInfo.Error)
	}

	if !versionInfo.UpdateAvailable {
		fmt.Printf("  %s CLI is up to date (%s)\n", ui.Checkmark(), version.GetVersion())
		return nil
	}

	fmt.Printf("  Updating CLI %s -> %s...\n", version.GetVersion(), versionInfo.LatestVersion)
	if err := version.SelfUpdate(ctx); err != nil {
		return fmt.Errorf("CLI update failed: %w\n  Try manual update: %s", err, version.GetUpdateInstructions())
	}
	fmt.Printf("  %s CLI updated. Restart sl to use the new version.\n", ui.Checkmark())
	return nil
}

// performTemplateUpdate applies embedded templates to the project without prompting.
func performTemplateUpdate() error {
	projectDir, _ := os.Getwd()
	cliVersion := version.GetVersion()
	templateStatus, err := templates.CheckTemplateStatus(projectDir, cliVersion)
	if err != nil || templateStatus == nil || !templateStatus.InProject {
		return fmt.Errorf("not in a SpecLedger project (no specledger.yaml found)")
	}

	// --template flag forces update regardless of version
	fmt.Printf("  Applying templates (v%s -> v%s)...\n", templateStatus.ProjectTemplateVersion, cliVersion)
	result, err := templates.UpdateTemplates(projectDir, cliVersion)
	if err != nil {
		return fmt.Errorf("template update failed: %w", err)
	}
	total := len(result.Updated) + len(result.Overwritten)
	fmt.Printf("  %s Updated %d templates (%d new, %d overwritten)\n",
		ui.Checkmark(), total, len(result.Updated), len(result.Overwritten))

	// Warn about stale files if any detected
	if len(result.Stale) > 0 {
		fmt.Printf("  %s Found %d stale templates (not deleted, manual cleanup recommended):\n",
			ui.Yellow("⚠"), len(result.Stale))
		for _, f := range result.Stale {
			fmt.Printf("    - %s\n", f)
		}
	}
	return nil
}

func outputDoctorJSON(check prerequisites.PrerequisiteCheck) error {
	output := DoctorOutput{
		Status: "pass",
		Tools:  []DoctorToolStatus{},
	}

	// Add all tools to output
	for _, result := range check.CoreResults {
		output.Tools = append(output.Tools, DoctorToolStatus{
			Name:      result.Tool.Name,
			Installed: result.Installed,
			Version:   result.Version,
			Path:      result.Path,
			Category:  string(result.Tool.Category),
		})
	}

	// Set status and missing tools
	if !check.AllCoreInstalled {
		output.Status = "fail"
		output.Missing = []string{}
		for _, tool := range check.MissingCore {
			output.Missing = append(output.Missing, tool.Name)
		}
		output.InstallInstructions = check.Instructions
	}

	// Add CLI version info
	output.CLIVersion = version.GetVersion()

	// Check for updates
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	versionInfo := version.CheckLatestVersion(ctx)
	if versionInfo.Error != "" {
		output.CLICheckError = versionInfo.Error
	} else {
		output.CLILatestVersion = versionInfo.LatestVersion
		output.CLIUpdateAvailable = versionInfo.UpdateAvailable
		if versionInfo.UpdateAvailable {
			output.CLIUpdateInstructions = version.GetUpdateInstructions()
		}
	}

	// Add template version info
	projectDir, _ := os.Getwd()
	templateStatus, _ := templates.CheckTemplateStatus(projectDir, version.GetVersion())
	if templateStatus != nil && templateStatus.InProject {
		output.TemplateVersion = templateStatus.ProjectTemplateVersion
		output.TemplateUpdateAvailable = templateStatus.UpdateAvailable
		output.TemplateCustomizedFiles = templateStatus.CustomizedFiles
	}

	// Marshal and print JSON
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

func outputDoctorHuman(check prerequisites.PrerequisiteCheck) error {
	// Header
	ui.PrintHeader("SpecLedger Doctor", "Environment Check", 58)

	// Core tools section
	fmt.Println(ui.Bold("Core Tools"))
	fmt.Println(ui.Cyan("──────────"))
	fmt.Println()

	for _, result := range check.CoreResults {
		name := result.Tool.DisplayName
		versionInfo := ""
		status := ui.Crossmark() + " "
		if result.Installed {
			status = ui.Checkmark() + " "
			if result.Version != "" {
				versionInfo = ui.Dim(fmt.Sprintf("(%s)", result.Version))
			}
		}
		fmt.Printf("  %s%s%s\n", status, ui.Bold(name), versionInfo)
	}
	fmt.Println()

	// CLI version section
	fmt.Println(ui.Bold("SpecLedger CLI"))
	fmt.Println(ui.Cyan("──────────────"))
	fmt.Println()

	cliVersion := version.GetVersion()
	fmt.Printf("  %s Version: %s", ui.Checkmark(), ui.Bold(cliVersion))

	// Check for updates (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	versionInfo := version.CheckLatestVersion(ctx)
	if versionInfo.Error != "" {
		fmt.Printf(" %s\n", ui.Dim(fmt.Sprintf("(check skipped: %s)", versionInfo.Error)))
	} else if versionInfo.UpdateAvailable {
		fmt.Printf(" %s\n", ui.Yellow(fmt.Sprintf("(latest: %s)", versionInfo.LatestVersion)))
		fmt.Println()
		fmt.Printf("  CLI update available: %s -> %s\n", cliVersion, versionInfo.LatestVersion)
		confirmed, err := tui.ConfirmPrompt("  Update CLI? [y/N]: ")
		if err != nil {
			fmt.Printf("  %s Failed to read confirmation: %v\n", ui.Red("✗"), err)
		} else if confirmed {
			fmt.Printf("  Updating CLI...\n")
			if err := version.SelfUpdate(ctx); err != nil {
				fmt.Printf("  %s Update failed: %v\n", ui.Red("✗"), err)
				fmt.Printf("  %s Try manual update:\n", ui.Dim("ℹ"))
				fmt.Printf("      %s\n", version.GetUpdateInstructions())
			} else {
				fmt.Printf("  %s CLI updated. Restart sl to use the new version.\n", ui.Checkmark())
			}
		} else {
			fmt.Printf("  Skipping CLI update\n")
		}
	} else {
		fmt.Printf(" %s\n", ui.Green("(latest)"))
	}
	fmt.Println()

	// Template status section (only if in a project)
	projectDir, _ := os.Getwd()
	templateStatus, _ := templates.CheckTemplateStatus(projectDir, cliVersion)
	if templateStatus != nil && templateStatus.InProject {
		fmt.Println(ui.Bold("Project Templates"))
		fmt.Println(ui.Cyan("─────────────────"))
		fmt.Println()

		if templateStatus.ProjectTemplateVersion == "" {
			fmt.Printf("  ⚠  Templates: %s\n", ui.Yellow("(version unknown)"))
		} else if templateStatus.UpdateAvailable {
			fmt.Printf("  ⚠  Templates: %s %s\n",
				ui.Dim(fmt.Sprintf("v%s", templateStatus.ProjectTemplateVersion)),
				ui.Yellow(fmt.Sprintf("(CLI: v%s)", cliVersion)))
		} else {
			fmt.Printf("  %s Templates: %s\n", ui.Checkmark(), ui.Green("current"))
		}

		// Ask user before applying template update
		if templateStatus.NeedsUpdate {
			fmt.Println()
			if hasUncommittedChanges(projectDir) {
				fmt.Printf("  %s Warning: Uncommitted changes in .claude/ will be overwritten\n", ui.Yellow("⚠"))
			}
			fmt.Printf("  Template update available: v%s -> v%s\n", templateStatus.ProjectTemplateVersion, cliVersion)
			confirmed, err := tui.ConfirmPrompt("  Apply template updates? [y/N]: ")
			if err != nil {
				fmt.Printf("  %s Failed to read confirmation: %v\n", ui.Red("✗"), err)
			} else if confirmed {
				fmt.Printf("  Applying templates...\n")
				result, err := templates.UpdateTemplates(projectDir, cliVersion)
				if err != nil {
					fmt.Printf("  %s Template update failed: %v\n", ui.Red("✗"), err)
				} else {
					total := len(result.Updated) + len(result.Overwritten)
					fmt.Printf("  %s Updated %d templates (%d new, %d overwritten)\n",
						ui.Checkmark(), total, len(result.Updated), len(result.Overwritten))
				}
			} else {
				fmt.Printf("  Skipping template update\n")
			}
		}
		fmt.Println()
	}

	// Overall status
	if check.AllCoreInstalled {
		ui.PrintBox("All core tools installed", ui.Green, 54)
		return nil
	}

	// Missing tools - print error and return error for exit code
	// SilenceUsage: true prevents Cobra from printing help message
	ui.PrintBox("Missing required tools", ui.Red, 54)
	fmt.Println()
	fmt.Println(check.Instructions)

	return fmt.Errorf("missing required tools")
}

// hasUncommittedChanges checks if there are uncommitted changes in .claude/ directory
func hasUncommittedChanges(projectDir string) bool {
	// Run git status to check for uncommitted changes in .claude/
	// This is a simple check - we just check if there are any changes
	cmd := exec.Command("git", "status", "--porcelain", ".claude/")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return false // If git fails, assume no uncommitted changes
	}
	return len(output) > 0
}
