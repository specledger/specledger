package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specledger/specledger/internal/agent"
)

func TestNewAgentFromDefinition(t *testing.T) {
	def := agent.Agent{
		Name:           "Test Agent",
		Command:        "test-agent",
		ConfigDir:      ".test",
		InstallCommand: "npm install -g test-agent",
	}

	ag := NewAgentFromDefinition(def)
	if ag.Name() != "Test Agent" {
		t.Errorf("Name() = %q, want %q", ag.Name(), "Test Agent")
	}
	if ag.Command() != "test-agent" {
		t.Errorf("Command() = %q, want %q", ag.Command(), "test-agent")
	}
	if ag.InstallCommand != "npm install -g test-agent" {
		t.Errorf("InstallCommand = %q, want %q", ag.InstallCommand, "npm install -g test-agent")
	}
}

func TestCheckInstalled_NotFound(t *testing.T) {
	def := agent.Agent{
		Name:           "Nonexistent Agent",
		Command:        "nonexistent-agent-cmd-xyz",
		ConfigDir:      ".nonexistent",
		InstallCommand: "npm install -g nonexistent-agent",
	}

	ag := NewAgentFromDefinition(def)
	err := ag.CheckInstalled()
	if err == nil {
		t.Error("CheckInstalled() expected error for nonexistent agent")
	}

	expectedMsg := "'nonexistent-agent-cmd-xyz' not found. Install: npm install -g nonexistent-agent"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestNewLauncherForAgent(t *testing.T) {
	def := agent.Agent{
		Name:           "Claude Code",
		Command:        "claude",
		ConfigDir:      ".claude",
		InstallCommand: "npm install -g @anthropic-ai/claude-code",
	}

	launcher := NewLauncherForAgent(def, "/tmp/test")
	if launcher.Name != "Claude Code" {
		t.Errorf("launcher.Name = %q, want %q", launcher.Name, "Claude Code")
	}
	if launcher.Command != "claude" {
		t.Errorf("launcher.Command = %q, want %q", launcher.Command, "claude")
	}
	if launcher.Dir != "/tmp/test" {
		t.Errorf("launcher.Dir = %q, want %q", launcher.Dir, "/tmp/test")
	}
}

func TestLaunchWithPrompt_EmptyCommand(t *testing.T) {
	l := &AgentLauncher{Command: ""}
	err := l.LaunchWithPrompt("test")
	if err == nil || err.Error() != "no agent command configured" {
		t.Errorf("LaunchWithPrompt() with empty command: got %v, want 'no agent command configured'", err)
	}
}

func TestLaunchWithPromptAndOptions_PreservesConfigEnv(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "env_output")

	l := &AgentLauncher{
		Name:    "Test",
		Command: "/bin/sh",
		Dir:     tmpDir,
		env: map[string]string{
			"TEST_LAUNCH_VAR": "preserved_value",
		},
		flags: []string{"-c"},
	}

	prompt := fmt.Sprintf("echo $TEST_LAUNCH_VAR > %s", outputFile)
	err := l.LaunchWithPromptAndOptions(prompt, LaunchOptions{})
	if err != nil {
		t.Fatalf("LaunchWithPromptAndOptions() error: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	got := strings.TrimSpace(string(data))
	if got != "preserved_value" {
		t.Errorf("config env var not preserved, got %q, want %q", got, "preserved_value")
	}
}

func TestLaunchWithPromptAndOptions_ModelOption(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "env_output")

	l := &AgentLauncher{
		Name:    "Test",
		Command: "/bin/sh",
		Dir:     tmpDir,
		env:     map[string]string{},
		flags:   []string{"-c"},
	}

	prompt := fmt.Sprintf("echo $ANTHROPIC_MODEL > %s", outputFile)
	err := l.LaunchWithPromptAndOptions(prompt, LaunchOptions{Model: "claude-opus-4"})
	if err != nil {
		t.Fatalf("LaunchWithPromptAndOptions() error: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	got := strings.TrimSpace(string(data))
	if got != "claude-opus-4" {
		t.Errorf("ANTHROPIC_MODEL = %q, want %q", got, "claude-opus-4")
	}
}

func TestLaunchWithPromptAndOptions_MaxOutputTokensOption(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "env_output")

	l := &AgentLauncher{
		Name:    "Test",
		Command: "/bin/sh",
		Dir:     tmpDir,
		env:     map[string]string{},
		flags:   []string{"-c"},
	}

	prompt := fmt.Sprintf("echo $CLAUDE_CODE_MAX_OUTPUT_TOKENS > %s", outputFile)
	err := l.LaunchWithPromptAndOptions(prompt, LaunchOptions{MaxOutputTokens: 16384})
	if err != nil {
		t.Fatalf("LaunchWithPromptAndOptions() error: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	got := strings.TrimSpace(string(data))
	if got != "16384" {
		t.Errorf("CLAUDE_CODE_MAX_OUTPUT_TOKENS = %q, want %q", got, "16384")
	}
}
