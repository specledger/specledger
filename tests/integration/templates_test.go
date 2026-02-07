package integration

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTemplatesIntegration tests the template copying functionality.
func TestTemplatesIntegration(t *testing.T) {
	// This is a placeholder for integration tests
	// Full integration tests will be added in user story implementation

	t.Run("template source initialization", func(t *testing.T) {
		// Test that NewEmbeddedSource works
		// This will be implemented in T020
	})
}

// createTestDir creates a temporary directory for testing.
func createTestDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// countFiles counts files in a directory recursively.
func countFiles(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}
