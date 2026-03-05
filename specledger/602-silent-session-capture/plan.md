# Implementation Plan: Silent Session Capture

**Branch**: `602-silent-session-capture` | **Date**: 2026-03-04 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specledger/602-silent-session-capture/spec.md`

## Summary (v2)

Create a `/specledger.commit` slash command for auth-aware commit workflow. Fix the existing PostToolUse hook to silently skip when no credentials or no project ID. Add dual error logging (local file + Supabase) when capture fails for authenticated users. The slash command is used by the agent when users ask to commit via chat.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), Supabase (GoTrue auth, PostgREST, Storage)
**Storage**: File-based (credentials.json, specledger.yaml, JSONL), Supabase (sessions, errors)
**Testing**: Go standard `testing` package, table-driven tests
**Target Platform**: Cross-platform CLI (Windows, macOS, Linux)
**Project Type**: Single project (Go CLI + slash command markdown)
**Performance Goals**: Silent skip paths complete instantly (single file check). Error logging is non-blocking.
**Constraints**: Capture/logging never blocks commit or push. Local log always succeeds.
**Scale/Scope**: 3-4 Go files modified, 1 new Go file, 1 new slash command markdown, 1 new Supabase table

## Constitution Check

- [x] **Specification-First**: Spec.md complete with 4 prioritized user stories
- [x] **Test-First**: Test cases defined for all user stories
- [x] **Code Quality**: golangci-lint, gofmt, go vet
- [x] **UX Consistency**: Auth decision matrix documented (skip silently / log error)
- [x] **Performance**: Silent paths return immediately, logging is non-blocking
- [x] **Observability**: Dual error logging (local + Supabase) is the core deliverable
- [ ] **Issue Tracking**: Epic to be created in /tasks phase

**Complexity Violations**: None.

## Project Structure

### Documentation (this feature)

```text
specledger/602-silent-session-capture/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── error-log-api.md
│   └── slash-command.md
└── checklists/
    └── requirements.md
```

### Source Code (files to create/modify)

```text
# New files
pkg/embedded/skills/commands/specledger.commit.md   # Slash command definition
pkg/cli/session/errorlog.go                          # Error logging (local + Supabase)

# Modified files
pkg/cli/session/capture.go          # Reorder auth check, silent skip, add error logging
pkg/cli/session/capture_test.go     # Tests for silent skip behavior
pkg/cli/session/queue.go            # Add error logging on retry failures
```

**Structure Decision**: Minimal additions. One new Go file for error logging, one new markdown file for the slash command. Rest is modifications to existing code.

## Implementation Approach

### Part 1: Slash Command (US1 + US2)

Create `pkg/embedded/skills/commands/specledger.commit.md` with:
- YAML frontmatter with description
- Workflow: check staged → commit → push → show summary
- Auth-aware: check `~/.specledger/credentials.json` existence
- Uses `$ARGUMENTS` for optional commit message
- Instructs agent to use this command when user asks to commit via chat

After creating, run `sl bootstrap` or copy to `.claude/commands/` to make it available.

### Part 2: Silent Skip in Capture() (US4)

In `pkg/cli/session/capture.go`:
1. Move `auth.LoadCredentials()` to right after tool success check (before project ID)
2. Return `result` with nil error when no credentials (silent skip)
3. Remove stderr warnings for missing project ID, return silently
4. No changes needed in `commands/session.go` (already handles nil error correctly)

### Part 3: Error Logging (US3)

Create `pkg/cli/session/errorlog.go`:
- `LogCaptureError(entry)` function
- Writes to `~/.specledger/capture-errors.log` (JSONL, append-only) first
- Then POSTs to Supabase `/rest/v1/session_capture_errors` (best-effort)
- Never blocks, never panics

Integrate into:
- `capture.go`: Call `LogCaptureError()` when upload or metadata creation fails
- `queue.go`: Call `LogCaptureError()` when `ProcessQueue()` retry fails

### Part 4: Supabase Table (infrastructure)

Create `session_capture_errors` table on Supabase with:
- Columns: id, user_id, project_id, session_id, error_message, feature_branch, commit_hash, retry_count, created_at
- RLS: users read/write own errors

This is a Supabase migration, separate from Go code.

## Complexity Tracking

> No violations. All changes are straightforward.
