package deps

import (
	"fmt"
	"os"
	"path/filepath"
)

// CacheDir returns the global cache directory for SpecLedger dependencies.
// Defaults to ~/.specledger/cache/, but can be overridden via SPECLEDGER_CACHE_DIR env var.
func CacheDir() (string, error) {
	if customPath := os.Getenv("SPECLEDGER_CACHE_DIR"); customPath != "" {
		return customPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".specledger", "cache"), nil
}

// CachePathForDependency returns the cache path for a specific dependency.
// Uses the alias if available, otherwise generates a directory name from the URL.
func CachePathForDependency(alias, url string) (string, error) {
	cacheDir, err := CacheDir()
	if err != nil {
		return "", err
	}

	dirName := alias
	if dirName == "" {
		dirName = generateDirName(url)
	}

	return filepath.Join(cacheDir, dirName), nil
}

// generateDirName generates a directory name from a Git URL.
func generateDirName(url string) string {
	// This is a placeholder - the actual implementation in pkg/cli/commands/deps.go
	// has the full logic. This can be refactored later to share code.
	return "dep"
}
