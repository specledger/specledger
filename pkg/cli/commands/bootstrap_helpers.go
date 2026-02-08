package commands

import (
	"fmt"

	"specledger/pkg/cli/playbooks"
	"specledger/pkg/cli/ui"
)

// applyEmbeddedPlaybooks copies embedded playbooks to the project directory.
func applyEmbeddedPlaybooks(projectPath string) error {
	ui.PrintSection("Copying Playbooks")
	fmt.Printf("Applying SpecLedger playbooks...\n")

	if err := playbooks.ApplyToProject(projectPath, ""); err != nil {
		// Playbooks are helpful but not critical - log warning and continue
		ui.PrintWarning(fmt.Sprintf("Playbook copying failed: %v", err))
		ui.PrintWarning("Project will be created without playbooks")
		return nil
	}

	fmt.Printf("%s Playbooks applied\n", ui.Checkmark())
	return nil
}
