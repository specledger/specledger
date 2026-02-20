package models

import (
	"fmt"
	"regexp"
	"strings"
)

// TemplateDefinition represents a project template with metadata and structure.
// Templates define predefined project structures for different use cases
// (e.g., Full-Stack, ML Image Processing, AI Chatbot).
type TemplateDefinition struct {
	// ID is the unique template identifier in kebab-case (e.g., "full-stack", "ml-image")
	ID string `yaml:"id"`

	// Name is the human-readable template name (e.g., "Full-Stack Application")
	Name string `yaml:"name"`

	// Description is a one-line description of the template's purpose
	Description string `yaml:"description"`

	// Characteristics lists key technologies used in the template (e.g., "Go", "React", "Kafka")
	Characteristics []string `yaml:"characteristics,omitempty"`

	// Path is the relative path within embedded templates/ directory
	Path string `yaml:"path"`

	// IsDefault indicates if this is the default template selection
	IsDefault bool `yaml:"is_default,omitempty"`
}

// kebabCasePattern validates kebab-case identifiers (lowercase letters, numbers, hyphens)
var kebabCasePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// Validate checks if the TemplateDefinition has valid field values.
// Returns an error if any validation rule fails.
func (t *TemplateDefinition) Validate() error {
	// Validate ID format (kebab-case)
	if !kebabCasePattern.MatchString(t.ID) {
		return fmt.Errorf("template ID must be kebab-case (lowercase letters, numbers, hyphens): got %q", t.ID)
	}

	// Validate name length (1-50 characters)
	if len(t.Name) == 0 || len(t.Name) > 50 {
		return fmt.Errorf("template name must be 1-50 characters: got %d", len(t.Name))
	}

	// Validate description length (1-200 characters)
	if len(t.Description) == 0 || len(t.Description) > 200 {
		return fmt.Errorf("template description must be 1-200 characters: got %d", len(t.Description))
	}

	// Validate characteristics (max 6 items)
	if len(t.Characteristics) > 6 {
		return fmt.Errorf("template can have max 6 characteristics: got %d", len(t.Characteristics))
	}

	// Validate path is non-empty
	if strings.TrimSpace(t.Path) == "" {
		return fmt.Errorf("template path cannot be empty")
	}

	return nil
}

// String returns a formatted string representation of the template for logging.
func (t *TemplateDefinition) String() string {
	defaultMarker := ""
	if t.IsDefault {
		defaultMarker = " (default)"
	}

	techStr := ""
	if len(t.Characteristics) > 0 {
		techStr = fmt.Sprintf(" [%s]", strings.Join(t.Characteristics, ", "))
	}

	return fmt.Sprintf("%s: %s%s%s", t.ID, t.Name, techStr, defaultMarker)
}
