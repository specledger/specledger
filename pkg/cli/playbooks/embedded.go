package playbooks

import (
	"fmt"
	"path/filepath"

	"github.com/specledger/specledger/pkg/models"
)

// EmbeddedSource implements PlaybookSource for playbooks stored in the embedded filesystem.
type EmbeddedSource struct {
	templatesDir string
	manifest     *PlaybookManifest
}

// NewEmbeddedSource creates a new EmbeddedSource.
func NewEmbeddedSource() (*EmbeddedSource, error) {
	templatesDir, err := PlaybooksDir()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize templates directory: %w", err)
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

// List returns all available playbooks from the embedded source.
func (s *EmbeddedSource) List() ([]Playbook, error) {
	if s.manifest == nil {
		return nil, fmt.Errorf("manifest not loaded")
	}
	return s.manifest.Playbooks, nil
}

// Copy copies the specified playbook to the destination directory.
func (s *EmbeddedSource) Copy(name string, destDir string, opts CopyOptions) (*CopyResult, error) {
	playbook, err := s.getPlaybook(name)
	if err != nil {
		return nil, err
	}

	return CopyPlaybooks(s.templatesDir, destDir, *playbook, opts)
}

// Exists checks if a playbook with the given name exists.
func (s *EmbeddedSource) Exists(name string) bool {
	_, err := s.getPlaybook(name)
	return err == nil
}

// getPlaybook retrieves a playbook by name.
func (s *EmbeddedSource) getPlaybook(name string) (*Playbook, error) {
	if s.manifest == nil {
		return nil, fmt.Errorf("manifest not loaded")
	}

	for _, pb := range s.manifest.Playbooks {
		if pb.Name == name {
			return &pb, nil
		}
	}

	return nil, fmt.Errorf("playbook not found: %s", name)
}

// GetPlaybook retrieves a playbook by name.
// Returns the playbook with the matching name, or an error if not found.
func (s *EmbeddedSource) GetPlaybook(name string) (*Playbook, error) {
	return s.getPlaybook(name)
}

// GetDefaultPlaybook returns the first available playbook.
// For now, there's only one playbook (specledger).
func (s *EmbeddedSource) GetDefaultPlaybook() (*Playbook, error) {
	if s.manifest == nil || len(s.manifest.Playbooks) == 0 {
		return nil, fmt.Errorf("no playbooks available")
	}

	return &s.manifest.Playbooks[0], nil
}

// ValidatePlaybooks checks that the templates directory exists and contains required files.
func (s *EmbeddedSource) ValidatePlaybooks() error {
	// Check manifest exists in embedded FS
	manifestPath := filepath.Join(s.templatesDir, "manifest.yaml")
	if !Exists(manifestPath) {
		return fmt.Errorf("manifest file not found: %s", manifestPath)
	}

	// Validate each playbook's path exists in embedded FS
	for _, pb := range s.manifest.Playbooks {
		playbookPath := filepath.Join(s.templatesDir, pb.Path)
		if !Exists(playbookPath) {
			return fmt.Errorf("playbook path not found in embedded filesystem: %s", playbookPath)
		}
	}

	return nil
}

// LoadTemplates returns all available project templates from the manifest.
// Templates are validated and cached for performance.
func (s *EmbeddedSource) LoadTemplates() ([]models.TemplateDefinition, error) {
	if s.manifest == nil {
		return nil, fmt.Errorf("manifest not loaded")
	}

	// Return templates from manifest (already validated during parsing)
	return s.manifest.Templates, nil
}

// GetTemplateByID retrieves a template by its ID.
// Returns the template if found, or an error if not found.
func (s *EmbeddedSource) GetTemplateByID(id string) (*models.TemplateDefinition, error) {
	if s.manifest == nil {
		return nil, fmt.Errorf("manifest not loaded")
	}

	for i := range s.manifest.Templates {
		if s.manifest.Templates[i].ID == id {
			return &s.manifest.Templates[i], nil
		}
	}

	return nil, fmt.Errorf("template not found: %s", id)
}

// GetDefaultTemplate returns the default template (marked with is_default: true).
// If no template is marked as default, returns the first template.
func (s *EmbeddedSource) GetDefaultTemplate() (*models.TemplateDefinition, error) {
	if s.manifest == nil || len(s.manifest.Templates) == 0 {
		return nil, fmt.Errorf("no templates available")
	}

	// Find template marked as default
	for i := range s.manifest.Templates {
		if s.manifest.Templates[i].IsDefault {
			return &s.manifest.Templates[i], nil
		}
	}

	// Fallback to first template if no default marked
	return &s.manifest.Templates[0], nil
}
