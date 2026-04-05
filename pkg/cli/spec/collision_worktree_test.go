package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetNextFeatureNum_Worktree(t *testing.T) {
	mainDir := t.TempDir()
	initExecRepo(t, mainDir)

	// Create a feature branch in main repo
	gitExec(t, mainDir, "branch", "020-existing-feature")

	// Create specledger dir with a feature
	specDir := filepath.Join(mainDir, "specledger")
	if err := os.Mkdir(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	featureDir := filepath.Join(specDir, "005-feature")
	if err := os.Mkdir(featureDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(featureDir, "spec.md"), []byte("# Spec"), 0644); err != nil {
		t.Fatal(err)
	}
	gitExec(t, mainDir, "add", "-A")
	gitExec(t, mainDir, "commit", "-m", "add spec dir")

	// Create worktree
	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitExec(t, mainDir, "worktree", "add", worktreePath, "-b", "wt-collision")

	// GetNextFeatureNum from worktree should see both dir (005) and branch (020)
	got, err := GetNextFeatureNum(worktreePath)
	if err != nil {
		t.Fatalf("GetNextFeatureNum in worktree: %v", err)
	}
	if got != "021" {
		t.Errorf("got %q, want %q (should see branch 020-existing-feature from worktree)", got, "021")
	}
}
