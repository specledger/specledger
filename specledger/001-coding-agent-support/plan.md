# Implementation Plan: Multi-Coding Agent Support

**Branch**: `001-coding-agent-support` | **Date**: 2026-03-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specledger/001-coding-agent-support/spec.md`

## Summary

Add multi-coding agent support to SpecLedger CLI, enabling users to launch any coding agent (Claude Code, OpenCode, Codex) via `sl code [<agent>]` command with configurable arguments and environment variables. Includes multi-agent selection during `sl new`/`sl init` with symlink-based sharing of commands/skills.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), YAML v3 (config), go-git/v5
**Storage**: File-based (YAML config files, JSONL for issues)
**Testing**: Go testing package (`_test.go` files)
**Target Platform**: macOS, Linux (CLI)
**Project Type**: Single CLI project
**Performance Goals**: Agent launch in <2 seconds
**Constraints**: CLI tool - minimal startup overhead
**Scale/Scope**: Single-user CLI tool

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with principles from `.specledger/memory/constitution.md`:

- [x] **Specification-First**: Spec.md complete with prioritized user stories
- [x] **Test-First**: Test strategy defined (unit tests for agent registry, integration tests for launch)
- [x] **Code Quality**: Linting via golangci-lint (existing project standard)
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: Agent launch <2 seconds defined in Success Criteria
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
│   └── agents.go            # NEW: agent registry
├── config/
│   ├── config.go            # EXISTING: extend AgentConfig
│   └── schema.go            # EXISTING: add new schema keys
└── tui/
    ├── sl_new.go            # EXISTING: multi-select for agents
    └── sl_init.go           # EXISTING: multi-select for agents

pkg/cli/playbooks/
└── copy.go                  # EXISTING: extend for symlink creation

internal/
└── agent/
    └── registry.go          # NEW: agent definitions and install info

tests/
├── unit/
│   └── launcher_test.go     # NEW: agent launch tests
└── integration/
    └── code_command_test.go # NEW: sl code integration tests
```

**Structure Decision**: Extends existing CLI structure. New `sl code` command in `pkg/cli/commands/`, agent registry in `pkg/cli/launcher/`, config schema extensions in `pkg/cli/config/`.

## Complexity Tracking

> No violations to justify

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none) | - | - |

## Implementation Phases

### Phase 1: Core Agent Launch (P1 Stories)

1. Create agent registry with Claude Code, OpenCode, Codex definitions
2. Extend `AgentConfig` with `Default` and `Arguments` fields
3. Implement `sl code [<agent>]` command
4. Add config schema keys: `agent.default`, `agent.arguments`

### Phase 2: Multi-Agent Setup (P2 Stories)

1. Extend TUI for multi-select in `sl new` and `sl init`
2. Implement symlink creation for `.agent/commands` and `.agent/skills`
3. Update constitution to store selected agents

### Phase 3: Polish & Edge Cases (P3 Stories)

1. Handle agent not installed gracefully
2. Provide install instructions per agent
3. Windows symlink compatibility check
