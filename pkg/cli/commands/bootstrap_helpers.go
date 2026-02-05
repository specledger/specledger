package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"specledger/pkg/cli/metadata"
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
		if strings.HasPrefix(trimmed, "# \"npm:@openspec/cli\"") {
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
