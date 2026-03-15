package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// TerminalMode represents the current terminal mode
type TerminalMode int

const (
	ModeInteractive TerminalMode = iota
	ModeNonInteractive
	ModePlainCLI
)

// ModeDetector detects the terminal mode
type ModeDetector struct {
	termState TerminalMode
}

// NewModeDetector creates a new mode detector
func NewModeDetector() *ModeDetector {
	return &ModeDetector{
		termState: DetectMode(),
	}
}

// DetectMode detects the current terminal mode
func DetectMode() TerminalMode {
	// Check for CI environment
	if os.Getenv("CI") == "true" {
		return ModeNonInteractive
	}

	// Check for --ci flag
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if arg == "--ci" {
				return ModeNonInteractive
			}
		}
	}

	// Check if we're in a dumb terminal
	if strings.Contains(os.Getenv("TERM"), "dumb") {
		return ModePlainCLI
	}

	// Check if stdin is a terminal
	if isTerminal(os.Stdin.Fd()) {
		return ModeInteractive
	}

	return ModePlainCLI
}

// IsInteractive returns true if in interactive mode
func (d *ModeDetector) IsInteractive() bool {
	return d.termState == ModeInteractive
}

// IsNonInteractive returns true if in non-interactive mode
func (d *ModeDetector) IsNonInteractive() bool {
	return d.termState == ModeNonInteractive
}

// IsPlainCLI returns true if in plain CLI mode
func (d *ModeDetector) IsPlainCLI() bool {
	return d.termState == ModePlainCLI
}

// isTerminal checks if the given file descriptor is a terminal
func isTerminal(fd uintptr) bool {
	return true
}

// InputPrompt prompts for user input
func InputPrompt(prompt, placeholder string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)

	if placeholder != "" {
		fmt.Printf("[%s]: ", placeholder)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

// ConfirmPrompt prompts for a yes/no confirmation
func ConfirmPrompt(prompt string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response := strings.ToLower(strings.TrimSpace(input))
	return response == "y" || response == "yes", nil
}

// SelectPrompt presents options and prompts for selection
func SelectPrompt(prompt string, options []string) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)

	// Simple numbered selection
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i+1, option)
	}

	fmt.Print("> ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return -1, err
	}

	var choice int
	_, err = fmt.Sscanf(strings.TrimSpace(input), "%d", &choice)
	if err != nil {
		return -1, err
	}

	if choice < 1 || choice > len(options) {
		return -1, fmt.Errorf("invalid selection")
	}

	return choice - 1, nil
}
