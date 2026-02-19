package main

import (
	"fmt"
	"os"

	"github.com/specledger/specledger/pkg/cli/commands"

	"github.com/spf13/cobra"
)

// Version variables set by GoReleaser during build
var (
	version   = "dev"
	commit    = "unknown"
	date      = "unknown"
	buildType = "development"
)

var rootCmd = &cobra.Command{
	Use:   "sl",
	Short: "SpecLedger CLI - Bootstrap projects and manage spec dependencies",
	Long: `SpecLedger (sl) helps you:

1. Create new projects with interactive TUI or flags
2. Initialize SpecLedger in existing repositories
3. Manage specification dependencies (add, remove, list)
4. Cache dependencies locally for offline use
5. View dependency graphs and relationships

Quick start:
 sl new              # Create a new project (interactive)
 sl init             # Initialize in existing repository
 sl deps list        # List dependencies
 sl deps add <url>   # Add a dependency`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	// Add subcommands - only the essential ones
	rootCmd.AddCommand(commands.VarBootstrapCmd)
	rootCmd.AddCommand(commands.VarInitCmd)
	rootCmd.AddCommand(commands.VarDepsCmd)
	rootCmd.AddCommand(commands.VarGraphCmd)
	rootCmd.AddCommand(commands.VarDoctorCmd)
	rootCmd.AddCommand(commands.VarPlaybookCmd)
	rootCmd.AddCommand(commands.VarAuthCmd)
	rootCmd.AddCommand(commands.VarSessionCmd)
	rootCmd.AddCommand(commands.VarIssueCmd)

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display the version, commit, and build date of SpecLedger",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("SpecLedger CLI (sl)\n")
			fmt.Printf("Version:  %s\n", version)
			fmt.Printf("Commit:   %s\n", commit)
			fmt.Printf("Built:    %s\n", date)
			fmt.Printf("Type:     %s\n", buildType)
		},
	})

	// Disable default command completion (sl specledger alias)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
