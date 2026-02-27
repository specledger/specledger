package mockup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanComponents_React(t *testing.T) {
	dir := t.TempDir()
	compDir := filepath.Join(dir, "src", "components")
	os.MkdirAll(compDir, 0755)

	// Create a React component file
	reactComp := `import React from 'react';

interface ButtonProps {
  variant: string;
  onClick: () => void;
  disabled?: boolean;
}

export default function Button({ variant, onClick, disabled }: ButtonProps) {
  return (
    <button className={variant} onClick={onClick} disabled={disabled}>
      Click me
    </button>
  );
}
`
	os.WriteFile(filepath.Join(compDir, "Button.tsx"), []byte(reactComp), 0600)

	result, err := ScanComponents(dir, FrameworkReact)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Components) == 0 {
		t.Fatal("expected at least one component")
	}

	found := false
	for _, c := range result.Components {
		if c.Name == "Button" {
			found = true
			if len(c.Props) == 0 {
				t.Error("expected Button to have props")
			}
			break
		}
	}
	if !found {
		t.Error("expected to find Button component")
	}
}

func TestScanComponents_Vue(t *testing.T) {
	dir := t.TempDir()
	compDir := filepath.Join(dir, "src", "components")
	os.MkdirAll(compDir, 0755)

	vueComp := `<template>
  <div class="card">
    <h2>{{ title }}</h2>
    <slot />
  </div>
</template>

<script setup>
defineProps<{
  title: string;
  elevated?: boolean;
}>()
</script>
`
	os.WriteFile(filepath.Join(compDir, "Card.vue"), []byte(vueComp), 0600)

	result, err := ScanComponents(dir, FrameworkVue)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Components) == 0 {
		t.Fatal("expected at least one component")
	}
	if result.Components[0].Name != "Card" {
		t.Errorf("expected component name Card, got %s", result.Components[0].Name)
	}
}

func TestScanComponents_EmptyProject(t *testing.T) {
	dir := t.TempDir()

	result, err := ScanComponents(dir, FrameworkReact)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Components) != 0 {
		t.Errorf("expected 0 components for empty project, got %d", len(result.Components))
	}
}

func TestScanComponents_SkipsNodeModules(t *testing.T) {
	dir := t.TempDir()
	nmDir := filepath.Join(dir, "node_modules", "some-package", "components")
	os.MkdirAll(nmDir, 0755)
	os.WriteFile(filepath.Join(nmDir, "Internal.tsx"), []byte(`export default function Internal() { return <div /> }`), 0600)

	// Also create a legitimate component
	compDir := filepath.Join(dir, "src", "components")
	os.MkdirAll(compDir, 0755)
	os.WriteFile(filepath.Join(compDir, "Real.tsx"), []byte(`export default function Real() { return <div /> }`), 0600)

	result, err := ScanComponents(dir, FrameworkReact)
	if err != nil {
		t.Fatal(err)
	}

	// Should only find Real, not Internal from node_modules
	for _, c := range result.Components {
		if c.Name == "Internal" {
			t.Error("should not scan node_modules components")
		}
	}
}

func TestScanResultToDesignSystem(t *testing.T) {
	sr := &ScanResult{
		Components: []Component{
			{Name: "Button", FilePath: "src/components/Button.tsx"},
		},
		Framework:     FrameworkReact,
		ComponentDirs: []string{"src/components"},
		ExternalLibs:  []string{"@mui/material"},
	}

	ds := ScanResultToDesignSystem(sr, FrameworkReact)

	if ds.Version != 1 {
		t.Errorf("Version = %d, want 1", ds.Version)
	}
	if ds.Framework != FrameworkReact {
		t.Errorf("Framework = %s, want react", ds.Framework)
	}
	if len(ds.Components) != 1 {
		t.Errorf("Components = %d, want 1", len(ds.Components))
	}
}

func TestComponentNameFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"src/components/Button.tsx", "Button"},
		{"src/components/user-card.vue", "UserCard"},
		{"src/components/Button/index.tsx", "Button"},
		{"src/app/hero.component.ts", "Hero"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := componentNameFromPath(tt.path)
			if got != tt.want {
				t.Errorf("componentNameFromPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
