package config

import (
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

type ConfigScope string

const (
	ScopeDefault       ConfigScope = "default"
	ScopeGlobal        ConfigScope = "global"
	ScopeProfile       ConfigScope = "profile"
	ScopeTeamLocal     ConfigScope = "local"
	ScopePersonalLocal ConfigScope = "personal"
)

type ResolvedValue struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Source    ConfigScope `json:"source"`
	Sensitive bool        `json:"sensitive"`
}

type ResolvedConfig struct {
	Values        map[string]*ResolvedValue `json:"values"`
	ActiveProfile string                    `json:"active_profile"`
}

func MergeConfigs(defaults, global, profile, teamLocal, personalLocal *AgentConfig) *ResolvedConfig {
	resolved := &ResolvedConfig{
		Values: make(map[string]*ResolvedValue),
	}

	layers := []struct {
		config *AgentConfig
		scope  ConfigScope
	}{
		{defaults, ScopeDefault},
		{global, ScopeGlobal},
		{profile, ScopeProfile},
		{teamLocal, ScopeTeamLocal},
		{personalLocal, ScopePersonalLocal},
	}

	for _, keyDef := range GetRegistry().List() {
		if keyDef.Key == "active-profile" {
			continue
		}

		var finalValue interface{}
		var finalSource ConfigScope = ScopeDefault
		var found bool

		for _, layer := range layers {
			if layer.config == nil {
				continue
			}

			value := getAgentConfigValue(layer.config, keyDef.Key)
			if value != nil && !isZero(value) {
				finalValue = value
				finalSource = layer.scope
				found = true
			}
		}

		if found {
			resolved.Values[keyDef.Key] = &ResolvedValue{
				Key:       keyDef.Key,
				Value:     finalValue,
				Source:    finalSource,
				Sensitive: keyDef.Sensitive,
			}
		}
	}

	return resolved
}

func getAgentConfigValue(cfg *AgentConfig, key string) interface{} {
	if cfg == nil {
		return nil
	}

	switch key {
	case "agent.base-url":
		return cfg.BaseURL
	case "agent.auth-token":
		return cfg.AuthToken
	case "agent.api-key":
		return cfg.APIKey
	case "agent.model":
		return cfg.Model
	case "agent.model.sonnet":
		return cfg.ModelSonnet
	case "agent.model.opus":
		return cfg.ModelOpus
	case "agent.model.haiku":
		return cfg.ModelHaiku
	case "agent.subagent-model":
		return cfg.SubagentModel
	case "agent.provider":
		return cfg.Provider
	case "agent.permission-mode":
		return cfg.PermissionMode
	case "agent.skip-permissions":
		return cfg.SkipPermissions
	case "agent.effort":
		return cfg.Effort
	case "agent.allowed-tools":
		return cfg.AllowedTools
	case "agent.env":
		return cfg.Env
	default:
		return nil
	}
}

func isZero(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice:
		return v.Len() == 0
	case reflect.Map:
		return v.Len() == 0
	case reflect.Ptr:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

func (r *ResolvedConfig) Get(key string) *ResolvedValue {
	return r.Values[key]
}

func (r *ResolvedConfig) GetValue(key string) interface{} {
	if v, ok := r.Values[key]; ok {
		return v.Value
	}
	return nil
}

func (r *ResolvedConfig) GetEnvVars() map[string]string {
	envVars := make(map[string]string)

	for _, keyDef := range GetRegistry().List() {
		if keyDef.EnvVar == "" {
			continue
		}

		if resolved := r.Get(keyDef.Key); resolved != nil {
			if str, ok := resolved.Value.(string); ok && str != "" {
				envVars[keyDef.EnvVar] = str
			}
		} else if keyDef.Default != nil {
			if str, ok := keyDef.Default.(string); ok && str != "" {
				envVars[keyDef.EnvVar] = str
			}
		}
	}

	if env := r.Get("agent.env"); env != nil {
		if envMap, ok := env.Value.(map[string]string); ok {
			for k, v := range envMap {
				envVars[k] = v
			}
		}
	}

	return envVars
}

func ResolveAgentEnv() map[string]string {
	globalCfg, _ := Load()
	if globalCfg == nil {
		globalCfg = DefaultConfig()
	}

	var teamLocal, personalLocal *AgentConfig

	teamMeta, err := loadProjectMetadata(".")
	if err == nil && teamMeta != nil && teamMeta.Agent != nil {
		teamLocal = teamMeta.Agent
	}

	personalMeta, err := loadPersonalMetadata(".")
	if err == nil && personalMeta != nil && personalMeta.Agent != nil {
		personalLocal = personalMeta.Agent
	}

	var profile *AgentConfig
	if globalCfg.ActiveProfile != "" && globalCfg.Profiles != nil {
		profile = globalCfg.Profiles[globalCfg.ActiveProfile]
	}

	resolved := MergeConfigs(
		DefaultAgentConfig(),
		globalCfg.Agent,
		profile,
		teamLocal,
		personalLocal,
	)

	return resolved.GetEnvVars()
}

func loadProjectMetadata(projectPath string) (*struct {
	Agent *AgentConfig `yaml:"agent,omitempty"`
}, error) {
	metaPath := filepath.Join(projectPath, "specledger", "specledger.yaml")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta struct {
		Agent *AgentConfig `yaml:"agent,omitempty"`
	}
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func loadPersonalMetadata(projectPath string) (*struct {
	Agent *AgentConfig `yaml:"agent,omitempty"`
}, error) {
	personalPath := filepath.Join(projectPath, "specledger", "specledger.local.yaml")
	data, err := os.ReadFile(personalPath)
	if err != nil {
		return nil, err
	}

	var meta struct {
		Agent *AgentConfig `yaml:"agent,omitempty"`
	}
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}
