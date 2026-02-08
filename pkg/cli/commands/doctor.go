package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"specledger/pkg/cli/metadata"
	"specledger/pkg/cli/prerequisites"
	"specledger/pkg/cli/ui"
)

var (
	doctorJSONOutput bool
)

// VarDoctorCmd represents the doctor command
var VarDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check installation status of required and optional tools",
	Long: `Check the installation status of all tools required by SpecLedger.

This command verifies that:
- Core tools (mise, bd, perles) are installed and accessible
- Framework tools (specify, openspec) are installed (optional)

Use --json flag for machine-readable output suitable for CI/CD pipelines.`,
	Example: `  sl doctor           # Human-readable output
  sl doctor --json    # JSON output for CI/CD`,
	RunE:         runDoctor,
	SilenceUsage: true, // Don't print usage on error
	SilenceErrors: true, // Don't print error message from return (we handle it in UI)
}

func init() {
	VarDoctorCmd.Flags().BoolVar(&doctorJSONOutput, "json", false, "Output results in JSON format")
}

// DoctorOutput represents the JSON output structure for doctor command
type DoctorOutput struct {
	Status              string             `json:"status"`
	Tools               []DoctorToolStatus `json:"tools"`
	Missing             []string           `json:"missing,omitempty"`
	InstallInstructions string             `json:"install_instructions,omitempty"`
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
	check := prerequisites.CheckPrerequisites()

	if doctorJSONOutput {
		return outputDoctorJSON(check)
	}

	return outputDoctorHuman(check)
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

	for _, result := range check.FrameworkResults {
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

	// Framework tools section
	fmt.Println(ui.Bold("SDD Framework Tools"))
	fmt.Println(ui.Cyan("──────────────────"))
	fmt.Println()

	for _, result := range check.FrameworkResults {
		name := result.Tool.DisplayName
		versionInfo := ""
		status := ui.Crossmark() + " "
		fwStatus := ""

		if result.Installed {
			status = ui.Checkmark() + " "
			if result.Version != "" {
				versionInfo = ui.Dim(fmt.Sprintf("(%s)", result.Version))
			}

			// Check if playbook is applied in current project
			projectDir, _ := os.Getwd()
			if metadata.HasYAMLMetadata(projectDir) {
				if meta, _ := metadata.LoadFromProject(projectDir); meta != nil {
					// Show playbook name instead of framework choice
					if meta.Playbook.Name != "" {
						fwStatus = fmt.Sprintf("(playbook: %s)", meta.Playbook.Name)
					} else {
						fwStatus = ui.Yellow("(no playbook)")
					}
				}
			}
		}
		fmt.Printf("  %s%s%s %s\n", status, ui.Bold(name), versionInfo, fwStatus)
	}
	fmt.Println()

	// Check if we're in a SpecLedger project and show framework init commands
	projectDir, err := os.Getwd()
	if err == nil && metadata.HasYAMLMetadata(projectDir) {
		meta, loadErr := metadata.LoadFromProject(projectDir)
		if loadErr == nil {
			showFrameworkInitCommands(check, meta)
		}
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

// showFrameworkInitCommands shows commands to initialize frameworks that need it
func showFrameworkInitCommands(check prerequisites.PrerequisiteCheck, meta *metadata.ProjectMetadata) {
	// Framework initialization commands are no longer needed
	// as we use playbooks instead of frameworks
	return
}

// isFrameworkInitialized checks if a framework is already initialized in the project
func isFrameworkInitialized(framework string) bool {
	dir, err := os.Getwd()
	if err != nil {
		return false
	}

	switch framework {
	case "specify":
		// Check for .specify directory or specify.yaml
		specifyDir := filepath.Join(dir, ".specify")
		specifyYaml := filepath.Join(dir, "specify.yaml")
		_, err1 := os.Stat(specifyDir)
		_, err2 := os.Stat(specifyYaml)
		return err1 == nil || err2 == nil
	case "openspec":
		// Check for .openspec directory or openspec.yaml
		openspecDir := filepath.Join(dir, ".openspec")
		openspecYaml := filepath.Join(dir, "openspec.yaml")
		_, err1 := os.Stat(openspecDir)
		_, err2 := os.Stat(openspecYaml)
		return err1 == nil || err2 == nil
	}
	return false
}
