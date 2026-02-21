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
- **Issue Tracking** - Built-in task tracking with `sl issue` commands (no external dependencies)
- **Spec Dependencies** - Manage and track specification dependencies across projects
- **Tool Checking** - Ensures all required tools are installed and configured
- **Workflow Orchestration** - End-to-end workflows from spec to deployment

## Features

- **All-in-One SDD**: Complete Specification-Driven Development workflow built-in
- **Interactive TUI**: Create projects with a beautiful terminal interface
- **Prerequisites Checking**: Automatically detect and install required tools (mise)
- **Dependency Management**: Add, remove, and list spec dependencies
- **YAML Metadata**: Modern, human-readable project configuration with `specledger.yaml`
- **Local Caching**: Dependencies are cached locally at `~/.specledger/cache` for offline use
- **LLM Integration**: Cached specs can be easily referenced by AI agents
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Browser-Based Auth**: Secure OAuth authentication via browser with automatic token refresh

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
| `sl version` | Show version, commit, and build information |

### Dependencies

Dependencies allow you to reference external specifications from other teams or projects. When you add a dependency, SpecLedger automatically downloads and caches the specifications for offline use and AI reference.

| Command | Description |
|---------|-------------|
| `sl deps list` | List all dependencies |
| `sl deps add <url>` | Add a dependency (auto-detects SpecLedger repos) |
| `sl deps add <url> --alias <name>` | Add with alias for AI reference paths |
| `sl deps add <url> --artifact-path <path>` | Add with manual artifact path for non-SpecLedger repos |
| `sl deps add <url> --alias <name> --link` | Add and create symlink for Claude Code |
| `sl deps remove <url>` | Remove a dependency |
| `sl deps resolve` | Download and cache dependencies |
| `sl deps resolve --link` | Resolve and create symlinks for Claude Code |
| `sl deps update` | Update dependencies to latest versions |
| `sl deps link` | Manually create symlinks for all dependencies |
| `sl deps unlink [alias]` | Remove symlinks for dependencies |

**Artifact Path**: For SpecLedger repositories, the `artifact_path` is auto-detected from the dependency's `specledger.yaml`. For non-SpecLedger repositories, use `--artifact-path` to specify where specifications are located (e.g., `docs/openapi/`).

**Reference Format**: Dependencies can be referenced using the `alias:artifact` syntax in specifications. For example, if you add a dependency with `--alias api`, you can reference its artifacts as `api:spec.md` or `api:contracts/user-api.proto`.

**Linking Dependencies**: To make dependency files available for Claude Code, use the `--link` flag when adding or resolving dependencies:
```bash
sl deps add git@github.com:org/specs --alias specs --link
sl deps resolve --link
```
Or manually link all dependencies: `sl deps link`

**Unlinking**: Use `sl deps unlink [alias]` to remove symlinks. Useful for cleaning up or re-linking dependencies.

### Issue Tracking

SpecLedger includes a built-in issue tracker for managing tasks within specs. Issues are stored per-spec in `specledger/<spec>/issues.jsonl`.

| Command | Description |
|---------|-------------|
| `sl issue create --title "..." --type task` | Create a new issue |
| `sl issue list` | List issues in current spec |
| `sl issue list --all` | List issues across all specs |
| `sl issue list --tree` | Show issues as dependency tree |
| `sl issue show <id>` | Show issue details |
| `sl issue show <id> --tree` | Show issue with dependency context |
| `sl issue ready` | List issues ready to work on (not blocked) |
| `sl issue ready --all` | Ready issues across all specs |
| `sl issue ready --json` | Ready issues as JSON (for scripting) |
| `sl issue update <id> --status in_progress` | Update issue status |
| `sl issue close <id> --reason "..."` | Close an issue |
| `sl issue link <from> blocks <to>` | Add dependency |
| `sl issue unlink <from> blocks <to>` | Remove dependency |
| `sl issue migrate` | Migrate from Beads format |

**Issue IDs**: Issues use deterministic IDs in format `SL-xxxxxx` (6 hex characters derived from SHA-256 hash).

**Spec Storage**: Issues are stored per-spec to avoid merge conflicts. Use `--all` flag to work across all specs.

**Ready State**: An issue is "ready" when it has status `open` or `in_progress` AND all issues blocking it are `closed`. Use `sl issue ready` to quickly find unblocked work.

**Tree View**: Use `--tree` flag to visualize dependencies. The tree shows which issues block others, helping you understand the critical path.

### Review Comments (`sl revise`)

Fetch unresolved review comments from the SpecLedger platform and address them interactively with an AI coding agent. Requires authentication (`sl auth login`).

| Command | Description |
|---------|-------------|
| `sl revise` | Interactive: auto-detect branch, fetch comments, launch agent |
| `sl revise <branch>` | Use a specific branch name |
| `sl revise --summary` | Print compact comment listing and exit |
| `sl revise --dry-run` | Write prompt to file instead of launching agent |
| `sl revise --auto <fixture.json>` | Non-interactive fixture-driven prompt generation |

**Interactive workflow:**
1. Detect or select the target branch
2. Fetch unresolved comments from SpecLedger (issue comments + review comments)
3. Select artifacts to work on (multi-select TUI)
4. Process each comment — provide guidance or skip
5. Generate a combined revision prompt
6. Open the prompt in your editor for refinement
7. Launch the configured AI coding agent
8. Offer to commit/push changes and resolve comments

### Playbooks

| Command | Description |
|---------|-------------|
| `sl playbook list` | List available SDD playbooks |
| `sl playbook list --json` | List playbooks in JSON format |

### Authentication

| Command | Description |
|---------|-------------|
| `sl auth login` | Sign in via browser (OAuth) |
| `sl auth login --token <token>` | Authenticate with access token (CI/headless) |
| `sl auth login --refresh <token>` | Authenticate with refresh token |
| `sl auth logout` | Sign out and clear stored credentials |
| `sl auth status` | Check authentication status and token expiry |
| `sl auth refresh` | Manually refresh the access token |
| `sl auth token` | Print access token (for scripts, auto-refreshes) |
| `sl auth supabase` | Show Supabase URL and anon key |

**Authentication Flow:**

The CLI uses browser-based OAuth authentication:

1. Run `sl auth login` to start the authentication flow
2. Your browser opens to the SpecLedger sign-in page
3. Complete authentication in the browser
4. Credentials are automatically saved to `~/.specledger/credentials.json`

For CI/CD or headless environments, use token-based authentication:
```bash
sl auth login --token "$SPECLEDGER_ACCESS_TOKEN"
sl auth login --refresh "$SPECLEDGER_REFRESH_TOKEN"
```

**Environment Variables:**

| Variable | Description |
|----------|-------------|
| `SPECLEDGER_AUTH_URL` | Override the authentication URL |
| `SPECLEDGER_SUPABASE_URL` | Override the Supabase project URL |
| `SPECLEDGER_SUPABASE_ANON_KEY` | Override the Supabase anon key |
| `SPECLEDGER_ENV` | Set to `dev` or `development` for local development |

## Claude Code Slash Commands

SpecLedger provides slash commands for [Claude Code](https://claude.ai/claude-code) integration:

### Specification Workflow

| Command | Description |
|---------|-------------|
| `/specledger.adopt` | Create/update spec from feature description |
| `/specledger.specify` | Create/update feature specification |
| `/specledger.clarify` | Ask clarification questions for spec |
| `/specledger.plan` | Generate implementation plan |
| `/specledger.tasks` | Generate actionable tasks from plan |
| `/specledger.implement` | Execute tasks from tasks.md |
| `/specledger.resume` | Resume implementation from where you left off |
| `/specledger.analyze` | Cross-artifact consistency analysis |
| `/specledger.audit` | Full codebase audit with dependency graphs |
| `/specledger.checklist` | Generate a custom checklist for the current feature |

### Dependencies

| Command | Description |
|---------|-------------|
| `/specledger.add-deps` | Add a new spec dependency |
| `/specledger.remove-deps` | Remove a spec dependency |

### Project Setup

| Command | Description |
|---------|-------------|
| `/specledger.onboard` | Guided onboarding from constitution to implementation |
| `/specledger.constitution` | Create or update the project constitution |
| `/specledger.help` | Show all available SpecLedger commands |

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
