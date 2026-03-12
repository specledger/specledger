package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	beginMarker = "# BEGIN SPECLEDGER PUSH HOOK"
	endMarker   = "# END SPECLEDGER PUSH HOOK"
	shebang     = "#!/bin/sh"
)

var hookBlock = strings.Join([]string{
	beginMarker,
	"# Installed by: sl hook install",
	"# Do not edit this block manually. Use 'sl hook uninstall' to remove.",
	`sl hook execute --event pre-push "$@" 2>/dev/null || true`,
	endMarker,
}, "\n")

// prePushHookPath returns the path to the pre-push hook file.
func prePushHookPath(gitDir string) string {
	return filepath.Join(gitDir, "hooks", "pre-push")
}

// InstallPushHook installs the SpecLedger pre-push hook into the git repository.
// If force is true, an existing SpecLedger block is replaced.
// Existing non-SpecLedger hook content is preserved.
func InstallPushHook(gitDir string, force bool) error {
	hookPath := prePushHookPath(gitDir)

	// Ensure hooks directory exists
	hooksDir := filepath.Dir(hookPath)
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	existing, err := os.ReadFile(hookPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read existing hook: %w", err)
	}

	content := string(existing)

	if strings.Contains(content, beginMarker) {
		if !force {
			return nil // Already installed, not forced
		}
		// Remove existing block
		content = removeMarkerBlock(content)
	}

	// Build new content
	var result strings.Builder
	if content == "" {
		result.WriteString(shebang + "\n")
	} else {
		result.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			result.WriteString("\n")
		}
	}
	result.WriteString(hookBlock + "\n")

	if err := os.WriteFile(hookPath, []byte(result.String()), 0755); err != nil {
		return fmt.Errorf("failed to write hook file: %w", err)
	}

	return nil
}

// UninstallPushHook removes the SpecLedger pre-push hook block.
// If the file becomes empty (or only shebang), it is deleted.
func UninstallPushHook(gitDir string) error {
	hookPath := prePushHookPath(gitDir)

	existing, err := os.ReadFile(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to uninstall
		}
		return fmt.Errorf("failed to read hook file: %w", err)
	}

	content := string(existing)
	if !strings.Contains(content, beginMarker) {
		return nil // Not installed
	}

	content = removeMarkerBlock(content)
	content = strings.TrimSpace(content)

	// If only shebang or empty remains, delete the file
	if content == "" || content == shebang {
		return os.Remove(hookPath)
	}

	return os.WriteFile(hookPath, []byte(content+"\n"), 0755)
}

// HasPushHook returns true if the SpecLedger push hook is installed.
func HasPushHook(gitDir string) bool {
	hookPath := prePushHookPath(gitDir)
	data, err := os.ReadFile(hookPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), beginMarker)
}

// removeMarkerBlock removes everything between BEGIN and END markers (inclusive).
func removeMarkerBlock(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlock := false

	for _, line := range lines {
		if strings.TrimSpace(line) == beginMarker {
			inBlock = true
			continue
		}
		if strings.TrimSpace(line) == endMarker {
			inBlock = false
			continue
		}
		if !inBlock {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
