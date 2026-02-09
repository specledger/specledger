package commands

import (
	"fmt"
	"os"
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

// ErrInvalidInput indicates invalid input
func ErrInvalidInput(message string) *CLIError {
	return NewCLIError(
		"Invalid input",
		message,
		nil,
		1,
	)
}

// ErrMissingDependency indicates a required dependency is not installed
func ErrMissingDependency(name string, installCmd string) *CLIError {
	return NewCLIError(
		fmt.Sprintf("Dependency '%s' not found", name),
		fmt.Sprintf("The required dependency '%s' is not installed on your system", name),
		[]string{
			fmt.Sprintf("Install using: %s", installCmd),
			"Or continue without TUI using --ci flag",
		},
		1,
	)
}

// ErrCommandNotFound indicates an invalid command
func ErrCommandNotFound(command string) *CLIError {
	return NewCLIError(
		"Unknown command",
		fmt.Sprintf("Command '%s' not found", command),
		nil,
		1,
	)
}

// ErrNotAProject indicates the current directory is not a SpecLedger project
func ErrNotAProject() *CLIError {
	return NewCLIError(
		"Not a SpecLedger project",
		"This command requires running from within a SpecLedger project",
		[]string{
			"Navigate to a SpecLedger project directory",
			"Run this command from a directory containing .github.com/specledger/specledger/",
		},
		2,
	)
}

// ErrCIRequired indicates TUI is required in non-interactive environment
func ErrCIRequired() *CLIError {
	return NewCLIError(
		"TUI mode requires interactive terminal",
		"Cannot use TUI in non-interactive environment",
		[]string{
			"Use --ci flag to force non-interactive mode",
			"Or run with proper terminal access",
		},
		1,
	)
}

// Error handling helper functions

// ExitWithError prints error message and exits
func ExitWithError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

// ExitWithCode prints error message and exits with specific code
func ExitWithCode(code int, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(code)
	}
}

// PrintSuccess prints success message
func PrintSuccess(title string, details ...string) {
	fmt.Printf("\n✓ %s\n", title)
	for _, detail := range details {
		fmt.Printf("  %s\n", detail)
	}
}
