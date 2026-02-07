package templates

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadManifest loads the template manifest from the templates directory.
func LoadManifest(templatesDir string) (*TemplateManifest, error) {
	manifestPath := filepath.Join(templatesDir, "manifest.yaml")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	return ParseManifest(data)
}

// ParseManifest parses the manifest YAML data.
func ParseManifest(data []byte) (*TemplateManifest, error) {
	var manifest TemplateManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Validate manifest version
	if manifest.Version == "" {
		manifest.Version = "1.0" // Default version
	}

	// Validate templates
	for i, tmpl := range manifest.Templates {
		if tmpl.Name == "" {
			return nil, fmt.Errorf("template %d: name is required", i)
		}
		if tmpl.Framework == "" {
			return nil, fmt.Errorf("template %s: framework is required", tmpl.Name)
		}
		if tmpl.Path == "" {
			tmpl.Path = tmpl.Name // Default path to template name
		}
		if tmpl.Version == "" {
			tmpl.Version = "1.0.0" // Default version
		}
		manifest.Templates[i] = tmpl
	}

	return &manifest, nil
}
