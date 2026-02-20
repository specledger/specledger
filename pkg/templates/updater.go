package templates

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/specledger/specledger/pkg/cli/metadata"
	"github.com/specledger/specledger/pkg/embedded"
)

// TemplateUpdateResult represents the result of a template update operation.
type TemplateUpdateResult struct {
	Updated    []string `json:"updated"`     // Files that were updated
	Overwritten []string `json:"overwritten"` // Files that existed and were overwritten
	Errors     []error  `json:"errors"`      // Any errors encountered
	NewVersion string   `json:"new_version"` // New template_version written to YAML
	Success    bool     `json:"success"`     // true if no fatal errors
}

// UpdateTemplates updates project templates from embedded files.
// All embedded templates are copied, overwriting any existing files.
func UpdateTemplates(projectDir, cliVersion string) (*TemplateUpdateResult, error) {
	result := &TemplateUpdateResult{
		Updated:     []string{},
		Overwritten: []string{},
		Errors:      []error{},
		NewVersion:  cliVersion,
		Success:     true,
	}

	claudeDir := filepath.Join(projectDir, ".claude")

	// Ensure .claude directory exists
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Walk the embedded skills FS and copy files
	err := fs.WalkDir(embedded.SkillsFS, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Get relative path within skills directory
		relPath := strings.TrimPrefix(path, "skills/")

		// Target path in project
		targetPath := filepath.Join(claudeDir, relPath)

		// Check if file exists (for tracking overwritten files)
		fileExists := false
		if _, err := os.Stat(targetPath); err == nil {
			fileExists = true
		}

		// Ensure parent directory exists
		parentDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to create directory %s: %w", parentDir, err))
			return nil
		}

		// Read embedded file
		content, err := embedded.SkillsFS.ReadFile(path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to read embedded %s: %w", relPath, err))
			return nil
		}

		// Determine file permissions
		// Scripts (files ending in .sh, or in commands/ directory) get executable permissions
		perm := os.FileMode(0644)
		if strings.HasSuffix(relPath, ".sh") || strings.HasPrefix(relPath, "commands/") {
			perm = 0755
		}

		// Write to project (overwrites if exists)
		if err := os.WriteFile(targetPath, content, perm); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to write %s: %w", relPath, err))
			return nil
		}

		if fileExists {
			result.Overwritten = append(result.Overwritten, relPath)
		} else {
			result.Updated = append(result.Updated, relPath)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking embedded files: %w", err)
	}

	// Update template_version in specledger.yaml
	if err := updateTemplateVersion(projectDir, cliVersion); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to update template_version: %w", err))
		result.Success = false
	}

	// Mark as not successful if there were any errors
	if len(result.Errors) > 0 {
		result.Success = false
	}

	return result, nil
}

// updateTemplateVersion updates the template_version field in specledger.yaml.
func updateTemplateVersion(projectDir, cliVersion string) error {
	// Load current metadata
	meta, err := metadata.LoadFromProject(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Update template version
	meta.TemplateVersion = cliVersion

	// Save metadata
	if err := metadata.SaveToProject(meta, projectDir); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}
