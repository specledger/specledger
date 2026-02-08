package prerequisites

import (
	"fmt"
	"os/exec"
	"strings"

	"specledger/pkg/cli/metadata"
)

// Tool represents a required or optional tool
type Tool struct {
	Name        string
	DisplayName string
	Category    metadata.ToolCategory
	VersionFlag string // Flag to get version (usually --version or -v)
	InstallCmd  string // Command to install via mise
	InstallURL  string // URL for manual installation instructions
}

// ToolCheckResult contains the result of checking a tool's installation status
type ToolCheckResult struct {
	Tool      Tool
	Installed bool
	Version   string
	Path      string
	Error     error
}

var (
	// CoreTools are required for SpecLedger to function
	CoreTools = []Tool{
		{
			Name:        "mise",
			DisplayName: "mise (version manager)",
			Category:    metadata.ToolCategoryCore,
			VersionFlag: "--version",
			InstallURL:  "https://mise.jdx.dev/getting-started.html",
		},
		{
			Name:        "bd",
			DisplayName: "beads (issue tracker)",
			Category:    metadata.ToolCategoryCore,
			VersionFlag: "--version",
			InstallCmd:  "mise install ubi:steveyegge/beads@0.28.0",
		},
		{
			Name:        "perles",
			DisplayName: "perles (workflow tool)",
			Category:    metadata.ToolCategoryCore,
			VersionFlag: "--version",
			InstallCmd:  "mise install ubi:zjrosen/perles@0.2.11",
		},
	}

	// FrameworkTools are optional SDD framework tools
	FrameworkTools = []Tool{
		{
			Name:        "specify",
			DisplayName: "specify (Spec Kit framework)",
			Category:    metadata.ToolCategoryFramework,
			VersionFlag: "--version",
			InstallCmd:  "mise install pipx:git+https://github.com/github/spec-kit.git",
		},
		{
			Name:        "openspec",
			DisplayName: "openspec (OpenSpec framework)",
			Category:    metadata.ToolCategoryFramework,
			VersionFlag: "--version",
			InstallCmd:  "mise install npm:@fission-ai/openspec",
		},
	}
)

// CheckTool checks if a tool is installed and gets its version
func CheckTool(tool Tool) ToolCheckResult {
	result := ToolCheckResult{
		Tool:      tool,
		Installed: false,
	}

	// Check if tool exists in PATH
	path, err := exec.LookPath(tool.Name)
	if err != nil {
		result.Error = err
		return result
	}

	result.Installed = true
	result.Path = path

	// Try to get version
	if tool.VersionFlag != "" {
		cmd := exec.Command(tool.Name, tool.VersionFlag)
		output, err := cmd.CombinedOutput()
		if err == nil {
			// Extract version from output (usually first line)
			version := strings.TrimSpace(string(output))
			lines := strings.Split(version, "\n")
			if len(lines) > 0 {
				result.Version = strings.TrimSpace(lines[0])
			}
		}
	}

	return result
}

// CheckAllTools checks both core and framework tools
func CheckAllTools() ([]ToolCheckResult, []ToolCheckResult) {
	coreResults := make([]ToolCheckResult, len(CoreTools))
	for i, tool := range CoreTools {
		coreResults[i] = CheckTool(tool)
	}

	frameworkResults := make([]ToolCheckResult, len(FrameworkTools))
	for i, tool := range FrameworkTools {
		frameworkResults[i] = CheckTool(tool)
	}

	return coreResults, frameworkResults
}

// CheckCoreTools checks only core required tools
func CheckCoreTools() []ToolCheckResult {
	results := make([]ToolCheckResult, len(CoreTools))
	for i, tool := range CoreTools {
		results[i] = CheckTool(tool)
	}
	return results
}

// AllCoreToolsInstalled returns true if all core tools are installed
func AllCoreToolsInstalled() bool {
	results := CheckCoreTools()
	for _, result := range results {
		if !result.Installed {
			return false
		}
	}
	return true
}

// GetMissingCoreTools returns a list of missing core tools
func GetMissingCoreTools() []Tool {
	results := CheckCoreTools()
	missing := []Tool{}
	for _, result := range results {
		if !result.Installed {
			missing = append(missing, result.Tool)
		}
	}
	return missing
}

// IsMiseInstalled checks specifically if mise is installed
func IsMiseInstalled() bool {
	_, err := exec.LookPath("mise")
	return err == nil
}

// IsCommandAvailable checks if any command is available in PATH
func IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// GetInstallInstructions returns human-readable install instructions for missing tools
func GetInstallInstructions(missing []Tool) string {
	if len(missing) == 0 {
		return "All required tools are installed!"
	}

	var sb strings.Builder
	sb.WriteString("Missing required tools:\n\n")

	for _, tool := range missing {
		sb.WriteString(fmt.Sprintf("  • %s\n", tool.DisplayName))
		if tool.Name == "mise" {
			// Special case: mise must be installed first
			sb.WriteString(fmt.Sprintf("    Install: %s\n", tool.InstallURL))
		} else if tool.InstallCmd != "" {
			sb.WriteString(fmt.Sprintf("    Install: %s\n", tool.InstallCmd))
		} else if tool.InstallURL != "" {
			sb.WriteString(fmt.Sprintf("    Install: %s\n", tool.InstallURL))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// PrerequisiteCheck represents the result of a prerequisite check
type PrerequisiteCheck struct {
	AllCoreInstalled bool
	MissingCore      []Tool
	CoreResults      []ToolCheckResult
	FrameworkResults []ToolCheckResult
	Instructions     string
}

// CheckPrerequisites performs a comprehensive check of all tools
func CheckPrerequisites() PrerequisiteCheck {
	coreResults, frameworkResults := CheckAllTools()

	missingCore := []Tool{}
	for _, result := range coreResults {
		if !result.Installed {
			missingCore = append(missingCore, result.Tool)
		}
	}

	return PrerequisiteCheck{
		AllCoreInstalled: len(missingCore) == 0,
		MissingCore:      missingCore,
		CoreResults:      coreResults,
		FrameworkResults: frameworkResults,
		Instructions:     GetInstallInstructions(missingCore),
	}
}

// EnsurePrerequisites checks for prerequisites and optionally prompts for installation
// If interactive is false (CI mode), it just reports status without prompting
func EnsurePrerequisites(interactive bool) error {
	check := CheckPrerequisites()

	if check.AllCoreInstalled {
		return nil
	}

	if !interactive {
		// CI mode: just return error with instructions
		return fmt.Errorf("missing required tools:\n%s", check.Instructions)
	}

	// Interactive mode: display missing tools
	fmt.Println("⚠️  Missing required tools detected:")
	fmt.Println()

	for _, tool := range check.MissingCore {
		fmt.Printf("  ✗ %s\n", tool.DisplayName)
	}
	fmt.Println()

	// Check if mise is missing - it's needed to install other tools
	miseInstalled := IsMiseInstalled()
	if !miseInstalled {
		fmt.Println("mise is required to install other tools.")
		fmt.Printf("Please install mise first: %s\n", CoreTools[0].InstallURL)
		fmt.Println()
		return fmt.Errorf("mise not installed")
	}

	// If mise is installed, we can offer to install other tools
	fmt.Println("You can install missing tools with:")
	fmt.Println()
	for _, tool := range check.MissingCore {
		if tool.Name != "mise" && tool.InstallCmd != "" {
			fmt.Printf("  %s\n", tool.InstallCmd)
		}
	}
	fmt.Println()

	return fmt.Errorf("missing required tools")
}

// FormatToolStatus formats a tool check result for display
func FormatToolStatus(result ToolCheckResult) string {
	if result.Installed {
		status := "✅"
		if result.Version != "" {
			return fmt.Sprintf("%s %s (%s)", status, result.Tool.DisplayName, result.Version)
		}
		return fmt.Sprintf("%s %s", status, result.Tool.DisplayName)
	}
	return fmt.Sprintf("❌ %s", result.Tool.DisplayName)
}
