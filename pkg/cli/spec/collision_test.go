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
	os.MkdirAll(filepath.Join(specDir, "001-first-feature"), 0755)
	os.MkdirAll(filepath.Join(specDir, "005-fifth-feature"), 0755)
	os.MkdirAll(filepath.Join(specDir, "003-third-feature"), 0755)

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
	os.MkdirAll(filepath.Join(specDir, "010-real-feature"), 0755)
	os.MkdirAll(filepath.Join(specDir, "migrated"), 0755)
	os.MkdirAll(filepath.Join(specDir, "not-a-feature"), 0755)
	// Create a file (not directory)
	os.WriteFile(filepath.Join(specDir, "specledger.yaml"), []byte(""), 0644)

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
	os.MkdirAll(filepath.Join(specDir, "001-first"), 0755)
	os.MkdirAll(filepath.Join(specDir, "002-second"), 0755)
	os.MkdirAll(filepath.Join(specDir, "003-third"), 0755)

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
	os.MkdirAll(filepath.Join(specDir, "001-first"), 0755)

	err := checkLocalFeatures(tmpDir, "002")
	if err != nil {
		t.Errorf("expected no collision, got: %v", err)
	}
}

func TestCheckFeatureCollision_HasCollision(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "specledger")
	os.MkdirAll(filepath.Join(specDir, "001-first"), 0755)

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
