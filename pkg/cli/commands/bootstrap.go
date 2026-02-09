package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/specledger/specledger/pkg/cli/config"
	"github.com/specledger/specledger/pkg/cli/logger"
	"github.com/specledger/specledger/pkg/cli/prerequisites"
	"github.com/specledger/specledger/pkg/cli/tui"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/spf13/cobra"
)

var (
	projectNameFlag string
	shortCodeFlag   string
	demoDirFlag     string
	ciFlag          bool
	// Init-specific flags
	initShortCodeFlag string
	initPlaybookFlag  string
	initForceFlag     bool
)

// VarBootstrapCmd is the bootstrap command
var VarBootstrapCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new SpecLedger project",
	Long: `Create a new SpecLedger project with all necessary infrastructure:

Interactive mode:
  sl new

Non-interactive mode (for CI/CD):
  sl new --ci --project-name <name> --short-code <code> --project-dir <path>

The bootstrap creates:
- .claude/ directory with skills and commands
- .beads/ directory for issue tracking
- github.com/specledger/specledger/ directory for specifications
- github.com/specledger/specledger/specledger.yaml file for project metadata`,

	// RunE is called when the command is executed
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create logger
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		l := logger.New(logger.Debug)

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
- .beads/ directory for issue tracking
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
		shouldOverwrite, err := tui.ConfirmPrompt(fmt.Sprintf("Directory '%s' already exists. Overwrite? [y/N]: ", projectName))
		if err != nil {
			return fmt.Errorf("failed to confirm overwrite: %w", err)
		}
		if !shouldOverwrite {
			return fmt.Errorf("bootstrap cancelled by user")
		}
	}

	// Create directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Setup SpecLedger project (playbooks, skills, metadata, git)
	_, _, _, err = setupSpecLedgerProject(projectPath, projectName, shortCode, playbookName, true)
	if err != nil {
		return err
	}

	// Success message
	ui.PrintHeader("Project Created Successfully", "", 60)
	fmt.Printf("  Path:       %s\n", ui.Bold(projectPath))
	fmt.Printf("  Beads:      %s\n", ui.Bold(shortCode))
	fmt.Println()
	fmt.Println(ui.Bold("Next steps:"))
	fmt.Printf("  %s    %s\n", ui.Cyan("cd"), projectPath)
	fmt.Printf("  %s  %s\n", ui.Cyan("sl doctor"), ui.Dim("# Check tool installation status"))
	fmt.Println()

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
		return ErrProjectExists(projectName)
	}

	// Create directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return ErrPermissionDenied(projectPath)
	}

	// Setup SpecLedger project (playbooks, skills, metadata, git)
	_, _, _, err := setupSpecLedgerProject(projectPath, projectName, shortCode, "", true)
	if err != nil {
		return err
	}

	// Success message
	ui.PrintHeader("Project Created Successfully", "", 60)
	fmt.Printf("  Path:       %s\n", ui.Bold(projectPath))
	fmt.Printf("  Beads:      %s\n", ui.Bold(shortCode))
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
			return fmt.Errorf("already initialized (github.com/specledger/specledger/specledger.yaml exists). Use --force to re-initialize")
		}
	}

	// Determine short code
	shortCode := initShortCodeFlag
	if shortCode == "" {
		// Try to derive from project name
		if len(projectName) >= 2 {
			shortCode = strings.ToLower(projectName[:2])
		} else {
			shortCode = "sl"
		}
	}

	// Limit short code to 4 characters
	if len(shortCode) > 4 {
		shortCode = shortCode[:4]
	}

	ui.PrintSection("Initializing SpecLedger")
	fmt.Printf("  Directory:  %s\n", ui.Bold(projectPath))
	fmt.Printf("  Project:    %s\n", ui.Bold(projectName))
	fmt.Printf("  Short Code: %s\n", ui.Bold(shortCode))
	if initPlaybookFlag != "" {
		fmt.Printf("  Playbook:   %s\n", ui.Bold(initPlaybookFlag))
	}
	fmt.Println()

	// Setup SpecLedger project (playbooks, skills, metadata, no git)
	// Note: initGit=false because we're in an existing repo
	_, _, _, err = setupSpecLedgerProject(projectPath, projectName, shortCode, initPlaybookFlag, false)
	if err != nil {
		return err
	}

	// Success message
	ui.PrintHeader("SpecLedger Initialized", "", 60)
	fmt.Printf("  Directory:  %s\n", ui.Bold(projectPath))
	fmt.Printf("  Beads:      %s\n", ui.Bold(shortCode))
	fmt.Printf("  Metadata:   %s\n", ui.Bold("github.com/specledger/specledger/specledger.yaml"))
	fmt.Println()
	ui.PrintSuccess("SpecLedger is ready to use!")
	fmt.Println(ui.Bold("Next steps:"))
	fmt.Printf("  %s             %s\n", ui.Cyan("sl deps list"), ui.Dim("# List dependencies"))
	fmt.Printf("  %s <repo-url> %s\n", ui.Cyan("sl deps add"), ui.Dim("# Add a dependency"))
	fmt.Println()

	return nil
}

func init() {
	// Flags for 'new' command
	VarBootstrapCmd.PersistentFlags().StringVarP(&projectNameFlag, "project-name", "n", "", "Project name")
	VarBootstrapCmd.PersistentFlags().StringVarP(&shortCodeFlag, "short-code", "s", "", "Short code (2-4 letters)")
	VarBootstrapCmd.PersistentFlags().StringVarP(&demoDirFlag, "project-dir", "d", "", "Project directory path")
	VarBootstrapCmd.PersistentFlags().BoolVarP(&ciFlag, "ci", "", false, "Force non-interactive mode (skip TUI)")

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
