package playbooks

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/specledger/specledger/internal/agent"
)

// CopyPlaybooks copies a playbook to the destination directory from the embedded filesystem.
// It copies files based on:
// 1. Structure items (files/directories copied to project root)
// 2. Commands (copied to .agents/commands/ or .claude/commands/ depending on AgentTargetDir)
// 3. Skills (copied to .agents/skills/ or .claude/skills/ depending on AgentTargetDir)
func CopyPlaybooks(srcDir, destDir string, playbook Playbook, opts CopyOptions) (*CopyResult, error) {
	startTime := time.Now()
	result := &CopyResult{}

	// Validate source directory exists in embedded FS
	// Use path.Join (forward slashes) for embedded FS paths
	srcPath := path.Join(srcDir, playbook.Path)
	if !Exists(srcPath) {
		return result, fmt.Errorf("playbook path not found in embedded filesystem: %s", playbook.Path)
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return result, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Build protected files map from playbook
	protectedMap := make(map[string]bool)
	for _, p := range playbook.Protected {
		protectedMap[p] = true
	}

	// Build mergeable files map from playbook
	mergeableMap := make(map[string]bool)
	for _, m := range playbook.Mergeable {
		mergeableMap[m] = true
	}

	// Determine target directory for commands and skills
	agentTargetDir := opts.AgentTargetDir
	if agentTargetDir == "" {
		agentTargetDir = ".claude"
	}

	// 1. Copy structure items (files/directories to project root)
	for _, structureItem := range playbook.Structure {
		// path.Join for embedded FS source, filepath.Join for local destination
		itemSrcPath := path.Join(srcPath, structureItem)
		itemDestPath := filepath.Join(destDir, structureItem)

		if err := copyStructureItem(itemSrcPath, itemDestPath, structureItem, opts, result, protectedMap, mergeableMap); err != nil {
			result.Errors = append(result.Errors, CopyError{
				Path:      structureItem,
				Err:       err,
				IsWarning: true,
			})
		}
	}

	// 2. Copy commands to {agentTargetDir}/commands/
	for _, cmd := range playbook.Commands {
		// path.Join for embedded FS source, filepath.Join for local destination
		srcFilePath := path.Join(srcPath, cmd.Path)
		destFilePath := filepath.Join(destDir, agentTargetDir, "commands", filepath.Base(cmd.Path))

		if err := copySingleFile(srcFilePath, destFilePath, opts, result, protectedMap); err != nil {
			result.Errors = append(result.Errors, CopyError{
				Path:      cmd.Path,
				Err:       err,
				IsWarning: false,
			})
		}
	}

	// 3. Copy skills to {agentTargetDir}/skills/
	for _, skill := range playbook.Skills {
		// path.Join for embedded FS source, filepath.Join for local destination
		srcFilePath := path.Join(srcPath, skill.Path)
		destFilePath := filepath.Join(destDir, agentTargetDir, skill.Path)

		if err := copySingleFile(srcFilePath, destFilePath, opts, result, protectedMap); err != nil {
			result.Errors = append(result.Errors, CopyError{
				Path:      skill.Path,
				Err:       err,
				IsWarning: false,
			})
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// copyStructureItem copies a structure item (file or directory) from embedded FS to destination.
func copyStructureItem(srcPath, destPath, structureItem string, opts CopyOptions, result *CopyResult, protectedFiles, mergeableFiles map[string]bool) error {
	// Check if source exists in embedded FS
	if !Exists(srcPath) {
		return fmt.Errorf("structure item not found: %s", srcPath)
	}

	// Check if it's a directory or file by trying to read it
	content, err := ReadFile(srcPath)
	if err != nil {
		// It's a directory - walk and copy all files
		return copyDirectory(srcPath, destPath, structureItem, opts, result, protectedFiles)
	}

	// Mergeable files use sentinel-based merge (bypasses protected and overwrite logic)
	if mergeableFiles[structureItem] {
		return mergeFile(srcPath, destPath, content, opts, result)
	}

	// It's a file - check if protected
	if protectedFiles[structureItem] {
		if opts.Verbose {
			fmt.Printf("Skipped protected file: %s\n", structureItem)
		}
		result.FilesSkipped++
		return nil
	}

	// Copy directly
	return copySingleFile(srcPath, destPath, opts, result, protectedFiles)
}

// copyDirectory recursively copies a directory from embedded FS to destination.
func copyDirectory(srcPath, destPath, structureItem string, opts CopyOptions, result *CopyResult, protectedFiles map[string]bool) error {
	// Create destination directory
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Walk through the embedded source directory
	return WalkPlaybooks(func(walkPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip if not under our source path
		if !strings.HasPrefix(walkPath, srcPath+"/") && walkPath != srcPath {
			return nil
		}

		// Skip directories (created as needed)
		if d.IsDir() {
			return nil
		}

		// Get relative path from source directory (embedded FS uses forward slashes)
		relPath := strings.TrimPrefix(walkPath, srcPath+"/")
		if relPath == "" || relPath == walkPath {
			return nil
		}

		// Construct the full project-relative path for protected file checking
		// e.g., structureItem=".specledger/" + relPath="memory/constitution.md"
		fullPath := path.Join(strings.TrimSuffix(structureItem, "/"), relPath)

		// Skip protected files that shouldn't be overwritten
		if protectedFiles[fullPath] || protectedFiles[path.Base(relPath)] {
			if opts.Verbose {
				fmt.Printf("Skipped protected file: %s\n", fullPath)
			}
			result.FilesSkipped++
			return nil
		}

		// Determine destination path (local filesystem uses filepath)
		fileDestPath := filepath.Join(destPath, filepath.FromSlash(relPath))

		return copySingleFile(walkPath, fileDestPath, opts, result, protectedFiles)
	})
}

// copySingleFile copies a single file from embedded FS to destination.
func copySingleFile(srcPath, destPath string, opts CopyOptions, result *CopyResult, protectedFiles map[string]bool) error {
	// Skip protected files that shouldn't be overwritten
	filename := path.Base(srcPath)
	if protectedFiles[filename] {
		if opts.Verbose {
			fmt.Printf("Skipped protected file: %s\n", srcPath)
		}
		result.FilesSkipped++
		return nil
	}

	content, err := ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded file: %w", err)
	}
	return copySingleFileFromContent(srcPath, destPath, content, opts, result)
}

// copySingleFileFromContent writes content to destination with appropriate permissions.
func copySingleFileFromContent(srcPath, destPath string, content []byte, opts CopyOptions, result *CopyResult) error {
	// Check if file already exists
	if _, err := os.Stat(destPath); err == nil {
		if !opts.Overwrite {
			result.FilesSkipped++
			if opts.Verbose {
				fmt.Printf("Skipped existing file: %s\n", destPath)
			}
			return nil
		}
	}

	// Create destination directory structure
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Determine permissions based on file type
	var perms fs.FileMode
	if IsExecutableFile(filepath.Base(destPath), content) {
		perms = 0755 // Executable: rwxr-xr-x
	} else {
		perms = 0644 // Regular: rw-r--r-
	}

	// Write to destination
	if !opts.DryRun {
		if err := os.WriteFile(destPath, content, perms); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	result.FilesCopied++
	if opts.Verbose {
		fmt.Printf("Copied: %s -> %s\n", srcPath, destPath)
	}

	return nil
}

// mergeFile merges embedded template content into an existing file using sentinel markers.
// If the destination file doesn't exist, it creates it with the sentinel block.
func mergeFile(srcPath, destPath string, templateContent []byte, opts CopyOptions, result *CopyResult) error {
	// Read existing file from disk (empty string if not exists)
	existing := ""
	if data, err := os.ReadFile(destPath); err == nil {
		existing = string(data)
	}

	// Merge using sentinel markers
	managed := strings.TrimRight(string(templateContent), "\n")
	merged := MergeSentinelSection(existing, managed)

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if !opts.DryRun {
		if err := os.WriteFile(destPath, []byte(merged), 0644); err != nil {
			return fmt.Errorf("failed to write merged file: %w", err)
		}
	}

	result.FilesMerged++
	if opts.Verbose {
		fmt.Printf("Merged: %s -> %s\n", srcPath, destPath)
	}

	return nil
}

// IsExecutableFile determines if a file should have execute permissions.
// Returns true if the file has a .sh extension or starts with a shebang (#!).
func IsExecutableFile(filename string, content []byte) bool {
	if strings.HasSuffix(filename, ".sh") {
		return true
	}

	if len(content) > 2 && content[0] == '#' && content[1] == '!' {
		return true
	}

	return false
}

func CreateAgentSharedDir(projectDir string, force bool) error {
	agentDir := filepath.Join(projectDir, ".agents")

	// Check if .agents/ already exists
	if _, err := os.Stat(agentDir); err == nil {
		if !force {
			return fmt.Errorf(".agents/ directory already exists. Use --force to proceed")
		}
		// With --force, we just proceed without deleting existing content
		// This preserves any custom commands/skills the user has added
	}

	// Ensure commands and skills directories exist (MkdirAll is idempotent)
	commandsDir := filepath.Join(agentDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .agents/commands: %w", err)
	}

	skillsDir := filepath.Join(agentDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .agents/skills: %w", err)
	}

	// Migrate from .claude/commands if it exists (only copy files that don't already exist)
	claudeCommandsDir := filepath.Join(projectDir, ".claude", "commands")
	if _, err := os.Stat(claudeCommandsDir); err == nil {
		entries, err := os.ReadDir(claudeCommandsDir)
		if err != nil {
			return fmt.Errorf("failed to read .claude/commands: %w", err)
		}
		for _, entry := range entries {
			src := filepath.Join(claudeCommandsDir, entry.Name())
			dst := filepath.Join(commandsDir, entry.Name())
			// Only copy if destination doesn't exist (preserve existing customizations)
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				if err := copyFileOrDir(src, dst); err != nil {
					return fmt.Errorf("failed to migrate %s: %w", entry.Name(), err)
				}
			}
		}
	}

	return nil
}

func LinkAgentToShared(projectDir string, agentNames []string, force bool) error {
	// Ensure .agents directories exist
	sharedCommandsDir := filepath.Join(projectDir, ".agents", "commands")
	sharedSkillsDir := filepath.Join(projectDir, ".agents", "skills")
	if err := os.MkdirAll(sharedCommandsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .agents/commands: %w", err)
	}
	if err := os.MkdirAll(sharedSkillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .agents/skills: %w", err)
	}

	// Build set of selected agents for quick lookup
	selectedSet := make(map[string]bool)
	for _, name := range agentNames {
		selectedSet[strings.ToLower(strings.TrimSpace(name))] = true
	}

	// Link ALL agents that have existing config directories or are selected
	// This ensures we don't leave orphaned directories with wrong symlinks
	for _, ag := range agent.All() {
		// Skip agents without a config dir or with special dirs (like .github)
		if ag.ConfigDir == "" || ag.ConfigDir == ".github" {
			continue
		}

		agentDir := filepath.Join(projectDir, ag.ConfigDir)
		commandsLink := filepath.Join(agentDir, "commands")
		skillsLink := filepath.Join(agentDir, "skills")

		// Check if this agent should be linked:
		// 1. It's in the selected list, OR
		// 2. It has an existing config directory (commands or skills)
		isSelected := selectedSet[strings.ToLower(ag.Name)] || selectedSet[strings.ToLower(ag.Command)]
		hasCommandsDir := dirExists(commandsLink)
		hasSkillsDir := dirExists(skillsLink)

		if !isSelected && !hasCommandsDir && !hasSkillsDir {
			continue // Skip agents that aren't selected and don't have existing dirs
		}

		if err := os.MkdirAll(agentDir, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", agentDir, err)
		}

		// Handle commands directory
		if err := migrateAndLink(commandsLink, sharedCommandsDir, force); err != nil {
			return fmt.Errorf("failed to link %s/commands: %w", ag.Name, err)
		}

		// Handle skills directory
		if err := migrateAndLink(skillsLink, sharedSkillsDir, force); err != nil {
			return fmt.Errorf("failed to link %s/skills: %w", ag.Name, err)
		}
	}

	return nil
}

// dirExists checks if a path exists and is a directory (or symlink to directory)
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// migrateAndLink handles migrating contents from an existing directory to shared,
// then creates a symlink to the shared directory.
func migrateAndLink(linkPath, sharedDir string, force bool) error {
	// Compute relative path from linkPath's parent to sharedDir for portable symlinks
	linkParent := filepath.Dir(linkPath)
	relSharedDir, err := filepath.Rel(linkParent, sharedDir)
	if err != nil {
		// Fall back to absolute path if relative computation fails
		relSharedDir = sharedDir
	}

	// Check if linkPath already exists
	info, err := os.Lstat(linkPath)
	if err == nil {
		// Already exists
		if info.Mode()&os.ModeSymlink != 0 {
			// It's already a symlink, check if it points to the right place
			target, _ := os.Readlink(linkPath)
			// Compare against relative path since that's what we use
			if target == relSharedDir {
				return nil // Already correctly linked
			}
		}

		if !force {
			return nil // Skip if not forcing
		}

		// If it's a directory (not a symlink), migrate contents first
		if info.IsDir() {
			// Migrate contents to shared directory
			entries, err := os.ReadDir(linkPath)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", linkPath, err)
			}
			for _, entry := range entries {
				src := filepath.Join(linkPath, entry.Name())
				dst := filepath.Join(sharedDir, entry.Name())
				// Only copy if destination doesn't exist
				if _, err := os.Stat(dst); os.IsNotExist(err) {
					if err := copyFileOrDir(src, dst); err != nil {
						// Log warning but continue
						fmt.Printf("Warning: failed to migrate %s: %v\n", src, err)
					}
				}
			}
		}

		// Remove existing file/directory/symlink
		if err := os.RemoveAll(linkPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", linkPath, err)
		}
	}

	// Create symlink using relative path for portability
	return agent.SymlinkOrCopy(relSharedDir, linkPath)
}

func copyFileOrDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDirRecursive(src, dst)
	}
	return copyFileContents(src, dst)
}

func copyDirRecursive(src, dst string) error {
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

func copyFileContents(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
}

// CleanupAgentSymlinks removes existing agent symlinks/directories for all known agents.
// This is called when --force is used to reset the agent configuration.
func CleanupAgentSymlinks(projectPath string) error {
	for _, ag := range agent.All() {
		// Skip agents without a config dir or with special dirs (like .github)
		if ag.ConfigDir == "" || ag.ConfigDir == ".github" {
			continue
		}

		agentDir := filepath.Join(projectPath, ag.ConfigDir)
		commandsDir := filepath.Join(agentDir, "commands")
		skillsDir := filepath.Join(agentDir, "skills")

		// Remove commands symlink/directory if it exists
		if _, err := os.Lstat(commandsDir); err == nil {
			if err := os.RemoveAll(commandsDir); err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", commandsDir, err)
			}
		}

		// Remove skills symlink/directory if it exists
		if _, err := os.Lstat(skillsDir); err == nil {
			if err := os.RemoveAll(skillsDir); err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", skillsDir, err)
			}
		}
	}

	return nil
}
