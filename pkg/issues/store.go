package issues

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
)

// Store-related errors
var (
	ErrIssueNotFound      = errors.New("issue not found")
	ErrIssueAlreadyExists = errors.New("issue already exists")
	ErrStoreLocked        = errors.New("store is locked by another process")
	ErrSpecDirNotFound    = errors.New("spec directory not found")
)

// Store manages JSONL file operations with file locking
type Store struct {
	path        string // Path to specledger directory
	specContext string // Current spec context (e.g., "010-my-feature")
	lock        *flock.Flock
	mu          sync.Mutex
}

// StoreOptions contains options for creating a new Store
type StoreOptions struct {
	BasePath    string // Base path to specledger directory (default: "specledger")
	SpecContext string // Spec context (e.g., "010-my-feature"), empty for cross-spec mode
}

// NewStore creates a new issue store for a specific spec context
func NewStore(opts StoreOptions) (*Store, error) {
	basePath := opts.BasePath
	if basePath == "" {
		basePath = "specledger"
	}

	// If no spec context, create a store for cross-spec operations
	if opts.SpecContext == "" {
		return &Store{
			path:        basePath,
			specContext: "",
			lock:        flock.New(filepath.Join(basePath, "issues.jsonl.lock")),
		}, nil
	}

	// Construct path to issues.jsonl
	issuesPath := filepath.Join(basePath, opts.SpecContext, "issues.jsonl")

	// Create lock file path
	lockPath := filepath.Join(basePath, opts.SpecContext, "issues.jsonl.lock")

	store := &Store{
		path:        issuesPath,
		specContext: opts.SpecContext,
		lock:        flock.New(lockPath),
	}

	return store, nil
}

// Path returns the path to the issues.jsonl file
func (s *Store) Path() string {
	return s.path
}

// Create creates a new issue in the store
func (s *Store) Create(issue *Issue) error {
	return s.WithLock(func() error {
		// Ensure spec context matches
		if issue.SpecContext != s.specContext {
			issue.SpecContext = s.specContext
		}

		// Validate issue
		if err := issue.Validate(); err != nil {
			return err
		}

		// Check if issue already exists
		existing, err := s.getByIDUnlocked(issue.ID)
		if err != nil && !errors.Is(err, ErrIssueNotFound) {
			return err
		}
		if existing != nil {
			return ErrIssueAlreadyExists
		}

		// Ensure directory exists
		dir := filepath.Dir(s.path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Append to file
		f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()

		data, err := json.Marshal(issue)
		if err != nil {
			return fmt.Errorf("failed to marshal issue: %w", err)
		}

		_, err = fmt.Fprintf(f, "%s\n", data)
		return err
	})
}

// Get retrieves an issue by ID
func (s *Store) Get(id string) (*Issue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getByIDUnlocked(id)
}

func (s *Store) getByIDUnlocked(id string) (*Issue, error) {
	issues, err := s.readAllUnlocked()
	if err != nil {
		return nil, err
	}

	for _, issue := range issues {
		if issue.ID == id {
			return issue, nil
		}
	}

	return nil, ErrIssueNotFound
}

// List returns issues matching the filter
func (s *Store) List(filter ListFilter) ([]Issue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	issues, err := s.readAllUnlocked()
	if err != nil {
		return nil, err
	}

	var result []Issue
	for _, issue := range issues {
		if s.matchesFilter(issue, filter) {
			result = append(result, *issue)
		}
	}

	return result, nil
}

// Update updates an existing issue
func (s *Store) Update(id string, update IssueUpdate) (*Issue, error) {
	return s.WithLockResult(func() (*Issue, error) {
		issues, err := s.readAllUnlocked()
		if err != nil {
			return nil, err
		}

		var found *Issue
		var foundIdx int
		for i, issue := range issues {
			if issue.ID == id {
				found = issue
				foundIdx = i
				break
			}
		}

		if found == nil {
			return nil, ErrIssueNotFound
		}

		// Apply updates
		if update.Title != nil {
			found.Title = *update.Title
		}
		if update.Description != nil {
			found.Description = *update.Description
		}
		if update.Status != nil {
			found.Status = *update.Status
			if *update.Status == StatusClosed && found.ClosedAt == nil {
				now := NowFunc()
				found.ClosedAt = &now
			}
		}
		if update.Priority != nil {
			found.Priority = *update.Priority
		}
		if update.IssueType != nil {
			found.IssueType = *update.IssueType
		}
		if update.Assignee != nil {
			found.Assignee = *update.Assignee
		}
		if update.Notes != nil {
			found.Notes = *update.Notes
		}
		if update.Design != nil {
			found.Design = *update.Design
		}
		if update.AcceptanceCriteria != nil {
			found.AcceptanceCriteria = *update.AcceptanceCriteria
		}
		if update.Labels != nil {
			found.Labels = *update.Labels
		}
		if len(update.AddLabels) > 0 {
			for _, label := range update.AddLabels {
				if !contains(found.Labels, label) {
					found.Labels = append(found.Labels, label)
				}
			}
		}
		if len(update.RemoveLabels) > 0 {
			var newLabels []string
			for _, label := range found.Labels {
				if !contains(update.RemoveLabels, label) {
					newLabels = append(newLabels, label)
				}
			}
			found.Labels = newLabels
		}
		if update.BlockedBy != nil {
			found.BlockedBy = *update.BlockedBy
		}
		if update.Blocks != nil {
			found.Blocks = *update.Blocks
		}
		if update.DefinitionOfDone != nil {
			found.DefinitionOfDone = update.DefinitionOfDone
		}
		if update.CheckDoDItem != "" && found.DefinitionOfDone != nil {
			found.DefinitionOfDone.CheckItem(update.CheckDoDItem)
		}
		if update.UncheckDoDItem != "" && found.DefinitionOfDone != nil {
			found.DefinitionOfDone.UncheckItem(update.UncheckDoDItem)
		}

		// Update timestamp
		found.UpdatedAt = NowFunc()

		// Validate
		if err := found.Validate(); err != nil {
			return nil, err
		}

		// Update in slice
		issues[foundIdx] = found

		// Write all issues back
		if err := s.writeAllUnlocked(issues); err != nil {
			return nil, err
		}

		return found, nil
	})
}

// Delete removes an issue from the store
func (s *Store) Delete(id string) error {
	return s.WithLock(func() error {
		issues, err := s.readAllUnlocked()
		if err != nil {
			return err
		}

		var newIssues []*Issue
		var found bool
		for _, issue := range issues {
			if issue.ID == id {
				found = true
				continue
			}
			newIssues = append(newIssues, issue)
		}

		if !found {
			return ErrIssueNotFound
		}

		return s.writeAllUnlocked(newIssues)
	})
}

// WithLock executes a function while holding the file lock
func (s *Store) WithLock(fn func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	locked, err := s.lock.TryLock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return ErrStoreLocked
	}
	defer func() { _ = s.lock.Unlock() }()

	return fn()
}

// WithLockResult executes a function while holding the file lock and returns a result
func (s *Store) WithLockResult(fn func() (*Issue, error)) (*Issue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	locked, err := s.lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return nil, ErrStoreLocked
	}
	defer func() { _ = s.lock.Unlock() }()

	return fn()
}

func (s *Store) readAllUnlocked() ([]*Issue, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Issue{}, nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var issues []*Issue
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var issue Issue
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			// Log warning but continue - skip invalid lines
			// In production, we might want to log this
			continue
		}

		issues = append(issues, &issue)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return issues, nil
}

func (s *Store) writeAllUnlocked(issues []*Issue) error {
	// Write to temp file first
	tmpPath := s.path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	writer := bufio.NewWriter(f)
	for _, issue := range issues {
		data, err := json.Marshal(issue)
		if err != nil {
			f.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("failed to marshal issue: %w", err)
		}
		if _, err := fmt.Fprintf(writer, "%s\n", data); err != nil {
			f.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("failed to write issue: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, s.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

func (s *Store) matchesFilter(issue *Issue, filter ListFilter) bool {
	if filter.Status != nil && issue.Status != *filter.Status {
		return false
	}
	if filter.IssueType != nil && issue.IssueType != *filter.IssueType {
		return false
	}
	if filter.Priority != nil && issue.Priority != *filter.Priority {
		return false
	}
	if filter.Blocked && len(issue.BlockedBy) == 0 {
		return false
	}
	for _, label := range filter.Labels {
		if !contains(issue.Labels, label) {
			return false
		}
	}
	return true
}

// NowFunc is a variable for testing purposes
var NowFunc = time.Now

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ListAllSpecs lists issues across all spec directories
func ListAllSpecs(basePath string, filter ListFilter) ([]Issue, error) {
	if basePath == "" {
		basePath = "specledger"
	}

	// Get all spec directories
	specs, err := listSpecDirs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list spec directories: %w", err)
	}

	var allIssues []Issue
	for _, spec := range specs {
		store, err := NewStore(StoreOptions{
			BasePath:    basePath,
			SpecContext: spec,
		})
		if err != nil {
			continue
		}

		issues, err := store.List(filter)
		if err != nil {
			continue
		}

		allIssues = append(allIssues, issues...)
	}

	return allIssues, nil
}

// listSpecDirs lists all spec directories in the base path
func listSpecDirs(basePath string) ([]string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	var specs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Check if it matches the spec pattern (###-name) or is a special directory
		if specBranchPattern.MatchString(name) || name == "migrated" {
			// Check if issues.jsonl exists
			issuesPath := filepath.Join(basePath, name, "issues.jsonl")
			if _, err := os.Stat(issuesPath); err == nil {
				specs = append(specs, name)
			}
		}
	}

	return specs, nil
}

// GetIssueAcrossSpecs searches for an issue across all specs
func GetIssueAcrossSpecs(id, basePath string) (*Issue, string, error) {
	if basePath == "" {
		basePath = "specledger"
	}

	specs, err := listSpecDirs(basePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list spec directories: %w", err)
	}

	for _, spec := range specs {
		store, err := NewStore(StoreOptions{
			BasePath:    basePath,
			SpecContext: spec,
		})
		if err != nil {
			continue
		}

		issue, err := store.Get(id)
		if err == nil {
			return issue, spec, nil
		}
	}

	return nil, "", ErrIssueNotFound
}

// ReadyIssue represents an issue that is ready to work on
type ReadyIssue struct {
	Issue     Issue     `json:"issue"`
	BlockedBy []Blocker `json:"blocked_by,omitempty"` // Empty for ready issues
}

// ListReady returns all issues that are ready to work on (not blocked by open dependencies).
// Ready issues have status open or in_progress and all their blockers are closed.
func (s *Store) ListReady(filter ListFilter) ([]ReadyIssue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	issues, err := s.readAllUnlocked()
	if err != nil {
		return nil, err
	}

	// Build lookup map for dependency resolution
	issueMap := make(map[string]*Issue)
	for _, issue := range issues {
		issueMap[issue.ID] = issue
	}

	var result []ReadyIssue
	for _, issue := range issues {
		// Check if ready
		if !issue.IsReady(issueMap) {
			continue
		}

		// Apply additional filters
		if filter.Status != nil && issue.Status != *filter.Status {
			continue
		}
		if filter.IssueType != nil && issue.IssueType != *filter.IssueType {
			continue
		}
		if filter.Priority != nil && issue.Priority != *filter.Priority {
			continue
		}
		for _, label := range filter.Labels {
			if !contains(issue.Labels, label) {
				continue
			}
		}

		result = append(result, ReadyIssue{
			Issue:     *issue,
			BlockedBy: []Blocker{}, // Empty for ready issues
		})
	}

	return result, nil
}

// ListReadyAcrossSpecs returns ready issues across all spec directories.
func ListReadyAcrossSpecs(basePath string, filter ListFilter) ([]ReadyIssue, error) {
	if basePath == "" {
		basePath = "specledger"
	}

	specs, err := listSpecDirs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list spec directories: %w", err)
	}

	var allReady []ReadyIssue
	for _, spec := range specs {
		store, err := NewStore(StoreOptions{
			BasePath:    basePath,
			SpecContext: spec,
		})
		if err != nil {
			continue
		}

		ready, err := store.ListReady(filter)
		if err != nil {
			continue
		}

		allReady = append(allReady, ready...)
	}

	return allReady, nil
}

// GetBlockedIssuesWithBlockers returns all issues that are currently blocked, along with their blocker details.
func (s *Store) GetBlockedIssuesWithBlockers() ([]ReadyIssue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	issues, err := s.readAllUnlocked()
	if err != nil {
		return nil, err
	}

	// Build lookup map for dependency resolution
	issueMap := make(map[string]*Issue)
	for _, issue := range issues {
		issueMap[issue.ID] = issue
	}

	var result []ReadyIssue
	for _, issue := range issues {
		// Skip closed issues
		if issue.Status == StatusClosed {
			continue
		}

		// Check if blocked (has open blockers)
		if issue.IsReady(issueMap) {
			continue
		}

		// Get blocker details
		blockers := issue.GetBlockers(issueMap)
		if len(blockers) == 0 {
			continue // Not actually blocked (no blocker references found)
		}

		result = append(result, ReadyIssue{
			Issue:     *issue,
			BlockedBy: blockers,
		})
	}

	return result, nil
}

// RepairResult contains the result of repairing an issues file
type RepairResult struct {
	ValidLines      int
	InvalidLines    int
	RecoveredIssues int
	SkippedLines    []SkippedLine
	BackupPath      string
}

// SkippedLine contains information about a skipped line during repair
type SkippedLine struct {
	LineNum int
	Reason  string
}

// RepairIssuesFile repairs a corrupted issues.jsonl file
func RepairIssuesFile(path string) (*RepairResult, error) {
	result := &RepairResult{}

	// Read the file
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrIssueNotFound
		}
		return nil, err
	}
	defer f.Close()

	var validIssues []Issue
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		result.ValidLines++

		var issue Issue
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			result.InvalidLines++
			result.SkippedLines = append(result.SkippedLines, SkippedLine{
				LineNum: lineNum,
				Reason:  "Invalid JSON",
			})
			continue
		}

		// Validate issue
		if err := issue.Validate(); err != nil {
			result.InvalidLines++
			result.SkippedLines = append(result.SkippedLines, SkippedLine{
				LineNum: lineNum,
				Reason:  err.Error(),
			})
			continue
		}

		validIssues = append(validIssues, issue)
		result.RecoveredIssues++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Create backup
	backupPath := path + ".bak"
	if err := copyFile(path, backupPath); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}
	result.BackupPath = backupPath

	// Write valid issues back
	tmpPath := path + ".tmp"
	tf, err := os.Create(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	writer := bufio.NewWriter(tf)
	for _, issue := range validIssues {
		data, err := json.Marshal(issue)
		if err != nil {
			tf.Close()
			os.Remove(tmpPath)
			return nil, fmt.Errorf("failed to marshal issue: %w", err)
		}
		if _, err := fmt.Fprintf(writer, "%s\n", data); err != nil {
			tf.Close()
			os.Remove(tmpPath)
			return nil, fmt.Errorf("failed to write issue: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		tf.Close()
		os.Remove(tmpPath)
		return nil, fmt.Errorf("failed to flush writer: %w", err)
	}

	tf.Close()

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("failed to rename file: %w", err)
	}

	return result, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
