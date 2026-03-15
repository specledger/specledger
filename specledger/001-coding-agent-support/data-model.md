# Data Model: Multi-Coding Agent Support

**Feature**: 001-coding-agent-support
**Date**: 2026-03-15
**Updated**: Aligned with spec clarifications

## Entities

### AgentDefinition

Represents a coding agent that can be launched.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Name | string | Display name (e.g., "Claude Code") | Required, unique |
| Command | string | CLI command to execute (e.g., "claude") | Required, must be valid executable name |
| ConfigDir | string | Agent's config directory name (e.g., ".claude") | Required, starts with "." |
| InstallCommand | string | Install command for error messages | Required |

**State**: Static (defined in code registry)

**Relationships**: Referenced by AgentConfig.Default

**Registry**:

| Name | Command | ConfigDir | InstallCommand |
|------|---------|-----------|----------------|
| Claude Code | `claude` | `.claude/` | `npm install -g @anthropic-ai/claude-code` |
| OpenCode | `opencode` | `.opencode/` | `go install github.com/opencode-ai/opencode@latest` |
| Copilot CLI | `github-copilot` | `.github/` | `npm install -g @github/copilot` |
| Codex | `codex` | `.codex/` | `npm install -g @openai/codex` |

### AgentConfig (Extended)

Extended configuration for agent behavior with per-agent settings.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Default | string | Default agent name to launch | Must match AgentDefinition.Name or empty |
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

**Existing fields preserved**: All existing AgentConfig fields remain unchanged.

### PerAgentConfig

Per-agent configuration for arguments and environment variables.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Arguments | string | Raw arguments to pass to this agent | Free-form string |
| Env | map[string]string | Environment variables for this agent | Key-value pairs |

**Config Key Pattern**: `agent.<name>.arguments`, `agent.<name>.env`

**Examples**:
- `agent.claude.arguments: "--dangerously-skip-permissions"`
- `agent.opencode.arguments: "--model gpt-4"`
- `agent.claude.env.ANTHROPIC_API_KEY: "sk-xxx"`

### SelectedAgents

Stores which agents were selected during project setup.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Agents | []string | List of selected agent names | At least one if not "None" |

**Storage**: In constitution.md under "Agent Preferences" section

**Format**: `- **Selected Agents**: claude, opencode, github-copilot`

## Configuration Hierarchy

```
~/.specledger/config.yaml          # Global config
├── agent.default: claude          # Default agent
├── agent.claude.arguments: ""     # Claude-specific args
├── agent.opencode.arguments: ""   # OpenCode-specific args
└── agent.claude.env: {}           # Claude env vars

specledger/specledger.yaml         # Project config (git-tracked)
├── agent.default: opencode        # Project override
└── agent.opencode.arguments: "--model gpt-4"

specledger/specledger.local.yaml   # Personal config (gitignored)
└── agent.claude.env.ANTHROPIC_API_KEY: sk-xxx
```

**Merge Order**: personal-local > team-local > global > default

## Directory Structure

### After Multi-Agent Setup (macOS/Linux)

```
project/
├── .agent/
│   ├── commands/              # Shared commands (source)
│   │   └── *.md
│   └── skills/                # Shared skills (source)
│       └── *.md
├── .claude/
│   ├── commands -> ../.agent/commands    # Symlink
│   ├── skills -> ../.agent/skills        # Symlink
│   ├── settings.json
│   └── settings.local.json
├── .opencode/
│   ├── commands -> ../.agent/commands    # Symlink
│   ├── skills -> ../.agent/skills        # Symlink
│   └── manifest.yaml
└── specledger/
    └── specledger.yaml
```

### After Multi-Agent Setup (Windows)

```
project/
├── .agent/
│   ├── commands/              # Shared commands (source)
│   │   └── *.md
│   └── skills/                # Shared skills (source)
│       └── *.md
├── .claude/
│   ├── commands/              # COPY of .agent/commands
│   ├── skills/                # COPY of .agent/skills
│   ├── settings.json
│   └── settings.local.json
└── specledger/
    └── specledger.yaml
```

**Note**: Windows users must re-run `sl init --force` or manually sync when `.agent/` contents change.

### Symlink/Copy Creation Rules

1. Create `.agent/commands` and `.agent/skills` directories
2. If `.agent/` already exists: Require `--force` flag to proceed
3. **macOS/Linux**: Create symlinks
   - Move existing `.claude/commands` content to `.agent/commands` (if exists)
   - Remove old `.claude/commands` directory
   - Create symlink: `.claude/commands -> ../.agent/commands`
4. **Windows**: Copy files
   - Copy `.agent/commands/*` to `.claude/commands/`
   - No automatic sync - user must re-run or manually copy

## Validation Rules

### agent.default
- Must be empty or match a known agent name
- Case-insensitive comparison
- Empty defaults to "Claude Code"

### agent.<name>.arguments
- No validation (pass-through to agent)
- Split on spaces for CLI args (respect quoted strings)
- Only applies when launching that specific agent

### Selected Agents
- Must be from known agent list: claude, opencode, github-copilot, codex
- "None" is valid option (no agents selected)
- At least one required if not "None"

### --force Flag
- Required when `.agent/` directory exists during setup
- Error message: "Error: .agent/ exists. Use --force to overwrite."
