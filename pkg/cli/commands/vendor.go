package commands

import (
	"github.com/spf13/cobra"
)

// VarVendorCmd represents the vendor command
var VarVendorCmd = &cobra.Command{
	Use:   "vendor",
	Short: "Vendor dependencies for offline use",
}

// VarVendorAllCmd represents the vendor command
var VarVendorAllCmd = &cobra.Command{
	Use:   "vendor --output <path>",
	Short: "Copy all dependencies to vendor directory",
	Long:  `Copy all external dependencies to the local vendor directory for offline use.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runVendorAll,
}

// VarVendorUpdateCmd represents the update command
var VarVendorUpdateCmd = &cobra.Command{
	Use:   "update [--vendor-path <path>] [--force]",
	Short: "Update vendored dependencies",
	Run:   runVendorUpdate,
}

// VarCleanCmd represents the clean command
var VarCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove vendored dependencies",
	Run:   runVendorClean,
}

func init() {
	VarVendorCmd.AddCommand(VarVendorAllCmd, VarVendorUpdateCmd, VarCleanCmd)

	VarVendorAllCmd.Flags().StringP("output", "o", "specs/vendor", "Output vendor directory path")
	VarVendorUpdateCmd.Flags().StringP("vendor-path", "p", "specs/vendor", "Vendor directory path")
	VarVendorUpdateCmd.Flags().BoolP("force", "f", false, "Force update all vendored specs")
}

func runVendorAll(cmd *cobra.Command, args []string) {
	output, _ := cmd.Flags().GetString("output")
	cmd.Printf("Vendor all dependencies to %s\n", output)
}

func runVendorUpdate(cmd *cobra.Command, args []string) {
	vendorPath, _ := cmd.Flags().GetString("vendor-path")
	force, _ := cmd.Flags().GetBool("force")
	cmd.Printf("Updating vendored dependencies in %s (force: %v)\n", vendorPath, force)
}

func runVendorClean(cmd *cobra.Command, args []string) {
	cmd.Println("Cleaning vendor directory...")
}
