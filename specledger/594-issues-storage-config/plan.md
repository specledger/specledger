# Implementation Plan: Issues Storage Configuration

**Branch**: `594-issues-storage-config` | **Date**: 2026-02-20 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specledger/594-issues-storage-config/spec.md`

## Summary

Refactor the issues storage system to:
1. Rename lock files from `.issues.jsonl.lock` to `issues.jsonl.lock` (remove leading dot)
2. Use `artifact_path` from `specledger.yaml` as the base directory for issue storage
3. Add lock file pattern to .gitignore (both project and embedded templates)

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), go-git v5, gofrs/flock (file locking), gopkg.in/yaml.v3
**Storage**: File-based (JSONL for issues, file locks for concurrency)
**Testing**: Go standard testing (`go test`), integration tests in `tests/`
**Target Platform**: macOS, Linux, Windows (CLI)
**Project Type**: Single CLI application
**Performance Goals**: N/A (file-based local operations)
**Constraints**: N/A
**Scale/Scope**: Per-project, per-spec issue tracking

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Specification-First**: Spec.md complete with prioritized user stories
- [x] **Test-First**: Test strategy defined (unit tests for store, integration tests for CLI)
- [x] **Code Quality**: Go standard formatting (gofmt), existing golangci-lint config
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: N/A (file-based local operations, no latency concerns)
- [x] **Observability**: N/A (CLI tool, errors returned to user)
- [x] **Issue Tracking**: Built-in `sl issue` tracking enabled

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/594-issues-storage-config/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (repository root)

```text
pkg/
├── issues/
│   ├── store.go         # MODIFY: Lock file naming, artifact path support
│   └── migrate.go       # REVIEW: May need artifact path support
├── cli/
│   ├── commands/
│   │   └── issue.go     # MODIFY: Load artifact_path from config
│   └── metadata/
│       └── schema.go    # EXISTING: Already has GetArtifactPath()
└── embedded/
    └── templates/
        └── specledger/  # REVIEW: Check for .gitignore template

tests/
├── integration/
│   └── bootstrap_test.go  # EXISTING: Integration test patterns
└── issues/
    └── issue_test.go      # ADD: Tests for artifact_path behavior

.gitignore              # MODIFY: Add issues.jsonl.lock pattern
```

**Structure Decision**: Existing single-project CLI structure. Changes confined to `pkg/issues/store.go`, `pkg/cli/commands/issue.go`, and `.gitignore`.

## Complexity Tracking

No violations requiring justification.
