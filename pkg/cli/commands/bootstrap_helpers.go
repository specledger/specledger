package commands

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"specledger/pkg/cli/playbooks"
	"specledger/pkg/cli/ui"
	"specledger/pkg/embedded"
)

// applyEmbeddedPlaybooks copies embedded playbooks to the project directory.
// If playbookName is empty, uses the default playbook.
// Returns the playbook name, version, and structure for metadata storage.
func applyEmbeddedPlaybooks(projectPath string, playbookName string) (string, string, []string, error) {
	ui.PrintSection("Copying Playbooks")
	fmt.Printf("Applying SpecLedger playbooks...\n")

	pbName, pbVersion, pbStructure, err := playbooks.ApplyToProject(projectPath, playbookName)
	if err != nil {
		// Playbooks are helpful but not critical - log warning and continue
		ui.PrintWarning(fmt.Sprintf("Playbook copying failed: %v", err))
		ui.PrintWarning("Project will be created without playbooks")
		return "", "", nil, nil
	}

	fmt.Printf("%s Playbooks applied\n", ui.Checkmark())

	// Trust mise.toml if it exists
	trustMiseConfig(projectPath)

	return pbName, pbVersion, pbStructure, nil
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

// applyEmbeddedSkills copies embedded skills and commands to the project.
// These provide Claude with context for SpecLedger capabilities.
func applyEmbeddedSkills(projectPath string) error {
	// Target directory is .claude in the project root
	targetDir := filepath.Join(projectPath, ".claude")

	// Walk through the skills embedded filesystem
	err := fs.WalkDir(embedded.SkillsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == "." || path == ".claude" {
			return nil
		}

		// Skip directories (they'll be created when files are written)
		if d.IsDir() {
			return nil
		}

		// Calculate destination path
		relPath := strings.TrimPrefix(path, ".")
		destPath := filepath.Join(targetDir, relPath)

		// Ensure parent directory exists
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destDir, err)
		}

		// Read file from embedded FS
		data, err := embedded.SkillsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// Write to destination
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to copy embedded skills: %w", err)
	}

	return nil
}
