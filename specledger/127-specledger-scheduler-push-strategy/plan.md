# Implementation Plan: Push-Triggered Scheduler Strategy

**Branch**: `127-specledger-scheduler-push-strategy` | **Date**: 2026-03-12 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specledger/127-specledger-scheduler-push-strategy/spec.md`

## Summary

Enable push-triggered implementation execution by installing a `pre-push` git hook that detects approved SpecLedger specs and spawns `sl implement` as a detached background process. `sl implement` is a thin Go command that delegates to the Claude CLI via `claude -p "/specledger.implement" --dangerously-skip-permissions` — the Claude session handles task reading and sequential execution. Includes a new `sl approve` command to gate spec readiness, a new `sl hook` command group for hook lifecycle management, and a PID-based execution lock to prevent duplicate runs. All new code follows existing Cobra CLI patterns and Go project conventions.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), go-git/v5 (git operations), gofrs/flock (file locking), encoding/json, Claude CLI (`claude` binary for implementation execution)
**Storage**: File-based (JSON for exec.lock, append-only text for push-hook.log, YAML for config)
**Testing**: Go testing package (`_test.go` files), table-driven tests
**Target Platform**: macOS (Darwin) and Linux
**Project Type**: Single CLI binary
**Performance Goals**: Hook adds <2s overhead to git push; hook install <5s (SC-002, SC-003)
**Constraints**: No server-side changes; works with existing specledger/ directory structure; cross-platform macOS + Linux
**Scale/Scope**: Single developer local workflow

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Note: Constitution is not yet filled in (template placeholders). Applying reasonable defaults based on project conventions observed in the codebase.

- [x] **Specification-First**: Spec.md complete with 4 prioritized user stories (P1-P3), 16 functional requirements, edge cases
- [x] **Test-First**: Test strategy defined - unit tests for lock management, hook script generation, status parsing; integration tests for install/uninstall cycle
- [x] **Code Quality**: Go standard tooling (gofmt, go vet); existing codebase uses these
- [x] **UX Consistency**: User flows documented in spec acceptance scenarios (US0-US3)
- [x] **Performance**: Metrics defined - <2s hook overhead (SC-003), <5s install (SC-002)
- [x] **Observability**: Logging to `.specledger/logs/push-hook.log` with structured entries (FR-010)
- [ ] **Issue Tracking**: Epic should be created with `sl issue create --type epic` and linked to spec

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/127-specledger-scheduler-push-strategy/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 research output
├── data-model.md        # Phase 1 data model
├── quickstart.md        # Phase 1 quickstart guide
├── contracts/
│   └── sl-hook-cli.md   # CLI contract for new commands
└── tasks.md             # Phase 2 output (via /specledger.tasks)
```

### Source Code (repository root)

```text
pkg/cli/commands/
├── approve.go           # NEW: sl approve command
├── hook.go              # NEW: sl hook install/uninstall/status/execute commands
├── implement.go         # NEW: sl implement command (thin wrapper → Claude CLI)
└── lock.go              # NEW: sl lock reset/status commands

pkg/cli/hooks/
├── claude.go            # EXISTING: Claude Code hook management (unchanged)
└── githook.go           # NEW: Git hook install/uninstall/status logic

pkg/cli/scheduler/
├── lock.go              # NEW: ExecutionLock create/check/remove/stale-detect
├── lock_test.go         # NEW: Lock management tests
├── detector.go          # NEW: Approved spec detection logic
├── detector_test.go     # NEW: Detection tests
├── executor.go          # NEW: Claude CLI process spawning (claude -p "/specledger.implement" --dangerously-skip-permissions)
└── executor_test.go     # NEW: Executor tests

pkg/cli/spec/
└── status.go            # EXISTING or NEW: Spec status read/write helpers

cmd/sl/main.go           # MODIFIED: Register VarApproveCmd, VarHookCmd, VarImplementCmd

tests/
├── hook_install_test.go     # NEW: Integration tests for hook lifecycle
└── approve_integration_test.go  # NEW: Integration tests for approve
```

**Structure Decision**: Follows existing single-project layout. New packages (`scheduler/`) keep hook execution logic separate from CLI command wiring. The `hooks/` package is extended with a new file for git hooks alongside the existing Claude Code hook file.

## Complexity Tracking

No violations - all patterns follow existing codebase conventions.

## Previous Work

- **Feature 010 (Checkpoint Session Capture)**: Established hook infrastructure in `pkg/cli/hooks/claude.go`. Pattern: LoadSettings -> Install/Uninstall/HasHook functions. This feature follows the same pattern but for git hooks instead of Claude Code hooks.
- **Feature 598 (SDD Workflow Streamline)**: Defined the 4-layer architecture (L0: Hooks, L1: CLI, L2: AI Commands, L3: Skills). Push hooks are L0; `sl approve`/`sl hook` are L1.
- **Existing Dependencies**: `gofrs/flock` (already in go.mod for issue store locking), `go-git/v5` (already used for git operations).
- **specledger.implement prompt**: `.claude/commands/specledger.implement.md` already exists and handles full task orchestration (phase-by-phase execution, progress tracking via `sl issue`, DoD verification). `sl implement` delegates to this prompt via `claude -p "/specledger.implement" --dangerously-skip-permissions`.

## External Dependencies

- **Claude CLI** (`claude` binary): Required at runtime for `sl implement`. Not a Go dependency — must be installed separately on the developer's machine. `sl implement` verifies availability via `exec.LookPath("claude")` before attempting execution.

## Phase Summary

### Phase 1: Approve Command (P1 - US0)
- `sl approve` command in `pkg/cli/commands/approve.go`
- Spec status parsing and writing in `pkg/cli/spec/status.go`
- Artifact validation (spec.md, plan.md, tasks.md exist and non-empty)
- Unit + integration tests

### Phase 2: Hook Management (P2 - US2)
- `sl hook install/uninstall/status` commands in `pkg/cli/commands/hook.go`
- Git hook file management in `pkg/cli/hooks/githook.go`
- Marker-based install/uninstall (preserves existing hooks)
- Unit + integration tests

### Phase 3: Push-Triggered Execution (P1 - US1)
- `sl hook execute` internal command (invoked by pre-push hook script)
- `sl implement` command in `pkg/cli/commands/implement.go`:
  - Thin Go wrapper that acquires execution lock, then spawns Claude CLI
  - Invokes: `claude -p "/specledger.implement" --dangerously-skip-permissions`
  - Manages lifecycle: lock acquisition → Claude CLI spawn → wait for completion → lock release
  - Commits generated code to `<feature-branch>/implement` sub-branch
  - Writes result summary to `.specledger/logs/<feature>-result.md`
- Execution lock management in `pkg/cli/scheduler/lock.go`
  - Lock check: if `exec.lock` exists, skip execution and log "already running"
  - No automatic stale lock detection — manual recovery only via `sl lock reset`
  - `sl lock reset` removes `exec.lock` unconditionally
  - `sl lock status` displays current lock info (PID, feature, started_at)
- Approved spec detection in `pkg/cli/scheduler/detector.go`
- Claude CLI process spawning in `pkg/cli/scheduler/executor.go`:
  - Verifies `claude` CLI is available in PATH
  - Spawns `claude -p "/specledger.implement" --dangerously-skip-permissions` with `SysProcAttr{Setpgid: true}`
  - Redirects stdout/stderr to `.specledger/logs/<feature>-claude.log`
  - Monitors process completion and triggers lock cleanup
- Error handling strategy (FR-011 — graceful failure, never block push):
  - Pre-push hook always exits 0 regardless of errors
  - All errors logged to `.specledger/logs/push-hook.log` with severity, timestamp, error type
  - Edge cases:
    - `sl` binary not found: hook script checks `command -v sl`, logs warning, exits 0
    - `claude` CLI not found: `sl implement` logs error, removes lock, exits
    - Malformed spec.md: detector logs parse error, skips spec
    - Lock held (active run): executor logs "already running", skips
    - Claude CLI spawn failure: logs error, removes lock, exits
  - `sl hook status` displays last 5 errors from push-hook.log
- Unit + integration tests

### Phase 4: Logging & Observability (P3 - US3)
- Structured log writing to `.specledger/logs/push-hook.log`
- Log rotation (keep last 50 entries per SC-005)
- Result summary at `.specledger/logs/<feature>-result.md`
