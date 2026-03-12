package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/specledger/specledger/pkg/cli/scheduler"
	"github.com/specledger/specledger/pkg/cli/spec"
	"github.com/spf13/cobra"
)

var implementFeatureFlag string

var VarImplementCmd = &cobra.Command{
	Use:   "implement",
	Short: "Execute implementation for an approved feature via Claude CLI",
	Long: `Execute implementation by delegating to the Claude CLI.

This command acquires the execution lock, spawns claude -p "/specledger.implement"
--dangerously-skip-permissions, waits for completion, and releases the lock.

Typically called by 'sl hook execute' as a background process after git push.

Examples:
  sl implement                        # Auto-detect feature from branch
  sl implement --feature 127-feature  # Implement a specific feature`,
	RunE: runImplement,
}

func init() {
	VarImplementCmd.Flags().StringVar(&implementFeatureFlag, "feature", "", "Feature spec context to implement (default: auto-detect from branch)")
}

func runImplement(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Resolve feature context
	opts := spec.DetectionOptions{SpecOverride: implementFeatureFlag}
	ctx, err := spec.DetectFeatureContextWithOptions(workDir, opts)
	if err != nil {
		return fmt.Errorf("failed to detect feature context: %w", err)
	}

	feature := ctx.Branch
	projectRoot := ctx.RepoRoot

	// Verify claude CLI is available
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude CLI not found in PATH. Install Claude Code CLI first")
	}

	// Verify the implement prompt exists
	promptFile := filepath.Join(projectRoot, ".claude", "commands", "specledger.implement.md")
	if _, err := os.Stat(promptFile); os.IsNotExist(err) {
		return fmt.Errorf("specledger.implement prompt not found: %s", promptFile)
	}

	// Acquire execution lock
	if err := scheduler.AcquireLock(projectRoot, feature); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %s\nRun 'sl lock reset' to clear if the process is no longer running.\n", err)
		os.Exit(1)
	}
	defer scheduler.ReleaseLock(projectRoot)

	logDir := filepath.Join(projectRoot, ".specledger", "logs")
	fmt.Fprintf(cmd.OutOrStdout(), "Implementing: %s\n", feature)
	fmt.Fprintf(cmd.OutOrStdout(), "Claude CLI log: %s\n", filepath.Join(logDir, feature+"-claude.log"))

	// Spawn Claude CLI and wait for completion
	proc, err := scheduler.SpawnClaudeCLI(feature, projectRoot, logDir)
	if err != nil {
		return fmt.Errorf("failed to spawn Claude CLI: %w", err)
	}

	// Wait for the process to complete
	state, err := proc.Wait()
	if err != nil {
		writeResultSummary(logDir, feature, "failed", err.Error())
		return fmt.Errorf("Claude CLI exited with error: %w", err)
	}

	exitCode := state.ExitCode()
	if exitCode == 0 {
		writeResultSummary(logDir, feature, "success", "Implementation completed successfully")
		fmt.Fprintln(cmd.OutOrStdout(), "Implementation complete.")
	} else {
		writeResultSummary(logDir, feature, "failed", fmt.Sprintf("exit code %d", exitCode))
		os.Exit(exitCode)
	}

	return nil
}

func writeResultSummary(logDir, feature, status, message string) {
	resultFile := filepath.Join(logDir, feature+"-result.md")
	content := fmt.Sprintf("# Implementation Result: %s\n\n**Status**: %s\n**Completed**: %s\n**Message**: %s\n",
		feature, status, time.Now().UTC().Format(time.RFC3339), message)
	_ = os.MkdirAll(logDir, 0755)
	_ = os.WriteFile(resultFile, []byte(content), 0644)
}
