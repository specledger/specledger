package prompt

import (
	"os"
	"testing"
)

func TestDetectEditor_EnvVar(t *testing.T) {
	original := os.Getenv("EDITOR")
	defer func() {
		if original != "" {
			os.Setenv("EDITOR", original)
		} else {
			os.Unsetenv("EDITOR")
		}
	}()

	os.Setenv("EDITOR", "test-editor-abc")
	got := DetectEditor()
	if got != "test-editor-abc" {
		t.Errorf("DetectEditor() = %q, want %q", got, "test-editor-abc")
	}
}

func TestDetectEditor_VisualFallback(t *testing.T) {
	originalEditor := os.Getenv("EDITOR")
	originalVisual := os.Getenv("VISUAL")
	defer func() {
		if originalEditor != "" {
			os.Setenv("EDITOR", originalEditor)
		} else {
			os.Unsetenv("EDITOR")
		}
		if originalVisual != "" {
			os.Setenv("VISUAL", originalVisual)
		} else {
			os.Unsetenv("VISUAL")
		}
	}()

	os.Unsetenv("EDITOR")
	os.Setenv("VISUAL", "test-visual-editor")
	got := DetectEditor()
	if got != "test-visual-editor" {
		t.Errorf("DetectEditor() = %q, want %q", got, "test-visual-editor")
	}
}

func TestDetectEditor_FallsBackToSystem(t *testing.T) {
	originalEditor := os.Getenv("EDITOR")
	originalVisual := os.Getenv("VISUAL")
	defer func() {
		if originalEditor != "" {
			os.Setenv("EDITOR", originalEditor)
		} else {
			os.Unsetenv("EDITOR")
		}
		if originalVisual != "" {
			os.Setenv("VISUAL", originalVisual)
		} else {
			os.Unsetenv("VISUAL")
		}
	}()

	os.Unsetenv("EDITOR")
	os.Unsetenv("VISUAL")

	got := DetectEditor()
	// Should return one of the fallback editors or empty string
	valid := map[string]bool{"vi": true, "nano": true, "vim": true, "": true}
	if !valid[got] {
		t.Errorf("DetectEditor() = %q, want one of vi/nano/vim/empty", got)
	}
}
