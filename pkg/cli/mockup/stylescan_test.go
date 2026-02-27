package mockup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanStyles_TailwindProject(t *testing.T) {
	dir := t.TempDir()

	// Create package.json with tailwindcss
	pkgJSON := `{"dependencies":{"tailwindcss":"^3.4.0","react":"^18.0.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0600)

	// Create tailwind config
	os.WriteFile(filepath.Join(dir, "tailwind.config.js"), []byte("module.exports = {}"), 0600)

	// Create globals.css with CSS variables
	os.MkdirAll(filepath.Join(dir, "src", "app"), 0755)
	globalCSS := `@tailwind base;
@tailwind components;
@tailwind utilities;

:root {
  --background: #ffffff;
  --foreground: #171717;
  --primary: #2563eb;
  --muted: #f5f5f5;
}

body {
  font-family: Inter, system-ui, sans-serif;
}`
	os.WriteFile(filepath.Join(dir, "src", "app", "globals.css"), []byte(globalCSS), 0600)

	info := ScanStyles(dir)

	if info.CSSFramework != "Tailwind CSS" {
		t.Errorf("CSSFramework = %q, want Tailwind CSS", info.CSSFramework)
	}
	if info.StylingApproach != "utility-first" {
		t.Errorf("StylingApproach = %q, want utility-first", info.StylingApproach)
	}
	if len(info.ThemeColors) == 0 {
		t.Error("expected ThemeColors to be populated")
	}
	if _, ok := info.ThemeColors["--primary"]; !ok {
		t.Error("expected --primary in ThemeColors")
	}
	if len(info.FontFamilies) == 0 {
		t.Error("expected FontFamilies to be populated")
	}
}

func TestScanStyles_EmotionProject(t *testing.T) {
	dir := t.TempDir()

	pkgJSON := `{"dependencies":{"@emotion/react":"^11.0.0","@emotion/styled":"^11.0.0","react":"^18.0.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0600)

	info := ScanStyles(dir)

	if info.CSSFramework != "Emotion" {
		t.Errorf("CSSFramework = %q, want Emotion", info.CSSFramework)
	}
	if info.StylingApproach != "css-in-js" {
		t.Errorf("StylingApproach = %q, want css-in-js", info.StylingApproach)
	}
}

func TestScanStyles_NoFramework(t *testing.T) {
	dir := t.TempDir()

	info := ScanStyles(dir)

	if info.CSSFramework != "" {
		t.Errorf("CSSFramework = %q, want empty", info.CSSFramework)
	}
	if info.StylingApproach != "traditional" {
		t.Errorf("StylingApproach = %q, want traditional", info.StylingApproach)
	}
}

func TestRenderMockupPrompt_WithStyle(t *testing.T) {
	ctx := &MockupPromptContext{
		SpecName:        "042-registration",
		SpecPath:        "specledger/042-registration/spec.md",
		SpecTitle:       "User Registration",
		Framework:       FrameworkReact,
		Format:          MockupFormatHTML,
		OutputDir:       "specledger/042-registration/mockup/",
		HasDesignSystem: false,
		HasStyle:        true,
		Style: &StyleInfo{
			CSSFramework:    "Tailwind CSS",
			StylingApproach: "utility-first",
			FontFamilies:    []string{"Inter, system-ui, sans-serif"},
			ThemeColors: map[string]string{
				"--primary":    "#2563eb",
				"--background": "#ffffff",
			},
		},
	}

	result, err := RenderMockupPrompt(ctx)
	if err != nil {
		t.Fatal(err)
	}

	checks := []string{
		"Tailwind CSS",
		"utility-first",
		"--primary",
		"#2563eb",
		"Inter",
		"Tailwind CDN",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("prompt missing expected content: %q", check)
		}
	}
}
