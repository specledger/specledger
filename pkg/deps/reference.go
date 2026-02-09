package deps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveReference resolves a dependency reference to a local file path.
//
// The reference format is: <alias>:<artifact-name>
//
// Resolution formula:
// <project.artifact_path> + <dependency.alias> + "/" + <artifact-name>
//
// Example:
//   project.artifact_path: specledger/
//   dependency.alias: platform
//   artifact_name: api.md
//   Result: specledger/platform/api.md
//
// Parameters:
//   - projectArtifactPath: The artifact_path from the project's specledger.yaml
//   - depAlias: The alias of the dependency
//   - artifactName: The name of the artifact file (e.g., "api.md", "openapi.yaml")
//   - projectRoot: The root directory of the project (for absolute path resolution)
//
// Returns:
//   - The resolved file path (relative to project root)
//   - An error if resolution fails
func ResolveReference(projectArtifactPath, depAlias, artifactName, projectRoot string) (string, error) {
	// Validate inputs
	if projectArtifactPath == "" {
		return "", errors.New("project artifact_path cannot be empty")
	}
	if depAlias == "" {
		return "", errors.New("dependency alias cannot be empty")
	}
	if artifactName == "" {
		return "", errors.New("artifact name cannot be empty")
	}

	// Clean and normalize the artifact path
	projectArtifactPath = strings.TrimSuffix(projectArtifactPath, "/")
	if !strings.HasSuffix(projectArtifactPath, "/") {
		projectArtifactPath += "/"
	}

	// Build the resolved path
	// Format: project.artifact_path + dependency.alias + "/" + artifact_name
	resolvedPath := projectArtifactPath + depAlias + "/" + artifactName

	// If projectRoot is provided, check if the resolved file exists
	if projectRoot != "" {
		fullPath := filepath.Join(projectRoot, resolvedPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return "", fmt.Errorf("artifact not found: %s (resolved from %s)", resolvedPath, depAlias+":"+artifactName)
		}
	}

	return resolvedPath, nil
}

// ResolveReferenceWithCache resolves a dependency reference, checking both the project's
// artifact directory and the dependency's cache directory.
//
// This is useful when artifacts from dependencies need to be copied or symlinked
// into the project's artifact directory.
//
// Parameters:
//   - projectArtifactPath: The artifact_path from the project's specledger.yaml
//   - depAlias: The alias of the dependency
//   - depArtifactPath: The artifact_path from the dependency's specledger.yaml
//   - artifactName: The name of the artifact file
//   - projectRoot: The root directory of the project
//   - cachePath: The cache directory where the dependency is cloned
//
// Returns:
//   - The project-local resolved path
//   - The cache-resolved path (in the dependency's clone)
//   - An error if resolution fails
func ResolveReferenceWithCache(projectArtifactPath, depAlias, depArtifactPath, artifactName, projectRoot, cachePath string) (string, string, error) {
	// Resolve project-local path
	projectPath, err := ResolveReference(projectArtifactPath, depAlias, artifactName, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve project path: %w", err)
	}

	// Resolve cache path (where the artifact exists in the cloned dependency)
	depArtifactPath = strings.TrimSuffix(depArtifactPath, "/")
	if !strings.HasSuffix(depArtifactPath, "/") {
		depArtifactPath += "/"
	}
	cacheResolvedPath := filepath.Join(cachePath, depArtifactPath, artifactName)

	// Check if cache path exists
	if _, err := os.Stat(cacheResolvedPath); os.IsNotExist(err) {
		return projectPath, "", fmt.Errorf("artifact not found in dependency cache: %s", cacheResolvedPath)
	}

	return projectPath, cacheResolvedPath, nil
}
