package templates

import (
	"fmt"
)

// RemoteSource implements TemplateSource for templates fetched from remote URLs.
// This is a stub for future implementation - the architecture supports it
// but the actual functionality is not yet implemented.
//
// Future implementation will support:
// - Git repositories (git@github.com:org/template.git)
// - HTTPS URLs to template archives
// - Template caching at ~/.specledger/template-cache/
// - Authentication for private repositories
type RemoteSource struct {
	baseURL string
	cacheDir string
}

// NewRemoteSource creates a new RemoteSource for fetching templates from a remote URL.
// Future implementation will support various URL formats.
func NewRemoteSource(baseURL string) (*RemoteSource, error) {
	// Validate URL format (placeholder for future implementation)
	if baseURL == "" {
		return nil, fmt.Errorf("base URL cannot be empty")
	}

	return &RemoteSource{
		baseURL: baseURL,
		// cacheDir will be set when cache is implemented
	}, nil
}

// List returns all available templates from the remote source.
// Future implementation will:
// 1. Fetch manifest from remote URL
// 2. Parse and return template list
// 3. Handle authentication and errors
func (s *RemoteSource) List() ([]Template, error) {
	// TODO: Implement remote template fetching
	// This will involve:
	// - HTTP request to {baseURL}/manifest.yaml
	// - Parse manifest YAML
	// - Return templates list
	return nil, fmt.Errorf("remote template source not yet implemented - use embedded templates instead")
}

// Copy copies the specified template from the remote source to the destination directory.
// Future implementation will:
// 1. Fetch template archive or clone git repository
// 2. Extract/copy to destination directory
// 3. Handle caching, versioning, and updates
func (s *RemoteSource) Copy(name string, destDir string, opts CopyOptions) (*CopyResult, error) {
	// TODO: Implement remote template copying
	// This will involve:
	// - Download template from remote source
	// - Extract to destination
	// - Update local cache
	return nil, fmt.Errorf("remote template copying not yet implemented - use embedded templates instead")
}

// Exists checks if a template with the given name exists in the remote source.
// Future implementation will check the remote manifest.
func (s *RemoteSource) Exists(name string) bool {
	// TODO: Implement remote template existence check
	return false
}

// RefreshCache updates the local cache of remote templates.
// Future implementation for when caching is added.
func (s *RemoteSource) RefreshCache() error {
	// TODO: Implement cache refresh
	return fmt.Errorf("template cache not yet implemented")
}
