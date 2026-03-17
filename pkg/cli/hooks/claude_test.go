package hooks

import (
	"encoding/json"
	"testing"
)

func TestClaudeSettingsJSON(t *testing.T) {
	t.Run("unmarshal with hooks", func(t *testing.T) {
		data := `{
			"hooks": {
				"PostToolUse": [
					{
						"matcher": "Bash",
						"hooks": [{"type": "command", "command": "sl session capture"}]
					}
				]
			},
			"otherField": "value"
		}`

		var settings ClaudeSettings
		if err := json.Unmarshal([]byte(data), &settings); err != nil {
			t.Fatalf("Unmarshal() error: %v", err)
		}

		if len(settings.Hooks) != 1 {
			t.Errorf("expected 1 hook type, got %d", len(settings.Hooks))
		}

		postToolUse := settings.Hooks["PostToolUse"]
		if len(postToolUse) != 1 {
			t.Fatalf("expected 1 PostToolUse matcher, got %d", len(postToolUse))
		}

		if postToolUse[0].Matcher != "Bash" {
			t.Errorf("expected matcher 'Bash', got %q", postToolUse[0].Matcher)
		}

		if len(settings.Other) == 0 {
			t.Error("expected Other fields to be preserved")
		}
	})

	t.Run("marshal with hooks", func(t *testing.T) {
		settings := &ClaudeSettings{
			Hooks: map[string][]HookMatcher{
				"PostToolUse": {
					{
						Matcher: "Bash",
						Hooks:   []Hook{{Type: "command", Command: "sl session capture"}},
					},
				},
			},
			Other: map[string]json.RawMessage{
				"otherField": json.RawMessage(`"value"`),
			},
		}

		data, err := json.Marshal(settings)
		if err != nil {
			t.Fatalf("Marshal() error: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal result error: %v", err)
		}

		if _, ok := result["hooks"]; !ok {
			t.Error("expected hooks field in output")
		}
		if _, ok := result["otherField"]; !ok {
			t.Error("expected otherField to be preserved")
		}
	})
}

func TestIsSessionCaptureCommand(t *testing.T) {
	tests := []struct {
		cmd      string
		expected bool
	}{
		{"sl session capture", true},
		{"/usr/local/bin/sl session capture", true},
		{"/home/user/go/bin/sl session capture", true},
		{"other command", false},
		{"sl", false},
		{"sl session", false},
		{"capture session", false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if got := isSessionCaptureCommand(tt.cmd); got != tt.expected {
				t.Errorf("isSessionCaptureCommand(%q) = %v, want %v", tt.cmd, got, tt.expected)
			}
		})
	}
}

func TestHasSessionCaptureHook(t *testing.T) {
	t.Run("has hook", func(t *testing.T) {
		settings := &ClaudeSettings{
			Hooks: map[string][]HookMatcher{
				"PostToolUse": {
					{
						Matcher: "Bash",
						Hooks:   []Hook{{Type: "command", Command: "sl session capture"}},
					},
				},
			},
		}

		if !HasSessionCaptureHook(settings) {
			t.Error("expected HasSessionCaptureHook to return true")
		}
	})

	t.Run("no hook", func(t *testing.T) {
		settings := &ClaudeSettings{
			Hooks: map[string][]HookMatcher{
				"PostToolUse": {
					{
						Matcher: "Bash",
						Hooks:   []Hook{{Type: "command", Command: "other command"}},
					},
				},
			},
		}

		if HasSessionCaptureHook(settings) {
			t.Error("expected HasSessionCaptureHook to return false")
		}
	})

	t.Run("no PostToolUse hooks", func(t *testing.T) {
		settings := &ClaudeSettings{
			Hooks: map[string][]HookMatcher{},
		}

		if HasSessionCaptureHook(settings) {
			t.Error("expected HasSessionCaptureHook to return false")
		}
	})

	t.Run("no Bash matcher", func(t *testing.T) {
		settings := &ClaudeSettings{
			Hooks: map[string][]HookMatcher{
				"PostToolUse": {
					{
						Matcher: "Other",
						Hooks:   []Hook{{Type: "command", Command: "sl session capture"}},
					},
				},
			},
		}

		if HasSessionCaptureHook(settings) {
			t.Error("expected HasSessionCaptureHook to return false for non-Bash matcher")
		}
	})
}
