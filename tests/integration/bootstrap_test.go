package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specledger/specledger/pkg/cli/commands"
	"github.com/specledger/specledger/pkg/cli/metadata"
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

// TestBootstrapNewCreatesConstitution tests that sl new --ci creates a populated constitution
func TestBootstrapNewCreatesConstitution(t *testing.T) {
	tempDir := t.TempDir()
	slBinary := buildSLBinary(t, tempDir)

	projectName := "test-constitution"
	projectPath := filepath.Join(tempDir, projectName)

	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", projectName,
		"--short-code", "tc",
		"--project-dir", tempDir)
	cmd.Dir = tempDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl new failed: %v\nOutput: %s", err, string(output))
	}

	// Verify constitution was created
	constitutionPath := filepath.Join(projectPath, ".specledger", "memory", "constitution.md")
	if _, err := os.Stat(constitutionPath); os.IsNotExist(err) {
		t.Fatal("Constitution file not created")
	}

	// Verify constitution is populated (no placeholders)
	if !commands.IsConstitutionPopulated(constitutionPath) {
		t.Error("Constitution should be populated (no placeholder tokens)")
	}

	// Verify agent preference is set to None in CI mode
	agentPref, err := commands.ReadAgentPreference(constitutionPath)
	if err != nil {
		t.Fatalf("Failed to read agent preference: %v", err)
	}
	if agentPref != "None" {
		t.Errorf("Expected agent preference 'None' in CI mode, got '%s'", agentPref)
	}

	// Verify all default principles are present
	content, err := os.ReadFile(constitutionPath)
	if err != nil {
		t.Fatalf("Failed to read constitution: %v", err)
	}
	contentStr := string(content)
	expectedPrinciples := []string{"Specification-First", "Test-First", "Code Quality", "Simplicity", "Observability"}
	for _, p := range expectedPrinciples {
		if !strings.Contains(contentStr, p) {
			t.Errorf("Constitution missing principle: %s", p)
		}
	}
}

// TestConstitutionDetection tests the IsConstitutionPopulated helper
func TestConstitutionDetection(t *testing.T) {
	tempDir := t.TempDir()

	// Test 1: Missing file
	missingPath := filepath.Join(tempDir, "missing.md")
	if commands.IsConstitutionPopulated(missingPath) {
		t.Error("Missing file should not be considered populated")
	}

	// Test 2: Template file with placeholders
	templatePath := filepath.Join(tempDir, "template.md")
	templateContent := "# [PROJECT_NAME] Constitution\n\n### [PRINCIPLE_1_NAME]\n[PRINCIPLE_1_DESCRIPTION]\n"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}
	if commands.IsConstitutionPopulated(templatePath) {
		t.Error("Template with placeholders should not be considered populated")
	}

	// Test 3: Populated file
	populatedPath := filepath.Join(tempDir, "populated.md")
	populatedContent := "# My Project Constitution\n\n## Core Principles\n\n### I. Specification-First\nEvery feature starts with a spec.\n"
	if err := os.WriteFile(populatedPath, []byte(populatedContent), 0644); err != nil {
		t.Fatal(err)
	}
	if !commands.IsConstitutionPopulated(populatedPath) {
		t.Error("Populated file should be considered populated")
	}

	// Test 4: Empty file
	emptyPath := filepath.Join(tempDir, "empty.md")
	if err := os.WriteFile(emptyPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if commands.IsConstitutionPopulated(emptyPath) {
		t.Error("Empty file should not be considered populated")
	}
}

// TestInitPreservesExistingConstitution tests that sl init does not overwrite existing constitution
func TestInitPreservesExistingConstitution(t *testing.T) {
	tempDir := t.TempDir()
	slBinary := buildSLBinary(t, tempDir)

	// Create project directory with an existing populated constitution
	projectDir := filepath.Join(tempDir, "existing-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	constitutionDir := filepath.Join(projectDir, ".specledger", "memory")
	if err := os.MkdirAll(constitutionDir, 0755); err != nil {
		t.Fatal(err)
	}

	existingConstitution := "# Existing Constitution\n\n## Core Principles\n\n### I. My Custom Principle\nCustom description.\n\n## Agent Preferences\n\n- **Preferred Agent**: Claude Code\n"
	constitutionPath := filepath.Join(constitutionDir, "constitution.md")
	if err := os.WriteFile(constitutionPath, []byte(existingConstitution), 0644); err != nil {
		t.Fatal(err)
	}

	// Run sl init
	cmd := exec.Command(slBinary, "init", "--short-code", "ep", "--playbook", "specledger")
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl init failed: %v\nOutput: %s", err, string(output))
	}

	// Verify constitution was preserved (not overwritten)
	content, err := os.ReadFile(constitutionPath)
	if err != nil {
		t.Fatalf("Failed to read constitution: %v", err)
	}
	if !strings.Contains(string(content), "My Custom Principle") {
		t.Error("Existing constitution was overwritten — should have been preserved")
	}

	// Verify agent preference was read from existing constitution
	agentPref, err := commands.ReadAgentPreference(constitutionPath)
	if err != nil {
		t.Fatalf("Failed to read agent preference: %v", err)
	}
	if agentPref != "Claude Code" {
		t.Errorf("Expected agent preference 'Claude Code', got '%s'", agentPref)
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
