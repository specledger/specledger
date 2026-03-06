package playbooks

import (
	"path"
	"strings"
	"testing"
)

// TestLoadManifestPathForwardSlash verifies LoadManifest works on all platforms.
// Regression guard: if filepath.Join is reintroduced, Windows produces
// "templates\manifest.yaml" which embed.FS cannot find.
func TestLoadManifestPathForwardSlash(t *testing.T) {
	manifest, err := LoadManifest("templates")
	if err != nil {
		t.Fatalf("LoadManifest failed (embed.FS path separator bug?): %v", err)
	}
	if len(manifest.Playbooks) == 0 {
		t.Error("expected at least one playbook in manifest")
	}
}

// TestManifestPathNeverHasBackslash verifies path.Join always produces forward slashes.
func TestManifestPathNeverHasBackslash(t *testing.T) {
	result := path.Join("templates", "manifest.yaml")
	if result != "templates/manifest.yaml" {
		t.Errorf("expected templates/manifest.yaml, got %q", result)
	}
	if strings.Contains(result, "\\") {
		t.Errorf("path contains backslash — will break embed.FS on Windows: %q", result)
	}
}
