package templates

import (
	"fmt"
	"os"
	"path/filepath"
)

// EmbeddedSource implements TemplateSource for templates stored in the local templates directory.
type EmbeddedSource struct {
	templatesDir string
	manifest     *TemplateManifest
}

// NewEmbeddedSource creates a new EmbeddedSource.
func NewEmbeddedSource() (*EmbeddedSource, error) {
	templatesDir, err := TemplatesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to find templates directory: %w", err)
	}

	manifest, err := LoadManifest(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	return &EmbeddedSource{
		templatesDir: templatesDir,
		manifest:     manifest,
	}, nil
}

// List returns all available templates from the embedded source.
func (s *EmbeddedSource) List() ([]Template, error) {
	if s.manifest == nil {
		return nil, fmt.Errorf("manifest not loaded")
	}
	return s.manifest.Templates, nil
}

// Copy copies the specified template to the destination directory.
func (s *EmbeddedSource) Copy(name string, destDir string, opts CopyOptions) (*CopyResult, error) {
	template, err := s.getTemplate(name)
	if err != nil {
		return nil, err
	}

	return CopyTemplates(s.templatesDir, destDir, *template, opts)
}

// Exists checks if a template with the given name exists.
func (s *EmbeddedSource) Exists(name string) bool {
	_, err := s.getTemplate(name)
	return err == nil
}

// getTemplate retrieves a template by name.
func (s *EmbeddedSource) getTemplate(name string) (*Template, error) {
	if s.manifest == nil {
		return nil, fmt.Errorf("manifest not loaded")
	}

	for _, tmpl := range s.manifest.Templates {
		if tmpl.Name == name {
			return &tmpl, nil
		}
	}

	return nil, fmt.Errorf("template not found: %s", name)
}

// GetTemplateByFramework returns the first template matching the given framework.
func (s *EmbeddedSource) GetTemplateByFramework(framework string) (*Template, error) {
	if s.manifest == nil {
		return nil, fmt.Errorf("manifest not loaded")
	}

	for _, tmpl := range s.manifest.Templates {
		if tmpl.Framework == framework || tmpl.Framework == "both" {
			return &tmpl, nil
		}
	}

	return nil, fmt.Errorf("no template found for framework: %s", framework)
}

// ValidateTemplates checks that the templates directory exists and contains required files.
func (s *EmbeddedSource) ValidateTemplates() error {
	// Check templates directory exists
	if info, err := os.Stat(s.templatesDir); err != nil {
		return fmt.Errorf("templates directory not found: %s: %w", s.templatesDir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("templates path is not a directory: %s", s.templatesDir)
	}

	// Check manifest exists
	manifestPath := filepath.Join(s.templatesDir, "manifest.yaml")
	if _, err := os.Stat(manifestPath); err != nil {
		return fmt.Errorf("manifest file not found: %s: %w", manifestPath, err)
	}

	// Validate each template's path exists
	for _, tmpl := range s.manifest.Templates {
		templatePath := filepath.Join(s.templatesDir, tmpl.Path)
		if _, err := os.Stat(templatePath); err != nil {
			return fmt.Errorf("template path not found: %s: %w", templatePath, err)
		}
	}

	return nil
}
