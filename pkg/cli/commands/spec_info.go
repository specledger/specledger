package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/specledger/specledger/pkg/cli/spec"
	"github.com/spf13/cobra"
)

type SpecInfoOutput struct {
	FeatureDir    string   `json:"FEATURE_DIR"`
	Branch        string   `json:"BRANCH"`
	FeatureSpec   string   `json:"FEATURE_SPEC"`
	AvailableDocs []string `json:"AVAILABLE_DOCS,omitempty"`
}

var specInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get feature paths and prerequisite validation",
	Long: `Get feature context information including paths and available documentation.

This command detects the current feature context from the git branch and
returns information about the feature directory, spec file, and other paths.

Output includes:
  FEATURE_DIR     - Path to the feature directory
  BRANCH          - Current feature branch name
  FEATURE_SPEC    - Path to the spec.md file
  AVAILABLE_DOCS  - List of other documentation files (research.md, etc.)

Examples:
  sl spec info --json                    # Output as JSON
  sl spec info --require-plan            # Error if plan.md missing
  sl spec info --require-tasks           # Error if tasks.md missing
  sl spec info --include-tasks           # Include tasks.md in AVAILABLE_DOCS
  sl spec info --paths-only              # Output minimal paths only`,
	RunE: runSpecInfo,
}

func init() {
	VarSpecCmd.AddCommand(specInfoCmd)

	specInfoCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	specInfoCmd.Flags().Bool("require-plan", false, "Error if plan.md does not exist")
	specInfoCmd.Flags().Bool("require-tasks", false, "Error if tasks.md does not exist")
	specInfoCmd.Flags().Bool("include-tasks", false, "Include tasks.md in AVAILABLE_DOCS")
	specInfoCmd.Flags().Bool("paths-only", false, "Output minimal paths only (no doc discovery)")
	specInfoCmd.Flags().String("spec", "", "Override feature spec name (bypasses detection)")
}

func runSpecInfo(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	requirePlan, _ := cmd.Flags().GetBool("require-plan")
	requireTasks, _ := cmd.Flags().GetBool("require-tasks")
	includeTasks, _ := cmd.Flags().GetBool("include-tasks")
	pathsOnly, _ := cmd.Flags().GetBool("paths-only")
	specOverride, _ := cmd.Flags().GetString("spec")

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

	if requirePlan {
		if !spec.FileExists(ctx.PlanFile) {
			return fmt.Errorf("plan.md not found at: %s", ctx.PlanFile)
		}
	}

	if requireTasks {
		if !spec.FileExists(ctx.TasksFile) {
			return fmt.Errorf("tasks.md not found at: %s", ctx.TasksFile)
		}
	}

	output := SpecInfoOutput{
		FeatureDir:  ctx.FeatureDir,
		Branch:      ctx.Branch,
		FeatureSpec: ctx.SpecFile,
	}

	if !pathsOnly {
		docs, err := spec.DiscoverDocs(ctx.FeatureDir)
		if err != nil {
			if !jsonOutput {
				fmt.Fprintf(os.Stderr, "Warning: failed to discover docs: %v\n", err)
			}
			docs = []string{}
		}

		if includeTasks && spec.FileExists(ctx.TasksFile) {
			docs = append(docs, "tasks.md")
		}

		output.AvailableDocs = docs
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	} else {
		fmt.Printf("FEATURE_DIR: %s\n", output.FeatureDir)
		fmt.Printf("BRANCH: %s\n", output.Branch)
		fmt.Printf("FEATURE_SPEC: %s\n", output.FeatureSpec)

		if len(output.AvailableDocs) > 0 {
			fmt.Printf("AVAILABLE_DOCS: %v\n", output.AvailableDocs)
		}
	}

	return nil
}

func NewSpecInfoCmd() *cobra.Command {
	return specInfoCmd
}
