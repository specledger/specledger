# Implementation Plan: Embedded Templates

**Branch**: `001-embedded-templates` | **Date**: 2026-02-07 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-embedded-templates/spec.md`

## Summary

SpecLedger currently has templates in the `templates/` folder but they are not being copied to new projects during `sl new` or `sl init`. This feature adds automatic template copying functionality so that when users create a new project, the embedded Spec Kit playbook templates (Claude Code commands, skills, bash scripts, file templates, and beads configuration) are automatically set up in the project.

**Technical Approach**:
1. Create a template management package (`pkg/cli/templates`) that handles reading embedded templates from the `templates/` folder
2. Integrate template copying into the existing `sl new` and `sl init` workflows
3. Add `sl template list` command to display available embedded templates
4. Design the architecture to support future remote template fetching without major refactoring

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**:
- github.com/spf13/cobra (CLI framework)
- gopkg.in/yaml.v3 (YAML parsing)
- Existing: specledger/pkg/cli/* packages

**Storage**: File system (templates embedded in codebase at `templates/`, copied to user projects)
**Testing**: Go testing (`go test`), integration tests with temporary directories
**Target Platform**: Cross-platform (Linux, macOS, Windows) - same as SpecLedger CLI
**Project Type**: CLI tool with embedded resources
**Performance Goals**: Template copying should complete in < 1 second for typical template sizes
**Constraints**:
- Must preserve directory structure when copying
- Must handle existing files gracefully (skip, warn, or offer options)
- Must validate template folder exists before project creation
- Architecture must support future remote templates without refactoring
**Scale/Scope**:
- Current: ~40 template files in `templates/` folder
- Expected growth: Additional playbooks (OpenSpec, custom) in future
- File sizes: Small text files (markdown, bash, yaml, JSON)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Note**: No active constitution file exists for SpecLedger project. Using general software engineering best practices as guidelines:

| Principle | Status | Notes |
|-----------|--------|-------|
| Simplicity | ✅ PASS | Straightforward file copying, no complex state |
| Testability | ✅ PASS | Can test with temp directories and fake templates |
| Minimal Dependencies | ✅ PASS | Uses Go standard library file operations |
| Clear Purpose | ✅ PASS | Single responsibility: template management |
| Future-Proof | ✅ PASS | Architecture supports remote templates (deferred) |

**No violations to justify** - proceeding to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/001-embedded-templates/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (Go interfaces)
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
specledger/
├── cmd/
│   └── main.go              # CLI entry point
├── pkg/
│   ├── cli/
│   │   ├── commands/        # Existing commands
│   │   │   ├── bootstrap.go         # Modified: Add template copying
│   │   │   ├── bootstrap_helpers.go # Modified: Add copy helpers
│   │   │   └── templates.go        # NEW: template list command
│   │   ├── templates/        # NEW: Template management package
│   │   │   ├── templates.go       # Core template copying logic
│   │   │   ├── manifest.go        # Template manifest parsing
│   │   │   └── copy.go            # File/directory copying utilities
│   │   └── ...
│   └── embedded/            # NEW: Embed templates in binary
│       └── templates.go        //go:embed templates/
├── templates/               # Embedded template source (existing)
│   ├── specledger/
│   │   ├── scripts/
│   │   ├── templates/
│   │   ├── memory/
│   │   └── spec-kit-version
│   ├── .claude/
│   ├── .beads/
│   └── ...
└── tests/
    └── integration/
        └── templates_test.go  # Integration tests for template copying
```

**Structure Decision**: Single project structure (CLI tool). The new `pkg/cli/templates` package handles all template operations, `pkg/embedded` contains the embedded filesystem using Go's `embed` package, and the existing `pkg/cli/commands` are modified to call the new template functionality.

## Complexity Tracking

> No constitution violations - this section intentionally left empty.
