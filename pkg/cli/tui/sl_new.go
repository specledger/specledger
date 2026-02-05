package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Colors and styles
var (
	colorPrimary    = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))  // Gold
	colorSuccess    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // Green
	colorWarning    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))  // Yellow
	colorError      = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))   // Red
	colorSubtle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	colorBackground = lipgloss.NewStyle().Background(lipgloss.Color("235"))
	colorBorder     = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))

	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	selectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
	unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// Step constants
const (
	stepProjectName = iota
	stepDirectory
	stepShortCode
	stepFramework
	stepConfirm
	stepComplete
)

// TUI model for sl new command
type Model struct {
	step             int
	textInput        textinput.Model
	answers          map[string]string
	showingError     string
	width            int
	selectedIdx      int                // Index for list selections
	frameworkChoices map[string]bool    // Checkbox states for framework selection
	quitting         bool
	defaultDir       string             // Default project directory
}

// InitialModel creates initial model with default directory
func InitialModel(defaultDir string) Model {
	ti := textinput.New()
	ti.Placeholder = "my-project"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 50

	return Model{
		step:             stepProjectName,
		textInput:        ti,
		answers:          make(map[string]string),
		frameworkChoices: make(map[string]bool),
		width:            80,
		defaultDir:       defaultDir,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyUp:
			if m.step == stepFramework {
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
			}
			return m, nil

		case tea.KeyDown:
			if m.step == stepFramework {
				frameworkOptions := getFrameworkCheckboxOptions()
				if m.selectedIdx < len(frameworkOptions)-1 {
					m.selectedIdx++
				}
			}
			return m, nil

		case tea.KeyRunes:
			// Handle space bar for checkbox toggle at framework step
			if m.step == stepFramework && len(msg.Runes) > 0 && msg.Runes[0] == ' ' {
				options := getFrameworkCheckboxOptions()
				if m.selectedIdx < len(options) {
					key := options[m.selectedIdx]
					m.frameworkChoices[key] = !m.frameworkChoices[key]
				}
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	// Update text input for text entry steps
	if m.step == stepProjectName || m.step == stepDirectory || m.step == stepShortCode {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleEnter processes Enter key based on current step
func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	m.showingError = ""

	switch m.step {
	case stepProjectName:
		value := strings.TrimSpace(m.textInput.Value())
		if value == "" {
			m.showingError = "Project name cannot be empty"
			return m, nil
		}
		m.answers["project_name"] = value
		m.step = stepDirectory

		// Setup text input for directory with default value
		m.textInput.SetValue(m.defaultDir)
		m.textInput.Placeholder = "/path/to/projects"
		m.textInput.CharLimit = 200
		m.textInput.Focus()
		return m, nil

	case stepDirectory:
		value := strings.TrimSpace(m.textInput.Value())
		if value == "" {
			m.showingError = "Directory cannot be empty"
			return m, nil
		}
		m.answers["project_dir"] = value
		m.step = stepShortCode

		// Setup text input for short code with default value
		projectName := m.answers["project_name"]
		m.textInput.SetValue(generateShortCode(projectName))
		m.textInput.Placeholder = "xy"
		m.textInput.CharLimit = 4
		m.textInput.Focus()
		return m, nil

	case stepShortCode:
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
		m.step = stepFramework
		m.selectedIdx = 0
		return m, nil

	case stepFramework:
		// Determine framework choice based on checkbox selections
		speckit := m.frameworkChoices["speckit"]
		openspec := m.frameworkChoices["openspec"]

		if speckit && openspec {
			m.answers["framework"] = "both"
		} else if speckit {
			m.answers["framework"] = "speckit"
		} else if openspec {
			m.answers["framework"] = "openspec"
		} else {
			m.answers["framework"] = "none"
		}

		m.step = stepConfirm
		return m, nil

	case stepConfirm:
		m.step = stepComplete
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if m.quitting && m.step == stepComplete {
		return m.viewComplete()
	}

	var s strings.Builder

	// Header
	s.WriteString("\n")
	s.WriteString(titleStyle.Render("SpecLedger Bootstrap"))
	s.WriteString("\n\n")

	// Step indicator
	s.WriteString(colorSubtle.Render(fmt.Sprintf("Step %d of 5", m.step+1)))
	s.WriteString("\n\n")

	// Step content
	switch m.step {
	case stepProjectName:
		s.WriteString(m.viewProjectName())
	case stepDirectory:
		s.WriteString(m.viewDirectory())
	case stepShortCode:
		s.WriteString(m.viewShortCode())
	case stepFramework:
		s.WriteString(m.viewFramework())
	case stepConfirm:
		s.WriteString(m.viewConfirm())
	}

	// Error message
	if m.showingError != "" {
		s.WriteString("\n\n")
		s.WriteString(colorError.Render("✗ " + m.showingError))
	}

	// Help text
	s.WriteString("\n\n")
	if m.step == stepFramework {
		s.WriteString(colorSubtle.Render("↑/↓: Navigate • Space: Toggle • Enter: Confirm • Ctrl+C: Cancel"))
	} else {
		s.WriteString(colorSubtle.Render("Enter: Continue • Ctrl+C: Cancel"))
	}
	s.WriteString("\n")

	return s.String()
}

func (m Model) viewProjectName() string {
	var s strings.Builder
	s.WriteString(colorPrimary.Render("Project Name"))
	s.WriteString("\n")
	s.WriteString(colorSubtle.Render("This will be your project directory name"))
	s.WriteString("\n\n")
	s.WriteString(m.textInput.View())
	return s.String()
}

func (m Model) viewDirectory() string {
	var s strings.Builder
	s.WriteString(colorPrimary.Render("Project Directory"))
	s.WriteString("\n")
	s.WriteString(colorSubtle.Render("Parent directory where the project will be created"))
	s.WriteString("\n\n")
	s.WriteString(m.textInput.View())
	return s.String()
}

func (m Model) viewShortCode() string {
	var s strings.Builder
	s.WriteString(colorPrimary.Render("Short Code"))
	s.WriteString("\n")
	s.WriteString(colorSubtle.Render("A 2-4 letter prefix for issues (e.g., 'sl' for specledger-123)"))
	s.WriteString("\n\n")
	s.WriteString(m.textInput.View())
	return s.String()
}

func (m Model) viewFramework() string {
	var s strings.Builder
	s.WriteString(colorPrimary.Render("Select SDD Framework(s)"))
	s.WriteString("\n")
	s.WriteString(colorSubtle.Render("Use Space to toggle, Enter to confirm"))
	s.WriteString("\n\n")

	options := getFrameworkCheckboxOptions()
	descriptions := getFrameworkCheckboxDescriptions()

	for i, option := range options {
		cursor := " "
		checkbox := "☐"
		style := unselectedStyle

		if i == m.selectedIdx {
			cursor = "›"
			style = selectedStyle
		}

		if m.frameworkChoices[option] {
			checkbox = "☑"
		}

		s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, checkbox, style.Render(getFrameworkDisplayName(option))))

		if i == m.selectedIdx {
			s.WriteString(colorSubtle.Render("  " + descriptions[i]))
			s.WriteString("\n")
		}
	}

	return s.String()
}

func (m Model) viewConfirm() string {
	var s strings.Builder
	s.WriteString(colorPrimary.Render("Confirm Configuration"))
	s.WriteString("\n\n")

	projectPath := m.answers["project_dir"] + "/" + m.answers["project_name"]
	s.WriteString(colorSuccess.Render("✓ ") + "Project Name: " + m.answers["project_name"] + "\n")
	s.WriteString(colorSuccess.Render("✓ ") + "Location: " + projectPath + "\n")
	s.WriteString(colorSuccess.Render("✓ ") + "Short Code: " + m.answers["short_code"] + "\n")

	// Display framework selection with details
	framework := m.answers["framework"]
	frameworkDisplay := ""
	switch framework {
	case "speckit":
		frameworkDisplay = "Spec Kit (GitHub)"
	case "openspec":
		frameworkDisplay = "OpenSpec"
	case "both":
		frameworkDisplay = "Both (Spec Kit + OpenSpec)"
	case "none":
		frameworkDisplay = "None"
	default:
		frameworkDisplay = framework
	}
	s.WriteString(colorSuccess.Render("✓ ") + "SDD Framework: " + frameworkDisplay + "\n")

	s.WriteString("\n")
	s.WriteString(colorSubtle.Render("Press Enter to create project"))

	return s.String()
}

func (m Model) viewComplete() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(colorSuccess.Bold(true).Render("✓ Bootstrap Complete!"))
	s.WriteString("\n\n")
	s.WriteString("Creating your SpecLedger project...")
	return s.String()
}

// Helper functions

// getFrameworkCheckboxOptions returns the keys for checkbox selection
func getFrameworkCheckboxOptions() []string {
	return []string{
		"speckit",
		"openspec",
	}
}

// getFrameworkDisplayName returns the display name for a framework key
func getFrameworkDisplayName(key string) string {
	switch key {
	case "speckit":
		return "Spec Kit (GitHub)"
	case "openspec":
		return "OpenSpec"
	default:
		return key
	}
}

// getFrameworkCheckboxDescriptions returns descriptions for each framework
func getFrameworkCheckboxDescriptions() []string {
	return []string{
		"Structured, phase-gated workflow with comprehensive tooling",
		"Lightweight, iterative workflow for agile teams",
	}
}

func generateShortCode(projectName string) string {
	// Generate a short code from project name
	projectName = strings.ToLower(projectName)
	if len(projectName) >= 2 {
		return projectName[:2]
	}
	if len(projectName) == 1 {
		return projectName
	}
	return ""
}

// Program wraps Bubble Tea program execution
type Program struct {
	teaProgram *tea.Program
	model      Model
}

// NewProgram creates a new interactive Bubble Tea program
func NewProgram(defaultDir string) *Program {
	m := InitialModel(defaultDir)
	p := tea.NewProgram(m)

	return &Program{
		teaProgram: p,
		model:      m,
	}
}

// Run runs the program to completion
func (p *Program) Run() (map[string]string, error) {
	result, err := p.teaProgram.Run()
	if err != nil {
		return nil, err
	}

	if model, ok := result.(Model); ok {
		p.model = model
	}

	return p.model.answers, nil
}
