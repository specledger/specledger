package mockup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteAndLoadDesignSystem(t *testing.T) {
	dir := t.TempDir()
	dsPath := filepath.Join(dir, "design-system.md")

	ds := &DesignSystem{
		Version:      1,
		Framework:    FrameworkReact,
		ExternalLibs: []string{"@mui/material"},
		Style: &StyleInfo{
			CSSFramework:    "Tailwind CSS",
			StylingApproach: "utility-first",
			ThemeColors: map[string]string{
				"--primary":   "#3b82f6",
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
	dsPath := filepath.Join(dir, "design-system.md")

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
	dsPath := filepath.Join(dir, "design-system.md")

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

	// Should contain Overview table
	if !strings.Contains(content, "## Overview") {
		t.Error("expected Overview heading")
	}
	// Should mention UI Libraries in overview row
	if !strings.Contains(content, "@mui/material") {
		t.Error("expected UI library in output")
	}
	// Should contain CSS info in overview
	if !strings.Contains(content, "Tailwind CSS") {
		t.Error("expected CSS framework in output")
	}
	// Should mention AI agent reads frontmatter
	if !strings.Contains(content, "AI agent") {
		t.Error("expected mention of AI agent")
	}
	// Color palette section should show preview, not full dump
	if !strings.Contains(content, "## Color Palette") {
		t.Error("expected Color Palette heading")
	}
}

func TestLoadDesignSystem_MalformedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	dsPath := filepath.Join(dir, "design-system.md")
	_ = os.WriteFile(dsPath, []byte("no frontmatter here"), 0600)

	_, err := LoadDesignSystem(dsPath)
	if err == nil {
		t.Error("expected error for malformed frontmatter")
	}
}

func TestLoadDesignSystem_EmptyStyle(t *testing.T) {
	dir := t.TempDir()
	dsPath := filepath.Join(dir, "design-system.md")

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
