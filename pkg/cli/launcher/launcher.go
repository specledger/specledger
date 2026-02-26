package launcher

import (
	"fmt"
	"os"
	"os/exec"
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
}

// NewAgentLauncher creates a launcher for the given agent in the given directory.
func NewAgentLauncher(agent AgentOption, dir string) *AgentLauncher {
	return &AgentLauncher{
		Name:    agent.Name,
		Command: agent.Command,
		Dir:     dir,
		env:     make(map[string]string),
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

func (l *AgentLauncher) BuildEnv() []string {
	result := os.Environ()
	for k, v := range l.env {
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

	// #nosec G204 -- l.Command is from a controlled DefaultAgents list, not user input
	cmd := exec.Command(l.Command)
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
	if l.Command == "" {
		return fmt.Errorf("no agent command configured")
	}

	// #nosec G204 -- l.Command is from a controlled DefaultAgents list, prompt is internal
	cmd := exec.Command(l.Command, prompt)
	cmd.Dir = l.Dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = l.BuildEnv()

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
