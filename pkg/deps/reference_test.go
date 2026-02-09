package deps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveReference(t *testing.T) {
	tests := []struct {
		name                string
		projectArtifactPath string
		depAlias            string
		artifactName        string
		projectRoot         string
		wantPath            string
		wantErr             bool
		setupFunc           func() string // Returns temp dir to clean up
	}{
		{
			name:                "valid simple reference",
			projectArtifactPath: "specledger/",
			depAlias:            "platform",
			artifactName:        "api.md",
			projectRoot:         "",
			wantPath:            "specledger/platform/api.md",
			wantErr:             false,
		},
		{
			name:                "valid nested artifact",
			projectArtifactPath: "specledger/",
			depAlias:            "platform",
			artifactName:        "auth/requirements.md",
			projectRoot:         "",
			wantPath:            "specledger/platform/auth/requirements.md",
			wantErr:             false,
		},
		{
			name:                "empty project artifact path",
			projectArtifactPath: "",
			depAlias:            "platform",
			artifactName:        "api.md",
			projectRoot:         "",
			wantPath:            "",
			wantErr:             true,
		},
		{
			name:                "empty dependency alias",
			projectArtifactPath: "specledger/",
			depAlias:            "",
			artifactName:        "api.md",
			projectRoot:         "",
			wantPath:            "",
			wantErr:             true,
		},
		{
			name:                "empty artifact name",
			projectArtifactPath: "specledger/",
			depAlias:            "platform",
			artifactName:        "",
			projectRoot:         "",
			wantPath:            "",
			wantErr:             true,
		},
		{
			name:                "artifact path without trailing slash",
			projectArtifactPath: "specledger",
			depAlias:            "platform",
			artifactName:        "api.md",
			projectRoot:         "",
			wantPath:            "specledger/platform/api.md",
			wantErr:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			if tt.setupFunc != nil {
				tempDir := tt.setupFunc()
				cleanup = func() { os.RemoveAll(tempDir) }
			}
			if cleanup != nil {
				defer cleanup()
			}

			gotPath, err := ResolveReference(tt.projectArtifactPath, tt.depAlias, tt.artifactName, tt.projectRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPath != tt.wantPath {
				t.Errorf("ResolveReference() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestResolveReferenceWithProjectRoot(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create the artifact directory structure
	artifactDir := filepath.Join(tempDir, "specledger", "platform")
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create a test artifact file
	artifactFile := filepath.Join(artifactDir, "api.md")
	if err := os.WriteFile(artifactFile, []byte("# API Spec"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test resolution with project root
	gotPath, err := ResolveReference("specledger/", "platform", "api.md", tempDir)
	if err != nil {
		t.Errorf("ResolveReference() error = %v", err)
		return
	}
	if gotPath != "specledger/platform/api.md" {
		t.Errorf("ResolveReference() = %v, want specledger/platform/api.md", gotPath)
	}
}

func TestResolveReferenceMissingArtifact(t *testing.T) {
	tempDir := t.TempDir()

	// Don't create the artifact file - it should error
	_, err := ResolveReference("specledger/", "platform", "api.md", tempDir)
	if err == nil {
		t.Error("ResolveReference() expected error for missing artifact, got nil")
	}
}
