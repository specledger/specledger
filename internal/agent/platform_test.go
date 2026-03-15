package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsWindows(t *testing.T) {
	result := IsWindows()
	expected := runtime.GOOS == "windows"
	if result != expected {
		t.Errorf("IsWindows() = %v, want %v (runtime.GOOS = %s)", result, expected, runtime.GOOS)
	}
}

func TestSupportsSymlinks(t *testing.T) {
	result := SupportsSymlinks()
	expected := runtime.GOOS != "windows"
	if result != expected {
		t.Errorf("SupportsSymlinks() = %v, want %v (runtime.GOOS = %s)", result, expected, runtime.GOOS)
	}
}

func TestSymlinkOrCopy_File(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "source.txt")
	srcContent := []byte("test content")
	// #nosec G306 -- Test files need to be readable
	if err := os.WriteFile(srcFile, srcContent, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dstFile := filepath.Join(tmpDir, "dest.txt")
	if err := SymlinkOrCopy(srcFile, dstFile); err != nil {
		t.Fatalf("SymlinkOrCopy() failed: %v", err)
	}

	if SupportsSymlinks() {
		linkTarget, err := os.Readlink(dstFile)
		if err != nil {
			t.Errorf("Expected symlink, but Readlink failed: %v", err)
		} else if linkTarget != srcFile {
			t.Errorf("Symlink target = %q, want %q", linkTarget, srcFile)
		}
	} else {
		dstContent, err := os.ReadFile(dstFile)
		if err != nil {
			t.Fatalf("Failed to read copied file: %v", err)
		}
		if string(dstContent) != string(srcContent) {
			t.Errorf("Copied content = %q, want %q", dstContent, srcContent)
		}
	}
}

func TestSymlinkOrCopy_Directory(t *testing.T) {
	if IsWindows() {
		t.Skip("Skipping directory copy test on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	// #nosec G306 -- Test files need to be readable
	if err := os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	// #nosec G306 -- Test files need to be readable
	if err := os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	dstDir := filepath.Join(tmpDir, "dst")
	if err := SymlinkOrCopy(srcDir, dstDir); err != nil {
		t.Fatalf("SymlinkOrCopy() failed: %v", err)
	}

	if SupportsSymlinks() {
		linkTarget, err := os.Readlink(dstDir)
		if err != nil {
			t.Errorf("Expected symlink, but Readlink failed: %v", err)
		} else if linkTarget != srcDir {
			t.Errorf("Symlink target = %q, want %q", linkTarget, srcDir)
		}
	} else {
		content1, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
		if err != nil || string(content1) != "content1" {
			t.Errorf("file1.txt not copied correctly")
		}
		content2, err := os.ReadFile(filepath.Join(dstDir, "subdir", "file2.txt"))
		if err != nil || string(content2) != "content2" {
			t.Errorf("subdir/file2.txt not copied correctly")
		}
	}
}

func TestSymlinkOrCopy_OverwritesExisting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "source.txt")
	// #nosec G306 -- Test files need to be readable
	if err := os.WriteFile(srcFile, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dstFile := filepath.Join(tmpDir, "dest.txt")
	// #nosec G306 -- Test files need to be readable
	if err := os.WriteFile(dstFile, []byte("old content"), 0644); err != nil {
		t.Fatalf("Failed to create dest file: %v", err)
	}

	if err := SymlinkOrCopy(srcFile, dstFile); err != nil {
		t.Fatalf("SymlinkOrCopy() failed: %v", err)
	}

	if SupportsSymlinks() {
		linkTarget, err := os.Readlink(dstFile)
		if err != nil || linkTarget != srcFile {
			t.Errorf("Expected symlink to source, got: %v, err: %v", linkTarget, err)
		}
	} else {
		content, err := os.ReadFile(dstFile)
		if err != nil || string(content) != "new content" {
			t.Errorf("Expected new content, got: %s, err: %v", content, err)
		}
	}
}
