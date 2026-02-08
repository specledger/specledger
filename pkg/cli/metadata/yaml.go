package metadata

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// MetadataVersion is the current schema version
	MetadataVersion = "1.0.0"

	// DefaultMetadataFile is the default filename for project metadata
	DefaultMetadataFile = "specledger/specledger.yaml"
)

// Load reads and parses a specledger.yaml file
func Load(path string) (*ProjectMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata ProjectMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	if err := metadata.Validate(); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// Save writes ProjectMetadata to a YAML file
func Save(metadata *ProjectMetadata, path string) error {
	// Update modified timestamp
	metadata.Project.Modified = time.Now()

	// Validate before saving
	if err := metadata.Validate(); err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal to YAML
	data, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}

	// Write to file
	// #nosec G306 -- metadata files need to be readable, 0644 is appropriate
	return os.WriteFile(path, data, 0644)
}

// LoadFromProject loads metadata from a project directory
// It looks for specledger/specledger.yaml relative to projectRoot
func LoadFromProject(projectRoot string) (*ProjectMetadata, error) {
	metadataPath := filepath.Join(projectRoot, DefaultMetadataFile)
	return Load(metadataPath)
}

// SaveToProject saves metadata to a project directory
// It writes to specledger/specledger.yaml relative to projectRoot
func SaveToProject(metadata *ProjectMetadata, projectRoot string) error {
	metadataPath := filepath.Join(projectRoot, DefaultMetadataFile)
	return Save(metadata, metadataPath)
}

// NewProjectMetadata creates a new ProjectMetadata with default values
func NewProjectMetadata(name, shortCode string, playbookName string, playbookVersion string, playbookStructure []string) *ProjectMetadata {
	now := time.Now()
	metadata := &ProjectMetadata{
		Version: MetadataVersion,
		Project: ProjectInfo{
			Name:      name,
			ShortCode: shortCode,
			Created:   now,
			Modified:  now,
			Version:   "0.1.0",
		},
		Playbook: PlaybookInfo{
			Name:      playbookName,
			Version:   playbookVersion,
			AppliedAt: &now,
			Structure: playbookStructure,
		},
		TaskTracker: TaskTrackerInfo{
			Choice:    TaskTrackerBeads,
			EnabledAt: &now,
		},
		Dependencies: []Dependency{},
	}

	return metadata
}

// HasYAMLMetadata checks if a project has the new YAML metadata file
func HasYAMLMetadata(projectRoot string) bool {
	yamlPath := filepath.Join(projectRoot, DefaultMetadataFile)
	_, err := os.Stat(yamlPath)
	return err == nil
}
