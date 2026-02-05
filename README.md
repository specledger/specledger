# SpecLedger

> Unified CLI for project bootstrap and specification dependency management

SpecLedger (`sl`) helps you create new projects and manage specification dependencies with a simple, intuitive CLI.

## Features

- **Interactive TUI**: Create projects with a beautiful terminal interface
- **Non-Interactive Mode**: Perfect for CI/CD and automation
- **Dependency Management**: Add, remove, and list spec dependencies
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
# Interactive mode (recommended)
sl new

# Non-interactive mode (for CI/CD)
sl new --ci --project-name myproject --short-code mp --project-dir ~/demos
```

### Manage Dependencies

```bash
# List dependencies
sl deps list

# Add a dependency
sl deps add git@github.com:org/api-spec

# Remove a dependency
sl deps remove git@github.com:org/api-spec

# Download and cache dependencies
sl deps resolve
```

## Commands

### Project Creation

| Command | Description |
|---------|-------------|
| `sl new` | Create a new project (interactive TUI) |
| `sl new --ci --project-name <name> --short-code <code>` | Create a project (non-interactive) |

### Dependency Management

| Command | Description |
|---------|-------------|
| `sl deps list` | List all dependencies |
| `sl deps add <url>` | Add a dependency |
| `sl deps remove <url>` | Remove a dependency |
| `sl deps resolve` | Download and cache dependencies |
| `sl deps update` | Update to latest versions |

### Visualization

| Command | Description |
|---------|-------------|
| `sl graph show` | Show dependency graph (coming soon) |
| `sl graph export` | Export graph to file (coming soon) |

## Usage Examples

### Creating a Project

```bash
# Interactive - guided by TUI
sl new

# Non-interactive - for CI/CD
sl new --ci --project-name my-api --short-code mapi

# With custom directory
sl new --ci --project-name my-api --short-code mapi --project-dir ~/projects
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

### Dependency Caching

Dependencies are automatically cached locally:

- Cache location: `~/.specledger/cache/`
- Cached specs can be referenced by LLMs
- Offline mode supported once cached

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
- **Beads Issues**: `.specledger/issues.jsonl`

## Troubleshooting

### "Not a SpecLedger project" Error

This error occurs when you try to run a spec-specific command outside a project directory.

```bash
# Solution 1: Navigate to a project directory
cd ~/demos/myproject

# Solution 2: Create a new project first
sl new --ci --project-name myproject --short-code mp
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
sl new --ci --project-name myproject --short-code mp
```

### Missing Dependencies

If gum or mise is missing, you can:

1. Install it manually:
   ```bash
   go install github.com/charmbracelet/gum@latest
   ```

2. Or run bootstrap without TUI:
   ```bash
   sl new --ci --project-name myproject --short-code mp
   ```

## Migration from Legacy Scripts

If you were using the old `sl` bash script or standalone `specledger` CLI:

| Legacy Command | New Command |
|----------------|-------------|
| `./sl new` | `sl new` |
| `./specledger deps list` | `sl deps list` |
| `./specledger new` | `sl new` |

### Backward Compatibility

The `specledger` command still works as an alias:

```bash
sl new      # or specledger new
sl deps list  # or specledger deps list
```

## Development

### Build for All Platforms

```bash
make build-all
```

This creates:
- `bin/sl-linux` - Linux (AMD64)
- `bin/sl-darwin` - macOS (AMD64)
- `bin/sl-windows.exe` - Windows (AMD64)
- `bin/sl-linux-arm64` - Linux (ARM64)
- `bin/sl-darwin-arm64` - macOS (ARM64)

### Run Tests

```bash
make test
```

### Format and Lint

```bash
make fmt   # Format code
make vet   # Run go vet
```

## Architecture

The CLI is built with:

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
│   │   ├── dependencies/ # Dependency registry
│   │   ├── logger/      # Logging
│   │   └── tui/         # Terminal UI utilities
│   └── ...
├── scripts/
│   └── install.sh       # Installation script
├── homebrew/            # Homebrew formula
├── Makefile             # Build automation
└── .goreleaser.yaml     # Release configuration
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- **Documentation**: [https://specledger.io/docs](https://specledger.io/docs)
- **Issues**: [GitHub Issues](https://github.com/specledger/specledger/issues)
- **Discussions**: [GitHub Discussions](https://github.com/specledger/specledger/discussions)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed changes.

## Release Notes

For the latest release notes, visit the [releases page](https://github.com/specledger/specledger/releases).
