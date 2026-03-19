package spec

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func initGitRepo(t *testing.T, dir string) *gogit.Repository {
	t.Helper()
	repo, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	// Create an initial commit so HEAD exists
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, ".gitkeep"))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err := wt.Add(".gitkeep"); err != nil {
		t.Fatal(err)
	}
	_, err = wt.Commit("init", &gogit.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@test.com", When: time.Now()},
	})
	if err != nil {
		t.Fatal(err)
	}
	return repo
}

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

func TestGetNextFeatureNum_WithGitBranches(t *testing.T) {
	dir := t.TempDir()

	// Init a git repo so openRepo succeeds and branch scanning is exercised
	initGitRepo(t, dir)

	specDir := filepath.Join(dir, "specledger")
	if err := os.Mkdir(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(specDir, "005-feature"), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := GetNextFeatureNum(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "006" {
		t.Errorf("got %q, want %q", got, "006")
	}
}

func TestGetNextFeatureNum_BranchNumHigherThanDir(t *testing.T) {
	dir := t.TempDir()

	// Init git repo and create a branch with a higher number
	repo := initGitRepo(t, dir)

	// Create a branch named 020-some-feature
	headRef, err := repo.Head()
	if err != nil {
		t.Fatal(err)
	}
	err = repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("020-some-feature"), headRef.Hash()))
	if err != nil {
		t.Fatal(err)
	}

	specDir := filepath.Join(dir, "specledger")
	if err := os.Mkdir(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(specDir, "005-feature"), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := GetNextFeatureNum(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "021" {
		t.Errorf("got %q, want %q", got, "021")
	}
}

func TestParseFeatureNum(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"001-my-feature", "001"},
		{"123-test", "123"},
		{"main", ""},
		{"", ""},
		{"no-number", "no"},
	}
	for _, tt := range tests {
		got := ParseFeatureNum(tt.input)
		if got != tt.want {
			t.Errorf("ParseFeatureNum(%q) = %q, want %q", tt.input, got, tt.want)
		}
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
