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

Examples:
  sl config set agent.base-url https://api.example.com
  sl config set --global agent.model sonnet
  sl config set --personal agent.auth-token sk-xxx
  sl config show
  sl config unset agent.base-url`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Keys are namespaced (e.g., agent.base-url, agent.model.sonnet).
Use --global to set in user-wide config.
Use --personal to set in gitignored personal config (recommended for secrets).`,
	Example: `  sl config set agent.base-url https://api.example.com
  sl config set --global agent.model sonnet
  sl config set --personal agent.auth-token sk-xxx`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
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
	Args: cobra.ExactArgs(1),
	RunE: runConfigUnset,
}

func init() {
	VarConfigCmd.AddCommand(configSetCmd)
	VarConfigCmd.AddCommand(configGetCmd)
	VarConfigCmd.AddCommand(configShowCmd)
	VarConfigCmd.AddCommand(configUnsetCmd)

	configSetCmd.Flags().BoolVar(&configGlobalFlag, "global", false, "Set in global config (~/.specledger/config.yaml)")
	configSetCmd.Flags().BoolVar(&configPersonalFlag, "personal", false, "Set in personal-local config (gitignored)")
	configUnsetCmd.Flags().BoolVar(&configGlobalFlag, "global", false, "Unset from global config")
	configUnsetCmd.Flags().BoolVar(&configPersonalFlag, "personal", false, "Unset from personal-local config")
}

func runConfigSet(cmd *cobra.Command, args []string) error {
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

	keysByCategory := config.GetRegistry().ListByCategory()
	for _, category := range []string{"Provider", "Models", "Launch Flags", "Environment", "Profiles"} {
		keys := keysByCategory[category]
		if len(keys) == 0 && category != "Environment" {
			continue
		}

		fmt.Printf("  %s\n", category)

		if category == "Environment" && cfg.Agent != nil && len(cfg.Agent.Env) > 0 {
			fmt.Printf("    %-25s\n", "agent.env")
			for envKey, envValue := range cfg.Agent.Env {
				fmt.Printf("      %-23s %-35s [local]\n", envKey, envValue)
			}
		}

		for _, keyDef := range keys {
			if keyDef.Key == "agent.env" {
				continue
			}
			value, scope := getAgentConfigValue(cfg, keyDef.Key)
			if value == "" {
				continue
			}

			displayValue := value
			if keyDef.Sensitive {
				displayValue = maskSensitive(value)
			}

			fmt.Printf("    %-25s %-35s [%s]\n", keyDef.Key, displayValue, scope)
		}
		fmt.Println()
	}

	if cfg.ActiveProfile != "" {
		fmt.Printf("  Active Profile: %s\n\n", cfg.ActiveProfile)
	}

	fmt.Println("General")
	fmt.Printf("    %-25s %-35s [global]\n", "tui_enabled", fmt.Sprintf("%v", cfg.TUIEnabled))
	fmt.Printf("    %-25s %-35s [global]\n", "log_level", cfg.LogLevel)

	return nil
}

func runConfigUnset(cmd *cobra.Command, args []string) error {
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
	if cfg.Agent == nil {
		cfg.Agent = config.DefaultAgentConfig()
	}

	switch key {
	case "agent.base-url":
		cfg.Agent.BaseURL = value
	case "agent.auth-token":
		cfg.Agent.AuthToken = value
	case "agent.api-key":
		cfg.Agent.APIKey = value
	case "agent.model":
		cfg.Agent.Model = value
	case "agent.model.sonnet":
		cfg.Agent.ModelSonnet = value
	case "agent.model.opus":
		cfg.Agent.ModelOpus = value
	case "agent.model.haiku":
		cfg.Agent.ModelHaiku = value
	case "agent.subagent-model":
		cfg.Agent.SubagentModel = value
	case "agent.provider":
		cfg.Agent.Provider = value
	case "agent.permission-mode":
		cfg.Agent.PermissionMode = value
	case "agent.skip-permissions":
		cfg.Agent.SkipPermissions = value == "true"
	case "agent.effort":
		cfg.Agent.Effort = value
	case "active-profile":
		cfg.ActiveProfile = value
	default:
		if strings.HasPrefix(key, "agent.env.") {
			envKey := strings.TrimPrefix(key, "agent.env.")
			if cfg.Agent.Env == nil {
				cfg.Agent.Env = make(map[string]string)
			}
			cfg.Agent.Env[envKey] = value
			return nil
		}
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func getAgentConfigValue(cfg *config.Config, key string) (string, string) {
	if cfg.Agent == nil {
		return "", "default"
	}

	scope := "local"

	switch key {
	case "agent.base-url":
		return cfg.Agent.BaseURL, scope
	case "agent.auth-token":
		return cfg.Agent.AuthToken, scope
	case "agent.api-key":
		return cfg.Agent.APIKey, scope
	case "agent.model":
		return cfg.Agent.Model, scope
	case "agent.model.sonnet":
		return cfg.Agent.ModelSonnet, scope
	case "agent.model.opus":
		return cfg.Agent.ModelOpus, scope
	case "agent.model.haiku":
		return cfg.Agent.ModelHaiku, scope
	case "agent.subagent-model":
		return cfg.Agent.SubagentModel, scope
	case "agent.provider":
		if cfg.Agent.Provider == "" {
			return "anthropic", "default"
		}
		return cfg.Agent.Provider, scope
	case "agent.permission-mode":
		if cfg.Agent.PermissionMode == "" {
			return "default", "default"
		}
		return cfg.Agent.PermissionMode, scope
	case "agent.skip-permissions":
		return fmt.Sprintf("%v", cfg.Agent.SkipPermissions), scope
	case "agent.effort":
		return cfg.Agent.Effort, scope
	case "active-profile":
		return cfg.ActiveProfile, scope
	default:
		if strings.HasPrefix(key, "agent.env.") {
			envKey := strings.TrimPrefix(key, "agent.env.")
			if cfg.Agent.Env != nil {
				return cfg.Agent.Env[envKey], scope
			}
		}
	}
	return "", "default"
}

func unsetAgentConfigValue(cfg *config.Config, key string) error {
	if cfg.Agent == nil {
		return fmt.Errorf("no config value set for: %s", key)
	}

	switch key {
	case "agent.base-url":
		cfg.Agent.BaseURL = ""
	case "agent.auth-token":
		cfg.Agent.AuthToken = ""
	case "agent.api-key":
		cfg.Agent.APIKey = ""
	case "agent.model":
		cfg.Agent.Model = ""
	case "agent.model.sonnet":
		cfg.Agent.ModelSonnet = ""
	case "agent.model.opus":
		cfg.Agent.ModelOpus = ""
	case "agent.model.haiku":
		cfg.Agent.ModelHaiku = ""
	case "agent.subagent-model":
		cfg.Agent.SubagentModel = ""
	case "agent.provider":
		cfg.Agent.Provider = ""
	case "agent.permission-mode":
		cfg.Agent.PermissionMode = ""
	case "agent.skip-permissions":
		cfg.Agent.SkipPermissions = false
	case "agent.effort":
		cfg.Agent.Effort = ""
	case "active-profile":
		cfg.ActiveProfile = ""
	default:
		if strings.HasPrefix(key, "agent.env.") {
			envKey := strings.TrimPrefix(key, "agent.env.")
			delete(cfg.Agent.Env, envKey)
			return nil
		}
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
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
