package issues

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
)

// Context-related errors
var (
	ErrNotFeatureBranch = errors.New("not on a feature branch. Use --spec flag or checkout a ###-branch")
	ErrNoGitRepo        = errors.New("not in a git repository")
)

// specBranchPattern matches branch names like "010-my-feature" or "591-issue-tracking"
var specBranchPattern = regexp.MustCompile(`^(\d{3,}-[a-z0-9-]+)$`)

// ContextDetector detects the current spec context from git branch
type ContextDetector struct {
	repoPath string
}

// NewContextDetector creates a new context detector
func NewContextDetector(repoPath string) *ContextDetector {
	if repoPath == "" {
		repoPath = "."
	}
	return &ContextDetector{repoPath: repoPath}
}

// DetectSpecContext returns the current spec context from the git branch name
func (d *ContextDetector) DetectSpecContext() (string, error) {
	repo, err := git.PlainOpen(d.repoPath)
	if err != nil {
		return "", ErrNoGitRepo
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get branch name from reference
	branchName := ref.Name().Short()
	if branchName == "" {
		// Try to get from full ref name
		name := ref.Name().String()
		if strings.HasPrefix(name, "refs/heads/") {
			branchName = strings.TrimPrefix(name, "refs/heads/")
		}
	}

	// Parse spec context from branch name
	specContext, ok := ParseSpecFromBranch(branchName)
	if !ok {
		return "", ErrNotFeatureBranch
	}

	return specContext, nil
}

// ParseSpecFromBranch extracts the spec context from a branch name
// Returns the spec context and true if valid, empty string and false otherwise
func ParseSpecFromBranch(branchName string) (string, bool) {
	// Extract just the branch name part (remove refs/heads/ prefix if present)
	name := strings.TrimPrefix(branchName, "refs/heads/")

	// Check if it matches the spec pattern
	if specBranchPattern.MatchString(name) {
		return name, true
	}

	return "", false
}

// DetectSpecContextFromPath attempts to detect spec context from the current directory
func DetectSpecContextFromPath(path string) (string, error) {
	detector := NewContextDetector(path)
	return detector.DetectSpecContext()
}

// GetBranchName returns the current git branch name
func (d *ContextDetector) GetBranchName() (string, error) {
	repo, err := git.PlainOpen(d.repoPath)
	if err != nil {
		return "", ErrNoGitRepo
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return ref.Name().Short(), nil
}

// IsFeatureBranch checks if the current branch is a feature branch
func (d *ContextDetector) IsFeatureBranch() (bool, error) {
	_, err := d.DetectSpecContext()
	if err != nil {
		if errors.Is(err, ErrNotFeatureBranch) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ValidateSpecContext validates a spec context string
func ValidateSpecContext(specContext string) error {
	if specContext == "" {
		return errors.New("spec context cannot be empty")
	}
	if !specBranchPattern.MatchString(specContext) {
		return fmt.Errorf("invalid spec context format: %s (expected ###-name pattern)", specContext)
	}
	return nil
}

// FormatNotFeatureBranchError returns a formatted error message with helpful context
func FormatNotFeatureBranchError(currentBranch string) string {
	if currentBranch == "" {
		return "Not on a feature branch. Use --spec flag or checkout a ###-branch."
	}
	return fmt.Sprintf("Current branch '%s' is not a feature branch. Use --spec flag or checkout a ###-branch.", currentBranch)
}
