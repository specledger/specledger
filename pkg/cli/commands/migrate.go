package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"specledger/pkg/cli/metadata"
	"specledger/pkg/cli/ui"
)

// VarMigrateCmd represents the migrate command
var VarMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate .mod files to YAML format",
	Long: `Migrate existing specledger.mod files to the new specledger.yaml format.

This command converts legacy .mod files to the new YAML metadata format.
The original .mod file is preserved for backup purposes.

Migration rules:
- Project name and short code are extracted from .mod comments
- Framework choice defaults to 'none' (you can edit the YAML later)
- Dependencies are preserved (if any were declared)
- Original .mod file is kept as specledger.spec.mod.backup

Examples:
  sl migrate           # Migrate in current project directory
  sl migrate --dry-run # Preview changes without writing`,
	RunE: runMigrate,
}

var migrateDryRun bool

func init() {
	VarMigrateCmd.Flags().BoolVarP(&migrateDryRun, "dry-run", "d", false, "Preview changes without writing files")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	// Get current directory (or specified directory)
	projectDir := "."
	if len(args) > 0 {
		projectDir = args[0]
	}

	// Resolve to absolute path
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if .mod file exists
	if !metadata.HasLegacyModFile(absDir) {
		return fmt.Errorf("no specledger.mod file found in %s", absDir)
	}

	// Check if YAML already exists
	if metadata.HasYAMLMetadata(absDir) {
		return fmt.Errorf("specledger.yaml already exists. Use --force to overwrite (not implemented yet)")
	}

	ui.PrintSection("Migrating to YAML Format")
	fmt.Printf("Directory: %s\n", ui.Bold(absDir))
	fmt.Println()

	// Parse the .mod file to show preview
	modPath := filepath.Join(absDir, "specledger", "specledger.mod")
	modData, err := metadata.ParseModFile(modPath)
	if err != nil {
		return fmt.Errorf("failed to parse .mod file: %w", err)
	}

	// Show what would be done
	if migrateDryRun {
		ui.PrintSection("Dry Run Preview")
		fmt.Printf("  Project:     %s\n", ui.Bold(modData.ProjectName))
		fmt.Printf("  Short Code:  %s\n", ui.Bold(modData.ShortCode))
		fmt.Printf("  Framework:   %s\n", ui.Bold("none (default for migrated projects)"))
		fmt.Printf("  Dependencies: %s\n", ui.Bold("0"))
		fmt.Println()
		fmt.Printf("Backup: %s\n", ui.Cyan("specledger.spec.mod.backup"))
		fmt.Println()
		return nil
	}

	// Perform migration
	meta, err := metadata.MigrateModToYAML(absDir)
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	// Backup the .mod file
	backupPath := filepath.Join(absDir, "specledger", "specledger.spec.mod.backup")
	if err := os.Rename(modPath, backupPath); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to backup .mod file: %v", err))
	} else {
		fmt.Printf("  Backup: %s\n", ui.Cyan(backupPath))
	}

	fmt.Println()
	ui.PrintSuccess("Migration Complete!")
	fmt.Printf("  Project:   %s (short code: %s)\n", ui.Bold(meta.Project.Name), ui.Bold(meta.Project.ShortCode))
	fmt.Printf("  Playbook:  %s\n", ui.Bold(meta.Playbook.Name))
	fmt.Println()
	fmt.Println(ui.Bold("Next steps:"))
	fmt.Println("  1. Review the generated specledger.yaml")
	fmt.Printf("  2. Optionally remove the backup file: %s\n", ui.Cyan("rm specledger.spec.mod.backup"))
	fmt.Println()

	return nil
}
