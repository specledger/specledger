package mockup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanAppStructure_NextJSAppRouter(t *testing.T) {
	dir := t.TempDir()

	// Create Next.js App Router structure
	for _, p := range []string{
		"app/layout.tsx",
		"app/page.tsx",
		"app/about/page.tsx",
		"app/dashboard/layout.tsx",
		"app/dashboard/page.tsx",
		"app/dashboard/settings/page.tsx",
		"app/api/users/route.ts",
		"app/api/auth/route.ts",
	} {
		full := filepath.Join(dir, p)
		_ = os.MkdirAll(filepath.Dir(full), 0755)
		_ = os.WriteFile(full, []byte("export default function() {}"), 0600)
	}

	app := ScanAppStructure(dir, FrameworkNextJS)
	if app == nil {
		t.Fatal("expected AppStructure, got nil")
	}
	if app.Router != "app-router" {
		t.Errorf("Router = %q, want app-router", app.Router)
	}
	if len(app.Layouts) < 2 {
		t.Errorf("expected at least 2 layouts, got %d: %v", len(app.Layouts), app.Layouts)
	}
}

func TestScanAppStructure_NextJSPagesRouter(t *testing.T) {
	dir := t.TempDir()

	// Create Next.js Pages Router structure (no app/ directory)
	for _, p := range []string{
		"pages/_app.tsx",
		"pages/_document.tsx",
		"pages/index.tsx",
		"pages/about.tsx",
		"pages/api/users.ts",
	} {
		full := filepath.Join(dir, p)
		_ = os.MkdirAll(filepath.Dir(full), 0755)
		_ = os.WriteFile(full, []byte("export default function() {}"), 0600)
	}

	app := ScanAppStructure(dir, FrameworkNextJS)
	if app == nil {
		t.Fatal("expected AppStructure, got nil")
	}
	if app.Router != "pages-router" {
		t.Errorf("Router = %q, want pages-router", app.Router)
	}
	if len(app.Layouts) < 1 {
		t.Errorf("expected at least 1 layout (_app.tsx), got %d: %v", len(app.Layouts), app.Layouts)
	}
}

func TestScanAppStructure_SvelteKit(t *testing.T) {
	dir := t.TempDir()

	for _, p := range []string{
		"src/routes/+layout.svelte",
		"src/routes/+page.svelte",
		"src/routes/about/+page.svelte",
		"src/routes/dashboard/+layout.svelte",
		"src/routes/dashboard/+page.svelte",
	} {
		full := filepath.Join(dir, p)
		_ = os.MkdirAll(filepath.Dir(full), 0755)
		_ = os.WriteFile(full, []byte("<script>...</script>"), 0600)
	}

	app := ScanAppStructure(dir, FrameworkSvelteKit)
	if app == nil {
		t.Fatal("expected AppStructure, got nil")
	}
	if app.Router != "file-based" {
		t.Errorf("Router = %q, want file-based", app.Router)
	}
	if len(app.Layouts) < 2 {
		t.Errorf("expected at least 2 layouts, got %d: %v", len(app.Layouts), app.Layouts)
	}
}

func TestScanAppStructure_Nuxt(t *testing.T) {
	dir := t.TempDir()

	for _, p := range []string{
		"layouts/default.vue",
		"pages/index.vue",
		"pages/about.vue",
		"pages/users/[id].vue",
		"server/api/hello.ts",
	} {
		full := filepath.Join(dir, p)
		_ = os.MkdirAll(filepath.Dir(full), 0755)
		_ = os.WriteFile(full, []byte("<template></template>"), 0600)
	}

	app := ScanAppStructure(dir, FrameworkNuxt)
	if app == nil {
		t.Fatal("expected AppStructure, got nil")
	}
	if len(app.Layouts) < 1 {
		t.Errorf("expected at least 1 layout, got %d", len(app.Layouts))
	}
}

func TestScanAppStructure_Astro(t *testing.T) {
	dir := t.TempDir()

	for _, p := range []string{
		"src/layouts/Base.astro",
		"src/pages/index.astro",
		"src/pages/about.astro",
		"src/pages/blog/[slug].astro",
	} {
		full := filepath.Join(dir, p)
		_ = os.MkdirAll(filepath.Dir(full), 0755)
		_ = os.WriteFile(full, []byte("---\n---\n<html></html>"), 0600)
	}

	app := ScanAppStructure(dir, FrameworkAstro)
	if app == nil {
		t.Fatal("expected AppStructure, got nil")
	}
	if len(app.Layouts) < 1 {
		t.Errorf("expected at least 1 layout, got %d", len(app.Layouts))
	}
}

func TestScanAppStructure_NoStructure(t *testing.T) {
	dir := t.TempDir()
	// Empty project
	app := ScanAppStructure(dir, FrameworkReact)
	if app != nil {
		t.Errorf("expected nil for empty project, got %+v", app)
	}
}

func TestScanAppStructure_UnknownFramework(t *testing.T) {
	dir := t.TempDir()
	app := ScanAppStructure(dir, FrameworkUnknown)
	if app != nil {
		t.Errorf("expected nil for unknown framework, got %+v", app)
	}
}

func TestScanAppStructure_DesignSystemRoundTrip(t *testing.T) {
	dir := t.TempDir()

	// Create a minimal Next.js structure
	for _, p := range []string{"app/layout.tsx", "app/page.tsx"} {
		full := filepath.Join(dir, p)
		_ = os.MkdirAll(filepath.Dir(full), 0755)
		_ = os.WriteFile(full, []byte("export default function() {}"), 0600)
	}

	app := ScanAppStructure(dir, FrameworkNextJS)
	dsPath := filepath.Join(dir, "design-system.md")
	ds := &DesignSystem{
		Version:      1,
		Framework:    FrameworkNextJS,
		AppStructure: app,
	}

	err := WriteDesignSystem(dsPath, ds)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadDesignSystem(dsPath)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.AppStructure == nil {
		t.Fatal("AppStructure is nil after round-trip")
	}
	if loaded.AppStructure.Router != "app-router" {
		t.Errorf("Router = %q, want app-router", loaded.AppStructure.Router)
	}
	if len(loaded.AppStructure.Layouts) != len(app.Layouts) {
		t.Errorf("Layouts count mismatch: got %d, want %d", len(loaded.AppStructure.Layouts), len(app.Layouts))
	}
}
