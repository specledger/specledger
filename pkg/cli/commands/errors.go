package commands

import (
	"fmt"
	"strings"
)

// CLIError represents a CLI error with actionable suggestions
type CLIError struct {
	Title       string
	Description string
	Suggestions []string
	ExitCode    int
}

// Error returns the error message
func (e *CLIError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ERROR: %s\n", e.Title))

	if e.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n", e.Description))
	}

	if len(e.Suggestions) > 0 {
		sb.WriteString("\nPossible solutions:\n")
		for i, suggestion := range e.Suggestions {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, suggestion))
		}
	}

	return sb.String()
}

// NewCLIError creates a new CLI error with suggestions
func NewCLIError(title, description string, suggestions []string, exitCode int) *CLIError {
	if suggestions == nil {
		suggestions = []string{}
	}
	return &CLIError{
		Title:       title,
		Description: description,
		Suggestions: suggestions,
		ExitCode:    exitCode,
	}
}

// Error scenarios

// ErrProjectExists indicates the project directory already exists
func ErrProjectExists(projectName string) *CLIError {
	return NewCLIError(
		"Project directory already exists",
		fmt.Sprintf("Directory '%s' already exists", projectName),
		[]string{
			"Choose a different project name",
			fmt.Sprintf("Remove the existing directory: rm -rf ~/demos/%s", projectName),
			fmt.Sprintf("Use an existing project: cd ~/demos/%s", projectName),
		},
		1,
	)
}

// ErrPermissionDenied indicates permission error
func ErrPermissionDenied(path string) *CLIError {
	return NewCLIError(
		"Permission denied",
		fmt.Sprintf("Cannot write to: %s", path),
		[]string{
			"Choose a different directory",
			"Remove write restrictions from current directory",
			"Use a different user account with proper permissions",
		},
		1,
	)
}

