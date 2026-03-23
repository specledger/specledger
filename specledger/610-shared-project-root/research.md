# Research: Shared Project Root Resolution

**Date**: 2026-03-23
**Feature**: 610-shared-project-root

## Prior Work

- **596-doctor-version-update**: Implemented the current `sl doctor` with template status, version check, and JSON output. The `os.Getwd()` pattern was introduced here.
- **598-mockup-command**: Extracted shared editor and prompt utilities to `pkg/cli/shared/` — establishes the pattern for shared utilities in this codebase.
- **609-gitattributes-merge**: Recent work on `sl doctor --template` behavior, including merge strategy. Same `os.Getwd()` pattern.
- **deps.go `findProjectRoot()`**: Working implementation since deps command was added. Uses `metadata.HasYAMLMetadata()` to walk up the directory tree.

## Decision: Where to place the shared utility

**Decision**: Place `FindProjectRoot()` in `pkg/cli/metadata/` package as an exported function.

**Rationale**:
- The `metadata` package already owns `HasYAMLMetadata()` which is the core check used by `findProjectRoot()`.
- The function is fundamentally about locating the metadata file — it belongs with the metadata package.
- The existing `findProjectRoot()` in `deps.go` already imports and uses `metadata.HasYAMLMetadata()`.
- This avoids creating a new package for a single function (YAGNI principle).

**Alternatives considered**:
- `pkg/cli/project/root.go` (new package): Rejected — creating a new package for one function violates YAGNI. The metadata package is the natural home.
- `pkg/cli/shared/`: Rejected — this package holds UI/editor utilities, not project discovery logic.
- Keep in `deps.go` and import: Rejected — `deps.go` is a command file, not a utility package. Other command files can't import from it without circular dependencies.

## Decision: Which `os.Getwd()` call sites to update

**Decision**: Update only call sites that need the project root for project-level operations. Leave `os.Getwd()` in place for commands that intentionally operate on the current directory.

**Rationale**: Not every `os.Getwd()` is a bug. Some commands (like `bootstrap`) create projects at the current directory. The distinction is: "I need the project root" vs. "I need the current directory."

**Call sites to update** (need project root):
| File | Line | Function | Current Usage |
|------|------|----------|---------------|
| `doctor.go` | 141 | `performTemplateUpdate()` | Uses cwd as project dir |
| `doctor.go` | 215 | `outputDoctorJSON()` | Uses cwd for template status |
| `doctor.go` | 296 | `outputDoctorHuman()` | Uses cwd for template status |
| `session.go` | 178, 189, 264 | Multiple functions | Uses cwd for session operations |
| `comment.go` | 128 | Comment command | Uses cwd for project context |
| `revise.go` | 60 | Revise command | Uses cwd for project context |
| `mockup.go` | 80, 352 | Mockup command | Uses cwd for project context |
| `spec_info.go` | 61 | Spec info command | Uses cwd for project context |
| `context_update.go` | 63 | Context update command | Uses cwd for project context |
| `spec_create.go` | 64 | Spec create command | Uses cwd for project context |
| `spec_setup_plan.go` | 51 | Plan setup command | Uses cwd for project context |

**Call sites to leave unchanged** (intentionally use cwd):
| File | Line | Function | Reason |
|------|------|----------|--------|
| `bootstrap.go` | 281 | Bootstrap/init | Creates project at cwd |
| `session/capture.go` | 519 | Session capture | Records current directory, not project root |
| `config_test.go` | 285, 345, 392 | Tests | Test helpers manipulating cwd |

## Decision: Error message format

**Decision**: Use a consistent error message: `"not in a SpecLedger project (no specledger.yaml found). Run 'sl init' to create one, or navigate to a project directory."`

**Rationale**: Constitution Principle IX (Fail Fast, Fix Forward) requires clear, actionable error messages. The current `deps.go` message references an internal path; the new message is user-friendly and suggests next steps.

## Decision: Function signature

**Decision**: `func FindProjectRoot() (string, error)` — exported, no parameters. Uses `os.Getwd()` internally.

**Rationale**: Mirrors the existing `findProjectRoot()` signature from `deps.go`. No parameters needed since it always starts from cwd. Exported so all command packages can use it.

For testability, also provide: `func FindProjectRootFrom(startDir string) (string, error)` — allows tests to specify a starting directory without changing the working directory.
