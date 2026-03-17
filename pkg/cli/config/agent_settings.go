package config

// AgentSettings represents per-agent configuration.
// This is the universal config structure that applies to all agents.
type AgentSettings struct {
	APIKey    string            `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	BaseURL   string            `yaml:"base_url,omitempty" json:"base_url,omitempty"`
	Model     string            `yaml:"model,omitempty" json:"model,omitempty"`
	Arguments []string          `yaml:"arguments,omitempty" json:"arguments,omitempty"`
	Env       map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

// ClaudeModelAliases defines Claude-specific model aliases.
// These map short names (sonnet, opus, haiku) to full model IDs.
type ClaudeModelAliases struct {
	Sonnet string `yaml:"sonnet,omitempty" json:"sonnet,omitempty"`
	Opus   string `yaml:"opus,omitempty" json:"opus,omitempty"`
	Haiku  string `yaml:"haiku,omitempty" json:"haiku,omitempty"`
}

// ClaudeSettings extends AgentSettings with Claude-specific options.
type ClaudeSettings struct {
	AgentSettings `yaml:",inline" json:",inline"`
	ModelAliases  *ClaudeModelAliases `yaml:"model_aliases,omitempty" json:"model_aliases,omitempty"`
}

// ConfigAgents holds all agent configurations with namespacing.
// This replaces the old single AgentConfig with per-agent settings.
type ConfigAgents struct {
	// Default agent to launch when no agent is specified
	Default string `yaml:"default,omitempty" json:"default,omitempty"`

	// Claude-specific settings (has extra fields for model aliases)
	Claude *ClaudeSettings `yaml:"claude,omitempty" json:"claude,omitempty"`

	// Other agents use the base AgentSettings
	OpenCode      *AgentSettings `yaml:"opencode,omitempty" json:"opencode,omitempty"`
	GitHubCopilot *AgentSettings `yaml:"github-copilot,omitempty" json:"github-copilot,omitempty"`
	Codex         *AgentSettings `yaml:"codex,omitempty" json:"codex,omitempty"`
}

// NewAgentSettings creates a new AgentSettings with initialized maps.
func NewAgentSettings() *AgentSettings {
	return &AgentSettings{
		Env: make(map[string]string),
	}
}

// NewClaudeSettings creates a new ClaudeSettings with initialized nested structs.
func NewClaudeSettings() *ClaudeSettings {
	return &ClaudeSettings{
		AgentSettings: AgentSettings{
			Env: make(map[string]string),
		},
		ModelAliases: &ClaudeModelAliases{},
	}
}

// NewConfigAgents creates a new ConfigAgents with default settings.
func NewConfigAgents() *ConfigAgents {
	return &ConfigAgents{
		Default: "claude",
	}
}

// GetAgentSettings returns the AgentSettings for the given agent name.
// For Claude, it returns the AgentSettings embedded in ClaudeSettings.
func (c *ConfigAgents) GetAgentSettings(agentName string) *AgentSettings {
	switch agentName {
	case "claude":
		if c.Claude != nil {
			return &c.Claude.AgentSettings
		}
		return nil
	case "opencode":
		return c.OpenCode
	case "github-copilot":
		return c.GitHubCopilot
	case "codex":
		return c.Codex
	default:
		return nil
	}
}

// SetAgentSettings sets the AgentSettings for the given agent name.
func (c *ConfigAgents) SetAgentSettings(agentName string, settings *AgentSettings) {
	switch agentName {
	case "claude":
		if c.Claude == nil {
			c.Claude = NewClaudeSettings()
		}
		c.Claude.AgentSettings = *settings
	case "opencode":
		c.OpenCode = settings
	case "github-copilot":
		c.GitHubCopilot = settings
	case "codex":
		c.Codex = settings
	}
}

// GetOrCreateAgentSettings gets existing settings or creates new ones.
func (c *ConfigAgents) GetOrCreateAgentSettings(agentName string) *AgentSettings {
	switch agentName {
	case "claude":
		if c.Claude == nil {
			c.Claude = NewClaudeSettings()
		}
		return &c.Claude.AgentSettings
	case "opencode":
		if c.OpenCode == nil {
			c.OpenCode = NewAgentSettings()
		}
		return c.OpenCode
	case "github-copilot":
		if c.GitHubCopilot == nil {
			c.GitHubCopilot = NewAgentSettings()
		}
		return c.GitHubCopilot
	case "codex":
		if c.Codex == nil {
			c.Codex = NewAgentSettings()
		}
		return c.Codex
	default:
		return nil
	}
}
