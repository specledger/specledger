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

### Phase 1: Core Agent Launch (P1 Stories)

1. Create agent registry with Claude Code, OpenCode, Copilot CLI, Codex definitions (including install commands)
2. Extend `AgentConfig` with per-agent `Arguments` and `Env` fields (e.g., `agent.claude.arguments`)
3. Implement `sl code [<agent>]` command with error messages including install commands
4. Add config schema keys: `agent.default`, `agent.<name>.arguments`, `agent.<name>.env`

### Phase 2: Multi-Agent Setup (P2 Stories)

1. Extend TUI for multi-select in `sl new` and `sl init`
2. Implement symlink creation for `.agent/commands` and `.agent/skills` on macOS/Linux
3. Implement file copy fallback for Windows
4. Add `--force` flag handling for existing `.agent/` directory
5. Update constitution to store selected agents

### Phase 3: Polish & Edge Cases (P3 Stories)

1. Handle agent binary not found with actionable install command
2. Test Windows file copy behavior
3. Document manual sync requirement for Windows users

## Clarifications Applied (2026-03-15)

| Decision | Choice | Impact on Plan |
|----------|--------|----------------|
| Agent list | 4 agents: Claude, OpenCode, Copilot CLI, Codex | FR-002, agent registry includes all 4 |
| Arguments | Per-agent only | Config schema: `agent.<name>.arguments` |
| Windows | Copy files instead of symlinks | New `platform.go` for OS detection + copy logic |
| Binary error | Include install command | Error messages include npm install command |
| Existing .agent/ | Require --force flag | Add --force flag to sl new/init |
