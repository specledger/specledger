<!--
SYNC IMPACT REPORT
==================
Version change:       (new) → 1.0.0
Modified principles:  N/A — initial ratification
Added sections:       Core Principles (I–IV), Technology Standards, Development Workflow, Agent Preferences, Governance
Removed sections:     N/A
Templates updated:
  ✅ .specledger/templates/plan-template.md  — Constitution Check already references these principles by name
  ✅ .specledger/templates/spec-template.md  — Mandatory sections align with Specification-First principle
  ✅ .specledger/templates/tasks-template.md — Task types (observability, testing, polish) align with principles
Deferred TODOs:       None
-->

# SpecLedger Constitution

## Core Principles

### I. Specification-First

Every feature MUST have a complete `spec.md` before any code is written.
Specifications define user stories with prioritized acceptance scenarios, functional
requirements, and measurable success criteria — not implementation details. The spec
is the single source of truth for what is being built and why.

- Specs MUST be reviewed and approved before planning begins.
- `sl issue` MUST be used to create an epic linked to the spec before implementation.
- Implementation that deviates from the approved spec MUST update the spec first.

### II. Simplicity (YAGNI)

The simplest solution that satisfies the spec MUST be chosen. Complexity must be
explicitly justified in the plan's Complexity Tracking table. Abstractions,
generalization, and extra configurability are prohibited unless directly required
by an approved user story.

- Prefer deleting code over adding it when a simpler path exists.
- No speculative features, no future-proofing, no premature abstractions.
- Three similar lines of code are better than a premature helper function.

### III. UX Consistency

All `sl` commands MUST follow consistent UX patterns so users can predict behavior
across the CLI surface.

- Errors MUST surface with actionable suggestions using the standard `CLIError` type.
- Complex multi-step flows MUST offer an interactive TUI (Bubble Tea) with a
  non-interactive flag-based fallback for CI/headless environments.
- Output styling MUST use the `pkg/cli/ui` color/formatting helpers (Lipgloss).
- Command naming and flag conventions MUST be consistent with existing commands.

### IV. Observability

Significant operations MUST produce debuggable output. Silent failures are
prohibited.

- All errors MUST be logged with enough context to reproduce the failure.
- Destructive operations (file writes, network calls, git mutations) MUST log
  before executing.
- The `pkg/cli/logger` package MUST be used for structured logging; no bare
  `fmt.Println` for error paths.
- Session capture (`sl session`) SHOULD be used to record AI-assisted development
  sessions for post-hoc analysis.

## Technology Standards

- **Language**: Go 1.24+ — follow idiomatic Go conventions (effective Go, `go vet`, `golangci-lint`).
- **CLI Framework**: Cobra (`github.com/spf13/cobra`) — all commands exported as `Var*Cmd` vars.
- **TUI Framework**: Bubble Tea (`github.com/charmbracelet/bubbletea`) + Bubbles + Lipgloss.
- **Git Operations**: `go-git/v5` — no shelling out to `git` binary.
- **Config/Metadata**: `gopkg.in/yaml.v3` for all YAML serialization.
- **File Locking**: `github.com/gofrs/flock` for concurrent JSONL writes.
- **Storage**: File-based (JSONL for issues, YAML for metadata). No mandatory external database.
- **Build & Release**: Make + GoReleaser v2 via GitHub Actions. Single binary output.
- **Testing**: `go test ./...` for unit tests; `tests/integration/` for integration tests.

## Development Workflow

- **Branch naming**: `<issue-number>-<short-description>` (e.g., `593-init-project-templates`).
- **Spec before branch**: Create spec and epic issue before opening a feature branch.
- **PR requirement**: All changes MUST go through a pull request; direct pushes to `main` are prohibited.
- **Constitution Check**: Every `plan.md` MUST include a completed Constitution Check section
  before Phase 0 research begins.
- **Issue tracking**: ALL tasks and phases MUST be tracked via `sl issue` CLI — not inline
  in markdown files.

## Agent Preferences

- **Preferred Agent**: Claude Code (claude-sonnet-4-6 or newer)

## Governance

This constitution supersedes all other coding practices, style guides, and informal
conventions within the SpecLedger repository.

- All PRs MUST verify compliance with all four Core Principles before merge.
- Amendments require: (1) a documented rationale, (2) a version bump per the
  semantic versioning policy below, (3) consistency propagation across all
  dependent templates.
- **Versioning policy**:
  - MAJOR — backward-incompatible removal or redefinition of a principle.
  - MINOR — new principle or section added, or materially expanded guidance.
  - PATCH — clarification, wording refinement, typo fix.
- The `CLAUDE.md` file at the repository root is auto-generated from feature plans
  and MUST NOT be edited manually; it does not override this constitution.

**Version**: 1.0.0 | **Ratified**: 2026-02-20 | **Last Amended**: 2026-02-20
