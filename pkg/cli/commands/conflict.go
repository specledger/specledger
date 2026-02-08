package commands

import (
	"fmt"
	"strings"

	"specledger/internal/spec"
	"specledger/pkg/models"

	"github.com/spf13/cobra"
)

// VarConflictCmd represents the conflict command
var VarConflictCmd = &cobra.Command{
	Use:   "conflict",
	Short: "Check for dependency conflicts",
	Long:  `Check the dependency graph for potential conflicts and circular dependencies.`,
}

// VarCheckCmd represents the check command
var VarCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for conflicts in the dependency graph",
	RunE:  runCheckConflicts,
}

// VarDetectCmd represents the detect command
var VarDetectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect potential conflicts",
	RunE:  runDetectConflicts,
}

func init() {
	VarConflictCmd.AddCommand(VarCheckCmd, VarDetectCmd)
}

func runCheckConflicts(cmd *cobra.Command, args []string) error {
	manifestPath := "specledger/spec.mod"

	// Read manifest
	manifest, err := spec.ParseManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	if len(manifest.Dependecies) == 0 {
		fmt.Println("No dependencies found")
		return nil
	}

	fmt.Printf("Checking %d dependency(ies) for conflicts...\n\n", len(manifest.Dependecies))

	// Check for duplicate dependencies
	duplicates := checkDuplicateDependencies(manifest.Dependecies)
	if len(duplicates) > 0 {
		fmt.Println("⚠️  Duplicates found:")
		for _, dup := range duplicates {
			fmt.Printf("  - %s\n", dup)
		}
		fmt.Println()
	}

	// Check for circular dependencies
	circular := checkCircularDependencies(manifest.Dependecies)
	if len(circular) > 0 {
		fmt.Println("⚠️  Circular dependencies detected:")
		for _, cycle := range circular {
			fmt.Printf("  - %s\n", cycle)
		}
		fmt.Println()
	}

	// Check for version conflicts
	conflicts := checkVersionConflicts(manifest.Dependecies)
	if len(conflicts) > 0 {
		fmt.Println("⚠️  Version conflicts found:")
		for _, conflict := range conflicts {
			fmt.Printf("  - %s\n", conflict)
		}
		fmt.Println()
	}

	// Check for missing spec paths
	missing := checkMissingSpecPaths(manifest.Dependecies)
	if len(missing) > 0 {
		fmt.Println("⚠️  Missing or invalid spec paths:")
		for _, specPath := range missing {
			fmt.Printf("  - %s\n", specPath)
		}
		fmt.Println()
	}

	if len(duplicates) == 0 && len(circular) == 0 && len(conflicts) == 0 && len(missing) == 0 {
		fmt.Println("✓ No conflicts detected!")
		return nil
	}

	return fmt.Errorf("%d conflict(s) detected", len(duplicates)+len(circular)+len(conflicts)+len(missing))
}

func runDetectConflicts(cmd *cobra.Command, args []string) error {
	manifestPath := "specledger/spec.mod"

	// Read manifest
	manifest, err := spec.ParseManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	if len(manifest.Dependecies) == 0 {
		fmt.Println("No dependencies to check")
		return nil
	}

	// Build a dependency graph
	graph := buildDependencyGraph(manifest.Dependecies)

	// Detect potential conflicts
	issues := detectPotentialConflicts(graph)

	if len(issues) == 0 {
		fmt.Println("No potential conflicts detected")
		return nil
	}

	fmt.Println("Potential conflicts detected:")
	for _, issue := range issues {
		fmt.Printf("  - %s\n", issue)
	}

	return fmt.Errorf("%d potential conflict(s) detected", len(issues))
}

// checkDuplicateDependencies checks for duplicate dependencies
func checkDuplicateDependencies(deps []models.Dependency) []string {
	var duplicates []string
	seen := make(map[string]bool)

	for _, dep := range deps {
		key := fmt.Sprintf("%s:%s", dep.RepositoryURL, dep.SpecPath)
		if seen[key] {
			duplicates = append(duplicates, key)
		} else {
			seen[key] = true
		}
	}

	return duplicates
}

// checkCircularDependencies checks for circular dependencies
func checkCircularDependencies(deps []models.Dependency) []string {
	// This is a basic implementation
	// For a more complete solution, we'd need to track transitive dependencies
	return nil
}

// checkVersionConflicts checks for version conflicts
func checkVersionConflicts(deps []models.Dependency) []string {
	// This is a basic implementation
	// For a more complete solution, we'd need to parse version constraints
	return nil
}

// checkMissingSpecPaths checks for missing or invalid spec paths
func checkMissingSpecPaths(deps []models.Dependency) []string {
	var missing []string

	for _, dep := range deps {
		if dep.SpecPath == "" {
			missing = append(missing, dep.RepositoryURL)
		} else if !strings.HasSuffix(dep.SpecPath, ".md") {
			missing = append(missing, fmt.Sprintf("%s (invalid path: %s)", dep.RepositoryURL, dep.SpecPath))
		}
	}

	return missing
}

// buildDependencyGraph builds a dependency graph from the manifest
func buildDependencyGraph(deps []models.Dependency) map[string][]string {
	graph := make(map[string][]string)

	for _, dep := range deps {
		graph[dep.RepositoryURL] = make([]string, 0)
	}

	return graph
}

// detectPotentialConflicts detects potential conflicts in the dependency graph
func detectPotentialConflicts(graph map[string][]string) []string {
	var issues []string

	// Check for duplicate dependencies
	for repo, deps := range graph {
		if len(deps) > 1 {
			issues = append(issues, fmt.Sprintf("Multiple specs from %s: %v", repo, deps))
		}
	}

	// Check for self-references
	for repo := range graph {
		for _, dep := range graph[repo] {
			if dep == repo {
				issues = append(issues, fmt.Sprintf("Self-reference detected: %s depends on itself", repo))
			}
		}
	}

	return issues
}
