package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// GetSessionStatePath returns the path to the session state file
func GetSessionStatePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("HOME")
	}
	return filepath.Join(homeDir, ".specledger", "session-state.json")
}

// LoadSessionState loads the session state from disk
func LoadSessionState() (*SessionState, error) {
	statePath := GetSessionStatePath()

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty state if file doesn't exist
			return &SessionState{Sessions: make(map[string]*SessionOffsetInfo)}, nil
		}
		return nil, fmt.Errorf("failed to read session state: %w", err)
	}

	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse session state: %w", err)
	}

	if state.Sessions == nil {
		state.Sessions = make(map[string]*SessionOffsetInfo)
	}

	return &state, nil
}

// SaveSessionState saves the session state to disk
func SaveSessionState(state *SessionState) error {
	statePath := GetSessionStatePath()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write session state: %w", err)
	}

	return nil
}

// ComputeDelta reads new lines from a transcript file since the last offset
func ComputeDelta(transcriptPath string, lastOffset int64) ([]Message, int64, error) {
	file, err := os.Open(transcriptPath)
	if err != nil {
		return nil, lastOffset, fmt.Errorf("failed to open transcript: %w", err)
	}
	defer file.Close()

	// Seek to last offset
	if lastOffset > 0 {
		_, err := file.Seek(lastOffset, io.SeekStart)
		if err != nil {
			return nil, lastOffset, fmt.Errorf("failed to seek to offset %d: %w", lastOffset, err)
		}
	}

	var messages []Message
	scanner := bufio.NewScanner(file)
	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // max 1MB per line

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var tl TranscriptLine
		if err := json.Unmarshal(line, &tl); err != nil {
			// Skip malformed lines
			continue
		}

		// Convert transcript line to message
		msg := transcriptLineToMessage(tl)
		if msg != nil {
			messages = append(messages, *msg)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, lastOffset, fmt.Errorf("error reading transcript: %w", err)
	}

	// Get new offset
	newOffset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, lastOffset, fmt.Errorf("failed to get current offset: %w", err)
	}

	return messages, newOffset, nil
}

// transcriptLineToMessage converts a transcript line to a message
func transcriptLineToMessage(tl TranscriptLine) *Message {
	// Only capture user and assistant message types
	if tl.Type != "user" && tl.Type != "assistant" {
		return nil
	}

	// Determine role - prefer nested message role, fallback to type
	role := tl.Type
	if tl.Message != nil && tl.Message.Role != "" {
		role = tl.Message.Role
	} else if tl.Role != "" {
		role = tl.Role
	}

	// Only capture user and assistant messages
	if role != "user" && role != "assistant" {
		return nil
	}

	// Extract content - check nested message first, then direct content
	var content string
	if tl.Message != nil && tl.Message.Content != nil {
		content = extractContent(tl.Message.Content)
	} else if tl.Content != "" {
		content = tl.Content
	}

	// Skip empty content
	if content == "" {
		return nil
	}

	timestamp := tl.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	return &Message{
		Role:      role,
		Content:   content,
		Timestamp: timestamp,
	}
}

// extractContent extracts string content from various formats
func extractContent(c interface{}) string {
	switch v := c.(type) {
	case string:
		return v
	case []interface{}:
		// Array of content blocks - extract text from each
		var parts []string
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if text, ok := m["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

// GetTranscriptSize returns the size of the transcript file
func GetTranscriptSize(transcriptPath string) (int64, error) {
	info, err := os.Stat(transcriptPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat transcript: %w", err)
	}
	return info.Size(), nil
}

// UpdateSessionOffset updates the offset tracking for a session
func UpdateSessionOffset(sessionID string, offset int64, commitHash string, transcriptPath string) error {
	state, err := LoadSessionState()
	if err != nil {
		return err
	}

	state.Sessions[sessionID] = &SessionOffsetInfo{
		LastOffset:     offset,
		LastCommit:     commitHash,
		TranscriptPath: transcriptPath,
	}

	return SaveSessionState(state)
}

// GetSessionOffset retrieves the last offset for a session
func GetSessionOffset(sessionID string) (*SessionOffsetInfo, error) {
	state, err := LoadSessionState()
	if err != nil {
		return nil, err
	}

	info, exists := state.Sessions[sessionID]
	if !exists {
		return &SessionOffsetInfo{LastOffset: 0}, nil
	}

	return info, nil
}

// specledgerCommandPattern matches /specledger.<command> invocations in transcript content
var specledgerCommandPattern = regexp.MustCompile(`/specledger\.(plan|implement|specify|tasks|clarify|audit|analyze|resume|adopt|onboard|checklist|revise|help|constitution|remove-deps|add-deps)`)

// ReadRawDelta reads raw bytes from a transcript file since the last offset.
// Returns individual lines as byte slices without JSON parsing (for fast pattern matching).
func ReadRawDelta(transcriptPath string, lastOffset int64) ([][]byte, int64, error) {
	file, err := os.Open(transcriptPath)
	if err != nil {
		return nil, lastOffset, fmt.Errorf("failed to open transcript: %w", err)
	}
	defer file.Close()

	if lastOffset > 0 {
		if _, err := file.Seek(lastOffset, io.SeekStart); err != nil {
			return nil, lastOffset, fmt.Errorf("failed to seek to offset %d: %w", lastOffset, err)
		}
	}

	var lines [][]byte
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // max 1MB per line

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		// Make a copy since scanner reuses the buffer
		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)
		lines = append(lines, lineCopy)
	}

	if err := scanner.Err(); err != nil {
		return nil, lastOffset, fmt.Errorf("error reading transcript: %w", err)
	}

	newOffset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, lastOffset, fmt.Errorf("failed to get current offset: %w", err)
	}

	return lines, newOffset, nil
}

// DetectSpecledgerCommand scans raw transcript lines for /specledger.* command patterns.
// Returns the command name (e.g., "plan", "implement") or empty string if no match.
func DetectSpecledgerCommand(lines [][]byte) string {
	for _, line := range lines {
		match := specledgerCommandPattern.FindSubmatch(line)
		if match != nil {
			return string(match[1])
		}
	}
	return ""
}

// lockFile represents a file lock for concurrent access protection
type lockFile struct {
	file *os.File
}

// acquireStateLock acquires an exclusive file lock on the session state file.
// This prevents race conditions between Stop and PostToolUse hooks.
func acquireStateLock() (*lockFile, error) {
	lockPath := GetSessionStatePath() + ".lock"

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory: %w", err)
	}

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	// Try exclusive lock with timeout-like behavior via LOCK_NB + retry
	for i := 0; i < 50; i++ { // 50 * 100ms = 5s max wait
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return &lockFile{file: f}, nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	f.Close()
	return nil, fmt.Errorf("timeout acquiring state lock")
}

// release releases the file lock
func (l *lockFile) release() {
	if l.file != nil {
		syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
		l.file.Close()
	}
}

// LoadSessionStateLocked loads session state with file locking
func LoadSessionStateLocked() (*SessionState, *lockFile, error) {
	lock, err := acquireStateLock()
	if err != nil {
		return nil, nil, err
	}

	state, err := LoadSessionState()
	if err != nil {
		lock.release()
		return nil, nil, err
	}

	return state, lock, nil
}

// SaveSessionStateLocked saves session state and releases the lock
func SaveSessionStateLocked(state *SessionState, lock *lockFile) error {
	defer lock.release()
	return SaveSessionState(state)
}

// UpdateSessionOffsetLocked updates offset tracking with file locking and optional commit tracking.
// If commitHash is non-empty, it is appended to CapturedCommits (capped at MaxCapturedCommits).
func UpdateSessionOffsetLocked(sessionID string, offset int64, commitHash string, transcriptPath string) error {
	state, lock, err := LoadSessionStateLocked()
	if err != nil {
		return err
	}

	info, exists := state.Sessions[sessionID]
	if !exists {
		info = &SessionOffsetInfo{}
		state.Sessions[sessionID] = info
	}

	info.LastOffset = offset
	info.LastCommit = commitHash
	info.TranscriptPath = transcriptPath

	// Track captured commits for post-commit dedup
	if commitHash != "" {
		info.CapturedCommits = appendCapturedCommit(info.CapturedCommits, commitHash)
	}

	return SaveSessionStateLocked(state, lock)
}

// appendCapturedCommit appends a commit hash, keeping the list capped at MaxCapturedCommits
func appendCapturedCommit(commits []string, hash string) []string {
	// Check for duplicates
	for _, c := range commits {
		if c == hash {
			return commits
		}
	}

	commits = append(commits, hash)

	// Trim oldest entries if over cap
	if len(commits) > MaxCapturedCommits {
		commits = commits[len(commits)-MaxCapturedCommits:]
	}

	return commits
}

// IsCommitCapturedInClaude checks if a commit hash was already captured inside Claude.
// Used by post-commit hook for deduplication.
func IsCommitCapturedInClaude(commitHash string) (bool, error) {
	state, err := LoadSessionState()
	if err != nil {
		return false, err
	}

	for _, info := range state.Sessions {
		for _, c := range info.CapturedCommits {
			if c == commitHash {
				return true, nil
			}
		}
	}

	return false, nil
}
