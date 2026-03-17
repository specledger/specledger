package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/specledger/specledger/internal/agent"
	"github.com/specledger/specledger/pkg/cli/auth"
	"github.com/specledger/specledger/pkg/cli/config"
	"github.com/specledger/specledger/pkg/cli/launcher"
	"github.com/specledger/specledger/pkg/cli/metadata"
	"github.com/specledger/specledger/pkg/cli/playbooks"
	"github.com/specledger/specledger/pkg/cli/revise"
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
// If agentTargetDir is set, commands and skills are copied there instead of .claude/
// Returns the playbook name, version, and structure for metadata storage.
func applyEmbeddedPlaybooks(projectPath string, playbookName string, force bool, agentTargetDir string) (string, string, []string, error) {
	ui.PrintSection("Copying Playbooks")
	fmt.Printf("Applying SpecLedger playbooks...\n")

	var pbName, pbVersion string
	var pbStructure []string
	var err error

	if agentTargetDir != "" {
		pbName, pbVersion, pbStructure, err = playbooks.ApplyToProjectWithAgentTarget(projectPath, playbookName, force, agentTargetDir)
	} else {
		pbName, pbVersion, pbStructure, err = playbooks.ApplyToProject(projectPath, playbookName, force)
	}

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

// setupSpecLedgerProject applies playbooks and creates metadata.
// Optionally initializes git based on flags.
// If force is true, existing files will be overwritten and agent symlinks will be fixed.
// selectedAgents is a comma-separated list of agent names to configure (e.g., "claude,opencode").
// Returns the playbook name, version, and structure for metadata storage.
func setupSpecLedgerProject(projectPath, projectName, shortCode, playbookName string, initGit bool, force bool, selectedAgents string) (string, string, []string, error) {
	// Determine agent target directory - use .agents if multiple agents selected
	agentTargetDir := ""
	if selectedAgents != "" && selectedAgents != "None" {
		agentTargetDir = ".agents"
	}

	// Apply embedded playbooks (commands/skills go to agentTargetDir if set)
	// When force=true, templates in .agents/ will be overwritten
	selectedPlaybookName, playbookVersion, playbookStructure, err := applyEmbeddedPlaybooks(projectPath, playbookName, force, agentTargetDir)
	if err != nil {
		// Playbook application failure is not fatal - log warning and continue
		fmt.Printf("Warning: playbook application had issues: %v\n", err)
	}

	// Setup multi-agent shared directories if agents selected
	if selectedAgents != "" && selectedAgents != "None" {
		agentNames := strings.Split(selectedAgents, ",")
		for i, name := range agentNames {
			agentNames[i] = strings.TrimSpace(name)
		}

		// Create .agents/commands and .agents/skills directories (if not already created by playbook)
		if err := playbooks.CreateAgentSharedDir(projectPath, force); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				ui.PrintWarning(fmt.Sprintf("Failed to create .agents directory: %v", err))
			}
		} else {
			fmt.Printf("%s Created .agents/ directory\n", ui.Checkmark())
		}

		// Link each selected agent to the shared directories
		// This will fix broken/incorrect symlinks without deleting .agents/ content
		if err := playbooks.LinkAgentToShared(projectPath, agentNames, force); err != nil {
			ui.PrintWarning(fmt.Sprintf("Failed to link agents: %v", err))
		} else {
			fmt.Printf("%s Linked agents: %s\n", ui.Checkmark(), strings.Join(agentNames, ", "))
		}
	}

	// Create YAML metadata with playbook info
	projectMetadata := metadata.NewProjectMetadata(projectName, shortCode, selectedPlaybookName, playbookVersion, playbookStructure, version.GetVersion())
	if err := metadata.SaveToProject(projectMetadata, projectPath); err != nil {
		return "", "", nil, fmt.Errorf("failed to create project metadata: %w", err)
	}

	// Run post-init script BEFORE git init (so generated files are included)
	runPostInitScript(projectPath, projectMetadata)

	// Initialize git if requested (bootstrap only)
	// This runs AFTER post-init so generated files are staged
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
func WriteDefaultConstitution(path string, principles []ConstitutionPrinciple, agentPref string, selectedAgents []string) error {
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

	if len(selectedAgents) > 0 {
		sb.WriteString("## Selected Agents\n\n")
		for _, ag := range selectedAgents {
			sb.WriteString(fmt.Sprintf("- %s\n", ag))
		}
		sb.WriteString("\n")
	}

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
			parts := strings.SplitN(line, "**Preferred Agent**:", 2)
			if len(parts) == 2 {
				pref := strings.TrimSpace(parts[1])
				if placeholderPattern.MatchString(pref) {
					return "", nil
				}
				return pref, nil
			}
		}
	}
	return "", nil
}

// ReadSelectedAgents extracts the selected agents from a populated constitution file.
// Returns empty slice and nil error if the file exists but has no selected agents.
func ReadSelectedAgents(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var agents []string
	inSelectedAgentsSection := false

	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, "## Selected Agents") {
			inSelectedAgentsSection = true
			continue
		}
		if inSelectedAgentsSection {
			if strings.HasPrefix(line, "## ") {
				break
			}
			if strings.HasPrefix(line, "- ") {
				agent := strings.TrimSpace(strings.TrimPrefix(line, "- "))
				if agent != "" && !placeholderPattern.MatchString(agent) {
					agents = append(agents, agent)
				}
			}
		}
	}
	return agents, nil
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
	if agentPref == "" || agentPref == "None" {
		fmt.Println()
		ui.PrintSuccess("Project setup complete!")
		fmt.Println(ui.Dim("  To start an AI coding session later, run your preferred agent in the project directory."))
		return nil
	}

	// Parse comma-separated agent preference (TUI may return "Claude Code,OpenCode")
	// Use the first agent in the list as the primary agent to launch
	primaryAgent := strings.TrimSpace(strings.Split(agentPref, ",")[0])

	// Look up agent using the registry (supports both display name and command)
	ag, found := agent.Lookup(primaryAgent)
	if !found {
		ui.PrintWarning(fmt.Sprintf("Unknown agent '%s'. Skipping agent launch.", primaryAgent))
		return nil
	}

	// Create launcher using agent definition
	l := launcher.NewLauncherForAgent(ag, projectDir)

	// Check if agent is installed
	wrappedAgent := launcher.NewAgentFromDefinition(ag)
	if err := wrappedAgent.CheckInstalled(); err != nil {
		fmt.Println()
		ui.PrintWarning(fmt.Sprintf("%s is not installed.", ag.Name))
		fmt.Printf("  Install: %s\n", ag.InstallCommand)
		fmt.Println(ui.Dim("  Project setup is complete. You can launch the agent manually after installing."))
		return nil
	}

	// Get agent settings from config (using same logic as 'sl code')
	settings := config.ResolveAgentSettings(ag.Command)
	if settings != nil {
		envVars := make(map[string]string)

		// Set arguments
		if len(settings.Arguments) > 0 {
			l.SetFlags(settings.Arguments)
		}

		// Map API key to agent's env var
		if settings.APIKey != "" && ag.APIKeyEnvVar != "" {
			envVars[ag.APIKeyEnvVar] = settings.APIKey
		}

		// Map base URL to agent's env var
		if settings.BaseURL != "" && ag.BaseURLEnvVar != "" {
			envVars[ag.BaseURLEnvVar] = settings.BaseURL
		}

		// Map model to agent's env var
		if settings.Model != "" && ag.ModelEnvVar != "" {
			envVars[ag.ModelEnvVar] = settings.Model
		}

		// Add per-agent env vars
		for k, v := range settings.EnvVars {
			envVars[k] = v
		}

		// Claude-specific: map model aliases to env vars
		if ag.Command == "claude" {
			if v, ok := settings.ModelAliases["sonnet"]; ok && v != "" {
				envVars["ANTHROPIC_DEFAULT_SONNET_MODEL"] = v
			}
			if v, ok := settings.ModelAliases["opus"]; ok && v != "" {
				envVars["ANTHROPIC_DEFAULT_OPUS_MODEL"] = v
			}
			if v, ok := settings.ModelAliases["haiku"]; ok && v != "" {
				envVars["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = v
			}
		}

		if len(envVars) > 0 {
			l.SetEnv(envVars)
		}
	}

	fmt.Println()
	ui.PrintSection("Launching " + ag.Name)
	fmt.Println(ui.Dim("  Type /specledger.onboard to start the guided workflow."))
	fmt.Println()

	if err := l.Launch(); err != nil {
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

// GitRemoteInfo contains parsed git remote information
type GitRemoteInfo struct {
	Owner string
	Name  string
}

// detectGitRemote parses the git origin remote URL to extract owner and repo name.
// Supports both HTTPS and SSH formats:
//   - https://github.com/owner/repo.git
//   - git@github.com:owner/repo.git
func detectGitRemote(projectPath string) (*GitRemoteInfo, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("no git remote 'origin' found (run: git remote add origin <url>)")
	}

	remoteURL := strings.TrimSpace(string(output))
	return parseGitRemoteURL(remoteURL)
}

// parseGitRemoteURL parses a git URL (HTTPS or SSH) to extract owner and repo name.
func parseGitRemoteURL(remoteURL string) (*GitRemoteInfo, error) {
	// SSH format: git@github.com:owner/repo.git
	sshPattern := regexp.MustCompile(`git@[^:]+:([^/]+)/(.+?)(?:\.git)?$`)
	if matches := sshPattern.FindStringSubmatch(remoteURL); len(matches) == 3 {
		return &GitRemoteInfo{
			Owner: matches[1],
			Name:  strings.TrimSuffix(matches[2], ".git"),
		}, nil
	}

	// HTTPS format: https://github.com/owner/repo.git
	httpsPattern := regexp.MustCompile(`https?://[^/]+/([^/]+)/(.+?)(?:\.git)?$`)
	if matches := httpsPattern.FindStringSubmatch(remoteURL); len(matches) == 3 {
		return &GitRemoteInfo{
			Owner: matches[1],
			Name:  strings.TrimSuffix(matches[2], ".git"),
		}, nil
	}

	return nil, fmt.Errorf("could not parse git remote URL: %s", remoteURL)
}

// lookupProjectID attempts to find the project ID from Supabase using git remote info.
// Returns empty string if lookup fails (no authentication, no remote, project not found).
// This is a non-fatal lookup - callers should gracefully handle empty results.
func lookupProjectID(projectPath string) string {
	// 1. Detect git remote to get owner/repo
	remoteInfo, err := detectGitRemote(projectPath)
	if err != nil {
		return "" // No git remote, skip silently
	}

	// 2. Check authentication
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return "" // Not authenticated, skip silently
	}

	// 3. Lookup project in Supabase via revise client
	client := revise.NewReviseClient(accessToken)
	project, err := client.GetProject(remoteInfo.Owner, remoteInfo.Name)
	if err != nil {
		return "" // Project not found in Supabase, skip silently
	}

	return project.ID
}

// runPostInitScript executes the template's init.sh script if it exists.
// This allows templates to perform post-initialization tasks.
// Passes specledger.yaml data as environment variables for use in scripts.
// The init.sh script is read from embedded templates (not copied to target project).
func runPostInitScript(projectPath string, projectMetadata *metadata.ProjectMetadata) {
	// Look for init.sh in the embedded templates for the selected playbook
	playbookName := projectMetadata.Playbook.Name
	if playbookName == "" {
		return
	}

	// Path to init.sh in embedded templates (must use forward slashes for embed.FS)
	initScriptPath := path.Join("templates", playbookName, "init.sh")

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

	// Execute the script with environment variables.
	// On Windows, .sh files cannot be run directly — find a Unix shell interpreter.
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		shell := findWindowsShell()
		if shell == "" {
			// No Unix shell available on Windows — skip post-init script gracefully
			return
		}
		cmd = exec.Command(shell, tmpFile.Name()) // #nosec G204 -- shell from LookPath, tmpFile from os.CreateTemp
	} else {
		cmd = exec.Command(tmpFile.Name()) // #nosec G204 -- tmpFile from os.CreateTemp, safe path
	}
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

// findWindowsShell looks for a Unix shell (bash or sh) on Windows,
// as shipped by Git for Windows or similar tools.
// Returns the path to the shell, or empty string if none found.
func findWindowsShell() string {
	for _, shell := range []string{"bash", "sh"} {
		if p, err := exec.LookPath(shell); err == nil {
			return p
		}
	}
	return ""
}
