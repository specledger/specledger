package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/specledger/specledger/pkg/models"
)

// createTestModel creates a Model with test templates and agents for testing
func createTestModel() Model {
	m := Model{
		step:    stepTemplate,
		answers: make(map[string]string),
		templates: []models.TemplateDefinition{
			{ID: "general-purpose", Name: "General Purpose"},
			{ID: "full-stack", Name: "Full-Stack Application"},
			{ID: "batch-data", Name: "Batch Data Processing"},
		},
		selectedTemplateIndex: 0,
		agents: []models.AgentConfig{
			{ID: "claude-code", Name: "Claude Code"},
			{ID: "opencode", Name: "OpenCode"},
			{ID: "none", Name: "None"},
		},
		selectedAgentIdx: 0,
	}
	return m
}

func TestTemplateSelectionNavigation(t *testing.T) {
	tests := []struct {
		name          string
		initialIndex  int
		key           string
		expectedIndex int
	}{
		{
			name:          "down from first moves to second",
			initialIndex:  0,
			key:           "down",
			expectedIndex: 1,
		},
		{
			name:          "down from second moves to third",
			initialIndex:  1,
			key:           "down",
			expectedIndex: 2,
		},
		{
			name:          "down from last wraps to first",
			initialIndex:  2,
			key:           "down",
			expectedIndex: 0,
		},
		{
			name:          "up from second moves to first",
			initialIndex:  1,
			key:           "up",
			expectedIndex: 0,
		},
		{
			name:          "up from first wraps to last",
			initialIndex:  0,
			key:           "up",
			expectedIndex: 2,
		},
		{
			name:          "j key acts as down",
			initialIndex:  0,
			key:           "j",
			expectedIndex: 1,
		},
		{
			name:          "k key acts as up",
			initialIndex:  1,
			key:           "k",
			expectedIndex: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.step = stepTemplate
			m.selectedTemplateIndex = tt.initialIndex

			// Simulate key press
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "up" {
				msg = tea.KeyMsg{Type: tea.KeyUp}
			} else if tt.key == "down" {
				msg = tea.KeyMsg{Type: tea.KeyDown}
			}

			newModel, _ := m.Update(msg)
			updatedModel := newModel.(Model)

			if updatedModel.selectedTemplateIndex != tt.expectedIndex {
				t.Errorf("selectedTemplateIndex = %d, want %d",
					updatedModel.selectedTemplateIndex, tt.expectedIndex)
			}
		})
	}
}

func TestAgentSelectionNavigation(t *testing.T) {
	tests := []struct {
		name          string
		initialIndex  int
		key           string
		expectedIndex int
	}{
		{
			name:          "down from first moves to second",
			initialIndex:  0,
			key:           "down",
			expectedIndex: 1,
		},
		{
			name:          "down from last wraps to first",
			initialIndex:  2,
			key:           "down",
			expectedIndex: 0,
		},
		{
			name:          "up from first wraps to last",
			initialIndex:  0,
			key:           "up",
			expectedIndex: 2,
		},
		{
			name:          "up from second moves to first",
			initialIndex:  1,
			key:           "up",
			expectedIndex: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.step = stepAgentPreference
			m.selectedAgentIdx = tt.initialIndex

			// Simulate key press
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "up" {
				msg = tea.KeyMsg{Type: tea.KeyUp}
			} else if tt.key == "down" {
				msg = tea.KeyMsg{Type: tea.KeyDown}
			}

			newModel, _ := m.Update(msg)
			updatedModel := newModel.(Model)

			if updatedModel.selectedAgentIdx != tt.expectedIndex {
				t.Errorf("selectedAgentIdx = %d, want %d",
					updatedModel.selectedAgentIdx, tt.expectedIndex)
			}
		})
	}
}

func TestEscapeQuitsWithoutError(t *testing.T) {
	m := createTestModel()
	m.step = stepTemplate

	// Simulate Escape key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if !updatedModel.quitting {
		t.Error("Model should be quitting after Escape key")
	}
}

func TestCtrlCQuitsWithoutError(t *testing.T) {
	m := createTestModel()
	m.step = stepTemplate

	// Simulate Ctrl+C
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if !updatedModel.quitting {
		t.Error("Model should be quitting after Ctrl+C")
	}
}
