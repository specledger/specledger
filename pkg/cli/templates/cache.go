package templates

import (
	"fmt"
	"os"
	"path/filepath"
)

// CacheDir returns the directory for caching remote templates.
// Future implementation will use this for storing downloaded templates.
func CacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".specledger", "template-cache")

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cacheDir, nil
}

// CachedTemplatePath returns the path where a cached template would be stored.
// Future implementation will use this for storing downloaded template archives.
func CachedTemplatePath(templateName, version string) (string, error) {
	cacheDir, err := CacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, fmt.Sprintf("%s-%s.tar.gz", templateName, version)), nil
}

// ClearCache removes all cached templates.
// Future implementation for cache management.
func ClearCache() error {
	cacheDir, err := CacheDir()
	if err != nil {
		return err
	}

	// Remove the entire cache directory
	return os.RemoveAll(cacheDir)
}
