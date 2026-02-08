package ref

import (
	"fmt"
	"regexp"
	"strings"
)

// Reference represents an external reference in a specification
type Reference struct {
	Text     string // The full text of the reference
	Markdown string // The markdown link text
	URL      string // The resolved URL
	Line     int    // Line number where reference was found
	Column   int    // Column number where reference was found
	Type     string // Type of reference: "markdown", "image", "code", "link"
}

// ReferenceResolver resolves external references in specifications
type ReferenceResolver struct {
	lockfilePath string
	dependencies map[string]string // alias -> repository URL mapping
}

// NewResolver creates a new reference resolver
func NewResolver(lockfilePath string) *ReferenceResolver {
	return &ReferenceResolver{
		lockfilePath: lockfilePath,
	}
}

// ParseSpec parses a specification file and extracts references
func (r *ReferenceResolver) ParseSpec(content string) ([]Reference, error) {
	// Extract all markdown links
	markdownLinks := extractMarkdownLinks(content)

	// Extract image links
	imageLinks := extractImageLinks(content)

	// Extract inline code references (custom syntax like `spec.example#section`)
	inlineRefs := extractInlineReferences(content)

	var references []Reference

	// Add markdown links
	for _, link := range markdownLinks {
		references = append(references, Reference{
			Text:     link.text,
			Markdown: link.markdown,
			URL:      link.url,
			Type:     "markdown",
		})
	}

	// Add image links
	for _, img := range imageLinks {
		references = append(references, Reference{
			Text:     img.text,
			Markdown: img.markdown,
			URL:      img.url,
			Type:     "image",
		})
	}

	// Add inline references
	for _, ref := range inlineRefs {
		references = append(references, Reference{
			Text:     ref.text,
			Markdown: ref.markdown,
			URL:      ref.url,
			Type:     "inline",
		})
	}

	return references, nil
}

// ValidateReferences validates all references against the lockfile
func (r *ReferenceResolver) ValidateReferences(references []Reference) []ValidationError {
	var errors []ValidationError

	for _, ref := range references {
		if err := r.validateReference(ref); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// validateReference validates a single reference
func (r *ReferenceResolver) validateReference(ref Reference) *ValidationError {
	// Skip image references for now (they can be ignored or validated separately)
	if ref.Type == "image" {
		return nil
	}

	// Try to resolve the reference
	resolvedURL, err := r.resolveReference(ref.URL)
	if err != nil {
		return &ValidationError{
			Reference: ref,
			Field:     "url",
			Message:   fmt.Sprintf("failed to resolve reference: %w", err),
		}
	}

	// Check if the URL matches any dependency
	matched := false
	for _, repoURL := range r.dependencies {
		if resolvedURL == repoURL {
			matched = true
			break
		}
	}

	if !matched {
		return &ValidationError{
			Reference: ref,
			Field:     "url",
			Message:   fmt.Sprintf("reference points to unknown dependency: %s", resolvedURL),
		}
	}

	return nil
}

// resolveReference tries to resolve a reference URL
func (r *ReferenceResolver) resolveReference(url string) (string, error) {
	// Try to parse as a full URL first
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url, nil
	}

	// Try to resolve relative paths
	if strings.HasPrefix(url, "#") {
		// This is an anchor reference
		return url, nil
	}

	// Try to match against known dependencies
	// This is a placeholder - actual resolution depends on the lockfile
	if depURL, ok := r.dependencies[url]; ok {
		return depURL, nil
	}

	return "", fmt.Errorf("unknown reference: %s", url)
}

// SetDependencies sets the dependencies map for validation
func (r *ReferenceResolver) SetDependencies(deps map[string]string) error {
	r.dependencies = deps
	return nil
}

// GetDependencies gets the dependencies map
func (r *ReferenceResolver) GetDependencies() map[string]string {
	return r.dependencies
}

// resolveReferenceByAlias tries to resolve a reference using its alias
func (r *ReferenceResolver) resolveReferenceByAlias(alias string) (string, error) {
	// This is a placeholder - actual resolution depends on the lockfile
	return "", fmt.Errorf("alias not found: %s", alias)
}

// extractMarkdownLinks extracts all markdown links from content
func extractMarkdownLinks(content string) []linkInfo {
	pattern := `\[([^\]]+)\]\(([^)]+)\)`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(content, -1)

	var links []linkInfo
	for _, match := range matches {
		if len(match) >= 3 {
			links = append(links, linkInfo{
				markdown: match[0],
				text:     match[1],
				url:      match[2],
			})
		}
	}
	return links
}

// extractImageLinks extracts all markdown images from content
func extractImageLinks(content string) []linkInfo {
	pattern := `!\[([^\]]*)\]\(([^)]+)\)`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(content, -1)

	var images []linkInfo
	for _, match := range matches {
		if len(match) >= 3 {
			images = append(images, linkInfo{
				markdown: match[0],
				text:     match[1],
				url:      match[2],
			})
		}
	}
	return images
}

// extractInlineReferences extracts custom inline references (e.g., spec.example#section)
func extractInlineReferences(content string) []linkInfo {
	// Pattern for spec-reference syntax: spec.alias#section or spec-url#section
	pattern := `spec\.([^\s\)\]+\[#?([^\]]+)?\]?)`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(content, -1)

	var refs []linkInfo
	for _, match := range matches {
		if len(match) >= 3 {
			alias := match[1]
			section := match[2]
			url := fmt.Sprintf("spec.%s", alias)
			if section != "" {
				url += "#" + section
			}
			refs = append(refs, linkInfo{
				markdown: match[0],
				text:     match[0],
				url:      url,
			})
		}
	}
	return refs
}

// linkInfo holds info about a parsed link
type linkInfo struct {
	markdown string
	text     string
	url      string
}

// ValidationError represents an error from reference validation
type ValidationError struct {
	Reference Reference
	Field     string
	Message   string
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
