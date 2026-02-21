package revise

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var featureBranchRe = regexp.MustCompile(`^\d{3,}-`)

// repoURLRe extracts owner/name from GitHub remote URLs (both HTTPS and SSH forms).
var repoURLRe = regexp.MustCompile(`github\.com[:/]([^/]+)/([^/\.]+?)(?:\.git)?$`)

// openRepo opens the git repository at repoPath, searching parent dirs for .git.
func openRepo(repoPath string) (*gogit.Repository, error) {
	repo, err := gogit.PlainOpenWithOptions(repoPath, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	return repo, nil
}

// GetRepoOwnerName parses the origin remote URL to extract the GitHub owner and repo name.
// Supports both HTTPS (https://github.com/owner/repo.git) and SSH (git@github.com:owner/repo.git) URLs.
func GetRepoOwnerName(repoPath string) (owner, name string, err error) {
	repo, err := openRepo(repoPath)
	if err != nil {
		return "", "", err
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return "", "", fmt.Errorf("no 'origin' remote found: %w", err)
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", "", fmt.Errorf("origin remote has no URLs")
	}

	m := repoURLRe.FindStringSubmatch(urls[0])
	if m == nil {
		return "", "", fmt.Errorf("cannot parse GitHub owner/repo from remote URL: %s", urls[0])
	}

	return m[1], m[2], nil
}

// BranchExists reports whether a local branch with the given name exists.
func BranchExists(repoPath, name string) (bool, error) {
	repo, err := openRepo(repoPath)
	if err != nil {
		return false, err
	}

	refName := plumbing.ReferenceName("refs/heads/" + name)
	_, err = repo.Reference(refName, true)
	if err == plumbing.ErrReferenceNotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to resolve branch ref: %w", err)
	}
	return true, nil
}

// GetCurrentBranch returns the short name of the current branch (e.g., "136-revise-comments").
func GetCurrentBranch(repoPath string) (string, error) {
	repo, err := openRepo(repoPath)
	if err != nil {
		return "", err
	}

	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}

	// Detached HEAD — return the commit hash
	return head.Hash().String()[:8], nil
}

// IsFeatureBranch reports whether name looks like a feature branch (e.g., "136-revise-comments").
func IsFeatureBranch(name string) bool {
	return featureBranchRe.MatchString(name)
}

// HasUncommittedChanges returns true if the working tree has uncommitted modifications.
func HasUncommittedChanges(repoPath string) (bool, error) {
	repo, err := openRepo(repoPath)
	if err != nil {
		return false, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	return !status.IsClean(), nil
}

// CheckoutBranch checks out a local branch by name using go-git.
// Use CheckoutRemoteTracking for branches that only exist on the remote.
func CheckoutBranch(repoPath, name string) error {
	repo, err := openRepo(repoPath)
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	refName := plumbing.ReferenceName("refs/heads/" + name)
	err = wt.Checkout(&gogit.CheckoutOptions{
		Branch: refName,
		Keep:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch %q: %w", name, err)
	}

	return nil
}

// StashChanges stashes all uncommitted changes.
// Uses exec (go-git does not implement stash; see go-git issue #606).
func StashChanges(repoPath string) error {
	// #nosec G204 — repoPath is from a controlled call site (working directory), not user input
	cmd := exec.Command("git", "stash", "push", "--include-untracked", "-m", "sl revise: auto-stash")
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git stash failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// StashPop re-applies the most recent stash.
// Uses exec (go-git does not implement stash; see go-git issue #606).
func StashPop(repoPath string) error {
	// #nosec G204 — repoPath is from a controlled call site
	cmd := exec.Command("git", "stash", "pop")
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git stash pop failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// CheckoutRemoteTracking checks out a remote-tracking branch, creating a local tracking branch.
// Uses exec because go-git requires a 4-step manual process for remote tracking checkout.
func CheckoutRemoteTracking(repoPath, name string) error {
	// #nosec G204 — repoPath is from a controlled call site, name is a validated branch name
	cmd := exec.Command("git", "checkout", "--track", "origin/"+name)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout --track failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// AddFiles stages the specified file paths in the repository.
func AddFiles(repoPath string, paths []string) error {
	repo, err := openRepo(repoPath)
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	for _, p := range paths {
		if _, err := wt.Add(p); err != nil {
			return fmt.Errorf("failed to stage %q: %w", p, err)
		}
	}

	return nil
}

// CommitChanges creates a commit with the staged changes.
// Author info is read from the repository's global git config.
func CommitChanges(repoPath, message string) error {
	repo, err := openRepo(repoPath)
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Read author info from global git config
	cfg, err := repo.ConfigScoped(gogitconfig.GlobalScope)
	if err != nil {
		return fmt.Errorf("failed to read git config: %w", err)
	}

	name := cfg.User.Name
	email := cfg.User.Email
	if name == "" {
		name = "SpecLedger"
	}
	if email == "" {
		email = "noreply@specledger.io"
	}

	_, err = wt.Commit(message, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  name,
			Email: email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// PushToRemote pushes the current branch to the origin remote.
// Uses exec for reliable credential helper and SSH agent support.
func PushToRemote(repoPath string) error {
	// #nosec G204 — repoPath is from a controlled call site
	cmd := exec.Command("git", "push")
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
