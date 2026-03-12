package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/specledger/specledger/pkg/cli/spec"
)

func createSpecDir(t *testing.T, feature string) string {
	t.Helper()
	dir := t.TempDir()
	specDir := filepath.Join(dir, "specledger", feature)
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	return specDir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestApproveWorkflow_SuccessfulApproval(t *testing.T) {
	specDir := createSpecDir(t, "127-test-feature")

	specContent := `# Feature Specification: Test Feature

**Feature Branch**: 127-test-feature
**Status**: Draft

## Requirements
Some requirements here.
`
	writeFile(t, filepath.Join(specDir, "spec.md"), specContent)
	writeFile(t, filepath.Join(specDir, "plan.md"), "# Plan\nSome plan content.\n")
	writeFile(t, filepath.Join(specDir, "tasks.md"), "# Tasks\nSome tasks.\n")

	// Read status — should be Draft
	status, err := spec.ReadStatus(specDir)
	if err != nil {
		t.Fatalf("ReadStatus() error: %v", err)
	}
	if status != "Draft" {
		t.Errorf("expected Draft, got %s", status)
	}

	// Write approved status
	if err := spec.WriteStatus(specDir, "Approved"); err != nil {
		t.Fatalf("WriteStatus() error: %v", err)
	}

	// Verify status changed
	status, err = spec.ReadStatus(specDir)
	if err != nil {
		t.Fatalf("ReadStatus() after approval error: %v", err)
	}
	if status != "Approved" {
		t.Errorf("expected Approved, got %s", status)
	}
}

func TestApproveWorkflow_MissingArtifacts(t *testing.T) {
	specDir := createSpecDir(t, "128-missing-artifacts")

	specContent := `# Feature Specification: Missing

**Status**: Draft
`
	writeFile(t, filepath.Join(specDir, "spec.md"), specContent)
	// plan.md and tasks.md intentionally missing

	// Verify artifacts are missing
	planPath := filepath.Join(specDir, "plan.md")
	if _, err := os.Stat(planPath); !os.IsNotExist(err) {
		t.Error("plan.md should not exist for this test")
	}

	tasksPath := filepath.Join(specDir, "tasks.md")
	if _, err := os.Stat(tasksPath); !os.IsNotExist(err) {
		t.Error("tasks.md should not exist for this test")
	}
}

func TestApproveWorkflow_AlreadyApproved(t *testing.T) {
	specDir := createSpecDir(t, "129-already-approved")

	specContent := `# Feature Specification: Already Approved

**Status**: Approved

## Requirements
Already approved spec.
`
	writeFile(t, filepath.Join(specDir, "spec.md"), specContent)

	status, err := spec.ReadStatus(specDir)
	if err != nil {
		t.Fatalf("ReadStatus() error: %v", err)
	}
	if status != "Approved" {
		t.Errorf("expected Approved, got %s", status)
	}

	// Writing Approved again should be idempotent
	if err := spec.WriteStatus(specDir, "Approved"); err != nil {
		t.Fatalf("WriteStatus() error: %v", err)
	}

	status, _ = spec.ReadStatus(specDir)
	if status != "Approved" {
		t.Errorf("expected Approved after re-write, got %s", status)
	}
}
