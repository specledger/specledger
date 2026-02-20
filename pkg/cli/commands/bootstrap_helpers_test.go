package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupAgentConfig_ClaudeCode(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup Claude Code agent config
	if err := setupAgentConfig(tmpDir, "claude-code"); err != nil {
		t.Fatalf("setupAgentConfig failed: %v", err)
	}

	// Verify .claude directory was created
	claudeDir := filepath.Join(tmpDir, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		t.Error(".claude directory was not created")
	}

	// Verify settings.json exists
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error(".claude/settings.json was not created")
	}

	// Verify settings.json content
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(content, &settings); err != nil {
		t.Fatalf("invalid JSON in settings.json: %v", err)
	}

	// Check saveTranscripts is true
	if saveTranscripts, ok := settings["saveTranscripts"].(bool); !ok || !saveTranscripts {
		t.Error("saveTranscripts should be true")
	}

	// Check hooks exist
	if _, ok := settings["hooks"]; !ok {
		t.Error("hooks configuration missing")
	}
}

func TestSetupAgentConfig_OpenCode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := setupAgentConfig(tmpDir, "opencode"); err != nil {
		t.Fatalf("setupAgentConfig failed: %v", err)
	}

	// Verify .opencode directory was created
	opencodeDir := filepath.Join(tmpDir, ".opencode")
	if _, err := os.Stat(opencodeDir); os.IsNotExist(err) {
		t.Error(".opencode directory was not created")
	}

	// Verify opencode.json exists
	configPath := filepath.Join(tmpDir, "opencode.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("opencode.json was not created")
	}
}

func TestSetupAgentConfig_None(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Should not error and should not create any directories
	if err := setupAgentConfig(tmpDir, "none"); err != nil {
		t.Fatalf("setupAgentConfig failed: %v", err)
	}

	// Verify no agent directories were created
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}

	if len(entries) > 0 {
		t.Errorf("expected no files/directories for 'none' agent, got %d", len(entries))
	}
}

func TestSetupAgentConfig_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Empty agent ID should not error
	if err := setupAgentConfig(tmpDir, ""); err != nil {
		t.Fatalf("setupAgentConfig failed: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}

	if len(entries) > 0 {
		t.Errorf("expected no files/directories for empty agent, got %d", len(entries))
	}
}
