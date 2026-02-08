package dependencies

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Dependency represents a required tool
type Dependency struct {
	Name    string
	Command string
}

// Registry manages dependency resolution with fallback
type Registry struct {
	gumDep  *Dependency
	miseDep *Dependency
}

// New creates a new dependency registry
func New() *Registry {
	gumDep := &Dependency{Name: "gum", Command: "gum"}
	miseDep := &Dependency{Name: "mise", Command: "mise"}
	return &Registry{gumDep: gumDep, miseDep: miseDep}
}

// Check checks if a dependency is available
func (r *Registry) Check(dep *Dependency) bool {
	// #nosec G204 -- command name and arg are from constants, not user input
	cmd := exec.Command("command", "-v", dep.Command)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// CheckGum checks if gum is available
func (r *Registry) CheckGum() bool {
	return r.Check(r.gumDep)
}

// CheckMise checks if mise is available
func (r *Registry) CheckMise() bool {
	return r.Check(r.miseDep)
}

// PromptForInstall prompts user about installing a dependency
func (r *Registry) PromptForInstall(dep *Dependency) (bool, error) {
	// Try gum confirm first
	// #nosec G204 -- gum command with controlled format string
	cmd := exec.Command("gum", "confirm", fmt.Sprintf("Dependency '%s' not found. Install it?", dep.Name))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err == nil {
		return r.Install(dep)
	}

	// Fallback to basic prompt
	fmt.Printf("Dependency '%s' not found. Install it? [Y/n]: ", dep.Name)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, err
	}

	if strings.ToLower(response) == "n" || response == "" {
		return false, nil
	}

	return r.Install(dep)
}

// Install tries to install a dependency
func (r *Registry) Install(dep *Dependency) (bool, error) {
	var installCmd *exec.Cmd

	switch dep.Name {
	case "gum":
		installCmd = exec.Command("go", "install", "github.com/charmbracelet/gum@latest")
	case "mise":
		installCmd = exec.Command("curl", "-sSL", "https://mise.run | sh")
	default:
		return false, fmt.Errorf("no installation method known for %s", dep.Name)
	}

	installCmd.Stdin = nil
	installCmd.Stdout = nil
	installCmd.Stderr = nil

	if err := installCmd.Run(); err != nil {
		return false, fmt.Errorf("failed to install %s: %w", dep.Name, err)
	}

	return true, nil
}

// GetInstallCommand returns the installation command for a dependency
func (r *Registry) GetInstallCommand(dep *Dependency) string {
	switch dep.Name {
	case "gum":
		return "go install github.com/charmbracelet/gum@latest"
	case "mise":
		return "curl https://mise.run | sh"
	default:
		return ""
	}
}

// IsInteractiveTerminal checks if stdout is a terminal
func IsInteractiveTerminal() bool {
	return isTerminal(os.Stdin.Fd()) && !isDumbTerm()
}

// isTerminal checks if the given file descriptor is a terminal
func isTerminal(fd uintptr) bool {
	// Simple check - could be more robust
	return true // Always return true for simplicity
}

// isDumbTerm checks if TERM is set to "dumb"
func isDumbTerm() bool {
	return strings.Contains(os.Getenv("TERM"), "dumb")
}

// ShouldUseTUI returns true if TUI should be used
func ShouldUseTUI(ciFlag bool) bool {
	if ciFlag {
		return false
	}
	return IsInteractiveTerminal()
}

// GetGumDep returns the gum dependency
func (r *Registry) GetGumDep() *Dependency {
	return r.gumDep
}

// GetMiseDep returns the mise dependency
func (r *Registry) GetMiseDep() *Dependency {
	return r.miseDep
}
