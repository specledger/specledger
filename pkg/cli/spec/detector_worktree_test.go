package spec

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

func TestDetectFeatureContext_Worktree(t *testing.T) {
	mainDir := t.TempDir()
	initExecRepo(t, mainDir)

	// Create specledger feature dir in main repo
	featureDir := filepath.Join(mainDir, "specledger", "100-worktree-feature")
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create minimal spec.md so the feature dir is valid
	if err := os.WriteFile(filepath.Join(featureDir, "spec.md"), []byte("# Spec"), 0644); err != nil {
		t.Fatal(err)
	}
	gitExec(t, mainDir, "add", "-A")
	gitExec(t, mainDir, "commit", "-m", "add spec")

	// Create worktree on a feature branch matching the spec pattern
	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitExec(t, mainDir, "worktree", "add", worktreePath, "-b", "100-worktree-feature")

	ctx, err := DetectFeatureContext(worktreePath)
	if err != nil {
		t.Fatalf("DetectFeatureContext in worktree: %v", err)
	}

	if ctx.Branch != "100-worktree-feature" {
		t.Errorf("Branch = %q, want %q", ctx.Branch, "100-worktree-feature")
	}
	if !ctx.HasGit {
		t.Error("expected HasGit = true")
	}
}

func TestDetectFeatureContext_Worktree_SpecOverride(t *testing.T) {
	mainDir := t.TempDir()
	initExecRepo(t, mainDir)

	// Create specledger feature dir
	featureDir := filepath.Join(mainDir, "specledger", "200-override-test")
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(featureDir, "spec.md"), []byte("# Spec"), 0644); err != nil {
		t.Fatal(err)
	}
	gitExec(t, mainDir, "add", "-A")
	gitExec(t, mainDir, "commit", "-m", "add spec")

	// Create worktree on a non-feature branch
	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitExec(t, mainDir, "worktree", "add", worktreePath, "-b", "random-branch")

	// SpecOverride should work from worktree
	ctx, err := DetectFeatureContextWithOptions(worktreePath, DetectionOptions{
		SpecOverride: "200-override-test",
	})
	if err != nil {
		t.Fatalf("DetectFeatureContextWithOptions with SpecOverride in worktree: %v", err)
	}

	if ctx.Branch != "200-override-test" {
		t.Errorf("Branch = %q, want %q", ctx.Branch, "200-override-test")
	}
}

func TestGetCurrentBranch_Worktree_Detector(t *testing.T) {
	mainDir := t.TempDir()
	initExecRepo(t, mainDir)

	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitExec(t, mainDir, "worktree", "add", worktreePath, "-b", "detector-wt-branch")

	branch, err := GetCurrentBranch(worktreePath)
	if err != nil {
		t.Fatalf("GetCurrentBranch (detector) in worktree: %v", err)
	}
	if branch != "detector-wt-branch" {
		t.Errorf("got %q, want %q", branch, "detector-wt-branch")
	}
}

func TestBranchExists_Worktree_Detector(t *testing.T) {
	mainDir := t.TempDir()
	initExecRepo(t, mainDir)
	gitExec(t, mainDir, "branch", "some-feature")

	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitExec(t, mainDir, "worktree", "add", worktreePath, "-b", "detector-wt")

	exists, err := BranchExists(worktreePath, "some-feature")
	if err != nil {
		t.Fatalf("BranchExists (detector) in worktree: %v", err)
	}
	if !exists {
		t.Error("expected some-feature to exist from worktree")
	}
}
