package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/getsentry/sentry-go"
)

const (
	// CaptureErrorsLogFile is the name of the local error log file
	CaptureErrorsLogFile = "capture-errors.log"
)

// CaptureErrorEntry represents a single error log entry
type CaptureErrorEntry struct {
	Timestamp     time.Time `json:"timestamp"`
	UserID        string    `json:"user_id"`
	ProjectID     string    `json:"project_id"`
	SessionID     string    `json:"session_id,omitempty"`
	ErrorMessage  string    `json:"error_message"`
	FeatureBranch string    `json:"feature_branch,omitempty"`
	CommitHash    string    `json:"commit_hash,omitempty"`
	RetryCount    int       `json:"retry_count"`
}

// GetCaptureErrorsLogPath returns the path to the local capture errors log file
func GetCaptureErrorsLogPath() string {
	return filepath.Join(GetBaseDir(), CaptureErrorsLogFile)
}

// LogCaptureError logs a capture error to both local file and Sentry.
// Local write happens first (guaranteed). Sentry is best-effort.
// This function never panics and never blocks the caller for long.
func LogCaptureError(entry CaptureErrorEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// 1. Write to local JSONL file first (guaranteed)
	logToLocalFile(entry)

	// 2. Send to Sentry (best-effort, non-blocking)
	logToSentry(entry)
}

// logToLocalFile appends a JSONL entry to the local capture errors log
func logToLocalFile(entry CaptureErrorEntry) {
	logPath := GetCaptureErrorsLogPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return // Can't create dir, silently fail
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return // Can't marshal, silently fail
	}

	// Append with newline (JSONL format)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.Write(append(data, '\n'))
}

// logToSentry sends the error to Sentry with structured context tags.
// Best-effort: if it fails, we already have the local log.
func logToSentry(entry CaptureErrorEntry) {
	defer func() {
		recover() //nolint:errcheck // swallow any panics — never crash the caller
	}()

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{ID: entry.UserID})
		scope.SetTag("project_id", entry.ProjectID)
		if entry.SessionID != "" {
			scope.SetTag("session_id", entry.SessionID)
		}
		if entry.FeatureBranch != "" {
			scope.SetTag("branch", entry.FeatureBranch)
		}
		if entry.CommitHash != "" {
			scope.SetTag("commit_hash", entry.CommitHash)
		}
		scope.SetExtra("retry_count", entry.RetryCount)
		sentry.CaptureException(fmt.Errorf("session capture: %s", entry.ErrorMessage))
	})
}
