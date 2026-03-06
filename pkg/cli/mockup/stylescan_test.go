package mockup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanStyles_TailwindProject(t *testing.T) {
	dir := t.TempDir()

	// Create package.json with tailwindcss
	pkgJSON := `{"dependencies":{"tailwindcss":"^3.4.0","react":"^18.0.0"}}`
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0600)

	// Create tailwind config
	_ = os.WriteFile(filepath.Join(dir, "tailwind.config.js"), []byte("module.exports = {}"), 0600)

	// Create globals.css with CSS variables
	_ = os.MkdirAll(filepath.Join(dir, "src", "app"), 0755)
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
	_ = os.WriteFile(filepath.Join(dir, "src", "app", "globals.css"), []byte(globalCSS), 0600)

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

func TestScanStyles_TailwindWithPostCSS(t *testing.T) {
	dir := t.TempDir()

	// Both tailwind and postcss config exist — tailwind should win
	_ = os.WriteFile(filepath.Join(dir, "tailwind.config.ts"), []byte("export default {}"), 0600)
	_ = os.WriteFile(filepath.Join(dir, "postcss.config.js"), []byte("module.exports = {}"), 0600)

	info := ScanStyles(dir)

	if info.CSSFramework != "Tailwind CSS" {
		t.Errorf("CSSFramework = %q, want Tailwind CSS (postcss should not override)", info.CSSFramework)
	}
}

func TestScanTailwindConfig_ExtractsColors(t *testing.T) {
	dir := t.TempDir()

	config := `import type { Config } from 'tailwindcss'
export default {
  theme: {
    extend: {
      colors: {
        brand: {
          50:  '#eef2ff',
          500: '#6366f1',
          900: '#312e81',
        },
        surface: {
          0:   '#ffffff',
          100: '#f3f4f6',
        },
      },
    },
  },
} satisfies Config`
	_ = os.WriteFile(filepath.Join(dir, "tailwind.config.ts"), []byte(config), 0600)

	info := &StyleInfo{ThemeColors: make(map[string]string)}
	scanTailwindConfig(dir, info)

	cases := map[string]string{
		"brand-50":    "#eef2ff",
		"brand-500":   "#6366f1",
		"brand-900":   "#312e81",
		"surface-0":   "#ffffff",
		"surface-100": "#f3f4f6",
	}
	for token, want := range cases {
		if got, ok := info.ThemeColors[token]; !ok {
			t.Errorf("missing color token %q", token)
		} else if got != want {
			t.Errorf("ThemeColors[%q] = %q, want %q", token, got, want)
		}
	}
}

func TestScanStyles_EmotionProject(t *testing.T) {
	dir := t.TempDir()

	pkgJSON := `{"dependencies":{"@emotion/react":"^11.0.0","@emotion/styled":"^11.0.0","react":"^18.0.0"}}`
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0600)

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

func TestScanStyles_TailwindV4(t *testing.T) {
	dir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(dir, "src", "app"), 0755)
	globalCSS := `@import "tailwindcss";

@theme {
  --color-primary: #3b82f6;
  --color-secondary: #64748b;
  --font-sans: "Inter", sans-serif;
}
`
	_ = os.WriteFile(filepath.Join(dir, "src", "app", "globals.css"), []byte(globalCSS), 0600)

	info := ScanStyles(dir)

	if info.CSSFramework != "Tailwind CSS v4" {
		t.Errorf("CSSFramework = %q, want Tailwind CSS v4", info.CSSFramework)
	}
	if info.StylingApproach != "utility-first" {
		t.Errorf("StylingApproach = %q, want utility-first", info.StylingApproach)
	}
	if got, ok := info.ThemeColors["--color-primary"]; !ok || got != "#3b82f6" {
		t.Errorf("ThemeColors[--color-primary] = %q, %v; want #3b82f6", got, ok)
	}
}

func TestScanStyles_UnoCSS(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "uno.config.ts"), []byte("export default {}"), 0600)

	info := ScanStyles(dir)

	if info.CSSFramework != "UnoCSS" {
		t.Errorf("CSSFramework = %q, want UnoCSS", info.CSSFramework)
	}
}

func TestScanStyles_PandaCSS(t *testing.T) {
	dir := t.TempDir()

	pkgJSON := `{"devDependencies":{"@pandacss/dev":"^0.30.0"}}`
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0600)
	_ = os.WriteFile(filepath.Join(dir, "panda.config.ts"), []byte("export default {}"), 0600)

	info := ScanStyles(dir)

	if info.CSSFramework != "Panda CSS" {
		t.Errorf("CSSFramework = %q, want Panda CSS", info.CSSFramework)
	}
	if info.StylingApproach != "utility-first" {
		t.Errorf("StylingApproach = %q, want utility-first", info.StylingApproach)
	}
}

func TestScanStyles_ComponentLibs(t *testing.T) {
	dir := t.TempDir()

	pkgJSON := `{"dependencies":{"react":"^18.0.0","@radix-ui/react-dialog":"^1.0.0","@radix-ui/react-popover":"^1.0.0","lucide-react":"^0.300.0","tailwindcss":"^3.4.0"}}`
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0600)

	info := ScanStyles(dir)

	found := map[string]bool{}
	for _, lib := range info.ComponentLibs {
		found[lib] = true
	}
	if !found["Radix UI"] {
		t.Errorf("expected Radix UI in ComponentLibs, got %v", info.ComponentLibs)
	}
	if !found["Lucide Icons"] {
		t.Errorf("expected Lucide Icons in ComponentLibs, got %v", info.ComponentLibs)
	}
}

func TestScanStyles_ShadcnByComponentsJSON(t *testing.T) {
	dir := t.TempDir()

	pkgJSON := `{"dependencies":{"react":"^18.0.0","tailwindcss":"^3.4.0"}}`
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0600)
	_ = os.WriteFile(filepath.Join(dir, "components.json"), []byte(`{"style":"default"}`), 0600)

	info := ScanStyles(dir)

	found := false
	for _, lib := range info.ComponentLibs {
		if lib == "shadcn/ui" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected shadcn/ui in ComponentLibs, got %v", info.ComponentLibs)
	}
}
