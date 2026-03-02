package spec

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func GetFeatureDir(repoRoot, branch string) string {
	return filepath.Join(repoRoot, "specledger", branch)
}

func GetSpecFile(featureDir string) string {
	return filepath.Join(featureDir, "spec.md")
}

func GetPlanFile(featureDir string) string {
	return filepath.Join(featureDir, "plan.md")
}

func GetTasksFile(featureDir string) string {
	return filepath.Join(featureDir, "tasks.md")
}

func DiscoverDocs(featureDir string) ([]string, error) {
	info, err := os.Stat(featureDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("feature directory does not exist: %s", featureDir)
		}
		return nil, fmt.Errorf("failed to access feature directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("feature path is not a directory: %s", featureDir)
	}

	var docs []string

	err = filepath.WalkDir(featureDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			relPath, err := filepath.Rel(featureDir, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			baseName := filepath.Base(relPath)
			if baseName != "spec.md" && baseName != "plan.md" && baseName != "tasks.md" {
				docs = append(docs, relPath)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk feature directory: %w", err)
	}

	return docs, nil
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func EnsureDir(path string) error {
	if DirExists(path) {
		return nil
	}

	return os.MkdirAll(path, 0755)
}
