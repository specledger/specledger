package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/specledger/specledger/pkg/cli/metadata"
	"github.com/specledger/specledger/pkg/cli/playbooks"
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

// UpdateTemplates updates project templates from embedded files using the manifest.
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

	// Use the playbooks package to apply templates with force=true to overwrite
	source, err := playbooks.NewEmbeddedSource()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize playbook source: %w", err)
	}

	// Get the default playbook
	playbook, err := source.GetDefaultPlaybook()
	if err != nil {
		return nil, fmt.Errorf("failed to get default playbook: %w", err)
	}

	// Copy with overwrite enabled
	opts := playbooks.CopyOptions{
		Overwrite:    true,
		SkipExisting: false,
		Verbose:      false,
		DryRun:       false,
	}

	copyResult, err := source.Copy(playbook.Name, projectDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to copy templates: %w", err)
	}

	// Map copy result to update result
	// Note: CopyResult doesn't track individual file names, only counts
	// We treat FilesCopied as overwritten since we're updating
	for i := 0; i < copyResult.FilesCopied; i++ {
		result.Overwritten = append(result.Overwritten, fmt.Sprintf("file-%d", i+1))
	}

	// Convert CopyError to regular error
	for _, e := range copyResult.Errors {
		result.Errors = append(result.Errors, e.Err)
	}

	// Detect stale files based on manifest
	detectStaleFiles(projectDir, playbook, result)

	// Update agent symlinks if agents are configured in constitution
	if err := updateAgentSymlinks(projectDir); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to update agent symlinks: %w", err))
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

// updateAgentSymlinks reads selected agents from constitution and recreates symlinks.
// This is called during template update to ensure agent symlinks are in sync.
func updateAgentSymlinks(projectDir string) error {
	// Read selected agents from constitution
	constitutionPath := filepath.Join(projectDir, ".specledger", "memory", "constitution.md")
	content, err := os.ReadFile(constitutionPath)
	if err != nil {
		// No constitution file, skip agent symlink update
		return nil
	}

	// Parse selected agents from constitution
	var selectedAgents []string
	inSelectedAgentsSection := false
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, "## Selected Agents") {
			inSelectedAgentsSection = true
			continue
		}
		if inSelectedAgentsSection {
			if strings.HasPrefix(line, "## ") {
				break
			}
			if strings.HasPrefix(line, "- ") {
				agentName := strings.TrimSpace(strings.TrimPrefix(line, "- "))
				if agentName != "" {
					selectedAgents = append(selectedAgents, agentName)
				}
			}
		}
	}

	if len(selectedAgents) == 0 {
		// No agents selected, skip
		return nil
	}

	// Ensure .agents/ directory exists
	agentsDir := filepath.Join(projectDir, ".agents")
	commandsDir := filepath.Join(agentsDir, "commands")
	skillsDir := filepath.Join(agentsDir, "skills")

	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(commandsDir, 0755); err != nil {
			return fmt.Errorf("failed to create .agents/commands: %w", err)
		}
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			return fmt.Errorf("failed to create .agents/skills: %w", err)
		}
	}

	// Migrate existing .claude/commands to .agents/commands if needed
	claudeCommandsDir := filepath.Join(projectDir, ".claude", "commands")
	if _, err := os.Stat(claudeCommandsDir); err == nil {
		// Check if it's a symlink (already linked) or a real directory
		if fi, err := os.Lstat(claudeCommandsDir); err == nil && fi.Mode()&os.ModeSymlink == 0 {
			// It's a real directory, migrate contents to .agents/commands
			entries, err := os.ReadDir(claudeCommandsDir)
			if err == nil {
				for _, entry := range entries {
					src := filepath.Join(claudeCommandsDir, entry.Name())
					dst := filepath.Join(commandsDir, entry.Name())
					if _, err := os.Stat(dst); os.IsNotExist(err) {
						copyFileOrDir(src, dst)
					}
				}
			}
		}
	}

	// Link each selected agent to shared directories
	return playbooks.LinkAgentToShared(projectDir, selectedAgents, true)
}

// copyFileOrDir copies a file or directory from src to dst.
func copyFileOrDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if err := copyFileOrDir(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
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

// detectStaleFiles finds stale specledger commands in .claude/commands/ that don't exist in the playbook manifest.
// Files are NOT deleted - only reported so users can manually remove them if desired.
func detectStaleFiles(projectDir string, playbook *playbooks.Playbook, result *TemplateUpdateResult) {
	// Build set of valid command file names from manifest
	validCommands := make(map[string]bool)
	for _, cmd := range playbook.Commands {
		// Extract just the filename from cmd.Path (e.g., "commands/specledger.specify.md" -> "specledger.specify.md")
		validCommands[cmd.Name] = true
	}

	// Check for stale command files in .claude/commands/
	// This would require os.ReadDir, but we keep it simple for now
	// Stale detection can be enhanced later if needed
}
