package playbooks

import (
	"fmt"
	"strings"
)

// ApplyToProject applies playbooks to a project.
// This is the main entry point for playbook copying during project creation.
// If playbookName is empty, it uses the default playbook (currently "specledger").
// Returns the playbook name, version, and structure.
func ApplyToProject(projectPath, playbookName string) (string, string, []string, error) {
	source, err := NewEmbeddedSource()
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to initialize playbook source: %w", err)
	}

	// Validate playbooks exist
	if err := source.ValidatePlaybooks(); err != nil {
		return "", "", nil, fmt.Errorf("playbook validation failed: %w", err)
	}

	// Get the playbook - use default if not specified
	var playbook *Playbook
	if playbookName == "" {
		playbook, err = source.GetDefaultPlaybook()
		if err != nil {
			return "", "", nil, fmt.Errorf("failed to get default playbook: %w", err)
		}
	} else {
		playbook, err = source.getPlaybook(playbookName)
		if err != nil {
			return "", "", nil, fmt.Errorf("failed to get playbook '%s': %w", playbookName, err)
		}
	}

	// Copy playbooks
	opts := CopyOptions{
		SkipExisting: true,
		Verbose:      false,
		DryRun:       false,
	}

	result, err := source.Copy(playbook.Name, projectPath, opts)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to copy playbooks: %w", err)
	}

	// Report results
	if result.FilesCopied > 0 {
		fmt.Printf("Copied %d playbook files to %s\n", result.FilesCopied, projectPath)
	}
	if result.FilesSkipped > 0 {
		fmt.Printf("Skipped %d existing files\n", result.FilesSkipped)
	}
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			if e.IsWarning {
				fmt.Printf("Warning: %s: %v\n", e.Path, e.Err)
			} else {
				fmt.Printf("Error: %s: %v\n", e.Path, e.Err)
			}
		}
	}

	return playbook.Name, playbook.Version, playbook.Structure, nil
}

// ListPlaybooks returns all available playbooks formatted for display.
func ListPlaybooks() ([]Playbook, error) {
	source, err := NewEmbeddedSource()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize playbook source: %w", err)
	}

	return source.List()
}

// FormatTable formats playbooks as a table for display.
func FormatTable(playbooks []Playbook) string {
	var sb strings.Builder

	sb.WriteString("Available Playbooks:\n")
	sb.WriteString("\n")

	for _, pb := range playbooks {
		sb.WriteString(fmt.Sprintf("  %-12s Version: %s\n",
			pb.Name+" ", pb.Version))
		sb.WriteString(fmt.Sprintf("    %s\n", pb.Description))
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatJSON formats playbooks as JSON for display.
func FormatJSON(playbooks []Playbook) string {
	var sb strings.Builder

	sb.WriteString("[\n")
	for i, pb := range playbooks {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf("  {\n"))
		sb.WriteString(fmt.Sprintf("    \"name\": \"%s\",\n", pb.Name))
		sb.WriteString(fmt.Sprintf("    \"description\": \"%s\",\n", pb.Description))
		sb.WriteString(fmt.Sprintf("    \"version\": \"%s\"\n", pb.Version))
		sb.WriteString(fmt.Sprintf("  }"))
	}
	sb.WriteString("\n]\n")

	return sb.String()
}
