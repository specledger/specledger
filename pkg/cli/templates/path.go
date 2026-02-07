package templates

import (
	"os"
	"path/filepath"
)

// TemplatesDir returns the path to the templates directory.
// It checks relative to the current working directory or executable location.
func TemplatesDir() (string, error) {
	// First, check if templates/ exists relative to current directory
	if dir, err := os.Getwd(); err == nil {
		templatesPath := filepath.Join(dir, "templates")
		if info, err := os.Stat(templatesPath); err == nil && info.IsDir() {
			return templatesPath, nil
		}
	}

	// Fallback: check relative to executable
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	templatesPath := filepath.Join(filepath.Dir(execPath), "templates")
	if info, err := os.Stat(templatesPath); err == nil && info.IsDir() {
		return templatesPath, nil
	}

	return "", os.ErrNotExist
}
