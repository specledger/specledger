# Implementation Plan: Multi-Coding Agent Support

**Branch**: `001-coding-agent-support` | **Date**: 2026-03-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specledger/001-coding-agent-support/spec.md`

## Summary

Add multi-coding agent support to SpecLedger CLI, enabling users to launch any coding agent (Claude Code, OpenCode, Copilot CLI, Codex) via `sl code [<agent>]` command with per-agent configurable arguments and environment variables. Includes multi-agent selection during `sl new`/`sl init` with symlink-based sharing on macOS/Linux and file copies on Windows.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), YAML v3 (config), go-git/v5
**Storage**: File-based (YAML config files, JSONL for issues)
**Testing**: Go testing package (`_test.go` files)
**Target Platform**: macOS, Linux, Windows (CLI)
**Project Type**: Single CLI project
**Performance Goals**: Sub-second agent launch overhead
**Constraints**: CLI tool - minimal startup overhead
**Scale/Scope**: Single-user CLI tool

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with principles from `.specledger/memory/constitution.md`:

- [x] **Specification-First**: Spec.md complete with prioritized user stories
- [x] **Test-First**: Test strategy defined (unit tests for agent registry, integration tests for launch)
- [x] **Code Quality**: Linting via golangci-lint (existing project standard)
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: Sub-second launch overhead defined in Success Criteria
- [x] **Observability**: Existing logger package for debug output
- [x] **Issue Tracking**: Will create epic with `sl issue create --type epic`

**Complexity Violations** (if any, justify in Complexity Tracking table below):
- None identified

## Project Structure

### Documentation (this feature)

```text
specledger/001-coding-agent-support/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (via /specledger.tasks)
```

### Source Code (repository root)

```text
pkg/cli/
├── commands/
│   └── code.go              # NEW: sl code command
├── launcher/
│   ├── launcher.go          # EXISTING: extend for multi-agent
│   └── agents.go            # NEW: agent registry (4 agents + install commands)
├── config/
│   ├── config.go            # EXISTING: extend AgentConfig
│   └── schema.go            # EXISTING: add per-agent schema keys
└── tui/
    ├── sl_new.go            # EXISTING: multi-select for agents
    └── sl_init.go           # EXISTING: multi-select for agents

pkg/cli/playbooks/
└── copy.go                  # EXISTING: extend for symlink/copy logic

internal/
└── agent/
    ├── registry.go          # NEW: agent definitions (name, CLI command, install command)
    └── platform.go          # NEW: Windows detection and copy fallback

tests/
├── unit/
│   └── launcher_test.go     # NEW: agent launch tests
└── integration/
    └── code_command_test.go # NEW: sl code integration tests
```

**Structure Decision**: Extends existing CLI structure. New `sl code` command in `pkg/cli/commands/`, agent registry in `pkg/cli/launcher/`, config schema extensions in `pkg/cli/config/`. Platform-specific logic in `internal/agent/platform.go`.

## Complexity Tracking

> No violations to justify

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none) | - | - |

## Implementation Phases

The implementation follows a 5-phase structure with parallel execution paths after setup:

```
Setup (Phase 0) ──┬──► P1 (US1+US2) ──┬──► Polish (Phase 4)
                  │                   │
                  ├──► P2 (US3+US4) ───┤
                  │                   │
                  └──► P3 (US5) ───────┘
```

### Phase 0: Setup - Agent Infrastructure

1. Create agent registry with Claude Code, OpenCode, Copilot CLI, Codex definitions (including install commands)
2. Add platform detection for Windows vs Unix (symlink vs copy)
3. Extend config schema for `agent.default`, `agent.<name>.arguments`, `agent.<name>.env`

### Phase 1: Core Agent Launch (P1 - US1+US2)

1. Implement `sl code [<agent>]` command with agent selection
2. Extend launcher for multi-agent support with env var injection
3. Add binary detection with install command error messages
4. Read per-agent config for arguments with project/global merge
5. Wire command to launcher

### Phase 2: Multi-Agent Setup (P2 - US3+US4)

1. Extend TUI for multi-select in `sl new` and `sl init`
2. Create `.agent/commands` and `.agent/skills` directory structure
3. Implement symlink/copy logic for agent directories
4. Store selected agents in constitution

### Phase 3: Per-Agent Custom Arguments (P3 - US5)

1. Verify per-agent arguments work correctly (multiple flags, quotes, special chars)
2. Verify per-agent env vars work correctly

### Phase 4: Polish - Cross-Cutting Concerns

1. Update documentation for `sl code`
2. Validate quickstart.md scenarios

## Clarifications Applied (2026-03-15)

| Decision | Choice | Impact on Plan |
|----------|--------|----------------|
| Agent list | 4 agents: Claude, OpenCode, Copilot CLI, Codex | FR-002, agent registry includes all 4 |
| Arguments | Per-agent only | Config schema: `agent.<name>.arguments` |
| Windows | Copy files instead of symlinks | New `platform.go` for OS detection + copy logic |
| Binary error | Include install command | Error messages include npm install command |
| Existing .agent/ | Require --force flag | Add --force flag to sl new/init |

## Config Schema Refactoring (2026-03-15)

### Decision: Namespaced Per-Agent Config

The config system was refactored from Claude-centric to namespaced per-agent settings.

**Before (Claude-centric):**
```yaml
agent:
  api-key: sk-xxx
  model: claude-sonnet
  model.sonnet: claude-sonnet-4-20250514
  skip-permissions: true
```

**After (Namespaced):**
```yaml
agents:
  default: claude
  claude:
    api_key: sk-xxx
    model: claude-sonnet-4-20250514
    model_aliases:
      sonnet: claude-sonnet-4-20250514
    arguments: "--dangerously-skip-permissions"
  opencode:
    api_key: sk-openai
    model: gpt-4
```

### Keys Removed (Claude-Specific, No Multi-Agent Equivalent)

| Removed Key | Reason |
|-------------|--------|
| `agent.provider` | Claude-specific (anthropic/bedrock/vertex) |
| `agent.subagent_model` | Claude-only concept |
| `agent.permission_mode` | Claude flag, use `agent.claude.arguments` |
| `agent.skip_permissions` | Claude flag, use `agent.claude.arguments` |
| `agent.effort` | Claude flag, use `agent.claude.arguments` |
| `agent.allowed_tools` | Claude flag, use `agent.claude.arguments` |
| `agent.auth_token` | Duplicate of `agent.claude.api_key` |
| `agent.model.sonnet/opus/haiku` | Moved to `agent.claude.model_aliases.*` |

### Env Var Mappings Per Agent

Added to Agent struct in registry:
```go
type Agent struct {
    Name           string
    Command        string
    ConfigDir      string
    InstallCommand string
    APIKeyEnvVar   string  // e.g., "ANTHROPIC_API_KEY"
    BaseURLEnvVar  string  // e.g., "ANTHROPIC_BASE_URL"
    ModelEnvVar    string  // e.g., "ANTHROPIC_MODEL"
}
```

### Files Changed

| File | Change |
|------|--------|
| `internal/agent/registry.go` | Added env var mappings to Agent struct |
| `pkg/cli/config/agent_settings.go` | NEW: AgentSettings, ClaudeSettings, ConfigAgents structs |
| `pkg/cli/config/schema.go` | Updated for per-agent key patterns |
| `pkg/cli/config/merge.go` | Added ResolveAgentSettings() |
| `pkg/cli/config/config.go` | Added Agents field |
| `pkg/cli/config/migration.go` | NEW: Migration logic |
| `pkg/cli/config/config_test.go` | Updated tests for new namespaced schema |
| `pkg/cli/commands/config.go` | Updated for new key patterns |
| `pkg/cli/commands/code.go` | Use ResolveAgentSettings() with env var mapping |

## Implementation Status (2026-03-15)

All phases complete:
- [x] Phase 1: Update Agent Registry (env var mappings)
- [x] Phase 2: Create AgentSettings struct
- [x] Phase 3: Update Schema Registry
- [x] Phase 4: Update Merge Logic
- [x] Phase 5: Update Config Commands
- [x] Phase 6: Update Config Struct
- [x] Phase 7: Update sl code Command
- [x] Phase 8: Add Migration Logic
- [x] Phase 9: Update Tests (70 tests passing)
- [x] Phase 10: Update Spec Artifacts
