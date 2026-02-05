package metadata

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewProjectMetadata(t *testing.T) {
	tests := []struct {
		name      string
		projName  string
		shortCode string
		framework FrameworkChoice
		wantAt    bool // whether InstalledAt should be set
	}{
		{"with speckit", "test-project", "tp", FrameworkSpecKit, true},
		{"with openspec", "test-project", "tp", FrameworkOpenSpec, true},
		{"with both", "test-project", "tp", FrameworkBoth, true},
		{"with none", "test-project", "tp", FrameworkNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := NewProjectMetadata(tt.projName, tt.shortCode, tt.framework)

			if metadata.Version != MetadataVersion {
				t.Errorf("expected version %s, got %s", MetadataVersion, metadata.Version)
			}

			if metadata.Project.Name != tt.projName {
				t.Errorf("expected name %s, got %s", tt.projName, metadata.Project.Name)
			}

			if metadata.Project.ShortCode != tt.shortCode {
				t.Errorf("expected short code %s, got %s", tt.shortCode, metadata.Project.ShortCode)
			}

			if metadata.Framework.Choice != tt.framework {
				t.Errorf("expected framework %s, got %s", tt.framework, metadata.Framework.Choice)
			}

			if tt.wantAt && metadata.Framework.InstalledAt == nil {
				t.Error("expected InstalledAt to be set")
			}

			if !tt.wantAt && metadata.Framework.InstalledAt != nil {
				t.Error("expected InstalledAt to be nil")
			}

			if metadata.Project.Version != "0.1.0" {
				t.Errorf("expected project version 0.1.0, got %s", metadata.Project.Version)
			}

			if len(metadata.Dependencies) != 0 {
				t.Errorf("expected empty dependencies, got %d", len(metadata.Dependencies))
			}
		})
	}
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
		Framework: FrameworkInfo{
			Choice: FrameworkSpecKit,
		},
		Dependencies: []Dependency{
			{
				URL:    "git@github.com:org/repo.git",
				Branch: "main",
				Path:   "spec.md",
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

	if loaded.Framework.Choice != metadata.Framework.Choice {
		t.Errorf("framework choice mismatch: got %s, want %s", loaded.Framework.Choice, metadata.Framework.Choice)
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
	metadata := NewProjectMetadata("test-project", "tp", FrameworkNone)

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
		Framework: FrameworkInfo{
			Choice: FrameworkNone,
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
