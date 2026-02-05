# Implementation Plan: Spec Dependency Linking

**Branch**: `002-spec-dependency-linking` | **Date**: 2026-01-30 | **Spec**: [link]
**Input**: Feature specification from `/specs/002-spec-dependency-linking/spec.md`

**Note**: This template is filled in by the `implementation planning` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement golang-style dependency locking and linking for specifications, allowing teams to declare external spec dependencies, resolve them with cryptographic verification, and reference specific sections across repositories. The system will provide dependency resolution with 30-second performance for 10 repositories, reference validation under 5 seconds, and secure handling of private repositories with token/SSH authentication.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.21+
**Primary Dependencies**: go-git/v4, cobra, viper, golang.org/x/crypto
**Storage**: File system (spec.mod, spec.sum, cache directory), Git integration
**Testing**: testify, httptest, go-unit with 90%+ coverage
**Target Platform**: Cross-platform CLI (darwin/amd64, linux/amd64, windows/amd64)
**Project Type**: Single CLI application
**Performance Goals**:
- SC-001: <10s single dependency resolution
- SC-002: <5s reference validation
- SC-003: <30s for 10 repositories
- Memory: <512MB for dependency resolution
**Constraints**:
- Cross-platform compatibility
- Single binary distribution
- Cryptographic verification (SHA-256)
- Secure credential management
- Offline capable with vendoring
**Scale/Scope**:
- 50 transitive dependencies (SC-010)
- 100MB cache limit
- 10k lines/second parsing

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with principles from `.specify/memory/constitution.md`:

- [x] **Specification-First**: Spec.md complete with prioritized user stories (P1-P5)
- [x] **Test-First**: Test strategy defined (unit, integration, performance tests)
- [x] **Code Quality**: Go linting with golangci-lab, formatting with gofmt
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: Metrics defined (SC-001 to SC-010 targets)
- [x] **Observability**: Structured logging with context, metrics for resolution times
- [x] **Issue Tracking**: Beads epic created and linked to spec

**Complexity Violations** (if any, justify in Complexity Tracking table below):
- None identified / [List violations and justifications]

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (implementation planning command output)
├── research.md          # Phase 0 output (implementation planning command)
├── data-model.md        # Phase 1 output (implementation planning command)
├── quickstart.md        # Phase 1 output (implementation planning command)
├── contracts/           # Phase 1 output (implementation planning command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by implementation planning)
```

### Source Code (repository root)

```text
cmd/
└── main.go              # Cobra CLI entry point

internal/
├── spec/
│   ├── parser.go        # spec.mod parsing
│   ├── resolver.go      # Dependency resolution
│   ├── validator.go     # Reference validation
│   ├── manifest.go      # SpecManifest handling
│   └── lockfile.go      # SpecLockfile handling
├── git/
│   ├── client.go        # Git client abstraction
│   ├── auth.go          # Authentication (OAuth2, SSH)
│   ├── repository.go    # Repository operations
│   └── cache.go         # Cache management
├── crypto/
│   ├── hash.go          # SHA-256 calculation
│   ├── verify.go        # Content verification
│   └── signer.go        # Digital signature support
├── config/
│   ├── file.go          # File-based config
│   ├── auth.go          # Token management
│   └── cache.go         # Cache configuration
├── graph/
│   ├── builder.go       # Graph construction
│   ├── resolver.go      # Conflict resolution
│   └── visualizer.go   # Graph visualization
└── vendor/
    ├── copier.go        # Vendoring operations
    └── syncer.go        # Vendor synchronization

pkg/
├── models/
│   ├── dependency.go
│   ├── lockfile.go
│   ├── reference.go
│   ├── graph.go
│   └── vendor.go
├── utils/
│   ├── logger.go
│   ├── progress.go
│   └── errors.go
└── cli/
    ├── commands/
    │   ├── deps.go      # Dependency commands
    │   ├── refs.go      # Reference commands
    │   ├── graph.go     # Graph commands
    │   └── vendor.go    # Vendor commands
    └── flags.go

tests/
├── unit/
│   ├── spec/
│   ├── git/
│   ├── crypto/
│   └── graph/
├── integration/
│   ├── resolution_test.go
│   ├── validation_test.go
│   └── vendoring_test.go
├── performance/
│   ├── resolution_bench.go
│   ├── validation_bench.go
│   └── graph_bench.go
└── fixtures/
    ├── repositories/
    └── specifications/
```

**Structure Decision**: Single Go CLI application with clear separation of concerns. Core logic in `internal/`, public APIs in `pkg/`, and CLI commands in `pkg/cli/`. Test structure mirrors source with unit, integration, and performance tests.

## Integration

**Existing Components**:
- **Beads Issue Tracker**: Already integrated via `bd` commands
- **Claude Skills**: `/specledger.*` commands already functional
- **Mise Configuration**: Will add `sl` tool to `mise.toml`
- **Bootstrap Script**: New functionality will extend existing `sl` script

**CLI Integration Points**:
- Commands will be accessible as `sl deps <command>`, `sl refs <command>`, etc.
- Maintains compatibility with existing `sl bootstrap` functionality
- Will extend the existing `sl` script with dependency management subcommands
- Configuration stored alongside existing tool configs

