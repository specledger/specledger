# Data Model: Multi-Coding Agent Support

**Feature**: 001-coding-agent-support
**Date**: 2026-03-07

## Entities

### AgentDefinition

Represents a coding agent that can be launched.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Name | string | Display name (e.g., "Claude Code") | Required, unique |
| Command | string | CLI command to execute (e.g., "claude") | Required, must be valid executable name |
| ConfigDir | string | Agent's config directory name (e.g., ".claude") | Required, starts with "." |
| InstallInstructions | string | How to install the agent | Required |

**State**: Static (defined in code registry)

**Relationships**: Referenced by AgentConfig.Default

### AgentConfig (Extended)

Extended configuration for agent behavior.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Default | string | Default agent name to launch | Must match AgentDefinition.Name or empty |
| Arguments | string | Raw arguments to pass to agent | Free-form string |
| BaseURL | string | API endpoint URL | Valid URL or empty |
| AuthToken | string | Auth token | Sensitive |
| APIKey | string | API key | Sensitive |
| Model | string | Default model name | - |
| ModelSonnet | string | Sonnet model override | - |
| ModelOpus | string | Opus model override | - |
| ModelHaiku | string | Haiku model override | - |
| SubagentModel | string | Model for subagents | - |
| Provider | string | Provider selection | Enum: anthropic, bedrock, vertex |
| PermissionMode | string | Permission mode | Enum: default, plan, bypassPermissions, acceptEdits, dontAsk |
| Effort | string | Effort level | Enum: low, medium, high |
| AllowedTools | []string | Tools allowed without prompts | - |
| Env | map[string]string | Environment variables | Key-value pairs |

**Existing fields preserved**: All existing AgentConfig fields remain unchanged.

**New fields**:
- `Default` - Determines which agent `sl code` launches when no argument provided
- `Arguments` - Generic argument string passed directly to agent CLI

### SelectedAgents

Stores which agents were selected during project setup.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Agents | []string | List of selected agent names | At least one if not "None" |

**Storage**: In constitution.md under "Agent Preferences" section

**Format**: `- **Selected Agents**: claude, opencode`

## Configuration Hierarchy

```
~/.specledger/config.yaml          # Global config
├── agent.default: claude          # Default agent
├── agent.arguments: ""            # Global arguments
└── agent.env: {}                  # Global env vars

specledger/specledger.yaml         # Project config (git-tracked)
├── agent.default: opencode        # Project override
└── agent.arguments: "--verbose"   # Project arguments

specledger/specledger.local.yaml   # Personal config (gitignored)
└── agent.auth-token: sk-xxx       # Sensitive values
```

**Merge Order**: personal-local > team-local > global > default

## Directory Structure

### After Multi-Agent Setup

```
project/
├── .agent/
│   ├── commands/              # Shared commands (source)
│   │   └── *.md
│   └── skills/                # Shared skills (source)
│       └── *.md
├── .claude/
│   ├── commands -> ../.agent/commands
│   ├── skills -> ../.agent/skills
│   ├── settings.json
│   └── settings.local.json
├── .opencode/
│   ├── commands -> ../.agent/commands
│   ├── skills -> ../.agent/skills
│   └── manifest.yaml
└── specledger/
    └── specledger.yaml
```

### Symlink Creation Rules

1. Create `.agent/commands` and `.agent/skills` directories
2. Move existing `.claude/commands` content to `.agent/commands` (if exists)
3. Remove old `.claude/commands` directory
4. Create symlink: `.claude/commands -> ../.agent/commands`
5. Repeat for skills and other selected agents

## Validation Rules

### agent.default
- Must be empty or match a known agent name
- Case-insensitive comparison
- Empty defaults to "Claude Code"

### agent.arguments
- No validation (pass-through to agent)
- Split on spaces for CLI args (respect quoted strings)

### Selected Agents
- Must be from known agent list
- "None" is valid option (no agents selected)
- At least one required if not "None"
