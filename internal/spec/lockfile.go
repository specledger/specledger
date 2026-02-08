package spec

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Lockfile represents the lockfile (spec.sum) that contains resolved dependencies
type Lockfile struct {
	Version   string
	Entries   []LockfileEntry
	Timestamp time.Time
	TotalSize int64
}

// LockfileEntry represents a single entry in the lockfile
type LockfileEntry struct {
	RepositoryURL string `json:"repository_url"`
	CommitHash    string `json:"commit_hash"`
	ContentHash   string `json:"content_hash"`
	SpecPath      string `json:"spec_path"`
	Branch        string `json:"branch,omitempty"`
	Size          int64  `json:"size"`
	FetchedAt     string `json:"fetched_at"`
}

// NewLockfile creates a new lockfile
func NewLockfile(version string) *Lockfile {
	return &Lockfile{
		Version:   version,
		Entries:   make([]LockfileEntry, 0),
		Timestamp: time.Now(),
	}
}

// CalculateSHA256 calculates the SHA-256 hash of the lockfile content
func (l *Lockfile) CalculateSHA256() (string, error) {
	data, err := json.Marshal(l)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// CalculateSHA256FromBytes calculates SHA-256 hash of byte data
func CalculateSHA256FromBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Verify verifies that the lockfile content matches the current content
func (l *Lockfile) Verify(manifest *Manifest) ([]string, error) {
	var issues []string

	// Check that all manifest dependencies have entries
	for _, dep := range manifest.Dependecies {
		found := false
		for _, entry := range l.Entries {
			if entry.RepositoryURL == dep.RepositoryURL && entry.SpecPath == dep.SpecPath {
				found = true
				break
			}
		}
		if !found {
			issues = append(issues, fmt.Sprintf("missing entry for dependency: %s %s", dep.RepositoryURL, dep.SpecPath))
		}
	}

	// Check that all entries have valid hashes
	for _, entry := range l.Entries {
		if entry.ContentHash == "" {
			issues = append(issues, fmt.Sprintf("empty content hash for: %s %s", entry.RepositoryURL, entry.SpecPath))
		}
		if len(entry.ContentHash) != 64 {
			issues = append(issues, fmt.Sprintf("invalid content hash length for: %s %s (got %d, want 64)", entry.RepositoryURL, entry.SpecPath, len(entry.ContentHash)))
		}
	}

	if len(issues) > 0 {
		return issues, fmt.Errorf("lockfile verification failed: %d issues found", len(issues))
	}

	return nil, nil
}

// Write writes the lockfile to disk
func (l *Lockfile) Write(path string) error {
	// Create directories if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lockfile: %w", err)
	}

	// Write to file
	// #nosec G306 -- lockfile needs to be readable, 0644 is appropriate
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write lockfile: %w", err)
	}

	return nil
}

// Read reads a lockfile from disk
func ReadLockfile(path string) (*Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read lockfile: %w", err)
	}

	var lockfile Lockfile
	if err := json.Unmarshal(data, &lockfile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lockfile: %w", err)
	}

	return &lockfile, nil
}

// AddEntry adds an entry to the lockfile
func (l *Lockfile) AddEntry(entry LockfileEntry) {
	l.Entries = append(l.Entries, entry)
	l.TotalSize += entry.Size
}

// RemoveEntry removes an entry from the lockfile
func (l *Lockfile) RemoveEntry(repoURL, specPath string) bool {
	for i, e := range l.Entries {
		if e.RepositoryURL == repoURL && e.SpecPath == specPath {
			l.Entries = append(l.Entries[:i], l.Entries[i+1:]...)
			l.TotalSize -= e.Size
			return true
		}
	}
	return false
}

// GetEntry retrieves an entry from the lockfile
func (l *Lockfile) GetEntry(repoURL, specPath string) (*LockfileEntry, bool) {
	for _, entry := range l.Entries {
		if entry.RepositoryURL == repoURL && entry.SpecPath == specPath {
			return &entry, true
		}
	}
	return nil, false
}

// GetRepositoryEntries retrieves all entries for a given repository
func (l *Lockfile) GetRepositoryEntries(repoURL string) []LockfileEntry {
	var entries []LockfileEntry
	for _, entry := range l.Entries {
		if entry.RepositoryURL == repoURL {
			entries = append(entries, entry)
		}
	}
	return entries
}

// GetContentHash retrieves the content hash for a dependency
func (l *Lockfile) GetContentHash(repoURL, specPath string) (string, bool) {
	entry, ok := l.GetEntry(repoURL, specPath)
	if !ok {
		return "", false
	}
	return entry.ContentHash, true
}
