package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// MaxRetries is the maximum number of upload retries
	MaxRetries = 3
	// RetryDelay is the initial delay between retries
	RetryDelay = 5 * time.Second
	// QueueIndexFile is the name of the queue index file
	QueueIndexFile = ".queue.json"
)

// GetBaseDir returns the base path for local session storage (~/.specledger)
func GetBaseDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("HOME")
	}
	return filepath.Join(homeDir, ".specledger")
}

// GetSessionDir returns the path for sessions: ~/.specledger/{project_id}/{spec_key}/
func GetSessionDir(projectID, specKey string) string {
	return filepath.Join(GetBaseDir(), projectID, specKey)
}

// GetSessionPath returns the full path for a session file
// Path: ~/.specledger/{project_id}/{spec_key}/{identifier}.json.gz
func GetSessionPath(projectID, specKey, identifier string) string {
	return filepath.Join(GetSessionDir(projectID, specKey), identifier+".json.gz")
}

// GetSessionMetaPath returns the path for session metadata file
func GetSessionMetaPath(projectID, specKey, identifier string) string {
	return filepath.Join(GetSessionDir(projectID, specKey), identifier+".meta.json")
}

// QueueIndex tracks pending uploads across all projects/branches
type QueueIndex struct {
	Pending []QueueRef `json:"pending"`
}

// QueueRef references a queued session by its location
type QueueRef struct {
	ProjectID  string    `json:"project_id"`
	SpecKey    string    `json:"spec_key"`
	Identifier string    `json:"identifier"`
	QueuedAt   time.Time `json:"queued_at"`
}

// Queue manages the local session upload queue
type Queue struct {
	baseDir string
}

// NewQueue creates a new queue manager
func NewQueue() *Queue {
	return &Queue{baseDir: GetBaseDir()}
}

// ensureDir ensures a directory exists
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// getIndexPath returns the path to the queue index file
func (q *Queue) getIndexPath() string {
	return filepath.Join(q.baseDir, QueueIndexFile)
}

// loadIndex loads the queue index
func (q *Queue) loadIndex() (*QueueIndex, error) {
	indexPath := q.getIndexPath()
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &QueueIndex{Pending: []QueueRef{}}, nil
		}
		return nil, fmt.Errorf("failed to read queue index: %w", err)
	}

	var index QueueIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse queue index: %w", err)
	}

	return &index, nil
}

// saveIndex saves the queue index
func (q *Queue) saveIndex(index *QueueIndex) error {
	if err := ensureDir(q.baseDir); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal queue index: %w", err)
	}

	if err := os.WriteFile(q.getIndexPath(), data, 0600); err != nil {
		return fmt.Errorf("failed to write queue index: %w", err)
	}

	return nil
}

// Enqueue adds a session to the upload queue
// Stores at: ~/.specledger/{project_id}/{spec_key}/{identifier}.json.gz
func (q *Queue) Enqueue(entry *QueueEntry, compressedData []byte) error {
	// Determine identifier (commit hash or task ID)
	var identifier string
	if entry.CommitHash != nil {
		identifier = *entry.CommitHash
	} else if entry.TaskID != nil {
		identifier = *entry.TaskID
	} else {
		identifier = entry.SessionID
	}

	// Create directory structure
	sessionDir := GetSessionDir(entry.ProjectID, entry.FeatureBranch)
	if err := ensureDir(sessionDir); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	// Write compressed data
	dataPath := GetSessionPath(entry.ProjectID, entry.FeatureBranch, identifier)
	if err := os.WriteFile(dataPath, compressedData, 0600); err != nil {
		return fmt.Errorf("failed to write session data: %w", err)
	}

	// Write metadata
	metaPath := GetSessionMetaPath(entry.ProjectID, entry.FeatureBranch, identifier)
	metaData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		_ = os.Remove(dataPath)
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := os.WriteFile(metaPath, metaData, 0600); err != nil {
		_ = os.Remove(dataPath)
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Add to queue index
	index, err := q.loadIndex()
	if err != nil {
		return fmt.Errorf("failed to load queue index: %w", err)
	}

	index.Pending = append(index.Pending, QueueRef{
		ProjectID:  entry.ProjectID,
		SpecKey:    entry.FeatureBranch,
		Identifier: identifier,
		QueuedAt:   time.Now(),
	})

	if err := q.saveIndex(index); err != nil {
		return fmt.Errorf("failed to save queue index: %w", err)
	}

	return nil
}

// Dequeue removes a session from the queue
func (q *Queue) Dequeue(projectID, specKey, identifier string) error {
	// Remove files
	dataPath := GetSessionPath(projectID, specKey, identifier)
	metaPath := GetSessionMetaPath(projectID, specKey, identifier)

	_ = os.Remove(dataPath)
	_ = os.Remove(metaPath)

	// Remove from queue index
	index, err := q.loadIndex()
	if err != nil {
		return err
	}

	newPending := make([]QueueRef, 0, len(index.Pending))
	for _, ref := range index.Pending {
		if ref.ProjectID != projectID || ref.SpecKey != specKey || ref.Identifier != identifier {
			newPending = append(newPending, ref)
		}
	}
	index.Pending = newPending

	return q.saveIndex(index)
}

// GetQueuedSession retrieves a queued session's data and metadata
func (q *Queue) GetQueuedSession(projectID, specKey, identifier string) ([]byte, *QueueEntry, error) {
	dataPath := GetSessionPath(projectID, specKey, identifier)
	metaPath := GetSessionMetaPath(projectID, specKey, identifier)

	// Read compressed data
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read session data: %w", err)
	}

	// Read metadata
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var entry QueueEntry
	if err := json.Unmarshal(metaData, &entry); err != nil {
		return nil, nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return data, &entry, nil
}

// List returns all pending queue references
func (q *Queue) List() ([]QueueRef, error) {
	index, err := q.loadIndex()
	if err != nil {
		return nil, err
	}

	// Filter out entries whose files no longer exist
	validRefs := make([]QueueRef, 0, len(index.Pending))
	for _, ref := range index.Pending {
		dataPath := GetSessionPath(ref.ProjectID, ref.SpecKey, ref.Identifier)
		if _, err := os.Stat(dataPath); err == nil {
			validRefs = append(validRefs, ref)
		}
	}

	// Update index if we removed any stale entries
	if len(validRefs) != len(index.Pending) {
		index.Pending = validRefs
		_ = q.saveIndex(index)
	}

	return validRefs, nil
}

// Count returns the number of sessions in the queue
func (q *Queue) Count() (int, error) {
	refs, err := q.List()
	if err != nil {
		return 0, err
	}
	return len(refs), nil
}

// ListEntries returns all queue entries with their metadata
func (q *Queue) ListEntries() ([]*QueueEntry, error) {
	refs, err := q.List()
	if err != nil {
		return nil, err
	}

	var entries []*QueueEntry
	for _, ref := range refs {
		_, entry, err := q.GetQueuedSession(ref.ProjectID, ref.SpecKey, ref.Identifier)
		if err != nil {
			continue // Skip entries we can't read
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// UpdateRetryCount updates the retry count for a queued session
func (q *Queue) UpdateRetryCount(projectID, specKey, identifier string) error {
	metaPath := GetSessionMetaPath(projectID, specKey, identifier)

	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	var entry QueueEntry
	if err := json.Unmarshal(metaData, &entry); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	entry.RetryCount++
	now := time.Now()
	entry.LastRetry = &now

	updatedData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, updatedData, 0600); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// ShouldRetry checks if a session should be retried
func (q *Queue) ShouldRetry(entry *QueueEntry) bool {
	if entry.RetryCount >= MaxRetries {
		return false
	}

	if entry.LastRetry != nil {
		// Exponential backoff
		delay := RetryDelay * time.Duration(1<<entry.RetryCount)
		if time.Since(*entry.LastRetry) < delay {
			return false
		}
	}

	return true
}

// ProcessQueue attempts to upload all queued sessions
func (q *Queue) ProcessQueue(accessToken string) (uploaded int, failed int, skipped int, errors []error) {
	refs, err := q.List()
	if err != nil {
		return 0, 0, 0, []error{err}
	}

	storage := NewStorageClient()
	metadata := NewMetadataClient()

	for _, ref := range refs {
		data, entry, err := q.GetQueuedSession(ref.ProjectID, ref.SpecKey, ref.Identifier)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get session %s: %w", ref.Identifier, err))
			failed++
			continue
		}

		if !q.ShouldRetry(entry) {
			skipped++
			continue
		}

		// Build storage path (same structure as local)
		storagePath := BuildStoragePath(ref.ProjectID, ref.SpecKey, ref.Identifier)

		// Try to upload
		_, err = storage.Upload(accessToken, storagePath, data)
		if err != nil {
			_ = q.UpdateRetryCount(ref.ProjectID, ref.SpecKey, ref.Identifier)
			errors = append(errors, fmt.Errorf("failed to upload session %s: %w", ref.Identifier, err))
			failed++
			continue
		}

		// Create metadata
		_, err = metadata.Create(accessToken, &CreateSessionInput{
			ProjectID:     entry.ProjectID,
			FeatureBranch: entry.FeatureBranch,
			CommitHash:    entry.CommitHash,
			TaskID:        entry.TaskID,
			AuthorID:      entry.AuthorID,
			StoragePath:   storagePath,
			Status:        entry.Status,
			SizeBytes:     int64(len(data)),
			RawSizeBytes:  0, // We don't track raw size in queue
			MessageCount:  0, // We don't track message count in queue
		})
		if err != nil {
			_ = q.UpdateRetryCount(ref.ProjectID, ref.SpecKey, ref.Identifier)
			errors = append(errors, fmt.Errorf("failed to create metadata for session %s: %w", ref.Identifier, err))
			failed++
			continue
		}

		// Success - remove from queue
		_ = q.Dequeue(ref.ProjectID, ref.SpecKey, ref.Identifier)
		uploaded++
	}

	return uploaded, failed, skipped, errors
}

// GetLocalSession retrieves a session from local storage (not necessarily queued)
func GetLocalSession(projectID, specKey, identifier string) ([]byte, error) {
	dataPath := GetSessionPath(projectID, specKey, identifier)
	return os.ReadFile(dataPath)
}

// SaveLocalSession saves a session to local storage
func SaveLocalSession(projectID, specKey, identifier string, data []byte) error {
	sessionDir := GetSessionDir(projectID, specKey)
	if err := ensureDir(sessionDir); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	dataPath := GetSessionPath(projectID, specKey, identifier)
	return os.WriteFile(dataPath, data, 0600)
}

// ListLocalSessions lists all local sessions for a project/branch
func ListLocalSessions(projectID, specKey string) ([]string, error) {
	sessionDir := GetSessionDir(projectID, specKey)
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var identifiers []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json.gz") && !strings.HasSuffix(name, ".meta.json") {
			identifier := strings.TrimSuffix(name, ".json.gz")
			identifiers = append(identifiers, identifier)
		}
	}

	return identifiers, nil
}
