package ref

import (
	"testing"
)

func TestNewResolver(t *testing.T) {
	resolver := NewResolver("/path/to/lockfile")
	if resolver == nil {
		t.Fatal("expected resolver to be created")
	}
	if resolver.lockfilePath != "/path/to/lockfile" {
		t.Errorf("expected lockfilePath '/path/to/lockfile', got %q", resolver.lockfilePath)
	}
}

func TestParseSpec(t *testing.T) {
	resolver := NewResolver("")

	t.Run("parse markdown links", func(t *testing.T) {
		content := `# Test Spec

This is a [link to docs](https://example.com/docs) and [another link](https://example.com/other).
`
		refs, err := resolver.ParseSpec(content)
		if err != nil {
			t.Fatalf("ParseSpec() error: %v", err)
		}

		if len(refs) < 2 {
			t.Errorf("expected at least 2 references, got %d", len(refs))
		}

		// Check first reference
		found := false
		for _, ref := range refs {
			if ref.URL == "https://example.com/docs" {
				found = true
				if ref.Type != "markdown" {
					t.Errorf("expected type 'markdown', got %q", ref.Type)
				}
				break
			}
		}
		if !found {
			t.Error("expected to find reference to https://example.com/docs")
		}
	})

	t.Run("parse image links", func(t *testing.T) {
		content := `# Test Spec

![diagram](images/diagram.png)

Some text here.
`
		refs, err := resolver.ParseSpec(content)
		if err != nil {
			t.Fatalf("ParseSpec() error: %v", err)
		}

		found := false
		for _, ref := range refs {
			if ref.Type == "image" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected to find image reference")
		}
	})

	t.Run("empty content", func(t *testing.T) {
		refs, err := resolver.ParseSpec("")
		if err != nil {
			t.Fatalf("ParseSpec() error: %v", err)
		}
		if len(refs) != 0 {
			t.Errorf("expected 0 references for empty content, got %d", len(refs))
		}
	})
}

func TestSetAndGetDependencies(t *testing.T) {
	resolver := NewResolver("")

	deps := map[string]string{
		"alias1": "https://github.com/repo1",
		"alias2": "https://github.com/repo2",
	}

	if err := resolver.SetDependencies(deps); err != nil {
		t.Fatalf("SetDependencies() error: %v", err)
	}

	got := resolver.GetDependencies()
	if len(got) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(got))
	}

	if got["alias1"] != "https://github.com/repo1" {
		t.Errorf("expected alias1 -> repo1, got %s", got["alias1"])
	}
}

func TestValidateReferences(t *testing.T) {
	t.Run("valid reference", func(t *testing.T) {
		resolver := NewResolver("")
		if err := resolver.SetDependencies(map[string]string{
			"test": "https://github.com/test/repo",
		}); err != nil {
			t.Fatalf("SetDependencies() error: %v", err)
		}

		refs := []Reference{
			{
				Text: "test",
				URL:  "https://github.com/test/repo",
				Type: "markdown",
			},
		}

		errors := resolver.ValidateReferences(refs)
		if len(errors) != 0 {
			t.Errorf("expected no errors, got %d", len(errors))
		}
	})

	t.Run("image reference is skipped", func(t *testing.T) {
		resolver := NewResolver("")

		refs := []Reference{
			{
				Text: "image",
				URL:  "local-image.png",
				Type: "image",
			},
		}

		errors := resolver.ValidateReferences(refs)
		if len(errors) != 0 {
			t.Errorf("expected no errors for image reference, got %d", len(errors))
		}
	})

	t.Run("unknown reference", func(t *testing.T) {
		resolver := NewResolver("")
		if err := resolver.SetDependencies(map[string]string{}); err != nil {
			t.Fatalf("SetDependencies() error: %v", err)
		}

		refs := []Reference{
			{
				Text: "unknown",
				URL:  "spec.unknown#section",
				Type: "inline",
			},
		}

		errors := resolver.ValidateReferences(refs)
		if len(errors) == 0 {
			t.Error("expected error for unknown reference")
		}
	})
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Reference: Reference{Text: "test"},
		Field:     "url",
		Message:   "invalid URL",
	}

	expected := "url: invalid URL"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestExtractMarkdownLinks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "single link",
			content:  "[text](https://example.com)",
			expected: 1,
		},
		{
			name:     "multiple links",
			content:  "[link1](url1) and [link2](url2)",
			expected: 2,
		},
		{
			name:     "no links",
			content:  "just plain text",
			expected: 0,
		},
		{
			name:     "link with spaces",
			content:  "[my link](https://example.com/path with spaces)",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			links := extractMarkdownLinks(tt.content)
			if len(links) != tt.expected {
				t.Errorf("expected %d links, got %d", tt.expected, len(links))
			}
		})
	}
}

func TestExtractImageLinks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "single image",
			content:  "![alt text](image.png)",
			expected: 1,
		},
		{
			name:     "multiple images",
			content:  "![img1](a.png) and ![img2](b.png)",
			expected: 2,
		},
		{
			name:     "no images",
			content:  "just plain text",
			expected: 0,
		},
		{
			name:     "distinguish from link",
			content:  "[link](url) and ![image](img.png)",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			images := extractImageLinks(tt.content)
			if len(images) != tt.expected {
				t.Errorf("expected %d images, got %d", tt.expected, len(images))
			}
		})
	}
}

func TestResolveReference(t *testing.T) {
	resolver := NewResolver("")
	if err := resolver.SetDependencies(map[string]string{
		"mydep": "https://github.com/org/repo",
	}); err != nil {
		t.Fatalf("SetDependencies() error: %v", err)
	}

	tests := []struct {
		url      string
		wantErr  bool
		expected string
	}{
		{
			url:      "https://example.com",
			wantErr:  false,
			expected: "https://example.com",
		},
		{
			url:      "http://example.com",
			wantErr:  false,
			expected: "http://example.com",
		},
		{
			url:      "#section",
			wantErr:  false,
			expected: "#section",
		},
		{
			url:      "mydep",
			wantErr:  false,
			expected: "https://github.com/org/repo",
		},
		{
			url:     "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result, err := resolver.resolveReference(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}
