# Implementation Plan: Fix Executable Permissions for Template Scripts

**Branch**: `135-fix-missing-chmod-x` | **Date**: 2026-02-17 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specledger/135-fix-missing-chmod-x/spec.md`

## Summary

Fix a bug where shell scripts copied from embedded templates during `sl init` and `sl new` lack execute permissions. The fix modifies two file copy functions to detect executable files (by `.sh` extension or shebang) and set `0755` permissions instead of `0644`.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: Existing: Cobra (CLI), embedded FS
**Storage**: N/A (file permissions only)
**Testing**: Go standard testing (`go test`)
**Target Platform**: Cross-platform (Linux, macOS, Windows via Git Bash/WSL)
**Project Type**: Single project (CLI tool)
**Performance Goals**: N/A (negligible overhead from permission check)
**Constraints**: Unix filesystem permissions assumed; Windows users typically use Git Bash
**Scale/Scope**: 2 functions to modify, ~20 lines of code change

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Specification-First**: Spec.md complete with prioritized user stories
- [x] **Test-First**: Test strategy defined (unit tests for permission detection, integration tests for `sl init`)
- [x] **Code Quality**: Existing Go tooling (gofmt, go vet)
- [x] **UX Consistency**: Acceptance scenarios document expected behavior
- [x] **Performance**: N/A - simple permission check adds no measurable overhead
- [x] **Observability**: N/A - file copy operation, errors returned as normal
- [x] **Issue Tracking**: Feature tracked via branch `135-fix-missing-chmod-x`

**Complexity Violations**: None - this is a simple bug fix

## Project Structure

### Documentation (this feature)

```text
specledger/135-fix-missing-chmod-x/
├── plan.md              # This file
├── spec.md              # Feature specification
├── checklists/          # Quality checklists
└── tasks.md             # Implementation tasks (created by /specledger.tasks)
```

### Source Code (repository root)

```text
pkg/
├── cli/
│   ├── commands/
│   │   └── bootstrap_helpers.go  # applyEmbeddedSkills() - needs fix
│   └── playbooks/
│       └── copy.go               # copyEmbeddedFile() - needs fix
└── embedded/
    └── templates/specledger/     # Template files (unchanged)

pkg/cli/playbooks/
├── copy_test.go        # Existing tests - may need updates
└── copy.go             # Primary fix location
```

**Structure Decision**: Single project structure. Changes isolated to two functions in `pkg/cli/` package.

## Complexity Tracking

No complexity violations - simple bug fix affecting two functions.

## Phase 0: Research Summary

**Not required** - the bug is well-understood:
- Root cause: `os.WriteFile()` called with `0644` (no execute bit)
- Affected code: `copyEmbeddedFile()` and `applyEmbeddedSkills()`
- Fix: Detect executable files and use `0755` permissions

## Phase 1: Design

### Files to Modify

| File | Function | Change |
|------|----------|--------|
| `pkg/cli/playbooks/copy.go` | `copyEmbeddedFile()` | Add executable detection, use `0755` for scripts |
| `pkg/cli/commands/bootstrap_helpers.go` | `applyEmbeddedSkills()` | Add executable detection, use `0755` for scripts |
| `pkg/cli/playbooks/copy_test.go` | Tests | Add tests for permission detection |

### Implementation Approach

1. **Create helper function** `isExecutableFile(filename string, content []byte) bool`:
   - Returns `true` if file has `.sh` extension
   - Returns `true` if first line starts with `#!` (shebang)
   - Returns `false` otherwise

2. **Modify `copyEmbeddedFile()`** in `copy.go`:
   - Call `isExecutableFile()` after reading content
   - Use `0755` if executable, `0644` otherwise

3. **Modify `applyEmbeddedSkills()`** in `bootstrap_helpers.go`:
   - Same logic as above for skills files

4. **Add unit tests**:
   - Test `.sh` extension detection
   - Test shebang detection
   - Test non-executable files remain `0644`

### Edge Cases Handled

| Case | Behavior |
|------|----------|
| File already has execute permission | Overwrite with correct permission (always set `0755`) |
| Filesystem doesn't support exec bit | `os.Chmod` succeeds, exec fails later (acceptable) |
| File with shebang but no `.sh` extension | Gets `0755` (FR-002) |
| Empty file | Check shebang safely (no crash) |

## External Dependencies

None - no external specifications referenced.
