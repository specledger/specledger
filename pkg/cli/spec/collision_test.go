package spec

import (
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	_, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func TestGetNextFeatureNum_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	num, err := GetNextFeatureNum(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if num != "001" {
		t.Errorf("expected 001, got %s", num)
	}
}

func TestGetNextFeatureNum_WithExistingFeatures(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "specledger")
	mustMkdirAll(t, filepath.Join(specDir, "001-first-feature"))
	mustMkdirAll(t, filepath.Join(specDir, "005-fifth-feature"))
	mustMkdirAll(t, filepath.Join(specDir, "003-third-feature"))

	num, err := GetNextFeatureNum(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if num != "006" {
		t.Errorf("expected 006, got %s", num)
	}
}

func TestGetNextFeatureNum_SkipsNonFeatureDirs(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "specledger")
	mustMkdirAll(t, filepath.Join(specDir, "010-real-feature"))
	mustMkdirAll(t, filepath.Join(specDir, "migrated"))
	mustMkdirAll(t, filepath.Join(specDir, "not-a-feature"))
	// Create a file (not directory)
	mustWriteFile(t, filepath.Join(specDir, "specledger.yaml"), []byte(""))

	num, err := GetNextFeatureNum(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if num != "011" {
		t.Errorf("expected 011, got %s", num)
	}
}

func TestGetNextAvailableNum_NoCollision(t *testing.T) {
	tmpDir := t.TempDir()
	// Init a git repo so branch checks don't fall through to parent repo
	initGitRepo(t, tmpDir)

	num, err := GetNextAvailableNum(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if num != "001" {
		t.Errorf("expected 001, got %s", num)
	}
}

func TestGetNextAvailableNum_SkipsCollisions(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)
	specDir := filepath.Join(tmpDir, "specledger")
	// Create features 001 through 003
	mustMkdirAll(t, filepath.Join(specDir, "001-first"))
	mustMkdirAll(t, filepath.Join(specDir, "002-second"))
	mustMkdirAll(t, filepath.Join(specDir, "003-third"))

	num, err := GetNextAvailableNum(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if num != "004" {
		t.Errorf("expected 004, got %s", num)
	}
}

func TestCheckFeatureCollision_NoCollision(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "specledger")
	mustMkdirAll(t, filepath.Join(specDir, "001-first"))

	err := checkLocalFeatures(tmpDir, "002")
	if err != nil {
		t.Errorf("expected no collision, got: %v", err)
	}
}

func TestCheckFeatureCollision_HasCollision(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "specledger")
	mustMkdirAll(t, filepath.Join(specDir, "001-first"))

	err := checkLocalFeatures(tmpDir, "001")
	if err == nil {
		t.Error("expected collision error, got nil")
	}
}

func TestParseFeatureNum(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"600-my-feature", "600"},
		{"001-first", "001"},
		{"main", ""},
		{"feature-no-num", "feature"},
	}

	for _, tt := range tests {
		got := ParseFeatureNum(tt.branch)
		if got != tt.want {
			t.Errorf("ParseFeatureNum(%q) = %q, want %q", tt.branch, got, tt.want)
		}
	}
}
