package playbooks

import (
	"time"

	"github.com/specledger/specledger/pkg/models"
)

// PlaybookSource represents a source of playbooks (embedded or remote).
type PlaybookSource interface {
	// List returns all available playbooks from this source.
	List() ([]Playbook, error)

	// Copy copies the specified playbook to the destination directory.
	// The name parameter must match a playbook name from List().
	// Options control the copy behavior (overwrite, skip, etc.).
	Copy(name string, destDir string, opts CopyOptions) (*CopyResult, error)

	// Exists checks if a playbook with the given name exists in this source.
	Exists(name string) bool
}

// Playbook represents a single playbook that can be applied to a project.
type Playbook struct {
	// Name is the unique identifier for this playbook (e.g., "specledger")
	Name string `yaml:"name"`

	// Description explains what this playbook provides to users
	Description string `yaml:"description"`

	// Framework is the SDD framework type (optional, for future use)
	Framework string `yaml:"framework,omitempty"`

	// Version is the playbook version (semantic versioning recommended)
	Version string `yaml:"version"`

	// Path is the relative path within the templates folder
	Path string `yaml:"path"`

	// Patterns are glob patterns for files to include (optional, defaults to all files)
	// If empty, all files in Path are included.
	Patterns []string `yaml:"patterns,omitempty"`

	// Structure describes the folder structure that this playbook creates
	// This is informational only, used to document what the playbook provides
	Structure []string `yaml:"structure,omitempty"`
}

// PlaybookManifest represents the manifest file that lists available playbooks and templates.
type PlaybookManifest struct {
	// Version is the manifest format version
	Version string

	// Playbooks is the list of available playbooks (legacy, for backward compatibility)
	Playbooks []Playbook

	// Templates is the list of available project templates (new in v1.1.0)
	Templates []models.TemplateDefinition `yaml:"templates,omitempty"`
}

// CopyOptions controls the behavior of playbook copying operations.
type CopyOptions struct {
	// DryRun if true, shows what would be copied without actually copying
	DryRun bool

	// Overwrite if true, overwrites existing files in the destination
	// Default is false (skip existing files)
	Overwrite bool

	// SkipExisting if true, skips existing files (default behavior)
	// Ignored if Overwrite is true
	SkipExisting bool

	// Verbose if true, prints each file being copied
	Verbose bool

	// Framework filters templates by framework type (optional)
	// If set, only templates matching this framework are copied
	Framework string
}

// CopyResult contains the results of a playbook copy operation.
type CopyResult struct {
	// FilesCopied is the count of files successfully copied
	FilesCopied int

	// FilesSkipped is the count of files skipped (already existed)
	FilesSkipped int

	// Errors is a list of errors encountered during copying
	Errors []CopyError

	// Duration is how long the copy operation took
	Duration time.Duration
}

// CopyError represents an error that occurred during copying with context.
type CopyError struct {
	// Path is the file path where the error occurred
	Path string

	// Err is the underlying error
	Err error

	// IsWarning indicates if this is a non-fatal warning
	IsWarning bool
}

// Error implements the error interface.
func (e *CopyError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the underlying error for errors.Is/As.
func (e *CopyError) Unwrap() error {
	return e.Err
}
