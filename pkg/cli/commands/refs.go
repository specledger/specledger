package commands

import (
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
	Run:   runValidateReferences,
}

// VarListCmd represents the list command
var VarListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all external references",
	Run:   runListReferences,
}

func init() {
	VarRefsCmd.AddCommand(VarValidateCmd, VarListCmd)

	VarValidateCmd.Flags().BoolP("strict", "s", false, "Treat warnings as errors")
	VarValidateCmd.Flags().StringP("spec-path", "p", "spec.md", "Path to the specification file")
}

func runValidateReferences(cmd *cobra.Command, args []string) {
	cmd.Println("Validating references...")
}

func runListReferences(cmd *cobra.Command, args []string) {
	cmd.Println("Listing references...")
}
