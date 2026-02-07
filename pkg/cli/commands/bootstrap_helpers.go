package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"specledger/pkg/cli/metadata"
	"specledger/pkg/cli/templates"
	"specledger/pkg/cli/ui"
)

// updateMiseToml updates mise.toml to enable the selected SDD framework
func updateMiseToml(projectPath string, framework metadata.FrameworkChoice) error {
	miseTomlPath := filepath.Join(projectPath, "mise.toml")

	// Read current mise.toml
	content, err := os.ReadFile(miseTomlPath)
	if err != nil {
		return fmt.Errorf("failed to read mise.toml: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var updatedLines []string

	for _, line := range lines {
		// Check if this line is a commented framework tool
		trimmed := strings.TrimSpace(line)

		// Handle Spec Kit
		if strings.HasPrefix(trimmed, "# \"pipx:git+https://github.com/github/spec-kit.git\"") {
			if framework == metadata.FrameworkSpecKit || framework == metadata.FrameworkBoth {
				// Uncomment Spec Kit
				updatedLines = append(updatedLines, strings.TrimPrefix(line, "# "))
			} else {
				// Keep commented
				updatedLines = append(updatedLines, line)
			}
			continue
		}

		// Handle OpenSpec
		if strings.HasPrefix(trimmed, "# \"npm:@fission-ai/openspec\"") {
			if framework == metadata.FrameworkOpenSpec || framework == metadata.FrameworkBoth {
				// Uncomment OpenSpec
				updatedLines = append(updatedLines, strings.TrimPrefix(line, "# "))
			} else {
				// Keep commented
				updatedLines = append(updatedLines, line)
			}
			continue
		}

		// Keep all other lines as-is
		updatedLines = append(updatedLines, line)
	}

	// Write updated content
	updatedContent := strings.Join(updatedLines, "\n")
	return os.WriteFile(miseTomlPath, []byte(updatedContent), 0644)
}

// initializeFramework initializes the chosen SDD framework with appropriate flags
func initializeFramework(projectPath string, framework metadata.FrameworkChoice) error {
	if framework == metadata.FrameworkNone {
		return nil
	}

	fmt.Println()
	ui.PrintSection("Initializing SDD Framework")

	// Install and initialize Spec Kit if chosen
	if framework == metadata.FrameworkSpecKit || framework == metadata.FrameworkBoth {
		// First try to install the tool via mise
		if err := installMiseTool(projectPath, "pipx:git+https://github.com/github/spec-kit.git"); err != nil {
			// mise install might fail if tool already installed elsewhere, continue anyway
			fmt.Printf("Note: %s\n", ui.Dim(err.Error()))
		}

		if err := runSpecifyInit(projectPath); err != nil {
			// Log warning but don't fail - user can initialize manually
			ui.PrintWarning(fmt.Sprintf("Spec Kit initialization failed: %v", err))
			ui.PrintWarning("You can initialize manually with: specify init --here --ai claude --force --script sh --no-git")
		} else {
			fmt.Printf("%s Spec Kit initialized\n", ui.Checkmark())
		}
	}

	// Install and initialize OpenSpec if chosen (and not both, where we prioritize Spec Kit)
	if framework == metadata.FrameworkOpenSpec {
		// First try to install the tool via mise
		if err := installMiseTool(projectPath, "npm:@fission-ai/openspec"); err != nil {
			// mise install might fail if tool already installed elsewhere, continue anyway
			fmt.Printf("Note: %s\n", ui.Dim(err.Error()))
		}

		if err := runOpenSpecInit(projectPath); err != nil {
			// Log warning but don't fail - user can initialize manually
			ui.PrintWarning(fmt.Sprintf("OpenSpec initialization failed: %v", err))
			ui.PrintWarning("You can initialize manually with: openspec init --force --tools claude")
		} else {
			fmt.Printf("%s OpenSpec initialized\n", ui.Checkmark())
		}
	}

	// For "both" framework, only initialize Spec Kit
	// User can manually initialize OpenSpec if needed

	fmt.Println()
	return nil
}

// runSpecifyInit runs "specify init --here --ai claude --force --script sh --no-git"
func runSpecifyInit(projectPath string) error {
	cmd := exec.Command("specify", "init", "--here", "--ai", "claude", "--force", "--script", "sh", "--no-git")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("specify init failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// runOpenSpecInit runs "openspec init --force --tools claude"
func runOpenSpecInit(projectPath string) error {
	cmd := exec.Command("openspec", "init", "--force", "--tools", "claude")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("openspec init failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// installMiseTool installs a tool using mise
func installMiseTool(projectPath, tool string) error {
	cmd := exec.Command("mise", "install", tool)
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mise install failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// applyEmbeddedTemplates copies embedded templates to the project directory.
func applyEmbeddedTemplates(projectPath string, framework metadata.FrameworkChoice) error {
	if framework == metadata.FrameworkNone {
		return nil
	}

	frameworkStr := string(framework)
	ui.PrintSection("Copying Templates")
	fmt.Printf("Applying %s templates...\n", ui.Bold(frameworkStr))

	if err := templates.ApplyToProject(projectPath, frameworkStr); err != nil {
		// Templates are helpful but not critical - log warning and continue
		ui.PrintWarning(fmt.Sprintf("Template copying failed: %v", err))
		ui.PrintWarning("Project will be created without templates")
		return nil
	}

	fmt.Printf("%s Templates applied\n", ui.Checkmark())
	return nil
}
