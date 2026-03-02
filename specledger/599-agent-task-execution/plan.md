# Implementation Plan: AI Agent Task Execution Service

**Branch**: `599-agent-task-execution` | **Date**: 2026-03-01 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specledger/599-agent-task-execution/spec.md`

## Summary

Create a service that executes SpecLedger tasks via a Goose AI agent, orchestrated through `sl agent run`. The system reads tasks from `issues.jsonl`, selects ready (unblocked) tasks in priority order, constructs execution context with task metadata and design notes, launches Goose as a subprocess, captures results, and updates task status. Builds on 597-agent-model-config for configuration and 597-issue-create-fields for the task model.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), gofrs/flock (file locking), go-git v5 (git operations), gopkg.in/yaml.v3 (config)
**Storage**: File-based — JSONL for issues (existing), JSON for agent run metadata, log files for output capture
**Testing**: `go test` via `make test` (unit tests), no integration test infrastructure yet
**Target Platform**: macOS/Linux CLI (darwin, linux)
**Project Type**: Single Go project — extends existing CLI
**Performance Goals**: <30s overhead per task (context building, status updates) beyond Goose processing time (SC-006)
**Constraints**: Sequential task execution for MVP, single local machine, clean git state required
**Scale/Scope**: Single user running tasks for one spec at a time; 10+ tasks per run in headless mode (SC-005)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Note: Constitution is an unfilled template (`.specledger/memory/constitution.md` contains placeholder content). Checking against template principles:

- [x] **Specification-First**: spec.md complete with 4 prioritized user stories (P1-P4), 17 functional requirements
- [x] **Test-First**: Test strategy deferred per spec assumption ("Test infrastructure is not yet available. DoD abstracts test requirements"). Unit tests for core logic will be written.
- [x] **Code Quality**: golangci-lint, gofmt, go vet (per CLAUDE.md CI checks)
- [x] **UX Consistency**: User flows documented in acceptance scenarios with Given/When/Then format
- [x] **Performance**: SC-006 defines <30s overhead per task
- [x] **Observability**: FR-007 requires execution logs per task, accessible via `sl agent status` / `sl agent logs`
- [x] **Issue Tracking**: Feature spec created with epic linkage to 597-agent-model-config and 597-issue-create-fields

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/599-agent-task-execution/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 research output
├── data-model.md        # Phase 1 data model
├── quickstart.md        # Phase 1 quickstart guide
├── contracts/
│   └── cli-interface.md # CLI command specifications
└── tasks.md             # Phase 2 output (created by /specledger.tasks)
```

### Source Code (repository root)

```text
pkg/cli/
├── agent/                    # NEW — Agent execution package
│   ├── types.go              # AgentRun, TaskResult, RunStatus types
│   ├── runner.go             # Core execution orchestrator
│   ├── context.go            # ExecutionContext builder (instruction file generation)
│   ├── goose.go              # Goose-specific subprocess management
│   ├── selector.go           # Task selection (priority + dependency ordering)
│   └── store.go              # AgentRun persistence (.agent-runs/ JSON files)
├── commands/
│   └── agent.go              # NEW — sl agent run/status/logs/stop commands
├── launcher/
│   └── launcher.go           # MODIFY — Add Goose to DefaultAgents, extend for non-interactive use
└── config/
    └── (existing)            # REUSE — ResolveAgentEnv(), GetEnvVars()

pkg/issues/
├── store.go                  # REUSE — ListReady(), Update(), Get()
└── dependencies.go           # REUSE — DetectCycles(), GetBlockedIssuesWithBlockers()

specledger/<spec>/
├── issues.jsonl              # EXISTING — Task definitions (read by agent)
├── agent-recipe.yaml         # NEW (optional) — P4 custom execution config
└── .agent-runs/              # NEW (gitignored) — Execution artifacts
    ├── agent.pid
    ├── latest.json
    └── <run-id>/
```

**Structure Decision**: Follows the existing single Go project layout with a new `pkg/cli/agent/` package for agent-specific logic. CLI commands go in `pkg/cli/commands/agent.go` consistent with existing commands (auth.go, issue.go, etc.). The `.agent-runs/` directory is a gitignored runtime artifact store, separate from spec documentation.

## Complexity Tracking

> No violations identified. The design reuses existing patterns (JSONL store, AgentLauncher, config merge) and adds one new package (`pkg/cli/agent/`) with focused responsibilities.
