# Research: Improve Issue Parent-Child Linking

**Date**: 2026-03-10
**Feature**: 606-issue-parent-linking

## Prior Work

- **591-issue-tracking-upgrade**: Built the issue tracking system. Defined Issue entity with `ParentID *string` field, JSONL store, and all CLI commands including `sl issue create --parent` and `sl issue update --parent`.
- **595-issue-tree-ready**: Added `sl issue list --tree` and `sl issue ready` commands. Tree view renders parent-child hierarchy using Unicode box drawing. Ready command computes which tasks are unblocked. Both depend on correct parent-child relationships.
- **594-issues-storage-config**: Configured issues storage to use `issues.jsonl` per spec directory with `.lock` file for concurrency.

## Findings

### Link Command Architecture

The `sl issue link` command in `pkg/cli/commands/issue.go` uses a switch on link type string:
- `blocks` → calls `store.AddDependency(from, to, LinkBlocks)` which updates `Blocks`/`BlockedBy` arrays bidirectionally
- `related` → calls `store.AddDependency(from, to, LinkRelated)` which adds soft links

Parent-child is conceptually different — it uses `ParentID` (single pointer) rather than arrays. The existing `store.Update()` already handles all parent validation:
- Self-parent rejection
- Parent existence check
- Circular parent chain detection via `wouldCreateParentCycle()`

### Decision: Implement `parent` via store.Update(), not AddDependency()

**Rationale**: Parent-child is hierarchical organization, not a dependency. Routing through `store.Update()` reuses all existing validation. No need to add a new `LinkType` constant or modify the dependency graph logic.

**Alternatives considered**:
- Adding `LinkParent` to `dependencies.go` — rejected because it would mix hierarchy with dependency semantics, complicate cycle detection (two separate cycle checks), and add unnecessary complexity
- Creating a separate `sl issue parent` command — rejected because `sl issue link` is the established interface for creating relationships

### Orphan Detection

An orphaned issue is straightforward to detect:
- Type is NOT `epic` (epics are root-level by design)
- `ParentID` is `nil`

The existing `ListFilter` struct has a `Blocked bool` field that follows the same pattern — add `Orphaned bool` alongside it.

### Bulk Reparent

The `sl issue update` command already supports `--parent`. The bulk reparent command is a thin wrapper that:
1. Validates the parent issue exists once
2. Loops through child IDs
3. Calls `store.Update()` for each with `ParentID` set
4. Collects errors and reports at the end (continue-on-error)

### Skill Instructions

The current `.claude/commands/specledger.tasks.md` includes `--parent` in CLI examples but does not explicitly state it's mandatory. The fix is to add bold/emphasized instructions and a validation step.
