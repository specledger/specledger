package agent

import (
	"strings"
)

type Agent struct {
	Name           string
	Command        string
	ConfigDir      string
	InstallCommand string

	// Env var mappings for config resolution
	APIKeyEnvVar  string // e.g., "ANTHROPIC_API_KEY" for Claude
	BaseURLEnvVar string // e.g., "ANTHROPIC_BASE_URL" for Claude
	ModelEnvVar   string // e.g., "ANTHROPIC_MODEL" for Claude
}

type Registry struct {
	agents map[string]Agent
}

var defaultRegistry *Registry

func init() {
	defaultRegistry = NewRegistry()
}

func NewRegistry() *Registry {
	agents := []Agent{
		{
			Name:           "Claude Code",
			Command:        "claude",
			ConfigDir:      ".claude",
			InstallCommand: "npm install -g @anthropic-ai/claude-code",
			APIKeyEnvVar:   "ANTHROPIC_AUTH_TOKEN",
			BaseURLEnvVar:  "ANTHROPIC_BASE_URL",
			ModelEnvVar:    "ANTHROPIC_MODEL",
		},
		{
			Name:           "OpenCode",
			Command:        "opencode",
			ConfigDir:      ".opencode",
			InstallCommand: "go install github.com/opencode-ai/opencode@latest",
			APIKeyEnvVar:   "",
			BaseURLEnvVar:  "",
			ModelEnvVar:    "", // OpenCode uses config file for model
		},
		{
			Name:           "Copilot CLI",
			Command:        "github-copilot",
			ConfigDir:      ".github",
			InstallCommand: "npm install -g @github/copilot",
			APIKeyEnvVar:   "GITHUB_TOKEN",
			BaseURLEnvVar:  "",
			ModelEnvVar:    "", // Copilot uses config file for model
		},
		{
			Name:           "Codex",
			Command:        "codex",
			ConfigDir:      ".codex",
			InstallCommand: "npm install -g @openai/codex",
			APIKeyEnvVar:   "OPENAI_API_KEY",
			BaseURLEnvVar:  "OPENAI_BASE_URL",
			ModelEnvVar:    "", // Codex uses config file for model
		},
	}

	agentMap := make(map[string]Agent)
	for _, a := range agents {
		key := strings.ToLower(a.Name)
		agentMap[key] = a
		agentMap[strings.ToLower(a.Command)] = a
	}

	return &Registry{agents: agentMap}
}

func (r *Registry) Lookup(name string) (Agent, bool) {
	key := strings.ToLower(strings.TrimSpace(name))
	a, ok := r.agents[key]
	return a, ok
}

func (r *Registry) All() []Agent {
	seen := make(map[string]bool)
	var result []Agent
	for _, a := range r.agents {
		cmdKey := strings.ToLower(a.Command)
		if !seen[cmdKey] {
			seen[cmdKey] = true
			result = append(result, a)
		}
	}
	return result
}

func Lookup(name string) (Agent, bool) {
	return defaultRegistry.Lookup(name)
}

func All() []Agent {
	return defaultRegistry.All()
}
