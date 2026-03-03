package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type FeatureContext struct {
	RepoRoot   string
	Branch     string
	FeatureDir string
	SpecFile   string
	PlanFile   string
	TasksFile  string
	HasGit     bool
}

type BranchMap struct {
	BranchToFeature map[string]string `json:"branchToFeature"`
}

func DetectFeatureContext(workDir string) (*FeatureContext, error) {
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

	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	var branch string
	if head.Name().IsBranch() {
		branch = head.Name().Short()
	} else {
		return nil, fmt.Errorf("detached HEAD state - please checkout a feature branch (got commit %s)", head.Hash().String()[:8])
	}

	featureBranch := branch
	if !isFeatureBranch(branch) {
		mappedFeature, err := checkBranchMap(repoRoot, branch)
		if err != nil {
			return nil, fmt.Errorf("not a feature branch: %q (expected pattern: ###-description). Create branch-map.json to map non-feature branches to feature contexts", branch)
		}
		if mappedFeature != "" {
			featureBranch = mappedFeature
		} else {
			return nil, fmt.Errorf("not a feature branch: %q (expected pattern: ###-description)", branch)
		}
	}

	featureDir := filepath.Join(repoRoot, "specledger", featureBranch)
	specFile := filepath.Join(featureDir, "spec.md")
	planFile := filepath.Join(featureDir, "plan.md")
	tasksFile := filepath.Join(featureDir, "tasks.md")

	return &FeatureContext{
		RepoRoot:   repoRoot,
		Branch:     branch,
		FeatureDir: featureDir,
		SpecFile:   specFile,
		PlanFile:   planFile,
		TasksFile:  tasksFile,
		HasGit:     true,
	}, nil
}

func checkBranchMap(repoRoot, branch string) (string, error) {
	branchMapPath := filepath.Join(repoRoot, ".specledger", "branch-map.json")

	data, err := os.ReadFile(branchMapPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read branch-map.json: %w", err)
	}

	var branchMap BranchMap
	if err := json.Unmarshal(data, &branchMap); err != nil {
		return "", fmt.Errorf("failed to parse branch-map.json: %w", err)
	}

	if branchMap.BranchToFeature == nil {
		return "", nil
	}

	return branchMap.BranchToFeature[branch], nil
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
