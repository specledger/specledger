package templates

import (
	"crypto/sha256"
	"embed"
	"os"
	"path/filepath"
)

// IsFileCustomized checks if a project file differs from the embedded original.
func IsFileCustomized(projectPath, relPath string, embeddedFS embed.FS) (bool, error) {
	projectContent, err := os.ReadFile(projectPath)
	if err != nil {
		return false, err
	}

	embeddedPath := filepath.Join("skills", relPath)
	embeddedContent, err := embeddedFS.ReadFile(embeddedPath)
	if err != nil {
		return true, nil
	}

	projectHash := sha256.Sum256(projectContent)
	embeddedHash := sha256.Sum256(embeddedContent)

	return projectHash != embeddedHash, nil
}
