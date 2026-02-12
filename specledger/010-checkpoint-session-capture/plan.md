# Implementation Plan: Checkpoint Session Capture

**Branch**: `010-checkpoint-session-capture` | **Date**: 2026-02-12 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specledger/010-checkpoint-session-capture/spec.md`

## Summary

Capture AI conversation segments during git commits and beads task execution, store them as compressed JSON in Supabase Storage, and record queryable metadata in the Supabase database. Sessions are captured automatically via Claude Code hooks (`PostToolUse` on Bash to detect commits). The AI is a first-class consumer — it queries past sessions by feature/task to maintain context continuity across sessions.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), net/http (Supabase REST + Storage API), compress/gzip (compression), encoding/json (serialization), go-git/v5 (commit detection)
**Storage**: Supabase Storage (session content as gzip JSON) + Supabase PostgreSQL (session metadata via PostgREST) + local filesystem (offline queue, delta state)
**Testing**: `go test` (unit + integration), contract tests for Supabase REST API responses
**Target Platform**: macOS, Linux (CLI)
**Project Type**: Single project (Go CLI)
**Performance Goals**: Session capture adds <5s to commit workflow (async upload); session query returns in <3s (SC-003, SC-004)
**Constraints**: Sessions up to 10 MB uncompressed; capture must not block commits; offline-capable with local queue
**Scale/Scope**: Per-project sessions, queryable by feature/task/commit/author/date

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution is not yet filled in for this project (template only). Checking against implicit principles:

- [x] **Specification-First**: Spec.md complete with 5 prioritized user stories and detailed acceptance scenarios
- [x] **Test-First**: Test strategy defined — contract tests for Supabase API, integration tests for capture flow, unit tests for delta computation and compression
- [x] **Code Quality**: Go standard tooling (`go vet`, `golangci-lint`); follows existing patterns in `pkg/cli/`
- [x] **UX Consistency**: User flows documented in spec acceptance scenarios; CLI commands follow existing `sl` patterns
- [x] **Performance**: SC-003 (<5s capture), SC-004 (<3s retrieval), SC-008 (<5s AI query) defined
- [x] **Observability**: Errors logged to stderr; `--json` flag for programmatic output; capture failures warn user
- [ ] **Issue Tracking**: Beads epic to be created during task generation phase

**Complexity Violations**: None identified.

## Previous Work

| Feature/Task | Relevance | Reusable Components |
|-------------|-----------|---------------------|
| 008-cli-auth | High | `auth.GetValidAccessToken()`, `auth.GetSupabaseURL()`, `auth.GetSupabaseAnonKey()`, HTTP client patterns |
| 009-command-system-enhancements | Medium | Supabase REST API call patterns, session.json storage pattern |
| Beads Issue Tracker | Medium | Task execution framework (`.beads/issues.jsonl`), `bd` CLI for task-linked sessions |

## Project Structure

### Documentation (this feature)

```text
specledger/010-checkpoint-session-capture/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── sessions-api.md  # REST API contracts
└── tasks.md             # Phase 2 output (/specledger.tasks command)
```

### Source Code (repository root)

```text
pkg/
├── cli/
│   ├── auth/                     # [EXISTING] Auth & Supabase credentials
│   │   ├── credentials.go        # Token management (reused)
│   │   └── client.go             # HTTP client, Supabase constants (reused)
│   ├── commands/                  # [EXISTING] Cobra commands
│   │   ├── session.go            # [NEW] sl session {capture,list,get,sync} commands
│   │   └── ...
│   └── session/                   # [NEW] Session capture business logic
│       ├── capture.go             # Hook handler: read stdin, detect commit, compute delta
│       ├── storage.go             # Supabase Storage upload/download/signed URLs
│       ├── metadata.go            # Supabase PostgREST CRUD for session metadata
│       ├── delta.go               # JSONL delta computation (offset tracking)
│       ├── queue.go               # Offline queue: local cache, retry logic
│       ├── compress.go            # gzip compress/decompress helpers
│       └── types.go               # Session, Message, HookInput, QueueEntry types
│
cmd/sl/
└── main.go                        # [MODIFY] Register session command

tests/
├── unit/
│   ├── delta_test.go              # Delta computation tests
│   ├── compress_test.go           # Compression tests
│   └── queue_test.go              # Queue management tests
├── contract/
│   └── session_api_test.go        # Supabase REST API contract tests
└── integration/
    └── capture_flow_test.go       # End-to-end capture flow test
```

**Structure Decision**: Single project, extending existing `pkg/cli/` structure. New `pkg/cli/session/` package for business logic, new `pkg/cli/commands/session.go` for Cobra commands. Follows established patterns from `pkg/cli/auth/` and `pkg/cli/commands/deps.go`.

## Architecture

### Capture Flow

```
Claude Code                    SpecLedger CLI                  Supabase
─────────────                  ──────────────                  ────────
     │                              │                              │
     │ [PostToolUse:Bash hook]      │                              │
     │──stdin JSON──────────────────▶                              │
     │                              │                              │
     │                     Parse hook input                        │
     │                     Check: was this a git commit?           │
     │                              │                              │
     │                     Read transcript_path JSONL              │
     │                     Compute delta (since last offset)       │
     │                     Build session JSON                      │
     │                     Gzip compress                           │
     │                              │                              │
     │                              │──Upload to Storage──────────▶│
     │                              │                              │
     │                              │──POST session metadata──────▶│
     │                              │                              │
     │                     Update local session-state.json         │
     │                              │                              │
     │◀─exit 0─────────────────────│                              │
     │                              │                              │
     │ [On upload failure]          │                              │
     │                     Queue locally                           │
     │                     (~/.specledger/session-queue/)           │
     │◀─exit 0 (never block)──────│                              │
```

### Commit Detection Strategy

The `PostToolUse` hook with `Bash` matcher fires after every Bash tool execution. The capture command inspects the tool input (available in hook JSON) to detect `git commit` commands:

1. Parse stdin JSON for `tool_input` containing the Bash command
2. Check if command matches `git commit` patterns (not `git commit --amend` rewriting)
3. If match: extract commit hash from `git rev-parse HEAD`, proceed with capture
4. If no match: exit 0 immediately (no-op)

This avoids false positives from non-commit Bash operations and keeps the hook lightweight.

### Session Delta Strategy

```
Transcript JSONL (append-only):
  Line 0: {"type":"user", ...}
  Line 1: {"type":"assistant", ...}
  ...
  Line 100: {"type":"user", ...}    ← last_offset from previous capture
  Line 101: {"type":"assistant",...} ┐
  Line 102: {"type":"user",...}      │ delta = lines 101..150
  ...                                │
  Line 150: {"type":"assistant",...}  ┘ ← new offset saved

session-state.json:
  {"sessions": {"abc-123": {"last_offset": 150, "last_commit": "def456..."}}}
```

## Complexity Tracking

No violations identified. The design follows established patterns:
- HTTP client for Supabase (same as auth package)
- File-based local state (same as credentials, session-queue)
- Cobra commands (same as deps, auth)
- No new external Go dependencies required beyond stdlib
