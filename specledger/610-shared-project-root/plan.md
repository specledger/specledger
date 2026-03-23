# Implementation Plan: Shared Project Root Resolution

**Branch**: `610-shared-project-root` | **Date**: 2026-03-23 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specledger/610-shared-project-root/spec.md`

## Summary

Extract the private `findProjectRoot()` function from `pkg/cli/commands/deps.go` into the existing `pkg/cli/metadata/` package as an exported `FindProjectRoot()` utility. Update `doctor.go` and all other commands that incorrectly use `os.Getwd()` for project context to use the shared utility instead. This fixes GitHub issue #81 where `sl doctor --template` fails from subdirectories.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: `pkg/cli/metadata` (HasYAMLMetadata), Cobra (CLI framework)
**Storage**: File-based (specledger.yaml detection)
**Testing**: `go test` with table-driven tests
**Target Platform**: Cross-platform (darwin, linux, windows)
**Project Type**: Single CLI application
**Performance Goals**: N/A — filesystem stat operations are negligible
**Constraints**: Must not break existing commands; must work on all supported platforms
**Scale/Scope**: ~15 call sites across 11 command files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Specification-First**: Spec.md complete with prioritized user stories
- [x] **Test-First**: Test strategy defined — table-driven unit tests for FindProjectRootFrom() with temp directory trees
- [x] **Code Quality**: `golangci-lint v2`, `gofmt` — already in Makefile
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: N/A — directory stat operations are sub-millisecond
- [x] **Observability**: Error messages are clear and actionable (Principle IX)
- [x] **Issue Tracking**: Epic to be created with `sl issue create --type epic`

**Complexity Violations**: None identified. This is a straightforward extract-and-replace refactor.

## Project Structure

### Documentation (this feature)

```text
specledger/610-shared-project-root/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 research output
├── data-model.md        # Phase 1 — minimal, no new entities
├── quickstart.md        # Phase 1 — validation scenarios
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (created by /specledger.tasks)
```

### Source Code (repository root)

```text
pkg/cli/metadata/
├── yaml.go              # Existing — HasYAMLMetadata lives here
├── root.go              # NEW — FindProjectRoot() and FindProjectRootFrom()
├── root_test.go         # NEW — table-driven tests
├── schema.go            # Existing
├── schema_test.go       # Existing
└── yaml_test.go         # Existing

pkg/cli/commands/
├── deps.go              # MODIFIED — remove findProjectRoot(), use metadata.FindProjectRoot()
├── doctor.go            # MODIFIED — replace 3x os.Getwd() with metadata.FindProjectRoot()
├── session.go           # MODIFIED — replace 3x os.Getwd() with metadata.FindProjectRoot()
├── comment.go           # MODIFIED — replace os.Getwd()
├── revise.go            # MODIFIED — replace os.Getwd()
├── mockup.go            # MODIFIED — replace 2x os.Getwd()
├── spec_info.go         # MODIFIED — replace os.Getwd()
├── context_update.go    # MODIFIED — replace os.Getwd()
├── spec_create.go       # MODIFIED — replace os.Getwd()
└── spec_setup_plan.go   # MODIFIED — replace os.Getwd()
```

**Structure Decision**: No new packages. The shared function lives in the existing `pkg/cli/metadata/` package alongside `HasYAMLMetadata()` which it depends on. This follows the established pattern of keeping related functionality together.

## Implementation Phases

### Phase 1: Extract Shared Utility (P1 — Core Fix)

1. Create `pkg/cli/metadata/root.go` with:
   - `FindProjectRootFrom(startDir string) (string, error)` — core logic, testable
   - `FindProjectRoot() (string, error)` — convenience wrapper using `os.Getwd()`
   - Consistent error message: `"not in a SpecLedger project (no specledger.yaml found). Run 'sl init' to create one, or navigate to a project directory."`

2. Create `pkg/cli/metadata/root_test.go` with table-driven tests:
   - Find root from project root directory
   - Find root from nested subdirectory (2-3 levels deep)
   - Error when no specledger.yaml exists anywhere
   - Error from filesystem root
   - Find root with symlinked directory

3. Update `pkg/cli/commands/deps.go`:
   - Remove private `findProjectRoot()` function
   - Replace all calls with `metadata.FindProjectRoot()`

### Phase 2: Fix Doctor Command (P1 — Issue #81)

4. Update `pkg/cli/commands/doctor.go`:
   - `performTemplateUpdate()` line 141: replace `os.Getwd()` with `metadata.FindProjectRoot()`
   - `outputDoctorJSON()` line 215: replace `os.Getwd()` with `metadata.FindProjectRoot()`
   - `outputDoctorHuman()` line 296: replace `os.Getwd()` with `metadata.FindProjectRoot()`

### Phase 3: Update Remaining Commands (P2 — Consistency)

5. Update all other command files that use `os.Getwd()` for project context:
   - `session.go` (3 call sites)
   - `comment.go` (1 call site)
   - `revise.go` (1 call site)
   - `mockup.go` (2 call sites)
   - `spec_info.go` (1 call site)
   - `context_update.go` (1 call site)
   - `spec_create.go` (1 call site)
   - `spec_setup_plan.go` (1 call site)

### Phase 4: Validation

6. Run `make test` to verify no regressions
7. Run `make lint` to verify code quality
8. Manual validation: run `sl doctor --template` from a subdirectory

## Complexity Tracking

> No violations — this is a simple extract-and-replace refactor with no new abstractions beyond what's strictly needed.
