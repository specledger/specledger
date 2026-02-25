package session

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/specledger/specledger/pkg/cli/auth"
	"gopkg.in/yaml.v3"
)

// gitCommitPattern matches git commit commands (not commit-graph, commit-tree, etc.)
var gitCommitPattern = regexp.MustCompile(`^\s*git\s+commit(\s|$)`)

// gitAmendPattern matches git commit --amend commands
var gitAmendPattern = regexp.MustCompile(`\s+--amend\b`)

// ParseHookInput parses the hook input from stdin
func ParseHookInput(data []byte) (*HookInput, error) {
	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse hook input: %w", err)
	}
	return &input, nil
}

// IsGitCommit checks if the tool input is a git commit command (but not amend)
func IsGitCommit(toolInput string) bool {
	if !gitCommitPattern.MatchString(toolInput) {
		return false
	}
	// Exclude amend commits (they create new sessions for the new hash)
	if gitAmendPattern.MatchString(toolInput) {
		return false
	}
	return true
}

// GetCurrentCommitHash gets the current HEAD commit hash
func GetCurrentCommitHash(workdir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workdir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch gets the current git branch name
func GetCurrentBranch(workdir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = workdir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// specledgerConfig represents the structure of specledger.yaml
type specledgerConfig struct {
	Project struct {
		ID string `yaml:"id"`
	} `yaml:"project"`
}

// GetProjectID attempts to get the project ID from specledger.yaml
func GetProjectID(workdir string) (string, error) {
	// Search for specledger/specledger.yaml in the workdir or parent directories
	searchPaths := []string{
		filepath.Join(workdir, "specledger", "specledger.yaml"),
		filepath.Join(workdir, "specledger.yaml"),
	}

	var configPath string
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			break
		}
	}

	if configPath == "" {
		return "", fmt.Errorf("no specledger.yaml found in %s", workdir)
	}

	// Read and parse the config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read specledger.yaml: %w", err)
	}

	var config specledgerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to parse specledger.yaml: %w", err)
	}

	if config.Project.ID == "" {
		return "", fmt.Errorf("project.id not found in specledger.yaml - please add it")
	}

	return config.Project.ID, nil
}

// Capture orchestrates the session capture flow
func Capture(input *HookInput) *CaptureResult {
	result := &CaptureResult{Captured: false}

	// Check if this is a git commit
	if !IsGitCommit(input.ToolInput.Command()) {
		return result // Not a commit, nothing to capture
	}

	// Verify the tool succeeded
	if !input.ToolSuccess() {
		return result // Commit failed, nothing to capture
	}

	// Check for transcript path
	if input.TranscriptPath == "" {
		result.Error = fmt.Errorf("no transcript path in hook input")
		return result
	}

	// Check transcript exists
	if _, err := os.Stat(input.TranscriptPath); err != nil {
		result.Error = fmt.Errorf("transcript not found: %w", err)
		return result
	}

	// Get commit hash
	commitHash, err := GetCurrentCommitHash(input.Cwd)
	if err != nil {
		result.Error = err
		return result
	}

	// Get branch name
	branch, err := GetCurrentBranch(input.Cwd)
	if err != nil {
		result.Error = err
		return result
	}

	// Get project ID
	projectID, err := GetProjectID(input.Cwd)
	if err != nil {
		// Give clear guidance on how to fix
		fmt.Fprintf(os.Stderr, "⚠️  Session capture skipped: %v\n", err)
		fmt.Fprintf(os.Stderr, "    To enable session capture, ensure specledger.yaml has project.id set.\n")
		fmt.Fprintf(os.Stderr, "    Run 'sl init' or add manually: project:\\n  id: <your-project-id>\n")
		result.Error = err
		return result
	}

	// Get last offset for this session
	offsetInfo, err := GetSessionOffset(input.SessionID)
	if err != nil {
		result.Error = fmt.Errorf("failed to get session offset: %w", err)
		return result
	}

	// Compute delta
	messages, newOffset, err := ComputeDelta(input.TranscriptPath, offsetInfo.LastOffset)
	if err != nil {
		result.Error = fmt.Errorf("failed to compute delta: %w", err)
		return result
	}

	// Check if there's anything to capture
	if len(messages) == 0 {
		return result // No new messages since last capture
	}

	// Get credentials
	creds, err := auth.LoadCredentials()
	if err != nil || creds == nil {
		result.Error = fmt.Errorf("not authenticated: run 'sl auth login'")
		return result
	}

	// Build session content
	sessionID := uuid.New().String()
	content := &SessionContent{
		Version:       "1.0",
		SessionID:     sessionID,
		FeatureBranch: branch,
		CommitHash:    commitHash,
		TaskID:        "",
		Author:        creds.UserEmail,
		CapturedAt:    time.Now(),
		Messages:      messages,
	}

	// Marshal and compress
	contentJSON, err := json.Marshal(content)
	if err != nil {
		result.Error = fmt.Errorf("failed to marshal session: %w", err)
		return result
	}

	rawSize := int64(len(contentJSON))
	if rawSize > MaxSessionSize {
		result.Error = fmt.Errorf("session too large: %d bytes (max %d)", rawSize, MaxSessionSize)
		return result
	}

	compressed, err := Compress(contentJSON)
	if err != nil {
		result.Error = fmt.Errorf("failed to compress session: %w", err)
		return result
	}

	result.SessionID = sessionID
	result.MessageCount = len(messages)
	result.SizeBytes = int64(len(compressed))
	result.RawSizeBytes = rawSize
	result.StoragePath = BuildStoragePath(projectID, branch, commitHash)

	// Try to get valid access token
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		// Queue for later upload
		return queueSession(result, compressed, projectID, branch, &commitHash, nil, creds.UserID)
	}

	// Upload to storage
	storage := NewStorageClient()
	_, err = storage.Upload(accessToken, result.StoragePath, compressed)
	if err != nil {
		// Queue for later upload
		return queueSession(result, compressed, projectID, branch, &commitHash, nil, creds.UserID)
	}

	// Create metadata
	metadata := NewMetadataClient()
	_, err = metadata.Create(accessToken, &CreateSessionInput{
		ProjectID:     projectID,
		FeatureBranch: branch,
		CommitHash:    &commitHash,
		TaskID:        nil,
		AuthorID:      creds.UserID,
		StoragePath:   result.StoragePath,
		Status:        StatusComplete,
		SizeBytes:     result.SizeBytes,
		RawSizeBytes:  result.RawSizeBytes,
		MessageCount:  result.MessageCount,
	})
	if err != nil {
		// Storage upload succeeded but metadata failed - still queue
		return queueSession(result, compressed, projectID, branch, &commitHash, nil, creds.UserID)
	}

	// Update offset tracking
	if err := UpdateSessionOffset(input.SessionID, newOffset, commitHash, input.TranscriptPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update session offset: %v\n", err)
	}

	result.Captured = true
	return result
}

// queueSession queues a session for later upload
func queueSession(result *CaptureResult, compressed []byte, projectID, branch string, commitHash, taskID *string, authorID string) *CaptureResult {
	queue := NewQueue()
	entry := &QueueEntry{
		SessionID:     result.SessionID,
		ProjectID:     projectID,
		FeatureBranch: branch,
		CommitHash:    commitHash,
		TaskID:        taskID,
		AuthorID:      authorID,
		Status:        StatusComplete,
		CreatedAt:     time.Now(),
		RetryCount:    0,
	}

	if err := queue.Enqueue(result.SessionID, compressed, entry); err != nil {
		result.Error = fmt.Errorf("failed to queue session: %w", err)
		return result
	}

	result.Queued = true
	return result
}

// CaptureFromStdin reads hook input from stdin and captures the session
func CaptureFromStdin() *CaptureResult {
	// Check if stdin is available (not a terminal)
	stat, err := os.Stdin.Stat()
	if err != nil {
		return &CaptureResult{Error: fmt.Errorf("failed to stat stdin: %w", err)}
	}

	// Check if stdin is a pipe or has data available
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return &CaptureResult{
			Error: fmt.Errorf("no input provided: this command reads hook JSON from stdin (use --test-mode for manual testing)"),
		}
	}

	// Read stdin with timeout using a channel
	type readResult struct {
		data []byte
		err  error
	}

	resultChan := make(chan readResult, 1)
	go func() {
		data, err := io.ReadAll(os.Stdin)
		resultChan <- readResult{data: data, err: err}
	}()

	// Wait for read or timeout (5 seconds)
	select {
	case result := <-resultChan:
		if result.err != nil {
			return &CaptureResult{Error: fmt.Errorf("failed to read stdin: %w", result.err)}
		}

		if len(result.data) == 0 {
			return &CaptureResult{Error: fmt.Errorf("no data received from stdin")}
		}

		// Parse hook input
		input, err := ParseHookInput(result.data)
		if err != nil {
			return &CaptureResult{Error: err}
		}

		return Capture(input)

	case <-time.After(5 * time.Second):
		return &CaptureResult{Error: fmt.Errorf("timeout waiting for stdin input (waited 5 seconds)")}
	}
}

// CaptureTestMode runs a test capture simulation for manual testing
func CaptureTestMode() *CaptureResult {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return &CaptureResult{Error: fmt.Errorf("failed to get working directory: %w", err)}
	}

	// Check if we're in a git repository
	if _, err := GetCurrentBranch(cwd); err != nil {
		return &CaptureResult{Error: fmt.Errorf("not in a git repository: %w", err)}
	}

	// Check if project has a project ID
	projectID, err := GetProjectID(cwd)
	if err != nil {
		return &CaptureResult{Error: fmt.Errorf("project setup incomplete: %w\n\nTo use session capture, ensure specledger.yaml has a project.id field", err)}
	}

	fmt.Printf("✓ Git repository detected\n")
	fmt.Printf("✓ Project ID found: %s\n", projectID)

	// Check for authentication
	creds, err := auth.LoadCredentials()
	if err != nil || creds == nil {
		return &CaptureResult{Error: fmt.Errorf("not authenticated. Run: sl auth login")}
	}
	fmt.Printf("✓ Authenticated as: %s\n", creds.UserEmail)

	// Check for Claude Code transcript in projects directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return &CaptureResult{Error: fmt.Errorf("failed to get home directory: %w", err)}
	}

	// Look for Claude Code projects directory (where transcripts are stored)
	projectsDir := filepath.Join(homeDir, ".claude", "projects")

	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		return &CaptureResult{
			Error: fmt.Errorf(`Claude Code projects directory not found at %s

Session capture requires an active Claude Code session.

To test:
1. Start a Claude Code session in this project
2. Run: sl session capture --test-mode`,
				projectsDir),
		}
	}

	fmt.Printf("✓ Claude Code projects directory found\n")

	// Find most recent transcript file in projects
	transcriptPath, sessionID, err := findMostRecentTranscript(projectsDir)
	if err != nil {
		return &CaptureResult{Error: fmt.Errorf("failed to find transcript: %w", err)}
	}

	fmt.Printf("✓ Found transcript: %s\n", transcriptPath)
	fmt.Printf("✓ Session ID: %s\n", sessionID)

	// Create simulated hook input
	fmt.Println("\n📝 Simulating git commit hook...")
	input := &HookInput{
		SessionID:      sessionID,
		TranscriptPath: transcriptPath,
		Cwd:            cwd,
		HookEventName:  "PostToolUse",
		ToolName:       "Bash",
		ToolInput:      ToolInput{Raw: json.RawMessage(`{"command":"git commit -m \"test\""}`)},
		ToolResponse:   ToolResponse{Interrupted: false}, // Simulate successful command
	}

	// Note: This won't actually capture because we need a real commit
	// Instead, show what would happen
	fmt.Println("\n⚠️  Test mode simulates the capture flow but won't create a real session.")
	fmt.Println("To capture a real session, make a git commit while using Claude Code.")

	// Validate the flow up to the commit check
	if !IsGitCommit(input.ToolInput.Command()) {
		return &CaptureResult{Error: fmt.Errorf("command validation failed (this is expected in test mode)")}
	}

	fmt.Println("✅ Session capture system is configured correctly!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Work on your code with Claude Code")
	fmt.Println("  2. Make a git commit")
	fmt.Println("  3. Session will be automatically captured")
	fmt.Println("\nTo view sessions: sl session list")

	return &CaptureResult{Captured: false}
}

// findMostRecentTranscript finds the most recent Claude Code transcript
// Searches in ~/.claude/projects/<project-slug>/<session-uuid>.jsonl
func findMostRecentTranscript(projectsDir string) (string, string, error) {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return "", "", err
	}

	var newestTranscript string
	var newestSessionID string
	var newestTime time.Time

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectDir := filepath.Join(projectsDir, entry.Name())
		files, err := os.ReadDir(projectDir)
		if err != nil {
			continue
		}

		// Find .jsonl files (session transcripts)
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".jsonl") {
				continue
			}

			transcriptPath := filepath.Join(projectDir, file.Name())
			info, err := os.Stat(transcriptPath)
			if err != nil {
				continue
			}

			if newestTranscript == "" || info.ModTime().After(newestTime) {
				newestTranscript = transcriptPath
				// Extract session ID from filename (remove .jsonl)
				newestSessionID = strings.TrimSuffix(file.Name(), ".jsonl")
				newestTime = info.ModTime()
			}
		}
	}

	if newestTranscript == "" {
		return "", "", fmt.Errorf("no transcript files found in %s", projectsDir)
	}

	return newestTranscript, newestSessionID, nil
}
