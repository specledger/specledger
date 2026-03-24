package metadata

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindProjectRoot walks from the current working directory upward
// to find the nearest directory containing specledger/specledger.yaml.
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return FindProjectRootFrom(dir)
}

// FindProjectRootFrom walks from the given directory upward
// to find the nearest directory containing specledger/specledger.yaml.
func FindProjectRootFrom(dir string) (string, error) {
	if HasYAMLMetadata(dir) {
		return dir, nil
	}

	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not in a SpecLedger project (no %s found)", DefaultMetadataFile)
		}
		dir = parent

		if HasYAMLMetadata(dir) {
			return dir, nil
		}
	}
}
