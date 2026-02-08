package spec

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"specledger/pkg/models"
)

// ResolveResult represents the result of a dependency resolution
type ResolveResult struct {
	Dependency  *models.Dependency
	CommitHash  string
	Content     []byte
	ContentHash string
	Size        int64
	Source      string
}

// Resolver resolves dependencies by fetching specs from Git repositories
type Resolver struct {
	cacheDir string
}

// NewResolver creates a new resolver with global cache directory
func NewResolver(projectCacheDir string) *Resolver {
	// Use ~/.specledger/cache for dependency caching
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		if u, err := user.Current(); err == nil {
			homeDir = u.HomeDir
		}
	}

	globalCache := filepath.Join(homeDir, ".specledger", "cache")

	return &Resolver{
		cacheDir: globalCache,
	}
}

// Resolve resolves all dependencies in the manifest
func (r *Resolver) Resolve(ctx context.Context, manifest *Manifest, noCache bool) ([]ResolveResult, error) {
	results := make([]ResolveResult, 0, len(manifest.Dependecies))

	for _, dep := range manifest.Dependecies {
		result, err := r.resolveDependency(ctx, dep, noCache)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependency %s: %w", dep.RepositoryURL, err)
		}
		results = append(results, *result)
	}

	return results, nil
}

// resolveDependency resolves a single dependency
func (r *Resolver) resolveDependency(ctx context.Context, dep models.Dependency, noCache bool) (*ResolveResult, error) {
	// Check cache first if not disabled
	if !noCache {
		cached, err := r.getCachedContent(dep)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	// Fetch from Git
	content, commitHash, err := r.fetchFromGit(ctx, dep)
	if err != nil {
		return nil, err
	}

	// Calculate SHA-256 hash
	hasher := sha256.New()
	hasher.Write(content)
	contentHash := hex.EncodeToString(hasher.Sum(nil))

	// Create result
	result := &ResolveResult{
		Dependency:  &dep,
		CommitHash:  commitHash,
		Content:     content,
		ContentHash: contentHash,
		Size:        int64(len(content)),
		Source:      "remote",
	}

	// Save to cache
	if err := r.saveToCache(result); err != nil {
		// Warn but don't fail
		fmt.Printf("Warning: failed to cache dependency: %v\n", err)
	}

	return result, nil
}

// fetchFromGit fetches spec content from a Git repository
func (r *Resolver) fetchFromGit(ctx context.Context, dep models.Dependency) ([]byte, string, error) {
	// Convert repo URL to cache path
	cachePath := r.getRepoCachePath(dep.RepositoryURL)

	// Clone or fetch the repository
	repo, err := git.PlainCloneContext(ctx, cachePath, false, &git.CloneOptions{
		URL:      dep.RepositoryURL,
		Depth:    1, // Shallow clone for performance
		Progress: nil,
	})

	if err != nil {
		// If already exists, open it
		if err == git.ErrRepositoryAlreadyExists {
			repo, err = git.PlainOpen(cachePath)
			if err != nil {
				return nil, "", fmt.Errorf("failed to open repository: %w", err)
			}
		} else {
			return nil, "", fmt.Errorf("failed to clone repository: %w", err)
		}
	}

	// Get the branch/commit
	var commitHash plumbing.Hash

	// Handle version (branch, tag, or commit)
	rev, err := repo.ResolveRevision(plumbing.Revision(dep.Version))
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve version %s: %w", dep.Version, err)
	}
	commitHash = *rev

	// Checkout the specific commit
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Hash:  commitHash,
		Force: true,
	}); err != nil {
		return nil, "", fmt.Errorf("failed to checkout commit: %w", err)
	}

	// Read the spec file from the checked out repository
	specPath := filepath.Join(cachePath, dep.SpecPath)
	content, err := os.ReadFile(specPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read spec file %s: %w", dep.SpecPath, err)
	}

	return content, commitHash.String(), nil
}

// getRepoCachePath converts a repo URL to a cache path
// Example: git@github.com:org/repo -> ~/.cache/specledger/github.com/org/repo
func (r *Resolver) getRepoCachePath(repoURL string) string {
	// Remove git@ prefix and .git suffix
	url := strings.TrimPrefix(repoURL, "git@")
	url = strings.TrimSuffix(url, ".git")

	// Replace : with / for SSH URLs
	url = strings.Replace(url, ":", "/", 1)

	// Convert to filesystem path
	return filepath.Join(r.cacheDir, url)
}

// getCachedContent gets content from cache
func (r *Resolver) getCachedContent(dep models.Dependency) (*ResolveResult, error) {
	repoCachePath := r.getRepoCachePath(dep.RepositoryURL)
	specPath := filepath.Join(repoCachePath, dep.SpecPath)

	content, err := os.ReadFile(specPath)
	if err != nil {
		return nil, err
	}

	// Calculate hash
	hasher := sha256.New()
	hasher.Write(content)
	contentHash := hex.EncodeToString(hasher.Sum(nil))

	return &ResolveResult{
		Dependency:  &dep,
		Content:     content,
		ContentHash: contentHash,
		Size:        int64(len(content)),
		Source:      "cache",
	}, nil
}

// saveToCache saves content to cache (already saved by git clone)
func (r *Resolver) saveToCache(result *ResolveResult) error {
	// Content is already saved by git clone in the repo cache directory
	// We just need to ensure the cache directory exists
	if r.cacheDir == "" {
		return fmt.Errorf("cache directory not set")
	}

	return os.MkdirAll(r.cacheDir, 0755)
}

// GetCachePath returns the cache path for a dependency
// This can be used by LLMs to read cached specs
func (r *Resolver) GetCachePath(dep models.Dependency, commitHash string) string {
	repoCachePath := r.getRepoCachePath(dep.RepositoryURL)
	return filepath.Join(repoCachePath, dep.SpecPath)
}

// GetGlobalCacheDir returns the global cache directory
func GetGlobalCacheDir() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		if u, err := user.Current(); err == nil {
			homeDir = u.HomeDir
		}
	}
	return filepath.Join(homeDir, ".specledger", "cache")
}
