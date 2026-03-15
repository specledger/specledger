package config

import (
	"fmt"
	"strings"

	"github.com/specledger/specledger/internal/agent"
)

type ConfigKeyType string

const (
	KeyTypeString     ConfigKeyType = "string"
	KeyTypeBool       ConfigKeyType = "bool"
	KeyTypeEnum       ConfigKeyType = "enum"
	KeyTypeStringList ConfigKeyType = "string-list"
	KeyTypeStringMap  ConfigKeyType = "string-map"
)

type ConfigKeyDef struct {
	Key         string        `json:"key"`
	Type        ConfigKeyType `json:"type"`
	EnvVar      string        `json:"env_var"`
	CLIFlag     string        `json:"cli_flag"`
	Default     interface{}   `json:"default"`
	Sensitive   bool          `json:"sensitive"`
	Description string        `json:"description"`
	EnumValues  []string      `json:"enum_values,omitempty"`
	Category    string        `json:"category"`
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
		// Universal agent keys
		{
			Key:         "agent.default",
			Type:        KeyTypeString,
			Default:     "claude",
			Description: "Default coding agent to launch (claude, opencode, github-copilot, codex)",
			Category:    "Agent",
		},
		{
			Key:         "agent.env",
			Type:        KeyTypeStringMap,
			Description: "Arbitrary env vars injected into agent (deprecated, use agent.<name>.env)",
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

	// Handle agent.env.<VAR> (legacy)
	if strings.HasPrefix(key, "agent.env.") {
		if def, ok := r.keys["agent.env"]; ok {
			return def, nil
		}
	}

	// Handle per-agent keys: agent.<name>.<field>
	if strings.HasPrefix(key, "agent.") {
		parts := strings.Split(key, ".")
		if len(parts) >= 3 {
			agentName := parts[1]
			ag, found := agent.Lookup(agentName)
			if !found {
				return nil, fmt.Errorf("unknown agent name: %s (valid: claude, opencode, github-copilot, codex)", agentName)
			}

			// agent.<name>.api_key
			if len(parts) == 3 && parts[2] == "api_key" {
				return &ConfigKeyDef{
					Key:         key,
					Type:        KeyTypeString,
					Sensitive:   true,
					Description: fmt.Sprintf("API key for %s agent", ag.Name),
					Category:    "Per-Agent",
				}, nil
			}

			// agent.<name>.base_url
			if len(parts) == 3 && parts[2] == "base_url" {
				return &ConfigKeyDef{
					Key:         key,
					Type:        KeyTypeString,
					Description: fmt.Sprintf("Custom endpoint URL for %s agent", ag.Name),
					Category:    "Per-Agent",
				}, nil
			}

			// agent.<name>.model
			if len(parts) == 3 && parts[2] == "model" {
				return &ConfigKeyDef{
					Key:         key,
					Type:        KeyTypeString,
					Description: fmt.Sprintf("Model selection for %s agent", ag.Name),
					Category:    "Per-Agent",
				}, nil
			}

			// agent.<name>.arguments
			if len(parts) == 3 && parts[2] == "arguments" {
				return &ConfigKeyDef{
					Key:         key,
					Type:        KeyTypeString,
					Description: fmt.Sprintf("CLI arguments for %s agent", ag.Name),
					Category:    "Per-Agent",
				}, nil
			}

			// agent.<name>.env
			if len(parts) == 3 && parts[2] == "env" {
				return &ConfigKeyDef{
					Key:         key,
					Type:        KeyTypeStringMap,
					Description: fmt.Sprintf("Environment variables for %s agent", ag.Name),
					Category:    "Per-Agent",
				}, nil
			}

			// agent.<name>.env.<VAR>
			if len(parts) == 4 && parts[2] == "env" {
				return &ConfigKeyDef{
					Key:         key,
					Type:        KeyTypeString,
					Description: fmt.Sprintf("Environment variable %s for %s agent", parts[3], ag.Name),
					Category:    "Per-Agent",
				}, nil
			}

			// Claude-specific: agent.claude.model_aliases.<alias>
			if agentName == "claude" && len(parts) == 4 && parts[2] == "model_aliases" {
				alias := parts[3]
				if alias == "sonnet" || alias == "opus" || alias == "haiku" {
					return &ConfigKeyDef{
						Key:         key,
						Type:        KeyTypeString,
						Description: fmt.Sprintf("Claude model alias for %s (e.g., claude-%s-4-20250514)", alias, alias),
						Category:    "Claude-Specific",
					}, nil
				}
				return nil, fmt.Errorf("unknown model alias: %s (valid: sonnet, opus, haiku)", alias)
			}
		}
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
	if _, ok := r.keys[key]; ok {
		return true
	}

	if !strings.HasPrefix(key, "agent.") {
		return false
	}

	// Legacy: agent.env.<VAR> (must check before agent name lookup)
	if strings.HasPrefix(key, "agent.env.") && len(strings.Split(key, ".")) == 3 {
		return true
	}

	parts := strings.Split(key, ".")
	if len(parts) < 3 {
		return false
	}

	agentName := parts[1]
	ag, found := agent.Lookup(agentName)
	if !found {
		return false
	}

	// Check for valid field patterns
	if len(parts) == 3 {
		field := parts[2]
		return field == "api_key" || field == "base_url" || field == "model" ||
			field == "arguments" || field == "env"
	}

	// agent.<name>.env.<VAR>
	if len(parts) == 4 && parts[2] == "env" {
		return true
	}

	// agent.claude.model_aliases.<alias>
	if agentName == "claude" && len(parts) == 4 && parts[2] == "model_aliases" {
		alias := parts[3]
		return alias == "sonnet" || alias == "opus" || alias == "haiku"
	}

	_ = ag // avoid unused variable warning

	return false
}

func (r *SchemaRegistry) FindSimilar(key string) []string {
	var similar []string
	keyLower := strings.ToLower(key)

	for k := range r.keys {
		kLower := strings.ToLower(k)
		if strings.Contains(kLower, keyLower) || strings.Contains(keyLower, kLower) {
			similar = append(similar, k)
			continue
		}

		parts := strings.Split(keyLower, ".")
		kParts := strings.Split(kLower, ".")
		if len(parts) == len(kParts) && len(parts) > 1 {
			matchScore := 0
			for i := range parts {
				if parts[i] == kParts[i] {
					matchScore += 2
				} else if levenshteinDistance(parts[i], kParts[i]) <= 2 {
					matchScore += 1
				}
			}
			if matchScore >= len(parts) {
				similar = append(similar, k)
			}
		}
	}

	return similar
}

func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	if a[0] == b[0] {
		return levenshteinDistance(a[1:], b[1:])
	}

	return 1 + min(
		levenshteinDistance(a[1:], b),
		levenshteinDistance(a, b[1:]),
		levenshteinDistance(a[1:], b[1:]),
	)
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
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
