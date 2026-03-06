package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func CheckFeatureCollision(repoRoot, featureNum string) error {
	if err := checkLocalFeatures(repoRoot, featureNum); err != nil {
		return err
	}

	if err := checkLocalBranches(repoRoot, featureNum); err != nil {
		return err
	}

	if err := checkRemoteBranches(repoRoot, featureNum); err != nil {
		return nil
	}

	return nil
}

func checkLocalFeatures(repoRoot, featureNum string) error {
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

	featurePattern := regexp.MustCompile(`^(\d{3,})-`)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		matches := featurePattern.FindStringSubmatch(entry.Name())
		if len(matches) > 1 && matches[1] == featureNum {
			return fmt.Errorf("feature number %s already exists locally: %s", featureNum, entry.Name())
		}
	}

	return nil
}

func checkLocalBranches(repoRoot, featureNum string) error {
	repo, err := openRepo(repoRoot)
	if err != nil {
		return err
	}

	branches, err := repo.Branches()
	if err != nil {
		return fmt.Errorf("failed to list local branches: %w", err)
	}

	featurePattern := regexp.MustCompile(`^` + featureNum + `-`)

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
		return fmt.Errorf("feature number %s already has a local branch: %s", featureNum, existingBranch)
	}

	return nil
}

func checkRemoteBranches(repoRoot, featureNum string) error {
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

	featurePattern := regexp.MustCompile(`^` + featureNum + `-`)

	for _, ref := range refs {
		if !ref.Name().IsBranch() {
			continue
		}

		branchName := ref.Name().Short()
		if featurePattern.MatchString(branchName) {
			return fmt.Errorf("feature number %s already has a remote branch: origin/%s", featureNum, branchName)
		}
	}

	return nil
}

func GetNextFeatureNum(repoRoot string) (string, error) {
	specledgerDir := filepath.Join(repoRoot, "specledger")

	info, err := os.Stat(specledgerDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "001", nil
		}
		return "", fmt.Errorf("failed to access specledger directory: %w", err)
	}

	if !info.IsDir() {
		return "001", nil
	}

	entries, err := os.ReadDir(specledgerDir)
	if err != nil {
		return "", fmt.Errorf("failed to read specledger directory: %w", err)
	}

	maxNum := 0
	featurePattern := regexp.MustCompile(`^(\d{3,})-`)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		matches := featurePattern.FindStringSubmatch(entry.Name())
		if len(matches) > 1 {
			var num int
			_, _ = fmt.Sscanf(matches[1], "%d", &num)
			if num > maxNum {
				maxNum = num
			}
		}
	}

	nextNum := maxNum + 1
	return fmt.Sprintf("%03d", nextNum), nil
}

func ParseFeatureNum(branchName string) string {
	parts := strings.SplitN(branchName, "-", 2)
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}
