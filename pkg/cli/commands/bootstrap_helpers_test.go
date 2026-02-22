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

	// Check hooks exist with PostToolUse configuration
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("hooks configuration missing")
	}

	// Verify PostToolUse hook exists
	postToolUse, ok := hooks["PostToolUse"]
	if !ok {
		t.Error("PostToolUse hook missing")
	}

	// Verify it's an array with at least one entry
	hookArray, ok := postToolUse.([]interface{})
	if !ok || len(hookArray) == 0 {
		t.Error("PostToolUse should have at least one hook configuration")
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

func TestApplyTemplateFiles_GeneralPurpose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// general-purpose should not copy any additional files
	count, err := applyTemplateFiles(tmpDir, "general-purpose")
	if err != nil {
		t.Fatalf("applyTemplateFiles failed: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 files for general-purpose, got %d", count)
	}
}

func TestApplyTemplateFiles_BatchData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// batch-data should copy template files
	count, err := applyTemplateFiles(tmpDir, "batch-data")
	if err != nil {
		t.Fatalf("applyTemplateFiles failed: %v", err)
	}

	if count == 0 {
		t.Error("expected files to be copied for batch-data template")
	}

	// Verify key directories exist
	expectedDirs := []string{"workflows", "cmd/worker", "cmd/starter"}
	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected directory missing: %s", dir)
		}
	}

	// Verify go.mod was created (transformed from go.mod.template)
	goModPath := filepath.Join(tmpDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Error("go.mod should be created from go.mod.template")
	}
}

func TestApplyTemplateFiles_FullStack(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// full-stack should copy template files
	count, err := applyTemplateFiles(tmpDir, "full-stack")
	if err != nil {
		t.Fatalf("applyTemplateFiles failed: %v", err)
	}

	if count == 0 {
		t.Error("expected files to be copied for full-stack template")
	}

	// Verify key directories exist
	expectedDirs := []string{"frontend", "frontend/src"}
	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected directory missing: %s", dir)
		}
	}
}

func TestApplyTemplateFiles_Unknown(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Unknown template should return error
	_, err = applyTemplateFiles(tmpDir, "nonexistent-template")
	if err == nil {
		t.Error("expected error for unknown template")
	}
}
