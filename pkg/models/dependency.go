package models

import (
	"fmt"
	"time"
)

// Dependency represents a single external specification dependency declaration
type Dependency struct {
	RepositoryURL string       // Git repository URL (HTTPS or SSH)
	Version       string       // Version constraint (branch, tag, commit hash, or semver)
	SpecPath      string       // Path to the specification file within the repository
	Alias         string       // Optional short alias for the dependency
	Pinned        bool         // Whether the dependency is pinned to a specific commit
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
	if len(alias) == 0 || len(alias) > 50 {
		return false
	}

	for _, ch := range alias {
		isValid := (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' ||
			ch == '_'

		if !isValid {
			return false
		}
	}

	return true
}

// DependencyManifest represents the dependency manifest file (spec.mod)
type DependencyManifest struct {
	Version      string       // Version of the manifest format
	Dependencies []Dependency // List of declared dependencies
	ID           string       // Unique identifier for this spec (for external references)
	Path         string       // File path to this spec (relative to repo root)
	CreatedAt    time.Time    // When the manifest was created
	UpdatedAt    time.Time    // When the manifest was last updated
}

// LockfileEntry represents a single resolved dependency entry in the lockfile
type LockfileEntry struct {
	RepositoryURL string
	CommitHash    string
	ContentHash   string // SHA-256 hash of the spec file content
	SpecPath      string
	Branch        string
	Size          int64
	FetchedAt     time.Time
}

// Lockfile represents the lockfile file (spec.sum) with cryptographic verification
type Lockfile struct {
	Version   string
	Entries   []LockfileEntry
	Timestamp time.Time
	TotalSize int64
}

// SpecFile represents a specification file
type SpecFile struct {
	Path     string
	Content  []byte
	Modified time.Time
}
