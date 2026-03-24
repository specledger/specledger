package metadata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRootFrom(t *testing.T) {
	t.Run("finds root at given directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create specledger/specledger.yaml
		metaDir := filepath.Join(tmpDir, "specledger")
		if err := os.MkdirAll(metaDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(metaDir, "specledger.yaml"), []byte("version: 1.0.0\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		root, err := FindProjectRootFrom(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != tmpDir {
			t.Errorf("expected %s, got %s", tmpDir, root)
		}
	})

	t.Run("finds root from subdirectory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create specledger/specledger.yaml at root
		metaDir := filepath.Join(tmpDir, "specledger")
		if err := os.MkdirAll(metaDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(metaDir, "specledger.yaml"), []byte("version: 1.0.0\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Create a deep subdirectory
		subDir := filepath.Join(tmpDir, "src", "pkg", "deep")
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatal(err)
		}

		root, err := FindProjectRootFrom(subDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != tmpDir {
			t.Errorf("expected %s, got %s", tmpDir, root)
		}
	})

	t.Run("returns error when no project root found", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := FindProjectRootFrom(tmpDir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
