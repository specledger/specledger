package commands

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/specledger/specledger/pkg/cli/launcher"
	"github.com/specledger/specledger/pkg/cli/metadata"
	"github.com/specledger/specledger/pkg/cli/playbooks"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/specledger/specledger/pkg/embedded"
	"github.com/specledger/specledger/pkg/version"
)

// detectArtifactPath detects the artifact path from existing directories in the project.
// It checks for common artifact directory names and returns the first match.
// If no match is found, returns the default "specledger/".
func detectArtifactPath(projectPath string) string {
	// Common artifact directory names to check, in priority order
	candidatePaths := []string{
		"specledger/",
		"specs/",
		"spec/",
		"docs/",
		"documentation/",
		"api/",
	}

	// Check if any candidate directory exists
	for _, candidate := range candidatePaths {
		candidateDir := filepath.Join(projectPath, strings.TrimSuffix(candidate, "/"))
		if info, err := os.Stat(candidateDir); err == nil && info.IsDir() {
			// Check if directory has any files (not empty)
			entries, err := os.ReadDir(candidateDir)
			if err == nil && len(entries) > 0 {
				return candidate
			}
		}
	}

	// Default if no existing directory found
	return "specledger/"
}

// applyEmbeddedPlaybooks copies embedded playbooks to the project directory.
// If playbookName is empty, uses the default playbook.
// If force is true, existing files will be overwritten.
// Returns the playbook name, version, and structure for metadata storage.
func applyEmbeddedPlaybooks(projectPath string, playbookName string, force bool) (string, string, []string, error) {
	ui.PrintSection("Copying Playbooks")
	fmt.Printf("Applying SpecLedger playbooks...\n")

	pbName, pbVersion, pbStructure, err := playbooks.ApplyToProject(projectPath, playbookName, force)
	if err != nil {
		// Playbooks are helpful but not critical - log warning and continue
		ui.PrintWarning(fmt.Sprintf("Playbook copying failed: %v", err))
		ui.PrintWarning("Project will be created without playbooks")
		return "", "", nil, nil
	}

	fmt.Printf("%s Playbooks applied\n", ui.Checkmark())

	// Trust mise.toml if it exists
	trustMiseConfig(projectPath)

	return pbName, pbVersion, pbStructure, nil
}

// trustMiseConfig runs `mise trust` on the project's mise.toml file.
func trustMiseConfig(projectPath string) {
	misePath := projectPath + "/mise.toml"
	cmd := exec.Command("mise", "trust", misePath)
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		// mise trust failing is not critical - mise will prompt user to trust on first use
		ui.PrintWarning(fmt.Sprintf("Could not trust mise.toml: %v", err))
		ui.PrintWarning("Run 'mise trust' to enable mise tools")
	}
}

// applyEmbeddedSkills copies embedded skills and commands to the project.
// These provide Claude with context for SpecLedger capabilities.
func applyEmbeddedSkills(projectPath string) error {
	// Target directory is .claude in the project root
	targetDir := filepath.Join(projectPath, ".claude")

	// Walk through the skills embedded filesystem
	err := fs.WalkDir(embedded.SkillsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory and skills wrapper
		if path == "." || path == "skills" {
			return nil
		}

		// Skip directories (they'll be created when files are written)
		if d.IsDir() {
			return nil
		}

		// Remove "skills/" prefix to get relative path from commands/ and skills/
		relPath := strings.TrimPrefix(path, "skills/")
		destPath := filepath.Join(targetDir, relPath)

		// Ensure parent directory exists
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destDir, err)
		}

		// Read file from embedded FS
		data, err := embedded.SkillsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// Determine permissions based on file type
		var perms fs.FileMode
		if playbooks.IsExecutableFile(filepath.Base(destPath), data) {
			perms = 0755 // Executable: rwxr-xr-x
		} else {
			perms = 0644 // Regular: rw-r--r--
		}

		// Write to destination with appropriate permissions
		if err := os.WriteFile(destPath, data, perms); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to copy embedded skills: %w", err)
	}

	return nil
}

// setupSpecLedgerProject applies playbooks, skills, and creates metadata.
// Optionally initializes git based on flags.
// If force is true, existing files will be overwritten.
// Returns the playbook name, version, and structure for metadata storage.
func setupSpecLedgerProject(projectPath, projectName, shortCode, playbookName string, initGit bool, force bool) (string, string, []string, error) {
	// Apply embedded playbooks
	selectedPlaybookName, playbookVersion, playbookStructure, err := applyEmbeddedPlaybooks(projectPath, playbookName, force)
	if err != nil {
		// Playbook application failure is not fatal - log warning and continue
		fmt.Printf("Warning: playbook application had issues: %v\n", err)
	}

	// Apply embedded skills
	if err := applyEmbeddedSkills(projectPath); err != nil {
		// Skills are helpful but not critical - log warning and continue
		fmt.Printf("Warning: skills installation had issues: %v\n", err)
	}

	// Create YAML metadata with playbook info
	projectMetadata := metadata.NewProjectMetadata(projectName, shortCode, selectedPlaybookName, playbookVersion, playbookStructure, version.GetVersion())
	if err := metadata.SaveToProject(projectMetadata, projectPath); err != nil {
		return "", "", nil, fmt.Errorf("failed to create project metadata: %w", err)
	}

	// Run post-init script BEFORE git init (so generated files are included)
	runPostInitScript(projectPath, projectMetadata)

	// Initialize git if requested (bootstrap only)
	// This runs AFTER post-init so generated files (like .beads/) are staged
	if initGit {
		if err := initializeGitRepo(projectPath); err != nil {
			return "", "", nil, fmt.Errorf("failed to initialize git: %w", err)
		}
	}

	return selectedPlaybookName, playbookVersion, playbookStructure, nil
}

// ConstitutionPrinciple represents a selectable guiding principle for the project constitution.
type ConstitutionPrinciple struct {
	Name        string
	Description string
	Selected    bool
}

// DefaultPrinciples returns the default set of constitution principles presented during sl new.
func DefaultPrinciples() []ConstitutionPrinciple {
	return []ConstitutionPrinciple{
		{Name: "Specification-First", Description: "Every feature starts with a spec before code", Selected: true},
		{Name: "Test-First", Description: "Tests written before implementation; TDD enforced", Selected: true},
		{Name: "Code Quality", Description: "Consistent formatting, linting, and review standards", Selected: true},
		{Name: "Simplicity", Description: "Start simple; avoid premature abstraction (YAGNI)", Selected: true},
		{Name: "Observability", Description: "Structured logging and metrics for debuggability", Selected: true},
	}
}

// placeholderPattern matches [ALL_CAPS_IDENTIFIER] tokens in the constitution template.
var placeholderPattern = regexp.MustCompile(`\[[A-Z_]{3,}\]`)

// IsConstitutionPopulated checks if the constitution file exists and is populated
// (no placeholder tokens remaining). Returns false if the file is missing, empty,
// or still contains [PLACEHOLDER] style tokens.
func IsConstitutionPopulated(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	if len(content) == 0 {
		return false
	}
	return !placeholderPattern.Match(content)
}

// WriteDefaultConstitution writes a populated constitution file with the given principles
// and agent preference, replacing template placeholders.
func WriteDefaultConstitution(path string, principles []ConstitutionPrinciple, agentPref string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create constitution directory: %w", err)
	}

	// Get project name from path context
	projectName := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(path))))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s Constitution\n\n", projectName))
	sb.WriteString("## Core Principles\n\n")

	for i, p := range principles {
		if !p.Selected {
			continue
		}
		sb.WriteString(fmt.Sprintf("### %s. %s\n", romanNumeral(i+1), p.Name))
		sb.WriteString(p.Description + "\n\n")
	}

	sb.WriteString("## Agent Preferences\n\n")
	sb.WriteString(fmt.Sprintf("- **Preferred Agent**: %s\n\n", agentPref))

	sb.WriteString("## Governance\n\n")
	sb.WriteString("Constitution supersedes all other practices. Amendments require documentation and team approval.\n\n")

	now := time.Now().Format("2006-01-02")
	sb.WriteString(fmt.Sprintf("**Version**: 1.0.0 | **Ratified**: %s | **Last Amended**: %s\n", now, now))

	// #nosec G306 -- constitution file needs to be readable, 0644 is appropriate
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write constitution: %w", err)
	}
	return nil
}

// ReadAgentPreference extracts the preferred agent from a populated constitution file.
// Returns empty string and nil error if the file exists but has no agent preference.
func ReadAgentPreference(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, "**Preferred Agent**:") {
			// Extract value after the label
			parts := strings.SplitN(line, "**Preferred Agent**:", 2)
			if len(parts) == 2 {
				pref := strings.TrimSpace(parts[1])
				// Strip any placeholder tokens
				if placeholderPattern.MatchString(pref) {
					return "", nil
				}
				return pref, nil
			}
		}
	}
	return "", nil
}

// romanNumeral converts 1-10 to Roman numerals for principle numbering.
func romanNumeral(n int) string {
	numerals := []string{"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X"}
	if n > 0 && n < len(numerals) {
		return numerals[n]
	}
	return fmt.Sprintf("%d", n)
}

// launchAgent checks agent availability and launches the selected agent after project setup.
// This is a non-fatal operation — project setup is still complete even if the agent
// cannot be launched.
func launchAgent(projectDir string, agentPref string) error {
	// No agent selected
	if agentPref == "" || agentPref == "None" {
		fmt.Println()
		ui.PrintSuccess("Project setup complete!")
		fmt.Println(ui.Dim("  To start an AI coding session later, run your preferred agent in the project directory."))
		return nil
	}

	// Find the matching agent option
	var agent launcher.AgentOption
	for _, a := range launcher.DefaultAgents {
		if a.Name == agentPref {
			agent = a
			break
		}
	}
	if agent.Command == "" {
		ui.PrintWarning(fmt.Sprintf("Unknown agent '%s'. Skipping agent launch.", agentPref))
		return nil
	}

	al := launcher.NewAgentLauncher(agent, projectDir)

	// Check availability
	if !al.IsAvailable() {
		fmt.Println()
		ui.PrintWarning(fmt.Sprintf("%s is not installed.", agent.Name))
		fmt.Printf("  %s\n", al.InstallInstructions())
		fmt.Println(ui.Dim("  Project setup is complete. You can launch the agent manually after installing."))
		return nil
	}

	// Launch the agent
	fmt.Println()
	ui.PrintSection("Launching " + agent.Name)
	fmt.Println(ui.Dim("  Type /specledger.onboard to start the guided workflow."))
	fmt.Println()

	if err := al.Launch(); err != nil {
		// Agent exit is non-fatal — project is already set up
		ui.PrintWarning(fmt.Sprintf("Agent exited: %v", err))
	}

	return nil
}

// shouldLaunchAgent returns true if agent launch is appropriate in the current environment.
func shouldLaunchAgent() bool {
	// Don't launch in CI environments
	if os.Getenv("CI") == "true" || os.Getenv("CI") == "1" {
		return false
	}
	return true
}

// runPostInitScript executes the template's init.sh script if it exists.
// This allows templates to perform post-initialization tasks like setting up beads.
// Passes specledger.yaml data as environment variables for use in scripts.
// The init.sh script is read from embedded templates (not copied to target project).
func runPostInitScript(projectPath string, projectMetadata *metadata.ProjectMetadata) {
	// Look for init.sh in the embedded templates for the selected playbook
	playbookName := projectMetadata.Playbook.Name
	if playbookName == "" {
		return
	}

	// Path to init.sh in embedded templates
	initScriptPath := filepath.Join("templates", playbookName, "init.sh")

	// Check if init.sh exists in embedded templates
	scriptContent, err := embedded.TemplatesFS.ReadFile(initScriptPath)
	if err != nil {
		// Script doesn't exist in this template, skip silently
		return
	}

	ui.PrintSection("Running Post-Init Script")

	// Write script to a temp file for execution
	tmpFile, err := os.CreateTemp("", "specledger-init-*.sh")
	if err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to create temp file for init script: %v", err))
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(scriptContent); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to write init script: %v", err))
		tmpFile.Close()
		return
	}
	tmpFile.Close()

	// Make the temp file executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to make init script executable: %v", err))
		return
	}

	// Execute the script with environment variables
	// #nosec G204 -- tmpFile.Name() is from os.CreateTemp, safe path
	cmd := exec.Command(tmpFile.Name())
	cmd.Dir = projectPath

	// Set environment variables from specledger.yaml for script use
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SPECLEDGER_PROJECT_ROOT=%s", projectPath),
		fmt.Sprintf("SPECLEDGER_PROJECT_NAME=%s", projectMetadata.Project.Name),
		fmt.Sprintf("SPECLEDGER_PROJECT_SHORT_CODE=%s", projectMetadata.Project.ShortCode),
		fmt.Sprintf("SPECLEDGER_PROJECT_VERSION=%s", projectMetadata.Project.Version),
		fmt.Sprintf("SPECLEDGER_PLAYBOOK_NAME=%s", projectMetadata.Playbook.Name),
		fmt.Sprintf("SPECLEDGER_PLAYBOOK_VERSION=%s", projectMetadata.Playbook.Version),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Post-init script failure is not fatal - log warning and continue
		ui.PrintWarning(fmt.Sprintf("Post-init script had issues: %v", err))
		fmt.Println(string(output))
	} else {
		fmt.Printf("%s Post-init completed\n", ui.Checkmark())
	}
}
