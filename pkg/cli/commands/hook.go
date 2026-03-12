package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/specledger/specledger/pkg/cli/hooks"
	"github.com/specledger/specledger/pkg/cli/scheduler"
	"github.com/spf13/cobra"
)

var (
	hookForceFlag bool
	hookJSONFlag  bool
	hookEventFlag string
)

var VarHookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Manage SpecLedger git push hooks",
	Long: `Install, uninstall, and check status of the SpecLedger pre-push git hook.

The push hook detects approved specs and triggers implementation automatically.

Commands:
  sl hook install    Install the pre-push hook
  sl hook uninstall  Remove the pre-push hook
  sl hook status     Check hook installation status
  sl hook execute    (internal) Called by the pre-push hook`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var hookInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the SpecLedger pre-push git hook",
	RunE:  runHookInstall,
}

var hookUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the SpecLedger pre-push git hook",
	RunE:  runHookUninstall,
}

var hookStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if the SpecLedger push hook is installed",
	RunE:  runHookStatus,
}

var hookExecuteCmd = &cobra.Command{
	Use:    "execute",
	Short:  "Execute hook logic (internal, called by pre-push hook)",
	Hidden: true,
	RunE:   runHookExecute,
}

func init() {
	VarHookCmd.AddCommand(hookInstallCmd)
	VarHookCmd.AddCommand(hookUninstallCmd)
	VarHookCmd.AddCommand(hookStatusCmd)
	VarHookCmd.AddCommand(hookExecuteCmd)

	hookInstallCmd.Flags().BoolVar(&hookForceFlag, "force", false, "Overwrite existing SpecLedger hook block")
	hookStatusCmd.Flags().BoolVar(&hookJSONFlag, "json", false, "Output as JSON")
	hookExecuteCmd.Flags().StringVar(&hookEventFlag, "event", "", "Git hook event type")
}

func findGitDir() (string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Walk up to find .git directory
	dir := workDir
	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return gitDir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("not a git repository (or any parent)")
}

func runHookInstall(cmd *cobra.Command, args []string) error {
	gitDir, err := findGitDir()
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
		os.Exit(1)
	}

	if err := hooks.InstallPushHook(gitDir, hookForceFlag); err != nil {
		return fmt.Errorf("failed to install hook: %w", err)
	}

	hookPath := filepath.Join(gitDir, "hooks", "pre-push")
	if hookForceFlag {
		fmt.Fprintf(cmd.OutOrStdout(), "Push hook reinstalled at %s\n", hookPath)
	} else if hooks.HasPushHook(gitDir) {
		fmt.Fprintf(cmd.OutOrStdout(), "Push hook installed at %s\n", hookPath)
	}
	return nil
}

func runHookUninstall(cmd *cobra.Command, args []string) error {
	gitDir, err := findGitDir()
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
		os.Exit(1)
	}

	installed := hooks.HasPushHook(gitDir)
	if err := hooks.UninstallPushHook(gitDir); err != nil {
		return fmt.Errorf("failed to uninstall hook: %w", err)
	}

	if installed {
		fmt.Fprintln(cmd.OutOrStdout(), "Push hook uninstalled.")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Push hook: not installed")
	}
	return nil
}

func runHookStatus(cmd *cobra.Command, args []string) error {
	gitDir, err := findGitDir()
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
		os.Exit(1)
	}

	installed := hooks.HasPushHook(gitDir)
	hookPath := filepath.Join(gitDir, "hooks", "pre-push")

	if hookJSONFlag {
		result := map[string]interface{}{
			"installed": installed,
			"path":      hookPath,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	if installed {
		fmt.Fprintf(cmd.OutOrStdout(), "Push hook: installed\nLocation: %s\n", hookPath)
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Push hook: not installed")
	}

	// Show recent errors if any
	workDir, _ := os.Getwd()
	logDir := filepath.Join(workDir, ".specledger", "logs")
	if errors := scheduler.ReadRecentErrors(logDir, 5); len(errors) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "\nRecent errors:")
		for _, e := range errors {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", e)
		}
	}

	return nil
}

func runHookExecute(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		hookLog(workDir, "error", "failed to get working directory: "+err.Error())
		return nil // Never block the push
	}

	// Detect approved spec on current branch
	result, err := scheduler.DetectApprovedSpec(workDir)
	if err != nil {
		hookLog(workDir, "error", "detection failed: "+err.Error())
		return nil
	}

	if result == nil {
		hookLog(workDir, "skip", "no feature context detected")
		return nil
	}

	if !result.Approved {
		hookLog(workDir, "skip", fmt.Sprintf("feature %s not approved", result.Feature))
		return nil
	}

	// Check execution lock
	if scheduler.IsLockHeld(result.RepoRoot) {
		lock, err := scheduler.CheckLock(result.RepoRoot)
		if err == nil && lock != nil {
			// Check if PID is still alive
			if isProcessAlive(lock.PID) {
				hookLog(workDir, "skip", fmt.Sprintf("already running (PID %d, feature: %s)", lock.PID, lock.Feature))
				return nil
			}
			// Stale lock — remove it
			hookLog(workDir, "warn", fmt.Sprintf("removing stale lock (PID %d no longer running)", lock.PID))
			_ = scheduler.ReleaseLock(result.RepoRoot)
		}
	}

	// Spawn sl implement as a detached background process
	if err := scheduler.SpawnImplement(result.Feature, result.RepoRoot); err != nil {
		hookLog(workDir, "error", "failed to spawn implement: "+err.Error())
		return nil
	}

	hookLog(workDir, "triggered", fmt.Sprintf("spawned sl implement for %s", result.Feature))
	return nil
}

// hookLog writes a structured log entry via the scheduler's WriteHookLog with rotation.
func hookLog(projectRoot, action, detail string) {
	logDir := filepath.Join(projectRoot, ".specledger", "logs")
	_ = scheduler.WriteHookLog(logDir, "", "", action, detail)
}

// isProcessAlive checks if a process with the given PID is still running.
func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds. Send signal 0 to check.
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
