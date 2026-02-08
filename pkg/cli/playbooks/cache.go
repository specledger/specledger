package playbooks

import (
	"fmt"
	"os"
	"path/filepath"
)

// CacheDir returns the directory for caching remote playbooks.
// Future implementation will use this for storing downloaded playbooks.
func CacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".specledger", "playbook-cache")

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cacheDir, nil
}

// CachedPlaybookPath returns the path where a cached playbook would be stored.
// Future implementation will use this for storing downloaded playbook archives.
func CachedPlaybookPath(playbookName, version string) (string, error) {
	cacheDir, err := CacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, fmt.Sprintf("%s-%s.tar.gz", playbookName, version)), nil
}

// ClearCache removes all cached playbooks.
// Future implementation for cache management.
func ClearCache() error {
	cacheDir, err := CacheDir()
	if err != nil {
		return err
	}

	// Remove the entire cache directory
	return os.RemoveAll(cacheDir)
}
