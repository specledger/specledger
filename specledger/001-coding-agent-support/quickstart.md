# Quickstart: Multi-Coding Agent Support

**Feature**: 001-coding-agent-support
**Date**: 2026-03-15
**Updated**: Aligned with spec clarifications

## Prerequisites

- SpecLedger CLI (`sl`) installed
- At least one coding agent installed (Claude Code, OpenCode, Copilot CLI, or Codex)

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

# Launch Copilot CLI
sl code github-copilot

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

### Pass Custom Arguments (Per-Agent)

```bash
# Configure arguments for Claude Code
sl config set agent.claude.arguments "--dangerously-skip-permissions"

# Configure arguments for OpenCode
sl config set agent.opencode.arguments "--model gpt-4"

# Multiple arguments
sl config set agent.claude.arguments "--dangerously-skip-permissions --verbose"
```

### Set Environment Variables

```bash
# Set environment variable for Claude
sl config set agent.claude.env.ANTHROPIC_API_KEY "sk-xxx"

# Set environment variable for OpenCode
sl config set agent.opencode.env.OPENAI_API_KEY "sk-xxx"
```

## Project Setup with Multiple Agents

### New Project

```bash
sl new
# When prompted for "AI Coding Agent":
# - Use Space to toggle multiple agents
# - Options: Claude Code, OpenCode, Copilot CLI, Codex
# - Press Enter to confirm selection
```

### Existing Project

```bash
cd your-project
sl init
# Select multiple agents when prompted
```

### Force Overwrite

If `.agent/` directory already exists:

```bash
sl init --force
# This will overwrite existing .agent/ directory
```

### What Gets Created

**macOS/Linux** (symlinks):

```
your-project/
├── .agent/
│   ├── commands/    # Shared commands (source)
│   └── skills/      # Shared skills (source)
├── .claude/
│   ├── commands -> ../.agent/commands
│   └── skills -> ../.agent/skills
└── .opencode/
    ├── commands -> ../.agent/commands
    └── skills -> ../.agent/skills
```

**Windows** (copies):

```
your-project/
├── .agent/
│   ├── commands/    # Shared commands (source)
│   └── skills/      # Shared skills (source)
├── .claude/
│   ├── commands/    # COPY of .agent/commands
│   └── skills/      # COPY of .agent/skills
└── .opencode/
    ├── commands/    # COPY of .agent/commands
    └── skills/      # COPY of .agent/skills
```

## View Current Configuration

```bash
# Show all agent settings
sl config show

# Get specific value
sl config get agent.default
sl config get agent.claude.arguments
```

## Troubleshooting

### Agent Not Found

```
Error: 'claude' not found. Install: npm install -g @anthropic-ai/claude-code
```

Install the missing agent using the command shown and try again.

**Install commands by agent**:

| Agent | Install Command |
|-------|-----------------|
| Claude Code | `npm install -g @anthropic-ai/claude-code` |
| OpenCode | `go install github.com/opencode-ai/opencode@latest` |
| Copilot CLI | `npm install -g @github/copilot` |
| Codex | `npm install -g @openai/codex` |

### .agent/ Directory Already Exists

```
Error: .agent/ exists. Use --force to overwrite.
```

Add the `--force` flag to proceed:
```bash
sl init --force
```

### Windows: Changes Not Reflected

On Windows, files are copied (not symlinked). If you update `.agent/commands/`, changes won't automatically appear in `.claude/commands/`.

**Solution**: Re-run setup or manually copy files:
```bash
# Re-run setup
sl init --force

# Or manually copy
xcopy /E /Y .agent\commands .claude\commands
```

### Config Not Applied

Check configuration precedence:
```bash
sl config show
# Values show their source: [global], [local], [personal]
```

Project settings override global settings. Personal-local (gitignored) overrides team-local.
