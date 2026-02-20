// Package templates provides template comparison and update utilities.
package templates

import (
	"os"
	"path/filepath"

	"github.com/specledger/specledger/pkg/cli/metadata"
	"github.com/specledger/specledger/pkg/embedded"
)

// TemplateStatus represents the state of project templates relative to current CLI.
type TemplateStatus struct {
	ProjectTemplateVersion string   `json:"project_template_version"` // Version stored in specledger.yaml
	CLIVersion             string   `json:"cli_version"`              // Current CLI version
	UpdateAvailable        bool     `json:"update_available"`         // true if versions differ
	CustomizedFiles        []string `json:"customized_files"`         // Files that differ from embedded originals
	TotalFiles             int      `json:"total_files"`              // Total template files in project
	NeedsUpdate            bool     `json:"needs_update"`             // true if UpdateAvailable and in project
	InProject              bool     `json:"in_project"`               // true if running in a SpecLedger project
}

// CheckTemplateStatus checks the template status for a project.
// Returns TemplateStatus with information about template version and customization.
func CheckTemplateStatus(projectDir, cliVersion string) (*TemplateStatus, error) {
	status := &TemplateStatus{
		CLIVersion:      cliVersion,
		CustomizedFiles: []string{},
		InProject:       false,
	}

	// Check if we're in a SpecLedger project
	if !metadata.HasYAMLMetadata(projectDir) {
		return status, nil
	}

	status.InProject = true

	// Load project metadata
	meta, err := metadata.LoadFromProject(projectDir)
	if err != nil {
		return nil, err
	}

	status.ProjectTemplateVersion = meta.TemplateVersion

	// Determine if update is needed
	// If template_version is empty, assume update is needed
	if meta.TemplateVersion == "" {
		status.UpdateAvailable = true
		status.NeedsUpdate = true
	} else {
		status.UpdateAvailable = meta.TemplateVersion != cliVersion
		status.NeedsUpdate = status.UpdateAvailable
	}

	// Count template files and find customized ones
	claudeDir := filepath.Join(projectDir, ".claude")
	if _, err := os.Stat(claudeDir); err == nil {
		_ = filepath.Walk(claudeDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			status.TotalFiles++

			// Check if file is customized
			relPath, err := filepath.Rel(claudeDir, path)
			if err != nil {
				return nil
			}

			isCustom, _ := IsFileCustomized(path, relPath, embedded.SkillsFS)
			if isCustom {
				status.CustomizedFiles = append(status.CustomizedFiles, relPath)
			}

			return nil
		})
	}

	return status, nil
}
