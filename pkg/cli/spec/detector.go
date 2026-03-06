package spec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/specledger/specledger/pkg/cli/metadata"
)

// DetectionOptions configures feature context detection behavior
type DetectionOptions struct {
	// SpecOverride bypasses all detection steps and uses this spec name directly (FR-013)
	SpecOverride string
	// Interactive enables the interactive prompt fallback (step 4)
	Interactive bool
}

type FeatureContext struct {
	RepoRoot   string
	Branch     string
	FeatureDir string
	SpecFile   string
	PlanFile   string
	TasksFile  string
	HasGit     bool
}

// DetectFeatureContext detects the current feature context using the 4-step fallback chain (FR-011)
// Steps: 1) env var/regex match → 2) yaml alias → 3) git heuristic → 4) error with available features
func DetectFeatureContext(workDir string) (*FeatureContext, error) {
	return DetectFeatureContextWithOptions(workDir, DetectionOptions{})
}

// DetectFeatureContextWithOptions detects feature context with configurable behavior
func DetectFeatureContextWithOptions(workDir string, opts DetectionOptions) (*FeatureContext, error) {
	repo, err := git.PlainOpenWithOptions(workDir, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	repoRoot := wt.Filesystem.Root()

	// Step 0: --spec flag override (FR-013) - highest priority
	if opts.SpecOverride != "" {
		return buildFeatureContext(repoRoot, opts.SpecOverride)
	}

	// Step 1: SPECIFY_FEATURE env var or regex match (FR-014)
	featureBranch := os.Getenv("SPECIFY_FEATURE")
	if featureBranch == "" {
		head, err := repo.Head()
		if err != nil {
			return nil, fmt.Errorf("failed to get HEAD: %w", err)
		}

		if head.Name().IsBranch() {
			currentBranch := head.Name().Short()
			// Check if branch matches NNN-xxx pattern (regex step)
			if isFeatureBranch(currentBranch) {
				featureBranch = currentBranch
			} else {
				// Step 2: YAML alias lookup (FR-012)
				alias, err := lookupBranchAlias(repoRoot, currentBranch)
				if err == nil && alias != "" {
					featureBranch = alias
				} else {
					// Step 3: Git heuristic - analyze touched files
					heuristicBranch, err := detectFeatureFromGitHistory(repo, repoRoot)
					if err == nil && heuristicBranch != "" {
						featureBranch = heuristicBranch
					} else {
						// Step 4: No detection possible - provide helpful error
						return nil, buildDetectionError(currentBranch, repoRoot)
					}
				}
			}
		} else {
			return nil, fmt.Errorf("detached HEAD state - please checkout a feature branch or set SPECIFY_FEATURE env var (got commit %s)", head.Hash().String()[:8])
		}
	}

	return buildFeatureContext(repoRoot, featureBranch)
}

// buildFeatureContext creates the FeatureContext for a given spec name
func buildFeatureContext(repoRoot, specName string) (*FeatureContext, error) {
	featureDir := filepath.Join(repoRoot, "specledger", specName)

	if !DirExists(featureDir) {
		availableFeatures := ListAvailableFeatures(repoRoot)
		if len(availableFeatures) > 0 {
			return nil, fmt.Errorf("feature directory not found: %s\n\nAvailable features:\n  - %s\n\nSet SPECIFY_FEATURE=<feature-name> or checkout matching branch", specName, strings.Join(availableFeatures, "\n  - "))
		}
		return nil, fmt.Errorf("feature directory not found: %s\n\nNo features available. Create one with: sl spec create", featureDir)
	}

	specFile := filepath.Join(featureDir, "spec.md")
	planFile := filepath.Join(featureDir, "plan.md")
	tasksFile := filepath.Join(featureDir, "tasks.md")

	return &FeatureContext{
		RepoRoot:   repoRoot,
		Branch:     specName,
		FeatureDir: featureDir,
		SpecFile:   specFile,
		PlanFile:   planFile,
		TasksFile:  tasksFile,
		HasGit:     true,
	}, nil
}

// lookupBranchAlias checks specledger.yaml for branch aliases (FR-012)
func lookupBranchAlias(repoRoot, branchName string) (string, error) {
	meta, err := metadata.LoadFromProject(repoRoot)
	if err != nil {
		return "", err
	}

	if meta.BranchAliases == nil {
		return "", fmt.Errorf("no branch aliases configured")
	}

	if alias, ok := meta.BranchAliases[branchName]; ok {
		return alias, nil
	}

	return "", fmt.Errorf("branch %q not found in aliases", branchName)
}

// detectFeatureFromGitHistory analyzes recent commits to find touched specledger dirs (step 3)
func detectFeatureFromGitHistory(repo *git.Repository, repoRoot string) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", err
	}

	// Get recent commits
	commitIter, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return "", err
	}
	defer commitIter.Close()

	// Track specledger dirs touched
	dirCounts := make(map[string]int)
	maxCommits := 20
	commitCount := 0

	for commitCount < maxCommits {
		commit, err := commitIter.Next()
		if err != nil {
			break
		}
		commitCount++

		// Get the tree for this commit
		tree, err := commit.Tree()
		if err != nil {
			continue
		}

		// Check for changes in specledger/ directories
		_ = tree.Files().ForEach(func(f *object.File) error {
			if strings.HasPrefix(f.Name, "specledger/") {
				// Extract the feature directory name
				parts := strings.Split(strings.TrimPrefix(f.Name, "specledger/"), "/")
				if len(parts) > 0 && strings.Contains(parts[0], "-") {
					dirCounts[parts[0]]++
				}
			}
			return nil
		})
	}

	// If exactly one feature dir was touched, use it
	if len(dirCounts) == 1 {
		for dir := range dirCounts {
			// Verify it's a valid feature directory
			if DirExists(filepath.Join(repoRoot, "specledger", dir)) {
				return dir, nil
			}
		}
	}

	return "", fmt.Errorf("could not determine feature from git history")
}

// buildDetectionError creates a helpful error message for detection failure
func buildDetectionError(currentBranch string, repoRoot string) error {
	availableFeatures := ListAvailableFeatures(repoRoot)

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("Could not detect feature for branch %q\n\n", currentBranch))
	msg.WriteString("Detection steps tried:\n")
	msg.WriteString("  1. Branch name pattern (NNN-xxx) - did not match\n")
	msg.WriteString("  2. YAML alias lookup - not configured\n")
	msg.WriteString("  3. Git heuristic - inconclusive\n\n")

	if len(availableFeatures) > 0 {
		msg.WriteString("Available features:\n")
		for _, f := range availableFeatures {
			msg.WriteString(fmt.Sprintf("  - %s\n", f))
		}
		msg.WriteString("\nTo fix, either:\n")
		msg.WriteString("  - Add an alias to specledger/specledger.yaml:\n")
		msg.WriteString(fmt.Sprintf("    branch_aliases:\n      %s: <feature-name>\n", currentBranch))
		msg.WriteString("  - Use --spec flag: sl <command> --spec <feature-name>")
	} else {
		msg.WriteString("No features available. Create one with: sl spec create")
	}

	return errors.New(msg.String())
}

func ListAvailableFeatures(repoRoot string) []string {
	specledgerDir := filepath.Join(repoRoot, "specledger")

	entries, err := os.ReadDir(specledgerDir)
	if err != nil {
		return nil
	}

	var features []string
	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(entry.Name(), "-") {
			features = append(features, entry.Name())
		}
	}

	return features
}

func isFeatureBranch(name string) bool {
	if len(name) < 4 {
		return false
	}

	for i, ch := range name {
		if ch == '-' {
			return i >= 3 && isAllDigits(name[:i])
		}
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return false
}

func isAllDigits(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return len(s) > 0
}

func GetFeatureNum(branch string) string {
	parts := strings.SplitN(branch, "-", 2)
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

func openRepo(repoPath string) (*git.Repository, error) {
	repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	return repo, nil
}

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

	return "", fmt.Errorf("detached HEAD state (commit %s)", head.Hash().String()[:8])
}

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
