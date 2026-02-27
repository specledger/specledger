# Implementation Plan: Advanced Agent Model Configuration

**Branch**: `597-agent-model-config` | **Date**: 2026-02-21 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specledger/597-agent-model-config/spec.md`

## Summary

Add persistent, layered configuration for agent launch environment variables (model names, base URLs, auth tokens, custom env vars) to the SpecLedger CLI. Extends the existing `Config` struct and `ProjectMetadata` with a config key registry, three-tier merge hierarchy (personal-local > team-local > global), named profiles, and a new `sl config` command with CLI subcommands and interactive TUI. The agent launcher gains environment variable injection from resolved configuration.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), Bubble Tea v1.3.10 + Bubbles v0.21.1 + Lipgloss v1.1.0 (TUI), huh v0.8.0 (forms), YAML v3 (config serialization), go-git v5
**Storage**: File-based — YAML for configuration (`~/.specledger/config.yaml` global, `specledger/specledger.yaml` team-local, `specledger/specledger.local.yaml` personal-local)
**Testing**: Go standard testing (`go test`), golangci-lint (govet, staticcheck, errcheck, ineffassign, unused, gosec)
**Target Platform**: macOS, Linux (distributed via Homebrew + GoReleaser)
**Project Type**: Single CLI project
**Performance Goals**: Sub-second config operations (file I/O only, no network)
**Constraints**: Must work offline. Personal-local config gitignored by default. CLI warns when sensitive values target git-tracked scope. No writes to Claude Code's own settings files.
**Scale/Scope**: ~20 config keys, 3 config files per scope, profiles embedded in config files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Note: Project constitution is in placeholder/template state (not yet populated). Checking against codebase-inferred standards.

- [x] **Specification-First**: spec.md complete with 4 prioritized user stories (P1, P2×2, P3), 13 functional requirements, acceptance scenarios for all stories
- [x] **Test-First**: Test strategy defined — unit tests for config merge/validation, integration tests for CLI commands, TUI snapshot tests
- [x] **Code Quality**: golangci-lint with govet, staticcheck, errcheck, ineffassign, unused, gosec (existing CI pipeline)
- [x] **UX Consistency**: Acceptance scenarios defined for all 4 user stories; quickstart.md defines UX flows
- [x] **Performance**: CLI tool — sub-second file I/O operations, no network dependency for config operations
- [x] **Observability**: CLI tool — errors reported via stderr with clear messages, no telemetry needed for config operations
- [ ] **Issue Tracking**: Beads epic needed for this feature (to be created during task generation)

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/597-agent-model-config/
├── plan.md              # This file
├── spec.md              # Feature specification (revised)
├── research.md          # Phase 0 output — consolidated research findings
├── research/            # Detailed research spikes
│   ├── 001-claude-cli-launch-options.md   # Claude Code env vars, flags, config hierarchy
│   └── 002-config-precedence-patterns.md  # Precedence patterns from established tools
├── data-model.md        # Phase 1 output — entity definitions and config schemas
├── quickstart.md        # Phase 1 output — UX walkthrough
├── revision-log.md      # Spec revision audit trail
└── tasks.md             # Phase 2 output (NOT created by /specledger.plan)
```

### Source Code (repository root)

```text
pkg/cli/
├── config/
│   ├── config.go            # EXTEND: Add AgentConfig nested struct with all agent keys
│   ├── schema.go            # NEW: Config key registry (type, env var mapping, default, validation)
│   ├── merge.go             # NEW: Multi-layer config merge logic (personal-local > team-local > global)
│   ├── profile.go           # NEW: Profile CRUD operations
│   ├── schema_test.go       # NEW: Unit tests for key registry and validation
│   ├── merge_test.go        # NEW: Unit tests for config merge precedence
│   └── profile_test.go      # NEW: Unit tests for profile operations
├── commands/
│   ├── config.go            # NEW: sl config command (set, get, show, unset subcommands)
│   ├── config_profile.go    # NEW: sl config profile subcommands (create, use, list, delete)
│   ├── bootstrap_helpers.go # MODIFY: Use resolved config for agent launch env vars
│   └── config_test.go       # NEW: Integration tests for config CLI
├── launcher/
│   └── launcher.go          # EXTEND: Add SetEnv/BuildEnv methods for env
│       # Consumers (no code duplication — all go through this package):
│       #   sl new   → bootstrap_helpers.go:launchAgent() → Launch()
│       #   sl init  → bootstrap_helpers.go:launchAgent() → Launch()
│       #   sl revise → revise.go → LaunchWithPrompt(prompt)
│       # TUI refs (agent selection only, no launch):
│       #   sl_new.go  → DefaultAgents list
│       #   sl_init.go → DefaultAgents list var injection
└── metadata/
    └── schema.go            # EXTEND: Add AgentConfig section to ProjectMetadata

cmd/sl/
└── main.go                  # MODIFY: Register VarConfigCmd
```

**Structure Decision**: Extends the existing single-project Go CLI structure. New files added to existing packages (`config/`, `commands/`, `launcher/`). No new top-level directories needed. Interactive TUI config editor descoped to a future spec (see `research/003-tui-framework-spike.md`).

## Design Decisions

### CLI Scope Flags

The `sl config set` and `sl config unset` commands accept scope-targeting flags:

| Flag | Target File | Git-Tracked | Description |
|------|-------------|-------------|-------------|
| *(default)* | `specledger/specledger.yaml` | **yes** | Team-local project config, shared via git |
| `--global` | `~/.specledger/config.yaml` | n/a | User-wide global defaults |
| `--personal` | `specledger/specledger.local.yaml` | **no** (gitignored) | Personal project overrides, not shared |

### Sensitive Field Guardrails

Fields tagged `sensitive:"true"` on the `AgentConfig` Go struct (currently `AuthToken`, `APIKey`) drive two behaviors:

1. **Display masking** — `sl config show` renders `****[last4]` (last 4 characters visible) instead of the full value
2. **Scope warning** — when `sl config set` stores a sensitive field in team-local scope (the default), the CLI emits a warning recommending `--personal` to avoid committing secrets to git. The user can proceed anyway (no `--force` required) but the warning is always shown.

This is best-effort guardrailing within SpecLedger. Teams should additionally adopt pre-commit secret detection (e.g., Yelp's `detect-secrets`, `gitleaks`, or `trufflehog`) as defense-in-depth. Recommending this in project README/onboarding is out of scope for this feature but is a natural follow-up.

## Complexity Tracking

> No violations identified. Feature extends existing patterns (Cobra commands, YAML config, Bubble Tea TUI) without introducing new architectural abstractions.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
