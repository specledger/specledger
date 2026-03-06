# Research: Multi-Coding Agent Support

**Feature**: 001-coding-agent-support
**Date**: 2026-03-07
**Status**: Complete

## Prior Work

### Related Codebase Components

| Component | Location | Relevance |
|-----------|----------|-----------|
| Agent Launcher | `pkg/cli/launcher/launcher.go` | Existing Claude Code launcher - extend for multi-agent |
| Config Schema | `pkg/cli/config/schema.go` | Agent config keys - extend with new fields |
| TUI Agent Selection | `pkg/cli/tui/sl_new.go`, `sl_init.go` | Single-select agent - change to multi-select |
| Bootstrap Helpers | `pkg/cli/commands/bootstrap_helpers.go` | `launchAgent()` function - reuse pattern |

### Related Issues

No existing issues reference multi-agent support. This is a new feature.

## Technology Decisions

### Agent Registry Design

**Decision**: Static registry with agent definitions in code

**Rationale**:
- Agents are relatively stable (Claude Code, OpenCode, Codex)
- No need for dynamic plugin system
- Simpler to maintain and test
- Consistent with existing `launcher.DefaultAgents` pattern

**Alternatives Considered**:
1. **Plugin-based registry** - Rejected: Over-engineered for 3-4 known agents
2. **Config-file based registry** - Rejected: Adds complexity, users can't add custom agents easily anyway

### Config Key Design

**Decision**: Add `agent.default` (string) and `agent.arguments` (string) to existing schema

**Rationale**:
- Minimal changes to existing config system
- `agent.arguments` as string allows any CLI args without schema changes
- Consistent with existing `agent.*` namespace

**Alternatives Considered**:
1. **Per-agent config sections** (`agent.claude.arguments`) - Rejected: Over-complicates for single active agent use case
2. **Individual flag keys** (`agent.skip-permissions`) - Rejected: User requested generic `agent.arguments` instead

### Symlink Strategy

**Decision**: Create `.agent/` as source of truth, symlink agent-specific dirs to it

**Rationale**:
- Single source of truth for commands/skills
- Existing pattern already in codebase (`.opencode/commands -> ../.claude/commands`)
- Easy to add new agents without duplicating files

**Structure**:
```
.agent/
├── commands/     # Shared commands
└── skills/       # Shared skills

.claude/
├── commands -> ../.agent/commands
└── skills -> ../.agent/skills

.opencode/
├── commands -> ../.agent/commands
└── skills -> ../.agent/skills
```

**Alternatives Considered**:
1. **Copy files to each agent dir** - Rejected: Maintenance nightmare, files drift apart
2. **Agent-agnostic location only** - Rejected: Agents expect their own config directories

### Multi-Select TUI Pattern

**Decision**: Use checkbox-style selection in existing Bubble Tea TUI

**Rationale**:
- Consistent with existing constitution principle selection
- Space to toggle, Enter to confirm
- Minimal learning curve

**Implementation**:
- Extend `InitModel` and `Model` with `selectedAgents []string`
- Change radio buttons to checkboxes for agent step
- Store comma-separated list in answers map

## Agent Definitions

| Agent | CLI Command | Config Dir | Install Instructions |
|-------|-------------|------------|---------------------|
| Claude Code | `claude` | `.claude/` | `npm install -g @anthropic-ai/claude-code` |
| OpenCode | `opencode` | `.opencode/` | `go install github.com/opencode-ai/opencode@latest` |
| Codex | `codex` | `.codex/` | `pip install openai-codex` (example) |

## Edge Case Handling

### Agent Not Installed

**Approach**: Check with `exec.LookPath()` before launch
- If not found: Print error with install instructions
- Exit code 1 with helpful message

### Windows Symlink Support

**Approach**: Detect Windows and warn if symlinks unavailable
- Windows 10+ Developer Mode supports symlinks
- Fall back to copying files on Windows if symlink fails
- Log warning about potential drift

### Conflicting Config

**Approach**: Project config overrides global (existing pattern)
- Document precedence: personal-local > team-local > global > default
- `sl config show` displays effective values with scope

## Open Questions

None - all NEEDS CLARIFICATION items resolved.

## References

- [Cobra CLI Documentation](https://github.com/spf13/cobra)
- [Bubble Tea TUI](https://github.com/charmbracelet/bubbletea)
- Existing launcher pattern: `pkg/cli/launcher/launcher.go`
