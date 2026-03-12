# Implementation Plan: Fix Embedded Skill Templates

**Branch**: `607-fix-skill-templates` | **Date**: 2026-03-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specledger/607-fix-skill-templates/spec.md`

## Summary

Fix critical issues in embedded skill templates: replace incorrect `sl-deps` content (currently duplicate of `sl-issue-tracking`), remove duplicate sections in `sl-audit` (~700 tokens wasted), enhance skill descriptions with trigger keywords, and remove aspirational content referencing non-existent commands.

## Technical Context

**Language/Version**: Go 1.24.2 (existing codebase)
**Primary Dependencies**: Cobra CLI, embedded templates in `pkg/embedded/`
**Storage**: N/A (static markdown files)
**Testing**: `go test ./pkg/embedded/...` (existing tests verify template embedding)
**Target Platform**: Cross-platform CLI tool
**Project Type**: Single project (CLI application)
**Performance Goals**: N/A (content changes only)
**Constraints**: Must maintain backward compatibility with skill loading mechanism
**Scale/Scope**: 4 skill files (~900 lines total), 1 manifest.yaml

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with principles from `.specledger/memory/constitution.md`:

- [x] **Specification-First**: Spec.md complete with prioritized user stories
- [x] **Test-First**: Test strategy defined (content verification via file comparison)
- [x] **Code Quality**: Linting/formatting tools identified (gofmt, existing CI)
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: Token reduction measurable (700+ tokens target)
- [x] **Observability**: N/A (static content changes)
- [x] **Issue Tracking**: Linked to GitHub issue #82

**Complexity Violations**: None identified - this is a content fix, not architectural change.

## Project Structure

### Documentation (this feature)

```text
specledger/607-fix-skill-templates/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (via /specledger.tasks)
```

### Source Code (repository root)

```text
pkg/embedded/templates/specledger/
├── manifest.yaml                    # Skill descriptions (update trigger keywords)
└── skills/
    ├── sl-audit/
    │   └── skill.md                 # Remove duplicate sections (lines 239-271)
    ├── sl-comment/
    │   └── skill.md                 # No changes needed (reference model)
    ├── sl-deps/
    │   └── skill.md                 # REPLACE content with correct sl deps guidance
    └── sl-issue-tracking/
        └── skill.md                 # No changes needed (correct content)
```

**Structure Decision**: Single project with embedded templates. Changes are isolated to `pkg/embedded/templates/specledger/skills/` and `manifest.yaml`.

## Complexity Tracking

> No violations - content fix only.

| Aspect | Status | Notes |
|--------|--------|-------|
| New code | None | Markdown content changes only |
| Architecture change | None | Existing template structure preserved |
| Dependencies | None | No new imports or packages |
