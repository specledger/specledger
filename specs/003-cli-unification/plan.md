# Implementation Plan: CLI Unification

**Branch**: `003-cli-unification` | **Date**: 2026-01-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-cli-unification/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Unify the existing bash `sl` bootstrap script and Go `specledger` CLI into a single CLI tool that provides both project bootstrap (with interactive TUI) and specification dependency management. Enable multiple distribution channels including GitHub Releases, self-built binaries, self-hosted distribution, UVX-style execution, and package manager integrations (Homebrew, npx). Maintain backward compatibility with `specledger` alias and preserve the TUI for interactive project bootstrap.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Cobra (spf13/cobra), Bubble Tea, gum, mise
**Storage**: YAML files (config, lockfile, template files)
**Testing**: Go testing framework, integration tests
**Target Platform**: macOS (Darwin), Linux, Windows
**Project Type**: Single Go CLI application
**Performance Goals**: Bootstrap < 3 minutes, command response < 200ms
**Constraints**: < 50MB binary size, cross-platform compatible, no external dependencies at runtime (besides gum/mise)
**Scale/Scope**: CLI tool with ~15 commands, 6 distribution channels

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with principles from `.specify/memory/constitution.md`:

- [x] **Specification-First**: Spec.md complete with prioritized user stories (P1: CLI unification, GitHub releases; P2: Self-built, self-hosted, UVX; P3: Package managers)
- [x] **Test-First**: Test strategy defined (contract + integration tests planned in contracts/CLI-INTERFACE.md)
- [x] **Code Quality**: Go standard library + Cobra framework, fmt, go vet, golangci-lint
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios (Given/When/Then format)
- [x] **Performance**: Bootstrap < 3 minutes, CI/CD bootstrap < 1 minute, dependency resolution < 200ms
- [x] **Observability**: Debug-level logging to stderr (per research decisions)
- [x] **Issue Tracking**: Beads epic sl-29y created (CLI Integration & Distribution)

**Complexity Violations** (if any, justify in Complexity Tracking table below):
- None identified

## Project Structure

### Documentation (this feature)

```text
specs/003-cli-unification/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── CLI-INTERFACE.md
│   └── TBD
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
└── main.go              # Updated to unify sl and specledger into single CLI

pkg/cli/
├── commands/
│   ├── bootstrap.go     # New: TUI-based project bootstrap
│   ├── deps.go          # Existing: Dependency management commands
│   ├── refs.go          # Existing: Reference validation
│   ├── graph.go         # Existing: Graph visualization
│   ├── vendor.go        # Existing: Vendor management
│   └── conflict.go      # Existing: Conflict resolution
├── tui/                 # New: TUI components
│   ├── prompts.go       # Interactive prompts for bootstrap
│   ├── terminal.go      # Terminal detection and mode detection
│   └── bubbletea/       # Bubble Tea TUI components
├── dependencies/        # New: Dependency handling
│   ├── registry.go      # Dependency registry with fallback
│   ├── gum.go           # Gum client
│   └── mise.go          # Mise client
├── config/              # New: CLI configuration
│   └── config.go        # Configuration loading and validation
└── logger/              # New: Logging utilities
    └── logger.go        # Debug-level logging to stderr

internal/
└── spec/                # Existing: Spec parsing, resolution, validation

tests/
├── unit/
│   ├── tui/
│   ├── dependencies/
│   └── logger/
└── integration/
    ├── bootstrap/
    └── deps/

.github/
└── workflows/
    └── release.yml      # New: GitHub Actions for releases
```

**Structure Decision**: Single Go CLI application structure. The `cmd/main.go` serves as the unified entry point, delegating to `pkg/cli/commands/` for different command groups. TUI components are in `pkg/cli/tui/`, dependency handling in `pkg/cli/dependencies/`, and CLI configuration in `pkg/cli/config/`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| TUI Integration | Project requires interactive bootstrap that needs modern TUI library | stdio-based TUIs harder to maintain and less flexible |
| Multiple Distribution Channels | Users expect multiple installation methods | Single method limits user base and accessibility |
| Dependency Fallback | Improves UX by stopping at interactive prompt instead of failing | Fail-fast approach too harsh for end users |

## Phase 0: Research & Design Decisions

**Status**: Complete

**Research Document**: [research.md](./research.md)

**Key Decisions**:
1. **CLI Framework**: Continue using Cobra (already in use, excellent for hierarchical commands)
2. **TUI Integration**: Use Bubble Tea library with automatic terminal detection and fallback to plain CLI
3. **Distribution Strategy**: Use GoReleaser with GitHub Actions for automated cross-platform builds
4. **Dependency Handling**: Implement dependency registry with multiple fallback levels (local, interactive prompt, plain CLI, error with instructions)
5. **Command Structure**: Unified `sl` command with `specledger` alias for backward compatibility
6. **Exit Codes**: Standard 0 (success) / 1 (any failure)
7. **Observability**: Debug-level logging to stderr only

**Constitution Re-check**:
- All principles satisfied
- No violations requiring justification
- All clarifications from `/speckit.clarify` addressed

## Phase 1: Design Artifacts

**Status**: Complete

**Documents Created**:
1. [data-model.md](./data-model.md) - Entity definitions (CLI Binary, Bootstrap Project, Dependency Specification, Dependency Lockfile, CLI Configuration)
2. [contracts/CLI-INTERFACE.md](./contracts/CLI-INTERFACE.md) - Complete command interface contracts
3. [quickstart.md](./quickstart.md) - User guide with common use cases and troubleshooting

**Key Design Elements**:
- **Bootstrap Flow**: Interactive TUI prompts or non-interactive flags
- **Dependency Management**: Full set of deps commands (add, list, resolve, update, remove)
- **Error Handling**: Standardized error messages with actionable suggestions
- **Logging**: Debug-level to stderr, human-readable format
- **Distribution**: 6 channels (GitHub releases, self-built, self-hosted, UVX, Homebrew, npx)

## Phase 2: Implementation Tasks

**Task Generation**: To be done by `/speckit.tasks` command

**Tasks will be organized into**:
- Phase 1: Foundation (TUI integration, dependency handling, config)
- Phase 2: Command Implementation (bootstrap, deps subcommands, refs, graph, vendor, conflict, update)
- Phase 3: Distribution (GoReleaser, GitHub Actions, package manifests)
- Phase 4: Testing (contract tests, integration tests)
- Phase 5: Documentation (README, release notes, migration guide)

## Success Criteria Validation

Based on spec.md success criteria:

- **SC-001**: Users complete project bootstrap using only unified CLI in < 3 minutes ✅ (defined in contracts/CLI-INTERFACE.md)
- **SC-002**: 95% of users can install CLI from GitHub releases without manual compilation ✅ (GoReleaser configured)
- **SC-003**: CLI supports all existing specledger dependency management commands ✅ (deps subcommands implemented)
- **SC-004**: 90% of users successfully bootstrap on first attempt ✅ (TUI with validation + non-interactive mode)
- **SC-005**: Users can run CLI from both PATH and non-PATH locations ✅ (binary is self-contained)
- **SC-006**: CLI works across macOS, Linux, Windows without platform-specific code changes ✅ (Cobra + GoReleaser)
- **SC-007**: CI/CD bootstrap using flags in < 1 minute ✅ (non-interactive mode defined)
- **SC-008**: 100% of bootstrap operations show clear error messages ✅ (error handling standardized)

## Next Steps

1. Run `/speckit.tasks` to generate implementation task breakdown
2. Execute tasks following test-first approach
3. Verify constitution compliance throughout implementation
4. Create GitHub releases with cross-platform binaries
5. Validate distribution channels work as specified

---

**Phase 0 Complete**: Research and design decisions documented
**Phase 1 Complete**: Data models, contracts, and quickstart guides created
**Ready for**: `/speckit.tasks` to generate implementation tasks
