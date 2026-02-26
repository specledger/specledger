package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	BaseURL           string            `yaml:"base-url" json:"base_url"`
	AuthToken         string            `yaml:"auth-token" json:"auth_token" sensitive:"true"`
	APIKey            string            `yaml:"api-key" json:"api_key" sensitive:"true"`
	Model             string            `yaml:"model" json:"model"`
	ModelSonnet       string            `yaml:"model.sonnet" json:"model_sonnet"`
	ModelOpus         string            `yaml:"model.opus" json:"model_opus"`
	ModelHaiku        string            `yaml:"model.haiku" json:"model_haiku"`
	SubagentModel     string            `yaml:"subagent-model" json:"subagent_model"`
	Provider          string            `yaml:"provider" json:"provider"`
	PermissionMode    string            `yaml:"permission-mode" json:"permission_mode"`
	SkipPermissions   bool              `yaml:"skip-permissions" json:"skip_permissions"`
	Effort            string            `yaml:"effort" json:"effort"`
	AllowedTools      []string          `yaml:"allowed-tools" json:"allowed_tools"`
	Env               map[string]string `yaml:"env" json:"env"`
}

func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		Provider:       "anthropic",
		PermissionMode: "default",
		Env:           make(map[string]string),
	}
}

type Config struct {
	DefaultProjectDir  string                   `yaml:"default_project_dir" json:"default_project_dir"`
	PreferredShell     string                   `yaml:"preferred_shell" json:"preferred_shell"`
	TUIEnabled         bool                     `yaml:"tui_enabled" json:"tui_enabled"`
	AutoInstallDeps    bool                     `yaml:"auto_install_deps" json:"auto_install_deps"`
	FallbackToPlainCLI bool                     `yaml:"fallback_to_plain_cli" json:"fallback_to_plain_cli"`
	LogLevel           string                   `yaml:"log_level" json:"log_level"`
	Theme              string                   `yaml:"theme" json:"theme"`
	Language           string                   `yaml:"language" json:"language"`
	Agent              *AgentConfig             `yaml:"agent,omitempty" json:"agent,omitempty"`
	Profiles           map[string]*AgentConfig  `yaml:"profiles,omitempty" json:"profiles,omitempty"`
	ActiveProfile      string                   `yaml:"active-profile,omitempty" json:"active_profile,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultProjectDir:  filepath.Join(os.Getenv("HOME"), "demos"),
		PreferredShell:     "zsh",
		TUIEnabled:         true,
		AutoInstallDeps:    false,
		FallbackToPlainCLI: true,
		LogLevel:           "debug",
		Theme:              "default",
		Language:           "en",
		Agent:              DefaultAgentConfig(),
		Profiles:           make(map[string]*AgentConfig),
	}
}

// Load loads the configuration from file
func Load() (*Config, error) {
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.DefaultProjectDir == "" {
		cfg.DefaultProjectDir = filepath.Join(os.Getenv("HOME"), "demos")
	}
	if cfg.PreferredShell == "" {
		cfg.PreferredShell = "zsh"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "debug"
	}
	if cfg.Agent == nil {
		cfg.Agent = DefaultAgentConfig()
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]*AgentConfig)
	}

	return &cfg, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath := getConfigPath()

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) HasSensitiveValues() bool {
	if c.Agent != nil {
		if c.Agent.AuthToken != "" || c.Agent.APIKey != "" {
			return true
		}
	}
	for _, profile := range c.Profiles {
		if profile != nil && (profile.AuthToken != "" || profile.APIKey != "") {
			return true
		}
	}
	return false
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	configDir := filepath.Join(os.Getenv("HOME"), ".specledger")
	return filepath.Join(configDir, "config.yaml")
}
