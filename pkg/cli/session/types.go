// Package session provides checkpoint session capture functionality
// for storing AI conversation segments linked to git commits and beads tasks.
package session

import (
	"encoding/json"
	"time"
)

// MaxSessionSize is the maximum uncompressed session size (10 MB)
const MaxSessionSize = 10 * 1024 * 1024

// MaxCapturedCommits is the maximum number of commit hashes tracked for deduplication
const MaxCapturedCommits = 50

// CaptureTrigger identifies what triggered the session capture
type CaptureTrigger string

const (
	TriggerPostToolUse CaptureTrigger = "post-tool-use"
	TriggerStop        CaptureTrigger = "stop"
	TriggerPostCommit  CaptureTrigger = "post-commit"
)

// CaptureTypeName identifies the type of capture record
type CaptureTypeName string

const (
	CaptureTypeCommit        CaptureTypeName = "commit"
	CaptureTypeSpecledgerCmd CaptureTypeName = "specledger-command"
	CaptureTypePostCommit    CaptureTypeName = "post-commit"
)

// SessionStatus represents the completion status of a session
type SessionStatus string

const (
	StatusComplete  SessionStatus = "complete"
	StatusRejected  SessionStatus = "rejected"
	StatusAbandoned SessionStatus = "abandoned"
)

// Message represents a single message in the conversation
type Message struct {
	Role      string    `json:"role"`      // "user" or "assistant"
	Content   string    `json:"content"`   // message content
	Timestamp time.Time `json:"timestamp"` // when the message was sent
}

// SessionContent represents the full session data stored in Supabase Storage
type SessionContent struct {
	Version       string    `json:"version"`                // schema version (e.g., "1.0")
	SessionID     string    `json:"session_id"`             // unique identifier
	FeatureBranch string    `json:"feature_branch"`         // e.g., "010-checkpoint-session-capture"
	CommitHash    string    `json:"commit_hash"`            // git commit hash (nullable for task sessions)
	TaskID        string    `json:"task_id"`                // beads task ID (nullable for commit sessions)
	Author        string    `json:"author"`                 // user email
	CapturedAt    time.Time `json:"captured_at"`            // when captured
	Messages      []Message `json:"messages"`               // conversation messages
	CaptureType   string    `json:"capture_type,omitempty"` // "commit", "specledger-command", "post-commit"
	CommandName   string    `json:"command_name,omitempty"` // specledger command name (e.g., "plan", "implement")
}

// SessionMetadata represents the queryable metadata stored in the database
type SessionMetadata struct {
	ID            string        `json:"id"`
	ProjectID     string        `json:"project_id"`
	FeatureBranch string        `json:"feature_branch"`
	CommitHash    *string       `json:"commit_hash,omitempty"`
	TaskID        *string       `json:"task_id,omitempty"`
	AuthorID      string        `json:"author_id"`
	StoragePath   string        `json:"storage_path"`
	Status        SessionStatus `json:"status"`
	SizeBytes     int64         `json:"size_bytes"`     // compressed size
	RawSizeBytes  int64         `json:"raw_size_bytes"` // uncompressed size
	MessageCount  int           `json:"message_count"`
	CreatedAt     time.Time     `json:"created_at"`
}

// ToolInput represents the tool_input field from Claude Code hooks.
// For Bash tools, this is {"command": "..."}.
// We use json.RawMessage to handle both object and string formats.
type ToolInput struct {
	Raw json.RawMessage
}

func (t *ToolInput) UnmarshalJSON(data []byte) error {
	t.Raw = data
	return nil
}

// Command extracts the command string from the tool input.
// Handles both object format {"command": "..."} and plain string format.
func (t *ToolInput) Command() string {
	if len(t.Raw) == 0 {
		return ""
	}
	// Try object format: {"command": "..."}
	var obj struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(t.Raw, &obj); err == nil && obj.Command != "" {
		return obj.Command
	}
	// Try plain string format
	var s string
	if err := json.Unmarshal(t.Raw, &s); err == nil {
		return s
	}
	return string(t.Raw)
}

// HookInput represents the JSON input from Claude Code hooks
type HookInput struct {
	SessionID      string    `json:"session_id"`
	TranscriptPath string    `json:"transcript_path"`
	Cwd            string    `json:"cwd"`
	HookEventName  string    `json:"hook_event_name"`
	ToolName       string    `json:"tool_name"`
	ToolInput      ToolInput `json:"tool_input"`       // the command that was run
	ToolOutput     string    `json:"tool_output"`      // output from the tool
	ToolDurationMs int64     `json:"tool_duration_ms"` // how long the tool took
	ToolSuccess    bool      `json:"tool_success"`     // whether the tool succeeded
	StopHookActive bool      `json:"stop_hook_active"` // true when invoked from Stop hook (different JSON shape)
}

// SessionState represents the local tracking state for delta computation
type SessionState struct {
	Sessions map[string]*SessionOffsetInfo `json:"sessions"`
}

// SessionOffsetInfo tracks the last captured position in the transcript
type SessionOffsetInfo struct {
	LastOffset      int64    `json:"last_offset"`                // byte offset in transcript file
	LastCommit      string   `json:"last_commit"`                // last commit hash captured
	TranscriptPath  string   `json:"transcript_path"`            // path to the transcript file
	CapturedCommits []string `json:"captured_commits,omitempty"` // commit hashes captured in-Claude (for post-commit dedup)
}

// QueueEntry represents a session queued for upload
type QueueEntry struct {
	SessionID     string        `json:"session_id"`
	ProjectID     string        `json:"project_id"`
	FeatureBranch string        `json:"feature_branch"`
	CommitHash    *string       `json:"commit_hash,omitempty"`
	TaskID        *string       `json:"task_id,omitempty"`
	AuthorID      string        `json:"author_id"`
	Status        SessionStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	RetryCount    int           `json:"retry_count"`
	LastRetry     *time.Time    `json:"last_retry,omitempty"`
}

// TranscriptLine represents a single line from the Claude Code transcript JSONL
type TranscriptLine struct {
	Type      string         `json:"type"`              // "user", "assistant", "tool_use", etc.
	Message   *TranscriptMsg `json:"message,omitempty"` // nested message object
	Content   string         `json:"content,omitempty"` // direct content (fallback)
	Timestamp time.Time      `json:"timestamp"`         // when recorded
	UUID      string         `json:"uuid,omitempty"`    // message UUID
	Role      string         `json:"role,omitempty"`    // alternative to type
}

// TranscriptMsg represents the nested message in Claude Code transcripts
type TranscriptMsg struct {
	Role    string      `json:"role"`              // "user" or "assistant"
	Content interface{} `json:"content,omitempty"` // string or array of content blocks
}

// CaptureResult represents the outcome of a capture operation
type CaptureResult struct {
	Captured     bool   // whether a session was captured
	SessionID    string // the session ID if captured
	StoragePath  string // path in storage
	MessageCount int    // number of messages in the session
	SizeBytes    int64  // compressed size
	RawSizeBytes int64  // uncompressed size
	Queued       bool   // whether it was queued for later upload
	Error        error  // any error that occurred
}
