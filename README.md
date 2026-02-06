# SpecLedger

> Unified CLI for project bootstrap with Spec Kit SDD framework

SpecLedger (`sl`) helps you create new projects with [Spec Kit](https://github.com/github/spec-kit) for specification-driven development. It provides a simple, intuitive CLI for project initialization and dependency management.

## Features

- **Interactive TUI**: Create projects with a beautiful terminal interface
- **Spec Kit Integration**: Built-in support for GitHub Spec Kit SDD framework
- **Prerequisites Checking**: Automatically detect and install required tools (mise, bd, perles)
- **Dependency Management**: Add, remove, and list spec dependencies
- **YAML Metadata**: Modern, human-readable project configuration with `specledger.yaml`
- **Migration Support**: Convert legacy `.mod` files to YAML with `sl migrate`
- **Local Caching**: Dependencies are cached locally at `~/.specledger/cache` for offline use
- **LLM Integration**: Cached specs can be easily referenced by AI agents
- **Cross-Platform**: Works on Linux, macOS, and Windows

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
# Interactive mode (recommended) - includes Spec Kit setup
sl new

# Non-interactive mode (for CI/CD)
sl new --ci --project-name myproject --short-code mp --project-dir ~/demos --framework speckit
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

## Spec Kit Integration

SpecLedger is designed to work seamlessly with [GitHub Spec Kit](https://github.com/github/spec-kit) for specification-driven development.

### What Spec Kit Provides

- **Structured Workflows**: Specifications → Plans → Tasks → Implementation
- **AI Integration**: Built-in support for Claude, ChatGPT, and other AI assistants
- **Phase Gates**: Clear transitions between specification, planning, and implementation
- **Artifact Management**: Track all project artifacts with git-based resolution

### Automatic Setup

When you create a project with `sl new --framework speckit`, SpecLedger:

1. Installs Spec Kit via mise
2. Initializes it with `specify init --here --ai claude --force --script sh --no-git`
3. Updates your `specledger.yaml` with the framework choice

### Using Spec Kit

After initialization:

```bash
# Create a new specification
specify spec create "User Authentication"

# Create a plan from specs
specify plan create

# Generate and track tasks
specify tasks generate
bd ready --limit 5
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

dependencies:
  - url: git@github.com:org/api-spec
    branch: main
    path: spec.md
    alias: api
    resolved_commit: abc123...
```

## Usage Examples

### Creating a Project with Spec Kit

```bash
# Interactive - guided by TUI (includes Spec Kit setup)
sl new

# Non-interactive - for CI/CD
sl new --ci --project-name my-api --short-code mapi --framework speckit

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
# Add a spec dependency
sl deps add git@github.com:org/auth-spec

# Add with specific branch and path
sl deps add git@github.com:org/api-spec v1.0 specs/api.md

# Add with alias for easy reference
sl deps add git@github.com:org/db-spec --alias db

# List all dependencies
sl deps list

# Remove a dependency
sl deps remove git@github.com:org/auth-spec

# Download and cache all dependencies
sl deps resolve
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

### Spec Kit Not Found

If Spec Kit isn't installed after project creation:

```bash
# Check tool status
sl doctor

# Install manually via mise
mise install pipx:git+https://github.com/github/spec-kit.git

# Initialize Spec Kit
specify init --here --ai claude --force --script sh --no-git
```

## Architecture

SpecLedger is a thin wrapper that delegates SDD workflows to external frameworks like [Spec Kit](https://github.com/github/spec-kit).

### Design Philosophy

1. **Framework Agnostic**: Supports multiple SDD frameworks (Spec Kit, OpenSpec)
2. **Bootstrap Focus**: Handles project creation and initial setup
3. **Dependency Management**: Tracks external spec dependencies
4. **Tool Detection**: Ensures required tools are installed
5. **Delegation**: Actual SDD workflows are handled by chosen frameworks

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

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/specledger/specledger/issues)
- **Discussions**: [GitHub Discussions](https://github.com/specledger/specledger/discussions)
- **Website**: [https://specledger.io](https://specledger.io)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed changes.
