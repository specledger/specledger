package commands

import (
	"fmt"

	"github.com/specledger/specledger/pkg/cli/playbooks"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/spf13/cobra"
)

// VarPlaybookCmd represents the playbook command
var VarPlaybookCmd = &cobra.Command{
	Use:   "playbook",
	Short: "Manage SDD playbooks",
	Long: `Manage embedded SDD playbooks for SpecLedger.

Playbooks contain Claude Code commands, skills, scripts, and templates
for SpecLedger projects.

Examples:
  sl playbook list              List all available playbooks
  sl playbook list --json       List playbooks in JSON format`,
}

func init() {
	VarPlaybookCmd.AddCommand(VarPlaybookListCmd)
}

// VarPlaybookListCmd represents the list command
var VarPlaybookListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List available SDD playbooks",
	Long:    `List all available embedded SDD playbooks with their names, descriptions, frameworks, and versions.`,
	Example: `  sl playbook list`,
	RunE:    runListPlaybooks,
}

var playbookJSONFlag bool

func init() {
	VarPlaybookListCmd.Flags().BoolVar(&playbookJSONFlag, "json", false, "Output in JSON format")
}

// runListPlaybooks lists all available playbooks
func runListPlaybooks(cmd *cobra.Command, args []string) error {
	ui.PrintSection("Available Playbooks")

	playbookList, err := playbooks.ListPlaybooks()
	if err != nil {
		return fmt.Errorf("failed to list playbooks: %w", err)
	}

	if playbookJSONFlag {
		fmt.Println(playbooks.FormatJSON(playbookList))
	} else {
		fmt.Println(playbooks.FormatTable(playbookList))
	}

	fmt.Println()
	fmt.Printf("Total: %d playbook(s)\n", len(playbookList))
	fmt.Println()
	fmt.Printf("Next: %s\n", ui.Cyan("sl new"))

	return nil
}
