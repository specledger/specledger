package metadata

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// LegacyModFile is the old .mod filename
	LegacyModFile = "specledger/specledger.mod"
)

// ModFileMetadata represents parsed data from legacy .mod file
type ModFileMetadata struct {
	ProjectName string
	ShortCode   string
	CreatedAt   time.Time
}

// ParseModFile reads and extracts metadata from legacy .mod format
func ParseModFile(path string) (*ModFileMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file creation time
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	metadata := &ModFileMetadata{
		CreatedAt: fileInfo.ModTime(),
	}

	// Parse comments for project name and short code
	scanner := bufio.NewScanner(file)
	projectNamePattern := regexp.MustCompile(`^#\s*Project:\s*(.+)$`)
	shortCodePattern := regexp.MustCompile(`^#\s*Short Code:\s*(.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if matches := projectNamePattern.FindStringSubmatch(line); matches != nil {
			metadata.ProjectName = strings.TrimSpace(matches[1])
		}

		if matches := shortCodePattern.FindStringSubmatch(line); matches != nil {
			metadata.ShortCode = strings.TrimSpace(matches[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Validate required fields were found
	if metadata.ProjectName == "" {
		return nil, errors.New("could not find project name in .mod file")
	}
	if metadata.ShortCode == "" {
		return nil, errors.New("could not find short code in .mod file")
	}

	return metadata, nil
}

// MigrateModToYAML converts a legacy .mod file to new YAML format
func MigrateModToYAML(projectRoot string) (*ProjectMetadata, error) {
	modPath := filepath.Join(projectRoot, LegacyModFile)

	// Check if .mod file exists
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return nil, errors.New("no specledger.mod file found to migrate")
	}

	// Check if YAML already exists
	yamlPath := filepath.Join(projectRoot, DefaultMetadataFile)
	if _, err := os.Stat(yamlPath); err == nil {
		return nil, errors.New("specledger.yaml already exists, migration not needed")
	}

	// Parse the .mod file
	modData, err := ParseModFile(modPath)
	if err != nil {
		return nil, err
	}

	// Create new YAML metadata
	now := time.Now()
	metadata := &ProjectMetadata{
		Version: MetadataVersion,
		Project: ProjectInfo{
			Name:      modData.ProjectName,
			ShortCode: modData.ShortCode,
			Created:   modData.CreatedAt,
			Modified:  now,
			Version:   "0.1.0",
		},
		Playbook: PlaybookInfo{
			Name:    "specledger", // Default playbook for migrated projects
			Version: "unknown",    // Version unknown for migrated projects
		},
		Dependencies: []Dependency{},
	}

	// Validate
	if err := metadata.Validate(); err != nil {
		return nil, err
	}

	// Save to YAML (don't delete .mod file)
	if err := SaveToProject(metadata, projectRoot); err != nil {
		return nil, err
	}

	return metadata, nil
}

// HasLegacyModFile checks if a project has a legacy .mod file
func HasLegacyModFile(projectRoot string) bool {
	modPath := filepath.Join(projectRoot, LegacyModFile)
	_, err := os.Stat(modPath)
	return err == nil
}

// HasYAMLMetadata checks if a project has the new YAML metadata
func HasYAMLMetadata(projectRoot string) bool {
	yamlPath := filepath.Join(projectRoot, DefaultMetadataFile)
	_, err := os.Stat(yamlPath)
	return err == nil
}
