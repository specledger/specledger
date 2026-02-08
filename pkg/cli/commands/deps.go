package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"specledger/pkg/cli/framework"
	"specledger/pkg/cli/metadata"
	"specledger/pkg/cli/ui"

	"github.com/spf13/cobra"
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
	Use:   "add <repo-url> [branch] [spec-path]",
	Short: "Add a dependency",
	Long:  `Add an external specification dependency to your project. The dependency will be tracked in specledger.yaml and cached locally for offline use.`,
	Example: `  sl deps add git@github.com:org/api-spec
  sl deps add git@github.com:org/api-spec v1.0 specledger/api.md
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
	Use:   "update [repo-url]",
	Short: "Update dependencies to latest versions",
	Long:  `Update dependencies to their latest versions. If no URL is given, updates all dependencies.`,
	Example: `  sl deps update                    # Update all
  sl deps update git@github.com:org/spec # Update one`,
	RunE: runUpdateDependencies,
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
	fmt.Printf("Resolving %s dependencies...\n", ui.Bold(fmt.Sprintf("%d", len(meta.Dependencies))))
	fmt.Println()

	// Check for --no-cache flag
	noCache, _ := cmd.Flags().GetBool("no-cache")

	// Resolve each dependency
	resolvedCount := 0
	for i, dep := range meta.Dependencies {
		// Use alias as directory name if available, otherwise generate from URL
		dirName := dep.Alias
		if dirName == "" {
			dirName = generateDirName(dep.URL)
		}

		// Determine cache location: project-local specledger/deps/ or global cache
		var cacheDir string
		if noCache {
			// Use project-local cache
			cacheDir = filepath.Join(projectDir, "specledger", "deps", dirName)
		} else {
			// Use global cache in user home directory
			homeDir, _ := os.UserHomeDir()
			cacheDir = filepath.Join(homeDir, ".specledger", "cache", dirName)
		}

		fmt.Printf("%s. %s\n", ui.Bold(fmt.Sprintf("%d", i+1)), ui.Bold(dep.URL))
		if dep.Alias != "" {
			fmt.Printf("   Alias:  %s\n", ui.Cyan(dep.Alias))
		}
		fmt.Printf("   Branch: %s\n", ui.Cyan(dep.Branch))

		// Check if already resolved (skip if --no-cache not set and commit exists)
		if dep.ResolvedCommit != "" && !noCache {
			// Verify the commit still exists in the cloned repo
			if _, err := os.Stat(cacheDir); err == nil {
				// Repo exists, verify commit
				cmd := exec.Command("git", "-C", cacheDir, "rev-parse", dep.ResolvedCommit+"^{commit}")
				if output, err := cmd.CombinedOutput(); err == nil {
					// Commit still valid
					resolvedCount++
					fmt.Printf("   Status: %s %s\n", ui.Green("✓"), ui.Gray(strings.TrimSpace(string(output))[:8]))
					fmt.Println()
					continue
				}
			}
		}

		// Clone or update the repository
		fmt.Printf("   Cache:  %s\n", ui.Cyan(cacheDir))
		fmt.Printf("   Status: %s...\n", ui.Yellow("cloning"))

		if err := cloneOrUpdateRepository(dep, cacheDir); err != nil {
			ui.PrintWarning(fmt.Sprintf("Failed to clone %s: %v", dep.URL, err))
			fmt.Println()
			continue
		}

		// Resolve current commit SHA
		cmd := exec.Command("git", "-C", cacheDir, "rev-parse", "HEAD")
		output, err := cmd.CombinedOutput()
		if err != nil {
			ui.PrintWarning(fmt.Sprintf("Failed to resolve commit: %v", err))
			fmt.Println()
			continue
		}

		commitSHA := strings.TrimSpace(string(output))
		meta.Dependencies[i].ResolvedCommit = commitSHA
		resolvedCount++

		fmt.Printf("   Status: %s %s\n", ui.Green("✓"), ui.Gray(commitSHA[:8]))
		fmt.Println()
	}

	// Save updated metadata
	if err := metadata.SaveToProject(meta, projectDir); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("Resolved %d/%d dependencies", resolvedCount, len(meta.Dependencies)))
	fmt.Println()
	if resolvedCount < len(meta.Dependencies) {
		ui.PrintWarning("Some dependencies failed to resolve")
	}
	fmt.Println()

	return nil
}

// cloneOrUpdateRepository clones a Git repository if it doesn't exist, or updates it if it does
func cloneOrUpdateRepository(dep metadata.Dependency, targetDir string) error {
	// Check if directory already exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		// Clone the repository
		args := []string{"clone", dep.URL, targetDir}
		if dep.Branch != "" && dep.Branch != "main" {
			args = append(args, "--branch", dep.Branch)
		}

		cmd := exec.Command("git", args...)
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}
	} else {
		// Repository exists, fetch and pull updates
		// Fetch latest changes
		cmd := exec.Command("git", "-C", targetDir, "fetch", "origin")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git fetch failed: %w", err)
		}

		// Checkout the specified branch
		branch := dep.Branch
		if branch == "" {
			branch = "main"
		}

		cmd = exec.Command("git", "-C", targetDir, "checkout", branch)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git checkout failed: %w", err)
		}

		// Pull latest changes
		cmd = exec.Command("git", "-C", targetDir, "pull", "origin", branch)
		cmd.Stdout = nil
		cmd.Stderr = nil
		_ = cmd.Run() // Pull might fail if no tracking branch, ignore for read-only access
	}

	return nil
}

// generateDirName generates a directory name from a Git URL
func generateDirName(url string) string {
	// Remove protocol and domain, extract repo name
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git@")

	// Replace : and / with -
	url = strings.ReplaceAll(url, ":", "-")
	url = strings.ReplaceAll(url, "/", "-")

	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	return url
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
