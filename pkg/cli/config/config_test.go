package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSchemaAgentDefault(t *testing.T) {
	def, err := LookupKey("agent.default")
	if err != nil {
		t.Fatalf("LookupKey(agent.default) failed: %v", err)
	}
	if def.Type != KeyTypeString {
		t.Errorf("expected KeyTypeString, got %s", def.Type)
	}
	if def.Default != "claude" {
		t.Errorf("expected default 'claude', got %v", def.Default)
	}
}

func TestSchemaPerAgentAPIKey(t *testing.T) {
	tests := []struct {
		key       string
		wantError bool
	}{
		{"agent.claude.api_key", false},
		{"agent.opencode.api_key", false},
		{"agent.github-copilot.api_key", false},
		{"agent.codex.api_key", false},
		{"agent.unknown.api_key", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			def, err := LookupKey(tt.key)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tt.key)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if def.Type != KeyTypeString {
					t.Errorf("expected KeyTypeString, got %s", def.Type)
				}
				if !def.Sensitive {
					t.Errorf("expected api_key to be sensitive")
				}
			}
		})
	}
}

func TestSchemaPerAgentBaseURL(t *testing.T) {
	tests := []struct {
		key       string
		wantError bool
	}{
		{"agent.claude.base_url", false},
		{"agent.opencode.base_url", false},
		{"agent.codex.base_url", false},
		{"agent.unknown.base_url", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			def, err := LookupKey(tt.key)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tt.key)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if def.Type != KeyTypeString {
					t.Errorf("expected KeyTypeString, got %s", def.Type)
				}
			}
		})
	}
}

func TestSchemaPerAgentModel(t *testing.T) {
	tests := []struct {
		key       string
		wantError bool
	}{
		{"agent.claude.model", false},
		{"agent.opencode.model", false},
		{"agent.codex.model", false},
		{"agent.unknown.model", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			def, err := LookupKey(tt.key)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tt.key)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if def.Type != KeyTypeString {
					t.Errorf("expected KeyTypeString, got %s", def.Type)
				}
			}
		})
	}
}

func TestSchemaPerAgentArguments(t *testing.T) {
	tests := []struct {
		key       string
		wantError bool
	}{
		{"agent.claude.arguments", false},
		{"agent.opencode.arguments", false},
		{"agent.github-copilot.arguments", false},
		{"agent.codex.arguments", false},
		{"agent.unknown.arguments", true},
		{"agent.CLAUDE.arguments", false}, // case-insensitive agent name
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			def, err := LookupKey(tt.key)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tt.key)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if def.Type != KeyTypeString {
					t.Errorf("expected KeyTypeString, got %s", def.Type)
				}
			}
		})
	}
}

func TestSchemaPerAgentEnv(t *testing.T) {
	tests := []struct {
		key       string
		wantError bool
	}{
		{"agent.claude.env", false},
		{"agent.opencode.env", false},
		{"agent.github-copilot.env", false},
		{"agent.codex.env", false},
		{"agent.unknown.env", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			def, err := LookupKey(tt.key)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tt.key)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if def.Type != KeyTypeStringMap {
					t.Errorf("expected KeyTypeStringMap, got %s", def.Type)
				}
			}
		})
	}
}

func TestSchemaClaudeModelAliases(t *testing.T) {
	tests := []struct {
		key       string
		wantError bool
	}{
		{"agent.claude.model_aliases.sonnet", false},
		{"agent.claude.model_aliases.opus", false},
		{"agent.claude.model_aliases.haiku", false},
		{"agent.claude.model_aliases.unknown", true},
		{"agent.opencode.model_aliases.sonnet", true}, // only claude has model_aliases
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			def, err := LookupKey(tt.key)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tt.key)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if def.Type != KeyTypeString {
					t.Errorf("expected KeyTypeString, got %s", def.Type)
				}
			}
		})
	}
}

func TestIsValidKeyPerAgent(t *testing.T) {
	tests := []struct {
		key   string
		valid bool
	}{
		// Universal keys
		{"agent.default", true},
		{"agent.env.CUSTOM", true}, // legacy

		// Per-agent keys
		{"agent.claude.api_key", true},
		{"agent.claude.base_url", true},
		{"agent.claude.model", true},
		{"agent.claude.arguments", true},
		{"agent.claude.env", true},
		{"agent.claude.env.CUSTOM_VAR", true},

		{"agent.opencode.api_key", true},
		{"agent.opencode.arguments", true},

		{"agent.github-copilot.api_key", true},
		{"agent.codex.arguments", true},

		// Claude-specific
		{"agent.claude.model_aliases.sonnet", true},
		{"agent.claude.model_aliases.opus", true},
		{"agent.claude.model_aliases.haiku", true},

		// Invalid
		{"agent.unknown.arguments", false},
		{"agent.unknown.env", false},
		{"agent.claude.model_aliases.unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			valid := IsValidKey(tt.key)
			if valid != tt.valid {
				t.Errorf("IsValidKey(%q) = %v, want %v", tt.key, valid, tt.valid)
			}
		})
	}
}

func TestResolveAgentSettings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Isolate from user's real config by setting HOME to temp dir
	t.Setenv("HOME", tmpDir)

	// Create specledger directory
	slDir := filepath.Join(tmpDir, "specledger")
	if err := os.MkdirAll(slDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create project config with claude settings
	projectConfig := `agents:
  claude:
    api_key: project-key
    model: claude-sonnet-4-20250514
    arguments:
      - --verbose
    env:
      CUSTOM_VAR: project-value
`
	if err := os.WriteFile(filepath.Join(slDir, "specledger.yaml"), []byte(projectConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original dir and change to temp dir
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Resolve settings
	settings := ResolveAgentSettings("claude")
	if settings == nil {
		t.Fatal("expected settings, got nil")
	}

	if settings.APIKey != "project-key" {
		t.Errorf("expected APIKey 'project-key', got %q", settings.APIKey)
	}
	if settings.Model != "claude-sonnet-4-20250514" {
		t.Errorf("expected Model 'claude-sonnet-4-20250514', got %q", settings.Model)
	}
	if len(settings.Arguments) != 1 || settings.Arguments[0] != "--verbose" {
		t.Errorf("expected Arguments [--verbose], got %v", settings.Arguments)
	}
	if settings.EnvVars["CUSTOM_VAR"] != "project-value" {
		t.Errorf("expected EnvVars[CUSTOM_VAR] 'project-value', got %q", settings.EnvVars["CUSTOM_VAR"])
	}
}

func TestResolveAgentSettingsPrecedence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Isolate from user's real config by setting HOME to temp dir
	t.Setenv("HOME", tmpDir)

	slDir := filepath.Join(tmpDir, "specledger")
	if err := os.MkdirAll(slDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Project config (lower precedence)
	projectConfig := `agents:
  claude:
    api_key: project-key
    model: project-model
`
	if err := os.WriteFile(filepath.Join(slDir, "specledger.yaml"), []byte(projectConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Personal config (higher precedence)
	personalConfig := `agents:
  claude:
    api_key: personal-key
`
	if err := os.WriteFile(filepath.Join(slDir, "specledger.local.yaml"), []byte(personalConfig), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	settings := ResolveAgentSettings("claude")
	if settings == nil {
		t.Fatal("expected settings, got nil")
	}

	// Personal should override project for api_key
	if settings.APIKey != "personal-key" {
		t.Errorf("expected APIKey 'personal-key' (from personal), got %q", settings.APIKey)
	}
	// Project value should still be used for model (not set in personal)
	if settings.Model != "project-model" {
		t.Errorf("expected Model 'project-model' (from project), got %q", settings.Model)
	}
}

func TestResolveAgentSettingsClaudeModelAliases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Isolate from user's real config by setting HOME to temp dir
	t.Setenv("HOME", tmpDir)

	slDir := filepath.Join(tmpDir, "specledger")
	if err := os.MkdirAll(slDir, 0755); err != nil {
		t.Fatal(err)
	}

	projectConfig := `agents:
  claude:
    model_aliases:
      sonnet: claude-sonnet-4-20250514
      opus: claude-opus-4-20250514
      haiku: claude-haiku-3-5-20241022
`
	if err := os.WriteFile(filepath.Join(slDir, "specledger.yaml"), []byte(projectConfig), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	settings := ResolveAgentSettings("claude")
	if settings == nil {
		t.Fatal("expected settings, got nil")
	}

	if settings.ModelAliases["sonnet"] != "claude-sonnet-4-20250514" {
		t.Errorf("expected sonnet alias, got %q", settings.ModelAliases["sonnet"])
	}
	if settings.ModelAliases["opus"] != "claude-opus-4-20250514" {
		t.Errorf("expected opus alias, got %q", settings.ModelAliases["opus"])
	}
	if settings.ModelAliases["haiku"] != "claude-haiku-3-5-20241022" {
		t.Errorf("expected haiku alias, got %q", settings.ModelAliases["haiku"])
	}
}

func TestProfileCRUD(t *testing.T) {
	cfg := DefaultConfig()

	err := cfg.CreateProfile("work", &AgentConfig{Model: "work-model"})
	if err != nil {
		t.Fatalf("CreateProfile failed: %v", err)
	}

	profile, err := cfg.GetProfile("work")
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}
	if profile.Model != "work-model" {
		t.Errorf("expected work-model, got %s", profile.Model)
	}

	profiles := cfg.ListProfiles()
	if len(profiles) != 1 || profiles[0] != "work" {
		t.Errorf("expected [work], got %v", profiles)
	}

	err = cfg.DeleteProfile("work")
	if err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}

	profiles = cfg.ListProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected empty list, got %v", profiles)
	}
}

func TestSetActiveProfile(t *testing.T) {
	cfg := DefaultConfig()

	if err := cfg.CreateProfile("profile1", &AgentConfig{Model: "model1"}); err != nil {
		t.Fatalf("CreateProfile failed: %v", err)
	}

	err := cfg.SetActiveProfile("profile1")
	if err != nil {
		t.Fatalf("SetActiveProfile failed: %v", err)
	}

	if cfg.ActiveProfile != "profile1" {
		t.Errorf("expected active profile 'profile1', got %q", cfg.ActiveProfile)
	}

	err = cfg.SetActiveProfile("")
	if err != nil {
		t.Fatalf("SetActiveProfile('') failed: %v", err)
	}

	if cfg.ActiveProfile != "" {
		t.Errorf("expected empty active profile, got %q", cfg.ActiveProfile)
	}
}

func TestPersonalConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	personalDir := filepath.Join(tmpDir, "specledger")
	if err := os.MkdirAll(personalDir, 0755); err != nil {
		t.Fatal(err)
	}

	personal := &PersonalConfig{
		Agent: &AgentConfig{
			APIKey: "sk-test-key",
		},
	}

	if err := personal.Save(tmpDir); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadPersonal(tmpDir)
	if err != nil {
		t.Fatalf("LoadPersonal failed: %v", err)
	}

	if loaded.Agent.APIKey != "sk-test-key" {
		t.Errorf("expected api key 'sk-test-key', got %q", loaded.Agent.APIKey)
	}
}

func TestParseArguments(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"--flag", []string{"--flag"}},
		{"--flag value", []string{"--flag", "value"}},
		{`--flag "value with spaces"`, []string{"--flag", "value with spaces"}},
		{`--flag 'single quotes'`, []string{"--flag", "single quotes"}},
		{"--one --two --three", []string{"--one", "--two", "--three"}},
		{`--msg "hello world" --count 5`, []string{"--msg", "hello world", "--count", "5"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseArguments(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseArguments(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parseArguments(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestConfigAgentsGetOrCreateAgentSettings(t *testing.T) {
	agents := NewConfigAgents()

	// Get settings for claude (should create ClaudeSettings)
	claudeSettings := agents.GetOrCreateAgentSettings("claude")
	if claudeSettings == nil {
		t.Fatal("expected claude settings, got nil")
	}
	if agents.Claude == nil {
		t.Error("expected Claude to be initialized")
	}

	// Get settings for opencode (should create AgentSettings)
	opencodeSettings := agents.GetOrCreateAgentSettings("opencode")
	if opencodeSettings == nil {
		t.Fatal("expected opencode settings, got nil")
	}
	if agents.OpenCode == nil {
		t.Error("expected OpenCode to be initialized")
	}

	// Get existing settings should return same instance
	claudeSettings2 := agents.GetOrCreateAgentSettings("claude")
	if claudeSettings != claudeSettings2 {
		t.Error("expected same instance for claude settings")
	}
}
