package commands

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/specledger/specledger/pkg/cli/metadata"
	"github.com/specledger/specledger/pkg/cli/playbooks"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/specledger/specledger/pkg/embedded"
)

// applyEmbeddedPlaybooks copies embedded playbooks to the project directory.
// If playbookName is empty, uses the default playbook.
// If force is true, existing files will be overwritten.
// Returns the playbook name, version, and structure for metadata storage.
func applyEmbeddedPlaybooks(projectPath string, playbookName string, force bool) (string, string, []string, error) {
	ui.PrintSection("Copying Playbooks")
	fmt.Printf("Applying SpecLedger playbooks...\n")

	pbName, pbVersion, pbStructure, err := playbooks.ApplyToProject(projectPath, playbookName, force)
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

		// Skip the root directory and skills wrapper
		if path == "." || path == "skills" {
			return nil
		}

		// Skip directories (they'll be created when files are written)
		if d.IsDir() {
			return nil
		}

		// Remove "skills/" prefix to get relative path from commands/ and skills/
		relPath := strings.TrimPrefix(path, "skills/")
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
		// #nosec G306 -- skill files need to be readable, 0644 is appropriate
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

// setupSpecLedgerProject applies playbooks, skills, and creates metadata.
// Optionally initializes git based on flags.
// If force is true, existing files will be overwritten.
// Returns the playbook name, version, and structure for metadata storage.
func setupSpecLedgerProject(projectPath, projectName, shortCode, playbookName string, initGit bool, force bool) (string, string, []string, error) {
	// Apply embedded playbooks
	selectedPlaybookName, playbookVersion, playbookStructure, err := applyEmbeddedPlaybooks(projectPath, playbookName, force)
	if err != nil {
		// Playbook application failure is not fatal - log warning and continue
		fmt.Printf("Warning: playbook application had issues: %v\n", err)
	}

	// Apply embedded skills
	if err := applyEmbeddedSkills(projectPath); err != nil {
		// Skills are helpful but not critical - log warning and continue
		fmt.Printf("Warning: skills installation had issues: %v\n", err)
	}

	// Create YAML metadata with playbook info
	projectMetadata := metadata.NewProjectMetadata(projectName, shortCode, selectedPlaybookName, playbookVersion, playbookStructure)
	if err := metadata.SaveToProject(projectMetadata, projectPath); err != nil {
		return "", "", nil, fmt.Errorf("failed to create project metadata: %w", err)
	}

	// Run post-init script BEFORE git init (so generated files are included)
	runPostInitScript(projectPath, projectMetadata)

	// Initialize git if requested (bootstrap only)
	// This runs AFTER post-init so generated files (like .beads/) are staged
	if initGit {
		if err := initializeGitRepo(projectPath); err != nil {
			return "", "", nil, fmt.Errorf("failed to initialize git: %w", err)
		}
	}

	return selectedPlaybookName, playbookVersion, playbookStructure, nil
}

// runPostInitScript executes the template's init.sh script if it exists.
// This allows templates to perform post-initialization tasks like setting up beads.
// Passes specledger.yaml data as environment variables for use in scripts.
// The init.sh script is read from embedded templates (not copied to target project).
func runPostInitScript(projectPath string, projectMetadata *metadata.ProjectMetadata) {
	// Look for init.sh in the embedded templates for the selected playbook
	playbookName := projectMetadata.Playbook.Name
	if playbookName == "" {
		return
	}

	// Path to init.sh in embedded templates
	initScriptPath := filepath.Join("templates", playbookName, "init.sh")

	// Check if init.sh exists in embedded templates
	scriptContent, err := embedded.TemplatesFS.ReadFile(initScriptPath)
	if err != nil {
		// Script doesn't exist in this template, skip silently
		return
	}

	ui.PrintSection("Running Post-Init Script")

	// Write script to a temp file for execution
	tmpFile, err := os.CreateTemp("", "specledger-init-*.sh")
	if err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to create temp file for init script: %v", err))
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(scriptContent); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to write init script: %v", err))
		tmpFile.Close()
		return
	}
	tmpFile.Close()

	// Make the temp file executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to make init script executable: %v", err))
		return
	}

	// Execute the script with environment variables
	// #nosec G204 -- tmpFile.Name() is from os.CreateTemp, safe path
	cmd := exec.Command(tmpFile.Name())
	cmd.Dir = projectPath

	// Set environment variables from specledger.yaml for script use
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SPECLEDGER_PROJECT_ROOT=%s", projectPath),
		fmt.Sprintf("SPECLEDGER_PROJECT_NAME=%s", projectMetadata.Project.Name),
		fmt.Sprintf("SPECLEDGER_PROJECT_SHORT_CODE=%s", projectMetadata.Project.ShortCode),
		fmt.Sprintf("SPECLEDGER_PROJECT_VERSION=%s", projectMetadata.Project.Version),
		fmt.Sprintf("SPECLEDGER_PLAYBOOK_NAME=%s", projectMetadata.Playbook.Name),
		fmt.Sprintf("SPECLEDGER_PLAYBOOK_VERSION=%s", projectMetadata.Playbook.Version),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Post-init script failure is not fatal - log warning and continue
		ui.PrintWarning(fmt.Sprintf("Post-init script had issues: %v", err))
		fmt.Println(string(output))
	} else {
		fmt.Printf("%s Post-init completed\n", ui.Checkmark())
	}
}
