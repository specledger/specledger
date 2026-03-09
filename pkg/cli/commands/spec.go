package commands

import (
	"github.com/spf13/cobra"
)

var VarSpecCmd = &cobra.Command{
	Use:   "spec",
	Short: "Manage feature specifications",
	Long: `Manage feature specifications and feature context.

Commands:
  info        Get feature paths and prerequisite validation
  create      Create a new feature branch and spec directory
  setup-plan  Copy plan template to feature directory

Examples:
  sl spec info --json                    # Get feature info as JSON
  sl spec create --number 600 --short-name "test-feature"  # Create new feature
  sl spec setup-plan                     # Setup plan.md from template`,
}

func NewSpecCmd() *cobra.Command {
	return VarSpecCmd
}
