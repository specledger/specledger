# SpecLedger Architecture

Project structure and design patterns.

## Overview

SpecLedger is a platform-agnostic bootstrap tool for SDD (Specification-Driven Development) projects. It handles project creation, dependency management, and framework integration while delegating specification workflows to external frameworks.

## Design Philosophy

1. **Framework Agnostic**: Supports multiple SDD frameworks without coupling
2. **Bootstrap Focus**: Handles only project creation and initial setup
3. **Thin Wrapper**: Delegates all SDD workflows to chosen frameworks
4. **Dependency Management**: Tracks external spec dependencies
5. **Tool Detection**: Ensures required tools are installed

## Project Structure

```
specledger/
├── cmd/
│   └── main.go              # CLI entry point
├── pkg/
│   ├── cli/                 # Command-line interface
│   │   ├── commands/        # Command implementations
│   │   ├── config/          # Configuration management
│   │   ├── dependencies/    # Dependency management
│   │   ├── framework/       # Framework integration
│   │   ├── playbooks/       # Framework workflows
│   │   ├── prerequisites/   # Tool checking
│   │   ├── tui/             # Terminal UI
│   │   ├── logger/          # Logging
│   │   └── ui/              # Formatting utilities
│   ├── embedded/           # Embedded templates
│   │   └── templates/       # Project templates
│   └── models/             # Data models
├── .specledger/            # Framework templates
│   ├── templates/          # Spec templates
│   └── scripts/            # Framework scripts
├── scripts/                # Installation scripts
├── homebrew/               # Homebrew formula
├── Makefile                # Build automation
├── go.mod                  # Go modules
├── .golangci.yml           # Linting configuration
├── .github/workflows/      # CI/CD
└── .goreleaser.yaml        # Release automation
```

## Components

### CLI Layer (`pkg/cli/`)

- **commands/**: Each command (new, init, deps, doctor, etc.)
- **config/**: Configuration file management
- **metadata/**: YAML metadata parsing and validation
- **dependencies/**: Dependency resolution and caching
- **framework/**: SDD framework integration
- **playbooks/**: Framework-specific workflow execution
- **prerequisites/**: Tool detection and installation
- **tui/**: Terminal UI components using Bubble Tea
- **logger/**: Structured logging
- **ui/**: Color and text formatting

### Embedded Templates (`pkg/embedded/`)

Project templates embedded for distribution:
- Template files for new projects
- Framework-specific configurations
- Initial documentation structure

### Framework Templates (`.specledger/`)

Templates for SpecLedger's own specification workflow:
- `spec-template.md`: Specification template
- `plan-template.md`: Implementation plan template
- `tasks-template.md`: Task list template
- Other workflow templates

## Key Design Decisions

### Why Framework Agnostic?

Different teams have different SDD preferences:
- **Spec Kit**: Structured, phase-gated approach
- **OpenSpec**: Lightweight, iterative approach
- **Custom**: Some teams have their own process

SpecLedger supports all these choices without imposing a specific workflow.

### Why Delegation?

SpecLedger focuses on what it does well:
- Project creation and structure
- Dependency management
- Tool verification

While frameworks handle:
- Specification workflows
- Implementation planning
- Task tracking
- Quality gates

### Why Embedded Templates?

Templates are embedded in the binary to:
- Ensure consistency across projects
- Allow distribution without external dependencies
- Enable offline project creation
- Simplify installation (single binary)

## Data Flow

### Project Creation Flow

```
User runs: sl new
    ↓
TUI collects: project name, short code, framework choice
    ↓
Template embedded: Creates project structure
    ↓
Framework detected: Installs chosen framework (mise)
    ↓
Framework initialized: Runs framework init command
    ↓
Project ready: specledger.yaml created with configuration
```

### Dependency Resolution Flow

```
User runs: sl deps add <url>
    ↓
Repository cloned: Temporarily to detect framework
    ↓
Framework detected: Spec Kit, OpenSpec, both, or none
    ↓
Import path generated: @alias or @reponame
    ↓
Added to YAML: Dependency metadata saved
    ↓
Cached locally: ~/.specledger/cache/@alias/
```

## Extension Points

SpecLedger can be extended with:

1. **New Frameworks**: Add framework detection and initialization
2. **New Prerequisites**: Add tools to check in `sl doctor`
3. **New Templates**: Add embedded project templates
4. **New Commands**: Add commands in `pkg/cli/commands/`

## Dependencies

- **Cobra**: CLI framework (command parsing, help text)
- **Bubble Tea**: Terminal UI (TUI for interactive mode)
- **go-git**: Git operations (cloning dependencies)
- **YAML v3**: Configuration parsing (specledger.yaml)
- **GoReleaser**: Release automation (multi-platform builds)
- **GitHub Actions**: CI/CD (testing, linting, releases)

## Build Process

1. **Source**: `go build ./cmd`
2. **Cross-compile**: Via GoReleaser for multiple platforms
3. **Package**: Homebrew formula, npm package
4. **Release**: GitHub Actions on tag push
