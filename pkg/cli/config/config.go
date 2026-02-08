package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration
type Config struct {
	DefaultProjectDir  string `yaml:"default_project_dir" json:"default_project_dir"`
	PreferredShell     string `yaml:"preferred_shell" json:"preferred_shell"`
	TUIEnabled         bool   `yaml:"tui_enabled" json:"tui_enabled"`
	AutoInstallDeps    bool   `yaml:"auto_install_deps" json:"auto_install_deps"`
	FallbackToPlainCLI bool   `yaml:"fallback_to_plain_cli" json:"fallback_to_plain_cli"`
	LogLevel           string `yaml:"log_level" json:"log_level"`
	Theme              string `yaml:"theme" json:"theme"`
	Language           string `yaml:"language" json:"language"`
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
	}
}

// Load loads the configuration from file
func Load() (*Config, error) {
	configPath := getConfigPath()

	// If config doesn't exist, return defaults
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

	// Ensure required fields are set
	if cfg.DefaultProjectDir == "" {
		cfg.DefaultProjectDir = filepath.Join(os.Getenv("HOME"), "demos")
	}
	if cfg.PreferredShell == "" {
		cfg.PreferredShell = "zsh"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "debug"
	}

	return &cfg, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath := getConfigPath()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// #nosec G306 -- config file needs to be readable, 0644 is appropriate
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "specledger")
	return filepath.Join(configDir, "config.yaml")
}
