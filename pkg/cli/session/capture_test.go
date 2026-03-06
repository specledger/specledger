package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIsGitCommit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// ===== TRUE POSITIVES: Should match =====
		{"simple commit", "git commit -m 'test'", true},
		{"commit with flags", "git commit -a -m 'test'", true},
		{"commit with double quotes", `git commit -m "test message"`, true},
		{"commit with leading space", "  git commit -m 'test'", true},
		{"commit with tab", "\tgit commit -m 'test'", true},
		{"commit bare", "git commit", true},
		{"commit with all flag", "git commit --all -m 'msg'", true},
		{"commit with signoff", "git commit -s -m 'msg'", true},
		{"commit with verbose", "git commit -v", true},
		{"commit with author", "git commit --author='Name <email>'", true},
		{"commit with date", "git commit --date='2026-01-01'", true},
		{"commit with no-verify", "git commit --no-verify -m 'msg'", true},
		{"commit with gpg-sign", "git commit -S -m 'msg'", true},
		{"commit with fixup", "git commit --fixup=abc123", true},
		{"commit heredoc style", "git commit -m \"$(cat <<'EOF'\nmsg\nEOF\n)\"", true},

		// ===== TRUE NEGATIVES: Should NOT match =====
		// Other git commands
		{"git add", "git add .", false},
		{"git push", "git push origin main", false},
		{"git status", "git status", false},
		{"git log", "git log --oneline", false},
		{"git diff", "git diff HEAD", false},
		{"git fetch", "git fetch origin", false},
		{"git pull", "git pull --rebase", false},
		{"git checkout", "git checkout main", false},
		{"git branch", "git branch -d old", false},
		{"git merge", "git merge feature", false},
		{"git rebase", "git rebase main", false},
		{"git stash", "git stash", false},
		{"git reset", "git reset --hard", false},

		// Similar but different commands
		{"git commit-graph", "git commit-graph write", false},
		{"git commit-tree", "git commit-tree abc123", false},

		// Amend variants (excluded by design)
		{"commit amend", "git commit --amend", false},
		{"commit amend with message", "git commit --amend -m 'fix'", false},
		{"commit amend no-edit", "git commit --amend --no-edit", false},
		{"commit with -a and amend", "git commit -a --amend", false},

		// ===== FALSE POSITIVE PREVENTION =====
		// Echo/print commands (should NOT trigger)
		{"echo git commit", `echo "git commit"`, false},
		{"echo single quote", `echo 'git commit -m test'`, false},
		{"printf git commit", `printf "git commit -m %s" msg`, false},

		// Comments (should NOT trigger)
		{"bash comment", "# git commit -m 'test'", false},
		{"inline comment", "ls # git commit", false},

		// Chained commands - git commit in chain SHOULD trigger
		{"and chain", "git add . && git commit -m 'test'", true},
		{"semicolon chain", "git status; git commit -m 'test'", true},
		{"or chain", "git add . || git commit -m 'test'", true},
		// Pipes and subshells should NOT trigger (different semantics)
		{"pipe", "echo test | git commit", false},
		{"subshell", "$(git commit -m 'test')", false},
		{"backtick subshell", "`git commit -m 'test'`", false},

		// String contains git commit but not a command
		{"grep for commit", "grep 'git commit' history.txt", false},
		{"cat heredoc", "cat <<EOF\ngit commit\nEOF", false},

		// Edge cases
		{"empty", "", false},
		{"whitespace only", "   ", false},
		{"git without commit", "git", false},
		{"commit without git", "commit -m 'test'", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitCommit(tt.input)
			if result != tt.expected {
				t.Errorf("IsGitCommit(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToolInputCommand(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected string
	}{
		{
			name:     "object format",
			raw:      `{"command":"git commit -m 'test'"}`,
			expected: "git commit -m 'test'",
		},
		{
			name:     "object with description",
			raw:      `{"command":"git status","description":"Check status"}`,
			expected: "git status",
		},
		{
			name:     "plain string",
			raw:      `"git commit -m 'test'"`,
			expected: "git commit -m 'test'",
		},
		{
			name:     "empty object",
			raw:      `{}`,
			expected: "{}", // Returns raw when no command field
		},
		{
			name:     "empty",
			raw:      ``,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := ToolInput{Raw: json.RawMessage(tt.raw)}
			result := ti.Command()
			if result != tt.expected {
				t.Errorf("Command() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestToolSuccess(t *testing.T) {
	tests := []struct {
		name        string
		interrupted bool
		expected    bool
	}{
		{"not interrupted", false, true},
		{"interrupted", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &HookInput{
				ToolResponse: ToolResponse{Interrupted: tt.interrupted},
			}
			result := input.ToolSuccess()
			if result != tt.expected {
				t.Errorf("ToolSuccess() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseHookInput(t *testing.T) {
	// Sample hook input from Claude Code
	sampleJSON := `{
		"session_id": "abc-123",
		"transcript_path": "/home/user/.claude/projects/test/abc.jsonl",
		"cwd": "/home/user/project",
		"permission_mode": "acceptEdits",
		"hook_event_name": "PostToolUse",
		"tool_name": "Bash",
		"tool_input": {"command": "git commit -m 'test'"},
		"tool_response": {"stdout": "success", "stderr": "", "interrupted": false},
		"tool_use_id": "toolu_123"
	}`

	input, err := ParseHookInput([]byte(sampleJSON))
	if err != nil {
		t.Fatalf("ParseHookInput failed: %v", err)
	}

	if input.SessionID != "abc-123" {
		t.Errorf("SessionID = %q, want %q", input.SessionID, "abc-123")
	}
	if input.ToolName != "Bash" {
		t.Errorf("ToolName = %q, want %q", input.ToolName, "Bash")
	}
	if input.ToolInput.Command() != "git commit -m 'test'" {
		t.Errorf("Command = %q, want %q", input.ToolInput.Command(), "git commit -m 'test'")
	}
	if !input.ToolSuccess() {
		t.Error("ToolSuccess() = false, want true")
	}
	if input.ToolResponse.Stdout != "success" {
		t.Errorf("Stdout = %q, want %q", input.ToolResponse.Stdout, "success")
	}
}

func TestParseHookInputInvalid(t *testing.T) {
	_, err := ParseHookInput([]byte("invalid json"))
	if err == nil {
		t.Error("ParseHookInput should fail for invalid JSON")
	}
}

// TestCaptureErrorsLogPath verifies the log path is under ~/.specledger/
func TestCaptureErrorsLogPath(t *testing.T) {
	path := GetCaptureErrorsLogPath()
	if !strings.Contains(path, ".specledger") {
		t.Errorf("GetCaptureErrorsLogPath() = %q, want path containing .specledger", path)
	}
	if !strings.HasSuffix(path, CaptureErrorsLogFile) {
		t.Errorf("GetCaptureErrorsLogPath() = %q, want suffix %q", path, CaptureErrorsLogFile)
	}
}

// TestLogCaptureErrorLocalFile verifies LogCaptureError writes JSONL to a local file
func TestLogCaptureErrorLocalFile(t *testing.T) {
	// Use a temp directory to avoid polluting the real log
	tmpDir := t.TempDir()
	origBaseDir := os.Getenv("HOME")

	// Override HOME so GetBaseDir() uses our temp dir
	// GetBaseDir falls back to HOME env var
	t.Setenv("HOME", tmpDir)
	// Also set USERPROFILE for Windows
	t.Setenv("USERPROFILE", tmpDir)

	logPath := filepath.Join(tmpDir, ".specledger", CaptureErrorsLogFile)

	entry := CaptureErrorEntry{
		Timestamp:     time.Now(),
		UserID:        "test-user-id",
		ProjectID:     "test-project-id",
		SessionID:     "test-session-id",
		ErrorMessage:  "test error message",
		FeatureBranch: "test-branch",
		CommitHash:    "abc123",
		RetryCount:    0,
	}

	// Write directly to local file (not through LogCaptureError which spawns goroutine)
	logToLocalFile(entry)

	// Verify file was created and contains the entry
	data, err := os.ReadFile(logPath)
	if err != nil {
		// Try the path GetCaptureErrorsLogPath returns (may differ on Windows)
		actualPath := GetCaptureErrorsLogPath()
		data, err = os.ReadFile(actualPath)
		if err != nil {
			t.Fatalf("Failed to read log file at %q or %q: %v (HOME was %q)", logPath, actualPath, err, origBaseDir)
		}
	}

	content := string(data)
	if !strings.Contains(content, "test-user-id") {
		t.Errorf("Log file missing user_id. Content: %s", content)
	}
	if !strings.Contains(content, "test error message") {
		t.Errorf("Log file missing error_message. Content: %s", content)
	}
	if !strings.Contains(content, "test-project-id") {
		t.Errorf("Log file missing project_id. Content: %s", content)
	}

	// Verify it's valid JSONL (single line, valid JSON)
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 JSONL line, got %d", len(lines))
	}

	var parsed CaptureErrorEntry
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Errorf("Failed to parse JSONL line as CaptureErrorEntry: %v", err)
	}
	if parsed.UserID != "test-user-id" {
		t.Errorf("Parsed UserID = %q, want %q", parsed.UserID, "test-user-id")
	}
}

// TestLogCaptureErrorAppends verifies multiple errors append to the same file
func TestLogCaptureErrorAppends(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	entry1 := CaptureErrorEntry{
		Timestamp:    time.Now(),
		UserID:       "user-1",
		ProjectID:    "proj-1",
		ErrorMessage: "first error",
	}
	entry2 := CaptureErrorEntry{
		Timestamp:    time.Now(),
		UserID:       "user-2",
		ProjectID:    "proj-2",
		ErrorMessage: "second error",
	}

	logToLocalFile(entry1)
	logToLocalFile(entry2)

	data, err := os.ReadFile(GetCaptureErrorsLogPath())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 JSONL lines, got %d", len(lines))
	}
}
