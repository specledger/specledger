package templates

import (
	"time"
)

// TemplateSource represents a source of templates (embedded or remote).
type TemplateSource interface {
	// List returns all available templates from this source.
	List() ([]Template, error)

	// Copy copies the specified template to the destination directory.
	// The name parameter must match a template name from List().
	// Options control the copy behavior (overwrite, skip, etc.).
	Copy(name string, destDir string, opts CopyOptions) (*CopyResult, error)

	// Exists checks if a template with the given name exists in this source.
	Exists(name string) bool
}

// Template represents a single template/playbook that can be applied to a project.
type Template struct {
	// Name is the unique identifier for this template (e.g., "speckit", "openspec")
	Name string

	// Description explains what this template provides to users
	Description string

	// Framework is the SDD framework type: "speckit", "openspec", "both", or "none"
	Framework string

	// Version is the template version (semantic versioning recommended)
	Version string

	// Path is the relative path within the templates folder
	Path string

	// Patterns are glob patterns for files to include (optional, defaults to all files)
	// If empty, all files in Path are included.
	Patterns []string
}

// TemplateManifest represents the manifest file that lists available templates.
type TemplateManifest struct {
	// Version is the manifest format version
	Version string

	// Templates is the list of available templates
	Templates []Template
}

// CopyOptions controls the behavior of template copying operations.
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

// CopyResult contains the results of a template copy operation.
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
