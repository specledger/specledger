package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestSpecCreateInWorktree verifies that `sl spec create` works inside a git worktree.
// Regression test for https://github.com/specledger/specledger/issues/159
func TestSpecCreateInWorktree(t *testing.T) {
	slBinary := buildSLBinary(t, t.TempDir())

	// Set up a project in a temp directory
	projectDir := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Initialize sl project
	initCmd := exec.Command(slBinary, "init", "--short-code", "wt", "--ci")
	initCmd.Dir = projectDir
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("sl init failed: %v\nOutput: %s", err, string(output))
	}

	// Initialize git repo with initial commit
	initGitRepo(t, projectDir)

	// Create a worktree
	worktreePath := filepath.Join(t.TempDir(), "worktree")
	addWorktree(t, projectDir, worktreePath, "worktree-branch")

	// Run sl spec create from the worktree
	createCmd := exec.Command(slBinary, "spec", "create", "--short-name", "worktree-test", "--number", "900", "--json")
	createCmd.Dir = worktreePath
	output, err := createCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl spec create in worktree failed: %v\nOutput: %s", err, string(output))
	}

	// Parse JSON output
	var result struct {
		BranchName string `json:"BRANCH_NAME"`
		FeatureDir string `json:"FEATURE_DIR"`
		FeatureNum string `json:"FEATURE_NUM"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nRaw: %s", err, string(output))
	}

	if result.FeatureNum != "900" {
		t.Errorf("FEATURE_NUM = %q, want %q", result.FeatureNum, "900")
	}
	if result.BranchName == "" {
		t.Error("BRANCH_NAME is empty")
	}
	if result.FeatureDir == "" {
		t.Error("FEATURE_DIR is empty")
	}

	// Verify the spec directory was created
	specDir := filepath.Join(worktreePath, "specledger", result.BranchName)
	if _, err := os.Stat(specDir); os.IsNotExist(err) {
		t.Errorf("spec directory not created at %s", specDir)
	}
}
