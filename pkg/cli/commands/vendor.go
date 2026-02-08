package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"specledger/internal/spec"

	"github.com/spf13/cobra"
)

// VarVendorCmd represents the vendor command
var VarVendorCmd = &cobra.Command{
	Use:   "vendor",
	Short: "Vendor dependencies for offline use",
}

// VarVendorAllCmd represents the vendor all command
var VarVendorAllCmd = &cobra.Command{
	Use:   "vendor --output <path>",
	Short: "Copy all dependencies to vendor directory",
	Long:  `Copy all external dependencies to the local vendor directory for offline use.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runVendorAll,
}

// VarVendorUpdateCmd represents the vendor update command
var VarVendorUpdateCmd = &cobra.Command{
	Use:   "update [--vendor-path <path>] [--force]",
	Short: "Update vendored dependencies",
	RunE:  runVendorUpdate,
}

// VarCleanCmd represents the clean command
var VarCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove vendored dependencies",
	RunE:  runVendorClean,
}

func init() {
	VarVendorCmd.AddCommand(VarVendorAllCmd, VarVendorUpdateCmd, VarCleanCmd)

	VarVendorAllCmd.Flags().StringP("output", "o", "specledger/vendor", "Output vendor directory path")
	VarVendorUpdateCmd.Flags().StringP("vendor-path", "p", "specledger/vendor", "Vendor directory path")
	VarVendorUpdateCmd.Flags().BoolP("force", "f", false, "Force update all vendored specs")
}

func runVendorAll(cmd *cobra.Command, args []string) error {
	vendorPath, _ := cmd.Flags().GetString("output")
	force, _ := cmd.Flags().GetBool("force")

	// Read lockfile
	lockfile, err := spec.ReadLockfile("specledger/spec.sum")
	if err != nil {
		return fmt.Errorf("failed to read lockfile: %w", err)
	}

	if len(lockfile.Entries) == 0 {
		return fmt.Errorf("no vendored specs found. Run 'sl deps resolve' first")
	}

	fmt.Printf("Vendoring %d spec(s) to %s...\n", len(lockfile.Entries), vendorPath)

	// Create vendor directory
	if err := os.MkdirAll(vendorPath, 0755); err != nil {
		return fmt.Errorf("failed to create vendor directory: %w", err)
	}

	// Vendor each entry
	for _, entry := range lockfile.Entries {
		if err := vendorEntry(entry, vendorPath, force); err != nil {
			return fmt.Errorf("failed to vendor %s: %w", entry.RepositoryURL, err)
		}
	}

	fmt.Printf("Successfully vendored %d spec(s)\n", len(lockfile.Entries))
	return nil
}

func runVendorUpdate(cmd *cobra.Command, args []string) error {
	vendorPath, _ := cmd.Flags().GetString("vendor-path")
	force, _ := cmd.Flags().GetBool("force")

	// Read lockfile
	lockfile, err := spec.ReadLockfile("specledger/spec.sum")
	if err != nil {
		return fmt.Errorf("failed to read lockfile: %w", err)
	}

	if len(lockfile.Entries) == 0 {
		return fmt.Errorf("no vendored specs found. Run 'sl deps resolve' first")
	}

	fmt.Printf("Updating %d vendored spec(s) in %s...\n", len(lockfile.Entries), vendorPath)

	// Update each entry
	for _, entry := range lockfile.Entries {
		if err := vendorEntry(entry, vendorPath, force); err != nil {
			return fmt.Errorf("failed to update %s: %w", entry.RepositoryURL, err)
		}
	}

	fmt.Printf("Successfully updated %d vendored spec(s)\n", len(lockfile.Entries))
	return nil
}

func runVendorClean(cmd *cobra.Command, args []string) error {
	vendorPath, _ := cmd.Flags().GetString("vendor-path")

	// Remove vendor directory
	if err := os.RemoveAll(vendorPath); err != nil {
		return fmt.Errorf("failed to remove vendor directory: %w", err)
	}

	fmt.Printf("Removed vendored specs from %s\n", vendorPath)
	return nil
}

// vendorEntry vendors a single spec entry
func vendorEntry(entry spec.LockfileEntry, vendorPath string, force bool) error {
	// Check if vendor directory exists for this dependency
	depVendorPath := filepath.Join(vendorPath, getVendorDirName(entry))

	// If force is not set and vendor already exists, skip
	if !force {
		if _, err := os.Stat(depVendorPath); err == nil {
			return nil
		}
	}

	// Create dependency vendor directory
	if err := os.MkdirAll(depVendorPath, 0755); err != nil {
		return fmt.Errorf("failed to create vendor directory: %w", err)
	}

	// Copy the spec file
	srcPath := filepath.Join(".spec-cache", getVendorDirName(entry), entry.SpecPath)
	dstPath := filepath.Join(depVendorPath, entry.SpecPath)

	if _, err := os.Stat(srcPath); err != nil {
		return fmt.Errorf("spec not found in cache: %s", srcPath)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to copy spec file: %w", err)
	}

	return nil
}

// getVendorDirName generates a vendor directory name from the repository URL
func getVendorDirName(entry spec.LockfileEntry) string {
	// Use Branch field if it looks like an alias (starts with #)
	if entry.Branch != "" && strings.HasPrefix(entry.Branch, "#") {
		return entry.Branch
	}

	// Extract repo name from URL
	parts := strings.Split(entry.RepositoryURL, "/")
	if len(parts) > 0 {
		repoName := parts[len(parts)-1]
		repoName = strings.TrimSuffix(repoName, ".git")
		return repoName
	}

	return "unknown"
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// ListVendorSpecs lists all vendored specs
func ListVendorSpecs(vendorPath string) ([]string, error) {
	var specs []string

	err := filepath.WalkDir(vendorPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && d.Name() == "spec.md" {
			relPath, err := filepath.Rel(vendorPath, path)
			if err != nil {
				return err
			}
			specs = append(specs, relPath)
		}

		return nil
	})

	return specs, err
}
