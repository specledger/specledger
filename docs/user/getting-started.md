# Getting Started with SpecLedger

This guide will help you install SpecLedger and create your first project.

## Prerequisites

- Go 1.24+ (if building from source)
- Git (for cloning repositories)
- Access to GitHub (for framework dependencies)

## Installation

### Homebrew (Recommended)

```bash
brew tap specledger/homebrew-specledger
brew install specledger
```

### Installation Script

```bash
# macOS/Linux
curl -fsSL https://raw.githubusercontent.com/specledger/specledger/main/scripts/install.sh | bash

# Windows (PowerShell)
irm https://raw.githubusercontent.com/specledger/specledger/main/scripts/install.ps1 | iex
```

### From Source

```bash
git clone https://github.com/specledger/specledger.git
cd specledger
make install
```

## Verify Installation

```bash
sl --version
sl doctor
```

## Create Your First Project

### Interactive Mode

```bash
sl new
```

Follow the TUI prompts to configure your project.

### Non-Interactive Mode

```bash
sl new --ci --project-name myproject --short-code mp --framework speckit
```

## Next Steps

- See [Commands Reference](commands.md) for all available commands
- See [SDD Framework](sdd-framework.md) for framework integration
- See [Contributor Setup](../contributor/setup.md) for development environment
