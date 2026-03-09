package embedded

import (
	"io/fs"
	"testing"
)

func TestTemplatesFSPopulated(t *testing.T) {
	// Test that TemplatesFS is populated with actual templates
	entries, err := TemplatesFS.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read TemplatesFS: %v", err)
	}

	if len(entries) == 0 {
		t.Error("TemplatesFS is empty - templates were not embedded")
	}

	// Check for expected top-level entries
	expectedEntries := map[string]bool{
		"templates":     true,  // directory
		"manifest.yaml": false, // this is a file
	}

	for _, entry := range entries {
		delete(expectedEntries, entry.Name())
	}

	for name := range expectedEntries {
		if expectedEntries[name] {
			t.Errorf("Expected %s to exist in TemplatesFS but it's missing", name)
		}
	}
}

func TestTemplatesFSHasSpecledgerPlaybook(t *testing.T) {
	// Check that specledger playbook template exists
	entries, err := TemplatesFS.ReadDir("templates")
	if err != nil {
		t.Fatalf("Failed to read templates directory: %v", err)
	}

	if len(entries) == 0 {
		t.Error("templates directory is empty in TemplatesFS")
	}

	// Look for specledger directory
	foundSpecledger := false
	for _, entry := range entries {
		if entry.Name() == "specledger" && entry.IsDir() {
			foundSpecledger = true
			break
		}
	}

	if !foundSpecledger {
		t.Error("Expected templates/specledger to exist but it's missing")
	}
}

func TestTemplatesFSHasManifest(t *testing.T) {
	// Check that manifest.yaml exists
	_, err := TemplatesFS.ReadFile("templates/manifest.yaml")
	if err != nil {
		t.Fatalf("Failed to read templates/manifest.yaml: %v", err)
	}

	// Also test reading from root (might be available)
	content, err := TemplatesFS.ReadFile("manifest.yaml")
	if err == nil && len(content) > 0 {
		t.Log("manifest.yaml is also available at root level")
	}
}

func TestTemplatesFSHasSpecledgerFiles(t *testing.T) {
	// Check that specledger playbook has expected files
	// Commands and skills are now directly under templates/specledger/
	expectedFiles := []string{
		"templates/specledger/init.sh",
		"templates/specledger/.specledger/FORK.md",
		"templates/specledger/commands/specledger.specify.md",
		"templates/specledger/skills/sl-audit/skill.md",
	}

	for _, file := range expectedFiles {
		_, err := TemplatesFS.ReadFile(file)
		if err != nil {
			t.Errorf("Expected %s to exist in TemplatesFS but got error: %v", file, err)
		}
	}
}

func TestTemplatesFSListing(t *testing.T) {
	// Test directory listing to see structure
	rootEntries, err := fs.ReadDir(TemplatesFS, ".")
	if err != nil {
		t.Fatalf("Failed to list root: %v", err)
	}

	t.Logf("Root entries: %d", len(rootEntries))
	for _, entry := range rootEntries {
		t.Logf("  - %s (dir: %v)", entry.Name(), entry.IsDir())
	}

	// List templates directory
	templatesEntries, err := fs.ReadDir(TemplatesFS, "templates")
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	t.Logf("Templates entries: %d", len(templatesEntries))
	for _, entry := range templatesEntries {
		t.Logf("  - %s (dir: %v)", entry.Name(), entry.IsDir())
	}
}

func TestTemplatesFSStructure(t *testing.T) {
	// Test that all expected directories exist in TemplatesFS
	// Commands and skills are now directly under templates/specledger/
	expectedDirs := []string{
		"templates/specledger/commands",
		"templates/specledger/skills",
		"templates/specledger/.specledger/templates",
	}

	for _, dir := range expectedDirs {
		entries, err := TemplatesFS.ReadDir(dir)
		if err != nil {
			t.Errorf("Expected directory %s to exist: %v", dir, err)
			continue
		}
		t.Logf("Directory %s has %d entries", dir, len(entries))
	}
}
