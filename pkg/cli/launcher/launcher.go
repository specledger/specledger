package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AgentOption represents an available AI coding agent choice.
type AgentOption struct {
	Name        string // Display name (e.g., "Claude Code")
	Command     string // CLI command (e.g., "claude")
	Description string // Short description for TUI
}

// DefaultAgents is the list of agent options presented during onboarding.
var DefaultAgents = []AgentOption{
	{
		Name:        "Claude Code",
		Command:     "claude",
		Description: "AI coding assistant with deep SpecLedger integration",
	},
	{
		Name:        "OpenCode",
		Command:     "opencode",
		Description: "Open-source AI coding assistant",
	},
	{
		Name:        "Copilot CLI",
		Command:     "github-copilot",
		Description: "GitHub Copilot command-line interface",
	},
	{
		Name:        "Codex",
		Command:     "codex",
		Description: "OpenAI Codex coding assistant",
	},
	{
		Name:        "None",
		Command:     "",
		Description: "Skip agent launch; use SpecLedger manually",
	},
}

// AgentLauncher checks availability and launches an AI coding agent
// as an interactive subprocess.
type AgentLauncher struct {
	Name    string            // Agent display name
	Command string            // CLI command to execute
	Dir     string            // Working directory for the agent process
	env     map[string]string // Environment variables to inject
	flags   []string          // CLI flags to pass to the command
}

// NewAgentLauncher creates a launcher for the given agent in the given directory.
func NewAgentLauncher(agent AgentOption, dir string) *AgentLauncher {
	return &AgentLauncher{
		Name:    agent.Name,
		Command: agent.Command,
		Dir:     dir,
		env:     make(map[string]string),
		flags:   []string{},
	}
}

func (l *AgentLauncher) SetEnv(envVars map[string]string) {
	if l.env == nil {
		l.env = make(map[string]string)
	}
	for k, v := range envVars {
		l.env[k] = v
	}
}

// SetFlags sets the CLI flags to pass when launching the agent.
func (l *AgentLauncher) SetFlags(flags []string) {
	l.flags = flags
}

func (l *AgentLauncher) BuildEnv() []string {
	// Start with current environment
	envMap := make(map[string]string)
	for _, entry := range os.Environ() {
		if idx := strings.Index(entry, "="); idx > 0 {
			envMap[entry[:idx]] = entry[idx+1:]
		}
	}

	// Override with launcher's env vars (these take precedence)
	for k, v := range l.env {
		envMap[k] = v
	}

	// Convert back to slice
	result := make([]string, 0, len(envMap))
	for k, v := range envMap {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// IsAvailable checks if the agent command exists in PATH.
func (l *AgentLauncher) IsAvailable() bool {
	if l.Command == "" {
		return false
	}
	_, err := exec.LookPath(l.Command)
	return err == nil
}

// Launch starts the agent as an interactive subprocess with stdio passthrough.
// This blocks until the agent process exits.
func (l *AgentLauncher) Launch() error {
	if l.Command == "" {
		return fmt.Errorf("no agent command configured")
	}

	args := l.flags
	// #nosec G204 -- l.Command is from a controlled DefaultAgents list, not user input
	cmd := exec.Command(l.Command, args...)
	cmd.Dir = l.Dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = l.BuildEnv()

	return cmd.Run()
}

// LaunchWithPrompt starts the agent as an interactive subprocess and passes prompt as a
// positional argument. Using a positional arg (not stdin) preserves TTY interactivity.
// This blocks until the agent process exits.
func (l *AgentLauncher) LaunchWithPrompt(prompt string) error {
	return l.LaunchWithPromptAndOptions(prompt, LaunchOptions{})
}

// LaunchOptions configures agent launch behavior.
type LaunchOptions struct {
	SkipPermissions bool   // Add --dangerously-skip-permissions flag
	Model           string // Set ANTHROPIC_MODEL env var (empty = use default)
	MaxOutputTokens int    // Set CLAUDE_CODE_MAX_OUTPUT_TOKENS env var (0 = use default)
}

// LaunchWithPromptAndOptions starts the agent with custom options.
func (l *AgentLauncher) LaunchWithPromptAndOptions(prompt string, opts LaunchOptions) error {
	if l.Command == "" {
		return fmt.Errorf("no agent command configured")
	}

	// Combine flags with prompt as final argument
	args := append(l.flags, prompt)
	// #nosec G204 -- l.Command is from a controlled DefaultAgents list, prompt is internal
	cmd := exec.Command(l.Command, args...)
	cmd.Dir = l.Dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = l.BuildEnv()

	// Set environment variables for agent
	cmd.Env = os.Environ()
	if opts.Model != "" {
		cmd.Env = append(cmd.Env, "ANTHROPIC_MODEL="+opts.Model)
	}
	if opts.MaxOutputTokens > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("CLAUDE_CODE_MAX_OUTPUT_TOKENS=%d", opts.MaxOutputTokens))
	}

	return cmd.Run()
}

// InstallInstructions returns help text for installing the agent.
func (l *AgentLauncher) InstallInstructions() string {
	switch l.Command {
	case "claude":
		return "Install Claude Code: npm install -g @anthropic-ai/claude-code"
	default:
		return fmt.Sprintf("Install %s and ensure '%s' is available in your PATH", l.Name, l.Command)
	}
}
