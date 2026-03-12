// Package scheduler provides push-triggered implementation execution logic
// including execution lock management, approved spec detection, and Claude CLI spawning.
package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const lockFileName = "exec.lock"

// ExecutionLock represents the lock file contents that prevent duplicate runs.
type ExecutionLock struct {
	PID       int    `json:"pid"`
	Feature   string `json:"feature"`
	StartedAt string `json:"started_at"`
}

// LockFilePath returns the path to the exec.lock file within the .specledger directory.
func LockFilePath(projectRoot string) string {
	return filepath.Join(projectRoot, ".specledger", lockFileName)
}

// AcquireLock creates the execution lock file with the current process PID.
// Returns an error if the lock already exists.
func AcquireLock(projectRoot, feature string) error {
	lockPath := LockFilePath(projectRoot)

	// Ensure .specledger directory exists
	dir := filepath.Dir(lockPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create .specledger directory: %w", err)
	}

	// Check if lock already exists
	if _, err := os.Stat(lockPath); err == nil {
		existing, readErr := CheckLock(projectRoot)
		if readErr == nil && existing != nil {
			return fmt.Errorf("execution lock held (PID %d, feature: %s, started: %s)",
				existing.PID, existing.Feature, existing.StartedAt)
		}
		return fmt.Errorf("execution lock file exists: %s", lockPath)
	}

	lock := ExecutionLock{
		PID:       os.Getpid(),
		Feature:   feature,
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock: %w", err)
	}

	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}

// CheckLock reads and parses the execution lock file.
// Returns nil if no lock file exists.
func CheckLock(projectRoot string) (*ExecutionLock, error) {
	lockPath := LockFilePath(projectRoot)
	data, err := os.ReadFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lock ExecutionLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	return &lock, nil
}

// ReleaseLock removes the execution lock file.
func ReleaseLock(projectRoot string) error {
	lockPath := LockFilePath(projectRoot)
	err := os.Remove(lockPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}
	return nil
}

// IsLockHeld returns true if an execution lock file exists.
func IsLockHeld(projectRoot string) bool {
	lockPath := LockFilePath(projectRoot)
	_, err := os.Stat(lockPath)
	return err == nil
}
