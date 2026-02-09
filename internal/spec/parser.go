package spec

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/specledger/specledger/pkg/models"
)

// Format version constants
const (
	ManifestVersion = "1.0.0"
)

// ParseManifest parses a spec.mod file
func ParseManifest(path string) (*Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest file: %w", err)
	}
	defer file.Close()

	manifest := &Manifest{
		Version:     ManifestVersion,
		Dependecies: make([]models.Dependency, 0),
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse dependency declaration
		dep, err := parseDependencyLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		manifest.Dependecies = append(manifest.Dependecies, *dep)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	manifest.ID = extractID(manifest.Dependecies)
	manifest.Path = path

	return manifest, nil
}

// parseDependencyLine parses a single dependency line
func parseDependencyLine(line string) (*models.Dependency, error) {
	// Format: require <repo-url> <version> <spec-path> [id <spec-id>]
	// Or: require <repo-url> <version> <spec-path> --alias <alias>

	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid dependency line, expected at least 3 parts: %s", line)
	}

	if parts[0] != "require" {
		return nil, fmt.Errorf("expected 'require' keyword, got: %s", parts[0])
	}

	repoURL := parts[1]
	version := parts[2]
	specPath := parts[3]

	dep := &models.Dependency{
		RepositoryURL: repoURL,
		Version:       version,
		SpecPath:      specPath,
	}

	// Check for optional alias
	if len(parts) > 4 && parts[4] == "--alias" {
		if len(parts) < 6 {
			return nil, fmt.Errorf("alias requires a value")
		}
		dep.Alias = parts[5]
	}

	return dep, nil
}

// extractID extracts a unique ID from the manifest dependencies
func extractID(deps []models.Dependency) string {
	// Try to find an existing ID
	for _, dep := range deps {
		if dep.Alias != "" && strings.HasPrefix(dep.Alias, "#") {
			return dep.Alias
		}
	}

	// Generate a default ID based on repo URLs
	if len(deps) == 0 {
		return "root"
	}

	// Simple hash of the first few repo URLs
	hash := ""
	for _, dep := range deps {
		hash += dep.RepositoryURL
	}

	// Return a truncated hash
	if len(hash) > 20 {
		hash = hash[:20]
	}

	return "spec-" + hash
}

// WriteManifest writes a manifest to a file
func WriteManifest(path string, manifest *Manifest) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create manifest file: %w", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "# SpecLedger Dependency Manifest v%s\n", manifest.Version)
	fmt.Fprintf(file, "# Generated at %s\n\n", manifest.UpdatedAt.Format(time.RFC3339))

	for _, dep := range manifest.Dependecies {
		line := fmt.Sprintf("require %s %s %s", dep.RepositoryURL, dep.Version, dep.SpecPath)
		if dep.Alias != "" {
			line += fmt.Sprintf(" --alias %s", dep.Alias)
		}
		fmt.Fprintln(file, line)
	}

	return nil
}

// Manifest represents the parsed spec.mod file
type Manifest struct {
	Version     string
	Dependecies []models.Dependency
	ID          string
	Path        string
	UpdatedAt   time.Time
}
