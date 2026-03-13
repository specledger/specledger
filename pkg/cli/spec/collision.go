package spec

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GenerateFeatureHash generates a random 6-character hex hash for feature identification.
// This eliminates collision risk when multiple people work on the same repo concurrently.
func GenerateFeatureHash() (string, error) {
	bytes := make([]byte, 3) // 3 bytes = 6 hex chars
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random hash: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CheckFeatureCollision checks if a feature ID (hash or number) already exists
// in local directories, local branches, or remote branches.
func CheckFeatureCollision(repoRoot, featureID string) error {
	if err := checkLocalFeatures(repoRoot, featureID); err != nil {
		return err
	}

	if err := checkLocalBranches(repoRoot, featureID); err != nil {
		return err
	}

	// Best-effort remote check — ignore errors (allow offline work)
	_ = checkRemoteBranches(repoRoot, featureID)

	return nil
}

func checkLocalFeatures(repoRoot, featureID string) error {
	specledgerDir := filepath.Join(repoRoot, "specledger")

	info, err := os.Stat(specledgerDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to access specledger directory: %w", err)
	}

	if !info.IsDir() {
		return nil
	}

	entries, err := os.ReadDir(specledgerDir)
	if err != nil {
		return fmt.Errorf("failed to read specledger directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		prefix := ParseFeatureID(entry.Name())
		if prefix != "" && prefix == featureID {
			return fmt.Errorf("feature ID %s already exists locally: %s", featureID, entry.Name())
		}
	}

	return nil
}

func checkLocalBranches(repoRoot, featureID string) error {
	repo, err := openRepo(repoRoot)
	if err != nil {
		return err
	}

	branches, err := repo.Branches()
	if err != nil {
		return fmt.Errorf("failed to list local branches: %w", err)
	}

	featurePattern := regexp.MustCompile(`^` + regexp.QuoteMeta(featureID) + `-`)

	found := false
	var existingBranch string

	err = branches.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()
		if featurePattern.MatchString(branchName) {
			found = true
			existingBranch = branchName
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to iterate branches: %w", err)
	}

	if found {
		return fmt.Errorf("feature ID %s already has a local branch: %s", featureID, existingBranch)
	}

	return nil
}

func checkRemoteBranches(repoRoot, featureID string) error {
	repo, err := openRepo(repoRoot)
	if err != nil {
		return err
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return nil
	}

	refs, err := remote.List(&gogit.ListOptions{})
	if err != nil {
		return nil
	}

	featurePattern := regexp.MustCompile(`^` + regexp.QuoteMeta(featureID) + `-`)

	for _, ref := range refs {
		if !ref.Name().IsBranch() {
			continue
		}

		branchName := ref.Name().Short()
		if featurePattern.MatchString(branchName) {
			return fmt.Errorf("feature ID %s already has a remote branch: origin/%s", featureID, branchName)
		}
	}

	return nil
}

// GenerateUniqueFeatureHash generates a hash and verifies it has no collisions.
// Retries up to 10 times (collision probability is ~1 in 16 million per attempt).
func GenerateUniqueFeatureHash(repoRoot string) (string, error) {
	for i := 0; i < 10; i++ {
		hash, err := GenerateFeatureHash()
		if err != nil {
			return "", err
		}
		if err := CheckFeatureCollision(repoRoot, hash); err == nil {
			return hash, nil
		}
	}
	return "", fmt.Errorf("could not generate unique feature hash after 10 attempts")
}

// ParseFeatureID extracts the feature ID prefix from a branch or directory name.
// Supports both legacy numeric format (e.g., "604" from "604-auto-spec-numbers")
// and hash format (e.g., "a3f2b1" from "a3f2b1-feature-name").
func ParseFeatureID(name string) string {
	parts := strings.SplitN(name, "-", 2)
	if len(parts) != 2 || parts[1] == "" {
		return ""
	}
	return parts[0]
}

// ParseFeatureNum is an alias for ParseFeatureID for backward compatibility.
func ParseFeatureNum(branchName string) string {
	return ParseFeatureID(branchName)
}
