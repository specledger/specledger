package spec

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/specledger/specledger/pkg/models"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateManifest validates a manifest
func ValidateManifest(manifest *Manifest) []error {
	var errors []error

	// Check version
	if manifest.Version == "" {
		errors = append(errors, &ValidationError{
			Field:   "version",
			Message: "version cannot be empty",
		})
	} else if !isValidVersion(manifest.Version) {
		errors = append(errors, &ValidationError{
			Field:   "version",
			Message: fmt.Sprintf("invalid version format: %s", manifest.Version),
		})
	}

	// Check dependencies
	for i, dep := range manifest.Dependecies {
		if depErr := validateDependency(&dep, i); depErr != nil {
			errors = append(errors, depErr)
		}
	}

	// Check for duplicate dependencies
	duplicates := findDuplicateDependencies(manifest.Dependecies)
	if len(duplicates) > 0 {
		errors = append(errors, &ValidationError{
			Field:   "dependencies",
			Message: fmt.Sprintf("duplicate dependencies found: %v", duplicates),
		})
	}

	return errors
}

// validateDependency validates a single dependency
func validateDependency(dep *models.Dependency, index int) *ValidationError {
	if dep.RepositoryURL == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("dependencies[%d].repository_url", index),
			Message: "repository URL cannot be empty",
		}
	}

	if !isValidURL(dep.RepositoryURL) {
		return &ValidationError{
			Field:   fmt.Sprintf("dependencies[%d].repository_url", index),
			Message: "invalid repository URL",
		}
	}

	if dep.Version == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("dependencies[%d].version", index),
			Message: "version cannot be empty",
		}
	}

	if !isValidVersion(dep.Version) {
		return &ValidationError{
			Field:   fmt.Sprintf("dependencies[%d].version", index),
			Message: fmt.Sprintf("invalid version format: %s", dep.Version),
		}
	}

	if dep.SpecPath == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("dependencies[%d].spec_path", index),
			Message: "spec path cannot be empty",
		}
	}

	if !isValidSpecPath(dep.SpecPath) {
		return &ValidationError{
			Field:   fmt.Sprintf("dependencies[%d].spec_path", index),
			Message: "invalid spec path",
		}
	}

	if dep.Alias != "" && !isValidAlias(dep.Alias) {
		return &ValidationError{
			Field:   fmt.Sprintf("dependencies[%d].alias", index),
			Message: "invalid alias format",
		}
	}

	return nil
}

// isValidURL checks if a string is a valid Git URL
func isValidURL(s string) bool {
	_, err := url.Parse(s)
	return err == nil
}

// isValidVersion checks if a version string is valid
func isValidVersion(v string) bool {
	// Allow: branch name, tag name, commit hash (40 chars), semver
	// Branch: alphanumeric, hyphens, underscores
	if len(v) > 0 && v[0] == '#' {
		return true // branch or tag
	}

	// Check if it's a semver pattern
	semverRegex := regexp.MustCompile(`^v?(\d+)(\.\d+)?(\.\d+)?$`)
	if semverRegex.MatchString(v) {
		return true
	}

	// Check if it's a commit hash (40 hex characters)
	hexRegex := regexp.MustCompile(`^[a-fA-F0-9]{40}$`)
	if hexRegex.MatchString(v) {
		return true
	}

	// Check if it's a simple identifier (branch name)
	if len(v) <= 50 {
		validChars := regexp.MustCompile(`^[a-zA-Z0-9_.\-]+$`)
		return validChars.MatchString(v)
	}

	return false
}

// isValidSpecPath checks if a spec path is valid
func isValidSpecPath(path string) bool {
	// Must be relative, not contain "..", must end with .md
	if strings.Contains(path, "..") {
		return false
	}

	if !strings.HasSuffix(path, ".md") {
		return false
	}

	// Must not be empty after extension
	base := strings.TrimSuffix(path, ".md")
	return len(base) > 0
}

// isValidAlias checks if an alias is valid
func isValidAlias(alias string) bool {
	// Can start with # or . for special purposes
	alias = strings.TrimPrefix(alias, "#")
	alias = strings.TrimPrefix(alias, ".")

	// Alphanumeric, hyphens, underscores, periods only, 1-50 chars
	return len(alias) > 0 && len(alias) <= 50 &&
		regexp.MustCompile(`^[a-zA-Z0-9_.\-]+$`).MatchString(alias)
}

// findDuplicateDependencies finds duplicate dependencies
func findDuplicateDependencies(deps []models.Dependency) []string {
	found := make(map[string]bool)
	duplicates := []string{}

	for _, dep := range deps {
		key := dep.RepositoryURL + ":" + dep.SpecPath
		if found[key] {
			duplicates = append(duplicates, key)
		} else {
			found[key] = true
		}
	}

	return duplicates
}
