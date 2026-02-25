// Package hooks provides Claude Code hook management functionality.
package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ClaudeSettings represents the structure of ~/.claude/settings.json
type ClaudeSettings struct {
	Hooks map[string][]HookMatcher `json:"hooks,omitempty"`
	// Preserve other fields
	Other map[string]json.RawMessage `json:"-"`
}

// HookMatcher represents a hook matcher in Claude settings
type HookMatcher struct {
	Matcher string `json:"matcher"`
	Hooks   []Hook `json:"hooks"`
}

// Hook represents a single hook configuration
type Hook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// Custom JSON marshaling to preserve unknown fields
func (s *ClaudeSettings) UnmarshalJSON(data []byte) error {
	// First unmarshal into a map to capture all fields
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	s.Other = make(map[string]json.RawMessage)

	// Extract hooks if present
	if hooksData, ok := raw["hooks"]; ok {
		if err := json.Unmarshal(hooksData, &s.Hooks); err != nil {
			return err
		}
		delete(raw, "hooks")
	}

	// Store remaining fields
	s.Other = raw
	return nil
}

func (s ClaudeSettings) MarshalJSON() ([]byte, error) {
	// Start with other fields
	result := make(map[string]interface{})
	for k, v := range s.Other {
		var val interface{}
		if err := json.Unmarshal(v, &val); err != nil {
			return nil, err
		}
		result[k] = val
	}

	// Add hooks if present
	if len(s.Hooks) > 0 {
		result["hooks"] = s.Hooks
	}

	// Return without indent - caller (SaveClaudeSettings) handles indentation
	return json.Marshal(result)
}

// getClaudeSettingsPath returns the path to ~/.claude/settings.json
func getClaudeSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

// LoadClaudeSettings loads settings from ~/.claude/settings.json
func LoadClaudeSettings() (*ClaudeSettings, error) {
	path, err := getClaudeSettingsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty settings if file doesn't exist
			return &ClaudeSettings{
				Hooks: make(map[string][]HookMatcher),
				Other: make(map[string]json.RawMessage),
			}, nil
		}
		return nil, fmt.Errorf("failed to read settings: %w", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	if settings.Hooks == nil {
		settings.Hooks = make(map[string][]HookMatcher)
	}
	if settings.Other == nil {
		settings.Other = make(map[string]json.RawMessage)
	}

	return &settings, nil
}

// SaveClaudeSettings saves settings to ~/.claude/settings.json
func SaveClaudeSettings(settings *ClaudeSettings) error {
	path, err := getClaudeSettingsPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

// HasSessionCaptureHook checks if the session capture hook is already installed
func HasSessionCaptureHook(settings *ClaudeSettings) bool {
	postToolUse, ok := settings.Hooks["PostToolUse"]
	if !ok {
		return false
	}

	for _, matcher := range postToolUse {
		if matcher.Matcher == "Bash" {
			for _, hook := range matcher.Hooks {
				if hook.Command == "sl session capture" {
					return true
				}
			}
		}
	}

	return false
}

// InstallSessionCaptureHook installs the PostToolUse hook for sl session capture.
// Returns (installed bool, err error) where installed is true if a new hook was added.
func InstallSessionCaptureHook() (bool, error) {
	settings, err := LoadClaudeSettings()
	if err != nil {
		return false, err
	}

	// Check if already installed
	if HasSessionCaptureHook(settings) {
		return false, nil // Already installed, nothing to do
	}

	// Create the hook
	sessionCaptureHook := Hook{
		Type:    "command",
		Command: "sl session capture",
	}

	// Find or create the Bash matcher in PostToolUse
	postToolUse := settings.Hooks["PostToolUse"]
	found := false
	for i, matcher := range postToolUse {
		if matcher.Matcher == "Bash" {
			// Add to existing Bash matcher
			postToolUse[i].Hooks = append(postToolUse[i].Hooks, sessionCaptureHook)
			found = true
			break
		}
	}

	if !found {
		// Create new Bash matcher
		postToolUse = append(postToolUse, HookMatcher{
			Matcher: "Bash",
			Hooks:   []Hook{sessionCaptureHook},
		})
	}

	settings.Hooks["PostToolUse"] = postToolUse

	if err := SaveClaudeSettings(settings); err != nil {
		return false, err
	}

	return true, nil
}

// UninstallSessionCaptureHook removes the session capture hook.
// Returns (removed bool, err error) where removed is true if the hook was found and removed.
func UninstallSessionCaptureHook() (bool, error) {
	settings, err := LoadClaudeSettings()
	if err != nil {
		return false, err
	}

	postToolUse, ok := settings.Hooks["PostToolUse"]
	if !ok {
		return false, nil // No PostToolUse hooks
	}

	removed := false
	newPostToolUse := make([]HookMatcher, 0, len(postToolUse))

	for _, matcher := range postToolUse {
		if matcher.Matcher == "Bash" {
			// Filter out sl session capture
			newHooks := make([]Hook, 0, len(matcher.Hooks))
			for _, hook := range matcher.Hooks {
				if hook.Command != "sl session capture" {
					newHooks = append(newHooks, hook)
				} else {
					removed = true
				}
			}

			// Only keep matcher if it still has hooks
			if len(newHooks) > 0 {
				matcher.Hooks = newHooks
				newPostToolUse = append(newPostToolUse, matcher)
			}
		} else {
			newPostToolUse = append(newPostToolUse, matcher)
		}
	}

	if !removed {
		return false, nil
	}

	if len(newPostToolUse) > 0 {
		settings.Hooks["PostToolUse"] = newPostToolUse
	} else {
		delete(settings.Hooks, "PostToolUse")
	}

	if err := SaveClaudeSettings(settings); err != nil {
		return false, err
	}

	return true, nil
}
