package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specledger/specledger/pkg/cli/playbooks"
)

// TestMiseTomlProtected tests that sl doctor --template does not overwrite customized mise.toml.
func TestMiseTomlProtected(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "mise-protected")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Init project — mise.toml should be created
	initProject(t, slBinary, projectDir)

	misePath := filepath.Join(projectDir, "mise.toml")
	if _, err := os.Stat(misePath); os.IsNotExist(err) {
		t.Fatal("mise.toml not created by sl init")
	}

	// Customize mise.toml with user content
	customContent := "# SpecLedger mise configuration\n\n[tools]\ngo = \"1.24\"\ngolangci-lint = \"2.11.4\"\n"
	if err := os.WriteFile(misePath, []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run sl doctor --template
	cmd := exec.Command(slBinary, "doctor", "--template")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl doctor --template failed: %v\nOutput: %s", err, string(output))
	}

	// Verify custom content is preserved
	afterDoctor, err := os.ReadFile(misePath)
	if err != nil {
		t.Fatalf("Failed to read mise.toml after doctor: %v", err)
	}

	got := string(afterDoctor)
	if got != customContent {
		t.Errorf("mise.toml was overwritten by doctor --template\ngot:\n%s\nwant:\n%s", got, customContent)
	}
}

// TestMiseTomlCreatedOnInit tests that mise.toml is created on first sl init even though it's protected.
func TestMiseTomlCreatedOnInit(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "mise-init")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Verify mise.toml does not exist before init
	misePath := filepath.Join(projectDir, "mise.toml")
	if _, err := os.Stat(misePath); err == nil {
		t.Fatal("mise.toml should not exist before init")
	}

	initProject(t, slBinary, projectDir)

	// Verify mise.toml was created
	if _, err := os.Stat(misePath); os.IsNotExist(err) {
		t.Error("mise.toml not created by sl init")
	}
}

// initGitRepo initializes a git repo with an initial commit so sl commands that require git work.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
		{"add", "-A"},
		{"commit", "-m", "init", "--allow-empty"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(output))
		}
	}
}

// TestContextUpdatePreservesUserContent tests that sl context update claude preserves existing CLAUDE.md content.
func TestContextUpdatePreservesUserContent(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "context-preserve")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Init project and git repo
	initProject(t, slBinary, projectDir)
	initGitRepo(t, projectDir)

	// Create a specledger spec with plan.md containing Technical Context
	specDir := filepath.Join(projectDir, "specledger", "test-feature")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	planContent := `# Implementation Plan

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: Cobra (CLI)
**Storage**: File-based
**Testing**: go test
`
	if err := os.WriteFile(filepath.Join(specDir, "plan.md"), []byte(planContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create existing CLAUDE.md with user content
	claudePath := filepath.Join(projectDir, "CLAUDE.md")
	userContent := "# My Project\n\n## Build Commands\n\nmake test\nmake lint\n\n## Architecture\n\nClean architecture with layers.\n"
	if err := os.WriteFile(claudePath, []byte(userContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run sl context update claude with --spec override
	cmd := exec.Command(slBinary, "context", "update", "claude", "--spec", "test-feature")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl context update claude failed: %v\nOutput: %s", err, string(output))
	}

	// Read updated file
	afterUpdate, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md after update: %v", err)
	}

	got := string(afterUpdate)

	// User content must be preserved
	if !strings.Contains(got, "# My Project") {
		t.Error("User content '# My Project' was lost")
	}
	if !strings.Contains(got, "make test") {
		t.Error("User content 'make test' was lost")
	}
	if !strings.Contains(got, "Clean architecture with layers") {
		t.Error("User content 'Clean architecture' was lost")
	}

	// Sentinel block must be present with managed tech
	if !strings.Contains(got, playbooks.HTMLMarkers.Begin) {
		t.Error("Missing sentinel begin marker")
	}
	if !strings.Contains(got, playbooks.HTMLMarkers.End) {
		t.Error("Missing sentinel end marker")
	}
	if !strings.Contains(got, "- Cobra (CLI)") {
		t.Error("Missing managed technology entry 'Cobra (CLI)'")
	}
	if !strings.Contains(got, "- Go 1.24") {
		t.Error("Missing managed technology entry 'Go 1.24'")
	}
}

// TestContextUpdateIdempotent tests that running sl context update claude twice produces identical output.
func TestContextUpdateIdempotent(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "context-idempotent")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	initProject(t, slBinary, projectDir)
	initGitRepo(t, projectDir)

	// Create spec with plan
	specDir := filepath.Join(projectDir, "specledger", "test-feature")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	planContent := `# Implementation Plan

## Technical Context

**Language/Version**: Python 3.12
**Primary Dependencies**: FastAPI, SQLAlchemy
**Testing**: pytest
`
	if err := os.WriteFile(filepath.Join(specDir, "plan.md"), []byte(planContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create existing CLAUDE.md
	claudePath := filepath.Join(projectDir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte("# Custom Project Docs\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// First update
	cmd1 := exec.Command(slBinary, "context", "update", "claude", "--spec", "test-feature")
	cmd1.Dir = projectDir
	if output, err := cmd1.CombinedOutput(); err != nil {
		t.Fatalf("First context update failed: %v\nOutput: %s", err, string(output))
	}
	first, _ := os.ReadFile(claudePath)

	// Second update
	cmd2 := exec.Command(slBinary, "context", "update", "claude", "--spec", "test-feature")
	cmd2.Dir = projectDir
	if output, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("Second context update failed: %v\nOutput: %s", err, string(output))
	}
	second, _ := os.ReadFile(claudePath)

	if string(first) != string(second) {
		t.Errorf("context update not idempotent:\nFirst:\n%s\nSecond:\n%s", first, second)
	}
}
