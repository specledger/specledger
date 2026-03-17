package spec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/specledger/specledger/pkg/models"
)

func TestParseManifest(t *testing.T) {
	t.Run("valid manifest", func(t *testing.T) {
		content := `# Test manifest
require https://github.com/test/repo main specledger/spec.md
require https://github.com/test/repo2 v1.0 specledger/spec2.md --alias my-dep
`
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "spec.mod")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		manifest, err := ParseManifest(path)
		if err != nil {
			t.Fatalf("ParseManifest() error: %v", err)
		}

		if len(manifest.Dependecies) != 2 {
			t.Errorf("expected 2 dependencies, got %d", len(manifest.Dependecies))
		}

		if manifest.Dependecies[0].RepositoryURL != "https://github.com/test/repo" {
			t.Errorf("expected repo URL 'https://github.com/test/repo', got %s", manifest.Dependecies[0].RepositoryURL)
		}

		if manifest.Dependecies[1].Alias != "my-dep" {
			t.Errorf("expected alias 'my-dep', got %s", manifest.Dependecies[1].Alias)
		}
	})

	t.Run("empty manifest", func(t *testing.T) {
		content := `# Empty manifest
`
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "spec.mod")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		manifest, err := ParseManifest(path)
		if err != nil {
			t.Fatalf("ParseManifest() error: %v", err)
		}

		if len(manifest.Dependecies) != 0 {
			t.Errorf("expected 0 dependencies, got %d", len(manifest.Dependecies))
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := ParseManifest("/nonexistent/spec.mod")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	t.Run("invalid dependency line", func(t *testing.T) {
		content := `require only-two-parts
`
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "spec.mod")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := ParseManifest(path)
		if err == nil {
			t.Error("expected error for invalid dependency line")
		}
	})

	t.Run("missing require keyword", func(t *testing.T) {
		content := `dependency https://github.com/test/repo main spec.md
`
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "spec.mod")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := ParseManifest(path)
		if err == nil {
			t.Error("expected error for missing require keyword")
		}
	})

	t.Run("alias without value", func(t *testing.T) {
		content := `require https://github.com/test/repo main spec.md --alias
`
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "spec.mod")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := ParseManifest(path)
		if err == nil {
			t.Error("expected error for alias without value")
		}
	})
}

func TestWriteManifest(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "spec.mod")

	manifest := &Manifest{
		Version:     ManifestVersion,
		Dependecies: []models.Dependency{},
	}

	if err := WriteManifest(path, manifest); err != nil {
		t.Fatalf("WriteManifest() error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected manifest file to be created")
	}
}

func TestExtractID(t *testing.T) {
	t.Run("with alias", func(t *testing.T) {
		deps := []models.Dependency{
			{RepositoryURL: "https://github.com/test/repo", Alias: "#my-spec"},
		}
		id := extractID(deps)
		if id != "#my-spec" {
			t.Errorf("expected id '#my-spec', got %s", id)
		}
	})

	t.Run("without alias empty deps", func(t *testing.T) {
		id := extractID([]models.Dependency{})
		if id != "root" {
			t.Errorf("expected id 'root', got %s", id)
		}
	})

	t.Run("without alias with deps", func(t *testing.T) {
		deps := []models.Dependency{
			{RepositoryURL: "https://github.com/test/repo"},
		}
		id := extractID(deps)
		if id == "" {
			t.Error("expected non-empty id")
		}
	})
}
