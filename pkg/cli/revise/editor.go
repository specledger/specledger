package revise

import (
	"fmt"
	"os"
	"os/exec"
)

// EditPrompt writes prompt to a temp file, opens it in the user's editor,
// and returns the (possibly modified) content after the editor exits.
func EditPrompt(prompt string) (string, error) {
	tmpFile, err := os.CreateTemp("", "sl-revise-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(prompt); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write prompt to temp file: %w", err)
	}
	tmpFile.Close()

	editor := detectEditor()
	if editor == "" {
		return "", fmt.Errorf("no editor found: set $EDITOR or $VISUAL environment variable")
	}

	// #nosec G204 — editor is from $EDITOR/$VISUAL env var, user-controlled
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited prompt: %w", err)
	}

	return string(content), nil
}

// detectEditor returns the user's preferred editor from the environment.
func detectEditor() string {
	for _, env := range []string{"EDITOR", "VISUAL"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	// Common fallbacks
	for _, candidate := range []string{"vi", "nano", "vim"} {
		if _, err := exec.LookPath(candidate); err == nil {
			return candidate
		}
	}
	return ""
}
