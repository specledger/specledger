package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type PersonalConfig struct {
	Agent         *AgentConfig `yaml:"agent,omitempty"`
	ActiveProfile string       `yaml:"active-profile,omitempty"`
}

func LoadPersonal(projectPath string) (*PersonalConfig, error) {
	personalPath := filepath.Join(projectPath, "specledger", "specledger.local.yaml")

	if _, err := os.Stat(personalPath); os.IsNotExist(err) {
		return &PersonalConfig{
			Agent: &AgentConfig{},
		}, nil
	}

	data, err := os.ReadFile(personalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read personal config: %w", err)
	}

	var cfg PersonalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse personal config: %w", err)
	}

	if cfg.Agent == nil {
		cfg.Agent = &AgentConfig{}
	}

	return &cfg, nil
}

func (c *PersonalConfig) Save(projectPath string) error {
	personalPath := filepath.Join(projectPath, "specledger", "specledger.local.yaml")

	if err := os.MkdirAll(filepath.Dir(personalPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal personal config: %w", err)
	}

	if err := os.WriteFile(personalPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write personal config: %w", err)
	}

	return nil
}

func (c *PersonalConfig) HasAgentConfig() bool {
	if c.Agent == nil {
		return false
	}
	return c.Agent.BaseURL != "" ||
		c.Agent.AuthToken != "" ||
		c.Agent.APIKey != "" ||
		c.Agent.Model != "" ||
		c.Agent.ModelSonnet != "" ||
		c.Agent.ModelOpus != "" ||
		c.Agent.ModelHaiku != "" ||
		c.Agent.SubagentModel != "" ||
		c.Agent.Provider != "" ||
		c.Agent.PermissionMode != "" ||
		c.Agent.Effort != "" ||
		len(c.Agent.AllowedTools) > 0 ||
		len(c.Agent.Env) > 0
}

func GetPersonalConfigPath(projectPath string) string {
	return filepath.Join(projectPath, "specledger", "specledger.local.yaml")
}

func (c *PersonalConfig) SetAgentConfigValue(key, value string) error {
	if c.Agent == nil {
		c.Agent = &AgentConfig{}
	}

	switch key {
	case "agent.base-url":
		c.Agent.BaseURL = value
	case "agent.auth-token":
		c.Agent.AuthToken = value
	case "agent.api-key":
		c.Agent.APIKey = value
	case "agent.model":
		c.Agent.Model = value
	case "agent.model.sonnet":
		c.Agent.ModelSonnet = value
	case "agent.model.opus":
		c.Agent.ModelOpus = value
	case "agent.model.haiku":
		c.Agent.ModelHaiku = value
	case "agent.subagent-model":
		c.Agent.SubagentModel = value
	case "agent.provider":
		c.Agent.Provider = value
	case "agent.permission-mode":
		c.Agent.PermissionMode = value
	case "agent.skip-permissions":
		c.Agent.SkipPermissions = value == "true"
	case "agent.effort":
		c.Agent.Effort = value
	case "active-profile":
		c.ActiveProfile = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func (c *PersonalConfig) UnsetAgentConfigValue(key string) error {
	if c.Agent == nil {
		return fmt.Errorf("no config value set for: %s", key)
	}

	switch key {
	case "agent.base-url":
		c.Agent.BaseURL = ""
	case "agent.auth-token":
		c.Agent.AuthToken = ""
	case "agent.api-key":
		c.Agent.APIKey = ""
	case "agent.model":
		c.Agent.Model = ""
	case "agent.model.sonnet":
		c.Agent.ModelSonnet = ""
	case "agent.model.opus":
		c.Agent.ModelOpus = ""
	case "agent.model.haiku":
		c.Agent.ModelHaiku = ""
	case "agent.subagent-model":
		c.Agent.SubagentModel = ""
	case "agent.provider":
		c.Agent.Provider = ""
	case "agent.permission-mode":
		c.Agent.PermissionMode = ""
	case "agent.skip-permissions":
		c.Agent.SkipPermissions = false
	case "agent.effort":
		c.Agent.Effort = ""
	case "active-profile":
		c.ActiveProfile = ""
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}
