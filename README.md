# SpecLedger

[![Build Status](https://img.shields.io/github/actions/workflow/status/specledger/specledger/ci.yml?branch=main)](https://github.com/specledger/specledger/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/specledger/specledger)](https://goreportcard.com/report/github.com/specledger/specledger)
[![Coverage](https://img.shields.io/codecov/c/github/specledger/specledger)](https://codecov.io/gh/specledger/specledger)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/github/v/release/specledger/specledger)](https://github.com/specledger/specledger/releases)

> Platform-agnostic CLI for project bootstrap and SDD framework setup

SpecLedger (`sl`) is a lightweight CLI that helps you create new projects and set up Specification-Driven Development frameworks like [Spec Kit](https://github.com/github/spec-kit) or [OpenSpec](https://github.com/fission-ai/openspec).

## Features

- **Framework Agnostic**: Choose Spec Kit, OpenSpec, both, or none for SDD workflows
- **Interactive TUI**: Create projects with a beautiful terminal interface
- **Prerequisites Checking**: Automatically detect and install required tools (mise, bd, perles)
- **Dependency Management**: Add, remove, and list spec dependencies
- **YAML Metadata**: Modern, human-readable project configuration with `specledger.yaml`
- **Migration Support**: Convert legacy `.mod` files to YAML with `sl migrate`
- **Local Caching**: Dependencies are cached locally at `~/.specledger/cache` for offline use
- **LLM Integration**: Cached specs can be easily referenced by AI agents
- **Cross-Platform**: Works on Linux, macOS, and Windows

## What SpecLedger Does (And Doesn't Do)

SpecLedger is a **thin wrapper** that handles:

✅ **Project creation** - Creates directory structure and configuration files
✅ **Framework setup** - Installs and initializes your chosen SDD framework
✅ **Tool checking** - Verifies required tools (mise, bd, perles) are installed
✅ **Dependency tracking** - Manages external specification dependencies
✅ **Metadata management** - Maintains project configuration in YAML format

❌ **NOT a full SDD framework** - Actual specification workflows are delegated to the framework you choose
❌ **NOT a task tracker** - Uses [beads](https://github.com/amelie/beads) for issue tracking
❌ **NOT a spec tool** - Relies on external frameworks for specification workflows

## Installation

### From GitHub Releases

#### macOS

```bash
# Using curl
curl -fsSL https://raw.githubusercontent.com/specledger/specledger/main/scripts/install.sh | bash

# Or using wget
wget -qO- https://raw.githubusercontent.com/specledger/specledger/main/scripts/install.sh | bash
```

#### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/specledger/specledger/main/scripts/install.sh | sudo bash

# Or using wget
wget -qO- https://raw.githubusercontent.com/specledger/specledger/main/scripts/install.sh | sudo bash
```

#### Windows

```powershell
# Using PowerShell
irm https://raw.githubusercontent.com/specledger/specledger/main/scripts/install.ps1 | iex
```

### From Source

```bash
# Clone the repository
git clone https://github.com/specledger/specledger.git
cd specledger

# Build and install the CLI
make install

# The CLI is installed to $GOPATH/bin or ~/go/bin
# Make sure $GOPATH/bin is in your PATH
sl --help
```

### Using Go Toolchain

```bash
# For local development (from source)
cd specledger
go install ./cmd/main.go

# This installs to $GOPATH/bin (usually ~/go/bin/sl)
# Add $GOPATH/bin to your PATH if not already:
export PATH=$PATH:$(go env GOPATH)/bin

# Verify installation:
sl --version
```

### Using Package Managers

#### Homebrew

```bash
brew tap specledger/homebrew-specledger
brew install specledger
```

#### npm / npx

```bash
npx @specledger/cli@latest --help
```

## Quick Start

### Create a New Project

```bash
# Interactive mode (recommended)
sl new

# Non-interactive mode (for CI/CD)
sl new --ci --project-name myproject --short-code mp --project-dir ~/demos
```

### Initialize in Existing Repository

```bash
# Initialize with Spec Kit framework
sl init --framework speckit
```

### Check Tool Status

```bash
# Check if all required tools are installed
sl doctor
```

## Commands

### Project Creation

| Command | Description |
|---------|-------------|
| `sl new` | Create a new project (interactive TUI) |
| `sl new --ci --project-name <name> --short-code <code>` | Create a project (non-interactive) |
| `sl init` | Initialize SpecLedger in an existing repository |
| `sl init --framework speckit` | Initialize with Spec Kit framework |

### Diagnostics

| Command | Description |
|---------|-------------|
| `sl doctor` | Check installation status of all required and optional tools |
| `sl doctor --json` | Get tool status in JSON format for CI/CD |

### Dependency Management

| Command | Description |
|---------|-------------|
| `sl deps list` | List all dependencies from `specledger.yaml` |
| `sl deps add <url>` | Add a dependency to `specledger.yaml` |
| `sl deps remove <url>` | Remove a dependency |
| `sl deps resolve` | Download and cache dependencies |
| `sl deps update` | Update to latest versions |

### Migration

| Command | Description |
|---------|-------------|
| `sl migrate` | Convert legacy `specledger.mod` to `specledger.yaml` |
| `sl migrate --dry-run` | Preview migration changes |

## SDD Framework Support

SpecLedger supports multiple SDD (Specification-Driven Development) frameworks. When creating a new project, you can choose:

| Framework | Description | Best For |
|-----------|-------------|----------|
| **Spec Kit** | GitHub Spec Kit - structured, phase-gated workflow with specifications, plans, and tasks | Teams that want structured development phases |
| **OpenSpec** | OpenSpec - lightweight, iterative specification workflow | Teams that prefer flexibility and iteration |
| **Both** | Use both frameworks as needed | Teams that want maximum flexibility |
| **None** | Use SpecLedger only for bootstrap and dependencies | Teams that have their own SDD process |

### What Happens When You Choose a Framework

When you create a project with `sl new --framework speckit`:

1. **Framework Installation** - Spec Kit is installed via mise
2. **Framework Initialization** - Runs `specify init --here --ai claude --force --script sh --no-git`
3. **Configuration** - Your `specledger.yaml` is updated with your framework choice
4. **Ready to Use** - Your project is ready for SDD workflows with the chosen framework

### After Framework Setup

Once your project is created, you use the framework's tools directly:

```bash
# With Spec Kit
specify spec create "User Authentication"
specify plan create
specify tasks generate
bd ready --limit 5

# With OpenSpec
openspec spec create "User Authentication"
# (OpenSpec workflows differ from Spec Kit)
```

## Project Metadata

SpecLedger projects use `specledger/specledger.yaml` for configuration:

```yaml
version: "1.0.0"

project:
  name: my-project
  short_code: mp
  created: "2026-02-05T10:30:00Z"
  modified: "2026-02-05T10:30:00Z"
  version: "0.1.0"

framework:
  choice: speckit  # speckit, openspec, both, or none
  installed_at: "2026-02-05T10:30:00Z"

task_tracker:
  choice: beads  # beads, none
  enabled_at: "2026-02-05T10:30:00Z"

dependencies:
  - url: git@github.com:org/api-spec
    branch: main
    path: spec.md
    alias: api
    resolved_commit: abc123...
    framework: speckit
    import_path: "@api"
```

## Usage Examples

### Creating a Project with Framework Setup

```bash
# Interactive - guided by TUI
sl new

# Non-interactive - with Spec Kit
sl new --ci --project-name my-api --short-code mapi --framework speckit

# Non-interactive - with OpenSpec
sl new --ci --project-name my-api --short-code mapi --framework openspec

# With custom directory
sl new --ci --project-name my-api --short-code mapi --project-dir ~/projects --framework speckit
```

### Initializing an Existing Repository

```bash
# Navigate to your repository
cd my-existing-project

# Initialize SpecLedger with Spec Kit
sl init --framework speckit --short-code ms
```

### Managing Dependencies

```bash
# Add a spec dependency (automatically detects framework type)
sl deps add git@github.com:org/auth-spec

# Add with specific branch and path
sl deps add git@github.com:org/api-spec v1.0 specledger/api.md

# Add with alias for easy reference (used for AI import paths)
sl deps add git@github.com:org/db-spec --alias db

# List all dependencies (shows detected framework and AI import path)
sl deps list

# Remove a dependency
sl deps remove git@github.com:org/auth-spec

# Download and cache all dependencies
sl deps resolve
```

#### Framework Detection and AI Import Paths

When you add a dependency, SpecLedger automatically:

1. **Detects Framework Type**: Clones the repo to identify whether it uses Spec Kit, OpenSpec, both, or none
2. **Generates Import Path**: Creates an `@alias` or `@reponame` import path for AI to reference

Example output:
```bash
$ sl deps add git@github.com:org/api-spec --alias api

Detecting Framework
───────────────────
Checking git@github.com:org/api-spec...
  Framework:  Spec Kit

✓ Dependency added
  Repository:  git@github.com:org/api-spec
  Alias:       api
  Branch:      main
  Path:        spec.md
  Framework:   Spec Kit
  Import Path: @api

Next: sl deps resolve
```

**AI Context**: The `import_path` field enables AI to reference dependencies like coding imports:
- In specifications: `See @api/spec.md for the data models`
- During generation: AI can read cached dependencies from `~/.specledger/cache/@alias/`

### Task Tracking

SpecLedger uses [Beads](https://github.com/amelie/beads) for task and issue tracking. When you create a new project, Beads is automatically configured with your project's short code.

```bash
# View all tasks
bd ls

# Show ready tasks (limit 5)
bd ready --limit 5

# Create a new task
bd new "Implement user authentication"

# Mark a task as done
bd done <task-id>
```

The task tracker configuration is stored in `specledger/specledger.yaml`:

```yaml
task_tracker:
  choice: beads  # Currently only beads is supported
  enabled_at: "2026-02-05T10:30:00Z"
```

### Checking Tool Status

```bash
# Check if all required tools are installed
sl doctor

# Get JSON output for CI/CD
sl doctor --json
```

### Migrating from Legacy Format

```bash
# Convert specledger.mod to specledger.yaml
sl migrate

# Preview migration without making changes
sl migrate --dry-run
```

## Configuration

Configuration is stored in `~/.config/specledger/config.yaml`:

```yaml
default_project_dir: ~/demos
preferred_shell: claude-code
tui_enabled: true
auto_install_deps: false
fallback_to_plain_cli: false
log_level: debug
theme: neobrutalist
language: en
```

### Configuration Locations

- **Config File**: `~/.config/specledger/config.yaml`
- **CLI Binary**: `/usr/local/bin/sl` or `$HOME/.local/bin/sl`
- **Project Metadata**: `specledger/specledger.yaml`
- **Beads Issues**: `.beads/issues.jsonl`

## Troubleshooting

### "Not a SpecLedger project" Error

This error occurs when you try to run a spec-specific command outside a project directory.

```bash
# Solution 1: Navigate to a project directory
cd ~/demos/myproject

# Solution 2: Create a new project first
sl new --ci --project-name myproject --short-code mp --framework speckit
```

### Permission Denied Errors

```bash
# If you see permission errors:
sudo bin/sl new --ci --project-name myproject --short-code mp

# Or run with sudo for the install script:
curl -fsSL https://raw.githubusercontent.com/specledger/specledger/main/scripts/install.sh | sudo bash
```

### TUI Not Working in CI/CD

The TUI requires an interactive terminal. Always use the `--ci` flag in non-interactive environments:

```bash
sl new --ci --project-name myproject --short-code mp --framework speckit
```

### Framework Not Found

If your chosen framework isn't installed after project creation:

```bash
# Check tool status
sl doctor

# Install manually via mise
mise install pipx:git+https://github.com/github/spec-kit.git
mise install npm:@fission-ai/openspec

# Initialize manually
specify init --here --ai claude --force --script sh --no-git
openspec init --force --tools claude
```

## Architecture

SpecLedger is a platform-agnostic bootstrap tool that delegates SDD workflows to external frameworks.

### Design Philosophy

1. **Framework Agnostic**: Supports multiple SDD frameworks without being tied to any single one
2. **Bootstrap Focus**: Handles project creation and initial setup only
3. **Dependency Management**: Tracks external spec dependencies
4. **Tool Detection**: Ensures required tools are installed
5. **Delegation**: All SDD workflows are handled by the chosen framework

### Tech Stack

- **Cobra**: Command-line interface framework
- **Bubble Tea**: Terminal UI library (for TUI)
- **Go Modules**: Dependency management
- **GoReleaser**: Automated releases and builds

### Project Structure

```
specledger/
├── cmd/
│   └── main.go          # CLI entry point
├── pkg/
│   ├── cli/
│   │   ├── commands/    # Command implementations
│   │   ├── config/      # Configuration management
│   │   ├── metadata/    # YAML metadata handling
│   │   ├── prerequisites/ # Tool checking
│   │   ├── tui/         # Terminal UI utilities
│   │   └── ui/          # Color and formatting utilities
│   └── ...
├── scripts/
│   └── install.sh       # Installation script
├── homebrew/            # Homebrew formula
├── Makefile             # Build automation
└── .goreleaser.yaml     # Release configuration
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a pull request

## Documentation

- **Documentation**: [https://specledger.io/docs](https://specledger.io/docs)
- **Spec Kit**: [https://github.com/github/spec-kit](https://github.com/github/spec-kit)
- **OpenSpec**: [https://github.com/fission-ai/openspec](https://github.com/fission-ai/openspec)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/specledger/specledger/issues)
- **Discussions**: [GitHub Discussions](https://github.com/specledger/specledger/discussions)
- **Website**: [https://specledger.io](https://specledger.io)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed changes.
