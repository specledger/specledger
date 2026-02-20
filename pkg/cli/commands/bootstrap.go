package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/specledger/specledger/pkg/cli/config"
	"github.com/specledger/specledger/pkg/cli/logger"
	"github.com/specledger/specledger/pkg/cli/metadata"
	"github.com/specledger/specledger/pkg/cli/playbooks"
	"github.com/specledger/specledger/pkg/cli/prerequisites"
	"github.com/specledger/specledger/pkg/cli/tui"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/specledger/specledger/pkg/models"
	"github.com/spf13/cobra"
)

var (
	projectNameFlag   string
	shortCodeFlag     string
	demoDirFlag       string
	ciFlag            bool
	templateFlag      string
	agentFlag         string
	listTemplatesFlag bool
	forceFlag         bool
	// Init-specific flags
	initShortCodeFlag string
	initPlaybookFlag  string
	initForceFlag     bool
)

// VarBootstrapCmd is the bootstrap command
var VarBootstrapCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new SpecLedger project",
	Long: `Create a new SpecLedger project with all necessary infrastructure.

Interactive mode (default):
  sl new

Non-interactive mode (for CI/CD):
  sl new --ci -n my-project -s mp -d /path --template full-stack --agent claude-code

List available templates:
  sl new --list-templates

Examples:
  sl new                                    # Interactive TUI mode
  sl new --template ml-image                # Pre-select ML template
  sl new --agent opencode                   # Pre-select OpenCode agent
  sl new --force                            # Overwrite existing directory
  sl new --ci -n app -s ap -d . -t general-purpose -a none

The bootstrap creates:
- Project template structure (based on --template selection)
- Agent config directory (.claude/ or .opencode/ based on --agent)
- specledger/ directory for specifications
- specledger/specledger.yaml with project metadata and UUID`,

	// RunE is called when the command is executed
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle --list-templates flag first
		if listTemplatesFlag {
			return listTemplates()
		}

		// Create logger
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		l := logger.New(logger.Debug)

		// Validate template flag if provided
		if templateFlag != "" {
			if _, err := playbooks.GetTemplateByID(templateFlag); err != nil {
				templates, _ := playbooks.LoadTemplates()
				var validIDs []string
				for _, t := range templates {
					validIDs = append(validIDs, t.ID)
				}
				return fmt.Errorf("unknown template: %q\nAvailable: %s\nUse --list-templates for details", templateFlag, strings.Join(validIDs, ", "))
			}
		}

		// Validate agent flag if provided
		if agentFlag != "" {
			if _, err := models.GetAgentByID(agentFlag); err != nil {
				return err
			}
		}

		// Check if non-interactive mode
		modeDetector := tui.NewModeDetector()
		if modeDetector.IsNonInteractive() || ciFlag {
			return runBootstrapNonInteractive(cmd, l, cfg)
		}

		// Interactive TUI mode
		return runBootstrapInteractive(l, cfg)
	},
}

// VarInitCmd is the init command
var VarInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize SpecLedger in an existing repository",
	Long: `Initialize SpecLedger in the current repository directory.

This adds SpecLedger to an existing project without creating a new directory.

Usage:
  sl init
  sl init --short-code abc
  sl init --playbook specledger

The init creates:
- .claude/ directory with skills
- Built-in issue tracking via sl issue commands
- github.com/specledger/specledger/ directory for specifications
- github.com/specledger/specledger/specledger.yaml file for project metadata`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(logger.Debug)
		return runInit(l)
	},
}

// runBootstrapInteractive runs the bootstrap with Bubble Tea TUI
func runBootstrapInteractive(l *logger.Logger, cfg *config.Config) error {
	// Determine default project directory
	defaultDir := cfg.DefaultProjectDir
	if demoDirFlag != "" {
		defaultDir = demoDirFlag
	}

	// Run Bubble Tea TUI with default directory
	tuiProgram := tui.NewProgram(defaultDir)
	answers, err := tuiProgram.Run()
	if err != nil {
		return fmt.Errorf("TUI exited with error: %w", err)
	}

	// Check if user cancelled (Ctrl+C)
	if len(answers) == 0 || answers["project_name"] == "" {
		return fmt.Errorf("bootstrap cancelled by user")
	}

	projectName := answers["project_name"]
	projectDir := answers["project_dir"]
	shortCode := answers["short_code"]
	playbookName := answers["playbook"]

	// Check prerequisites before starting
	fmt.Println()
	ui.PrintSection("Checking Prerequisites")
	if err := prerequisites.EnsurePrerequisites(true); err != nil {
		// Continue anyway - prerequisites are helpful but not blocking
		fmt.Printf("%s %v\n", ui.WarningIcon(), err)
		fmt.Println("Continuing with bootstrap...")
	} else {
		fmt.Printf("%s All prerequisites installed\n", ui.Checkmark())
	}

	// Create project path
	projectPath := filepath.Join(projectDir, projectName)

	// Check if directory already exists
	if _, err := os.Stat(projectPath); err == nil {
		shouldOverwrite := forceFlag
		if !forceFlag {
			var promptErr error
			shouldOverwrite, promptErr = tui.ConfirmPrompt(fmt.Sprintf("Directory '%s' already exists. Overwrite? [y/N]: ", projectName))
			if promptErr != nil {
				return fmt.Errorf("failed to confirm overwrite: %w", promptErr)
			}
		}
		if !shouldOverwrite {
			return fmt.Errorf("bootstrap cancelled by user")
		}
		// Remove existing directory
		if err := os.RemoveAll(projectPath); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
		fmt.Printf("%s Removed existing directory: %s\n", ui.WarningIcon(), projectPath)
	}

	// Create directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Setup SpecLedger project (playbooks, skills, metadata, git)
	_, _, _, err = setupSpecLedgerProject(projectPath, projectName, shortCode, playbookName, true, false)
	if err != nil {
		return err
	}

	// Write populated constitution with selected principles
	constitutionPath := filepath.Join(projectPath, ".specledger", "memory", "constitution.md")
	agentPref := answers["agent_preference"]
	if agentPref == "" {
		agentPref = "None"
	}

	// Parse selected principles from TUI
	selectedPrinciples := DefaultPrinciples()
	if principleNames, ok := answers["constitution_principles"]; ok && principleNames != "" {
		selected := make(map[string]bool)
		for _, name := range strings.Split(principleNames, ",") {
			selected[name] = true
		}
		for i := range selectedPrinciples {
			selectedPrinciples[i].Selected = selected[selectedPrinciples[i].Name]
		}
	}

	if err := WriteDefaultConstitution(constitutionPath, selectedPrinciples, agentPref); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to write constitution: %v", err))
	} else {
		fmt.Printf("%s Constitution created\n", ui.Checkmark())
	}

	// Setup agent config directory based on selection
	agentID := answers["agent"]
	if agentID == "" {
		// Fall back to agent_preference for backward compatibility
		agentID = mapAgentPreferenceToID(agentPref)
	}
	if err := setupAgentConfig(projectPath, agentID); err != nil {
		ui.PrintWarning(fmt.Sprintf("Agent config setup issue: %v", err))
	}

	// Update metadata with template and agent selections
	templateID := answers["template"]
	if templateID != "" || agentID != "" {
		projectMetadata, err := metadata.LoadFromProject(projectPath)
		if err == nil {
			if templateID != "" {
				projectMetadata.Project.Template = templateID
			}
			if agentID != "" && agentID != "none" {
				projectMetadata.Project.Agent = agentID
			}
			if err := metadata.SaveToProject(projectMetadata, projectPath); err != nil {
				ui.PrintWarning(fmt.Sprintf("Failed to update metadata: %v", err))
			}
		}
	}

	// Success message
	ui.PrintHeader("Project Created Successfully", "", 60)
	fmt.Printf("  Path:        %s\n", ui.Bold(projectPath))
	fmt.Printf("  Short Code:  %s\n", ui.Bold(shortCode))
	fmt.Println()

	// Launch agent if selected and in interactive mode
	if shouldLaunchAgent() && !ciFlag {
		if err := launchAgent(projectPath, agentPref); err != nil {
			ui.PrintWarning(fmt.Sprintf("Agent launch issue: %v", err))
		}
	} else if agentPref != "None" && agentPref != "" {
		fmt.Println(ui.Bold("Next steps:"))
		fmt.Printf("  %s    %s\n", ui.Cyan("cd"), projectPath)
		fmt.Printf("  %s %s\n", ui.Cyan(agentPref), ui.Dim("# Launch your coding agent"))
		fmt.Println()
	}

	return nil
}

// runBootstrapNonInteractive runs bootstrap without TUI
func runBootstrapNonInteractive(cmd *cobra.Command, l *logger.Logger, cfg *config.Config) error {
	// Validate required flags
	if projectNameFlag == "" {
		return fmt.Errorf("--project-name flag is required in non-interactive mode")
	}

	if shortCodeFlag == "" {
		return fmt.Errorf("--short-code flag is required in non-interactive mode")
	}

	// Check prerequisites in CI mode (non-interactive)
	if err := prerequisites.EnsurePrerequisites(false); err != nil {
		// In CI mode, prerequisites are required
		return fmt.Errorf("prerequisites check failed: %w", err)
	}

	projectName := projectNameFlag
	shortCode := strings.ToLower(shortCodeFlag)

	// Limit short code to 4 characters
	if len(shortCode) > 4 {
		shortCode = shortCode[:4]
	}

	// Get demo directory
	demoDir := cfg.DefaultProjectDir
	if demoDirFlag != "" {
		demoDir = demoDirFlag
	}

	projectPath := filepath.Join(demoDir, projectName)

	// Check if directory already exists
	if _, err := os.Stat(projectPath); err == nil {
		if forceFlag {
			// Remove existing directory when --force is used
			if err := os.RemoveAll(projectPath); err != nil {
				return fmt.Errorf("failed to remove existing directory: %w", err)
			}
			fmt.Printf("%s Removed existing directory: %s\n", ui.WarningIcon(), projectPath)
		} else {
			return ErrProjectExists(projectName)
		}
	}

	// Create directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return ErrPermissionDenied(projectPath)
	}

	// Setup SpecLedger project (playbooks, skills, metadata, git)
	_, _, _, err := setupSpecLedgerProject(projectPath, projectName, shortCode, "", true, false)
	if err != nil {
		return err
	}

	// Use agent from flag or default to "none" in CI mode
	agentID := agentFlag
	if agentID == "" {
		agentID = "none"
	}
	agentPref := "None"
	if agent, err := models.GetAgentByID(agentID); err == nil {
		agentPref = agent.Name
	}

	// Write default constitution in CI mode (all principles selected)
	constitutionPath := filepath.Join(projectPath, ".specledger", "memory", "constitution.md")
	if err := WriteDefaultConstitution(constitutionPath, DefaultPrinciples(), agentPref); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to write constitution: %v", err))
	}

	// Setup agent config directory
	if err := setupAgentConfig(projectPath, agentID); err != nil {
		ui.PrintWarning(fmt.Sprintf("Agent config setup issue: %v", err))
	}

	// Update metadata with template and agent selections
	templateID := templateFlag
	if templateID == "" {
		templateID = "general-purpose" // Default template
	}
	if templateID != "" || agentID != "" {
		projectMetadata, err := metadata.LoadFromProject(projectPath)
		if err == nil {
			projectMetadata.Project.Template = templateID
			if agentID != "" && agentID != "none" {
				projectMetadata.Project.Agent = agentID
			}
			if err := metadata.SaveToProject(projectMetadata, projectPath); err != nil {
				ui.PrintWarning(fmt.Sprintf("Failed to update metadata: %v", err))
			}
		}
	}

	// Success message
	ui.PrintHeader("Project Created Successfully", "", 60)
	fmt.Printf("  Path:       %s\n", ui.Bold(projectPath))
	fmt.Printf("  Short Code: %s\n", ui.Bold(shortCode))
	if templateID != "" {
		fmt.Printf("  Template:   %s\n", ui.Bold(templateID))
	}
	if agentID != "" && agentID != "none" {
		fmt.Printf("  Agent:      %s\n", ui.Bold(agentPref))
	}
	fmt.Println()
	fmt.Println(ui.Bold("Next steps:"))
	fmt.Printf("  %s    %s\n", ui.Cyan("cd"), projectPath)
	fmt.Printf("  %s  %s\n", ui.Cyan("sl doctor"), ui.Dim("# Check tool status"))
	fmt.Println()

	return nil
}

// runInit initializes SpecLedger in the current directory
func runInit(l *logger.Logger) error {
	// Get current directory
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get project name from directory name
	projectName := filepath.Base(projectPath)

	// Check if already initialized
	if !initForceFlag {
		if _, err := os.Stat(filepath.Join(projectPath, "specledger", "specledger.yaml")); err == nil {
			fmt.Println("SpecLedger is already initialized in this directory.")
			fmt.Println(ui.Dim("  Use --force to re-initialize."))
			return nil
		}
	}

	// Check for existing populated constitution
	constitutionPath := filepath.Join(projectPath, ".specledger", "memory", "constitution.md")
	hasConstitution := IsConstitutionPopulated(constitutionPath)

	// Read existing agent preference if constitution exists
	var existingAgentPref string
	if hasConstitution {
		existingAgentPref, _ = ReadAgentPreference(constitutionPath)
	}

	// Determine what config is missing
	shortCode := initShortCodeFlag
	playbookName := initPlaybookFlag
	agentPref := existingAgentPref

	// Check if interactive mode
	modeDetector := tui.NewModeDetector()
	isInteractive := modeDetector.IsInteractive() && !ciFlag

	if isInteractive {
		// Build missing config
		missingConfig := tui.MissingConfig{
			NeedsShortCode:       shortCode == "",
			NeedsPlaybook:        playbookName == "",
			NeedsAgentPreference: existingAgentPref == "",
			ExistingAgentPref:    existingAgentPref,
		}

		// Only show TUI if there's something to ask
		if missingConfig.NeedsShortCode || missingConfig.NeedsPlaybook || missingConfig.NeedsAgentPreference {
			initProgram := tui.NewInitProgram(missingConfig, projectName)
			answers, err := initProgram.Run()
			if err != nil {
				return fmt.Errorf("TUI exited with error: %w", err)
			}

			// Apply TUI answers
			if v, ok := answers["short_code"]; ok && v != "" {
				shortCode = v
			}
			if v, ok := answers["playbook"]; ok && v != "" {
				playbookName = v
			}
			if v, ok := answers["agent_preference"]; ok && v != "" {
				agentPref = v
			}
		}
	}

	// Default short code if still empty
	if shortCode == "" {
		if len(projectName) >= 2 {
			shortCode = strings.ToLower(projectName[:2])
		} else {
			shortCode = "sl"
		}
	}
	if len(shortCode) > 4 {
		shortCode = shortCode[:4]
	}

	ui.PrintSection("Initializing SpecLedger")
	fmt.Printf("  Directory:  %s\n", ui.Bold(projectPath))
	fmt.Printf("  Project:    %s\n", ui.Bold(projectName))
	fmt.Printf("  Short Code: %s\n", ui.Bold(shortCode))
	if playbookName != "" {
		fmt.Printf("  Playbook:   %s\n", ui.Bold(playbookName))
	}
	if hasConstitution {
		fmt.Printf("  Constitution: %s\n", ui.Bold("exists (preserved)"))
	} else {
		fmt.Printf("  Constitution: %s\n", ui.Dim("not found (will be created by AI agent)"))
	}
	fmt.Println()

	// Setup SpecLedger project (playbooks, skills, metadata, no git)
	_, _, _, err = setupSpecLedgerProject(projectPath, projectName, shortCode, playbookName, false, initForceFlag)
	if err != nil {
		return err
	}

	// Detect and update artifact_path if existing directories found
	projectMetadata, err := metadata.LoadFromProject(projectPath)
	if err == nil {
		detectedPath := detectArtifactPath(projectPath)
		if detectedPath != "specledger/" && detectedPath != projectMetadata.ArtifactPath {
			projectMetadata.ArtifactPath = detectedPath
			if err := metadata.SaveToProject(projectMetadata, projectPath); err == nil {
				fmt.Printf("  Artifact Path: %s (detected)\n", ui.Bold(detectedPath))
			}
		}
	}

	// Success message
	ui.PrintHeader("SpecLedger Initialized", "", 60)
	fmt.Printf("  Directory:  %s\n", ui.Bold(projectPath))
	fmt.Printf("  Short Code: %s\n", ui.Bold(shortCode))
	fmt.Printf("  Metadata:   %s\n", ui.Bold("github.com/specledger/specledger/specledger.yaml"))
	fmt.Println()
	ui.PrintSuccess("SpecLedger is ready to use!")
	fmt.Println(ui.Bold("Next steps:"))
	fmt.Printf("  %s             %s\n", ui.Cyan("sl deps list"), ui.Dim("# List dependencies"))
	fmt.Printf("  %s <repo-url> %s\n", ui.Cyan("sl deps add"), ui.Dim("# Add a dependency"))
	fmt.Println()

	// Launch agent if selected and in interactive mode
	if isInteractive && shouldLaunchAgent() {
		if err := launchAgent(projectPath, agentPref); err != nil {
			ui.PrintWarning(fmt.Sprintf("Agent launch issue: %v", err))
		}
	} else {
		ui.PrintSuccess("SpecLedger is ready to use!")
	}

	return nil
}

// listTemplates displays all available project templates
func listTemplates() error {
	templates, err := playbooks.LoadTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	ui.PrintHeader("Available Project Templates", "", 60)
	fmt.Println()

	for _, tmpl := range templates {
		defaultMarker := ""
		if tmpl.IsDefault {
			defaultMarker = ui.Dim(" (default)")
		}

		fmt.Printf("  %s%s\n", ui.Bold(tmpl.ID), defaultMarker)
		fmt.Printf("    %s\n", tmpl.Description)
		if len(tmpl.Characteristics) > 0 {
			fmt.Printf("    %s %s\n", ui.Dim("Tech:"), strings.Join(tmpl.Characteristics, ", "))
		}
		fmt.Println()
	}

	fmt.Println(ui.Dim("Usage: sl new --template <id>"))
	return nil
}

func init() {
	// Flags for 'new' command
	VarBootstrapCmd.PersistentFlags().StringVarP(&projectNameFlag, "project-name", "n", "", "Project name")
	VarBootstrapCmd.PersistentFlags().StringVarP(&shortCodeFlag, "short-code", "s", "", "Short code (2-4 letters)")
	VarBootstrapCmd.PersistentFlags().StringVarP(&demoDirFlag, "project-dir", "d", "", "Project directory path")
	VarBootstrapCmd.PersistentFlags().BoolVarP(&ciFlag, "ci", "", false, "Force non-interactive mode (skip TUI)")
	VarBootstrapCmd.PersistentFlags().StringVarP(&templateFlag, "template", "t", "", "Project template ID (e.g., full-stack, ml-image)")
	VarBootstrapCmd.PersistentFlags().StringVarP(&agentFlag, "agent", "a", "", "Coding agent ID (claude-code, opencode, none)")
	VarBootstrapCmd.PersistentFlags().BoolVar(&listTemplatesFlag, "list-templates", false, "List available project templates")
	VarBootstrapCmd.PersistentFlags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing project directory")

	// Flags for 'init' command
	VarInitCmd.PersistentFlags().StringVarP(&initShortCodeFlag, "short-code", "s", "", "Short code for issue IDs (2-4 letters)")
	VarInitCmd.PersistentFlags().StringVarP(&initPlaybookFlag, "playbook", "p", "", "Playbook to apply (default: specledger)")
	VarInitCmd.PersistentFlags().BoolVarP(&initForceFlag, "force", "", false, "Force initialize even if SpecLedger files exist")
}

// initializeGitRepo initializes a git repository in the project directory
// Note: Only runs git init and git add, does NOT commit to support bootstrapping into existing repos
func initializeGitRepo(projectPath string) error {
	// Run git init
	cmd := exec.Command("git", "init")
	cmd.Dir = projectPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %w\nOutput: %s", err, string(output))
	}

	// Run git add . to stage new files (ignore errors for existing repos)
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = projectPath
	_, _ = cmd.CombinedOutput() // Ignore errors - user might have custom .gitignore

	return nil
}
