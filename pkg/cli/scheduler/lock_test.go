package scheduler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAcquireLock(t *testing.T) {
	root := t.TempDir()
	feature := "127-test-feature"

	if err := AcquireLock(root, feature); err != nil {
		t.Fatalf("AcquireLock() error: %v", err)
	}

	// Verify lock file exists
	lockPath := LockFilePath(root)
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Fatal("lock file not created")
	}

	// Verify lock contents
	lock, err := CheckLock(root)
	if err != nil {
		t.Fatalf("CheckLock() error: %v", err)
	}
	if lock == nil {
		t.Fatal("CheckLock() returned nil")
	}
	if lock.PID != os.Getpid() {
		t.Errorf("PID = %d, want %d", lock.PID, os.Getpid())
	}
	if lock.Feature != feature {
		t.Errorf("Feature = %q, want %q", lock.Feature, feature)
	}
	if lock.StartedAt == "" {
		t.Error("StartedAt is empty")
	}
}

func TestAcquireLock_AlreadyHeld(t *testing.T) {
	root := t.TempDir()
	feature := "127-test-feature"

	if err := AcquireLock(root, feature); err != nil {
		t.Fatalf("first AcquireLock() error: %v", err)
	}

	err := AcquireLock(root, feature)
	if err == nil {
		t.Error("expected error when lock already held")
	}
}

func TestCheckLock_NoFile(t *testing.T) {
	root := t.TempDir()
	lock, err := CheckLock(root)
	if err != nil {
		t.Fatalf("CheckLock() error: %v", err)
	}
	if lock != nil {
		t.Error("expected nil lock when no file exists")
	}
}

func TestReleaseLock(t *testing.T) {
	root := t.TempDir()
	feature := "127-test-feature"

	if err := AcquireLock(root, feature); err != nil {
		t.Fatalf("AcquireLock() error: %v", err)
	}

	if err := ReleaseLock(root); err != nil {
		t.Fatalf("ReleaseLock() error: %v", err)
	}

	lockPath := LockFilePath(root)
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("lock file still exists after release")
	}
}

func TestReleaseLock_NoFile(t *testing.T) {
	root := t.TempDir()
	if err := ReleaseLock(root); err != nil {
		t.Fatalf("ReleaseLock() should not error when no file: %v", err)
	}
}

func TestIsLockHeld(t *testing.T) {
	root := t.TempDir()

	if IsLockHeld(root) {
		t.Error("IsLockHeld() should be false with no lock")
	}

	if err := AcquireLock(root, "test"); err != nil {
		t.Fatal(err)
	}

	if !IsLockHeld(root) {
		t.Error("IsLockHeld() should be true after acquire")
	}
}

func TestLockFilePath(t *testing.T) {
	got := LockFilePath("/project")
	want := filepath.Join("/project", ".specledger", "exec.lock")
	if got != want {
		t.Errorf("LockFilePath() = %q, want %q", got, want)
	}
}
