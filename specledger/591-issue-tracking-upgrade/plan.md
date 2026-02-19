# Implementation Plan: Built-In Issue Tracker

**Branch**: `591-issue-tracking-upgrade` | **Date**: 2026-02-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specledger/591-issue-tracking-upgrade/spec.md`

## Summary

Replace Beads+Perles issue tracking with a built-in JSONL-based issue tracker integrated into the `sl` CLI. Issues will be stored per-spec at `specledger/<spec>/issues.jsonl` with globally unique SHA-256 based IDs (`SL-<6-char-hex>`). This eliminates daemon dependencies and external tool coupling while maintaining backward compatibility for migration.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: Cobra (CLI framework), go-git v5 (branch detection), crypto/sha256 (ID generation)
**Storage**: File-based JSONL at `specledger/<spec>/issues.jsonl` (per-spec storage)
**Testing**: Go testing package + testify for assertions
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: Single project (Go CLI)
**Performance Goals**: Issue operations < 100ms for up to 1000 issues per spec (no daemon overhead)
**Constraints**: No daemon/background processes; direct file I/O only; file locking for concurrent writes
**Scale/Scope**: Up to 1000 issues per spec; 100,000+ total issues across all specs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Note: The constitution.md file is currently a template. Proceeding with standard Go best practices.

- [x] **Specification-First**: Spec.md complete with prioritized user stories (8 user stories with P1-P3 priorities)
- [x] **Test-First**: Test strategy defined (unit tests for issue operations, integration tests for CLI commands)
- [x] **Code Quality**: Go standard formatting (gofmt), linting (golangci-lint)
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: Metrics defined (< 100ms operations, no daemon overhead)
- [x] **Observability**: Error messages with context, --json flag for structured output
- [x] **Issue Tracking**: This feature *creates* the issue tracking system

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/591-issue-tracking-upgrade/
├── spec.md              # Feature specification
├── plan.md              # This file (/specledger.plan command output)
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── contracts/           # Phase 1 output (not needed - internal CLI only)
```

### Source Code (repository root)

```text
pkg/
├── cli/
│   └── commands/
│       └── issue.go         # NEW: sl issue subcommand (create, list, update, close, show, migrate, link, repair)
├── issues/
│   ├── store.go             # NEW: JSONL file operations (read, write, append, lock)
│   ├── issue.go             # NEW: Issue struct and field definitions
│   ├── id.go                # NEW: SHA-256 based ID generation
│   ├── migrate.go           # NEW: Beads migration logic + cleanup (.beads/, mise.toml)
│   ├── duplicate.go         # NEW: Duplicate detection (string similarity)
│   ├── dependencies.go      # NEW: Dependency management and cycle detection
│   └── definition_of_done.go # NEW: Definition of done validation
├── cli/
│   └── prerequisites/
│       └── checker.go       # MODIFY: Remove bd and perles from CoreTools
pkg/embedded/
├── templates/
│   └── specledger/
│       ├── init.sh          # MODIFY: Remove beads setup call
│       └── .specledger/scripts/bash/
│           └── setup-beads.sh # DELETE: No longer needed
│       └── .claude/skills/bd-issue-tracking/ # DELETE/MODIFY: Replace with sl-issue-tracking
└── skills/commands/
    ├── specledger.implement.md # MODIFY: Update to use sl issue commands
    └── specledger.tasks.md     # MODIFY: Update to use sl issue commands

tests/
├── issues/
│   ├── store_test.go        # Unit tests for JSONL operations
│   ├── issue_test.go        # Unit tests for issue struct
│   ├── id_test.go           # Unit tests for ID generation
│   ├── migrate_test.go      # Unit tests for migration
│   └── duplicate_test.go    # Unit tests for duplicate detection
└── integration/
    └── issue_cli_test.go    # Integration tests for sl issue commands
```

**Structure Decision**: Single Go project structure following existing patterns in `pkg/cli/commands/`. New `pkg/issues/` package encapsulates all issue tracking logic as a reusable library.

## Phase 0: Research Summary

See [research.md](./research.md) for complete findings.

**Key Decisions**:
1. **ID Format**: `SL-<6-char-hex>` from SHA-256(spec_context + title + created_at) - deterministic, collision-resistant
2. **Storage**: Per-spec JSONL files - eliminates cross-branch conflicts, aligns with feature workflow
3. **Duplicate Detection**: Levenshtein distance for title similarity (> 80% match = warning)
4. **Migration Strategy**: Map Beads issues to spec directories, then cleanup .beads/ and mise.toml

## Phase 1: Design Artifacts

- [data-model.md](./data-model.md) - Issue entity, IssueStore, DefinitionOfDone, Dependency models
- [quickstart.md](./quickstart.md) - CLI usage examples and workflow scenarios

## Complexity Tracking

> No violations identified - standard single-project Go CLI structure.
