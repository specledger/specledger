# Implementation Plan: Improve Issue Parent-Child Linking

**Branch**: `606-issue-parent-linking` | **Date**: 2026-03-10 | **Spec**: [spec.md](./spec.md)

## Summary

Add `parent` as a link type in `sl issue link`, add `--orphaned` filter to `sl issue list`, add a bulk `sl issue reparent` command, and update AI skill instructions to enforce `--parent` on all non-epic issue creation.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), JSONL file store (pkg/issues)
**Storage**: JSONL files per spec (`specledger/<spec>/issues.jsonl`)
**Project Type**: CLI tool
**Target Platform**: Cross-platform (macOS, Linux, Windows)

## Constitution Check

Constitution is uninitialized (template placeholders only). No gates to evaluate.

## Phase 0: Research Summary

### Current Architecture

**Link command** (`pkg/cli/commands/issue.go` lines 175-188, 1589-1632):
- Supports `blocks` and `related` link types via `store.AddDependency()`
- Validates both issue IDs exist, includes cycle detection
- Updates bidirectional arrays (`Blocks`/`BlockedBy`)

**Parent-child** is a separate mechanism from blocking:
- `ParentID *string` field on Issue entity (`pkg/issues/issue.go` line 44)
- Set via `sl issue update --parent` or `sl issue create --parent`
- Validated in `store.Update()` (lines 286-323): self-parent check, existence check, cycle detection via `wouldCreateParentCycle()`
- `store.GetChildren(parentID)` exists for querying children

**List command** (`pkg/cli/commands/issue.go` lines 372-464):
- Uses `ListFilter` struct with `Status`, `IssueType`, `Priority`, `Labels`, `SpecContext`, `All`, `Blocked` fields
- No orphan detection currently exists

**Skill file**: `.claude/commands/specledger.tasks.md`
- Contains task generation instructions including CLI examples with `--parent`
- But the emphasis on `--parent` being **required** is not strong enough

### Decisions

**US1 — `parent` link type**: Extend the link command's switch/case to handle `parent` by calling the existing `store.Update()` with `ParentID` set, rather than going through `AddDependency()`. This reuses all existing validation (existence, self-parent, cycle detection). No new link type constant needed in `dependencies.go`.

- **Rationale**: Parent-child is conceptually different from blocking dependencies. Routing through `store.Update()` reuses all existing validation logic.
- **Alternative considered**: Adding `LinkParent` constant to `dependencies.go` and modifying `AddDependency()` — rejected because parent-child is not a dependency and mixing the two would complicate the dependency graph.

**US2 — `--orphaned` flag**: Add `Orphaned bool` field to `ListFilter` and filter in the list command's output logic. An orphaned issue is a non-epic issue with `ParentID == nil`.

- **Rationale**: Simple filter addition consistent with existing `--blocked` flag pattern.

**US3 — `sl issue reparent`**: New subcommand under `issue` that takes `<parent-id> <child-id>...` and loops through children, calling `store.Update()` for each with `ParentID` set.

- **Rationale**: Reuses existing update/validation logic. Continue-on-error pattern matches FR-006.

**US4 — Skill instructions**: Update `.claude/commands/specledger.tasks.md` to add bold/emphasized text about `--parent` being **MANDATORY** for all non-epic issues, with a validation reminder.

### Previous Work

- **591-issue-tracking-upgrade**: Built the entire issue tracking system including the Issue entity, JSONL store, and CLI commands.
- **595-issue-tree-ready**: Added tree view and ready command that depend on proper parent-child hierarchy.
- **594-issues-storage-config**: Configured issues storage paths.

## Phase 1: Design

### Files to Modify

| File | Change | User Story |
|------|--------|------------|
| `pkg/cli/commands/issue.go` | Add `parent` case to link command switch | US1 |
| `pkg/cli/commands/issue.go` | Add `--orphaned` flag to list command | US2 |
| `pkg/issues/issue.go` | Add `Orphaned` field to `ListFilter` | US2 |
| `pkg/issues/store.go` | Add orphan filtering logic to `List()` | US2 |
| `pkg/cli/commands/issue.go` | Add `reparent` subcommand | US3 |
| `.claude/commands/specledger.tasks.md` | Emphasize `--parent` as mandatory | US4 |

### No New Files Needed

All changes are extensions to existing files. No new packages, entities, or data model changes required. The `Issue` entity already has `ParentID` and all validation logic.

### Key Implementation Details

**US1 — Link command `parent` type**:
```
// In the link command's switch on linkType:
case "parent":
    // Reuse store.Update() which already validates existence, self-parent, cycles
    store.Update(fromID, IssueUpdate{ParentID: &toID})
```

**US2 — Orphan detection logic**:
```
// An issue is orphaned if:
// 1. It is NOT an epic (epics are root-level by nature)
// 2. Its ParentID is nil
```

**US3 — Reparent command signature**:
```
sl issue reparent <parent-id> <child-id> [child-id...]
```
Loops through children, updates each, reports successes and failures at the end.

## Phase 2: Work Breakdown

### US1: Link Parent via Issue Link Command (P1)
- **Task 1**: Add `parent` case to link command switch in `pkg/cli/commands/issue.go`
- **Depends on**: Nothing

### US4: AI Agent Skill Instructions (P1)
- **Task 2**: Update `.claude/commands/specledger.tasks.md` to emphasize `--parent` as mandatory
- **Depends on**: Nothing (can be done in parallel with all other tasks)

### US2: Orphaned Issue Detection (P2)
- **Task 3**: Add `Orphaned bool` to `ListFilter` in `pkg/issues/issue.go`
- **Task 4**: Add orphan filtering logic to `List()` in `pkg/issues/store.go`
- **Task 5**: Add `--orphaned` flag to list command in `pkg/cli/commands/issue.go`
- **Depends on**: Task 3 → Task 4 → Task 5 (sequential, same data flow)

### US3: Bulk Reparent Command (P3)
- **Task 6**: Add `reparent` subcommand to `pkg/cli/commands/issue.go`
- **Depends on**: Nothing (reuses existing store.Update logic)

## Success Criteria

- [ ] SC-001: `sl issue link <child> parent <parent>` sets parent-child relationship
- [ ] SC-002: `sl issue list --orphaned` shows only non-epic issues without parents
- [ ] SC-003: `sl issue reparent <parent> <child1> <child2> ...` sets parent for all children in one command
- [ ] SC-004: Skill instructions explicitly mark `--parent` as mandatory for non-epic issues
