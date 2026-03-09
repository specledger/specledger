package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/specledger/specledger/pkg/cli/context"
	"github.com/specledger/specledger/pkg/cli/spec"
	"github.com/spf13/cobra"
)

type ContextUpdateOutput struct {
	UpdatedFiles []string `json:"UPDATED_FILES"`
}

var contextUpdateCmd = &cobra.Command{
	Use:   "update [agent]",
	Short: "Update agent context files with plan metadata",
	Long: `Update AI agent context files with Technical Context from plan.md.

This command reads the Technical Context section from plan.md and updates
the specified agent file (e.g., CLAUDE.md, GEMINI.md) with the metadata.

The update preserves any manual additions between the MANUAL ADDITIONS markers
and deduplicates entries to avoid repetition.

Supported agents:
  claude, gemini, copilot, cursor, qwen, windsurf, kilocode, auggie, roo,
  codebuddy, qoder, shai, amazonq, ibmbob, opencode, codex

Examples:
  sl context update claude              # Update CLAUDE.md
  sl context update gemini --json       # Update GEMINI.md with JSON output
  sl context update copilot             # Update .github/agents/copilot-instructions.md
  sl context update --agent claude      # Alternative flag syntax`,
	Args: cobra.MaximumNArgs(1),
	RunE: runContextUpdate,
}

func init() {
	VarContextCmd.AddCommand(contextUpdateCmd)

	contextUpdateCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	contextUpdateCmd.Flags().String("agent", "", "Agent type (claude, gemini, copilot, etc.)")
	contextUpdateCmd.Flags().String("spec", "", "Override feature spec name (bypasses detection)")
}

func runContextUpdate(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	agentFlag, _ := cmd.Flags().GetString("agent")
	specOverride, _ := cmd.Flags().GetString("spec")

	agentType := agentFlag
	if len(args) > 0 && agentType == "" {
		agentType = args[0]
	}

	if agentType == "" {
		agentType = "claude"
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	opts := spec.DetectionOptions{
		SpecOverride: specOverride,
	}

	ctx, err := spec.DetectFeatureContextWithOptions(workDir, opts)
	if err != nil {
		return fmt.Errorf("failed to detect feature context: %w", err)
	}

	if !spec.FileExists(ctx.PlanFile) {
		return fmt.Errorf("plan.md not found at: %s", ctx.PlanFile)
	}

	techCtx, err := context.ParseTechnicalContext(ctx.PlanFile)
	if err != nil {
		return fmt.Errorf("failed to parse plan.md: %w", err)
	}

	updater := context.NewAgentUpdater(agentType, ctx.RepoRoot)

	if err := updater.Update(techCtx); err != nil {
		return fmt.Errorf("failed to update agent file: %w", err)
	}

	output := ContextUpdateOutput{
		UpdatedFiles: []string{updater.FilePath},
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	} else {
		fmt.Printf("Updated agent file: %s\n", updater.FilePath)
		fmt.Printf("Agent type: %s\n", agentType)
		fmt.Printf("Feature: %s\n", ctx.Branch)
	}

	return nil
}

func NewContextUpdateCmd() *cobra.Command {
	return contextUpdateCmd
}
