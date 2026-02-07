package templates

import (
	"fmt"
	"strings"
)

// ApplyToProject applies templates to a project based on the framework.
// This is the main entry point for template copying during project creation.
func ApplyToProject(projectPath, framework string) error {
	source, err := NewEmbeddedSource()
	if err != nil {
		return fmt.Errorf("failed to initialize template source: %w", err)
	}

	// Validate templates exist
	if err := source.ValidateTemplates(); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Get template by framework
	template, err := source.GetTemplateByFramework(framework)
	if err != nil {
		return fmt.Errorf("failed to find template for framework %s: %w", framework, err)
	}

	// Copy templates
	opts := CopyOptions{
		SkipExisting: true,
		Verbose:      false,
		DryRun:       false,
		Framework:    framework,
	}

	result, err := source.Copy(template.Name, projectPath, opts)
	if err != nil {
		return fmt.Errorf("failed to copy templates: %w", err)
	}

	// Report results
	if result.FilesCopied > 0 {
		fmt.Printf("Copied %d template files to %s\n", result.FilesCopied, projectPath)
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

	return nil
}

// ListTemplates returns all available templates formatted for display.
func ListTemplates() ([]Template, error) {
	source, err := NewEmbeddedSource()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize template source: %w", err)
	}

	return source.List()
}

// FormatTable formats templates as a table for display.
func FormatTable(templates []Template) string {
	var sb strings.Builder

	sb.WriteString("Available Templates:\n")
	sb.WriteString("\n")

	for _, tmpl := range templates {
		sb.WriteString(fmt.Sprintf("  %-12s Framework: %-10s Version: %s\n",
			tmpl.Name+" ", tmpl.Framework, tmpl.Version))
		sb.WriteString(fmt.Sprintf("    %s\n", tmpl.Description))
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatJSON formats templates as JSON for display.
func FormatJSON(templates []Template) string {
	var sb strings.Builder

	sb.WriteString("[\n")
	for i, tmpl := range templates {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf("  {\n"))
		sb.WriteString(fmt.Sprintf("    \"name\": \"%s\",\n", tmpl.Name))
		sb.WriteString(fmt.Sprintf("    \"description\": \"%s\",\n", tmpl.Description))
		sb.WriteString(fmt.Sprintf("    \"framework\": \"%s\",\n", tmpl.Framework))
		sb.WriteString(fmt.Sprintf("    \"version\": \"%s\"\n", tmpl.Version))
		sb.WriteString(fmt.Sprintf("  }"))
	}
	sb.WriteString("\n]\n")

	return sb.String()
}
