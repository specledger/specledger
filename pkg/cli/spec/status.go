package spec

import (
	"fmt"
	"os"
	"strings"
)

const statusPrefix = "**Status**:"

// ReadStatus reads the Status field value from a spec.md file.
// Returns the status string (e.g., "Draft", "Approved") or an error.
func ReadStatus(specDir string) (string, error) {
	specFile := GetSpecFile(specDir)
	data, err := os.ReadFile(specFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("spec.md not found: %s", specFile)
		}
		return "", fmt.Errorf("failed to read spec.md: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, statusPrefix) {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, statusPrefix))
			if value == "" {
				return "", fmt.Errorf("status field is empty in %s", specFile)
			}
			return value, nil
		}
	}

	return "", fmt.Errorf("status field not found in %s", specFile)
}

// WriteStatus replaces the Status field value in a spec.md file.
func WriteStatus(specDir string, newStatus string) error {
	specFile := GetSpecFile(specDir)
	data, err := os.ReadFile(specFile)
	if err != nil {
		return fmt.Errorf("failed to read spec.md: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, statusPrefix) {
			lines[i] = statusPrefix + " " + newStatus
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("status field not found in %s", specFile)
	}

	return os.WriteFile(specFile, []byte(strings.Join(lines, "\n")), 0644)
}
