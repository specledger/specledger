package spec

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/specledger/specledger/pkg/models"
)

func TestNewLockfile(t *testing.T) {
	lf := NewLockfile("1.0")
	if lf.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", lf.Version)
	}
	if len(lf.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(lf.Entries))
	}
}

func TestLockfileAddEntry(t *testing.T) {
	lf := NewLockfile("1.0")
	entry := LockfileEntry{
		RepositoryURL: "https://github.com/test/repo",
		CommitHash:    "abc123",
		ContentHash:   "def456",
		SpecPath:      "specledger/spec.md",
		Size:          1024,
		FetchedAt:     time.Now().Format(time.RFC3339),
	}

	lf.AddEntry(entry)

	if len(lf.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(lf.Entries))
	}
	if lf.TotalSize != 1024 {
		t.Errorf("expected total size 1024, got %d", lf.TotalSize)
	}
}

func TestLockfileRemoveEntry(t *testing.T) {
	lf := NewLockfile("1.0")
	lf.AddEntry(LockfileEntry{
		RepositoryURL: "https://github.com/test/repo",
		SpecPath:      "specledger/spec.md",
		Size:          1024,
	})

	removed := lf.RemoveEntry("https://github.com/test/repo", "specledger/spec.md")
	if !removed {
		t.Error("expected entry to be removed")
	}
	if len(lf.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(lf.Entries))
	}
	if lf.TotalSize != 0 {
		t.Errorf("expected total size 0, got %d", lf.TotalSize)
	}
}

func TestLockfileRemoveEntryNotFound(t *testing.T) {
	lf := NewLockfile("1.0")
	removed := lf.RemoveEntry("https://github.com/nonexistent", "spec.md")
	if removed {
		t.Error("expected entry not to be removed")
	}
}

func TestLockfileGetEntry(t *testing.T) {
	lf := NewLockfile("1.0")
	lf.AddEntry(LockfileEntry{
		RepositoryURL: "https://github.com/test/repo",
		SpecPath:      "specledger/spec.md",
		ContentHash:   "abc123",
	})

	entry, found := lf.GetEntry("https://github.com/test/repo", "specledger/spec.md")
	if !found {
		t.Fatal("expected entry to be found")
	}
	if entry.ContentHash != "abc123" {
		t.Errorf("expected content hash abc123, got %s", entry.ContentHash)
	}
}

func TestLockfileGetEntryNotFound(t *testing.T) {
	lf := NewLockfile("1.0")
	_, found := lf.GetEntry("https://github.com/nonexistent", "spec.md")
	if found {
		t.Error("expected entry not to be found")
	}
}

func TestLockfileGetRepositoryEntries(t *testing.T) {
	lf := NewLockfile("1.0")
	lf.AddEntry(LockfileEntry{
		RepositoryURL: "https://github.com/test/repo",
		SpecPath:      "specledger/spec1.md",
	})
	lf.AddEntry(LockfileEntry{
		RepositoryURL: "https://github.com/test/repo",
		SpecPath:      "specledger/spec2.md",
	})
	lf.AddEntry(LockfileEntry{
		RepositoryURL: "https://github.com/other/repo",
		SpecPath:      "specledger/spec3.md",
	})

	entries := lf.GetRepositoryEntries("https://github.com/test/repo")
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestLockfileGetContentHash(t *testing.T) {
	lf := NewLockfile("1.0")
	lf.AddEntry(LockfileEntry{
		RepositoryURL: "https://github.com/test/repo",
		SpecPath:      "specledger/spec.md",
		ContentHash:   "abc123def456",
	})

	hash, found := lf.GetContentHash("https://github.com/test/repo", "specledger/spec.md")
	if !found {
		t.Fatal("expected hash to be found")
	}
	if hash != "abc123def456" {
		t.Errorf("expected hash abc123def456, got %s", hash)
	}
}

func TestLockfileGetContentHashNotFound(t *testing.T) {
	lf := NewLockfile("1.0")
	_, found := lf.GetContentHash("https://github.com/nonexistent", "spec.md")
	if found {
		t.Error("expected hash not to be found")
	}
}

func TestCalculateSHA256FromBytes(t *testing.T) {
	data := []byte("test data")
	hash := CalculateSHA256FromBytes(data)
	if len(hash) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash))
	}
	if hash == "" {
		t.Error("hash should not be empty")
	}
}

func TestLockfileCalculateSHA256(t *testing.T) {
	lf := NewLockfile("1.0")
	hash, err := lf.CalculateSHA256()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash))
	}
}

func TestLockfileWriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "spec.sum")

	lf := NewLockfile("1.0")
	lf.AddEntry(LockfileEntry{
		RepositoryURL: "https://github.com/test/repo",
		CommitHash:    "abc123",
		ContentHash:   "def4567890123456789012345678901234567890123456789012345678901234",
		SpecPath:      "specledger/spec.md",
		Size:          1024,
		FetchedAt:     time.Now().Format(time.RFC3339),
	})

	if err := lf.Write(path); err != nil {
		t.Fatalf("failed to write lockfile: %v", err)
	}

	read, err := ReadLockfile(path)
	if err != nil {
		t.Fatalf("failed to read lockfile: %v", err)
	}

	if read.Version != lf.Version {
		t.Errorf("expected version %s, got %s", lf.Version, read.Version)
	}
	if len(read.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(read.Entries))
	}
	if read.TotalSize != 1024 {
		t.Errorf("expected total size 1024, got %d", read.TotalSize)
	}
}

func TestLockfileReadNotFound(t *testing.T) {
	_, err := ReadLockfile("/nonexistent/path/spec.sum")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLockfileVerify(t *testing.T) {
	t.Run("valid lockfile", func(t *testing.T) {
		lf := NewLockfile("1.0")
		lf.AddEntry(LockfileEntry{
			RepositoryURL: "https://github.com/test/repo",
			ContentHash:   "abc123def4567890123456789012345678901234567890123456789012345678",
			SpecPath:      "specledger/spec.md",
		})

		manifest := &Manifest{
			Dependecies: []models.Dependency{},
		}

		issues, err := lf.Verify(manifest)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues, got %v", issues)
		}
	})

	t.Run("empty content hash", func(t *testing.T) {
		lf := NewLockfile("1.0")
		lf.AddEntry(LockfileEntry{
			RepositoryURL: "https://github.com/test/repo",
			ContentHash:   "",
			SpecPath:      "specledger/spec.md",
		})

		manifest := &Manifest{}
		issues, err := lf.Verify(manifest)
		if err == nil {
			t.Error("expected error for empty content hash")
		}
		if len(issues) == 0 {
			t.Error("expected issues for empty content hash")
		}
	})

	t.Run("invalid content hash length", func(t *testing.T) {
		lf := NewLockfile("1.0")
		lf.AddEntry(LockfileEntry{
			RepositoryURL: "https://github.com/test/repo",
			ContentHash:   "tooshort",
			SpecPath:      "specledger/spec.md",
		})

		manifest := &Manifest{}
		issues, err := lf.Verify(manifest)
		if err == nil {
			t.Error("expected error for invalid hash length")
		}
		if len(issues) == 0 {
			t.Error("expected issues for invalid hash length")
		}
	})
}

func TestLockfileWriteCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nested", "dir", "spec.sum")

	lf := NewLockfile("1.0")
	if err := lf.Write(path); err != nil {
		t.Fatalf("failed to write lockfile: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected lockfile to be created")
	}
}
