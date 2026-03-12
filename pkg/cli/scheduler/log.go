package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	hookLogFileName = "push-hook.log"
	maxLogEntries   = 50
)

// WriteHookLog appends a structured log entry to the push-hook.log file.
// It auto-creates the log directory and rotates entries to keep at most maxLogEntries.
func WriteHookLog(logDir, branch, feature, action, detail string) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, hookLogFileName)
	timestamp := time.Now().UTC().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s] branch=%s feature=%s action=%s detail=%s",
		timestamp, branch, feature, action, detail)

	// Read existing entries
	var entries []string
	if data, err := os.ReadFile(logPath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) != "" {
				entries = append(entries, line)
			}
		}
	}

	entries = append(entries, entry)

	// Rotate: keep last maxLogEntries
	if len(entries) > maxLogEntries {
		entries = entries[len(entries)-maxLogEntries:]
	}

	content := strings.Join(entries, "\n") + "\n"
	return os.WriteFile(logPath, []byte(content), 0644)
}

// ReadRecentErrors returns the last N error-level entries from the push-hook.log.
func ReadRecentErrors(logDir string, count int) []string {
	logPath := filepath.Join(logDir, hookLogFileName)
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil
	}

	var errors []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "action=error") {
			errors = append(errors, line)
		}
	}

	if len(errors) > count {
		errors = errors[len(errors)-count:]
	}
	return errors
}
