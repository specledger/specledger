# SpecLedger

> Unified CLI for project bootstrap and specification dependency management

SpecLedger provides a single CLI tool that integrates project bootstrap (with interactive TUI) and specification dependency management. Use `sl` for all operations, with full backward compatibility with the legacy `specledger` alias.

## Features

- **Interactive TUI**: Beautiful terminal user interface for project bootstrap
- **Non-Interactive Mode**: Perfect for CI/CD and automated workflows
- **Specification Dependency Management**: Track and manage spec dependencies
- **Multiple Distribution Channels**: Install from GitHub Releases, build from source, or use package managers
- **Cross-Platform**: Supports Linux, macOS, and Windows
- **Clear Error Messages**: Actionable suggestions when things go wrong

## Installation

### From GitHub Releases

#### macOS

```bash
# Using curl
curl -fsSL https://raw.githubusercontent.com/your-org/specledger/main/scripts/install.sh | bash

# Or using wget
wget -qO- https://raw.githubusercontent.com/your-org/specledger/main/scripts/install.sh | bash
```

#### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/your-org/specledger/main/scripts/install.sh | sudo bash

# Or using wget
wget -qO- https://raw.githubusercontent.com/your-org/specledger/main/scripts/install.sh | sudo bash
```

#### Windows

```powershell
# Using PowerShell
irm https://raw.githubusercontent.com/your-org/specledger/main/scripts/install.ps1 | iex
```

### From Source

```bash
# Clone the repository
git clone https://github.com/your-org/specledger.git
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
brew tap your-org/homebrew-specledger
brew install specledger
```

#### npm / npx

```bash
npx @specledger/cli@latest --help
```

## Quick Start

### Interactive Bootstrap

```bash
sl new
```

This starts an interactive TUI that will guide you through:

1. Project name
2. Short code (for Beads prefix)
3. Playbook selection
4. Agent shell selection

### Non-Interactive Bootstrap

```bash
sl new --ci --project-name myproject --short-code mp
```

### List Dependencies

```bash
cd myproject
sl deps list
```

### Update Dependencies

```bash
sl deps update
```

### Check for Conflicts

```bash
sl deps conflict
```

## Commands

### Bootstrap Commands

| Command | Description |
|---------|-------------|
| `sl new` | Start interactive TUI for project bootstrap |
| `sl new --project-name <name> --short-code <code>` | Non-interactive bootstrap |

### Dependency Management Commands

| Command | Description |
|---------|-------------|
| `sl deps list` | List all specification dependencies |
| `sl deps add <spec>` | Add a specification dependency |
| `sl deps remove <spec>` | Remove a specification dependency |
| `sl deps update` | Update dependencies to latest compatible versions |
| `sl deps conflict` | Check for dependency conflicts |
| `sl deps vendor` | Vendor dependencies for offline use |

### Reference Management Commands

| Command | Description |
|---------|-------------|
| `sl refs validate` | Validate external specification references |

### Graph Commands

| Command | Description |
|---------|-------------|
| `sl graph deps` | Display dependency graph |
| `sl graph refs` | Display reference graph |

### Other Commands

| Command | Description |
|---------|-------------|
| `sl conflict` | Check for dependency conflicts |
| `sl update` | Update dependencies to latest compatible versions |
| `sl vendor` | Vendor dependencies for offline use |
| `sl --help` | Show help for all commands |
| `sl --version` | Print version information |

## Usage Examples

### Creating a New Project

#### Interactive (Recommended)

```bash
sl new
# Follow the prompts in the TUI
```

#### Using Flags (CI/CD)

```bash
sl new --ci \
  --project-name my-api \
  --short-code mapi \
  --playbook "General SWE" \
  --shell claude-code
```

### Managing Dependencies

```bash
# List current dependencies
sl deps list

# Add a new dependency
sl deps add github.com/org/project-spec

# Remove a dependency
sl deps remove github.com/org/project-spec

# Update all dependencies
sl deps update

# Check for conflicts
sl deps conflict
```

### Viewing Dependency Graphs

```bash
# Show dependency graph
sl graph deps

# Show reference graph
sl graph refs
```

### Using Non-Interactive Mode

Perfect for CI/CD pipelines:

```bash
# In CI/CD environment, always use --ci flag
sl new --ci --project-name ci-project --short-code ci
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
curl -fsSL https://raw.githubusercontent.com/your-org/specledger/main/scripts/install.sh | sudo bash
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

- **Documentation**: [docs.specledger.dev](https://docs.specledger.dev)
- **Issues**: [GitHub Issues](https://github.com/your-org/specledger/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/specledger/discussions)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed changes.

## Release Notes

For the latest release notes, visit the [releases page](https://github.com/your-org/specledger/releases).
