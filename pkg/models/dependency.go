package models

import (
	"fmt"
	"strings"
	"time"
)

// Dependency represents a single external specification dependency declaration
type Dependency struct {
	RepositoryURL string   // Git repository URL (HTTPS or SSH)
	Version       string   // Version constraint (branch, tag, commit hash, or semver)
	SpecPath      string   // Path to the specification file within the repository
	Alias         string   // Optional short alias for the dependency
	Pinned        bool     // Whether the dependency is pinned to a specific commit
	Transitive    []Dependency // Transitive dependencies discovered from external spec.mod
}

// String returns a string representation of the dependency
func (d *Dependency) String() string {
	if d.Alias != "" {
		return fmt.Sprintf("%s -> %s", d.Alias, d.RepositoryURL)
	}
	return d.RepositoryURL
}

// Validate checks if the dependency is valid
func (d *Dependency) Validate() error {
	if d.RepositoryURL == "" {
		return fmt.Errorf("repository URL cannot be empty")
	}

	if d.Version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	if d.SpecPath == "" {
		return fmt.Errorf("spec path cannot be empty")
	}

	if d.Alias != "" {
		if !isValidAlias(d.Alias) {
			return fmt.Errorf("invalid alias: must be alphanumeric with hyphens or underscores")
		}
	}

	return nil
}

// isValidAlias checks if the alias is valid
func isValidAlias(alias string) bool {
	// Alphanumeric, hyphens, underscores only
	return len(alias) > 0 && len(alias) <= 50 &&
		strings.HasPrefix(alias, "#") ||
		(alias[0] == '#' && len(alias) > 1) ||
		(alias[0] == '.' && len(alias) > 1)
}

// DependencyManifest represents the dependency manifest file (spec.mod)
type DependencyManifest struct {
	Version     string           // Version of the manifest format
	Dependencies []Dependency   // List of declared dependencies
	ID          string           // Unique identifier for this spec (for external references)
	Path        string           // File path to this spec (relative to repo root)
	CreatedAt   time.Time        // When the manifest was created
	UpdatedAt   time.Time        // When the manifest was last updated
}

// NewDependencyManifest creates a new dependency manifest
func NewDependencyManifest(version, id, path string) *DependencyManifest {
	now := time.Now()
	return &DependencyManifest{
		Version:    version,
		Dependencies: make([]Dependency, 0),
		ID:         id,
		Path:       path,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// AddDependency adds a dependency to the manifest
func (m *DependencyManifest) AddDependency(dep Dependency) error {
	if err := dep.Validate(); err != nil {
		return err
	}

	// Check for duplicate dependencies
	for _, d := range m.Dependencies {
		if d.RepositoryURL == dep.RepositoryURL && d.SpecPath == dep.SpecPath {
			return fmt.Errorf("duplicate dependency: %s %s", dep.RepositoryURL, dep.SpecPath)
		}
	}

	m.Dependencies = append(m.Dependencies, dep)
	m.UpdatedAt = time.Now()
	return nil
}

// RemoveDependency removes a dependency from the manifest
func (m *DependencyManifest) RemoveDependency(repoURL, specPath string) bool {
	for i, d := range m.Dependencies {
		if d.RepositoryURL == repoURL && d.SpecPath == specPath {
			m.Dependencies = append(m.Dependencies[:i], m.Dependencies[i+1:]...)
			m.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// LockfileEntry represents a single resolved dependency entry in the lockfile
type LockfileEntry struct {
	RepositoryURL  string
	CommitHash     string
	ContentHash    string // SHA-256 hash of the spec file content
	SpecPath       string
	Branch         string
	Size           int64
	FetchedAt      time.Time
}

// Lockfile represents the lockfile file (spec.sum) with cryptographic verification
type Lockfile struct {
	Version    string
	Entries    []LockfileEntry
	Timestamp  time.Time
	TotalSize  int64
}

// NewLockfile creates a new lockfile
func NewLockfile(version string) *Lockfile {
	return &Lockfile{
		Version:   version,
		Entries:   make([]LockfileEntry, 0),
		Timestamp: time.Now(),
	}
}

// AddEntry adds an entry to the lockfile
func (l *Lockfile) AddEntry(entry LockfileEntry) {
	l.Entries = append(l.Entries, entry)
	l.TotalSize += entry.Size
}

// SpecFile represents a specification file
type SpecFile struct {
	Path     string
	Content  []byte
	Modified time.Time
}
