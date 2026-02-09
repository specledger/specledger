package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// findRepoRoot finds the repository root directory by looking for go.mod
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up directories looking for go.mod
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// buildSLBinary builds the sl binary and returns its path
func buildSLBinary(t testing.TB, tempDir string) string {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("Failed to find repo root: %v", err)
	}

	slBinary := filepath.Join(tempDir, "sl")
	buildCmd := exec.Command("go", "build", "-o", slBinary, filepath.Join(repoRoot, "cmd", "sl", "main.go"))
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build sl binary: %v\nOutput: %s", err, string(output))
	}
	return slBinary
}
