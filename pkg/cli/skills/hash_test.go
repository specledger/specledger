package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeFolderHash_Deterministic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.txt", "hello")
	writeFile(t, dir, "b.txt", "world")

	hash1, err := ComputeFolderHash(dir)
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := ComputeFolderHash(dir)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 != hash2 {
		t.Errorf("non-deterministic: %s != %s", hash1, hash2)
	}
}

func TestComputeFolderHash_SkipsGitAndNodeModules(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "SKILL.md", "content")
	writeFile(t, dir, ".git/config", "git stuff")
	writeFile(t, dir, "node_modules/pkg/index.js", "module")

	hash, err := ComputeFolderHash(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Now create a dir with the same file but no .git/node_modules
	dir2 := t.TempDir()
	writeFile(t, dir2, "SKILL.md", "content")

	hash2, err := ComputeFolderHash(dir2)
	if err != nil {
		t.Fatal(err)
	}

	if hash != hash2 {
		t.Errorf("skipped dirs affected hash: %s != %s", hash, hash2)
	}
}

func TestComputeFolderHash_RenameChangesHash(t *testing.T) {
	dir1 := t.TempDir()
	writeFile(t, dir1, "a.txt", "content")

	dir2 := t.TempDir()
	writeFile(t, dir2, "b.txt", "content")

	hash1, err := ComputeFolderHash(dir1)
	if err != nil {
		t.Fatal(err)
	}
	hash2, err := ComputeFolderHash(dir2)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 == hash2 {
		t.Error("rename did not change hash")
	}
}

func TestComputeFolderHash_ContentChangeDetected(t *testing.T) {
	dir1 := t.TempDir()
	writeFile(t, dir1, "a.txt", "content-v1")

	dir2 := t.TempDir()
	writeFile(t, dir2, "a.txt", "content-v2")

	hash1, err := ComputeFolderHash(dir1)
	if err != nil {
		t.Fatal(err)
	}
	hash2, err := ComputeFolderHash(dir2)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 == hash2 {
		t.Error("content change not detected")
	}
}

func TestComputeFolderHash_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	hash, err := ComputeFolderHash(dir)
	if err != nil {
		t.Fatal(err)
	}
	// SHA-256 of empty input
	if hash == "" {
		t.Error("empty dir returned empty hash")
	}
}

func TestComputeFolderHash_NestedFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "sub/deep/file.txt", "nested")
	writeFile(t, dir, "root.txt", "top")

	hash, err := ComputeFolderHash(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(hash) != 64 { // SHA-256 hex is 64 chars
		t.Errorf("hash length = %d, want 64", len(hash))
	}
}

func writeFile(t *testing.T, base, rel, content string) {
	t.Helper()
	path := filepath.Join(base, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
