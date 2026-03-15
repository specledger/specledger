package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// MigrateConfig migrates old AgentConfig format to new ConfigAgents format.
// This is called automatically when loading config.
func (c *Config) MigrateConfig() {
	// Skip if no old config or already migrated
	if c.Agent == nil {
		return
	}

	// Check if new config already exists
	if c.Agents != nil && c.Agents.Claude != nil {
		return // Already migrated
	}

	// Initialize new config
	if c.Agents == nil {
		c.Agents = NewConfigAgents()
	}

	// Migrate Claude settings
	if c.Agent.APIKey != "" || c.Agent.BaseURL != "" ||
		c.Agent.Model != "" || c.Agent.ModelSonnet != "" ||
		c.Agent.ModelOpus != "" || c.Agent.ModelHaiku != "" ||
		len(c.Agent.Env) > 0 {

		if c.Agents.Claude == nil {
			c.Agents.Claude = NewClaudeSettings()
		}

		// Migrate base settings
		c.Agents.Claude.APIKey = c.Agent.APIKey
		c.Agents.Claude.BaseURL = c.Agent.BaseURL
		c.Agents.Claude.Model = c.Agent.Model

		// Migrate env vars
		if c.Agent.Env != nil {
			if c.Agents.Claude.Env == nil {
				c.Agents.Claude.Env = make(map[string]string)
			}
			for k, v := range c.Agent.Env {
				c.Agents.Claude.Env[k] = v
			}
		}

		// Migrate model aliases
		if c.Agent.ModelSonnet != "" || c.Agent.ModelOpus != "" || c.Agent.ModelHaiku != "" {
			if c.Agents.Claude.ModelAliases == nil {
				c.Agents.Claude.ModelAliases = &ClaudeModelAliases{}
			}
			c.Agents.Claude.ModelAliases.Sonnet = c.Agent.ModelSonnet
			c.Agents.Claude.ModelAliases.Opus = c.Agent.ModelOpus
			c.Agents.Claude.ModelAliases.Haiku = c.Agent.ModelHaiku
		}
	}

	// Set default agent
	if c.Agents.Default == "" {
		c.Agents.Default = "claude"
	}

	// Print migration notice (only once)
	printMigrationNotice()
}

var migrationNoticePrinted = false

func printMigrationNotice() {
	if migrationNoticePrinted {
		return
	}
	migrationNoticePrinted = true

	fmt.Println("Notice: Config migrated to new format.")
	fmt.Println("Old keys like 'agent.api-key' are now 'agent.claude.api_key'.")
	fmt.Println("Run 'sl config show' to see the new structure.")
	fmt.Println()
}

// MigrateConfigFile performs a one-time migration of the config file.
// This writes the migrated config back to disk.
func MigrateConfigFile() error {
	cfg, err := Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if migration is needed
	if cfg.Agent == nil || (cfg.Agents != nil && cfg.Agents.Claude != nil) {
		return nil // No migration needed
	}

	// Perform migration
	cfg.MigrateConfig()

	// Save migrated config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save migrated config: %w", err)
	}

	return nil
}

// LoadWithMigration loads config and performs migration if needed.
func LoadWithMigration() (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	// Perform in-memory migration
	cfg.MigrateConfig()

	return cfg, nil
}

// migrateProjectConfig migrates project-level config files.
func migrateProjectConfig(projectPath string) error {
	teamPath := filepath.Join(projectPath, "specledger", "specledger.yaml")
	personalPath := filepath.Join(projectPath, "specledger", "specledger.local.yaml")

	// Migrate team config
	if err := migrateYamlFile(teamPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to migrate team config: %w", err)
	}

	// Migrate personal config
	if err := migrateYamlFile(personalPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to migrate personal config: %w", err)
	}

	return nil
}

// migrateYamlFile migrates a single YAML config file.
func migrateYamlFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Check if migration is needed
	agentRaw, hasOldAgent := raw["agent"]
	if !hasOldAgent {
		return nil // No old config
	}

	_, hasNewAgents := raw["agents"]
	if hasNewAgents {
		return nil // Already migrated
	}

	// Perform migration
	agentMap, ok := agentRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	// Create new structure
	newClaude := make(map[string]interface{})
	if v, ok := agentMap["api-key"]; ok {
		newClaude["api_key"] = v
	}
	if v, ok := agentMap["base-url"]; ok {
		newClaude["base_url"] = v
	}
	if v, ok := agentMap["model"]; ok {
		newClaude["model"] = v
	}
	if v, ok := agentMap["env"]; ok {
		newClaude["env"] = v
	}

	// Migrate model aliases
	modelAliases := make(map[string]interface{})
	if v, ok := agentMap["model.sonnet"]; ok {
		modelAliases["sonnet"] = v
	}
	if v, ok := agentMap["model.opus"]; ok {
		modelAliases["opus"] = v
	}
	if v, ok := agentMap["model.haiku"]; ok {
		modelAliases["haiku"] = v
	}
	if len(modelAliases) > 0 {
		newClaude["model_aliases"] = modelAliases
	}

	// Build new agents structure
	agents := map[string]interface{}{
		"default": "claude",
	}
	if len(newClaude) > 0 {
		agents["claude"] = newClaude
	}

	// Update raw map
	raw["agents"] = agents
	delete(raw, "agent")

	// Write back
	newData, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}

	return os.WriteFile(path, newData, 0600)
}
