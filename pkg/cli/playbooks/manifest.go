package playbooks

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadManifest loads the playbook manifest from the embedded templates directory.
func LoadManifest(templatesDir string) (*PlaybookManifest, error) {
	manifestPath := filepath.Join(templatesDir, "manifest.yaml")

	data, err := ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	return ParseManifest(data)
}

// ParseManifest parses the manifest YAML data.
func ParseManifest(data []byte) (*PlaybookManifest, error) {
	var manifest PlaybookManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Validate manifest version
	if manifest.Version == "" {
		manifest.Version = "1.0" // Default version
	}

	// Validate playbooks
	for i, pb := range manifest.Playbooks {
		if pb.Name == "" {
			return nil, fmt.Errorf("playbook %d: name is required", i)
		}
		if pb.Path == "" {
			pb.Path = pb.Name // Default path to playbook name
		}
		if pb.Version == "" {
			pb.Version = "1.0.0" // Default version
		}
		// Framework is optional - default to empty (works for all frameworks)
		if pb.Framework == "" {
			pb.Framework = "all"
		}
		manifest.Playbooks[i] = pb
	}

	// Validate templates (new in v1.1.0)
	for i, tmpl := range manifest.Templates {
		if err := tmpl.Validate(); err != nil {
			return nil, fmt.Errorf("template %d (%s): %w", i, tmpl.ID, err)
		}
		manifest.Templates[i] = tmpl
	}

	return &manifest, nil
}
