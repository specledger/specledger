# Implementation Plan: CLI Authentication and Comment Management

**Branch**: `009-add-login-and-comment-commands` | **Date**: 2026-02-10 | **Spec**: [spec.md](../.specledger/specs/009-add-login-and-comment-commands/spec.md)
**Input**: Feature specification from `/specledger/009-add-login-and-comment-commands/spec.md`

## Summary

Implement four Claude Code slash commands for Specledger authentication and comment management. This feature enables users to authenticate via browser token paste (`/specledger.login`, `/specledger.logout`), view comments on specification files (`/specledger.comment`), and resolve comments (`/specledger.resolve-comment`). Commands interact with Supabase backend via REST API with JWT authorization.

## Technical Context

**Language/Version**: Markdown (Claude Code slash commands), Shell (curl, file ops)
**Primary Dependencies**: Claude Code CLI, Supabase REST API (PostgREST), curl
**Storage**: File-based (`~/.specledger/session.json`), Supabase (comments table)
**Testing**: Manual testing via Claude Code execution
**Target Platform**: macOS, Linux (WSL compatible)
**Project Type**: CLI tool with Claude Code agent integration
**Performance Goals**: N/A (interactive user commands, sub-5s response)
**Constraints**: No secrets in code, file permission 600, RLS-based authorization
**Scale/Scope**: Single-user local session, shared Supabase comments backend

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with principles:

- [x] **Specification-First**: Spec.md complete with 4 prioritized user stories (P1-P4)
- [x] **Test-First**: Test strategy defined (manual testing via Claude Code + curl)
- [x] **Code Quality**: N/A - markdown commands, not compiled code
- [x] **UX Consistency**: User flows documented in spec acceptance scenarios
- [x] **Performance**: Sub-5s response for all commands
- [x] **Observability**: Error messages guide users to resolution
- [x] **Issue Tracking**: Linked to 008-cli-auth prior work

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/009-add-login-and-comment-commands/
├── plan.md              # This file
├── research.md          # Phase 0: Technical decisions
├── data-model.md        # Phase 1: Session and Comment entities
├── quickstart.md        # Phase 1: Developer guide
└── contracts/
    ├── comments-api.yaml     # OpenAPI schema for Supabase
    └── session.schema.json   # JSON Schema for session file
```

### Source Code (repository root)

```text
.claude/commands/
├── specledger.login.md          # Login command definition
├── specledger.logout.md         # Logout command definition
├── specledger.comment.md        # View comments command
└── specledger.resolve-comment.md # Resolve comment command

~/.specledger/
└── session.json                 # User session (created at runtime)
```

**Structure Decision**: This feature uses Claude Code slash commands (markdown files) instead of Go code. The commands are already created in `.claude/commands/` directory. No new Go code is required.

## Implementation Notes

### What's Already Done

The `.claude/commands/` directory already contains all four command definitions:

| File | Command | Status |
|------|---------|--------|
| specledger.login.md | `/specledger.login` | Defined |
| specledger.logout.md | `/specledger.logout` | Defined |
| specledger.comment.md | `/specledger.comment` | Defined |
| specledger.resolve-comment.md | `/specledger.resolve-comment` | Defined |

### What Needs Testing

1. **Login flow** with actual Auth UI
2. **Comment fetching** from Supabase with real data
3. **Comment resolution** via Supabase PATCH
4. **Error handling** for common failure cases

### Backend Requirements (Out of Scope)

The Supabase backend must have:
- `comments` table with schema from data-model.md
- RLS policies for read/update based on JWT
- Auth UI deployed at `https://specledger.io/cli/auth`

## Complexity Tracking

> No violations - feature uses simple markdown commands and standard REST API

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | - | - |

## Next Steps

1. **Validate commands** - Run `/specledger.login` through complete flow
2. **Test with Supabase** - Verify API integration with real backend
3. **Error handling** - Test edge cases (expired token, network errors)
4. **Document** - Update any missing error messages

---

## Phase 0 Output

See [research.md](./research.md) for technical decisions and prior work analysis.

## Phase 1 Output

- [data-model.md](./data-model.md) - Session and Comment entity definitions
- [contracts/comments-api.yaml](./contracts/comments-api.yaml) - OpenAPI schema
- [contracts/session.schema.json](./contracts/session.schema.json) - JSON Schema
- [quickstart.md](./quickstart.md) - Developer guide
