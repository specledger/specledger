package spec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"specledger/pkg/models"
)

// ResolveResult represents the result of a dependency resolution
type ResolveResult struct {
	Dependency *models.Dependency
	CommitHash string
	Content    []byte
	ContentHash string
	Size       int64
	Source     string
}

// Resolver resolves dependencies by fetching specs from Git repositories
type Resolver struct {
	cacheDir string
}

// NewResolver creates a new resolver
func NewResolver(cacheDir string) *Resolver {
	return &Resolver{
		cacheDir: cacheDir,
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
		if err == nil {
			return cached, nil
		}
	}

	// Fetch from Git
	content, commitHash, err := r.fetchFromGit(ctx, dep)
	if err != nil {
		return nil, err
	}

	// Calculate hash
	contentHash := fmt.Sprintf("%x", hash(content))

	// Create result
	result := &ResolveResult{
		Dependency:    &dep,
		CommitHash:    commitHash,
		Content:       content,
		ContentHash:   contentHash,
		Size:          int64(len(content)),
		Source:        "remote",
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
	// Parse repo URL to get owner, repo, and branch
	_, _, err := parseGitURL(dep.RepositoryURL)
	if err != nil {
		return nil, "", err
	}

	// Clone or fetch the repository
	repo, err := git.PlainCloneContext(ctx, r.cacheDir+"/"+dep.Alias, false, &git.CloneOptions{
		URL: dep.RepositoryURL,
		Depth: 1, // Shallow clone for performance
	})

	if err != nil {
		return nil, "", fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get the branch/commit
	var commitHash plumbing.Hash

	// Handle special cases
	if strings.HasPrefix(dep.Version, "#") {
		// It's a branch or tag name
		branch := strings.TrimPrefix(dep.Version, "#")
		rev, err := repo.ResolveRevision(plumbing.Revision(branch))
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolve branch %s: %w", branch, err)
		}
		commitHash = *rev
	} else {
		// It's a commit hash
		rev, err := repo.ResolveRevision(plumbing.Revision(dep.Version))
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolve commit %s: %w", dep.Version, err)
		}
		commitHash = *rev
	}

	// Read the spec file
	specPath := filepath.Join(r.cacheDir+"/"+dep.Alias, dep.SpecPath)
	content, err := os.ReadFile(specPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read spec file: %w", err)
	}

	return content, commitHash.String(), nil
}

// parseGitURL parses a Git URL into its components
func parseGitURL(url string) (string, string, error) {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repository URL: %s", url)
	}

	// Get repo name
	repoName := parts[len(parts)-1]
	if strings.Contains(repoName, ".git") {
		repoName = strings.TrimSuffix(repoName, ".git")
	}

	// Get owner (second to last part for GitHub, or path)
	owner := parts[len(parts)-2]
	if len(parts) > 2 && parts[0] == "github.com" {
		owner = parts[1]
	}

	return owner, repoName, nil
}

// hash calculates a hash of the content
func hash(data []byte) string {
	// Simple hash for now
	// In production, use a proper hash algorithm like SHA-256
	return fmt.Sprintf("%x", len(data))
}

// getCachedContent gets content from cache
func (r *Resolver) getCachedContent(dep models.Dependency) (*ResolveResult, error) {
	cachePath := filepath.Join(r.cacheDir, dep.Alias, dep.SpecPath)
	content, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	// Verify content hash (placeholder)
	return &ResolveResult{
		Dependency:    &dep,
		Content:       content,
		Source:        "cache",
	}, nil
}

// saveToCache saves content to cache
func (r *Resolver) saveToCache(result *ResolveResult) error {
	if r.cacheDir == "" {
		return fmt.Errorf("cache directory not set")
	}

	cachePath := filepath.Join(r.cacheDir, result.Dependency.Alias, result.Dependency.SpecPath)

	// Create directories if needed
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	return os.WriteFile(cachePath, result.Content, 0644)
}
