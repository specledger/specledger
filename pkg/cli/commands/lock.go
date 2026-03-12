package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/specledger/specledger/pkg/cli/scheduler"
	"github.com/spf13/cobra"
)

var lockJSONFlag bool

var VarLockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Manage the implementation execution lock",
	Long: `View and manage the execution lock that prevents duplicate sl implement runs.

Commands:
  sl lock status  Show current lock information
  sl lock reset   Manually remove a stale lock`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var lockStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display current execution lock information",
	RunE:  runLockStatus,
}

var lockResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Manually remove the execution lock",
	Long:  "Remove .specledger/exec.lock unconditionally. Use when a lock is left behind after a crash.",
	RunE:  runLockReset,
}

func init() {
	VarLockCmd.AddCommand(lockStatusCmd)
	VarLockCmd.AddCommand(lockResetCmd)

	lockStatusCmd.Flags().BoolVar(&lockJSONFlag, "json", false, "Output as JSON")
}

func runLockStatus(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	lock, err := scheduler.CheckLock(workDir)
	if err != nil {
		return fmt.Errorf("failed to read lock: %w", err)
	}

	if lock == nil {
		if lockJSONFlag {
			fmt.Fprintln(cmd.OutOrStdout(), `{"held": false}`)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "No active execution lock.")
		}
		return nil
	}

	if lockJSONFlag {
		result := map[string]interface{}{
			"held":       true,
			"pid":        lock.PID,
			"feature":    lock.Feature,
			"started_at": lock.StartedAt,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Lock held:\n  PID: %d\n  Feature: %s\n  Started: %s\n",
			lock.PID, lock.Feature, lock.StartedAt)
	}
	return nil
}

func runLockReset(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if !scheduler.IsLockHeld(workDir) {
		fmt.Fprintln(cmd.OutOrStdout(), "No lock found.")
		return nil
	}

	if err := scheduler.ReleaseLock(workDir); err != nil {
		return fmt.Errorf("failed to remove lock: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Execution lock removed.")
	return nil
}
