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
# Install via one-line script (auto-detects Intel/Apple Silicon)
curl -fsSL https://raw.githubusercontent.com/specledger/specledger/main/scripts/install.sh | bash
```

The install script will:
- Auto-detect your architecture (Intel/AMD64 or Apple Silicon/ARM64)
- Download the correct binary for macOS
- Verify the checksum before installation
- Install to `~/.local/bin` (or `/usr/local/bin` with sudo)

### Homebrew (macOS)

```bash
brew tap specledger/homebrew-specledger
brew install specledger
```

### From Source

```bash
git clone https://github.com/specledger/specledger.git
cd specledger
make install
```

### Go Install

**Requires v1.0.5 or later** (earlier versions had incorrect module paths):

```bash
go install github.com/specledger/specledger/cmd/sl@latest
```

### Binary Download

Download the latest release from [GitHub Releases](https://github.com/specledger/specledger/releases/latest).

Available binaries:
- `specledger_VERSION_darwin_amd64.tar.gz` - macOS Intel
- `specledger_VERSION_darwin_arm64.tar.gz` - macOS Apple Silicon

### Troubleshooting

**Installation script fails?**
- Make sure you have `curl` or `wget` installed
- Check that `~/.local/bin` is writable or install with sudo

**Binary not found after installation?**
- Add `~/.local/bin` to your PATH or start a new shell
- For Homebrew: `brew info specledger` to see installation location

**Go install fails?**
- Make sure you have Go 1.24+ installed
- Check that `$GOPATH/bin` or `$GOBIN` is in your PATH
- Ensure you're using v1.0.5 or later: `go install github.com/specledger/specledger/cmd/sl@v1.0.5`
- Older versions (v1.0.1-v1.0.4) installed binary as 'cmd' instead of 'sl'

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

Dependencies allow you to reference external specifications from other teams or projects. When you add a dependency, SpecLedger automatically downloads and caches the specifications for offline use and AI reference.

| Command | Description |
|---------|-------------|
| `sl deps list` | List all dependencies |
| `sl deps add <url>` | Add a dependency (auto-detects SpecLedger repos) |
| `sl deps add <url> --alias <name>` | Add with alias for AI reference paths |
| `sl deps add <url> --artifact-path <path>` | Add with manual artifact path for non-SpecLedger repos |
| `sl deps remove <url>` | Remove a dependency |
| `sl deps resolve` | Download and cache dependencies |
| `sl deps update` | Update dependencies to latest versions |
| `sl deps link` | Manually create symlinks (auto-linked on add/resolve) |

**Artifact Path**: For SpecLedger repositories, the `artifact_path` is auto-detected from the dependency's `specledger.yaml`. For non-SpecLedger repositories, use `--artifact-path` to specify where specifications are located (e.g., `docs/openapi/`).

**Reference Format**: Dependencies can be referenced using the `alias:artifact` syntax in specifications. For example, if you add a dependency with `--alias api`, you can reference its artifacts as `api:spec.md` or `api:contracts/user-api.proto`.

**Auto-Linking**: Dependencies are automatically linked when added or resolved. Symlinks are created from `~/.specledger/cache/<alias>/` to `specledger/<alias>/` making files available for Claude Code. If a conflict exists (non-empty directory), linking is skipped to avoid data loss - use `sl deps link --force` to override.

### Workflows

| Command | Description |
|---------|-------------|
| `sl playbook` | Run SDD playbook workflows |
| `sl graph show` | Show dependency graph |
| `sl graph export` | Export dependency graph |

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
