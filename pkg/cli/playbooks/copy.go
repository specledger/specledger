package playbooks

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CopyPlaybooks copies a playbook to the destination directory from the embedded filesystem.
// It applies the glob patterns from the playbook to filter which files to copy.
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

	// If no patterns specified, copy all files
	patterns := playbook.Patterns
	if len(patterns) == 0 {
		patterns = []string{"*"}
	}

	// Walk through the embedded source directory
	err := WalkPlaybooks(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, CopyError{
				Path:      path,
				Err:       err,
				IsWarning: false,
			})
			return nil // Continue walking
		}

		// Skip directories themselves (we'll create them as needed)
		if d.IsDir() {
			return nil
		}

		// Skip files not in our playbook path
		if !strings.HasPrefix(path, srcPath+"/") && path != srcPath {
			return nil
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			result.Errors = append(result.Errors, CopyError{
				Path:      path,
				Err:       err,
				IsWarning: true,
			})
			return nil
		}

		// Skip if relPath starts with ".." (file is outside playbook path)
		if strings.HasPrefix(relPath, "..") {
			return nil
		}

		// Skip init.sh - it's executed during init but not copied to target project
		if filepath.Base(path) == "init.sh" {
			return nil
		}

		// Check if file matches any pattern
		if !matchesPattern(relPath, patterns) {
			return nil
		}

		// Determine destination path
		destPath := filepath.Join(destDir, relPath)

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
			result.Errors = append(result.Errors, CopyError{
				Path:      destPath,
				Err:       fmt.Errorf("failed to create directory: %w", err),
				IsWarning: false,
			})
			return nil
		}

		// Copy file from embedded FS
		if !opts.DryRun {
			if err := copyEmbeddedFile(path, destPath); err != nil {
				result.Errors = append(result.Errors, CopyError{
					Path:      path,
					Err:       err,
					IsWarning: false,
				})
				return nil
			}
		}

		result.FilesCopied++
		if opts.Verbose {
			fmt.Printf("Copied: %s -> %s\n", relPath, destPath)
		}

		return nil
	})

	if err != nil {
		result.Errors = append(result.Errors, CopyError{
			Path:      srcPath,
			Err:       err,
			IsWarning: false,
		})
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// matchesPattern checks if a path matches any of the given patterns.
func matchesPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// Simple glob matching
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
		// Check for directory patterns (e.g., "specledger/**")
		if strings.Contains(pattern, "**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
		// Check for recursive patterns (e.g., ".claude/**")
		if strings.HasPrefix(pattern, ".") && strings.Contains(path, "/") {
			prefix := strings.Split(pattern, "/")[0]
			if strings.HasPrefix(path, prefix+"/") || path == prefix {
				return true
			}
		}
	}
	return false // Default: exclude files that don't match any pattern
}

// copyEmbeddedFile copies a single file from embedded FS to dest.
func copyEmbeddedFile(src, dest string) error {
	// Read from embedded filesystem
	srcFile, err := ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read embedded file: %w", err)
	}

	// Write to destination
	return os.WriteFile(dest, srcFile, 0644)
}
