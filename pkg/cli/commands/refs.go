package commands

import (
	"fmt"
	"os"
	"strings"

	"specledger/internal/ref"
	"specledger/internal/spec"

	"github.com/spf13/cobra"
)

// VarRefsCmd represents the refs command
var VarRefsCmd = &cobra.Command{
	Use:   "refs",
	Short: "Validate external specification references",
}

// VarValidateCmd represents the validate command
var VarValidateCmd = &cobra.Command{
	Use:   "validate [--strict] [--spec-path <path>]",
	Short: "Validate all external references in a specification",
	Long:  `Validate all external references in spec.md files against resolved dependencies.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runValidateReferences,
}

// VarListCmd represents the list command
var VarListCmd = &cobra.Command{
	Use:   "list [--spec-path <path>]",
	Short: "List all external references",
	RunE:  runListReferences,
}

func init() {
	VarRefsCmd.AddCommand(VarValidateCmd, VarListCmd)

	VarValidateCmd.Flags().BoolP("strict", "s", false, "Treat warnings as errors")
	VarValidateCmd.Flags().StringP("spec-path", "p", "spec.md", "Path to the specification file")
	VarListCmd.Flags().StringP("spec-path", "p", "spec.md", "Path to the specification file")
}

func runValidateReferences(cmd *cobra.Command, args []string) error {
	specPath, _ := cmd.Flags().GetString("spec-path")
	strict, _ := cmd.Flags().GetBool("strict")

	// Read spec file
	content, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}

	// Read lockfile
	lockfile, err := spec.ReadLockfile("specledger/spec.sum")
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no lockfile found. Run 'sl deps resolve' first")
		}
		return fmt.Errorf("failed to read lockfile: %w", err)
	}

	// Build dependency mapping from lockfile entries
	dependencies := make(map[string]string)
	for _, entry := range lockfile.Entries {
		// Try to use alias if available, otherwise use repository URL
		alias := getAliasForEntry(entry, lockfile.Entries)
		if alias != "" {
			dependencies[alias] = entry.RepositoryURL
		} else {
			dependencies[entry.RepositoryURL] = entry.RepositoryURL
		}
	}

	// Create resolver and set dependencies
	resolver := ref.NewResolver("specledger/spec.sum")
	resolver.SetDependencies(dependencies)

	// Parse references
	references, err := resolver.ParseSpec(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse references: %w", err)
	}

	fmt.Printf("Found %d references in %s\n", len(references), specPath)
	fmt.Println()

	// Validate references
	errors := resolver.ValidateReferences(references)

	if len(errors) > 0 {
		fmt.Println("Validation errors:")
		for _, err := range errors {
			fmt.Printf("  ✗ %s\n", err.Error())
		}
		fmt.Printf("\n%d error(s) found\n", len(errors))

		if strict {
			return fmt.Errorf("%d validation error(s) found (strict mode)", len(errors))
		}
		return fmt.Errorf("%d validation error(s) found", len(errors))
	}

	fmt.Println("All references validated successfully! ✓")
	return nil
}

func runListReferences(cmd *cobra.Command, args []string) error {
	specPath, _ := cmd.Flags().GetString("spec-path")

	// Read spec file
	content, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}

	// Read lockfile
	lockfile, err := spec.ReadLockfile("specledger/spec.sum")
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no lockfile found. Run 'sl deps resolve' first")
		}
		return fmt.Errorf("failed to read lockfile: %w", err)
	}

	// Build dependency mapping from lockfile entries
	dependencies := make(map[string]string)
	for _, entry := range lockfile.Entries {
		alias := getAliasForEntry(entry, lockfile.Entries)
		if alias != "" {
			dependencies[alias] = entry.RepositoryURL
		} else {
			dependencies[entry.RepositoryURL] = entry.RepositoryURL
		}
	}

	// Create resolver and set dependencies
	resolver := ref.NewResolver("specledger/spec.sum")
	resolver.SetDependencies(dependencies)

	// Parse references
	references, err := resolver.ParseSpec(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse references: %w", err)
	}

	fmt.Printf("References in %s (%d total):\n\n", specPath, len(references))

	for i, ref := range references {
		fmt.Printf("%d. [%s] %s\n", i+1, ref.Type, ref.URL)
		fmt.Printf("   Text: %s\n", ref.Text)
		fmt.Printf("   Line: %d\n\n", ref.Line)
	}

	return nil
}

// getAliasForEntry finds an alias for a given entry
func getAliasForEntry(entry spec.LockfileEntry, allEntries []spec.LockfileEntry) string {
	for _, e := range allEntries {
		if e.RepositoryURL == entry.RepositoryURL && e.SpecPath == entry.SpecPath {
			// Check if the entry has an explicit alias in Branch field (we use it as alias)
			if e.Branch != "" && strings.HasPrefix(e.Branch, "#") {
				return e.Branch
			}
		}
	}
	return ""
}
