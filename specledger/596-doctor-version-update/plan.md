# Implementation Plan: Doctor Version and Template Update

**Branch**: `596-doctor-version-update` | **Date**: 2026-02-20 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specledger/596-doctor-version-update/spec.md`

## Summary

Enhance `sl doctor` to check CLI version against GitHub Releases API, display update instructions when outdated, and proactively offer to update project templates when version mismatch is detected. Template version tracking will be stored in `specledger.yaml` metadata.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), gopkg.in/yaml.v3, net/http (GitHub API)
**Storage**: File-based (specledger.yaml for metadata, embedded FS for templates)
**Testing**: Go testing framework (`go test ./...`)
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: Single project (Go CLI)
**Performance Goals**: Version check < 3 seconds, template update < 10 seconds
**Constraints**: Must work offline gracefully, must preserve customized template files
**Scale/Scope**: Single CLI binary, unlimited projects

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Specification-First**: Spec.md complete with 3 prioritized user stories
- [x] **Test-First**: Test strategy defined (unit tests for version check, template comparison, integration tests for doctor command)
- [x] **Code Quality**: Go standard formatting (gofmt), golangci-lint
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: Metrics defined (< 3s for version check, < 10s for template update)
- [x] **Observability**: Standard Go logging, clear error messages for network failures
- [x] **Issue Tracking**: Issues will be created via `sl issue create` for tracking

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/596-doctor-version-update/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/specledger.tasks)
```

### Source Code (repository root)

```text
pkg/
├── cli/
│   └── commands/
│       └── doctor.go        # MODIFY - Add version check, template update prompt
│
├── version/
│   ├── checker.go           # NEW - GitHub API version checking
│   └── checker_test.go      # NEW - Unit tests
│
├── templates/
│   ├── updater.go           # NEW - Template update logic
│   ├── updater_test.go      # NEW - Unit tests
│   └── diff.go              # NEW - Template diff/comparison
│
├── cli/
│   └── metadata/
│       └── schema.go        # MODIFY - Add TemplateVersion field to ProjectMetadata
│
pkg/embedded/
└── embedded.go              # EXISTING - Embedded templates (no changes)
```

**Structure Decision**: Create new `pkg/version/` package for version checking logic and `pkg/templates/` package for template update logic. Modify existing `doctor.go` and `schema.go`. This follows the existing pattern of separating concerns into packages.

## Complexity Tracking

No violations to track.

## Phase 0: Research Summary

See [research.md](./research.md) for detailed findings.

### Key Decisions

1. **Version Check API**: Use GitHub Releases API (`/repos/owner/repo/releases/latest`) with 5-second timeout
2. **Template Version Storage**: Add `template_version` field to `ProjectMetadata.TemplateVersion` in specledger.yaml
3. **Custom File Detection**: Compare file checksums (SHA-256) against embedded originals to detect modifications
4. **Interactive Prompt**: Use Bubble Tea TUI for update prompt (consistent with existing `sl init` patterns)

## Phase 1: Design Artifacts

See [data-model.md](./data-model.md) for entity details.
See [quickstart.md](./quickstart.md) for usage scenarios.
