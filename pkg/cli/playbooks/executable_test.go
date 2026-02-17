package playbooks

import (
	"testing"
)

func TestIsExecutableFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  []byte
		expected bool
	}{
		{
			name:     "shell script with .sh extension",
			filename: "script.sh",
			content:  []byte("#!/bin/bash\necho hello"),
			expected: true,
		},
		{
			name:     "shell script without shebang",
			filename: "setup.sh",
			content:  []byte("echo hello"),
			expected: true,
		},
		{
			name:     "file with shebang but no .sh extension",
			filename: "script",
			content:  []byte("#!/usr/bin/env python3\nprint('hello')"),
			expected: true,
		},
		{
			name:     "regular markdown file",
			filename: "README.md",
			content:  []byte("# Title\n\nContent"),
			expected: false,
		},
		{
			name:     "regular yaml file",
			filename: "config.yaml",
			content:  []byte("key: value"),
			expected: false,
		},
		{
			name:     "empty file with .sh extension",
			filename: "empty.sh",
			content:  []byte{},
			expected: true,
		},
		{
			name:     "empty file without .sh extension",
			filename: "empty",
			content:  []byte{},
			expected: false,
		},
		{
			name:     "file starting with hash but not shebang",
			filename: "config",
			content:  []byte("# This is a comment\nkey: value"),
			expected: false,
		},
		{
			name:     "JSON file",
			filename: "data.json",
			content:  []byte(`{"key": "value"}`),
			expected: false,
		},
		{
			name:     "Python script with .sh extension",
			filename: "wrapper.sh",
			content:  []byte("#!/usr/bin/env python3\nimport sys"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExecutableFile(tt.filename, tt.content)
			if result != tt.expected {
				t.Errorf("IsExecutableFile(%q, %q) = %v, expected %v", tt.filename, tt.content, result, tt.expected)
			}
		})
	}
}
