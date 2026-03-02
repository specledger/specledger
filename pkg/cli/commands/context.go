package commands

import (
	"github.com/spf13/cobra"
)

var VarContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage agent context files",
	Long: `Manage AI agent context files with plan metadata.

Commands:
  update  Update agent files with Technical Context from plan.md

Examples:
  sl context update claude      # Update CLAUDE.md
  sl context update gemini      # Update GEMINI.md
  sl context update copilot     # Update .github/agents/copilot-instructions.md
  sl context update claude --json  # Output as JSON`,
}

func NewContextCmd() *cobra.Command {
	return VarContextCmd
}
