package spec_test

import (
	"strings"
	"testing"
	"specledger/internal/spec"
)

func TestParseManifest(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid manifest",
			content: `# Test Spec
require https://github.com/example/repo main specs/test.md --alias testdep
require https://github.com/other/repo v1.2.3 specs/other.md`,
			wantErr: false,
		},
		{
			name: "empty manifest",
			content: `# Empty`,
			wantErr: false,
		},
		{
			name:    "missing require keyword",
			content: `require: https://github.com/example/repo main specs/test.md`,
			wantErr: true,
		},
		{
			name:    "insufficient parts",
			content: `require github.com/example/repo`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write content to temp file
			tmpFile := writeTestManifest(t, tt.content)

			// Parse manifest
			got, err := spec.ParseManifest(tmpFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got.Dependecies) != 2 {
				t.Errorf("ParseManifest() got %d dependencies, want 2", len(got.Dependecies))
			}
		})
	}
}

func TestWriteManifest(t *testing.T) {
	tests := []struct {
		name    string
		manifest *spec.Manifest
		wantErr bool
	}{
		{
			name: "basic manifest",
			manifest: &spec.Manifest{
				Version:    "1.0.0",
				Dependecies: []models.Dependency{
					{RepositoryURL: "https://github.com/example/repo", Version: "main", SpecPath: "specs/test.md", Alias: "testdep"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := writeTestManifest(t, "")
			defer os.Remove(tmpFile)

			err := spec.WriteManifest(tmpFile, tt.manifest)

			if (err != nil) != tt.wantErr {
				t.Errorf("WriteManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Read back and verify
			got, err := spec.ParseManifest(tmpFile)
			if err != nil {
				t.Errorf("ParseManifest() after write = %v", err)
				return
			}

			if len(got.Dependecies) != len(tt.manifest.Dependecies) {
				t.Errorf("WriteManifest() wrote %d dependencies, want %d", len(got.Dependecies), len(tt.manifest.Dependecies))
			}
		})
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"main", true},
		{"v1.0.0", true},
		{"v1.2.3-beta", true},
		{"#main", true},
		{"#production", true},
		{"abc123", true},
		{"123abc", true},
		{"valid-version_123", true},
		{"invalid!@#", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := spec.IsValidVersion(tt.version)
			if got != tt.want {
				t.Errorf("IsValidVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestIsValidSpecPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"spec.md", true},
		{"specs/feature/spec.md", true},
		{"../../etc/passwd", false},
		{"spec..md", false},
		{"", false},
		{"spec", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := spec.IsValidSpecPath(tt.path)
			if got != tt.want {
				t.Errorf("IsValidSpecPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestFindDuplicateDependencies(t *testing.T) {
	tests := []struct {
		name     string
		deps     []models.Dependency
		wantDupes []string
	}{
		{
			name: "no duplicates",
			deps: []models.Dependency{
				{RepositoryURL: "https://github.com/a/repo", SpecPath: "spec.md"},
				{RepositoryURL: "https://github.com/b/repo", SpecPath: "spec.md"},
			},
			wantDupes: nil,
		},
		{
			name: "one duplicate",
			deps: []models.Dependency{
				{RepositoryURL: "https://github.com/a/repo", SpecPath: "spec.md"},
				{RepositoryURL: "https://github.com/a/repo", SpecPath: "spec.md"},
			},
			wantDupes: []string{"https://github.com/a/repo:spec.md"},
		},
		{
			name: "multiple duplicates",
			deps: []models.Dependency{
				{RepositoryURL: "https://github.com/a/repo", SpecPath: "spec.md"},
				{RepositoryURL: "https://github.com/a/repo", SpecPath: "spec.md"},
				{RepositoryURL: "https://github.com/b/repo", SpecPath: "other.md"},
				{RepositoryURL: "https://github.com/b/repo", SpecPath: "other.md"},
			},
			wantDupes: []string{"https://github.com/a/repo:spec.md", "https://github.com/b/repo:other.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := spec.FindDuplicateDependencies(tt.deps)
			if len(got) != len(tt.wantDupes) {
				t.Errorf("FindDuplicateDependencies() got %v, want %v", got, tt.wantDupes)
				return
			}

			for i, d := range got {
				if d != tt.wantDupes[i] {
					t.Errorf("FindDuplicateDependencies()[%d] = %v, want %v", i, d, tt.wantDupes[i])
				}
			}
		})
	}
}

// Helper function to write test manifest to temp file
func writeTestManifest(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "spec-*.mod")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	if content != "" {
		if _, err := tmpFile.WriteString(content); err != nil {
			t.Fatal(err)
		}
	}

	return tmpFile.Name()
}

// Add missing import
import (
	"os"
	"specledger/pkg/models"
)
