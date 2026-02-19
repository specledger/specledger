package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/specledger/specledger/pkg/cli/launcher"
)

// MissingConfig describes which configuration steps are still needed.
type MissingConfig struct {
	NeedsShortCode       bool
	NeedsPlaybook        bool
	NeedsAgentPreference bool
	ExistingAgentPref    string // Pre-select this agent if set
}

// InitStepType identifies the kind of input a step requires.
type InitStepType int

const (
	initStepTextInput InitStepType = iota
	initStepListSelect
)

// InitStep describes a single step in the sl init TUI.
type InitStep struct {
	Key      string       // "short_code", "playbook", "agent_preference"
	StepType InitStepType
}

// InitModel is the Bubble Tea model for sl init.
type InitModel struct {
	steps       []InitStep
	currentStep int
	answers     map[string]string
	textInput   textinput.Model
	selectedIdx int
	width       int
	quitting    bool
	showingError string
	projectName string

	// Agent preference options
	agentOptions     []launcher.AgentOption
	selectedAgentIdx int
}

// NewInitModel creates an InitModel that presents only the missing configuration steps.
func NewInitModel(config MissingConfig, projectName string) InitModel {
	var steps []InitStep

	if config.NeedsShortCode {
		steps = append(steps, InitStep{Key: "short_code", StepType: initStepTextInput})
	}
	if config.NeedsPlaybook {
		steps = append(steps, InitStep{Key: "playbook", StepType: initStepListSelect})
	}
	if config.NeedsAgentPreference {
		steps = append(steps, InitStep{Key: "agent_preference", StepType: initStepListSelect})
	}

	ti := textinput.New()
	ti.CharLimit = 4
	ti.Width = 20
	ti.Focus()

	// Auto-derive short code from project name
	defaultShortCode := "sl"
	if len(projectName) >= 2 {
		defaultShortCode = strings.ToLower(projectName[:2])
	}
	ti.SetValue(defaultShortCode)
	ti.Placeholder = "xy"

	// Pre-select existing agent preference if available
	selectedAgent := 0
	if config.ExistingAgentPref != "" {
		for i, a := range launcher.DefaultAgents {
			if a.Name == config.ExistingAgentPref {
				selectedAgent = i
				break
			}
		}
	}

	return InitModel{
		steps:            steps,
		currentStep:      0,
		answers:          make(map[string]string),
		textInput:        ti,
		selectedIdx:      0,
		width:            80,
		projectName:      projectName,
		agentOptions:     launcher.DefaultAgents,
		selectedAgentIdx: selectedAgent,
	}
}

// Init initializes the model.
func (m InitModel) Init() tea.Cmd {
	if len(m.steps) > 0 && m.steps[0].StepType == initStepTextInput {
		return textinput.Blink
	}
	return nil
}

// Update handles messages.
func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if len(m.steps) == 0 {
		m.quitting = true
		return m, tea.Quit
	}

	currentStepDef := m.steps[m.currentStep]

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle list navigation
		if currentStepDef.StepType == initStepListSelect {
			switch msg.String() {
			case "up", "k":
				switch currentStepDef.Key {
				case "playbook":
					if m.selectedIdx > 0 {
						m.selectedIdx--
					}
				case "agent_preference":
					if m.selectedAgentIdx > 0 {
						m.selectedAgentIdx--
					}
				}
				return m, nil
			case "down", "j":
				switch currentStepDef.Key {
				case "playbook":
					playbooks := getAvailablePlaybooks()
					if m.selectedIdx < len(playbooks)-1 {
						m.selectedIdx++
					}
				case "agent_preference":
					if m.selectedAgentIdx < len(m.agentOptions)-1 {
						m.selectedAgentIdx++
					}
				}
				return m, nil
			}
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	// Update text input for text entry steps
	if currentStepDef.StepType == initStepTextInput {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m InitModel) handleEnter() (tea.Model, tea.Cmd) {
	m.showingError = ""
	currentStepDef := m.steps[m.currentStep]

	switch currentStepDef.Key {
	case "short_code":
		value := strings.TrimSpace(m.textInput.Value())
		if value == "" {
			m.showingError = "Short code cannot be empty"
			return m, nil
		}
		if len(value) > 4 {
			m.showingError = "Short code must be 4 characters or less"
			return m, nil
		}
		m.answers["short_code"] = strings.ToLower(value)

	case "playbook":
		playbooks := getAvailablePlaybooks()
		if m.selectedIdx >= 0 && m.selectedIdx < len(playbooks) {
			m.answers["playbook"] = playbooks[m.selectedIdx]
		}

	case "agent_preference":
		if m.selectedAgentIdx >= 0 && m.selectedAgentIdx < len(m.agentOptions) {
			m.answers["agent_preference"] = m.agentOptions[m.selectedAgentIdx].Name
		}
	}

	// Advance to next step or complete
	m.currentStep++
	if m.currentStep >= len(m.steps) {
		m.quitting = true
		return m, tea.Quit
	}

	// Reset selection for next step
	m.selectedIdx = 0
	m.selectedAgentIdx = 0

	return m, nil
}

// View renders the TUI.
func (m InitModel) View() string {
	if m.quitting {
		return ""
	}
	if len(m.steps) == 0 {
		return ""
	}

	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(titleStyle.Render("SpecLedger Init"))
	s.WriteString("\n\n")

	s.WriteString(colorSubtle.Render(fmt.Sprintf("Step %d of %d", m.currentStep+1, len(m.steps))))
	s.WriteString("\n\n")

	currentStepDef := m.steps[m.currentStep]

	switch currentStepDef.Key {
	case "short_code":
		s.WriteString(m.viewInitShortCode())
	case "playbook":
		s.WriteString(m.viewInitPlaybook())
	case "agent_preference":
		s.WriteString(m.viewInitAgentPreference())
	}

	if m.showingError != "" {
		s.WriteString("\n\n")
		s.WriteString(colorError.Render("✗ " + m.showingError))
	}

	s.WriteString("\n\n")
	if currentStepDef.StepType == initStepListSelect {
		s.WriteString(colorSubtle.Render("↑/↓: Select • Enter: Confirm • Ctrl+C: Cancel"))
	} else {
		s.WriteString(colorSubtle.Render("Enter: Continue • Ctrl+C: Cancel"))
	}
	s.WriteString("\n")

	return s.String()
}

func (m InitModel) viewInitShortCode() string {
	var s strings.Builder
	s.WriteString(colorPrimary.Render("Short Code"))
	s.WriteString("\n")
	s.WriteString(colorSubtle.Render(fmt.Sprintf("A 2-4 letter prefix for issues (project: %s)", m.projectName)))
	s.WriteString("\n\n")
	s.WriteString(m.textInput.View())
	return s.String()
}

func (m InitModel) viewInitPlaybook() string {
	var s strings.Builder
	s.WriteString(colorPrimary.Render("Select Playbook"))
	s.WriteString("\n")
	s.WriteString(colorSubtle.Render("Choose the playbook to apply to your project"))
	s.WriteString("\n\n")

	playbooks := getAvailablePlaybooks()
	descriptions := getPlaybookDescriptions()

	for i, playbook := range playbooks {
		cursor := " "
		radio := "○"
		style := unselectedStyle

		if i == m.selectedIdx {
			cursor = "›"
			style = selectedStyle
			radio = "◉"
		}

		s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, radio, style.Render(playbook)))
		if i == m.selectedIdx {
			s.WriteString(colorSubtle.Render("  " + descriptions[i]))
			s.WriteString("\n")
		}
	}

	return s.String()
}

func (m InitModel) viewInitAgentPreference() string {
	var s strings.Builder
	s.WriteString(colorPrimary.Render("AI Coding Agent"))
	s.WriteString("\n")
	s.WriteString(colorSubtle.Render("Choose an AI agent to launch after setup"))
	s.WriteString("\n\n")

	for i, agent := range m.agentOptions {
		cursor := " "
		radio := "○"
		style := unselectedStyle

		if i == m.selectedAgentIdx {
			cursor = "›"
			style = selectedStyle
			radio = "◉"
		}

		s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, radio, style.Render(agent.Name)))
		if i == m.selectedAgentIdx {
			s.WriteString(colorSubtle.Render("  " + agent.Description))
			s.WriteString("\n")
		}
	}

	return s.String()
}

// InitProgram wraps the init TUI execution.
type InitProgram struct {
	teaProgram *tea.Program
	model      InitModel
}

// NewInitProgram creates a new sl init TUI program.
func NewInitProgram(config MissingConfig, projectName string) *InitProgram {
	m := NewInitModel(config, projectName)
	p := tea.NewProgram(m)

	return &InitProgram{
		teaProgram: p,
		model:      m,
	}
}

// Run runs the init program to completion.
func (p *InitProgram) Run() (map[string]string, error) {
	result, err := p.teaProgram.Run()
	if err != nil {
		return nil, err
	}

	if model, ok := result.(InitModel); ok {
		p.model = model
	}

	return p.model.answers, nil
}
