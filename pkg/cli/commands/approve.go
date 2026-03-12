package commands

import (
	"fmt"
	"os"

	"github.com/specledger/specledger/pkg/cli/spec"
	"github.com/spf13/cobra"
)

var approveSpecFlag string

var VarApproveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Mark a feature spec as approved for implementation",
	Long: `Mark a feature spec as approved, gating it for push-triggered implementation.

Validates that all required artifacts (spec.md, plan.md, tasks.md) exist and are
non-empty before setting **Status**: Approved in spec.md.

Examples:
  sl approve                    # Auto-detect spec from current branch
  sl approve --spec 127-feature # Approve a specific spec`,
	RunE: runApprove,
}

func init() {
	VarApproveCmd.Flags().StringVar(&approveSpecFlag, "spec", "", "Feature spec context to approve (default: auto-detect from branch)")
}

func runApprove(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	opts := spec.DetectionOptions{
		SpecOverride: approveSpecFlag,
	}

	ctx, err := spec.DetectFeatureContextWithOptions(workDir, opts)
	if err != nil {
		return fmt.Errorf("spec not found: %w", err)
	}

	// Validate required artifacts exist and are non-empty
	var missing []string
	artifacts := []struct {
		name string
		path string
	}{
		{"spec.md", ctx.SpecFile},
		{"plan.md", ctx.PlanFile},
		{"tasks.md", ctx.TasksFile},
	}

	for _, a := range artifacts {
		info, statErr := os.Stat(a.path)
		if statErr != nil {
			missing = append(missing, fmt.Sprintf("  - %s (not found)", a.name))
		} else if info.Size() == 0 {
			missing = append(missing, fmt.Sprintf("  - %s (empty)", a.name))
		}
	}

	if len(missing) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: cannot approve - missing artifacts:\n")
		for _, m := range missing {
			fmt.Fprintln(cmd.ErrOrStderr(), m)
		}
		os.Exit(1)
	}

	// Read current status
	status, err := spec.ReadStatus(ctx.FeatureDir)
	if err != nil {
		return fmt.Errorf("failed to read status: %w", err)
	}

	if status == "Approved" {
		fmt.Fprintf(cmd.OutOrStdout(), "Already approved: %s\n", ctx.Branch)
		return nil
	}

	// Update status to Approved
	if err := spec.WriteStatus(ctx.FeatureDir, "Approved"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Approved: %s\n", ctx.Branch)
	return nil
}
