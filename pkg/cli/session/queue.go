package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// MaxRetries is the maximum number of upload retries
	MaxRetries = 3
	// RetryDelay is the initial delay between retries
	RetryDelay = 5 * time.Second
)

// GetQueueDir returns the path to the session queue directory
func GetQueueDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("HOME")
	}
	return filepath.Join(homeDir, ".specledger", "session-queue")
}

// Queue manages the local session upload queue
type Queue struct {
	dir string
}

// NewQueue creates a new queue manager
func NewQueue() *Queue {
	return &Queue{dir: GetQueueDir()}
}

// ensureDir ensures the queue directory exists
func (q *Queue) ensureDir() error {
	return os.MkdirAll(q.dir, 0755)
}

// Enqueue adds a session to the upload queue
func (q *Queue) Enqueue(sessionID string, compressedData []byte, entry *QueueEntry) error {
	if err := q.ensureDir(); err != nil {
		return fmt.Errorf("failed to create queue directory: %w", err)
	}

	// Write compressed data
	dataPath := filepath.Join(q.dir, sessionID+".json.gz")
	if err := os.WriteFile(dataPath, compressedData, 0600); err != nil {
		return fmt.Errorf("failed to write session data: %w", err)
	}

	// Write metadata
	metaPath := filepath.Join(q.dir, sessionID+".meta.json")
	metaData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := os.WriteFile(metaPath, metaData, 0600); err != nil {
		// Clean up data file
		_ = os.Remove(dataPath)
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// Dequeue removes a session from the queue
func (q *Queue) Dequeue(sessionID string) error {
	dataPath := filepath.Join(q.dir, sessionID+".json.gz")
	metaPath := filepath.Join(q.dir, sessionID+".meta.json")

	_ = os.Remove(dataPath)
	_ = os.Remove(metaPath)

	return nil
}

// GetQueuedSession retrieves a queued session's data and metadata
func (q *Queue) GetQueuedSession(sessionID string) ([]byte, *QueueEntry, error) {
	dataPath := filepath.Join(q.dir, sessionID+".json.gz")
	metaPath := filepath.Join(q.dir, sessionID+".meta.json")

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

// List returns all session IDs in the queue
func (q *Queue) List() ([]string, error) {
	if err := q.ensureDir(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(q.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read queue directory: %w", err)
	}

	var sessionIDs []string
	seen := make(map[string]bool)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		var sessionID string

		if len(name) > 8 && name[len(name)-8:] == ".json.gz" {
			sessionID = name[:len(name)-8]
		} else if len(name) > 10 && name[len(name)-10:] == ".meta.json" {
			sessionID = name[:len(name)-10]
		}

		if sessionID != "" && !seen[sessionID] {
			// Only include if both files exist
			dataPath := filepath.Join(q.dir, sessionID+".json.gz")
			metaPath := filepath.Join(q.dir, sessionID+".meta.json")
			if _, err := os.Stat(dataPath); err == nil {
				if _, err := os.Stat(metaPath); err == nil {
					sessionIDs = append(sessionIDs, sessionID)
					seen[sessionID] = true
				}
			}
		}
	}

	return sessionIDs, nil
}

// Count returns the number of sessions in the queue
func (q *Queue) Count() (int, error) {
	sessionIDs, err := q.List()
	if err != nil {
		return 0, err
	}
	return len(sessionIDs), nil
}

// UpdateRetryCount updates the retry count for a queued session
func (q *Queue) UpdateRetryCount(sessionID string) error {
	metaPath := filepath.Join(q.dir, sessionID+".meta.json")

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
	sessionIDs, err := q.List()
	if err != nil {
		return 0, 0, 0, []error{err}
	}

	storage := NewStorageClient()
	metadata := NewMetadataClient()

	for _, sessionID := range sessionIDs {
		data, entry, err := q.GetQueuedSession(sessionID)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get session %s: %w", sessionID, err))
			failed++
			continue
		}

		if !q.ShouldRetry(entry) {
			skipped++
			continue
		}

		// Build storage path
		var identifier string
		if entry.CommitHash != nil {
			identifier = *entry.CommitHash
		} else if entry.TaskID != nil {
			identifier = *entry.TaskID
		} else {
			identifier = sessionID
		}
		storagePath := BuildStoragePath(entry.ProjectID, entry.FeatureBranch, identifier)

		// Try to upload
		_, err = storage.Upload(accessToken, storagePath, data)
		if err != nil {
			_ = q.UpdateRetryCount(sessionID)
			errors = append(errors, fmt.Errorf("failed to upload session %s: %w", sessionID, err))
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
			_ = q.UpdateRetryCount(sessionID)
			errors = append(errors, fmt.Errorf("failed to create metadata for session %s: %w", sessionID, err))
			failed++
			continue
		}

		// Success - remove from queue
		_ = q.Dequeue(sessionID)
		uploaded++
	}

	return uploaded, failed, skipped, errors
}
