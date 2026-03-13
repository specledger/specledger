package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// SpawnClaudeCLI starts the Claude CLI as a detached process that executes
// the /specledger.implement prompt. Returns the started process or an error.
func SpawnClaudeCLI(feature, projectRoot, logDir string) (*os.Process, error) {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return nil, fmt.Errorf("claude CLI not found in PATH: %w", err)
	}

	// Verify the implement prompt exists
	promptFile := filepath.Join(projectRoot, ".claude", "commands", "specledger.implement.md")
	if _, err := os.Stat(promptFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("specledger.implement prompt not found: %s", promptFile)
	}

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file for claude output
	logFile := filepath.Join(logDir, feature+"-claude.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open claude log file: %w", err)
	}

	cmd := exec.Command(claudePath, "-p", "\"/specledger.implement\"", "--dangerously-skip-permissions")
	cmd.Dir = projectRoot
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to start claude CLI: %w", err)
	}

	// Don't close the log file — the process needs it.
	// It will be closed when the process exits.
	return cmd.Process, nil
}

// SpawnImplement starts 'sl implement --feature <feature>' as a detached background process.
// This is called by the hook execute command to run implementation without blocking the push.
func SpawnImplement(feature, projectRoot string) error {
	slPath, err := exec.LookPath("sl")
	if err != nil {
		return fmt.Errorf("sl binary not found in PATH: %w", err)
	}

	// Create log directory
	logDir := filepath.Join(projectRoot, ".specledger", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile := filepath.Join(logDir, feature+"-implement.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open implement log: %w", err)
	}

	cmd := exec.Command(slPath, "implement", "--feature", feature)
	cmd.Dir = projectRoot
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		f.Close()
		return fmt.Errorf("failed to spawn sl implement: %w", err)
	}

	// Detach — don't wait for the process
	return nil
}
