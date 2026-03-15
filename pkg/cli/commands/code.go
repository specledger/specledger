package commands

import (
	"fmt"
	"os"

	"github.com/specledger/specledger/internal/agent"
	"github.com/specledger/specledger/pkg/cli/config"
	"github.com/specledger/specledger/pkg/cli/launcher"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/spf13/cobra"
)

var VarCodeCmd = &cobra.Command{
	Use:   "code [agent]",
	Short: "Launch a coding agent",
	Long: `Launch an AI coding agent in the current directory.

Supported agents:
  claude         Claude Code (default)
  opencode       OpenCode
  github-copilot GitHub Copilot CLI
  codex          OpenAI Codex

The agent is launched with configuration from:
  - Global config: ~/.specledger/config.yaml
  - Project config: specledger/specledger.yaml
  - Personal config: specledger/specledger.local.yaml

Examples:
  sl code              Launch default agent (claude)
  sl code claude       Launch Claude Code
  sl code opencode     Launch OpenCode`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCode,
}

func runCode(cmd *cobra.Command, args []string) error {
	// Get default agent from config
	cfg, _ := config.Load()
	defaultAgent := "claude"
	if cfg != nil && cfg.Agents != nil && cfg.Agents.Default != "" {
		defaultAgent = cfg.Agents.Default
	}

	agentName := defaultAgent
	if len(args) > 0 {
		agentName = args[0]
	}

	ag, found := agent.Lookup(agentName)
	if !found {
		return fmt.Errorf("unknown agent: %s\nValid agents: claude, opencode, github-copilot, codex", agentName)
	}

	wrappedAgent := launcher.NewAgentFromDefinition(ag)
	if err := wrappedAgent.CheckInstalled(); err != nil {
		ui.PrintError(err.Error())
		return fmt.Errorf("agent not available")
	}

	l := launcher.NewLauncherForAgent(ag, ".")

	// Get per-agent settings
	settings := config.ResolveAgentSettings(agentName)
	if settings != nil {
		// Set arguments
		if len(settings.Arguments) > 0 {
			l.SetFlags(settings.Arguments)
		}

		// Map config values to env vars using agent's env var mappings
		envVars := make(map[string]string)

		// Map API key to agent's env var
		if settings.APIKey != "" && ag.APIKeyEnvVar != "" {
			envVars[ag.APIKeyEnvVar] = settings.APIKey
		}

		// Map base URL to agent's env var
		if settings.BaseURL != "" && ag.BaseURLEnvVar != "" {
			envVars[ag.BaseURLEnvVar] = settings.BaseURL
		}

		// Map model to agent's env var
		if settings.Model != "" && ag.ModelEnvVar != "" {
			envVars[ag.ModelEnvVar] = settings.Model
		}

		// Add per-agent env vars
		for k, v := range settings.EnvVars {
			envVars[k] = v
		}

		// Claude-specific: map model aliases to env vars
		if agentName == "claude" {
			if v, ok := settings.ModelAliases["sonnet"]; ok && v != "" {
				envVars["ANTHROPIC_DEFAULT_SONNET_MODEL"] = v
			}
			if v, ok := settings.ModelAliases["opus"]; ok && v != "" {
				envVars["ANTHROPIC_DEFAULT_OPUS_MODEL"] = v
			}
			if v, ok := settings.ModelAliases["haiku"]; ok && v != "" {
				envVars["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = v
			}
		}

		if len(envVars) > 0 {
			l.SetEnv(envVars)
		}
	}

	fmt.Println(ui.Info(fmt.Sprintf("Launching %s...", ag.Name)))
	if err := l.Launch(); err != nil {
		return fmt.Errorf("failed to launch %s: %w", ag.Name, err)
	}

	return nil
}

func init() {
	VarCodeCmd.SetOut(os.Stdout)
	VarCodeCmd.SetErr(os.Stderr)
}
