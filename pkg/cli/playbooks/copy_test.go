package playbooks

import "testing"

func TestTransformTemplatePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "go.mod.template transforms to go.mod",
			input:    "go.mod.template",
			expected: "go.mod",
		},
		{
			name:     "nested go.mod.template transforms correctly",
			input:    "backend/go.mod.template",
			expected: "backend/go.mod",
		},
		{
			name:     "regular file unchanged",
			input:    "main.go",
			expected: "main.go",
		},
		{
			name:     "README unchanged",
			input:    "README.md",
			expected: "README.md",
		},
		{
			name:     "other .template file unchanged",
			input:    "config.template",
			expected: "config.template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformTemplatePath(tt.input)
			if result != tt.expected {
				t.Errorf("transformTemplatePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTransformTemplateContent(t *testing.T) {
	tests := []struct {
		name     string
		destPath string
		content  string
		expected string
	}{
		{
			name:     "removes go:build ignore from Go files",
			destPath: "main.go",
			content:  "//go:build ignore\n\npackage main\n\nfunc main() {}\n",
			expected: "package main\n\nfunc main() {}\n",
		},
		{
			name:     "preserves content without go:build ignore",
			destPath: "main.go",
			content:  "package main\n\nfunc main() {}\n",
			expected: "package main\n\nfunc main() {}\n",
		},
		{
			name:     "ignores non-Go files",
			destPath: "README.md",
			content:  "//go:build ignore\n\n# README\n",
			expected: "//go:build ignore\n\n# README\n",
		},
		{
			name:     "handles empty content",
			destPath: "empty.go",
			content:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformTemplateContent(tt.destPath, []byte(tt.content))
			if string(result) != tt.expected {
				t.Errorf("transformTemplateContent(%q, %q) = %q, want %q",
					tt.destPath, tt.content, string(result), tt.expected)
			}
		})
	}
}
