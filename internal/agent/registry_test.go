package agent

import (
	"testing"
)

func TestLookup(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string
		wantFound bool
	}{
		{"exact match", "claude", "Claude Code", true},
		{"case insensitive", "CLAUDE", "Claude Code", true},
		{"by command", "opencode", "OpenCode", true},
		{"copilot command", "github-copilot", "Copilot CLI", true},
		{"codex", "codex", "Codex", true},
		{"not found", "unknown", "", false},
		{"empty string", "", "", false},
		{"whitespace", "  claude  ", "Claude Code", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := Lookup(tt.input)
			if found != tt.wantFound {
				t.Errorf("Lookup(%q) found = %v, want %v", tt.input, found, tt.wantFound)
			}
			if found && got.Name != tt.wantName {
				t.Errorf("Lookup(%q) Name = %q, want %q", tt.input, got.Name, tt.wantName)
			}
		})
	}
}

func TestAll(t *testing.T) {
	agents := All()
	if len(agents) != 4 {
		t.Errorf("All() returned %d agents, want 4", len(agents))
	}

	names := make(map[string]bool)
	for _, a := range agents {
		names[a.Name] = true
	}

	expected := []string{"Claude Code", "OpenCode", "Copilot CLI", "Codex"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("All() missing agent %q", name)
		}
	}
}

func TestRegistryInstallCommands(t *testing.T) {
	tests := []struct {
		command     string
		wantInstall string
	}{
		{"claude", "npm install -g @anthropic-ai/claude-code"},
		{"opencode", "go install github.com/opencode-ai/opencode@latest"},
		{"github-copilot", "npm install -g @github/copilot"},
		{"codex", "npm install -g @openai/codex"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			agent, found := Lookup(tt.command)
			if !found {
				t.Fatalf("Lookup(%q) not found", tt.command)
			}
			if agent.InstallCommand != tt.wantInstall {
				t.Errorf("InstallCommand = %q, want %q", agent.InstallCommand, tt.wantInstall)
			}
		})
	}
}

func TestAgentConfigDirs(t *testing.T) {
	tests := []struct {
		command   string
		configDir string
	}{
		{"claude", ".claude"},
		{"opencode", ".opencode"},
		{"github-copilot", ".github"},
		{"codex", ".codex"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			agent, found := Lookup(tt.command)
			if !found {
				t.Fatalf("Lookup(%q) not found", tt.command)
			}
			if agent.ConfigDir != tt.configDir {
				t.Errorf("ConfigDir = %q, want %q", agent.ConfigDir, tt.configDir)
			}
		})
	}
}
