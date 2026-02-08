package commands

import (
	"fmt"
	"os/exec"

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

	// Trust mise.toml if it exists
	trustMiseConfig(projectPath)

	return nil
}

// trustMiseConfig runs `mise trust` on the project's mise.toml file.
func trustMiseConfig(projectPath string) {
	misePath := projectPath + "/mise.toml"
	cmd := exec.Command("mise", "trust", misePath)
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		// mise trust failing is not critical - mise will prompt user to trust on first use
		ui.PrintWarning(fmt.Sprintf("Could not trust mise.toml: %v", err))
		ui.PrintWarning("Run 'mise trust' to enable mise tools")
	}
}
