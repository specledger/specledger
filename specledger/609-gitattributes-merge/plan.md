# Implementation Plan: Gitattributes Merge

**Branch**: `609-gitattributes-merge` | **Date**: 2026-03-12 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specledger/609-gitattributes-merge/spec.md`

## Summary

PRs are cluttered with machine-generated specledger artifacts (`issues.jsonl`, `tasks.md`) that GitHub doesn't collapse. This feature populates the `.gitattributes` template with `linguist-generated` markers and changes the copy logic to use sentinel-based merging — preserving existing user content while keeping the managed section up-to-date across `sl init` and `sl doctor` update commands.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), Go embed FS, GoReleaser (build/release)
**Storage**: Embedded filesystem (`pkg/embedded/`) + local file I/O
**Testing**: `go test` with table-driven tests
**Target Platform**: Cross-platform CLI (darwin, linux, windows)
**Project Type**: Single project (Go CLI)
**Performance Goals**: N/A (file operation, runs once during init)
**Constraints**: Must be idempotent, must preserve user content in `.gitattributes`
**Scale/Scope**: 5 files modified, 3 new files (merge.go, merge_test.go, gitattributes_test.go)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution is not yet configured for this project (template placeholders only). Checking against default gates:

- [x] **Specification-First**: Spec.md complete with 3 prioritized user stories, clarifications resolved
- [x] **Test-First**: Test strategy defined — unit tests for merge function (table-driven), manual integration test via `sl init`
- [x] **Code Quality**: Go standard tooling (`go vet`, `go fmt`, `golangci-lint` via CI)
- [x] **UX Consistency**: User flows documented in spec acceptance scenarios (US1-US3)
- [x] **Performance**: N/A — single file operation during init, no performance concerns
- [x] **Observability**: Verbose mode already exists in CopyOptions for logging merge operations
- [x] **Issue Tracking**: GitHub Issue #74 exists; Epic SL-7bb372 created with 6 phases and 19 tasks

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/609-gitattributes-merge/
├── plan.md              # This file
├── research.md          # Phase 0: prior work + spike findings
├── spec.md              # Feature specification
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (/specledger.tasks)
```

### Source Code (repository root)

```text
pkg/cli/playbooks/
├── copy.go              # MODIFY: wire mergeableMap, add mergeFile()
├── template.go          # MODIFY: add Mergeable field to Playbook, FilesMerged to CopyResult
├── templates.go         # MODIFY: report merged files count
├── merge.go             # NEW: MergeSentinelSection() pure function
└── merge_test.go        # NEW: table-driven tests for merge logic

pkg/embedded/templates/
├── manifest.yaml        # MODIFY: add mergeable list
└── specledger/
    └── .gitattributes   # MODIFY: populate with linguist-generated patterns

tests/integration/
└── gitattributes_test.go # NEW: 8 integration tests (sl init + sl doctor flows)
```

**Structure Decision**: This is a Go CLI project. All changes are within the existing `pkg/cli/playbooks/` and `pkg/embedded/templates/` packages, plus integration tests in `tests/integration/`. No new packages or structural changes needed.
