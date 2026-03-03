package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeConfigs(t *testing.T) {
	tests := []struct {
		name          string
		defaults      *AgentConfig
		global        *AgentConfig
		profile       *AgentConfig
		teamLocal     *AgentConfig
		personalLocal *AgentConfig
		expectedKey   string
		expectedValue string
		expectedScope ConfigScope
	}{
		{
			name:          "default only",
			defaults:      DefaultAgentConfig(),
			global:        nil,
			profile:       nil,
			teamLocal:     nil,
			personalLocal: nil,
			expectedKey:   "agent.provider",
			expectedValue: "anthropic",
			expectedScope: ScopeDefault,
		},
		{
			name: "global overrides default",
			defaults: &AgentConfig{
				Provider: "anthropic",
			},
			global: &AgentConfig{
				Provider: "bedrock",
			},
			expectedKey:   "agent.provider",
			expectedValue: "bedrock",
			expectedScope: ScopeGlobal,
		},
		{
			name:     "profile overrides global",
			defaults: &AgentConfig{},
			global: &AgentConfig{
				Model: "claude-sonnet",
			},
			profile: &AgentConfig{
				Model: "claude-opus",
			},
			expectedKey:   "agent.model",
			expectedValue: "claude-opus",
			expectedScope: ScopeProfile,
		},
		{
			name:     "team-local overrides profile",
			defaults: &AgentConfig{},
			global: &AgentConfig{
				Model: "global-model",
			},
			profile: &AgentConfig{
				Model: "profile-model",
			},
			teamLocal: &AgentConfig{
				Model: "team-model",
			},
			expectedKey:   "agent.model",
			expectedValue: "team-model",
			expectedScope: ScopeTeamLocal,
		},
		{
			name:     "personal-local highest precedence",
			defaults: &AgentConfig{},
			global: &AgentConfig{
				Model: "global-model",
			},
			profile: &AgentConfig{
				Model: "profile-model",
			},
			teamLocal: &AgentConfig{
				Model: "team-model",
			},
			personalLocal: &AgentConfig{
				Model: "personal-model",
			},
			expectedKey:   "agent.model",
			expectedValue: "personal-model",
			expectedScope: ScopePersonalLocal,
		},
		{
			name: "empty value does not override",
			defaults: &AgentConfig{
				Model: "default-model",
			},
			global: &AgentConfig{
				Model: "",
			},
			expectedKey:   "agent.model",
			expectedValue: "default-model",
			expectedScope: ScopeDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := MergeConfigs(tt.defaults, tt.global, tt.profile, tt.teamLocal, tt.personalLocal)

			value := resolved.Get(tt.expectedKey)
			if value == nil {
				t.Fatalf("expected key %s not found", tt.expectedKey)
			}

			if str, ok := value.Value.(string); !ok || str != tt.expectedValue {
				t.Errorf("expected value %q, got %v", tt.expectedValue, value.Value)
			}

			if value.Source != tt.expectedScope {
				t.Errorf("expected scope %s, got %s", tt.expectedScope, value.Source)
			}
		})
	}
}

func TestGetEnvVars(t *testing.T) {
	tests := []struct {
		name       string
		config     *AgentConfig
		expectedKV map[string]string
	}{
		{
			name: "base-url mapped to env",
			config: &AgentConfig{
				BaseURL: "https://api.test.com",
			},
			expectedKV: map[string]string{
				"ANTHROPIC_BASE_URL": "https://api.test.com",
			},
		},
		{
			name: "model mapped to env",
			config: &AgentConfig{
				Model: "claude-sonnet-4",
			},
			expectedKV: map[string]string{
				"ANTHROPIC_MODEL": "claude-sonnet-4",
			},
		},
		{
			name: "agent.env injected directly",
			config: &AgentConfig{
				Env: map[string]string{
					"CUSTOM_VAR": "custom-value",
				},
			},
			expectedKV: map[string]string{
				"CUSTOM_VAR": "custom-value",
			},
		},
		{
			name: "multiple values merged",
			config: &AgentConfig{
				BaseURL: "https://api.test.com",
				Model:   "claude-opus",
				Env: map[string]string{
					"EXTRA_VAR": "extra",
				},
			},
			expectedKV: map[string]string{
				"ANTHROPIC_BASE_URL": "https://api.test.com",
				"ANTHROPIC_MODEL":    "claude-opus",
				"EXTRA_VAR":          "extra",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := MergeConfigs(DefaultAgentConfig(), tt.config, nil, nil, nil)
			envVars := resolved.GetEnvVars()

			for key, expectedValue := range tt.expectedKV {
				if actualValue, ok := envVars[key]; !ok {
					t.Errorf("expected env var %s not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("env var %s: expected %q, got %q", key, expectedValue, actualValue)
				}
			}
		})
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
			AuthToken: "sk-test-token",
		},
	}

	if err := personal.Save(tmpDir); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadPersonal(tmpDir)
	if err != nil {
		t.Fatalf("LoadPersonal failed: %v", err)
	}

	if loaded.Agent.AuthToken != "sk-test-token" {
		t.Errorf("expected auth token 'sk-test-token', got %q", loaded.Agent.AuthToken)
	}
}
