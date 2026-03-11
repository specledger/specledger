package memory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
)

// Store manages JSONL file operations for knowledge entries with file locking.
type Store struct {
	path string // Path to entries.jsonl
	lock *flock.Flock
	mu   sync.Mutex
}

// NewStore creates a new knowledge entry store.
// basePath is typically ".specledger/memory/cache".
func NewStore(basePath string) (*Store, error) {
	if basePath == "" {
		basePath = filepath.Join(".specledger", "memory", "cache")
	}

	entriesPath := filepath.Join(basePath, "entries.jsonl")
	lockPath := filepath.Join(basePath, "entries.jsonl.lock")

	return &Store{
		path: entriesPath,
		lock: flock.New(lockPath),
	}, nil
}

// Path returns the path to the entries.jsonl file.
func (s *Store) Path() string {
	return s.path
}

// Create adds a new knowledge entry to the store.
func (s *Store) Create(entry *KnowledgeEntry) error {
	return s.withLock(func() error {
		if err := entry.Validate(); err != nil {
			return err
		}

		// Check for duplicates
		existing, err := s.getByIDUnlocked(entry.ID)
		if err != nil && err != ErrEntryNotFound {
			return err
		}
		if existing != nil {
			return ErrEntryAlreadyExists
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

		data, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal entry: %w", err)
		}

		_, err = fmt.Fprintf(f, "%s\n", data)
		return err
	})
}

// Get retrieves a knowledge entry by ID.
func (s *Store) Get(id string) (*KnowledgeEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getByIDUnlocked(id)
}

func (s *Store) getByIDUnlocked(id string) (*KnowledgeEntry, error) {
	entries, err := s.readAllUnlocked()
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.ID == id {
			return entry, nil
		}
	}

	return nil, ErrEntryNotFound
}

// List returns all knowledge entries, optionally filtered by status.
func (s *Store) List(status *EntryStatus) ([]*KnowledgeEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := s.readAllUnlocked()
	if err != nil {
		return nil, err
	}

	if status == nil {
		return entries, nil
	}

	var filtered []*KnowledgeEntry
	for _, entry := range entries {
		if entry.Status == *status {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

// ListPromoted returns all promoted entries.
func (s *Store) ListPromoted() ([]*KnowledgeEntry, error) {
	status := StatusPromoted
	return s.List(&status)
}

// Update modifies an existing knowledge entry.
func (s *Store) Update(entry *KnowledgeEntry) error {
	return s.withLock(func() error {
		entries, err := s.readAllUnlocked()
		if err != nil {
			return err
		}

		found := false
		for i, e := range entries {
			if e.ID == entry.ID {
				entry.UpdatedAt = time.Now()
				if err := entry.Validate(); err != nil {
					return err
				}
				entries[i] = entry
				found = true
				break
			}
		}

		if !found {
			return ErrEntryNotFound
		}

		return s.writeAllUnlocked(entries)
	})
}

// Delete removes a knowledge entry by ID.
func (s *Store) Delete(id string) error {
	return s.withLock(func() error {
		entries, err := s.readAllUnlocked()
		if err != nil {
			return err
		}

		var newEntries []*KnowledgeEntry
		found := false
		for _, entry := range entries {
			if entry.ID == id {
				found = true
				continue
			}
			newEntries = append(newEntries, entry)
		}

		if !found {
			return ErrEntryNotFound
		}

		return s.writeAllUnlocked(newEntries)
	})
}

// FindSimilar finds entries with similar titles (case-insensitive substring match).
func (s *Store) FindSimilar(title string) ([]*KnowledgeEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := s.readAllUnlocked()
	if err != nil {
		return nil, err
	}

	titleLower := strings.ToLower(title)
	var similar []*KnowledgeEntry
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Title), titleLower) ||
			strings.Contains(titleLower, strings.ToLower(entry.Title)) {
			similar = append(similar, entry)
		}
	}
	return similar, nil
}

// Merge combines a new entry with an existing one, incrementing recurrence.
func (s *Store) Merge(existingID string, newEntry *KnowledgeEntry) error {
	return s.withLock(func() error {
		entries, err := s.readAllUnlocked()
		if err != nil {
			return err
		}

		for i, entry := range entries {
			if entry.ID == existingID {
				entry.RecurrenceCount++
				entry.Scores.Recurrence = clampScore(entry.Scores.Recurrence + 1.0)
				entry.Scores.Composite = CalculateComposite(entry.Scores)
				entry.UpdatedAt = time.Now()

				// Append source info if different
				if newEntry.SourceSessionID != "" && newEntry.SourceSessionID != entry.SourceSessionID {
					entry.SourceSessionID = newEntry.SourceSessionID
				}

				entries[i] = entry
				return s.writeAllUnlocked(entries)
			}
		}

		return ErrEntryNotFound
	})
}

// Promote sets an entry's status to promoted.
func (s *Store) Promote(id string) (*KnowledgeEntry, error) {
	var result *KnowledgeEntry
	err := s.withLock(func() error {
		entries, err := s.readAllUnlocked()
		if err != nil {
			return err
		}

		for i, entry := range entries {
			if entry.ID == id {
				entry.Status = StatusPromoted
				entry.UpdatedAt = time.Now()
				entries[i] = entry
				result = entry
				return s.writeAllUnlocked(entries)
			}
		}
		return ErrEntryNotFound
	})
	return result, err
}

// Demote sets an entry's status back to candidate.
func (s *Store) Demote(id string) (*KnowledgeEntry, error) {
	var result *KnowledgeEntry
	err := s.withLock(func() error {
		entries, err := s.readAllUnlocked()
		if err != nil {
			return err
		}

		for i, entry := range entries {
			if entry.ID == id {
				entry.Status = StatusCandidate
				entry.UpdatedAt = time.Now()
				entries[i] = entry
				result = entry
				return s.writeAllUnlocked(entries)
			}
		}
		return ErrEntryNotFound
	})
	return result, err
}

func (s *Store) withLock(fn func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	locked, err := s.lock.TryLock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("store is locked by another process")
	}
	defer func() { _ = s.lock.Unlock() }()

	return fn()
}

func (s *Store) readAllUnlocked() ([]*KnowledgeEntry, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*KnowledgeEntry{}, nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var entries []*KnowledgeEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry KnowledgeEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip invalid lines
		}

		entries = append(entries, &entry)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return entries, nil
}

func (s *Store) writeAllUnlocked(entries []*KnowledgeEntry) error {
	tmpPath := s.path + ".tmp"

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	writer := bufio.NewWriter(f)
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			f.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("failed to marshal entry: %w", err)
		}
		if _, err := fmt.Fprintf(writer, "%s\n", data); err != nil {
			f.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("failed to write entry: %w", err)
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

	if err := os.Rename(tmpPath, s.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// clampScore ensures a score stays within [0, 10].
func clampScore(v float64) float64 {
	if v < MinScore {
		return MinScore
	}
	if v > MaxScore {
		return MaxScore
	}
	return v
}
