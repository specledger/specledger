package commands

import (
	"fmt"
	"strings"

	"github.com/specledger/specledger/pkg/cli/config"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage agent configuration profiles",
	Long: `Manage named profiles that bundle multiple agent configuration values.

Profiles allow you to quickly switch between different AI provider configurations
(e.g., "work" profile for corporate gateway, "personal" for default API).

The active profile's values are applied as a layer in the config merge precedence:
personal-local > team-local > profile > global > default

Commands:
  sl config profile create <name>           Create a new profile
  sl config profile use <name>              Activate a profile
  sl config profile use --none              Deactivate all profiles
  sl config profile list                    List all profiles
  sl config profile delete <name>           Delete a profile

Examples:
  sl config profile create work
  sl config profile use work
  sl config profile list
  sl config profile delete work`,
}

var profileCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new profile",
	Long:  `Create a new named profile with default agent configuration.`,
	Example: `  sl config profile create work
  sl config profile create local`,
	Args: cobra.ExactArgs(1),
	RunE: runProfileCreate,
}

var profileUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Activate a profile",
	Long: `Activate a profile by name. Use --none to deactivate all profiles.
When a profile is active, its values are merged into the agent configuration.`,
	Example: `  sl config profile use work
  sl config profile use --none`,
	Args: cobra.MaximumNArgs(1),
	RunE: runProfileUse,
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Long:  `List all named profiles with their configuration summary.`,
	Example: `  sl config profile list`,
	Args: cobra.NoArgs,
	RunE: runProfileList,
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a profile",
	Long:  `Delete a named profile. Cannot delete the currently active profile.`,
	Example: `  sl config profile delete work`,
	Args: cobra.ExactArgs(1),
	RunE: runProfileDelete,
}

var profileUseNoneFlag bool

func init() {
	VarConfigCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileUseCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileDeleteCmd)

	profileUseCmd.Flags().BoolVar(&profileUseNoneFlag, "none", false, "Deactivate all profiles")
}

func runProfileCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.CreateProfile(name, config.DefaultAgentConfig()); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("Created profile '%s'", name))
	return nil
}

func runProfileUse(cmd *cobra.Command, args []string) error {
	if profileUseNoneFlag {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		cfg.SetActiveProfile("")
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		ui.PrintSuccess("Deactivated all profiles")
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("profile name required (or use --none to deactivate)")
	}

	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.SetActiveProfile(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("Activated profile '%s'", name))
	return nil
}

func runProfileList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	profiles := cfg.ListProfiles()
	if len(profiles) == 0 {
		fmt.Println("No profiles configured.")
		fmt.Println()
		fmt.Println("Create a profile with:")
		fmt.Println("  sl config profile create <name>")
		return nil
	}

	fmt.Println("Profiles:")
	fmt.Println()

	for _, name := range profiles {
		active := ""
		if name == cfg.ActiveProfile {
			active = " (active)"
		}

		profile, _ := cfg.GetProfile(name)
		summary := getProfileSummary(profile)

		fmt.Printf("  %s%s\n", name, active)
		if summary != "" {
			fmt.Printf("    %s\n", summary)
		}
	}

	if cfg.ActiveProfile == "" {
		fmt.Println()
		fmt.Println("No active profile. Use: sl config profile use <name>")
	}

	return nil
}

func runProfileDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.DeleteProfile(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("Deleted profile '%s'", name))
	return nil
}

func getProfileSummary(profile *config.AgentConfig) string {
	if profile == nil {
		return ""
	}

	var parts []string

	if profile.BaseURL != "" {
		parts = append(parts, profile.BaseURL)
	}
	if profile.Model != "" {
		parts = append(parts, fmt.Sprintf("model=%s", profile.Model))
	}
	if profile.Provider != "" && profile.Provider != "anthropic" {
		parts = append(parts, fmt.Sprintf("provider=%s", profile.Provider))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, ", ")
}
