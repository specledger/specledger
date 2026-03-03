package config

import (
	"fmt"
)

func (c *Config) CreateProfile(name string, agent *AgentConfig) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	if c.Profiles == nil {
		c.Profiles = make(map[string]*AgentConfig)
	}

	if _, exists := c.Profiles[name]; exists {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	c.Profiles[name] = agent
	return nil
}

func (c *Config) DeleteProfile(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	if c.Profiles == nil {
		return fmt.Errorf("no profiles exist")
	}

	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	delete(c.Profiles, name)

	if c.ActiveProfile == name {
		c.ActiveProfile = ""
	}

	return nil
}

func (c *Config) ListProfiles() []string {
	if c.Profiles == nil {
		return []string{}
	}

	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	return names
}

func (c *Config) GetProfile(name string) (*AgentConfig, error) {
	if name == "" {
		return nil, fmt.Errorf("profile name cannot be empty")
	}

	if c.Profiles == nil {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}

	profile, exists := c.Profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}

	return profile, nil
}

func (c *Config) SetActiveProfile(name string) error {
	if name == "" {
		c.ActiveProfile = ""
		return nil
	}

	if c.Profiles == nil {
		return fmt.Errorf("profile '%s' not found", name)
	}

	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	c.ActiveProfile = name
	return nil
}

func (c *Config) GetActiveProfile() string {
	return c.ActiveProfile
}

func (c *Config) HasProfiles() bool {
	return len(c.Profiles) > 0
}

func (c *Config) ProfileExists(name string) bool {
	if c.Profiles == nil {
		return false
	}
	_, exists := c.Profiles[name]
	return exists
}
