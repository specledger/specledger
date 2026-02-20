package metadata

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewProjectMetadata(t *testing.T) {
	t.Run("creates specledger playbook metadata", func(t *testing.T) {
		metadata := NewProjectMetadata("test-project", "tp", "specledger", "1.0.0", []string{".beads/", ".claude/"}, "1.0.0")

		if metadata.Version != MetadataVersion {
			t.Errorf("expected version %s, got %s", MetadataVersion, metadata.Version)
		}

		if metadata.Project.Name != "test-project" {
			t.Errorf("expected name test-project, got %s", metadata.Project.Name)
		}

		if metadata.Project.ShortCode != "tp" {
			t.Errorf("expected short code tp, got %s", metadata.Project.ShortCode)
		}

		if metadata.Playbook.Name != "specledger" {
			t.Errorf("expected playbook specledger, got %s", metadata.Playbook.Name)
		}

		if metadata.Playbook.Version != "1.0.0" {
			t.Errorf("expected playbook version 1.0.0, got %s", metadata.Playbook.Version)
		}

		if metadata.Playbook.AppliedAt == nil {
			t.Error("expected AppliedAt to be set")
		}

		if len(metadata.Playbook.Structure) != 2 {
			t.Errorf("expected 2 structure items, got %d", len(metadata.Playbook.Structure))
		}

		if metadata.Project.Version != "0.1.0" {
			t.Errorf("expected project version 0.1.0, got %s", metadata.Project.Version)
		}

		if len(metadata.Dependencies) != 0 {
			t.Errorf("expected empty dependencies, got %d", len(metadata.Dependencies))
		}
	})
}

func TestSaveAndLoad(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test metadata
	now := time.Now()
	metadata := &ProjectMetadata{
		Version: "1.0.0",
		Project: ProjectInfo{
			Name:      "test-project",
			ShortCode: "tp",
			Created:   now,
			Modified:  now,
			Version:   "0.1.0",
		},
		Playbook: PlaybookInfo{
			Name:    "specledger",
			Version: "1.0.0",
		},
		Dependencies: []Dependency{
			{
				URL:    "git@github.com:org/repo.git",
				Branch: "main",
				Alias:  "test",
			},
		},
	}

	// Save to file
	yamlPath := filepath.Join(tmpDir, "test.yaml")
	if err := Save(metadata, yamlPath); err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Fatal("YAML file was not created")
	}

	// Load from file
	loaded, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("failed to load metadata: %v", err)
	}

	// Verify loaded data matches (ignoring Modified timestamp which gets updated on save)
	if loaded.Version != metadata.Version {
		t.Errorf("version mismatch: got %s, want %s", loaded.Version, metadata.Version)
	}

	if loaded.Project.Name != metadata.Project.Name {
		t.Errorf("name mismatch: got %s, want %s", loaded.Project.Name, metadata.Project.Name)
	}

	if loaded.Project.ShortCode != metadata.Project.ShortCode {
		t.Errorf("short code mismatch: got %s, want %s", loaded.Project.ShortCode, metadata.Project.ShortCode)
	}

	if loaded.Playbook.Name != metadata.Playbook.Name {
		t.Errorf("playbook name mismatch: got %s, want %s", loaded.Playbook.Name, metadata.Playbook.Name)
	}

	if len(loaded.Dependencies) != len(metadata.Dependencies) {
		t.Fatalf("dependency count mismatch: got %d, want %d", len(loaded.Dependencies), len(metadata.Dependencies))
	}

	if loaded.Dependencies[0].URL != metadata.Dependencies[0].URL {
		t.Errorf("dependency URL mismatch: got %s, want %s", loaded.Dependencies[0].URL, metadata.Dependencies[0].URL)
	}
}

func TestSaveToProjectAndLoadFromProject(t *testing.T) {
	// Create temporary project directory
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test metadata
	metadata := NewProjectMetadata("test-project", "tp", "specledger", "1.0.0", []string{}, "1.0.0")

	// Save to project
	if err := SaveToProject(metadata, tmpDir); err != nil {
		t.Fatalf("failed to save to project: %v", err)
	}

	// Verify specledger directory was created
	specledgerDir := filepath.Join(tmpDir, "specledger")
	if _, err := os.Stat(specledgerDir); os.IsNotExist(err) {
		t.Fatal("specledger directory was not created")
	}

	// Verify YAML file exists
	yamlPath := filepath.Join(tmpDir, DefaultMetadataFile)
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Fatal("YAML file was not created in correct location")
	}

	// Load from project
	loaded, err := LoadFromProject(tmpDir)
	if err != nil {
		t.Fatalf("failed to load from project: %v", err)
	}

	if loaded.Project.Name != metadata.Project.Name {
		t.Errorf("name mismatch: got %s, want %s", loaded.Project.Name, metadata.Project.Name)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write invalid YAML
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	invalidYAML := `
version: "1.0.0"
project:
  name: "test"
  short_code: "x"  # invalid: too short
  created: "2026-02-05T10:30:00Z"
  modified: "2026-02-05T10:30:00Z"
  version: "0.1.0"
framework:
  choice: none
`
	// #nosec G306 -- test file, 0644 is appropriate
	if err := os.WriteFile(yamlPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write invalid YAML: %v", err)
	}

	// Attempt to load
	_, err = Load(yamlPath)
	if err == nil {
		t.Error("expected error when loading invalid YAML, got nil")
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/file.yaml")
	if err == nil {
		t.Error("expected error when loading nonexistent file, got nil")
	}
}

func TestSaveUpdatesModifiedTimestamp(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create metadata with past modified time
	past := time.Now().Add(-time.Hour)
	metadata := &ProjectMetadata{
		Version: "1.0.0",
		Project: ProjectInfo{
			Name:      "test",
			ShortCode: "ts",
			Created:   past,
			Modified:  past,
			Version:   "0.1.0",
		},
		Playbook: PlaybookInfo{
			Name:    "specledger",
			Version: "1.0.0",
		},
		Dependencies: []Dependency{},
	}

	yamlPath := filepath.Join(tmpDir, "test.yaml")

	// Save
	if err := Save(metadata, yamlPath); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Modified timestamp should be updated
	if !metadata.Project.Modified.After(past) {
		t.Error("expected Modified timestamp to be updated to current time")
	}

	// Load and verify
	loaded, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if !loaded.Project.Modified.After(past) {
		t.Error("expected loaded Modified timestamp to be after original")
	}
}
