package metadata

import (
	"errors"
	"regexp"
	"time"
)

// ProjectMetadata represents specledger.yaml
type ProjectMetadata struct {
	Version      string          `yaml:"version"`
	Project      ProjectInfo     `yaml:"project"`
	Playbook     PlaybookInfo    `yaml:"playbook"`
	TaskTracker  TaskTrackerInfo `yaml:"task_tracker,omitempty"`
	Dependencies []Dependency    `yaml:"dependencies,omitempty"`
}

// ProjectInfo contains project identification
type ProjectInfo struct {
	Name      string    `yaml:"name"`
	ShortCode string    `yaml:"short_code"`
	Created   time.Time `yaml:"created"`
	Modified  time.Time `yaml:"modified"`
	Version   string    `yaml:"version"`
}

// TaskTrackerInfo records task/issue tracker choice
type TaskTrackerInfo struct {
	Choice    TaskTrackerChoice `yaml:"choice"`
	EnabledAt *time.Time        `yaml:"enabled_at,omitempty"`
}

// PlaybookInfo records the playbook applied to this project
type PlaybookInfo struct {
	Name      string     `yaml:"name"`                 // Name of the playbook (e.g., "specledger")
	Version   string     `yaml:"version"`              // Version of the playbook that was applied
	AppliedAt *time.Time `yaml:"applied_at,omitempty"` // When the playbook was applied
	Structure []string   `yaml:"structure,omitempty"`  // Folder structure created by the playbook
}

// FrameworkInfo records SDD framework choice (deprecated, kept for migration compatibility)
type FrameworkInfo struct {
	Choice      FrameworkChoice `yaml:"choice"`
	InstalledAt *time.Time      `yaml:"installed_at,omitempty"`
}

// FrameworkChoice is an enum (deprecated)
type FrameworkChoice string

const (
	FrameworkSpecKit  FrameworkChoice = "speckit"
	FrameworkOpenSpec FrameworkChoice = "openspec"
	FrameworkBoth     FrameworkChoice = "both"
	FrameworkNone     FrameworkChoice = "none"
)

// TaskTrackerChoice is an enum for task/issue tracking systems
type TaskTrackerChoice string

const (
	TaskTrackerBeads TaskTrackerChoice = "beads"
	TaskTrackerNone  TaskTrackerChoice = "none"
)

// Dependency represents an external spec dependency
type Dependency struct {
	URL            string          `yaml:"url"`
	Branch         string          `yaml:"branch,omitempty"`
	Path           string          `yaml:"path,omitempty"`
	Alias          string          `yaml:"alias,omitempty"`
	ResolvedCommit string          `yaml:"resolved_commit,omitempty"`
	Framework      FrameworkChoice `yaml:"framework,omitempty"`   // speckit, openspec, both, none
	ImportPath     string          `yaml:"import_path,omitempty"` // @alias/spec format for AI imports
}

// ToolStatus represents runtime tool detection (not persisted)
type ToolStatus struct {
	Name      string
	Installed bool
	Version   string
	Path      string
	Category  ToolCategory
}

// ToolCategory is an enum
type ToolCategory string

const (
	ToolCategoryCore      ToolCategory = "core"
	ToolCategoryFramework ToolCategory = "framework"
)

// Validation functions

// ValidateProjectName validates project name format
func ValidateProjectName(name string) error {
	if len(name) == 0 {
		return errors.New("project name cannot be empty")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(name) {
		return errors.New("project name must contain only alphanumeric characters and hyphens")
	}
	return nil
}

// ValidateShortCode validates short code format
func ValidateShortCode(code string) error {
	if len(code) < 2 || len(code) > 10 {
		return errors.New("short code must be 2-10 characters")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(code) {
		return errors.New("short code must contain only alphanumeric characters")
	}
	return nil
}

// ValidateGitURL validates git URL format (SSH, HTTPS, or local path)
func ValidateGitURL(url string) error {
	sshPattern := `^git@[^:]+:[^/]+/.+\.git$|^git@[^:]+:[^/]+/[^/]+$`
	httpsPattern := `^https://[^/]+/[^/]+/.+$`
	localPathPattern := `^/|^./|^../`

	if regexp.MustCompile(sshPattern).MatchString(url) {
		return nil
	}
	if regexp.MustCompile(httpsPattern).MatchString(url) {
		return nil
	}
	if regexp.MustCompile(localPathPattern).MatchString(url) {
		return nil
	}

	return errors.New("url must be valid git SSH, HTTPS URL, or local file path")
}

// ValidateCommitSHA validates commit SHA format (40-character hex)
func ValidateCommitSHA(sha string) error {
	if len(sha) != 40 {
		return errors.New("commit SHA must be 40 characters")
	}
	if !regexp.MustCompile(`^[a-f0-9]+$`).MatchString(sha) {
		return errors.New("commit SHA must contain only hexadecimal characters")
	}
	return nil
}

// Validate validates the entire ProjectMetadata structure
func (m *ProjectMetadata) Validate() error {
	if m.Version != "1.0.0" {
		return errors.New("metadata version must be 1.0.0")
	}

	if err := ValidateProjectName(m.Project.Name); err != nil {
		return err
	}

	if err := ValidateShortCode(m.Project.ShortCode); err != nil {
		return err
	}

	if m.Project.Modified.Before(m.Project.Created) {
		return errors.New("modified timestamp must be after created timestamp")
	}

	// Validate playbook name is present
	if m.Playbook.Name == "" {
		return errors.New("playbook name is required")
	}

	for i, dep := range m.Dependencies {
		if err := ValidateGitURL(dep.URL); err != nil {
			return errors.New("dependency " + string(rune(i)) + ": " + err.Error())
		}
		if dep.ResolvedCommit != "" {
			if err := ValidateCommitSHA(dep.ResolvedCommit); err != nil {
				return errors.New("dependency " + string(rune(i)) + ": " + err.Error())
			}
		}
	}

	return nil
}
