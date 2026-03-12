package scheduler

import (
	"fmt"

	"github.com/specledger/specledger/pkg/cli/spec"
)

// DetectionResult holds the outcome of scanning for an approved spec on the current branch.
type DetectionResult struct {
	Feature    string
	SpecDir    string
	Approved   bool
	RepoRoot   string
}

// DetectApprovedSpec checks whether the current branch has an approved SpecLedger spec.
// Returns nil result (no error) for non-feature branches or specs without approval.
func DetectApprovedSpec(projectRoot string) (*DetectionResult, error) {
	ctx, err := spec.DetectFeatureContext(projectRoot)
	if err != nil {
		// Detection failure (non-feature branch, no spec, etc.) is not an error for the hook —
		// it just means there's nothing to trigger.
		return nil, nil
	}

	branch := ctx.Branch
	if !isFeatureBranchName(branch) {
		return &DetectionResult{
			Feature:  branch,
			SpecDir:  ctx.FeatureDir,
			Approved: false,
			RepoRoot: ctx.RepoRoot,
		}, nil
	}

	status, err := spec.ReadStatus(ctx.FeatureDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec status for %s: %w", branch, err)
	}

	return &DetectionResult{
		Feature:  branch,
		SpecDir:  ctx.FeatureDir,
		Approved: status == "Approved",
		RepoRoot: ctx.RepoRoot,
	}, nil
}

// isFeatureBranchName checks if a branch name matches the NNN-feature-name pattern.
func isFeatureBranchName(name string) bool {
	if len(name) < 4 {
		return false
	}
	for i, ch := range name {
		if ch == '-' {
			return i >= 3
		}
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return false
}
