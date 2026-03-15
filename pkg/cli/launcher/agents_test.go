package launcher

import (
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
