package models

import (
	"fmt"
	"strings"
)

// AgentConfig represents the configuration for an AI coding agent.
// Supports Claude Code, OpenCode, or None (manual development).
type AgentConfig struct {
	// ID is the unique agent identifier in kebab-case (e.g., "claude-code", "opencode")
	ID string

	// Name is the human-readable agent name (e.g., "Claude Code")
	Name string

	// Description is a one-line description of the agent
	Description string

	// ConfigDir is the configuration directory name (e.g., ".claude", ".opencode")
	// Empty string for "None" agent
	ConfigDir string
}

// Validate checks if the AgentConfig has valid field values.
// Returns an error if any validation rule fails.
func (a *AgentConfig) Validate() error {
	// Validate ID format (kebab-case)
	if !kebabCasePattern.MatchString(a.ID) && a.ID != "none" {
		return fmt.Errorf("agent ID must be kebab-case or 'none': got %q", a.ID)
	}

	// Validate name length (1-50 characters)
	if len(a.Name) == 0 || len(a.Name) > 50 {
		return fmt.Errorf("agent name must be 1-50 characters: got %d", len(a.Name))
	}

	// Validate description length (1-200 characters)
	if len(a.Description) == 0 || len(a.Description) > 200 {
		return fmt.Errorf("agent description must be 1-200 characters: got %d", len(a.Description))
	}

	// Validate ConfigDir format if not empty
	if a.ConfigDir != "" && !strings.HasPrefix(a.ConfigDir, ".") {
		return fmt.Errorf("agent config directory must start with '.' or be empty: got %q", a.ConfigDir)
	}

	return nil
}

// HasConfig returns true if the agent has a configuration directory.
func (a *AgentConfig) HasConfig() bool {
	return a.ConfigDir != ""
}

// SupportedAgents returns the list of all supported coding agents.
func SupportedAgents() []AgentConfig {
	return []AgentConfig{
		{
			ID:          "claude-code",
			Name:        "Claude Code",
			Description: "Anthropic's official CLI for Claude with session capture",
			ConfigDir:   ".claude",
		},
		{
			ID:          "opencode",
			Name:        "OpenCode",
			Description: "Open source coding agent compatible with Claude Code",
			ConfigDir:   ".opencode",
		},
		{
			ID:          "none",
			Name:        "None",
			Description: "No AI agent - manual development only",
			ConfigDir:   "",
		},
	}
}

// GetAgentByID finds an agent configuration by ID.
// Returns the agent config and nil error if found, or nil and an error if not found.
func GetAgentByID(id string) (*AgentConfig, error) {
	agents := SupportedAgents()
	for i := range agents {
		if agents[i].ID == id {
			return &agents[i], nil
		}
	}

	// Build list of valid IDs for error message
	var validIDs []string
	for _, a := range agents {
		validIDs = append(validIDs, a.ID)
	}

	return nil, fmt.Errorf("unknown agent ID: %q (valid: %s)", id, strings.Join(validIDs, ", "))
}

// DefaultAgent returns the default coding agent (Claude Code).
func DefaultAgent() AgentConfig {
	return SupportedAgents()[0] // claude-code is first
}
