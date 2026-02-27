package mockup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFramework_NextJS(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "next.config.js"), []byte("module.exports = {}"), 0600)

	result, err := DetectFramework(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsFrontend {
		t.Error("expected IsFrontend = true")
	}
	if result.Framework != FrameworkNextJS {
		t.Errorf("expected framework NextJS, got %s", result.Framework)
	}
	if result.Confidence != 99 {
		t.Errorf("expected confidence 99, got %d", result.Confidence)
	}
}

func TestDetectFramework_Angular(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "angular.json"), []byte("{}"), 0600)

	result, err := DetectFramework(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Framework != FrameworkAngular {
		t.Errorf("expected framework Angular, got %s", result.Framework)
	}
}

func TestDetectFramework_PackageJSON(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"dependencies": {"react": "^18.0.0", "react-dom": "^18.0.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0600)

	result, err := DetectFramework(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsFrontend {
		t.Error("expected IsFrontend = true")
	}
	if result.Framework != FrameworkReact {
		t.Errorf("expected framework React, got %s", result.Framework)
	}
	if result.Confidence != 85 {
		t.Errorf("expected confidence 85, got %d", result.Confidence)
	}
}

func TestDetectFramework_FileExtension(t *testing.T) {
	dir := t.TempDir()
	compDir := filepath.Join(dir, "src")
	os.MkdirAll(compDir, 0755)
	os.WriteFile(filepath.Join(compDir, "App.vue"), []byte("<template></template>"), 0600)

	result, err := DetectFramework(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsFrontend {
		t.Error("expected IsFrontend = true")
	}
	if result.Framework != FrameworkVue {
		t.Errorf("expected framework Vue, got %s", result.Framework)
	}
}

func TestDetectFramework_NotFrontend(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0600)

	result, err := DetectFramework(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsFrontend {
		t.Error("expected IsFrontend = false for Go project")
	}
	if result.Framework != FrameworkUnknown {
		t.Errorf("expected framework Unknown, got %s", result.Framework)
	}
}

func TestDetectFramework_SvelteKit(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "svelte.config.js"), []byte("export default {}"), 0600)

	result, err := DetectFramework(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Framework != FrameworkSvelteKit {
		t.Errorf("expected framework SvelteKit, got %s", result.Framework)
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"button", "Button"},
		{"user-card", "UserCard"},
		{"app_header", "AppHeader"},
		{"my-complex-component", "MyComplexComponent"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
