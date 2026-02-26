package config

import (
	"fmt"
	"strings"
)

type ConfigKeyType string

const (
	KeyTypeString      ConfigKeyType = "string"
	KeyTypeBool        ConfigKeyType = "bool"
	KeyTypeEnum        ConfigKeyType = "enum"
	KeyTypeStringList  ConfigKeyType = "string-list"
	KeyTypeStringMap   ConfigKeyType = "string-map"
)

type ConfigKeyDef struct {
	Key          string        `json:"key"`
	Type         ConfigKeyType `json:"type"`
	EnvVar       string        `json:"env_var"`
	CLIFlag      string        `json:"cli_flag"`
	Default      interface{}   `json:"default"`
	Sensitive    bool          `json:"sensitive"`
	Description  string        `json:"description"`
	EnumValues   []string      `json:"enum_values,omitempty"`
	Category     string        `json:"category"`
}

type SchemaRegistry struct {
	keys map[string]*ConfigKeyDef
}

var registry *SchemaRegistry

func init() {
	registry = &SchemaRegistry{
		keys: make(map[string]*ConfigKeyDef),
	}
	registerAgentKeys()
}

func registerAgentKeys() {
	agentKeys := []*ConfigKeyDef{
		{
			Key:         "agent.base-url",
			Type:        KeyTypeString,
			EnvVar:      "ANTHROPIC_BASE_URL",
			Description: "Custom API endpoint URL",
			Category:    "Provider",
		},
		{
			Key:         "agent.auth-token",
			Type:        KeyTypeString,
			EnvVar:      "ANTHROPIC_AUTH_TOKEN",
			Sensitive:   true,
			Description: "Auth token (sensitive, masked)",
			Category:    "Provider",
		},
		{
			Key:         "agent.api-key",
			Type:        KeyTypeString,
			EnvVar:      "ANTHROPIC_API_KEY",
			Sensitive:   true,
			Description: "API key (sensitive, masked)",
			Category:    "Provider",
		},
		{
			Key:         "agent.model",
			Type:        KeyTypeString,
			EnvVar:      "ANTHROPIC_MODEL",
			Description: "Default model (alias or full name)",
			Category:    "Models",
		},
		{
			Key:         "agent.model.sonnet",
			Type:        KeyTypeString,
			EnvVar:      "ANTHROPIC_DEFAULT_SONNET_MODEL",
			Description: "Model for sonnet alias",
			Category:    "Models",
		},
		{
			Key:         "agent.model.opus",
			Type:        KeyTypeString,
			EnvVar:      "ANTHROPIC_DEFAULT_OPUS_MODEL",
			Description: "Model for opus alias",
			Category:    "Models",
		},
		{
			Key:         "agent.model.haiku",
			Type:        KeyTypeString,
			EnvVar:      "ANTHROPIC_DEFAULT_HAIKU_MODEL",
			Description: "Model for haiku alias",
			Category:    "Models",
		},
		{
			Key:         "agent.subagent-model",
			Type:        KeyTypeString,
			EnvVar:      "CLAUDE_CODE_SUBAGENT_MODEL",
			Description: "Model for subagents",
			Category:    "Models",
		},
		{
			Key:         "agent.provider",
			Type:        KeyTypeEnum,
			EnvVar:      "",
			Default:     "anthropic",
			EnumValues:  []string{"anthropic", "bedrock", "vertex"},
			Description: "Provider selection",
			Category:    "Provider",
		},
		{
			Key:         "agent.permission-mode",
			Type:        KeyTypeEnum,
			CLIFlag:     "--permission-mode",
			Default:     "default",
			EnumValues:  []string{"default", "plan", "bypassPermissions", "acceptEdits", "dontAsk"},
			Description: "Permission mode for agent",
			Category:    "Launch Flags",
		},
		{
			Key:         "agent.skip-permissions",
			Type:        KeyTypeBool,
			CLIFlag:     "--dangerously-skip-permissions",
			Default:     false,
			Description: "Skip permission prompts",
			Category:    "Launch Flags",
		},
		{
			Key:         "agent.effort",
			Type:        KeyTypeEnum,
			CLIFlag:     "--effort",
			EnumValues:  []string{"low", "medium", "high"},
			Description: "Effort level",
			Category:    "Launch Flags",
		},
		{
			Key:         "agent.allowed-tools",
			Type:        KeyTypeStringList,
			CLIFlag:     "--allowedTools",
			Description: "Tools allowed without prompts",
			Category:    "Launch Flags",
		},
		{
			Key:         "agent.env",
			Type:        KeyTypeStringMap,
			Description: "Arbitrary env vars injected into agent",
			Category:    "Environment",
		},
		{
			Key:         "active-profile",
			Type:        KeyTypeString,
			Description: "Currently active profile name",
			Category:    "Profiles",
		},
	}

	for _, key := range agentKeys {
		registry.keys[key.Key] = key
	}
}

func (r *SchemaRegistry) Lookup(key string) (*ConfigKeyDef, error) {
	if def, ok := r.keys[key]; ok {
		return def, nil
	}
	return nil, fmt.Errorf("unknown config key: %s", key)
}

func (r *SchemaRegistry) List() []*ConfigKeyDef {
	result := make([]*ConfigKeyDef, 0, len(r.keys))
	for _, def := range r.keys {
		result = append(result, def)
	}
	return result
}

func (r *SchemaRegistry) ListByCategory() map[string][]*ConfigKeyDef {
	result := make(map[string][]*ConfigKeyDef)
	for _, def := range r.keys {
		result[def.Category] = append(result[def.Category], def)
	}
	return result
}

func (r *SchemaRegistry) IsValidKey(key string) bool {
	_, ok := r.keys[key]
	return ok
}

func (r *SchemaRegistry) FindSimilar(key string) []string {
	var similar []string
	keyLower := strings.ToLower(key)
	for k := range r.keys {
		if strings.Contains(strings.ToLower(k), keyLower) || strings.Contains(keyLower, strings.ToLower(k)) {
			similar = append(similar, k)
		}
	}
	return similar
}

func GetRegistry() *SchemaRegistry {
	return registry
}

func LookupKey(key string) (*ConfigKeyDef, error) {
	return registry.Lookup(key)
}

func IsValidKey(key string) bool {
	return registry.IsValidKey(key)
}
