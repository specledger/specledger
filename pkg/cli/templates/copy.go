package templates

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CopyTemplates copies a template to the destination directory.
// It applies the glob patterns from the template to filter which files to copy.
func CopyTemplates(srcDir, destDir string, template Template, opts CopyOptions) (*CopyResult, error) {
	startTime := time.Now()
	result := &CopyResult{}

	// Validate source directory exists
	srcPath := filepath.Join(srcDir, template.Path)
	if info, err := os.Stat(srcPath); err != nil {
		return result, fmt.Errorf("template path not found: %s: %w", template.Path, err)
	} else if !info.IsDir() {
		return result, fmt.Errorf("template path is not a directory: %s", template.Path)
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return result, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// If no patterns specified, copy all files
	patterns := template.Patterns
	if len(patterns) == 0 {
		patterns = []string{"*"}
	}

	// Walk through the source directory
	err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, CopyError{
				Path:      path,
				Err:       err,
				IsWarning: false,
			})
			return nil // Continue walking
		}

		// Skip directories themselves (we'll create them as needed)
		if info.IsDir() {
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

		// Copy file
		if !opts.DryRun {
			if err := copyFile(path, destPath); err != nil {
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
			if strings.HasPrefix(path, prefix) || filepath.HasPrefix(path, prefix) {
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
	return true // Default: include all files
}

// copyFile copies a single file from src to dest.
func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}
