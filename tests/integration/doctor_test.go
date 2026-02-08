package integration

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// DoctorOutput represents the JSON output structure from sl doctor
type DoctorOutput struct {
	Status              string         `json:"status"`
	Tools               []DoctorTool   `json:"tools"`
	Missing             []string       `json:"missing,omitempty"`
	InstallInstructions string         `json:"install_instructions,omitempty"`
}

type DoctorTool struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
	Path      string `json:"path,omitempty"`
	Category  string `json:"category"`
}

// TestDoctorCommand tests that sl doctor command runs successfully
func TestDoctorCommand(t *testing.T) {
	tempDir := t.TempDir()

	// Build the sl binary
	slBinary := buildSLBinary(t, tempDir)

	// Run sl doctor (human-readable output)
	cmd := exec.Command(slBinary, "doctor")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// sl doctor may exit with error if core tools are missing
		// That's expected behavior - just check it ran
		t.Logf("sl doctor exited with error (expected if tools missing): %v", err)
	}

	outputStr := string(output)

	// Verify output contains expected sections
	expectedSections := []string{
		"SpecLedger Doctor",
		"Core Tools",
		"SDD Framework Tools",
	}

	for _, section := range expectedSections {
		if !strings.Contains(outputStr, section) {
			t.Errorf("Expected output to contain '%s', got: %s", section, outputStr)
		}
	}

	// Verify checkmarks or X marks are used
	if !strings.Contains(outputStr, "✓") && !strings.Contains(outputStr, "✗") {
		// May use different unicode or ASCII
		if !strings.Contains(outputStr, "[") && !strings.Contains(outputStr, "]") {
			t.Error("Expected output to use status indicators")
		}
	}
}

// TestDoctorJSONOutput tests the --json flag for sl doctor
func TestDoctorJSONOutput(t *testing.T) {
	tempDir := t.TempDir()

	slBinary := buildSLBinary(t, tempDir)

	// Run sl doctor --json
	cmd := exec.Command(slBinary, "doctor", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// May exit with error if tools missing - that's OK
		t.Logf("sl doctor --json exited with error (may be expected): %v", err)
	}

	// Parse JSON output
	var doctorOutput DoctorOutput
	if err := json.Unmarshal(output, &doctorOutput); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, string(output))
	}

	// Verify JSON structure
	if doctorOutput.Status != "pass" && doctorOutput.Status != "fail" {
		t.Errorf("Expected status 'pass' or 'fail', got '%s'", doctorOutput.Status)
	}

	if len(doctorOutput.Tools) == 0 {
		t.Error("Expected tools array to have at least one entry")
	}

	// Verify tool categories
	hasCoreTools := false
	hasFrameworkTools := false
	for _, tool := range doctorOutput.Tools {
		if tool.Category == "core" {
			hasCoreTools = true
		}
		if tool.Category == "framework" {
			hasFrameworkTools = true
		}
	}

	if !hasCoreTools {
		t.Error("Expected at least one core tool in output")
	}

	if !hasFrameworkTools {
		t.Error("Expected at least one framework tool in output")
	}

	// Verify status consistency
	if doctorOutput.Status == "fail" {
		if len(doctorOutput.Missing) == 0 && doctorOutput.InstallInstructions == "" {
			t.Error("Expected missing tools or install instructions when status is 'fail'")
		}
	}
}

// TestDoctorToolDetection tests that sl doctor detects the expected tools
func TestDoctorToolDetection(t *testing.T) {
	tempDir := t.TempDir()

	slBinary := buildSLBinary(t, tempDir)

	// Run sl doctor --json for programmatic checking
	cmd := exec.Command(slBinary, "doctor", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("sl doctor --json exited (may be expected): %v", err)
	}

	var doctorOutput DoctorOutput
	if err := json.Unmarshal(output, &doctorOutput); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify expected core tools are checked
	expectedCoreTools := []string{"mise", "bd", "perles"}
	expectedFrameworkTools := []string{"specify", "openspec"}

	// Check that all expected tools are present in output
	for _, expectedTool := range expectedCoreTools {
		found := false
		for _, tool := range doctorOutput.Tools {
			if tool.Name == expectedTool && tool.Category == "core" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected core tool '%s' not found in output", expectedTool)
		}
	}

	for _, expectedTool := range expectedFrameworkTools {
		found := false
		for _, tool := range doctorOutput.Tools {
			if tool.Name == expectedTool && tool.Category == "framework" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected framework tool '%s' not found in output", expectedTool)
		}
	}
}

// TestDoctorExitCode tests that sl doctor returns appropriate exit codes
func TestDoctorExitCode(t *testing.T) {
	tempDir := t.TempDir()

	slBinary := buildSLBinary(t, tempDir)

	// Run sl doctor
	cmd := exec.Command(slBinary, "doctor")
	err := cmd.Run()

	// Check if core tools are installed
	miseInstalled := toolExists("mise")
	bdInstalled := toolExists("bd")
	perlesInstalled := toolExists("perles")

	allCoreInstalled := miseInstalled && bdInstalled && perlesInstalled

	// If all core tools are installed, command should succeed
	// If any are missing, command should fail
	if allCoreInstalled {
		if err != nil {
			t.Errorf("Expected sl doctor to succeed when all core tools are installed, got: %v", err)
		}
	} else {
		if err == nil {
			t.Error("Expected sl doctor to fail when core tools are missing")
		}
	}
}

// toolExists checks if a command is available in PATH
func toolExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// TestDoctorInProjectDirectory tests sl doctor works from within a SpecLedger project
func TestDoctorInProjectDirectory(t *testing.T) {
	tempDir := t.TempDir()

	slBinary := buildSLBinary(t, tempDir)

	// Create a SpecLedger project first
	projectPath := filepath.Join(tempDir, "test-project")
	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", "test-project",
		"--short-code", "tp",
		"--project-dir", tempDir)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create test project: %v\nOutput: %s", err, string(output))
	}

	// Run sl doctor from within the project
	cmd = exec.Command(slBinary, "doctor")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()

	// Should work regardless of project location
	outputStr := string(output)
	if !strings.Contains(outputStr, "SpecLedger Doctor") {
		t.Errorf("Expected 'SpecLedger Doctor' in output, got: %s", outputStr)
	}

	t.Logf("sl doctor output from project directory: %s", outputStr)
	if err != nil {
		t.Logf("sl doctor exited with error (expected if tools missing): %v", err)
	}
}
