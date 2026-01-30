# Quickstart: CLI Unification

**Feature**: CLI Unification (003-cli-unification)
**Date**: 2026-01-30

## Overview

This quickstart guide demonstrates how to use the unified SpecLedger CLI tool. The CLI provides both project bootstrap and specification dependency management in a single tool.

## Prerequisites

- Go 1.21+ (for building from source)
- git (for cloning repositories)
- curl or wget (for downloading releases)
- bash shell (for Makefile)

## Installation

### Method 1: GitHub Releases (Recommended)

**macOS**:
```bash
curl -sSL https://github.com/org/specledger/releases/latest/download/install.sh | sh
```

**Linux**:
```bash
curl -sSL https://github.com/org/specledger/releases/latest/download/install.sh | sh
```

**Windows**:
```powershell
powershell -c "Invoke-WebRequest -Uri 'https://github.com/org/specledger/releases/latest/download/install.ps1' -OutFile 'install.ps1'; .\install.ps1"
```

### Method 2: Build from Source

```bash
git clone https://github.com/org/specledger.git
cd specledger
make build
./bin/sl --version
```

### Method 3: Self-Hosted Binary

```bash
# Download from GitHub Releases
curl -LO https://github.com/org/specledger/releases/latest/download/sl-darwin-arm64.tar.gz

# Extract
tar -xzf sl-darwin-arm64.tar.gz

# Run (binary works from any directory)
./sl --help
```

### Method 4: Homebrew (macOS)

```bash
brew install specledger
```

### Method 5: npx (JavaScript/TypeScript users)

```bash
npx @specledger/cli new --project-name myproject --short-code myp
```

---

## Quick Start

### 1. Bootstrap a New Project

**Interactive Mode** (TUI):
```bash
sl new
```

This will start an interactive TUI prompting you for:
- Project name
- Short code (Beads prefix)
- Playbook type
- Agent shell

**Non-Interactive Mode** (CI/CD):
```bash
sl new --project-name myproject --short-code myp
```

After successful bootstrap:
```
✓ Project created: ~/demos/myproject
✓ Beads prefix: myp

Next steps:
  cd ~/demos/myproject
  claude
```

### 2. Manage Dependencies

**Add a dependency**:
```bash
cd ~/demos/myproject
sl deps add git@github.com:org/spec.git
```

**List all dependencies**:
```bash
sl deps list
```

**Resolve dependencies**:
```bash
sl deps resolve
```

**Update dependencies**:
```bash
sl deps update
```

**Remove a dependency**:
```bash
sl deps remove git@github.com:org/spec.git specs/spec.md
```

### 3. View Dependency Graph

```bash
sl graph show
```

### 4. Validate References

```bash
sl refs validate
```

---

## Common Use Cases

### Create a New Project in CI/CD

```yaml
# .github/workflows/spec-ledger.yml
name: Bootstrap SpecLedger Project

on: [workflow_dispatch]

jobs:
  bootstrap:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Download CLI
        run: |
          curl -LO https://github.com/org/specledger/releases/latest/download/sl-linux-amd64.tar.gz
          tar -xzf sl-linux-amd64.tar.gz
          sudo mv sl /usr/local/bin/
      - name: Create Project
        run: sl new --project-name myproject --short-code myp
```

### Update All Dependencies

```bash
# Update all dependencies
sl deps update

# Or update specific repository
sl deps update git@github.com:org/specific.git
```

### Vendor Specifications for Offline Use

```bash
sl vendor add git@github.com:org/spec.git
```

### Check for CLI Updates

```bash
sl update self
```

---

## User Stories Walkthrough

### User Story 1: Single Unified CLI Tool

**Scenario**: You want to use SpecLedger for both project setup and dependency management.

**Steps**:
```bash
# Bootstrap new project
sl new --project-name myproject --short-code myp

# Add dependencies
cd ~/demos/myproject
sl deps add git@github.com:org/spec-a.git
sl deps add git@github.com:org/spec-b.git

# List dependencies
sl deps list

# View graph
sl graph show
```

**Result**: All commands use the same `sl` binary - no need to learn multiple tools.

---

### User Story 2: GitHub Releases

**Scenario**: You want to install the CLI without manual compilation.

**Steps** (macOS):
```bash
# Run installer
curl -sSL https://github.com/org/specledger/releases/latest/download/install.sh | sh

# Verify installation
sl --version
# Output: sl version 1.0.0
```

**Result**: CLI installed from official GitHub releases.

---

### User Story 3: Self-Built Binaries

**Scenario**: You want to build and install locally.

**Steps**:
```bash
# Clone and build
git clone https://github.com/org/specledger.git
cd specledger
make build

# Run the built binary
./bin/sl --version
# Output: sl version 1.0.0
```

**Result**: Binary built from source works identically to released version.

---

### User Story 4: Self-Hosted / Local Binaries

**Scenario**: You want to run CLI from a non-PATH location.

**Steps**:
```bash
# Download binary
curl -LO https://github.com/org/specledger/releases/latest/download/sl-darwin-arm64.tar.gz

# Extract
tar -xzf sl-darwin-arm64.tar.gz

# Run from current directory
./sl --help
./sl new --project-name myproject --short-code myp
```

**Result**: Binary works without installing to PATH.

---

### User Story 5: UVX Style

**Scenario**: You want to try the CLI without installation.

**Steps**:
```bash
# Run standalone
./sl --help
# (assuming standalone executable is pre-provided)
```

**Result**: Can execute CLI directly without setup.

---

### User Story 6: Package Manager Integration

**Scenario**: You prefer using Homebrew for installation.

**Steps** (macOS):
```bash
# Install via Homebrew
brew install specledger

# Verify
sl --version
# Output: sl version 1.0.0
```

**Result**: CLI installed via package manager with standard management.

---

## Advanced Features

### Non-Interactive Bootstrap

```bash
sl new --project-name myproject --short-code myp --playbook default --shell claude-code
```

### Dry Run Mode

```bash
sl deps add git@github.com:org/spec.git --dry-run
```

### Custom Configuration

```bash
# Edit user config
nano ~/.config/specledger/config.yaml

# Configuration file location:
# - Linux/Mac: ~/.config/specledger/config.yaml
# - Windows: %USERPROFILE%\.config\specledger\config.yaml
```

---

## Troubleshooting

### TUI Not Working

If gum is not found:
```
ERROR: gum not found

Install from:
  macOS:  brew install gum
  Linux:  go install github.com/charmbracelet/gum@latest

You can continue without TUI using --ci flag.
```

**Solution 1**: Install gum
```bash
brew install gum
```

**Solution 2**: Use non-interactive mode
```bash
sl new --ci --project-name myproject --short-code myp
```

### Permission Denied

If you can't write to ~/demos:
```
ERROR: Permission denied: ~/demos/myproject

Please choose a different location.
```

**Solution**: Use current directory or provide full path
```bash
sl new --project-dir ./myproject --project-name myproject --short-code myp
```

### Not a SpecLedger Project

If you run deps commands outside a project:
```
ERROR: Not a SpecLedger project

This command requires running from within a SpecLedger project.

Expected to find .specledger directory.
```

**Solution**: Navigate to a SpecLedger project
```bash
cd ~/demos/myproject
sl deps list
```

---

## Next Steps

1. **Learn more**: See [CLAUDE.md](../../CLAUDE.md) for best practices
2. **Read documentation**: See [AGENTS.md](../../AGENTS.md) for workflow details
3. **Get help**: Run `sl --help` for command reference

---

## Support

- **GitHub Issues**: https://github.com/org/specledger/issues
- **Documentation**: https://github.com/org/specledger/wiki
- **Community**: Join the Discord server

---

## Version Compatibility

| CLI Version | Features |
|-------------|----------|
| 1.0.0 | Initial release with bootstrap, deps, refs, graph, vendor, conflict, update commands |

---

## Migration Guide

### From Old `sl` Script to Unified CLI

**Before**:
```bash
./sl  # Bootstrap script
specledger deps list  # Separate CLI
```

**After**:
```bash
sl new  # Unified CLI for everything
sl deps list  # Same command, single binary
```

**Backward Compatibility**:
- `specledger` command still works as an alias to `sl`
- All existing scripts continue to work

---

## Summary

The unified CLI provides:
- ✅ Single binary for all SpecLedger operations
- ✅ Interactive TUI for project bootstrap
- ✅ Non-interactive mode for CI/CD
- ✅ Multiple distribution methods (GitHub, source, self-hosted, package managers)
- ✅ Complete dependency management
- ✅ Backward compatibility with existing `sl` script

Start by running `sl new --help` to explore all options.
