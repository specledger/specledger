package commands

import (
	"fmt"
	"strings"

	"github.com/specledger/specledger/pkg/cli/config"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/spf13/cobra"
)

var (
	configGlobalFlag   bool
	configPersonalFlag bool
)

var VarConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long: `Manage SpecLedger CLI configuration.

Configuration is stored in a three-tier hierarchy:
  - Global: ~/.specledger/config.yaml (user-wide defaults)
  - Team-local: specledger/specledger.yaml (project-level, git-tracked)
  - Personal-local: specledger/specledger.local.yaml (project-level, gitignored)

Precedence: personal-local > team-local > global > default

Commands:
  sl config set <key> <value>   Set a configuration value
  sl config get <key>           Get a configuration value
  sl config show                Show all configuration values
  sl config unset <key>         Remove a configuration value

Per-Agent Keys (namespaced):
  agent.<name>.api_key          API key for the agent
  agent.<name>.base_url         Custom API endpoint
  agent.<name>.model            Model to use
  agent.<name>.arguments        CLI arguments to pass
  agent.<name>.env.<VAR>        Environment variable

Claude-Specific Keys:
  agent.claude.model_aliases.sonnet   Sonnet model alias
  agent.claude.model_aliases.opus     Opus model alias
  agent.claude.model_aliases.haiku    Haiku model alias

Examples:
  sl config set agent.claude.api_key sk-ant-xxx
  sl config set agent.claude.model claude-sonnet-4-20250514
  sl config set agent.codex.api_key sk-xxx
  sl config set --global agent.default claude
  sl config set agent.claude.arguments "--dangerously-skip-permissions"
  sl config show`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Keys are namespaced per agent (e.g., agent.claude.api_key, agent.claude.model).
Use --global to set in user-wide config.
Use --personal to set in gitignored personal config (recommended for secrets).`,
	Example: `  sl config set agent.claude.api_key sk-ant-xxx
  sl config set --global agent.default claude
  sl config set --personal agent.claude.api_key sk-ant-xxx
  sl config set agent.claude.model claude-sonnet-4-20250514
  sl config set agent.codex.api_key sk-xxx
  sl config set agent.claude.arguments "--dangerously-skip-permissions"`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Filter out --global and --personal flags from args
		filtered := filterConfigFlags(args)
		if len(filtered) != 2 {
			return fmt.Errorf("accepts 2 arg(s), received %d", len(filtered))
		}
		return nil
	},
	RunE:               runConfigSet,
	DisableFlagParsing: true, // Allow values that look like flags (e.g., --dangerously-skip-permissions)
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long:  `Get a single configuration value with its scope indicator.`,
	Example: `  sl config get agent.base-url
  sl config get agent.model`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

var configShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show all configuration values",
	Long:    `Show all configuration values with their scope indicators.`,
	Example: `  sl config show`,
	Args:    cobra.NoArgs,
	RunE:    runConfigShow,
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Remove a configuration value",
	Long: `Remove a configuration value.

The value will fall back to the next layer in the precedence hierarchy.`,
	Example: `  sl config unset agent.base-url
  sl config unset --personal agent.auth-token`,
	Args: func(cmd *cobra.Command, args []string) error {
		filtered := filterConfigFlags(args)
		if len(filtered) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(filtered))
		}
		return nil
	},
	RunE:               runConfigUnset,
	DisableFlagParsing: true, // Allow keys that look like flags
}

func init() {
	VarConfigCmd.AddCommand(configSetCmd)
	VarConfigCmd.AddCommand(configGetCmd)
	VarConfigCmd.AddCommand(configShowCmd)
	VarConfigCmd.AddCommand(configUnsetCmd)

	// Flags are manually parsed in runConfigSet/runConfigUnset due to DisableFlagParsing
}

// filterConfigFlags extracts --global and --personal flags from args and sets the flag variables.
// Returns the remaining positional arguments.
func filterConfigFlags(args []string) []string {
	var filtered []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--global" {
			configGlobalFlag = true
		} else if args[i] == "--personal" {
			configPersonalFlag = true
		} else {
			filtered = append(filtered, args[i])
		}
	}
	return filtered
}

// parseArgumentsValue parses a command-line arguments string into a slice.
// Handles quoted strings and space-separated arguments.
func parseArgumentsValue(s string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range s {
		switch {
		case !inQuotes && (r == '"' || r == '\''):
			inQuotes = true
			quoteChar = r
		case inQuotes && r == quoteChar:
			inQuotes = false
			quoteChar = 0
		case !inQuotes && r == ' ':
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	// Reset flags (in case of repeated calls in tests)
	configGlobalFlag = false
	configPersonalFlag = false

	// Filter out --global and --personal flags
	args = filterConfigFlags(args)
	if len(args) != 2 {
		return fmt.Errorf("accepts 2 arg(s), received %d", len(args))
	}

	key := args[0]
	value := args[1]

	keyDef, err := config.LookupKey(key)
	if err != nil {
		similar := config.GetRegistry().FindSimilar(key)
		if len(similar) > 0 {
			return fmt.Errorf("%w\n\nDid you mean one of these?\n  %s", err, strings.Join(similar, "\n  "))
		}
		return err
	}

	scope := determineScope()

	if keyDef.Sensitive && scope == "local" && !configPersonalFlag {
		ui.PrintWarning("Storing sensitive value in git-tracked config. Consider using --personal to store in gitignored file.")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := setAgentConfigValue(cfg, key, value); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("Set %s = %s [%s]", key, maskIfSensitive(value, keyDef.Sensitive), scope))
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	if _, err := config.LookupKey(key); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	value, scope := getAgentConfigValue(cfg, key)
	if value == "" {
		fmt.Printf("%s: (not set)\n", key)
		return nil
	}

	keyDef, _ := config.LookupKey(key)
	displayValue := value
	if keyDef != nil && keyDef.Sensitive {
		displayValue = maskSensitive(value)
	}

	fmt.Printf("%s = %s [%s]\n", key, displayValue, scope)
	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Agent Configuration")
	fmt.Println()

	// Show default agent
	defaultAgent := "claude"
	if cfg.Agents != nil && cfg.Agents.Default != "" {
		defaultAgent = cfg.Agents.Default
	}
	fmt.Printf("  Default Agent: %s\n\n", defaultAgent)

	// Show per-agent configuration
	agentNames := []string{"claude", "opencode", "github-copilot", "codex"}
	for _, agentName := range agentNames {
		var settings *config.AgentSettings
		if cfg.Agents != nil {
			settings = cfg.Agents.GetAgentSettings(agentName)
		}

		if settings == nil {
			continue
		}

		fmt.Printf("  %s\n", agentName)
		if settings.APIKey != "" {
			fmt.Printf("    %-23s %-35s [local]\n", "api_key", maskSensitive(settings.APIKey))
		}
		if settings.BaseURL != "" {
			fmt.Printf("    %-23s %-35s [local]\n", "base_url", settings.BaseURL)
		}
		if settings.Model != "" {
			fmt.Printf("    %-23s %-35s [local]\n", "model", settings.Model)
		}
		if len(settings.Arguments) > 0 {
			fmt.Printf("    %-23s %-35s [local]\n", "arguments", strings.Join(settings.Arguments, " "))
		}
		if len(settings.Env) > 0 {
			fmt.Printf("    %-23s\n", "env")
			for envKey, envValue := range settings.Env {
				fmt.Printf("      %-21s %-35s [local]\n", envKey, envValue)
			}
		}

		// Claude-specific: model aliases
		if agentName == "claude" && cfg.Agents.Claude != nil && cfg.Agents.Claude.ModelAliases != nil {
			ma := cfg.Agents.Claude.ModelAliases
			if ma.Sonnet != "" || ma.Opus != "" || ma.Haiku != "" {
				fmt.Printf("    %-23s\n", "model_aliases")
				if ma.Sonnet != "" {
					fmt.Printf("      %-21s %-35s [local]\n", "sonnet", ma.Sonnet)
				}
				if ma.Opus != "" {
					fmt.Printf("      %-21s %-35s [local]\n", "opus", ma.Opus)
				}
				if ma.Haiku != "" {
					fmt.Printf("      %-21s %-35s [local]\n", "haiku", ma.Haiku)
				}
			}
		}
		fmt.Println()
	}

	// Show legacy agent config if present (for backward compatibility)
	if cfg.Agent != nil && (cfg.Agent.APIKey != "" || cfg.Agent.Model != "") {
		fmt.Println("  Legacy (deprecated)")
		if cfg.Agent.APIKey != "" {
			fmt.Printf("    %-23s %-35s [local]\n", "api-key", maskSensitive(cfg.Agent.APIKey))
		}
		if cfg.Agent.Model != "" {
			fmt.Printf("    %-23s %-35s [local]\n", "model", cfg.Agent.Model)
		}
		fmt.Println()
	}

	fmt.Println("General")
	fmt.Printf("    %-25s %-35s [global]\n", "tui_enabled", fmt.Sprintf("%v", cfg.TUIEnabled))
	fmt.Printf("    %-25s %-35s [global]\n", "log_level", cfg.LogLevel)

	return nil
}

func runConfigUnset(cmd *cobra.Command, args []string) error {
	// Reset flags (in case of repeated calls in tests)
	configGlobalFlag = false
	configPersonalFlag = false

	// Filter out --global and --personal flags
	args = filterConfigFlags(args)
	if len(args) != 1 {
		return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
	}

	key := args[0]

	if _, err := config.LookupKey(key); err != nil {
		return err
	}

	scope := determineScope()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := unsetAgentConfigValue(cfg, key); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("Unset %s [%s]", key, scope))
	return nil
}

func determineScope() string {
	if configGlobalFlag {
		return "global"
	}
	if configPersonalFlag {
		return "personal"
	}
	return "local"
}

func setAgentConfigValue(cfg *config.Config, key, value string) error {
	// Initialize Agents if nil
	if cfg.Agents == nil {
		cfg.Agents = config.NewConfigAgents()
	}

	// Handle agent.default
	if key == "agent.default" {
		cfg.Agents.Default = value
		return nil
	}

	// Handle per-agent keys: agent.<name>.<field>
	if strings.HasPrefix(key, "agent.") {
		parts := strings.Split(key, ".")
		if len(parts) >= 3 {
			agentName := parts[1]

			// Validate agent name
			if agentName != "claude" && agentName != "opencode" && agentName != "github-copilot" && agentName != "codex" {
				return fmt.Errorf("unknown agent name: %s (valid: claude, opencode, github-copilot, codex)", agentName)
			}

			// Handle agent.<name>.api_key
			if len(parts) == 3 && parts[2] == "api_key" {
				settings := cfg.Agents.GetOrCreateAgentSettings(agentName)
				settings.APIKey = value
				return nil
			}

			// Handle agent.<name>.base_url
			if len(parts) == 3 && parts[2] == "base_url" {
				settings := cfg.Agents.GetOrCreateAgentSettings(agentName)
				settings.BaseURL = value
				return nil
			}

			// Handle agent.<name>.model
			if len(parts) == 3 && parts[2] == "model" {
				settings := cfg.Agents.GetOrCreateAgentSettings(agentName)
				settings.Model = value
				return nil
			}

			// Handle agent.<name>.arguments
			if len(parts) == 3 && parts[2] == "arguments" {
				settings := cfg.Agents.GetOrCreateAgentSettings(agentName)
				settings.Arguments = parseArgumentsValue(value)
				return nil
			}

			// Handle agent.<name>.env.<VAR>
			if len(parts) == 4 && parts[2] == "env" {
				envVar := parts[3]
				settings := cfg.Agents.GetOrCreateAgentSettings(agentName)
				if settings.Env == nil {
					settings.Env = make(map[string]string)
				}
				settings.Env[envVar] = value
				return nil
			}

			// Handle agent.claude.model_aliases.<alias>
			if agentName == "claude" && len(parts) == 4 && parts[2] == "model_aliases" {
				alias := parts[3]
				if alias != "sonnet" && alias != "opus" && alias != "haiku" {
					return fmt.Errorf("unknown model alias: %s (valid: sonnet, opus, haiku)", alias)
				}
				if cfg.Agents.Claude == nil {
					cfg.Agents.Claude = config.NewClaudeSettings()
				}
				if cfg.Agents.Claude.ModelAliases == nil {
					cfg.Agents.Claude.ModelAliases = &config.ClaudeModelAliases{}
				}
				switch alias {
				case "sonnet":
					cfg.Agents.Claude.ModelAliases.Sonnet = value
				case "opus":
					cfg.Agents.Claude.ModelAliases.Opus = value
				case "haiku":
					cfg.Agents.Claude.ModelAliases.Haiku = value
				}
				return nil
			}
		}

		// Handle legacy agent.env.<VAR> (maps to claude for backward compat)
		if strings.HasPrefix(key, "agent.env.") {
			envKey, _ := strings.CutPrefix(key, "agent.env.")
			if cfg.Agent == nil {
				cfg.Agent = config.DefaultAgentConfig()
			}
			if cfg.Agent.Env == nil {
				cfg.Agent.Env = make(map[string]string)
			}
			cfg.Agent.Env[envKey] = value
			return nil
		}
	}

	// Handle active-profile
	if key == "active-profile" {
		cfg.ActiveProfile = value
		return nil
	}

	return fmt.Errorf("unknown config key: %s", key)
}

func getAgentConfigValue(cfg *config.Config, key string) (string, string) {
	scope := "local"

	// Handle agent.default
	if key == "agent.default" {
		if cfg.Agents != nil && cfg.Agents.Default != "" {
			return cfg.Agents.Default, scope
		}
		return "claude", "default"
	}

	// Handle per-agent keys: agent.<name>.<field>
	if strings.HasPrefix(key, "agent.") {
		parts := strings.Split(key, ".")
		if len(parts) >= 3 {
			agentName := parts[1]

			// Get agent settings
			var settings *config.AgentSettings
			if cfg.Agents != nil {
				settings = cfg.Agents.GetAgentSettings(agentName)
			}

			// Handle agent.<name>.api_key
			if len(parts) == 3 && parts[2] == "api_key" {
				if settings != nil && settings.APIKey != "" {
					return settings.APIKey, scope
				}
				return "", "default"
			}

			// Handle agent.<name>.base_url
			if len(parts) == 3 && parts[2] == "base_url" {
				if settings != nil && settings.BaseURL != "" {
					return settings.BaseURL, scope
				}
				return "", "default"
			}

			// Handle agent.<name>.model
			if len(parts) == 3 && parts[2] == "model" {
				if settings != nil && settings.Model != "" {
					return settings.Model, scope
				}
				return "", "default"
			}

			// Handle agent.<name>.arguments
			if len(parts) == 3 && parts[2] == "arguments" {
				if settings != nil && len(settings.Arguments) > 0 {
					return strings.Join(settings.Arguments, " "), scope
				}
				return "", "default"
			}

			// Handle agent.<name>.env.<VAR>
			if len(parts) == 4 && parts[2] == "env" {
				envVar := parts[3]
				if settings != nil && settings.Env != nil {
					if v, ok := settings.Env[envVar]; ok {
						return v, scope
					}
				}
				return "", "default"
			}

			// Handle agent.claude.model_aliases.<alias>
			if agentName == "claude" && len(parts) == 4 && parts[2] == "model_aliases" {
				alias := parts[3]
				if cfg.Agents != nil && cfg.Agents.Claude != nil && cfg.Agents.Claude.ModelAliases != nil {
					switch alias {
					case "sonnet":
						if cfg.Agents.Claude.ModelAliases.Sonnet != "" {
							return cfg.Agents.Claude.ModelAliases.Sonnet, scope
						}
					case "opus":
						if cfg.Agents.Claude.ModelAliases.Opus != "" {
							return cfg.Agents.Claude.ModelAliases.Opus, scope
						}
					case "haiku":
						if cfg.Agents.Claude.ModelAliases.Haiku != "" {
							return cfg.Agents.Claude.ModelAliases.Haiku, scope
						}
					}
				}
				return "", "default"
			}
		}

		// Handle legacy agent.env.<VAR> (maps to claude for backward compat)
		if strings.HasPrefix(key, "agent.env.") {
			envKey, _ := strings.CutPrefix(key, "agent.env.")
			if cfg.Agent != nil && cfg.Agent.Env != nil {
				return cfg.Agent.Env[envKey], scope
			}
		}
	}

	// Handle active-profile
	if key == "active-profile" {
		return cfg.ActiveProfile, scope
	}

	return "", "default"
}

func unsetAgentConfigValue(cfg *config.Config, key string) error {
	// Handle agent.default
	if key == "agent.default" {
		if cfg.Agents != nil {
			cfg.Agents.Default = ""
		}
		return nil
	}

	// Handle per-agent keys: agent.<name>.<field>
	if strings.HasPrefix(key, "agent.") {
		parts := strings.Split(key, ".")
		if len(parts) >= 3 {
			agentName := parts[1]

			// Get agent settings
			var settings *config.AgentSettings
			if cfg.Agents != nil {
				settings = cfg.Agents.GetAgentSettings(agentName)
			}

			// Handle agent.<name>.api_key
			if len(parts) == 3 && parts[2] == "api_key" {
				if settings != nil {
					settings.APIKey = ""
				}
				return nil
			}

			// Handle agent.<name>.base_url
			if len(parts) == 3 && parts[2] == "base_url" {
				if settings != nil {
					settings.BaseURL = ""
				}
				return nil
			}

			// Handle agent.<name>.model
			if len(parts) == 3 && parts[2] == "model" {
				if settings != nil {
					settings.Model = ""
				}
				return nil
			}

			// Handle agent.<name>.arguments
			if len(parts) == 3 && parts[2] == "arguments" {
				if settings != nil {
					settings.Arguments = nil
				}
				return nil
			}

			// Handle agent.<name>.env.<VAR>
			if len(parts) == 4 && parts[2] == "env" {
				envVar := parts[3]
				if settings != nil && settings.Env != nil {
					delete(settings.Env, envVar)
				}
				return nil
			}

			// Handle agent.claude.model_aliases.<alias>
			if agentName == "claude" && len(parts) == 4 && parts[2] == "model_aliases" {
				alias := parts[3]
				if cfg.Agents != nil && cfg.Agents.Claude != nil && cfg.Agents.Claude.ModelAliases != nil {
					switch alias {
					case "sonnet":
						cfg.Agents.Claude.ModelAliases.Sonnet = ""
					case "opus":
						cfg.Agents.Claude.ModelAliases.Opus = ""
					case "haiku":
						cfg.Agents.Claude.ModelAliases.Haiku = ""
					}
				}
				return nil
			}
		}

		// Handle legacy agent.env.<VAR>
		if strings.HasPrefix(key, "agent.env.") {
			envKey, _ := strings.CutPrefix(key, "agent.env.")
			if cfg.Agent != nil && cfg.Agent.Env != nil {
				delete(cfg.Agent.Env, envKey)
			}
			return nil
		}
	}

	// Handle active-profile
	if key == "active-profile" {
		cfg.ActiveProfile = ""
		return nil
	}

	return fmt.Errorf("unknown config key: %s", key)
}

func maskSensitive(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return "****" + value[len(value)-4:]
}

func maskIfSensitive(value string, sensitive bool) string {
	if sensitive {
		return maskSensitive(value)
	}
	return value
}
