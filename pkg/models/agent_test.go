package models

import (
	"testing"
)

func TestSupportedAgents(t *testing.T) {
	agents := SupportedAgents()

	// Verify we have 3 agents
	if len(agents) != 3 {
		t.Errorf("SupportedAgents() returned %d agents, want 3", len(agents))
	}

	// Verify expected agent IDs
	expectedIDs := map[string]bool{
		"claude-code": false,
		"opencode":    false,
		"none":        false,
	}

	for _, agent := range agents {
		if _, ok := expectedIDs[agent.ID]; !ok {
			t.Errorf("Unexpected agent ID: %q", agent.ID)
		}
		expectedIDs[agent.ID] = true
	}

	for id, found := range expectedIDs {
		if !found {
			t.Errorf("Missing expected agent ID: %q", id)
		}
	}
}

func TestGetAgentByID(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
	}{
		{"claude-code", false},
		{"opencode", false},
		{"none", false},
		{"unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			agent, err := GetAgentByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAgentByID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}
			if !tt.wantErr && agent.ID != tt.id {
				t.Errorf("GetAgentByID(%q) returned agent with ID = %q", tt.id, agent.ID)
			}
		})
	}
}

func TestAgentHasConfig(t *testing.T) {
	tests := []struct {
		id        string
		hasConfig bool
	}{
		{"claude-code", true},
		{"opencode", true},
		{"none", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			agent, err := GetAgentByID(tt.id)
			if err != nil {
				t.Fatalf("GetAgentByID(%q) error = %v", tt.id, err)
			}
			if agent.HasConfig() != tt.hasConfig {
				t.Errorf("Agent %q HasConfig() = %v, want %v", tt.id, agent.HasConfig(), tt.hasConfig)
			}
		})
	}
}

func TestDefaultAgent(t *testing.T) {
	agent := DefaultAgent()

	// Default agent should be claude-code
	if agent.ID != "claude-code" {
		t.Errorf("DefaultAgent() ID = %q, want %q", agent.ID, "claude-code")
	}

	if !agent.HasConfig() {
		t.Error("DefaultAgent() should have config")
	}
}

func TestAgentValidation(t *testing.T) {
	for _, agent := range SupportedAgents() {
		if err := agent.Validate(); err != nil {
			t.Errorf("Agent %q validation failed: %v", agent.ID, err)
		}
	}
}
