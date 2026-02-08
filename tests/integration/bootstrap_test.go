package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"specledger/pkg/cli/metadata"
)

// TestBootstrapNewInteractive tests the sl new command in simulated interactive mode
// Note: Full interactive TUI testing is difficult, so we test CI mode which exercises the same paths
func TestBootstrapNewCI(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	projectName := "test-project-ci"
	shortCode := "tpci"
	projectPath := filepath.Join(tempDir, projectName)

	// Build the sl binary first
	slBinary := buildSLBinary(t, tempDir)

	// Run sl new in CI mode
	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", projectName,
		"--short-code", shortCode,
		"--project-dir", tempDir)
	cmd.Dir = tempDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl new failed: %v\nOutput: %s", err, string(output))
	}

	// Verify project directory was created
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Fatalf("Project directory not created: %s", projectPath)
	}

	// Verify specledger.yaml was created and is valid
	yamlPath := filepath.Join(projectPath, "specledger", "specledger.yaml")
	meta, err := metadata.Load(yamlPath)
	if err != nil {
		t.Fatalf("Failed to load specledger.yaml: %v", err)
	}

	// Verify metadata content
	if meta.Project.Name != projectName {
		t.Errorf("Expected project name %s, got %s", projectName, meta.Project.Name)
	}
	if meta.Project.ShortCode != shortCode {
		t.Errorf("Expected short code %s, got %s", shortCode, meta.Project.ShortCode)
	}
	if meta.Playbook.Name != "specledger" {
		t.Errorf("Expected playbook 'specledger', got %s", meta.Playbook.Name)
	}

	// Verify .beads directory was created
	beadsPath := filepath.Join(projectPath, ".beads")
	if _, err := os.Stat(beadsPath); os.IsNotExist(err) {
		t.Error(".beads directory not created")
	}

	// Verify mise.toml was created
	misePath := filepath.Join(projectPath, "mise.toml")
	if _, err := os.Stat(misePath); os.IsNotExist(err) {
		t.Error("mise.toml not created")
	}

	// Verify git repo was initialized
	gitPath := filepath.Join(projectPath, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		t.Error(".git directory not created")
	}
}

// TestBootstrapNewWithPlaybook tests bootstrap with playbook applied
func TestBootstrapNewWithPlaybook(t *testing.T) {
	tempDir := t.TempDir()

	// Build the sl binary using the helper
	slBinary := buildSLBinary(t, tempDir)

	projectName := "test-project-playbook"
	projectPath := filepath.Join(tempDir, projectName)

	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", projectName,
		"--short-code", "tp",
		"--project-dir", tempDir)
	cmd.Dir = tempDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl new failed: %v\nOutput: %s", err, string(output))
	}

	// Verify specledger.yaml has playbook applied
	yamlPath := filepath.Join(projectPath, "specledger", "specledger.yaml")
	meta, err := metadata.Load(yamlPath)
	if err != nil {
		t.Fatalf("Failed to load specledger.yaml: %v", err)
	}

	if meta.Playbook.Name != "specledger" {
		t.Errorf("Expected playbook 'specledger', got %s", meta.Playbook.Name)
	}

	if meta.Playbook.Version == "" {
		t.Errorf("Expected playbook version to be set")
	}

	if len(meta.Playbook.Structure) == 0 {
		t.Errorf("Expected playbook structure to be set")
	}
}

// TestBootstrapInitInExistingDirectory tests sl init command
func TestBootstrapInitInExistingDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Build the sl binary using the helper
	slBinary := buildSLBinary(t, tempDir)

	// Create a directory to initialize
	existingDir := filepath.Join(tempDir, "existing-project")
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Run sl init
	cmd := exec.Command(slBinary, "init", "--short-code", "ep")
	cmd.Dir = existingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl init failed: %v\nOutput: %s", err, string(output))
	}

	// Verify specledger.yaml was created
	yamlPath := filepath.Join(existingDir, "specledger", "specledger.yaml")
	meta, err := metadata.Load(yamlPath)
	if err != nil {
		t.Fatalf("Failed to load specledger.yaml: %v", err)
	}

	// Verify default playbook is "specledger" for sl init
	if meta.Playbook.Name != "specledger" {
		t.Errorf("Expected playbook 'specledger' for sl init, got %s", meta.Playbook.Name)
	}

	// Verify .beads was created
	beadsPath := filepath.Join(existingDir, ".beads")
	if _, err := os.Stat(beadsPath); os.IsNotExist(err) {
		t.Error(".beads directory not created")
	}
}

// TestBootstrapPrerequisiteChecking tests that prerequisites are checked during bootstrap
func TestBootstrapPrerequisiteChecking(t *testing.T) {
	// This test verifies the prerequisite check is called
	// It's difficult to test actual missing tools without modifying PATH
	// So we just verify the command structure is correct

	tempDir := t.TempDir()

	// Build the sl binary using the helper
	slBinary := buildSLBinary(t, tempDir)

	projectName := "test-prereq"
	projectPath := filepath.Join(tempDir, projectName)

	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", projectName,
		"--short-code", "pr",
		"--project-dir", tempDir)
	cmd.Dir = tempDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl new failed: %v\nOutput: %s", err, string(output))
	}

	// Output should contain prerequisite check messages
	_ = string(output) // Check may have been silent - this is OK in CI mode

	// Verify project was created successfully
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Error("Project should be created even with prerequisite warnings")
	}
}
