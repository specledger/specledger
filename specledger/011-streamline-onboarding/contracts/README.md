# Contracts: 011-streamline-onboarding

This feature is a CLI-only change with no external APIs, RPCs, or protocol contracts.

## Internal Interfaces

### AgentLauncher Interface

```go
// Package: pkg/cli/launcher
type AgentLauncher interface {
    // IsAvailable checks if the agent command exists in PATH
    IsAvailable() bool
    // Launch starts the agent as an interactive subprocess
    Launch() error
    // InstallInstructions returns help text for installing the agent
    InstallInstructions() string
}
```

### Constitution Checker

```go
// Package: pkg/cli/commands (bootstrap_helpers.go)

// IsConstitutionPopulated checks if the constitution file is populated
// (no placeholder tokens remaining)
func IsConstitutionPopulated(path string) bool

// WriteDefaultConstitution writes a constitution with the given principles
// and agent preference, replacing template placeholders
func WriteDefaultConstitution(path string, principles []ConstitutionPrinciple, agentPref string) error

// ReadAgentPreference extracts the preferred agent from a populated constitution
func ReadAgentPreference(path string) (string, error)
```

### TUI Init Model

```go
// Package: pkg/cli/tui

// NewInitModel creates a TUI model for sl init with only missing config steps
func NewInitModel(missingConfig MissingConfig) InitModel

type MissingConfig struct {
    NeedsShortCode       bool
    NeedsPlaybook        bool
    NeedsAgentPreference bool
}
```
