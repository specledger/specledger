package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"specledger/internal/spec"
	"specledger/pkg/models"
)

// VarDepsCmd represents the deps command
var VarDepsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Manage specification dependencies",
}

// VarAddCmd represents the add command
var VarAddCmd = &cobra.Command{
	Use:   "add <repo-url> [branch] [spec-path] [--alias <name>]",
	Short: "Add a dependency to the manifest",
	Long:  `Add an external specification dependency to the current project's spec.mod file.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAddDependency,
}

// VarDepsListCmd represents the list command
var VarDepsListCmd = &cobra.Command{
	Use:   "list [--include-transitive]",
	Short: "List all declared dependencies",
	RunE:  runListDependencies,
}

// VarResolveCmd represents the resolve command
var VarResolveCmd = &cobra.Command{
	Use:   "resolve [--no-cache] [--deep]",
	Short: "Resolve all dependencies and generate spec.sum",
	Long:  `Fetch external specifications, validate versions, and generate the lockfile with cryptographic hashes.`,
	RunE:  runResolveDependencies,
}

// VarDepsUpdateCmd represents the update command
var VarDepsUpdateCmd = &cobra.Command{
	Use:   "update [--force] [repo-url]",
	Short: "Update dependencies to latest compatible versions",
	RunE:  runUpdateDependencies,
}

// VarRemoveCmd represents the remove command
var VarRemoveCmd = &cobra.Command{
	Use:   "remove <repo-url> <spec-path>",
	Short: "Remove a dependency from the manifest",
	Args:  cobra.ExactArgs(2),
	RunE:  runRemoveDependency,
}

func init() {
	VarDepsCmd.AddCommand(VarAddCmd, VarDepsListCmd, VarResolveCmd, VarDepsUpdateCmd, VarRemoveCmd)

	VarAddCmd.Flags().StringP("alias", "a", "", "Optional alias for the dependency")
	VarDepsListCmd.Flags().BoolP("include-transitive", "t", false, "Include transitive dependencies")
	VarResolveCmd.Flags().BoolP("no-cache", "n", false, "Ignore cached specifications")
	VarResolveCmd.Flags().BoolP("deep", "d", false, "Fetch full git history")
	VarDepsUpdateCmd.Flags().BoolP("force", "f", false, "Force update all dependencies")
}

func runAddDependency(cmd *cobra.Command, args []string) error {
	// Extract flags
	alias, _ := cmd.Flags().GetString("alias")

	// Parse arguments
	repoURL := args[0]
	version := "main" // default
	specPath := "spec.md"

	if len(args) >= 2 {
		version = args[1]
	}
	if len(args) >= 3 {
		specPath = args[2]
	}

	// Validate URL
	if !isValidURL(repoURL) {
		return fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	// Create dependency
	dep := models.Dependency{
		RepositoryURL: repoURL,
		Version:       version,
		SpecPath:      specPath,
		Alias:         alias,
	}

	// Validate
	if err := dep.Validate(); err != nil {
		return fmt.Errorf("invalid dependency: %w", err)
	}

	// Read existing manifest
	manifestPath := "specs/spec.mod"
	manifest, err := spec.ParseManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Add dependency (manually append for now)
	manifest.Dependecies = append(manifest.Dependecies, dep)
	manifest.UpdatedAt = time.Now()

	// Write manifest
	if err := spec.WriteManifest(manifestPath, manifest); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	fmt.Printf("Added dependency: %s -> %s\n", repoURL, specPath)
	if alias != "" {
		fmt.Printf("  Alias: %s\n", alias)
	}

	return nil
}

func runListDependencies(cmd *cobra.Command, args []string) error {
	manifestPath := "specs/spec.mod"
	manifest, err := spec.ParseManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	fmt.Printf("Dependencies (%d):\n", len(manifest.Dependecies))
	fmt.Println()

	for i, dep := range manifest.Dependecies {
		fmt.Printf("%d. %s\n", i+1, dep.RepositoryURL)
		fmt.Printf("   Version: %s\n", dep.Version)
		fmt.Printf("   Spec: %s\n", dep.SpecPath)
		if dep.Alias != "" {
			fmt.Printf("   Alias: %s\n", dep.Alias)
		}
		fmt.Println()
	}

	return nil
}

func runResolveDependencies(cmd *cobra.Command, args []string) error {
	noCache, _ := cmd.Flags().GetBool("no-cache")
	_, _ = cmd.Flags().GetBool("deep") // deep flag is not used yet

	manifestPath := "specs/spec.mod"
	lockfilePath := "specs/spec.sum"

	// Read manifest
	manifest, err := spec.ParseManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Validate manifest
	errors := spec.ValidateManifest(manifest)
	if len(errors) > 0 {
		fmt.Println("Validation errors:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err.Error())
		}
		return fmt.Errorf("%d validation errors found", len(errors))
	}

	fmt.Printf("Resolving %d dependencies...\n", len(manifest.Dependecies))

	// Create resolver
	resolver := spec.NewResolver(".spec-cache")

	// Resolve dependencies
	results, err := resolver.Resolve(cmd.Context(), manifest, noCache)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	fmt.Println()

	// Create lockfile
	lockfile := spec.NewLockfile(spec.ManifestVersion)
	for _, result := range results {
		entry := spec.LockfileEntry{
			RepositoryURL: result.Dependency.RepositoryURL,
			CommitHash:    result.CommitHash,
			ContentHash:   result.ContentHash,
			SpecPath:      result.Dependency.SpecPath,
			Branch:        result.Dependency.Version,
			Size:          result.Size,
			FetchedAt:     time.Now().Format(time.RFC3339),
		}
		lockfile.AddEntry(entry)

		fmt.Printf("✓ Resolved: %s\n", result.Dependency.RepositoryURL)
		fmt.Printf("  Commit: %s\n", result.CommitHash)
		fmt.Printf("  Spec: %s\n", result.Dependency.SpecPath)
		fmt.Printf("  Hash: %s\n", result.ContentHash)
		fmt.Println()
	}

	// Write lockfile
	if err := lockfile.Write(lockfilePath); err != nil {
		return fmt.Errorf("failed to write lockfile: %w", err)
	}

	fmt.Printf("Lockfile written to: %s\n", lockfilePath)
	fmt.Printf("Total size: %d bytes\n", lockfile.TotalSize)

	return nil
}

func runUpdateDependencies(cmd *cobra.Command, args []string) error {
	fmt.Println("Updating dependencies...")
	fmt.Println("This feature will be implemented in a future version.")
	return nil
}

func runRemoveDependency(cmd *cobra.Command, args []string) error {
	repoURL := args[0]
	specPath := args[1]

	manifestPath := "specs/spec.mod"
	manifest, err := spec.ParseManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	removed := false
	for i, dep := range manifest.Dependecies {
		if dep.RepositoryURL == repoURL && dep.SpecPath == specPath {
			manifest.Dependecies = append(manifest.Dependecies[:i], manifest.Dependecies[i+1:]...)
			removed = true
			break
		}
	}

	if !removed {
		return fmt.Errorf("dependency not found: %s %s", repoURL, specPath)
	}

	manifest.UpdatedAt = time.Now()

	// Write manifest
	if err := spec.WriteManifest(manifestPath, manifest); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	fmt.Printf("Removed dependency: %s from %s\n", repoURL, specPath)

	return nil
}

func isValidURL(s string) bool {
	// Simple check for common Git URLs
	return len(s) > 0 && (strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "git@"))
}
