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

func TestGenerateFeatureHash(t *testing.T) {
	hash, err := GenerateFeatureHash()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 6 {
		t.Errorf("expected 6-char hash, got %q (len %d)", hash, len(hash))
	}
	// Verify it's valid hex
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("invalid hex character %c in hash %q", c, hash)
		}
	}
}

func TestGenerateFeatureHash_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		hash, err := GenerateFeatureHash()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if seen[hash] {
			t.Errorf("duplicate hash generated: %s", hash)
		}
		seen[hash] = true
	}
}

func TestGenerateUniqueFeatureHash(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	hash, err := GenerateUniqueFeatureHash(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 6 {
		t.Errorf("expected 6-char hash, got %q", hash)
	}
}

func TestCheckFeatureCollision_NoCollision(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "specledger")
	mustMkdirAll(t, filepath.Join(specDir, "a1b2c3-first"))

	err := checkLocalFeatures(tmpDir, "d4e5f6")
	if err != nil {
		t.Errorf("expected no collision, got: %v", err)
	}
}

func TestCheckFeatureCollision_HasCollision(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "specledger")
	mustMkdirAll(t, filepath.Join(specDir, "a1b2c3-first"))

	err := checkLocalFeatures(tmpDir, "a1b2c3")
	if err == nil {
		t.Error("expected collision error, got nil")
	}
}

func TestCheckFeatureCollision_LegacyNumeric(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "specledger")
	mustMkdirAll(t, filepath.Join(specDir, "001-first"))

	err := checkLocalFeatures(tmpDir, "001")
	if err == nil {
		t.Error("expected collision error for legacy numeric, got nil")
	}
}

func TestCheckFeatureCollision_NoSpecledgerDir(t *testing.T) {
	tmpDir := t.TempDir()

	err := checkLocalFeatures(tmpDir, "a1b2c3")
	if err != nil {
		t.Errorf("expected no error for missing specledger dir, got: %v", err)
	}
}

func TestParseFeatureID(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"a3f2b1-my-feature", "a3f2b1"},
		{"604-auto-spec-numbers", "604"},
		{"001-first", "001"},
		{"main", ""},
		{"feature", ""},
	}

	for _, tt := range tests {
		got := ParseFeatureID(tt.name)
		if got != tt.want {
			t.Errorf("ParseFeatureID(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestParseFeatureNum_BackwardCompat(t *testing.T) {
	// ParseFeatureNum should work the same as ParseFeatureID
	got := ParseFeatureNum("604-auto-spec")
	if got != "604" {
		t.Errorf("ParseFeatureNum(\"604-auto-spec\") = %q, want \"604\"", got)
	}
}
