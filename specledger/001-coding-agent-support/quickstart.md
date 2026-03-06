# Quickstart: Multi-Coding Agent Support

**Feature**: 001-coding-agent-support
**Date**: 2026-03-07

## Prerequisites

- SpecLedger CLI (`sl`) installed
- At least one coding agent installed (Claude Code, OpenCode, or Codex)

## Basic Usage

### Launch Default Agent

```bash
# Launch the configured default agent
sl code

# If no default configured, launches Claude Code
```

### Launch Specific Agent

```bash
# Launch Claude Code
sl code claude

# Launch OpenCode
sl code opencode

# Launch Codex
sl code codex
```

## Configuration

### Set Default Agent (Global)

```bash
# Set Claude Code as default across all projects
sl config set --global agent.default claude

# Set OpenCode as default
sl config set --global agent.default opencode
```

### Set Default Agent (Project)

```bash
# Set project-specific default
sl config set agent.default opencode
```

### Pass Custom Arguments

```bash
# Configure arguments to pass to agent
sl config set agent.arguments "--dangerously-skip-permissions"

# Multiple arguments
sl config set agent.arguments "--dangerously-skip-permissions --verbose"
```

### Set Environment Variables

```bash
# Set custom environment variable
sl config set agent.env.CUSTOM_VAR value

# Set model override
sl config set agent.model opus
```

## Project Setup with Multiple Agents

### New Project

```bash
sl new
# When prompted for "AI Coding Agent":
# - Use Space to toggle multiple agents
# - Press Enter to confirm selection
```

### Existing Project

```bash
cd your-project
sl init
# Select multiple agents when prompted
```

### What Gets Created

When you select multiple agents:

```
your-project/
├── .agent/
│   ├── commands/    # Shared commands
│   └── skills/      # Shared skills
├── .claude/
│   ├── commands -> ../.agent/commands
│   └── skills -> ../.agent/skills
└── .opencode/
    ├── commands -> ../.agent/commands
    └── skills -> ../.agent/skills
```

## View Current Configuration

```bash
# Show all agent settings
sl config show

# Get specific value
sl config get agent.default
sl config get agent.arguments
```

## Troubleshooting

### Agent Not Found

```
Error: claude is not installed.
Install Claude Code: npm install -g @anthropic-ai/claude-code
```

Install the missing agent and try again.

### Symlink Issues on Windows

Windows requires Developer Mode or admin privileges for symlinks. If symlinks fail:
- Enable Developer Mode in Windows Settings
- Or run terminal as Administrator

### Config Not Applied

Check configuration precedence:
```bash
sl config show
# Values show their source: [global], [local], [personal]
```

Project settings override global settings. Personal-local (gitignored) overrides team-local.
