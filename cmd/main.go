package main

import (
	"os"

	"github.com/spf13/cobra"
	"specledger/pkg/cli/commands"
)

var rootCmd = &cobra.Command{
	Use:   "specledger",
	Short: "SpecLedger - Specification dependency management",
	Long: `SpecLedger is a tool for managing external specification dependencies
across repositories. It enables teams to declare dependencies, resolve them
with cryptographic verification, and reference specific sections from external
specifications.`,
	Version: "1.0.0",
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add command groups
	rootCmd.AddGroup(&cobra.Group{
		ID:    "core",
		Title: "Core Commands",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "deps",
		Title: "Dependencies",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "refs",
		Title: "References",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "graph",
		Title: "Graph",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "vendor",
		Title: "Vendor",
	})

	// Add subcommands
	rootCmd.AddCommand(commands.VarDepsCmd)
	rootCmd.AddCommand(commands.VarRefsCmd)
	rootCmd.AddCommand(commands.VarGraphCmd)
	rootCmd.AddCommand(commands.VarVendorCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
