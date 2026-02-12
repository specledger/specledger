package session

import (
	"encoding/json"
	"fmt"
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

// gitCommitPattern matches git commit commands
var gitCommitPattern = regexp.MustCompile(`^\s*git\s+commit\b`)

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
	if !IsGitCommit(input.ToolInput) {
		return result // Not a commit, nothing to capture
	}

	// Verify the tool succeeded
	if !input.ToolSuccess {
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
		// Log warning but don't fail - queue for later
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
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
	// Read stdin
	data, err := os.ReadFile("/dev/stdin")
	if err != nil {
		return &CaptureResult{Error: fmt.Errorf("failed to read stdin: %w", err)}
	}

	// Parse hook input
	input, err := ParseHookInput(data)
	if err != nil {
		return &CaptureResult{Error: err}
	}

	return Capture(input)
}
