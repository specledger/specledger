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
	agentName := "claude"
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

	arguments := config.GetAgentArguments(ag.Command)
	if len(arguments) > 0 {
		l.SetFlags(arguments)
	}

	resolved := config.ResolveAgentConfig()
	envVars := resolved.GetEnvVars()
	if len(envVars) > 0 {
		l.SetEnv(envVars)
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
