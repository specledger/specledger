package playbooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllTemplatesHaveRequiredStructure(t *testing.T) {
	// Table-driven test for all 7 templates
	tests := []struct {
		templateID    string
		expectedDirs  []string
		expectedFiles []string
	}{
		{
			templateID:    "general-purpose",
			expectedDirs:  []string{".claude", ".specledger"},
			expectedFiles: []string{"AGENTS.md", "mise.toml"},
		},
		{
			templateID:    "full-stack",
			expectedDirs:  []string{"backend", "frontend", "backend/cmd/server", "frontend/src"},
			expectedFiles: []string{"README.md", "docker-compose.yml"},
		},
		{
			templateID:    "batch-data",
			expectedDirs:  []string{"workflows", "cmd/worker", "cmd/starter", "internal/extractors"},
			expectedFiles: []string{"README.md", "go.mod.template", "docker-compose.yml"},
		},
		{
			templateID:    "realtime-workflow",
			expectedDirs:  []string{"cmd/worker", "internal/workflows", "internal/activities"},
			expectedFiles: []string{"README.md", "go.mod.template", "docker-compose.yml"},
		},
		{
			templateID:    "ml-image",
			expectedDirs:  []string{"src/data", "src/models", "src/training", "data/raw"},
			expectedFiles: []string{"README.md", "requirements.txt", "pyproject.toml"},
		},
		{
			templateID:    "realtime-data",
			expectedDirs:  []string{"cmd/producer", "cmd/consumer", "internal/kafka"},
			expectedFiles: []string{"README.md", "go.mod.template"},
		},
		{
			templateID:    "ai-chatbot",
			expectedDirs:  []string{"src/agents", "src/tools", "src/integrations"},
			expectedFiles: []string{"README.md", "requirements.txt", "langgraph.json"},
		},
		{
			templateID:    "adk-chatbot",
			expectedDirs:  []string{"cmd/chatbot", "cmd/server", "internal/agents", "internal/tools"},
			expectedFiles: []string{"README.md", "go.mod.template", "AGENTS.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.templateID, func(t *testing.T) {
			// Get template
			tmpl, err := GetTemplateByID(tt.templateID)
			if err != nil {
				t.Fatalf("GetTemplateByID(%q) error: %v", tt.templateID, err)
			}

			// Verify template has required fields
			if tmpl.ID == "" {
				t.Error("template ID is empty")
			}
			if tmpl.Name == "" {
				t.Error("template Name is empty")
			}
			if tmpl.Description == "" {
				t.Error("template Description is empty")
			}
			if tmpl.Path == "" {
				t.Error("template Path is empty")
			}

			// Template path is relative to the embedded FS root
			// The embedded FS has paths prefixed with "templates/"
			templatePath := filepath.Join("templates", tmpl.Path)

			// Verify template directory exists in embedded FS
			if !Exists(templatePath) {
				t.Errorf("template directory does not exist: %s", templatePath)
			}

			// Check expected directories exist
			for _, dir := range tt.expectedDirs {
				dirPath := filepath.Join(templatePath, dir)
				if !Exists(dirPath) {
					t.Errorf("expected directory missing: %s", dir)
				}
			}

			// Check expected files exist
			for _, file := range tt.expectedFiles {
				filePath := filepath.Join(templatePath, file)
				if !Exists(filePath) {
					t.Errorf("expected file missing: %s", file)
				}
			}
		})
	}
}

func TestTemplateCount(t *testing.T) {
	templates, err := LoadTemplates()
	if err != nil {
		t.Fatalf("LoadTemplates() error: %v", err)
	}

	expectedCount := 8
	if len(templates) != expectedCount {
		t.Errorf("expected %d templates, got %d", expectedCount, len(templates))
	}
}

func TestDefaultTemplateIsGeneralPurpose(t *testing.T) {
	tmpl, err := GetDefaultTemplate()
	if err != nil {
		t.Fatalf("GetDefaultTemplate() error: %v", err)
	}

	if tmpl.ID != "general-purpose" {
		t.Errorf("default template should be general-purpose, got %s", tmpl.ID)
	}

	if !tmpl.IsDefault {
		t.Error("default template should have IsDefault = true")
	}
}

func TestTemplateCopy(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("skipping template copy test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "specledger-template-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy general-purpose template (uses specledger playbook)
	source, err := NewEmbeddedSource()
	if err != nil {
		t.Fatalf("NewEmbeddedSource() error: %v", err)
	}

	result, err := source.Copy("specledger", tmpDir, CopyOptions{
		SkipExisting: true,
		Verbose:      false,
	})
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	if result.FilesCopied == 0 {
		t.Error("no files were copied")
	}

	// Verify key files exist
	expectedFiles := []string{
		"AGENTS.md",
		"mise.toml",
		".gitattributes",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file not copied: %s", file)
		}
	}
}
