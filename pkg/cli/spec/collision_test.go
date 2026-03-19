package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetNextFeatureNum_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	// No specledger directory at all — should return "001"
	got, err := GetNextFeatureNum(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "001" {
		t.Errorf("got %q, want %q", got, "001")
	}
}

func TestGetNextFeatureNum_WithExistingDirs(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "specledger")
	if err := os.Mkdir(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"001-first", "002-second", "010-tenth"} {
		if err := os.Mkdir(filepath.Join(specDir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}

	got, err := GetNextFeatureNum(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "011" {
		t.Errorf("got %q, want %q", got, "011")
	}
}

func TestGetNextFeatureNum_NonNumericDirsIgnored(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "specledger")
	if err := os.Mkdir(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"003-feature", ".specledger", "notafeature"} {
		if err := os.Mkdir(filepath.Join(specDir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}

	got, err := GetNextFeatureNum(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "004" {
		t.Errorf("got %q, want %q", got, "004")
	}
}

func TestGetNextFeatureNum_FilesIgnored(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "specledger")
	if err := os.Mkdir(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file (not dir) with numeric prefix — should be ignored
	if err := os.WriteFile(filepath.Join(specDir, "005-file.md"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	// Create a real feature dir
	if err := os.Mkdir(filepath.Join(specDir, "002-real"), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := GetNextFeatureNum(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "003" {
		t.Errorf("got %q, want %q", got, "003")
	}
}

func TestGetNextFeatureNum_FormatsWithLeadingZeros(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "specledger")
	if err := os.Mkdir(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(filepath.Join(specDir, "099-feature"), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := GetNextFeatureNum(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "100" {
		t.Errorf("got %q, want %q", got, "100")
	}
}
