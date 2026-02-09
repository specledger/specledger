// Package deps provides Git operations using go-git/v5 for dependency management.
package deps

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// CloneOptions contains options for cloning a dependency repository.
type CloneOptions struct {
	URL       string // Repository URL
	Branch    string // Branch to clone (default "main")
	TargetDir string // Directory to clone to
	Shallow   bool   // Whether to do a shallow clone
}

// Clone clones a Git repository using go-git/v5.
// Returns the cloned repository and the resolved commit SHA.
func Clone(opts CloneOptions) (*git.Repository, string, error) {
	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(opts.TargetDir), 0755); err != nil {
		return nil, "", fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Determine branch
	branch := opts.Branch
	if branch == "" {
		branch = "main"
	}

	// Clone options
	cloneOpts := &git.CloneOptions{
		URL:          opts.URL,
		Progress:     nil,
		Tags:         git.NoTags,
		NoCheckout:   false,
		SingleBranch: true,
	}

	// Set shallow clone option
	if opts.Shallow {
		cloneOpts.Depth = 1
	}

	// Set reference name for branch
	if branch != "" && branch != "main" {
		refName := plumbing.ReferenceName("refs/heads/" + branch)
		cloneOpts.ReferenceName = refName
	} else {
		cloneOpts.ReferenceName = plumbing.ReferenceName("refs/heads/main")
	}

	// Set auth based on URL format
	auth, err := getAuthForURL(opts.URL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to determine auth method: %w", err)
	}
	if auth != nil {
		cloneOpts.Auth = auth
	}

	// Clone the repository
	repo, err := git.PlainClone(opts.TargetDir, false, cloneOpts)
	if err != nil {
		if err == git.ErrRepositoryAlreadyExists {
			// Repository already exists, open it
			repo, err = git.PlainOpen(opts.TargetDir)
			if err != nil {
				return nil, "", fmt.Errorf("failed to open existing repository: %w", err)
			}
		} else {
			return nil, "", fmt.Errorf("failed to clone repository: %w", err)
		}
	}

	// Get the HEAD commit
	head, err := repo.Head()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	commitSHA := head.Hash().String()

	return repo, commitSHA, nil
}

// OpenRepository opens an existing Git repository.
func OpenRepository(path string) (*git.Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	return repo, nil
}

// Fetch fetches the latest changes from a repository.
func Fetch(repo *git.Repository, branch string) error {
	// Get the remote
	remotes, err := repo.Remotes()
	if err != nil || len(remotes) == 0 {
		return fmt.Errorf("no remotes found")
	}

	// Use the first remote (usually "origin")
	remote := remotes[0]

	// Fetch options
	fetchOpts := &git.FetchOptions{
		RemoteURL: remote.Config().URLs[0],
		Progress:  nil,
		Tags:      git.NoTags,
	}

	// Set auth based on URL format
	auth, err := getAuthForURL(remote.Config().URLs[0])
	if err != nil {
		return fmt.Errorf("failed to determine auth method: %w", err)
	}
	if auth != nil {
		fetchOpts.Auth = auth
	}

	// Fetch from remote
	if err := remote.Fetch(fetchOpts); err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	return nil
}

// Checkout checks out a specific commit or branch in a repository.
func Checkout(repo *git.Repository, ref string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Try to resolve as a commit hash first
	hash := plumbing.NewHash(ref)
	if !hash.IsZero() {
		// Checkout by hash
		if err := worktree.Checkout(&git.CheckoutOptions{
			Hash:  hash,
			Force: true,
		}); err == nil {
			return nil
		}
	}

	// Try to checkout by branch reference
	refName := plumbing.ReferenceName("refs/heads/" + ref)
	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: refName,
		Force:  true,
	}); err != nil {
		return fmt.Errorf("failed to checkout %s: %w", ref, err)
	}

	return nil
}

// Pull pulls the latest changes from a repository's remote branch.
func Pull(repo *git.Repository, branch string) (string, error) {
	// Fetch latest changes
	if err := Fetch(repo, branch); err != nil {
		return "", err
	}

	// Determine branch
	if branch == "" {
		branch = "main"
	}

	// Get the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get the remote branch reference
	remoteRefName := plumbing.ReferenceName("refs/remotes/origin/" + branch)
	remoteHash, err := repo.ResolveRevision(plumbing.Revision(remoteRefName))
	if err != nil {
		return "", fmt.Errorf("failed to resolve remote branch: %w", err)
	}

	// Checkout the remote branch
	if err := worktree.Checkout(&git.CheckoutOptions{
		Hash:  *remoteHash,
		Force: true,
	}); err != nil {
		return "", fmt.Errorf("failed to checkout remote branch: %w", err)
	}

	return remoteHash.String(), nil
}

// ResolveHead resolves the current HEAD commit of a repository.
func ResolveHead(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return head.Hash().String(), nil
}

// ResolveRemoteCommit resolves the commit SHA of a remote branch.
func ResolveRemoteCommit(repo *git.Repository, branch string) (string, error) {
	// Determine branch
	if branch == "" {
		branch = "main"
	}

	// Fetch latest changes
	if err := Fetch(repo, branch); err != nil {
		return "", err
	}

	// Get the remote branch reference
	remoteRefName := plumbing.ReferenceName("refs/remotes/origin/" + branch)
	hash, err := repo.ResolveRevision(plumbing.Revision(remoteRefName))
	if err != nil {
		return "", fmt.Errorf("failed to resolve remote branch: %w", err)
	}

	return hash.String(), nil
}

// Log returns commit log between two revisions.
// limit specifies the maximum number of commits to return (0 for unlimited).
func Log(repo *git.Repository, from, to string, limit int) (string, error) {
	// Parse the commit hashes
	toHash := plumbing.NewHash(to)

	// Get the commit iterator
	commitIter, err := repo.Log(&git.LogOptions{
		From:  toHash,
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get commit log: %w", err)
	}
	defer commitIter.Close()

	// Collect commits
	var commits []string
	count := 0

	// Iterate through commits
	for {
		commit, err := commitIter.Next()
		if err != nil {
			break // No more commits
		}

		// Stop if we've reached the "from" commit
		if commit.Hash.String() == from {
			break
		}

		// Add commit to list
		commits = append(commits, fmt.Sprintf("%s %s", commit.Hash.String()[:8], commit.Message))
		count++

		// Check limit
		if limit > 0 && count >= limit {
			break
		}
	}

	// If we didn't find any commits
	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found")
	}

	// Join commits with newlines
	result := ""
	for i, commit := range commits {
		if i > 0 {
			result += "\n"
		}
		result += commit
	}

	return result, nil
}

// getAuthForURL determines the appropriate authentication method for a Git URL.
// Returns nil for public repositories, SSH auth for git@ URLs.
func getAuthForURL(url string) (transport.AuthMethod, error) {
	// For SSH URLs (git@github.com:org/repo), try to use SSH auth
	if len(url) > 4 && url[:4] == "git@" {
		// Try to get the current user's home directory
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			if u, err := user.Current(); err == nil {
				homeDir = u.HomeDir
			}
		}

		// Try common SSH key locations
		sshKeys := []string{
			filepath.Join(homeDir, ".ssh", "id_rsa"),
			filepath.Join(homeDir, ".ssh", "id_ed25519"),
			filepath.Join(homeDir, ".ssh", "id_ecdsa"),
		}

		for _, keyPath := range sshKeys {
			if _, err := os.Stat(keyPath); err == nil {
				// Key file exists, try to use it
				auth, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
				if err == nil {
					return auth, nil
				}
			}
		}

		// No SSH key found, return nil
		// The clone might still work if git-credential helper is configured
		return nil, nil
	}

	// For HTTPS URLs, no auth needed for public repos
	// Users can set up git credentials for private repos
	return nil, nil
}
