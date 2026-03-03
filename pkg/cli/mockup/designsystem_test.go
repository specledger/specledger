package mockup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteAndLoadDesignSystem(t *testing.T) {
	dir := t.TempDir()
	dsPath := filepath.Join(dir, "design_system.md")

	ds := &DesignSystem{
		Version:      1,
		Framework:    FrameworkReact,
		ExternalLibs: []string{"@mui/material"},
		Style: &StyleInfo{
			CSSFramework:    "Tailwind CSS",
			StylingApproach: "utility-first",
			ThemeColors: map[string]string{
				"--primary": "#3b82f6",
				"--secondary": "#64748b",
			},
			FontFamilies: []string{"Inter, sans-serif"},
			CSSVariables: []string{"--primary", "--secondary", "--background"},
		},
	}

	// Write
	err := WriteDesignSystem(dsPath, ds)
	if err != nil {
		t.Fatal(err)
	}

	// Read back
	loaded, err := LoadDesignSystem(dsPath)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Version != 1 {
		t.Errorf("Version = %d, want 1", loaded.Version)
	}
	if loaded.Framework != FrameworkReact {
		t.Errorf("Framework = %s, want react", loaded.Framework)
	}
	if loaded.Style == nil {
		t.Fatal("Style is nil, expected StyleInfo")
	}
	if loaded.Style.CSSFramework != "Tailwind CSS" {
		t.Errorf("CSSFramework = %q, want Tailwind CSS", loaded.Style.CSSFramework)
	}
}

func TestWriteDesignSystem_WithStyleInfo(t *testing.T) {
	dir := t.TempDir()
	dsPath := filepath.Join(dir, "design_system.md")

	ds := &DesignSystem{
		Version:   1,
		Framework: FrameworkVue,
		Style: &StyleInfo{
			CSSFramework:    "Bootstrap",
			StylingApproach: "utility-first",
			ThemeColors: map[string]string{
				"--bs-primary": "#0d6efd",
			},
		},
	}

	err := WriteDesignSystem(dsPath, ds)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(dsPath)
	content := string(data)

	if !strings.Contains(content, "Bootstrap") {
		t.Error("expected Bootstrap in output")
	}
	if !strings.Contains(content, "--bs-primary") {
		t.Error("expected theme color in output")
	}
}

func TestWriteDesignSystem_MarkdownFormat(t *testing.T) {
	dir := t.TempDir()
	dsPath := filepath.Join(dir, "design_system.md")

	ds := &DesignSystem{
		Version:      1,
		Framework:    FrameworkReact,
		ExternalLibs: []string{"@mui/material", "@chakra-ui/react"},
		Style: &StyleInfo{
			CSSFramework:    "Tailwind CSS",
			StylingApproach: "utility-first",
			ThemeColors: map[string]string{
				"--primary": "#3b82f6",
			},
			FontFamilies: []string{"Inter"},
			CSSVariables: []string{"--primary", "--background"},
		},
	}

	err := WriteDesignSystem(dsPath, ds)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(dsPath)
	content := string(data)

	// Should contain UI Libraries section
	if !strings.Contains(content, "## UI Libraries") {
		t.Error("expected UI Libraries heading")
	}
	// Should contain Styling section
	if !strings.Contains(content, "## Styling") {
		t.Error("expected Styling heading")
	}
	// Should mention AI agent for components
	if !strings.Contains(content, "AI agent") {
		t.Error("expected mention of AI agent for component discovery")
	}
}

func TestLoadDesignSystem_MalformedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	dsPath := filepath.Join(dir, "design_system.md")
	os.WriteFile(dsPath, []byte("no frontmatter here"), 0600)

	_, err := LoadDesignSystem(dsPath)
	if err == nil {
		t.Error("expected error for malformed frontmatter")
	}
}

func TestLoadDesignSystem_EmptyStyle(t *testing.T) {
	dir := t.TempDir()
	dsPath := filepath.Join(dir, "design_system.md")

	ds := &DesignSystem{
		Version:   1,
		Framework: FrameworkSvelte,
	}

	err := WriteDesignSystem(dsPath, ds)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadDesignSystem(dsPath)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Framework != FrameworkSvelte {
		t.Errorf("Framework = %s, want svelte", loaded.Framework)
	}
}
