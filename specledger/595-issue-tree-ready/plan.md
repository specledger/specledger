# Implementation Plan: Issue Tree View and Ready Command

**Branch**: `595-issue-tree-ready` | **Date**: 2026-02-20 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specledger/595-issue-tree-ready/spec.md`

## Summary

Implement hierarchical tree view for issue dependencies and add `sl issue ready` command to list unblocked issues. Integrate ready state into `/specledger.implement` workflow for intelligent task selection.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), go-git v5 (branch detection), gofrs/flock (file locking), gopkg.in/yaml.v3
**Storage**: File-based JSONL at `specledger/<spec>/issues.jsonl`
**Testing**: Go testing framework (`go test ./...`)
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: Single project (Go CLI)
**Performance Goals**: Tree view < 2 seconds for 100 issues, ready command < 1 second
**Constraints**: Must integrate with existing issue store, maintain backward compatibility
**Scale/Scope**: Up to 100 issues per spec, unlimited specs per project

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Specification-First**: Spec.md complete with 4 prioritized user stories
- [x] **Test-First**: Test strategy defined (unit tests for tree rendering, ready state computation)
- [x] **Code Quality**: Go standard formatting (gofmt), golangci-lint
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: Metrics defined (< 2s for tree, < 1s for ready)
- [x] **Observability**: Standard Go logging, error messages for cycles/broken refs
- [x] **Issue Tracking**: Issues will be created via `sl issue create` for tracking

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/595-issue-tree-ready/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/specledger.tasks)
```

### Source Code (repository root)

```text
pkg/
├── issues/
│   ├── issue.go         # Existing - Issue entity
│   ├── store.go         # Existing - Store operations
│   ├── dependencies.go  # Existing - Dependency management
│   └── tree.go          # NEW - Tree rendering logic
│
├── cli/
│   └── commands/
│       └── issue.go     # MODIFY - Add ready command, implement tree rendering
│
pkg/embedded/skills/commands/
└── specledger.implement.md  # MODIFY - Use sl issue ready for task selection
```

**Structure Decision**: Extend existing `pkg/issues/` package with tree rendering. Add ready command to existing `issue.go` CLI. Update embedded implement skill.

## Complexity Tracking

No violations to track.

## Phase 0: Research Summary

See [research.md](./research.md) for detailed findings.

### Key Decisions

1. **Tree Rendering**: Use ASCII tree characters (├─, └─, │) for terminal compatibility
2. **Ready State Computation**: Compute at query time (no caching) for simplicity
3. **Implement Integration**: Modify prompt template to use `sl issue ready` command

## Phase 1: Design Artifacts

See [data-model.md](./data-model.md) for entity details.
See [quickstart.md](./quickstart.md) for usage scenarios.
