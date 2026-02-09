package framework

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/specledger/specledger/pkg/cli/metadata"
)

// DetectFramework detects which SDD framework a remote repository uses
func DetectFramework(repoURL string) (metadata.FrameworkChoice, error) {
	// Clone the repo to a temporary directory
	tempDir, err := os.MkdirTemp("", "specledger-detect-*")
	if err != nil {
		return metadata.FrameworkNone, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone the repository
	repoPath := filepath.Join(tempDir, "repo")
	cmd := exec.Command("git", "clone", "--depth=1", repoURL, repoPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return metadata.FrameworkNone, fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	// Check for Spec Kit indicators
	if hasSpecKitMarkers(repoPath) {
		return metadata.FrameworkSpecKit, nil
	}

	// Check for OpenSpec indicators
	if hasOpenSpecMarkers(repoPath) {
		return metadata.FrameworkOpenSpec, nil
	}

	// Default to none if no framework detected
	return metadata.FrameworkNone, nil
}

// hasSpecKitMarkers checks if a repository uses Spec Kit
func hasSpecKitMarkers(repoPath string) bool {
	// Check for .specify directory
	specifyDir := filepath.Join(repoPath, ".specify")
	if info, err := os.Stat(specifyDir); err == nil && info.IsDir() {
		return true
	}

	// Check for specify.yaml file
	specifyYaml := filepath.Join(repoPath, "specify.yaml")
	if _, err := os.Stat(specifyYaml); err == nil {
		return true
	}

	// Check for spec-kit-version file
	versionFile := filepath.Join(repoPath, "spec-kit-version")
	if _, err := os.Stat(versionFile); err == nil {
		return true
	}

	// Check for SPECKIT.md or similar documentation
	matches, _ := filepath.Glob(filepath.Join(repoPath, "*[Ss][Pp][Ee][Cc][Kk][Ii][Tt]*.md"))
	return len(matches) > 0
}

// hasOpenSpecMarkers checks if a repository uses OpenSpec
func hasOpenSpecMarkers(repoPath string) bool {
	// Check for .openspec directory
	openspecDir := filepath.Join(repoPath, ".openspec")
	if info, err := os.Stat(openspecDir); err == nil && info.IsDir() {
		return true
	}

	// Check for openspec.yaml file
	openspecYaml := filepath.Join(repoPath, "openspec.yaml")
	if _, err := os.Stat(openspecYaml); err == nil {
		return true
	}

	// Check for OPENSPEC.md or similar documentation
	matches, _ := filepath.Glob(filepath.Join(repoPath, "*[Oo][Pp][Ee][Nn][Ss][Pp][Ee][Cc]*.md"))
	return len(matches) > 0
}

// DetectFrameworkFromContent detects framework from raw content (without cloning)
func DetectFrameworkFromContent(content []byte) metadata.FrameworkChoice {
	// Look for framework-specific patterns
	if bytes.Contains(content, []byte(".specify")) ||
		bytes.Contains(content, []byte("specify.yaml")) ||
		bytes.Contains(content, []byte("spec-kit-version")) ||
		(bytes.Contains(content, []byte("Spec Kit")) && bytes.Contains(content, []byte("framework"))) {
		return metadata.FrameworkSpecKit
	}

	if bytes.Contains(content, []byte(".openspec")) ||
		bytes.Contains(content, []byte("openspec.yaml")) ||
		(bytes.Contains(content, []byte("OpenSpec")) && bytes.Contains(content, []byte("framework"))) {
		return metadata.FrameworkOpenSpec
	}

	return metadata.FrameworkNone
}

// GetFrameworkImportPath generates an import path for AI context
func GetFrameworkImportPath(dep metadata.Dependency) string {
	if dep.ImportPath != "" {
		return dep.ImportPath
	}

	if dep.Alias != "" {
		return fmt.Sprintf("@%s", dep.Alias)
	}

	// Generate from URL
	return fmt.Sprintf("@%s", extractRepoName(dep.URL))
}

// extractRepoName extracts a short name from a git URL
func extractRepoName(url string) string {
	// Remove git@ prefix if present
	url = strings.TrimPrefix(url, "git@")

	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Extract the last part of the path
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return url
}
