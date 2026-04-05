package issues

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// requireGit skips the test if git is not available on PATH.
func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH, skipping")
	}
}

// gitExec runs a git command in dir, failing the test on error.
func gitExec(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(output))
	}
}

// initExecRepo creates a git repo with an initial commit using exec.
func initExecRepo(t *testing.T, dir string) {
	t.Helper()
	requireGit(t)
	gitExec(t, dir, "init", "-b", "main")
	gitExec(t, dir, "config", "user.email", "test@test.com")
	gitExec(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	gitExec(t, dir, "add", "-A")
	gitExec(t, dir, "commit", "-m", "init")
}

func TestDetectSpecContext_Worktree(t *testing.T) {
	mainDir := t.TempDir()
	initExecRepo(t, mainDir)

	// Create worktree on a spec-pattern branch
	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitExec(t, mainDir, "worktree", "add", worktreePath, "-b", "300-context-feature")

	detector := NewContextDetector(worktreePath)
	spec, err := detector.DetectSpecContext()
	if err != nil {
		t.Fatalf("DetectSpecContext in worktree: %v", err)
	}
	if spec != "300-context-feature" {
		t.Errorf("got spec %q, want %q", spec, "300-context-feature")
	}
}

func TestDetectSpecContext_Worktree_NonFeatureBranch(t *testing.T) {
	mainDir := t.TempDir()
	initExecRepo(t, mainDir)

	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitExec(t, mainDir, "worktree", "add", worktreePath, "-b", "not-a-feature")

	detector := NewContextDetector(worktreePath)
	_, err := detector.DetectSpecContext()
	if err == nil {
		t.Fatal("expected error for non-feature branch in worktree")
	}
}

func TestGetBranchName_Worktree(t *testing.T) {
	mainDir := t.TempDir()
	initExecRepo(t, mainDir)

	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitExec(t, mainDir, "worktree", "add", worktreePath, "-b", "wt-branch-name")

	detector := NewContextDetector(worktreePath)
	branch, err := detector.GetBranchName()
	if err != nil {
		t.Fatalf("GetBranchName in worktree: %v", err)
	}
	if branch != "wt-branch-name" {
		t.Errorf("got %q, want %q", branch, "wt-branch-name")
	}
}

func TestIsFeatureBranch_Worktree(t *testing.T) {
	mainDir := t.TempDir()
	initExecRepo(t, mainDir)

	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitExec(t, mainDir, "worktree", "add", worktreePath, "-b", "400-feature-test")

	detector := NewContextDetector(worktreePath)
	isFeature, err := detector.IsFeatureBranch()
	if err != nil {
		t.Fatalf("IsFeatureBranch in worktree: %v", err)
	}
	if !isFeature {
		t.Error("expected 400-feature-test to be recognized as feature branch from worktree")
	}
}
