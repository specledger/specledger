package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTechnicalContext(t *testing.T) {
	t.Run("valid plan with all fields", func(t *testing.T) {
		content := `# Plan

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: Cobra CLI
**Storage**: YAML files
**Testing**: go test
**Target Platform**: Linux/macOS/Windows
**Project Type**: CLI Tool
**Performance Goals**: Fast startup
**Constraints**: Single binary
**Scale/Scope**: Small teams

## Next Section
`
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "plan.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		ctx, err := ParseTechnicalContext(path)
		if err != nil {
			t.Fatalf("ParseTechnicalContext() error: %v", err)
		}

		if ctx.Language != "Go 1.24" {
			t.Errorf("expected Language 'Go 1.24', got %q", ctx.Language)
		}
		if ctx.PrimaryDeps != "Cobra CLI" {
			t.Errorf("expected PrimaryDeps 'Cobra CLI', got %q", ctx.PrimaryDeps)
		}
		if ctx.Storage != "YAML files" {
			t.Errorf("expected Storage 'YAML files', got %q", ctx.Storage)
		}
		if ctx.Testing != "go test" {
			t.Errorf("expected Testing 'go test', got %q", ctx.Testing)
		}
		if ctx.TargetPlatform != "Linux/macOS/Windows" {
			t.Errorf("expected TargetPlatform 'Linux/macOS/Windows', got %q", ctx.TargetPlatform)
		}
		if ctx.ProjectType != "CLI Tool" {
			t.Errorf("expected ProjectType 'CLI Tool', got %q", ctx.ProjectType)
		}
		if ctx.PerformanceGoals != "Fast startup" {
			t.Errorf("expected PerformanceGoals 'Fast startup', got %q", ctx.PerformanceGoals)
		}
		if ctx.Constraints != "Single binary" {
			t.Errorf("expected Constraints 'Single binary', got %q", ctx.Constraints)
		}
		if ctx.Scale != "Small teams" {
			t.Errorf("expected Scale 'Small teams', got %q", ctx.Scale)
		}
	})

	t.Run("partial fields", func(t *testing.T) {
		content := `# Plan

## Technical Context

**Language/Version**: Go 1.24
**Testing**: go test

## Next Section
`
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "plan.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		ctx, err := ParseTechnicalContext(path)
		if err != nil {
			t.Fatalf("ParseTechnicalContext() error: %v", err)
		}

		if ctx.Language != "Go 1.24" {
			t.Errorf("expected Language 'Go 1.24', got %q", ctx.Language)
		}
		if ctx.Testing != "go test" {
			t.Errorf("expected Testing 'go test', got %q", ctx.Testing)
		}
		if ctx.Storage != "" {
			t.Errorf("expected empty Storage, got %q", ctx.Storage)
		}
	})

	t.Run("no technical context section", func(t *testing.T) {
		content := `# Plan

## Other Section

Some content
`
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "plan.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := ParseTechnicalContext(path)
		if err == nil {
			t.Error("expected error for missing technical context")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := ParseTechnicalContext("/nonexistent/plan.md")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}

func TestTechnicalContextString(t *testing.T) {
	ctx := &TechnicalContext{
		Language:       "Go 1.24",
		PrimaryDeps:    "Cobra CLI",
		Storage:        "YAML files",
		Testing:        "go test",
		TargetPlatform: "Linux",
		ProjectType:    "CLI Tool",
	}

	result := ctx.String()
	if result == "" {
		t.Error("String() should not be empty")
	}
}

func TestTechnicalContextToMarkdown(t *testing.T) {
	ctx := &TechnicalContext{
		Language:         "Go 1.24",
		PrimaryDeps:      "Cobra CLI",
		Storage:          "YAML files",
		Testing:          "go test",
		TargetPlatform:   "Linux",
		ProjectType:      "CLI Tool",
		PerformanceGoals: "Fast",
		Constraints:      "Single binary",
		Scale:            "Small",
	}

	result := ctx.ToMarkdown()
	if result == "" {
		t.Error("ToMarkdown() should not be empty")
	}
}
