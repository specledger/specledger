package commands

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"specledger/pkg/cli/config"
	"specledger/pkg/cli/logger"
	"specledger/pkg/cli/metadata"
	"specledger/pkg/cli/prerequisites"
	"specledger/pkg/cli/tui"
	"specledger/pkg/cli/ui"
	"specledger/pkg/embedded"
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
- specledger/ directory for specifications
- specledger/specledger.yaml file for project metadata`,

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
- specledger/ directory for specifications
- specledger/specledger.yaml file for project metadata`,
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

	// Apply embedded playbooks (use TUI selection or default)
	selectedPlaybookName, playbookVersion, playbookStructure, err := applyEmbeddedPlaybooks(projectPath, playbookName)
	if err != nil {
		// Playbook application failure is not fatal - log and continue
		fmt.Printf("Warning: playbook application had issues: %v\n", err)
	}

	// Create YAML metadata with playbook info
	projectMetadata := metadata.NewProjectMetadata(projectName, shortCode, selectedPlaybookName, playbookVersion, playbookStructure)
	if err := metadata.SaveToProject(projectMetadata, projectPath); err != nil {
		return fmt.Errorf("failed to create project metadata: %w", err)
	}

	// Initialize git repo (but don't commit - user might bootstrap into existing repo)
	if err := initializeGitRepo(projectPath); err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
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

	// Apply embedded playbooks (use default for CI mode)
	selectedPlaybookName, playbookVersion, playbookStructure, err := applyEmbeddedPlaybooks(projectPath, "")
	if err != nil {
		// Playbook application failure is not fatal - log and continue
		fmt.Printf("Warning: playbook application had issues: %v\n", err)
	}

	// Create YAML metadata with playbook info
	projectMetadata := metadata.NewProjectMetadata(projectName, shortCode, selectedPlaybookName, playbookVersion, playbookStructure)
	if err := metadata.SaveToProject(projectMetadata, projectPath); err != nil {
		return fmt.Errorf("failed to create project metadata: %w", err)
	}

	// Initialize git repo (but don't commit - user might bootstrap into existing repo)
	if err := initializeGitRepo(projectPath); err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
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
			return fmt.Errorf("already initialized (specledger/specledger.yaml exists). Use --force to re-initialize")
		}
		if _, err := os.Stat(filepath.Join(projectPath, ".beads")); err == nil {
			return fmt.Errorf("already initialized (.beads exists). Use --force to re-initialize")
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

	// Apply embedded playbooks (use flag if provided, otherwise default)
	selectedPlaybookName, playbookVersion, playbookStructure, err := applyEmbeddedPlaybooks(projectPath, initPlaybookFlag)
	if err != nil {
		// Playbook application failure is not fatal - log and continue
		fmt.Printf("Warning: playbook application had issues: %v\n", err)
	}

	// Create YAML metadata with playbook info
	projectMetadata := metadata.NewProjectMetadata(projectName, shortCode, selectedPlaybookName, playbookVersion, playbookStructure)
	if err := metadata.SaveToProject(projectMetadata, projectPath); err != nil {
		return fmt.Errorf("failed to create project metadata: %w", err)
	}

	// Success message
	ui.PrintHeader("SpecLedger Initialized", "", 60)
	fmt.Printf("  Directory:  %s\n", ui.Bold(projectPath))
	fmt.Printf("  Beads:      %s\n", ui.Bold(shortCode))
	fmt.Printf("  Metadata:   %s\n", ui.Bold("specledger/specledger.yaml"))
	fmt.Println()
	ui.PrintSuccess("SpecLedger is ready to use!")
	fmt.Println(ui.Bold("Next steps:"))
	fmt.Printf("  %s             %s\n", ui.Cyan("sl deps list"), ui.Dim("# List dependencies"))
	fmt.Printf("  %s <repo-url> %s\n", ui.Cyan("sl deps add"), ui.Dim("# Add a dependency"))
	fmt.Printf("  %s             %s\n", ui.Cyan("bd create"), ui.Dim("Create an issue"))
	fmt.Printf("  %s             %s\n", ui.Cyan("bd ready"), ui.Dim("Find work to do"))
	fmt.Println()

	return nil
}

// copyInitTemplates copies SpecLedger templates for init (excludes new-project specific files)
func copyInitTemplates(projectPath, shortCode, projectName string) error {
	// Files and directories to exclude from init
	excludePaths := map[string]bool{
		"specledger/FORK.md":          true, // FORK.md is for new projects
		"specledger/memory":           true,
		"specledger/scripts":          true,
		"spec-kit-version":            true,
		"specledger/spec-kit-version": true,
		"specledger/templates":        true,
		"AGENTS.md":                   true, // Don't overwrite existing AGENTS.md if present
	}

	// Walk through the embedded filesystem
	err := fs.WalkDir(embedded.TemplatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path (remove "templates/" prefix)
		relPath := strings.TrimPrefix(path, "templates/")
		if relPath == "" || relPath == "." {
			return nil
		}

		// Check if this path should be excluded
		if excludePaths[relPath] {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(projectPath, relPath)

		// For files that already exist, skip unless it's .beads/config
		if _, err := os.Stat(destPath); err == nil {
			if filepath.Base(path) != "config.yaml" || filepath.Dir(path) != "templates/.beads" {
				// File exists, skip it
				return nil
			}
		}

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read file from embedded FS
		data, err := embedded.TemplatesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// For .beads/config.yaml, replace the prefix
		if filepath.Base(path) == "config.yaml" && filepath.Dir(path) == "templates/.beads" {
			data = []byte(strings.ReplaceAll(string(data), "issue-prefix: \"sl\"", fmt.Sprintf("issue-prefix: \"%s\"", shortCode)))
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}

		// If we just copied mise.toml, run mise trust
		if filepath.Base(path) == "mise.toml" {
			cmd := exec.Command("mise", "trust")
			cmd.Dir = projectPath
			_ = cmd.Run() // Ignore errors if mise is not installed
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk embedded templates: %w", err)
	}

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

// copyTemplates copies SpecLedger template files to the new project using embedded templates
func copyTemplates(projectPath, shortCode, projectName string) error {
	// Files and directories to exclude from copying
	excludePaths := map[string]bool{
		"specledger/FORK.md":          true,
		"specledger/memory":           true,
		"specledger/scripts":          true,
		"spec-kit-version":            true,
		"specledger/spec-kit-version": true,
		"specledger/templates":        true,
		// Don't exclude specledger directory itself - we want it!
	}

	// Walk through the embedded filesystem
	err := fs.WalkDir(embedded.TemplatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path (remove "templates/" prefix)
		relPath := strings.TrimPrefix(path, "templates/")
		if relPath == "" || relPath == "." {
			return nil
		}

		// Check if this path should be excluded
		if excludePaths[relPath] {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(projectPath, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read file from embedded FS
		data, err := embedded.TemplatesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// For .beads/config.yaml, replace the prefix
		if filepath.Base(path) == "config.yaml" && filepath.Dir(path) == "templates/.beads" {
			data = []byte(strings.ReplaceAll(string(data), "issue-prefix: \"sl\"", fmt.Sprintf("issue-prefix: \"%s\"", shortCode)))
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}

		// If we just copied mise.toml, run mise trust
		if filepath.Base(path) == "mise.toml" {
			cmd := exec.Command("mise", "trust")
			cmd.Dir = projectPath
			_ = cmd.Run() // Ignore errors if mise is not installed
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk embedded templates: %w", err)
	}

	// Note: specledger.yaml is now created separately via metadata.SaveToProject()
	// No longer creating .mod file here
	return nil
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
