package templates

import (
	"crypto/sha256"
	"embed"
	"os"
	"path/filepath"
	"strings"
)

// IsFileCustomized checks if a project file differs from the embedded original.
func IsFileCustomized(projectPath, relPath string, embeddedFS embed.FS) (bool, error) {
	// Read project file
	projectContent, err := os.ReadFile(projectPath)
	if err != nil {
		return false, err
	}

	// Read embedded file
	embeddedPath := filepath.Join("skills", relPath)
	embeddedContent, err := embeddedFS.ReadFile(embeddedPath)
	if err != nil {
		// File exists in project but not in embedded - definitely customized
		return true, nil
	}

	// Compare hashes
	projectHash := sha256.Sum256(projectContent)
	embeddedHash := sha256.Sum256(embeddedContent)

	return projectHash != embeddedHash, nil
}

// FindCustomizedFiles finds all customized files in the .claude directory.
func FindCustomizedFiles(projectDir string, embeddedFS embed.FS) ([]string, error) {
	customized := []string{}

	claudeDir := filepath.Join(projectDir, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return customized, nil
	}

	err := filepath.Walk(claudeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(claudeDir, path)
		if err != nil {
			return nil
		}

		// Normalize path for cross-platform compatibility
		relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		// Check if customized
		isCustom, err := IsFileCustomized(path, relPath, embeddedFS)
		if err != nil {
			// Skip files we can't read
			return nil
		}

		if isCustom {
			customized = append(customized, relPath)
		}

		return nil
	})

	return customized, err
}

// ComputeChecksum computes the SHA-256 checksum of a file.
func ComputeChecksum(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(content)
	return hash[:], nil
}

// ComputeEmbeddedChecksum computes the SHA-256 checksum of an embedded file.
func ComputeEmbeddedChecksum(relPath string, embeddedFS embed.FS) ([]byte, error) {
	embeddedPath := filepath.Join("skills", relPath)
	content, err := embeddedFS.ReadFile(embeddedPath)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(content)
	return hash[:], nil
}
