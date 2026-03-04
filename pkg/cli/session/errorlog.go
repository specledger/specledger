package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/specledger/specledger/pkg/cli/auth"
)

const (
	// CaptureErrorsLogFile is the name of the local error log file
	CaptureErrorsLogFile = "capture-errors.log"
	// SessionCaptureErrorsTable is the Supabase table for error logging
	SessionCaptureErrorsTable = "session_capture_errors"
	// ErrorLogTimeout is the HTTP timeout for Supabase error logging
	ErrorLogTimeout = 10 * time.Second
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

// LogCaptureError logs a capture error to both local file and Supabase.
// Local write happens first (guaranteed). Supabase is best-effort.
// This function never panics and never blocks the caller for long.
func LogCaptureError(entry CaptureErrorEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// 1. Write to local JSONL file first (guaranteed)
	logToLocalFile(entry)

	// 2. Attempt Supabase logging (best-effort, with short timeout)
	// Note: must be synchronous — goroutine gets killed on process exit
	logToSupabase(entry)
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

// logToSupabase posts the error entry to Supabase session_capture_errors table.
// Best-effort: if it fails, we already have the local log.
func logToSupabase(entry CaptureErrorEntry) {
	defer func() {
		recover() //nolint:errcheck // swallow any panics — never crash the caller
	}()

	// Get access token (try current token first)
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		// Try force refresh once
		accessToken, err = auth.ForceRefreshAccessToken()
		if err != nil {
			return // Can't authenticate, give up
		}
	}

	supabaseURL := auth.GetSupabaseURL()
	anonKey := auth.GetSupabaseAnonKey()

	// Build request body (Supabase PostgREST format)
	body := map[string]interface{}{
		"user_id":       entry.UserID,
		"project_id":    entry.ProjectID,
		"error_message": entry.ErrorMessage,
		"retry_count":   entry.RetryCount,
	}
	if entry.SessionID != "" {
		body["session_id"] = entry.SessionID
	}
	if entry.FeatureBranch != "" {
		body["feature_branch"] = entry.FeatureBranch
	}
	if entry.CommitHash != "" {
		body["commit_hash"] = entry.CommitHash
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return
	}

	url := fmt.Sprintf("%s/rest/v1/%s", supabaseURL, SessionCaptureErrorsTable)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", anonKey)
	req.Header.Set("Prefer", "return=minimal")

	client := &http.Client{Timeout: ErrorLogTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// If 401, try refresh once and retry
	if resp.StatusCode == http.StatusUnauthorized {
		newToken, err := auth.ForceRefreshAccessToken()
		if err != nil {
			return
		}
		req.Header.Set("Authorization", "Bearer "+newToken)
		resp2, err := client.Do(req)
		if err != nil {
			return
		}
		defer resp2.Body.Close()
	}
}
