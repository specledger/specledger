# SpecLedger

[![Build Status](https://img.shields.io/github/actions/workflow/status/specledger/specledger/ci.yml?branch=main)](https://github.com/specledger/specledger/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/specledger/specledger)](https://goreportcard.com/report/github.com/specledger/specledger)
[![Coverage](https://img.shields.io/codecov/c/github/specledger/specledger)](https://codecov.io/gh/specledger/specledger)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/github/v/release/specledger/specledger)](https://github.com/specledger/specledger/releases)

> All-in-one SDD Playbook for modern development teams

SpecLedger (`sl`) is a comprehensive Specification-Driven Development playbook that unifies project creation, customizable workflows, issue tracking, and specification dependency management.

**Documentation**: [https://specledger.io/docs](https://specledger.io/docs) | **Website**: [https://specledger.io](https://specledger.io)

## What is SpecLedger?

SpecLedger is an **all-in-one SDD playbook** that provides:

- **Easy Bootstrap** - Create new projects with a single command
- **Customizable Playbooks** - Support for multiple SDD playbook workflows
- **Issue Tracking** - Built-in task tracking with Beads integration
- **Spec Dependencies** - Manage and track specification dependencies across projects
- **Tool Checking** - Ensures all required tools are installed and configured
- **Workflow Orchestration** - End-to-end workflows from spec to deployment

## Features

- **All-in-One SDD**: Complete Specification-Driven Development workflow built-in
- **Interactive TUI**: Create projects with a beautiful terminal interface
- **Prerequisites Checking**: Automatically detect and install required tools (mise, bd, perles)
- **Dependency Management**: Add, remove, and list spec dependencies
- **YAML Metadata**: Modern, human-readable project configuration with `specledger.yaml`
- **Local Caching**: Dependencies are cached locally at `~/.specledger/cache` for offline use
- **LLM Integration**: Cached specs can be easily referenced by AI agents
- **Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### Quick Install (Recommended)

```bash
# Install via Homebrew
brew tap specledger/homebrew-specledger
brew install specledger

# Or install via one-line script
curl -fsSL https://raw.githubusercontent.com/specledger/specledger/main/scripts/install.sh | bash
```

### From Source

```bash
git clone https://github.com/specledger/specledger.git
cd specledger
make install
```

### Package Managers

```bash
# Homebrew
brew tap specledger/homebrew-specledger
brew install specledger

# npm / npx
npx @specledger/cli@latest --help
```

## Quick Start

```bash
# Create a new project (interactive mode)
sl new

# Create a project (non-interactive mode)
sl new --ci --project-name myproject --short-code mp

# Initialize in existing repository
sl init

# Check required tools
sl doctor

# Manage dependencies
sl deps add git@github.com:org/api-spec
sl deps list
sl deps resolve
```

## Commands

### Project Creation

| Command | Description |
|---------|-------------|
| `sl new` | Create a new project (interactive TUI) |
| `sl new --ci --project-name <name> --short-code <code>` | Create a project (non-interactive) |
| `sl init` | Initialize SpecLedger in an existing repository |

### Diagnostics

| Command | Description |
|---------|-------------|
| `sl doctor` | Check installation status of all required tools |
| `sl doctor --json` | Get tool status in JSON format for CI/CD |

### Dependencies

| Command | Description |
|---------|-------------|
| `sl deps list` | List all dependencies |
| `sl deps add <url>` | Add a dependency |
| `sl deps add <url> --alias <name>` | Add with alias for AI import paths |
| `sl deps remove <url>` | Remove a dependency |
| `sl deps resolve` | Download and cache dependencies |

### Workflows

| Command | Description |
|---------|-------------|
| `sl playbook` | Run SDD playbook workflows |
| `sl graph` | Show dependency graph |
| `sl refs` | Manage reference resolution |

### Utilities

| Command | Description |
|---------|-------------|
| `sl migrate` | Convert legacy `specledger.mod` to `specledger.yaml` |
| `sl vendor list` | List vendored dependencies |
| `sl vendor add <url>` | Add a vendored dependency |

## Documentation

Full documentation is available at [https://specledger.io/docs](https://specledger.io/docs)

- **Getting Started**: Installation and first project setup
- **User Guide**: Complete command reference and workflows
- **Contributing**: Development setup and contribution guidelines
- **Governance**: Project governance and decision-making

## Tech Stack

- **Go 1.24+** - Core language
- **Cobra** - Command-line interface
- **Bubble Tea** - Terminal UI
- **go-git** - Git operations
- **YAML v3** - Configuration parsing

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- **Documentation**: [https://specledger.io/docs](https://specledger.io/docs)
- **Issues**: [GitHub Issues](https://github.com/specledger/specledger/issues)
- **Discussions**: [GitHub Discussions](https://github.com/specledger/specledger/discussions)
- **Website**: [https://specledger.io](https://specledger.io)
