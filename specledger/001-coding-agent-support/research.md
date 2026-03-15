# Research: Multi-Coding Agent Support

**Feature**: 001-coding-agent-support
**Date**: 2026-03-15
**Status**: Complete (Updated with clarifications)

## Prior Work

### Related Codebase Components

| Component | Location | Relevance |
|-----------|----------|-----------|
| Agent Launcher | `pkg/cli/launcher/launcher.go` | Existing Claude Code launcher - extend for multi-agent |
| Config Schema | `pkg/cli/config/schema.go` | Agent config keys - extend with per-agent fields |
| TUI Agent Selection | `pkg/cli/tui/sl_new.go`, `sl_init.go` | Single-select agent - change to multi-select |
| Bootstrap Helpers | `pkg/cli/commands/bootstrap_helpers.go` | `launchAgent()` function - reuse pattern |

### Related Issues

No existing issues reference multi-agent support. This is a new feature.

## Technology Decisions

### Agent Registry Design

**Decision**: Static registry with agent definitions in code

**Rationale**:
- Agents are relatively stable (Claude Code, OpenCode, Copilot CLI, Codex)
- No need for dynamic plugin system
- Simpler to maintain and test
- Consistent with existing `launcher.DefaultAgents` pattern

**Alternatives Considered**:
1. **Plugin-based registry** - Rejected: Over-engineered for 4 known agents
2. **Config-file based registry** - Rejected: Adds complexity, users can't add custom agents easily anyway

### Config Key Design

**Decision**: Per-agent configuration keys (`agent.<name>.arguments`, `agent.<name>.env`)

**Rationale**:
- Different agents have different CLI arguments
- Per-agent config is clearer and easier to understand
- Simpler to implement - no global/per-agent merge logic needed
- YAGNI: No proven use case for global arguments that apply to all agents

**Alternatives Considered**:
1. **Global `agent.arguments` only** - Rejected: Different agents have different flags, would cause issues
2. **Both global + per-agent override** - Rejected: YAGNI, adds complexity without clear benefit

### Symlink Strategy

**Decision**: Create `.agent/` as source of truth, symlink (macOS/Linux) or copy (Windows) to agent-specific dirs

**Rationale**:
- Single source of truth for commands/skills
- Symlinks work well on Unix systems
- Windows file copies provide compatibility without Developer Mode requirement

**Structure (macOS/Linux)**:
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

**Structure (Windows)**:
```
.agent/
├── commands/     # Shared commands
└── skills/       # Shared skills

.claude/
├── commands/     # COPY of .agent/commands
└── skills/       # COPY of .agent/skills
```

**Alternatives Considered**:
1. **Symlinks only (fail on Windows)** - Rejected: Poor Windows UX
2. **Require Windows Developer Mode** - Rejected: Adds friction for Windows users
3. **Copy files to each agent dir (all platforms)** - Rejected: Maintenance nightmare on Unix, files drift apart

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

| Name | CLI Command | Config Dir | Install Command |
|------|-------------|------------|-----------------|
| Claude Code | `claude` | `.claude/` | `npm install -g @anthropic-ai/claude-code` |
| OpenCode | `opencode` | `.opencode/` | `go install github.com/opencode-ai/opencode@latest` |
| Copilot CLI | `github-copilot` | `.github/` | `npm install -g @github/copilot` |
| Codex | `codex` | `.codex/` | `npm install -g @openai/codex` |

## Edge Case Handling

### Agent Binary Not Found

**Approach**: Check with `exec.LookPath()` before launch
- If not found: Print error with actionable install command
- Example: `Error: 'claude' not found. Install: npm install -g @anthropic-ai/claude-code`
- Exit code 1

### Windows Symlink Support

**Approach**: Detect Windows and use file copy instead of symlinks
- No Developer Mode required
- Files are copied from `.agent/` to each agent directory
- User warned that changes to `.agent/` require manual sync or re-run
- Simpler than trying to detect Developer Mode

### Conflicting Config

**Approach**: Project config overrides global (existing pattern)
- Document precedence: personal-local > team-local > global > default
- `sl config show` displays effective values with scope

### Existing .agent/ Directory

**Approach**: Require `--force` flag to overwrite
- Error message: `Error: .agent/ exists. Use --force to overwrite.`
- Prevents accidental data loss
- Consistent with safety-first approach

## Clarifications Applied (2026-03-15)

| Question | Decision | Impact |
|----------|----------|--------|
| Agent list | 4 agents: Claude, OpenCode, Copilot CLI, Codex | Registry includes all 4 |
| Arguments config | Per-agent only | Config schema: `agent.<name>.arguments` |
| Windows handling | Copy files, no symlinks | Platform detection + copy logic |
| Binary error | Include install command | Actionable error messages |
| Existing .agent/ | Require --force | Safety check added |

## References

- [Cobra CLI Documentation](https://github.com/spf13/cobra)
- [Bubble Tea TUI](https://github.com/charmbracelet/bubbletea)
- Existing launcher pattern: `pkg/cli/launcher/launcher.go`
