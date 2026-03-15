package playbooks

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/specledger/specledger/internal/agent"
)

// CopyPlaybooks copies a playbook to the destination directory from the embedded filesystem.
// It copies files based on:
// 1. Structure items (files/directories copied to project root)
// 2. Commands (copied to .claude/commands/)
// 3. Skills (copied to .claude/skills/)
func CopyPlaybooks(srcDir, destDir string, playbook Playbook, opts CopyOptions) (*CopyResult, error) {
	startTime := time.Now()
	result := &CopyResult{}

	// Validate source directory exists in embedded FS
	srcPath := filepath.Join(srcDir, playbook.Path)
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

	// 1. Copy structure items (files/directories to project root)
	for _, structureItem := range playbook.Structure {
		itemSrcPath := filepath.Join(srcPath, structureItem)
		itemDestPath := filepath.Join(destDir, structureItem)

		if err := copyStructureItem(itemSrcPath, itemDestPath, srcPath, structureItem, opts, result, protectedMap); err != nil {
			result.Errors = append(result.Errors, CopyError{
				Path:      structureItem,
				Err:       err,
				IsWarning: true,
			})
		}
	}

	// 2. Copy commands to .claude/commands/
	for _, cmd := range playbook.Commands {
		srcFilePath := filepath.Join(srcPath, cmd.Path)
		destFilePath := filepath.Join(destDir, ".claude", "commands", filepath.Base(cmd.Path))

		if err := copySingleFile(srcFilePath, destFilePath, opts, result, protectedMap); err != nil {
			result.Errors = append(result.Errors, CopyError{
				Path:      cmd.Path,
				Err:       err,
				IsWarning: false,
			})
		}
	}

	// 3. Copy skills to .claude/skills/
	for _, skill := range playbook.Skills {
		// For skills with subdirectories like "skills/sl-audit/skill.md"
		// we preserve the skill directory structure
		srcFilePath := filepath.Join(srcPath, skill.Path)
		destFilePath := filepath.Join(destDir, ".claude", skill.Path)

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
func copyStructureItem(srcPath, destPath, playbookSrcPath, structureItem string, opts CopyOptions, result *CopyResult, protectedFiles map[string]bool) error {
	// Check if source exists in embedded FS
	if !Exists(srcPath) {
		return fmt.Errorf("structure item not found: %s", srcPath)
	}

	// Check if it's a directory or file by trying to read it
	_, err := ReadFile(srcPath)
	if err != nil {
		// It's a directory - walk and copy all files
		return copyDirectory(srcPath, destPath, playbookSrcPath, structureItem, opts, result, protectedFiles)
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
func copyDirectory(srcPath, destPath, playbookSrcPath, structureItem string, opts CopyOptions, result *CopyResult, protectedFiles map[string]bool) error {
	// Create destination directory
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Walk through the embedded source directory
	return WalkPlaybooks(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip if not under our source path
		if !strings.HasPrefix(path, srcPath+"/") && path != srcPath {
			return nil
		}

		// Skip directories (created as needed)
		if d.IsDir() {
			return nil
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		// Construct the full project-relative path for protected file checking
		// e.g., structureItem=".specledger/" + relPath="memory/constitution.md" -> ".specledger/memory/constitution.md"
		fullPath := filepath.Join(structureItem, relPath)
		// Clean up path separators (handles trailing slashes)
		fullPath = filepath.ToSlash(fullPath)
		if strings.HasSuffix(structureItem, "/") {
			fullPath = strings.TrimSuffix(structureItem, "/") + "/" + relPath
		}

		// Skip protected files that shouldn't be overwritten
		if protectedFiles[fullPath] || protectedFiles[filepath.Base(relPath)] {
			if opts.Verbose {
				fmt.Printf("Skipped protected file: %s\n", fullPath)
			}
			result.FilesSkipped++
			return nil
		}

		// Determine destination path
		fileDestPath := filepath.Join(destPath, relPath)

		return copySingleFile(path, fileDestPath, opts, result, protectedFiles)
	})
}

// copySingleFile copies a single file from embedded FS to destination.
func copySingleFile(srcPath, destPath string, opts CopyOptions, result *CopyResult, protectedFiles map[string]bool) error {
	// Skip protected files that shouldn't be overwritten
	filename := filepath.Base(srcPath)
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
	agentDir := filepath.Join(projectDir, ".agent")

	if _, err := os.Stat(agentDir); err == nil {
		if !force {
			return fmt.Errorf(".agent/ directory already exists. Use --force to overwrite")
		}
		if err := os.RemoveAll(agentDir); err != nil {
			return fmt.Errorf("failed to remove existing .agent directory: %w", err)
		}
	}

	commandsDir := filepath.Join(agentDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .agent/commands: %w", err)
	}

	skillsDir := filepath.Join(agentDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .agent/skills: %w", err)
	}

	claudeCommandsDir := filepath.Join(projectDir, ".claude", "commands")
	if _, err := os.Stat(claudeCommandsDir); err == nil {
		entries, err := os.ReadDir(claudeCommandsDir)
		if err != nil {
			return fmt.Errorf("failed to read .claude/commands: %w", err)
		}
		for _, entry := range entries {
			src := filepath.Join(claudeCommandsDir, entry.Name())
			dst := filepath.Join(commandsDir, entry.Name())
			if err := copyFileOrDir(src, dst); err != nil {
				return fmt.Errorf("failed to migrate %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

func LinkAgentToShared(projectDir string, agentNames []string, force bool) error {
	for _, agentName := range agentNames {
		ag, found := agent.Lookup(agentName)
		if !found {
			continue
		}

		if ag.ConfigDir == "" || ag.ConfigDir == ".github" {
			continue
		}

		agentDir := filepath.Join(projectDir, ag.ConfigDir)
		sharedCommandsDir := filepath.Join(projectDir, ".agent", "commands")
		sharedSkillsDir := filepath.Join(projectDir, ".agent", "skills")

		if err := os.MkdirAll(agentDir, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", agentDir, err)
		}

		commandsLink := filepath.Join(agentDir, "commands")
		if _, err := os.Stat(commandsLink); err == nil {
			if !force {
				continue
			}
			os.Remove(commandsLink)
		}
		if err := agent.SymlinkOrCopy(sharedCommandsDir, commandsLink); err != nil {
			return fmt.Errorf("failed to link %s/commands: %w", ag.Name, err)
		}

		skillsLink := filepath.Join(agentDir, "skills")
		if _, err := os.Stat(skillsLink); err == nil {
			if !force {
				continue
			}
			os.Remove(skillsLink)
		}
		if err := agent.SymlinkOrCopy(sharedSkillsDir, skillsLink); err != nil {
			return fmt.Errorf("failed to link %s/skills: %w", ag.Name, err)
		}
	}

	return nil
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
