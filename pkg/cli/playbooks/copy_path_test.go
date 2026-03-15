package playbooks

import (
	"testing"
)

// TestEmbedFSPathSeparator verifies that embed.FS paths must use forward slashes.
// Regression guard: if filepath.Join is reintroduced in copy.go, Windows will build
// srcPath with backslashes and all playbook files will be silently skipped.
func TestEmbedFSPathSeparator(t *testing.T) {
	// Forward slash must resolve in embed.FS
	if !Exists("templates/specledger") {
		t.Error("templates/specledger should exist in embed.FS (forward slash required)")
	}

	// Backslash must NOT resolve in embed.FS (catches regression to filepath.Join)
	if Exists(`templates\specledger`) {
		t.Error(`templates\specledger should NOT exist in embed.FS — backslash paths are invalid`)
	}
}
