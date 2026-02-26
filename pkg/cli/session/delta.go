package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	// Increase buffer size for long lines (Claude transcripts can have very long lines)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024) // max 10MB per line

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
