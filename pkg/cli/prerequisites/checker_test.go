package prerequisites

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/specledger/specledger/pkg/cli/metadata"
)

func TestCheckTool(t *testing.T) {
	// Test with a tool that should exist on most systems
	tool := Tool{
		Name:        "go",
		DisplayName: "Go compiler",
		Category:    metadata.ToolCategoryCore,
		VersionFlag: "version",
	}

	result := CheckTool(tool)

	if !result.Installed {
		t.Skip("go not installed, skipping test")
	}

	if result.Path == "" {
		t.Error("expected path to be set for installed tool")
	}

	if result.Version == "" {
		t.Error("expected version to be retrieved")
	}

	if !strings.Contains(result.Version, "go") {
		t.Errorf("expected version to contain 'go', got: %s", result.Version)
	}
}

func TestCheckToolNotInstalled(t *testing.T) {
	tool := Tool{
		Name:        "nonexistent-tool-12345",
		DisplayName: "Nonexistent Tool",
		Category:    metadata.ToolCategoryCore,
		VersionFlag: "--version",
	}

	result := CheckTool(tool)

	if result.Installed {
		t.Error("expected tool to not be installed")
	}

	if result.Path != "" {
		t.Error("expected path to be empty for missing tool")
	}

	if result.Version != "" {
		t.Error("expected version to be empty for missing tool")
	}

	if result.Error == nil {
		t.Error("expected error for missing tool")
	}
}

func TestIsCommandAvailable(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"existing command", "go", true},
		{"nonexistent command", "nonexistent-cmd-12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if the command we expect to exist doesn't exist
			if tt.expected {
				if _, err := exec.LookPath(tt.command); err != nil {
					t.Skipf("%s not available, skipping", tt.command)
				}
			}

			result := IsCommandAvailable(tt.command)
			if result != tt.expected {
				t.Errorf("IsCommandAvailable(%q) = %v, expected %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestIsMiseInstalled(t *testing.T) {
	// This test will pass or skip based on whether mise is installed
	result := IsMiseInstalled()

	_, err := exec.LookPath("mise")
	expected := err == nil

	if result != expected {
		t.Errorf("IsMiseInstalled() = %v, expected %v", result, expected)
	}
}

func TestGetMissingCoreTools(t *testing.T) {
	missing := GetMissingCoreTools()

	// missing should be a valid list (could be empty or contain tools)
	if missing == nil {
		t.Error("expected non-nil slice")
	}

	// All items in missing should be core tools
	for _, tool := range missing {
		found := false
		for _, coreTool := range CoreTools {
			if tool.Name == coreTool.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("tool %s not found in CoreTools list", tool.Name)
		}
	}
}

func TestAllCoreToolsInstalled(t *testing.T) {
	result := AllCoreToolsInstalled()

	// Verify consistency with GetMissingCoreTools
	missing := GetMissingCoreTools()
	expected := len(missing) == 0

	if result != expected {
		t.Errorf("AllCoreToolsInstalled() = %v, but missing tools count = %d", result, len(missing))
	}
}

func TestCheckAllTools(t *testing.T) {
	coreResults, frameworkResults := CheckAllTools()

	if len(coreResults) != len(CoreTools) {
		t.Errorf("expected %d core results, got %d", len(CoreTools), len(coreResults))
	}

	if len(frameworkResults) != len(FrameworkTools) {
		t.Errorf("expected %d framework results, got %d", len(FrameworkTools), len(frameworkResults))
	}

	// Verify all core tools are checked
	for i, result := range coreResults {
		if result.Tool.Name != CoreTools[i].Name {
			t.Errorf("core result %d: expected tool %s, got %s", i, CoreTools[i].Name, result.Tool.Name)
		}
	}

	// Verify all framework tools are checked
	for i, result := range frameworkResults {
		if result.Tool.Name != FrameworkTools[i].Name {
			t.Errorf("framework result %d: expected tool %s, got %s", i, FrameworkTools[i].Name, result.Tool.Name)
		}
	}
}

func TestCheckCoreTools(t *testing.T) {
	results := CheckCoreTools()

	if len(results) != len(CoreTools) {
		t.Errorf("expected %d results, got %d", len(CoreTools), len(results))
	}

	for i, result := range results {
		if result.Tool.Name != CoreTools[i].Name {
			t.Errorf("result %d: expected tool %s, got %s", i, CoreTools[i].Name, result.Tool.Name)
		}

		if result.Tool.Category != metadata.ToolCategoryCore {
			t.Errorf("tool %s: expected category core, got %s", result.Tool.Name, result.Tool.Category)
		}
	}
}

func TestGetInstallInstructions(t *testing.T) {
	tests := []struct {
		name     string
		missing  []Tool
		contains []string
	}{
		{
			"no missing tools",
			[]Tool{},
			[]string{"All required tools"},
		},
		{
			"mise missing",
			[]Tool{CoreTools[0]}, // mise
			[]string{"mise", "https://mise.jdx.dev"},
		},
		{
			"bd missing",
			[]Tool{CoreTools[1]}, // bd
			[]string{"beads", "mise install"},
		},
		{
			"multiple tools missing",
			[]Tool{CoreTools[1], CoreTools[2]}, // bd and perles
			[]string{"beads", "perles", "mise install"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instructions := GetInstallInstructions(tt.missing)

			for _, substr := range tt.contains {
				if !strings.Contains(instructions, substr) {
					t.Errorf("expected instructions to contain %q, got:\n%s", substr, instructions)
				}
			}
		})
	}
}

func TestCheckPrerequisites(t *testing.T) {
	check := CheckPrerequisites()

	// Verify consistency
	if check.AllCoreInstalled && len(check.MissingCore) > 0 {
		t.Error("AllCoreInstalled is true but MissingCore is not empty")
	}

	if !check.AllCoreInstalled && len(check.MissingCore) == 0 {
		t.Error("AllCoreInstalled is false but MissingCore is empty")
	}

	// Verify results are populated
	if len(check.CoreResults) != len(CoreTools) {
		t.Errorf("expected %d core results, got %d", len(CoreTools), len(check.CoreResults))
	}

	if len(check.FrameworkResults) != len(FrameworkTools) {
		t.Errorf("expected %d framework results, got %d", len(FrameworkTools), len(check.FrameworkResults))
	}

	// Verify instructions are provided if tools are missing
	if !check.AllCoreInstalled && check.Instructions == "" {
		t.Error("expected instructions when core tools are missing")
	}

	if check.AllCoreInstalled && !strings.Contains(check.Instructions, "All required tools") {
		t.Error("expected success message when all tools are installed")
	}
}

func TestEnsurePrerequisitesNonInteractive(t *testing.T) {
	// CI mode should not prompt, just return error if tools missing
	err := EnsurePrerequisites(false)

	// If all core tools are installed, should return nil
	if AllCoreToolsInstalled() {
		if err != nil {
			t.Errorf("expected nil error when all tools installed, got: %v", err)
		}
	} else {
		// If tools are missing, should return error with instructions
		if err == nil {
			t.Error("expected error when tools missing in CI mode")
		}

		if !strings.Contains(err.Error(), "missing required tools") {
			t.Errorf("expected error message to mention missing tools, got: %v", err)
		}
	}
}

func TestFormatToolStatus(t *testing.T) {
	tests := []struct {
		name     string
		result   ToolCheckResult
		contains []string
	}{
		{
			"installed with version",
			ToolCheckResult{
				Tool:      Tool{DisplayName: "test-tool"},
				Installed: true,
				Version:   "1.0.0",
			},
			[]string{"✅", "test-tool", "1.0.0"},
		},
		{
			"installed without version",
			ToolCheckResult{
				Tool:      Tool{DisplayName: "test-tool"},
				Installed: true,
			},
			[]string{"✅", "test-tool"},
		},
		{
			"not installed",
			ToolCheckResult{
				Tool:      Tool{DisplayName: "test-tool"},
				Installed: false,
			},
			[]string{"❌", "test-tool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatToolStatus(tt.result)

			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("expected output to contain %q, got: %s", substr, output)
				}
			}
		})
	}
}

func TestToolStructure(t *testing.T) {
	// Verify all core tools have required fields
	for _, tool := range CoreTools {
		if tool.Name == "" {
			t.Error("core tool missing Name field")
		}
		if tool.DisplayName == "" {
			t.Error("core tool missing DisplayName field")
		}
		if tool.Category != metadata.ToolCategoryCore {
			t.Errorf("core tool %s has wrong category: %s", tool.Name, tool.Category)
		}
		if tool.Name == "mise" && tool.InstallURL == "" {
			t.Error("mise tool should have InstallURL")
		}
		if tool.Name != "mise" && tool.InstallCmd == "" {
			t.Errorf("non-mise core tool %s should have InstallCmd", tool.Name)
		}
	}

	// Verify all framework tools have required fields
	for _, tool := range FrameworkTools {
		if tool.Name == "" {
			t.Error("framework tool missing Name field")
		}
		if tool.DisplayName == "" {
			t.Error("framework tool missing DisplayName field")
		}
		if tool.Category != metadata.ToolCategoryFramework {
			t.Errorf("framework tool %s has wrong category: %s", tool.Name, tool.Category)
		}
		if tool.InstallCmd == "" {
			t.Errorf("framework tool %s should have InstallCmd", tool.Name)
		}
	}
}
