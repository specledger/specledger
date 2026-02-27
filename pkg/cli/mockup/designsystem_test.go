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
		Version:       1,
		Framework:     FrameworkReact,
		ComponentDirs: []string{"src/components"},
		ExternalLibs:  []string{"@mui/material"},
		Components: []Component{
			{
				Name:        "Button",
				FilePath:    "src/components/Button.tsx",
				Description: "Primary action button",
				Props: []PropInfo{
					{Name: "variant", Type: "string"},
					{Name: "onClick", Type: "func"},
				},
			},
			{
				Name:        "Card",
				FilePath:    "src/components/Card.tsx",
				Description: "Container component",
			},
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
	if len(loaded.Components) != 2 {
		t.Errorf("Components count = %d, want 2", len(loaded.Components))
	}
	if loaded.Components[0].Name != "Button" {
		t.Errorf("First component = %q, want Button", loaded.Components[0].Name)
	}
}

func TestWriteDesignSystem_WithManualEntries(t *testing.T) {
	dir := t.TempDir()
	dsPath := filepath.Join(dir, "design_system.md")

	ds := &DesignSystem{
		Version:   1,
		Framework: FrameworkVue,
		Components: []Component{
			{Name: "Header", FilePath: "src/components/Header.vue"},
		},
		ManualEntries: []Component{
			{Name: "CustomWidget", Description: "A manually added component"},
		},
	}

	err := WriteDesignSystem(dsPath, ds)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(dsPath)
	content := string(data)

	if !strings.Contains(content, "<!-- MANUAL -->") {
		t.Error("expected MANUAL markers in output")
	}
	if !strings.Contains(content, "CustomWidget") {
		t.Error("expected manual entry in output")
	}
}

func TestWriteDesignSystem_TreeFormat(t *testing.T) {
	dir := t.TempDir()
	dsPath := filepath.Join(dir, "design_system.md")

	ds := &DesignSystem{
		Version:       1,
		Framework:     FrameworkReact,
		ComponentDirs: []string{"src/components"},
		Components: []Component{
			{Name: "Button", FilePath: "src/components/Button.tsx", Props: []PropInfo{{Name: "variant", Type: "string"}}},
			{Name: "LoginForm", FilePath: "src/components/auth/LoginForm.tsx", Props: []PropInfo{{Name: "onSubmit", Type: "func"}}},
			{Name: "Header", FilePath: "src/components/layout/Header.tsx"},
		},
	}

	err := WriteDesignSystem(dsPath, ds)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(dsPath)
	content := string(data)

	// Should contain tree connectors
	if !strings.Contains(content, "├──") && !strings.Contains(content, "└──") {
		t.Error("expected tree connectors in output")
	}
	// Should show component arrow syntax
	if !strings.Contains(content, "→") {
		t.Error("expected → arrows in tree output")
	}
	// Should show props inline
	if !strings.Contains(content, "variant: string") {
		t.Error("expected props in tree output")
	}
	// Should show directory structure
	if !strings.Contains(content, "auth/") {
		t.Error("expected auth/ directory in tree")
	}

	// Round-trip: load back and verify
	loaded, err := LoadDesignSystem(dsPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Components) != 3 {
		t.Errorf("loaded %d components, want 3", len(loaded.Components))
	}
	// Check props survived round-trip
	for _, c := range loaded.Components {
		if c.Name == "Button" {
			if len(c.Props) == 0 {
				t.Error("Button props lost in round-trip")
			} else if c.Props[0].Name != "variant" {
				t.Errorf("Button prop name = %q, want variant", c.Props[0].Name)
			}
		}
		if c.Name == "LoginForm" {
			if c.FilePath != "src/components/auth/LoginForm.tsx" {
				t.Errorf("LoginForm path = %q, want src/components/auth/LoginForm.tsx", c.FilePath)
			}
		}
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

func TestMergeDesignSystem(t *testing.T) {
	existing := &DesignSystem{
		Version:   1,
		Framework: FrameworkReact,
		Components: []Component{
			{Name: "Button", FilePath: "src/components/Button.tsx"},
			{Name: "OldCard", FilePath: "src/components/OldCard.tsx"},
		},
		ManualEntries: []Component{
			{Name: "CustomWidget"},
		},
	}

	scanResult := &ScanResult{
		Components: []Component{
			{Name: "Button", FilePath: "src/components/Button.tsx"},
			{Name: "NewDialog", FilePath: "src/components/NewDialog.tsx"},
		},
		ComponentDirs: []string{"src/components"},
		ExternalLibs:  []string{"@mui/material"},
	}

	added, removed := MergeDesignSystem(existing, scanResult)

	if added != 1 {
		t.Errorf("added = %d, want 1", added)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	// ManualEntries should be preserved
	if len(existing.ManualEntries) != 1 {
		t.Errorf("ManualEntries count = %d, want 1", len(existing.ManualEntries))
	}

	// Components should match scan result
	if len(existing.Components) != 2 {
		t.Errorf("Components count = %d, want 2", len(existing.Components))
	}
}
