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
	Updated     []string `json:"updated"`     // Files that were updated (new)
	Overwritten []string `json:"overwritten"` // Files that existed and were overwritten
	Stale       []string `json:"stale"`       // Files detected as stale (not deleted, just reported)
	Errors      []error  `json:"errors"`      // Any errors encountered
	NewVersion  string   `json:"new_version"` // New template_version written to YAML
	Success     bool     `json:"success"`     // true if no fatal errors
}

// UpdateTemplates updates project templates from embedded files.
// All embedded templates are copied, overwriting any existing files.
// Stale files (specledger.*.md in commands/) are detected but NOT deleted to preserve custom content.
func UpdateTemplates(projectDir, cliVersion string) (*TemplateUpdateResult, error) {
	result := &TemplateUpdateResult{
		Updated:     []string{},
		Overwritten: []string{},
		Stale:       []string{},
		Errors:      []error{},
		NewVersion:  cliVersion,
		Success:     true,
	}

	claudeDir := filepath.Join(projectDir, ".claude")

	// Ensure .claude directory exists
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Track all embedded file paths for stale detection
	embeddedPaths := make(map[string]bool)

	// Walk the embedded templates FS and copy files
	err := fs.WalkDir(embedded.TemplatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Get relative path within templates/specledger directory
		// The embedded templates are in templates/specledger/, so we strip that prefix
		// to get paths like .claude/commands/specledger.specify.md
		relPath := strings.TrimPrefix(path, "templates/specledger/")
		embeddedPaths[relPath] = true

		// Target path in project (relPath already includes .claude/ prefix)
		targetPath := filepath.Join(projectDir, relPath)

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
		content, err := embedded.TemplatesFS.ReadFile(path)
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

	// Detect stale specledger commands (but don't delete - user may have custom content)
	detectStaleFiles(claudeDir, embeddedPaths, result)

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

// detectStaleFiles finds stale specledger commands in .claude/commands/ that don't exist in embedded templates.
// Files are NOT deleted - only reported so users can manually remove them if desired.
func detectStaleFiles(claudeDir string, embeddedPaths map[string]bool, result *TemplateUpdateResult) {
	commandsDir := filepath.Join(claudeDir, "commands")

	// Only check commands directory for stale specledger.*.md files
	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		return // Directory doesn't exist or can't read, skip stale detection
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Only check specledger.*.md files (owned by playbook)
		if !strings.HasPrefix(name, "specledger.") || !strings.HasSuffix(name, ".md") {
			continue
		}

		relPath := "commands/" + name
		if !embeddedPaths[relPath] {
			result.Stale = append(result.Stale, relPath)
		}
	}
}
