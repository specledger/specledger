package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"specledger/pkg/cli/framework"
	"specledger/pkg/cli/metadata"
	"specledger/pkg/cli/ui"
)

// VarDepsCmd represents the deps command
var VarDepsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Manage specification dependencies",
	Long: `Manage external specification dependencies for your project.

Dependencies are stored in specledger/specledger.yaml and cached locally for offline use.

Examples:
  sl deps list                           # List all dependencies
  sl deps add git@github.com:org/spec    # Add a dependency
  sl deps remove git@github.com:org/spec # Remove a dependency`,
}

// VarAddCmd represents the add command
var VarAddCmd = &cobra.Command{
	Use:     "add <repo-url> [branch] [spec-path]",
	Short:   "Add a dependency",
	Long:    `Add an external specification dependency to your project. The dependency will be tracked in specledger.yaml and cached locally for offline use.`,
	Example: `  sl deps add git@github.com:org/api-spec
  sl deps add git@github.com:org/api-spec v1.0 specs/api.md
  sl deps add git@github.com:org/api-spec main spec.md --alias api`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAddDependency,
}

// VarDepsListCmd represents the list command
var VarDepsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all dependencies",
	Long:    `List all declared dependencies from specledger.yaml, showing their repository, version, and resolved status.`,
	Example: `  sl deps list`,
	RunE:    runListDependencies,
}

// VarRemoveCmd represents the remove command
var VarRemoveCmd = &cobra.Command{
	Use:     "remove <repo-url>",
	Short:   "Remove a dependency",
	Long:    `Remove a dependency from specledger.yaml. The local cache will be kept for future use.`,
	Example: `  sl deps remove git@github.com:org/api-spec`,
	Args:    cobra.ExactArgs(1),
	RunE:    runRemoveDependency,
}

// VarResolveCmd represents the resolve command
var VarResolveCmd = &cobra.Command{
	Use:     "resolve",
	Short:   "Download and cache dependencies",
	Long:    `Download all dependencies from specledger.yaml and cache them locally at ~/.specledger/cache/.`,
	Example: `  sl deps resolve`,
	RunE:    runResolveDependencies,
}

// VarDepsUpdateCmd represents the update command
var VarDepsUpdateCmd = &cobra.Command{
	Use:     "update [repo-url]",
	Short:   "Update dependencies to latest versions",
	Long:    `Update dependencies to their latest versions. If no URL is given, updates all dependencies.`,
	Example: `  sl deps update                    # Update all
  sl deps update git@github.com:org/spec # Update one`,
	RunE:    runUpdateDependencies,
}

func init() {
	VarDepsCmd.AddCommand(VarAddCmd, VarDepsListCmd, VarResolveCmd, VarDepsUpdateCmd, VarRemoveCmd)

	VarAddCmd.Flags().StringP("alias", "a", "", "Optional alias for the dependency")
	VarResolveCmd.Flags().BoolP("no-cache", "n", false, "Ignore cached specifications")
}

func runAddDependency(cmd *cobra.Command, args []string) error {
	// Get current directory or find project root
	projectDir, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Load existing metadata
	meta, err := metadata.LoadFromProject(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Extract flags
	alias, _ := cmd.Flags().GetString("alias")

	// Parse arguments
	repoURL := args[0]
	branch := "main" // default
	specPath := "spec.md"

	if len(args) >= 2 {
		branch = args[1]
	}
	if len(args) >= 3 {
		specPath = args[2]
	}

	// Validate URL
	if !isValidGitURL(repoURL) {
		return fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	// Detect framework type
	frameworkType := metadata.FrameworkNone
	ui.PrintSection("Detecting Framework")
	fmt.Printf("Checking %s...\n", ui.Bold(repoURL))

	detectedFramework, err := framework.DetectFramework(repoURL)
	if err != nil {
		ui.PrintWarning(fmt.Sprintf("Could not detect framework: %v", err))
		ui.PrintWarning("Continuing with 'none' as framework type")
	} else {
		frameworkType = detectedFramework
	}

	// Display detected framework
	frameworkDisplay := "None"
	switch frameworkType {
	case metadata.FrameworkSpecKit:
		frameworkDisplay = ui.Cyan("Spec Kit")
	case metadata.FrameworkOpenSpec:
		frameworkDisplay = ui.Cyan("OpenSpec")
	case metadata.FrameworkBoth:
		frameworkDisplay = ui.Cyan("Both")
	}
	fmt.Printf("  Framework:  %s\n", frameworkDisplay)
	fmt.Println()

	// Create dependency
	dep := metadata.Dependency{
		URL:       repoURL,
		Branch:    branch,
		Path:      specPath,
		Alias:     alias,
		Framework: frameworkType,
	}

	// Generate import path for AI context
	importPath := framework.GetFrameworkImportPath(dep)
	dep.ImportPath = importPath

	// Check for duplicates
	for _, existing := range meta.Dependencies {
		if existing.URL == repoURL {
			return fmt.Errorf("dependency already exists: %s", repoURL)
		}
		if alias != "" && existing.Alias == alias {
			return fmt.Errorf("alias already exists: %s", alias)
		}
	}

	// Add dependency
	meta.Dependencies = append(meta.Dependencies, dep)

	// Save metadata
	if err := metadata.SaveToProject(meta, projectDir); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	ui.PrintSuccess("Dependency added")
	fmt.Printf("  Repository:  %s\n", ui.Bold(repoURL))
	if alias != "" {
		fmt.Printf("  Alias:       %s\n", ui.Bold(alias))
	}
	fmt.Printf("  Branch:      %s\n", ui.Bold(branch))
	fmt.Printf("  Path:        %s\n", ui.Bold(specPath))
	fmt.Printf("  Framework:   %s\n", frameworkDisplay)
	fmt.Printf("  Import Path: %s\n", ui.Cyan(importPath))
	fmt.Println()
	fmt.Printf("Next: %s\n", ui.Cyan("sl deps resolve"))
	fmt.Println()

	return nil
}

func runListDependencies(cmd *cobra.Command, args []string) error {
	projectDir, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	meta, err := metadata.LoadFromProject(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	if len(meta.Dependencies) == 0 {
		ui.PrintSection("Dependencies")
		fmt.Println("No dependencies declared.")
		fmt.Println()
		fmt.Println(ui.Bold("Add dependencies with:"))
		fmt.Printf("  %s\n", ui.Cyan("sl deps add git@github.com:org/spec"))
		fmt.Println()
		return nil
	}

	ui.PrintHeader("Dependencies", fmt.Sprintf("%d total", len(meta.Dependencies)), 70)
	fmt.Println()

	for i, dep := range meta.Dependencies {
		fmt.Printf("%s. %s\n", ui.Bold(fmt.Sprintf("%d", i+1)), ui.Bold(dep.URL))
		if dep.Branch != "" && dep.Branch != "main" {
			fmt.Printf("   Branch:  %s\n", ui.Cyan(dep.Branch))
		}
		if dep.Path != "" && dep.Path != "spec.md" {
			fmt.Printf("   Path:    %s\n", ui.Cyan(dep.Path))
		}
		if dep.Alias != "" {
			fmt.Printf("   Alias:   %s\n", ui.Cyan(dep.Alias))
		}
		if dep.Framework != "" && dep.Framework != metadata.FrameworkNone {
			frameworkDisplay := string(dep.Framework)
			fmt.Printf("   Framework: %s\n", ui.Cyan(frameworkDisplay))
		}
		if dep.ImportPath != "" {
			fmt.Printf("   Import:    %s\n", ui.Yellow(dep.ImportPath))
		}
		if dep.ResolvedCommit != "" {
			fmt.Printf("   Status:  %s %s\n", ui.Green("✓"), ui.Gray(dep.ResolvedCommit[:8]))
		} else {
			fmt.Printf("   Status:  %s (run %s)\n", ui.Yellow("not resolved"), ui.Cyan("sl deps resolve"))
		}
		fmt.Println()
	}

	return nil
}

func runRemoveDependency(cmd *cobra.Command, args []string) error {
	projectDir, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	meta, err := metadata.LoadFromProject(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	target := args[0]

	// Find and remove dependency
	removed := false
	removedIndex := -1

	for i, dep := range meta.Dependencies {
		if dep.URL == target || dep.Alias == target {
			removedIndex = i
			removed = true
			break
		}
	}

	if !removed {
		return fmt.Errorf("dependency not found: %s", target)
	}

	// Remove from slice
	meta.Dependencies = append(meta.Dependencies[:removedIndex], meta.Dependencies[removedIndex+1:]...)

	// Save metadata
	if err := metadata.SaveToProject(meta, projectDir); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	ui.PrintSuccess("Dependency removed")
	fmt.Printf("  %s\n", ui.Bold(target))
	fmt.Println()

	return nil
}

func runResolveDependencies(cmd *cobra.Command, args []string) error {
	projectDir, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	meta, err := metadata.LoadFromProject(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	if len(meta.Dependencies) == 0 {
		ui.PrintWarning("No dependencies to resolve")
		return nil
	}

	ui.PrintSection("Resolving Dependencies")
	// TODO: Implement actual dependency resolution
	// For now, just show what would be resolved
	fmt.Printf("Would resolve %s dependencies:\n", ui.Bold(fmt.Sprintf("%d", len(meta.Dependencies))))
	fmt.Println()
	for _, dep := range meta.Dependencies {
		fmt.Printf("  - %s", ui.Bold(dep.URL))
		if dep.Alias != "" {
			fmt.Printf(" (alias: %s)", ui.Cyan(dep.Alias))
		}
		fmt.Println()
	}
	fmt.Println()
	ui.PrintWarning("Dependency resolution not yet implemented")
	fmt.Println("Dependencies are tracked in specledger/specledger.yaml")
	fmt.Println()

	return nil
}

func runUpdateDependencies(cmd *cobra.Command, args []string) error {
	projectDir, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	meta, err := metadata.LoadFromProject(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	if len(meta.Dependencies) == 0 {
		return fmt.Errorf("no dependencies to update")
	}

	ui.PrintSection("Checking for Updates")
	fmt.Printf("Checking %s dependencies for updates...\n", ui.Bold(fmt.Sprintf("%d", len(meta.Dependencies))))
	fmt.Println()

	for _, dep := range meta.Dependencies {
		// TODO: Implement actual update checking
		if dep.ResolvedCommit != "" {
			fmt.Printf("  %s: at %s\n", ui.Bold(dep.URL), ui.Gray(dep.ResolvedCommit[:8]))
		} else {
			fmt.Printf("  %s: %s\n", ui.Bold(dep.URL), ui.Yellow("not resolved yet"))
		}
	}
	fmt.Println()
	ui.PrintWarning("Dependency updates not yet implemented")
	fmt.Println()

	return nil
}

func isValidGitURL(s string) bool {
	// Simple check for common Git URLs and local paths
	return len(s) > 0 && (strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "git@") ||
		strings.HasPrefix(s, "/") || // Local absolute path
		strings.HasPrefix(s, "./") || // Local relative path
		strings.HasPrefix(s, "../"))
}

func findProjectRoot() (string, error) {
	// Start from current directory and work up
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Check current directory
	if metadata.HasYAMLMetadata(dir) {
		return dir, nil
	}

	// Check parent directories
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			return "", fmt.Errorf("not in a SpecLedger project (no specledger/specledger.yaml found)")
		}
		dir = parent

		if metadata.HasYAMLMetadata(dir) {
			return dir, nil
		}
	}
}
