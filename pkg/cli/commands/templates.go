package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"specledger/pkg/cli/templates"
	"specledger/pkg/cli/ui"
)

// VarTemplateCmd represents the template command
var VarTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage template playbooks",
	Long: `Manage embedded template playbooks for SpecLedger.

Examples:
  sl template list              List all available templates
  sl template list --json       List templates in JSON format`,
}

func init() {
	VarTemplateCmd.AddCommand(VarTemplateListCmd)
}

// VarTemplateListCmd represents the list command
var VarTemplateListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List available template playbooks",
	Long:    `List all available embedded template playbooks with their names, descriptions, frameworks, and versions.`,
	Example: `  sl template list`,
	RunE:    runListTemplates,
}

var templateJSONFlag bool

func init() {
	VarTemplateListCmd.Flags().BoolVar(&templateJSONFlag, "json", false, "Output in JSON format")
}

// runListTemplates lists all available templates
func runListTemplates(cmd *cobra.Command, args []string) error {
	ui.PrintSection("Available Templates")

	templatesList, err := templates.ListTemplates()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if templateJSONFlag {
		fmt.Println(templates.FormatJSON(templatesList))
	} else {
		fmt.Println(templates.FormatTable(templatesList))
	}

	fmt.Println()
	fmt.Printf("Total: %d template(s)\n", len(templatesList))
	fmt.Println()
	fmt.Printf("Next: %s\n", ui.Cyan("sl new --framework <name>"))

	return nil
}
