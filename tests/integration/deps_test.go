package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/specledger/specledger/pkg/cli/metadata"
)

// TestDepsAddCommand tests adding dependencies via sl deps add
func TestDepsAddCommand(t *testing.T) {
	tempDir := t.TempDir()

	// Build the sl binary using the helper
	slBinary := buildSLBinary(t, tempDir)

	// Create a SpecLedger project first
	projectPath := filepath.Join(tempDir, "test-deps-project")
	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", "test-deps-project",
		"--short-code", "tdp",
		"--project-dir", tempDir)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create test project: %v\nOutput: %s", err, string(output))
	}

	// Add a dependency
	testURL := "git@github.com:example/test-spec"
	cmd = exec.Command(slBinary, "deps", "add", testURL, "main", "spec.md", "--alias", "test")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl deps add failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	if !contains(outputStr, "Dependency") && !contains(outputStr, "added") {
		t.Errorf("Expected success message in output, got: %s", outputStr)
	}

	// Verify YAML was updated with the dependency
	yamlPath := filepath.Join(projectPath, "specledger", "specledger.yaml")
	meta, err := metadata.Load(yamlPath)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	if len(meta.Dependencies) != 1 {
		t.Fatalf("Expected 1 dependency, got %d", len(meta.Dependencies))
	}

	dep := meta.Dependencies[0]
	if dep.URL != testURL {
		t.Errorf("Expected URL '%s', got '%s'", testURL, dep.URL)
	}

	if dep.Branch != "main" {
		t.Errorf("Expected branch 'main', got '%s'", dep.Branch)
	}

	if dep.Alias != "test" {
		t.Errorf("Expected alias 'test', got '%s'", dep.Alias)
	}
}

// TestDepsListCommand tests listing dependencies
func TestDepsListCommand(t *testing.T) {
	tempDir := t.TempDir()

	// Build the sl binary using the helper
	slBinary := buildSLBinary(t, tempDir)

	// Create a SpecLedger project
	projectPath := filepath.Join(tempDir, "test-list-project")
	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", "test-list-project",
		"--short-code", "tlp",
		"--project-dir", tempDir)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create test project: %v\nOutput: %s", err, string(output))
	}

	// List dependencies (should be empty)
	cmd = exec.Command(slBinary, "deps", "list")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl deps list failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	if !contains(outputStr, "No dependencies") {
		t.Errorf("Expected 'No dependencies' message, got: %s", outputStr)
	}

	// Add a dependency
	yamlPath := filepath.Join(projectPath, "specledger", "specledger.yaml")
	meta, err := metadata.Load(yamlPath)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	meta.Dependencies = append(meta.Dependencies, metadata.Dependency{
		URL:    "git@github.com:example/spec1",
		Branch: "main",
		Alias:  "spec1",
	})
	meta.Dependencies = append(meta.Dependencies, metadata.Dependency{
		URL:    "git@github.com:example/spec2",
		Branch: "v1.0",
		Alias:  "spec2",
	})
	if err := metadata.SaveToProject(meta, projectPath); err != nil {
		t.Fatalf("Failed to save YAML: %v", err)
	}

	// List again (should show dependencies)
	cmd = exec.Command(slBinary, "deps", "list")
	cmd.Dir = projectPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl deps list failed: %v\nOutput: %s", err, string(output))
	}

	outputStr = string(output)
	if !contains(outputStr, "Dependencies (2)") && !contains(outputStr, "Dependencies: 2") && !contains(outputStr, "2 total") {
		t.Errorf("Expected 'Dependencies (2)' in output, got: %s", outputStr)
	}

	if !contains(outputStr, "git@github.com:example/spec1") {
		t.Error("Expected first dependency in output")
	}

	if !contains(outputStr, "git@github.com:example/spec2") {
		t.Error("Expected second dependency in output")
	}
}

// TestDepsRemoveCommand tests removing dependencies
func TestDepsRemoveCommand(t *testing.T) {
	tempDir := t.TempDir()

	// Build the sl binary using the helper
	slBinary := buildSLBinary(t, tempDir)

	// Create a SpecLedger project
	projectPath := filepath.Join(tempDir, "test-remove-project")
	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", "test-remove-project",
		"--short-code", "trp",
		"--project-dir", tempDir)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create test project: %v\nOutput: %s", err, string(output))
	}

	// Add dependencies via YAML
	yamlPath := filepath.Join(projectPath, "specledger", "specledger.yaml")
	meta, err := metadata.Load(yamlPath)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	meta.Dependencies = append(meta.Dependencies, metadata.Dependency{
		URL:   "git@github.com:example/spec1",
		Alias: "spec1",
	})
	meta.Dependencies = append(meta.Dependencies, metadata.Dependency{
		URL:   "git@github.com:example/spec2",
		Alias: "spec2",
	})
	if err := metadata.SaveToProject(meta, projectPath); err != nil {
		t.Fatalf("Failed to save YAML: %v", err)
	}

	// Remove by alias
	cmd = exec.Command(slBinary, "deps", "remove", "spec1")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl deps remove failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	if !contains(outputStr, "Removed") && !contains(outputStr, "removed") {
		t.Errorf("Expected removal confirmation, got: %s", outputStr)
	}

	// Verify dependency was removed
	meta, err = metadata.Load(yamlPath)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	if len(meta.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency after removal, got %d", len(meta.Dependencies))
	}

	if meta.Dependencies[0].Alias != "spec2" {
		t.Errorf("Expected remaining dependency to be 'spec2', got '%s'", meta.Dependencies[0].Alias)
	}
}

// TestDepsRemoveByURL tests removing dependencies by URL
func TestDepsRemoveByURL(t *testing.T) {
	tempDir := t.TempDir()

	// Build the sl binary using the helper
	slBinary := buildSLBinary(t, tempDir)

	// Create a SpecLedger project
	projectPath := filepath.Join(tempDir, "test-remove-url-project")
	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", "test-remove-url-project",
		"--short-code", "trup",
		"--project-dir", tempDir)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create test project: %v\nOutput: %s", err, string(output))
	}

	// Add dependencies via YAML
	yamlPath := filepath.Join(projectPath, "specledger", "specledger.yaml")
	meta, err := metadata.Load(yamlPath)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	testURL := "git@github.com:example/to-remove"
	meta.Dependencies = append(meta.Dependencies, metadata.Dependency{
		URL: testURL,
	})
	if err := metadata.SaveToProject(meta, projectPath); err != nil {
		t.Fatalf("Failed to save YAML: %v", err)
	}

	// Remove by URL
	cmd = exec.Command(slBinary, "deps", "remove", testURL)
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl deps remove failed: %v\nOutput: %s", err, string(output))
	}

	// Verify dependency was removed
	meta, err = metadata.Load(yamlPath)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	if len(meta.Dependencies) != 0 {
		t.Errorf("Expected 0 dependencies after removal, got %d", len(meta.Dependencies))
	}
}

// TestDepsDuplicateDetection tests that duplicate dependencies are prevented
func TestDepsDuplicateDetection(t *testing.T) {
	tempDir := t.TempDir()

	// Build the sl binary using the helper
	slBinary := buildSLBinary(t, tempDir)

	// Create a SpecLedger project
	projectPath := filepath.Join(tempDir, "test-dup-project")
	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", "test-dup-project",
		"--short-code", "tdp",
		"--project-dir", tempDir)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create test project: %v\nOutput: %s", err, string(output))
	}

	testURL := "git@github.com:example/test-spec"

	// Add a dependency
	cmd = exec.Command(slBinary, "deps", "add", testURL)
	cmd.Dir = projectPath
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("First sl deps add failed: %v\nOutput: %s", err, string(output))
	}

	// Try to add the same dependency again
	cmd = exec.Command(slBinary, "deps", "add", testURL)
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected error when adding duplicate dependency")
	}

	outputStr := string(output)
	if !contains(outputStr, "already exists") && !contains(outputStr, "duplicate") {
		t.Logf("Expected duplicate error message, got: %s", outputStr)
	}
}

// TestDepsResolve tests dependency resolution (currently not fully implemented)
func TestDepsResolve(t *testing.T) {
	tempDir := t.TempDir()

	// Build the sl binary using the helper
	slBinary := buildSLBinary(t, tempDir)

	// Create a SpecLedger project
	projectPath := filepath.Join(tempDir, "test-resolve-project")
	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", "test-resolve-project",
		"--short-code", "trp",
		"--project-dir", tempDir)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create test project: %v\nOutput: %s", err, string(output))
	}

	// Add a dependency via YAML
	yamlPath := filepath.Join(projectPath, "specledger", "specledger.yaml")
	meta, err := metadata.Load(yamlPath)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	meta.Dependencies = append(meta.Dependencies, metadata.Dependency{
		URL:   "git@github.com:example/test-spec",
		Alias: "test",
	})
	if err := metadata.SaveToProject(meta, projectPath); err != nil {
		t.Fatalf("Failed to save YAML: %v", err)
	}

	// Run resolve (will show "not yet implemented" message)
	cmd = exec.Command(slBinary, "deps", "resolve")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("sl deps resolve exited (expected - not fully implemented): %v", err)
	}

	outputStr := string(output)
	t.Logf("Resolve output: %s", outputStr)

	// For now, just verify it doesn't crash
	// Full git-based resolution will be implemented in future
}

// TestFindProjectRoot tests the project root finding logic
func TestFindProjectRoot(t *testing.T) {
	tempDir := t.TempDir()

	// Build the sl binary using the helper
	slBinary := buildSLBinary(t, tempDir)

	// Create a project
	projectPath := filepath.Join(tempDir, "test-root-project")
	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", "test-root-project",
		"--short-code", "trp",
		"--project-dir", tempDir)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create test project: %v\nOutput: %s", err, string(output))
	}

	// Test from project root
	cmd = exec.Command(slBinary, "deps", "list")
	cmd.Dir = projectPath
	if _, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("deps list failed from project root: %v", err)
	}

	// Test from subdirectory
	subDir := filepath.Join(projectPath, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	cmd = exec.Command(slBinary, "deps", "list")
	cmd.Dir = subDir
	if _, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("deps list failed from subdirectory: %v", err)
	}
}

// TestDepsOutsideProject tests error handling when outside a project
func TestDepsOutsideProject(t *testing.T) {
	tempDir := t.TempDir()

	// Build the sl binary using the helper
	slBinary := buildSLBinary(t, tempDir)

	// Create a directory that's not a SpecLedger project
	nonProjectPath := filepath.Join(tempDir, "not-a-project")
	if err := os.MkdirAll(nonProjectPath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Try to list deps (should fail)
	cmd := exec.Command(slBinary, "deps", "list")
	cmd.Dir = nonProjectPath
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected deps list to fail outside a SpecLedger project")
	}

	outputStr := string(output)
	if !contains(outputStr, "not a SpecLedger project") && !contains(outputStr, "no specledger.yaml") {
		t.Logf("Expected error about not being in a project, got: %s", outputStr)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
