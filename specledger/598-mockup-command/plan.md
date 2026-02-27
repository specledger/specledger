# Implementation Plan: Mockup Command

**Branch**: `598-mockup-command` | **Date**: 2026-02-27 | **Spec**: [specledger/598-mockup-command/spec.md](spec.md)
**Input**: Feature specification from `specledger/598-mockup-command/spec.md`

## Summary

Add `sl mockup <spec-name>` and `sl mockup update` commands that detect frontend frameworks, scan codebases for UI components, build a design system index (`specledger/design_system.md`), and generate ASCII mockups from feature specs mapped to the project's actual components. Extends `sl init` to auto-create the design system for frontend projects during onboarding.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), go-git v5 (git), gopkg.in/yaml.v3 (YAML parsing)
**Storage**: File-based — Markdown with YAML frontmatter (`design_system.md`), Markdown (`mockup.md`)
**Testing**: `go test` (unit tests), table-driven tests following existing patterns
**Target Platform**: macOS, Linux (CLI binary via GoReleaser)
**Project Type**: Single CLI binary
**Performance Goals**: Mockup generation <30s (SC-001/SC-005), design system update <10s (SC-007)
**Constraints**: Offline-capable (no external API calls for detection/scanning), skip `node_modules`/`vendor`
**Scale/Scope**: Scan up to ~1000 component files, 4 framework families (React/Vue/Svelte/Angular)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution is in template state (not yet ratified). Proceeding with best-practice gates:

- [x] **Specification-First**: Spec.md complete with 5 prioritized user stories (P1-P3)
- [x] **Test-First**: Test strategy defined — unit tests for detector, scanner, generator; integration tests for end-to-end CLI flow
- [x] **Code Quality**: golangci-lint configured in CI (`.golangci.yml`), `gofmt` enforced
- [x] **UX Consistency**: User flows documented in acceptance scenarios (US1-US5)
- [x] **Performance**: Metrics defined — <30s mockup, <10s update (SC-001, SC-005, SC-007)
- [x] **Observability**: CLI output with progress indicators (checkmarks, timing), `--json` flag for machine output
- [ ] **Issue Tracking**: No issue tracker epic required (constitution not ratified)

**Complexity Violations**: None identified

### Post-Design Re-check

- [x] Data model covers all entities from spec (DetectionResult, Component, DesignSystem, Mockup, ScanResult)
- [x] CLI contract covers both commands (`sl mockup <spec-name>`, `sl mockup update`) with flags, exit codes, error messages
- [x] Integration point with `sl init` documented in contracts/cli.md
- [x] No new external Go dependencies required — uses stdlib + existing yaml.v3

## Project Structure

### Documentation (this feature)

```text
specledger/598-mockup-command/
├── plan.md              # This file
├── research.md          # Phase 0 output (complete)
├── data-model.md        # Phase 1 output (complete)
├── contracts/
│   └── cli.md           # Phase 1 output (complete)
└── tasks.md             # Phase 2 output (created by /specledger.tasks)
```

### Source Code (repository root)

```text
pkg/cli/
├── commands/
│   └── mockup.go              # Cobra command definitions (VarMockupCmd, mockupUpdateCmd)
└── mockup/                    # Domain logic package (NEW)
    ├── detector.go            # Frontend framework detection (Tier 1-3 heuristics)
    ├── detector_test.go       # Detection unit tests
    ├── scanner.go             # Component scanning per framework
    ├── scanner_test.go        # Scanner unit tests
    ├── designsystem.go        # Design system index read/write/merge
    ├── designsystem_test.go   # Design system I/O tests
    ├── generator.go           # Mockup generation from spec + design system
    ├── generator_test.go      # Generator unit tests
    └── types.go               # Shared types (FrameworkType, Component, etc.)

pkg/cli/commands/
└── bootstrap.go               # MODIFIED — add frontend detection + design system init
```

**Structure Decision**: Follows the established pattern of domain logic in `pkg/cli/<feature>/` (like `revise/`, `session/`, `playbooks/`) with command definitions in `pkg/cli/commands/mockup.go`. All new types live in `pkg/cli/mockup/types.go` to avoid circular imports.

## Previous Work

| Spec | Relevance | Reuse |
|------|-----------|-------|
| **597-issue-create-fields** | Most recent CLI command | Cobra patterns, flag handling, `--json` output |
| **011-streamline-onboarding** | Extends onboarding flow | `sl init` integration point in `bootstrap.go` |
| **591-issue-tracking-upgrade** | File-based storage | `pkg/issues/context.go` for spec detection |
| **136-revise-comments** | Domain logic separation | `pkg/cli/revise/` package pattern |
| **596-doctor-version-update** | File scanning | File traversal patterns |

## Architecture

### Command Flow

```
sl mockup <spec-name>
  │
  ├─ 1. Validate spec exists (specledger/<spec-name>/spec.md)
  ├─ 2. Detect frontend framework (detector.go)
  │     ├─ Tier 1: Config files (next.config.js, angular.json, etc.)
  │     ├─ Tier 2: package.json dependencies
  │     └─ Tier 3: File extension scan
  ├─ 3. Check --force flag if not frontend
  ├─ 4. Load or generate design system (designsystem.go)
  │     ├─ If exists: parse YAML frontmatter + markdown
  │     └─ If missing: scan components (scanner.go) → write file
  ├─ 5. Parse spec user stories and requirements
  ├─ 6. Generate mockup screens (generator.go)
  │     ├─ Map UI needs to design system components
  │     ├─ Generate ASCII layouts per screen
  │     └─ Build component mapping table
  └─ 7. Write mockup.md to feature directory

sl mockup update
  │
  ├─ 1. Validate design_system.md exists
  ├─ 2. Load existing design system (preserve manual entries)
  ├─ 3. Rescan components (scanner.go)
  ├─ 4. Merge: add new, remove stale, keep manual
  └─ 5. Write updated design_system.md
```

### Key Design Decisions

1. **Domain package at `pkg/cli/mockup/`** — Follows `revise/`, `session/` pattern. Keeps detection, scanning, and generation logic separate from Cobra command wiring.

2. **Tiered framework detection** — Config files first (99% confidence), package.json fallback, file extension last resort. Returns `DetectionResult` with confidence score; `IsFrontend` only if confidence >= 70.

3. **YAML frontmatter + Markdown for design_system.md** — Machine-parseable metadata (version, framework, last_scanned) in frontmatter, human-readable component index in markdown body. Supports manual edits via `<!-- MANUAL -->` markers.

4. **Glob + regex scanning per framework** — Framework-specific glob patterns (`**/*.tsx`, `**/*.vue`, etc.) with content-based component identification. Skips `node_modules/`, `vendor/`, `.git/`, `dist/`, `build/`.

5. **ASCII mockup output** — Version-controllable, diffable, no rendering dependencies. Each screen is a text box with component annotations linking to design system entries.

6. **No new external dependencies** — Uses stdlib (`path/filepath`, `os`, `regexp`, `strings`, `encoding/json`) plus existing `gopkg.in/yaml.v3`. Avoids adding bloat.

### Codebase Integration Points

| Integration | File | Change |
|-------------|------|--------|
| Command registration | `cmd/sl/main.go` | Add `rootCmd.AddCommand(commands.VarMockupCmd)` |
| Command definition | `pkg/cli/commands/mockup.go` | New file — `VarMockupCmd` + `mockupUpdateCmd` |
| Init flow | `pkg/cli/commands/bootstrap.go` | Add frontend detection + design system init after base setup |
| Spec context | `pkg/issues/context.go` | Reuse `NewContextDetector` for auto-detecting spec name |
| UI output | `pkg/cli/ui/` | Reuse `ui.Checkmark()`, `ui.Bold()` for consistent output |

## Complexity Tracking

> No violations identified. Single package addition follows established patterns.

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Package count | +1 (`pkg/cli/mockup/`) | Follows existing domain package pattern |
| New files | ~9 (4 source + 4 test + 1 command) | Standard for a feature of this scope |
| External deps | 0 new | All needed functionality in stdlib + yaml.v3 |
| Modified files | 2 (`main.go`, `bootstrap.go`) | Minimal integration surface |
