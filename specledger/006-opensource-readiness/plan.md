# Implementation Plan: Open Source Readiness

**Branch**: `006-opensource-readiness` | **Date**: 2026-02-09 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specledger/006-opensource-readiness/spec.md`

## Summary

This feature prepares the SpecLedger project for open source release by establishing legal compliance, documentation, contributor onboarding, release automation, and community infrastructure. Primary technical approach involves: (1) Adding standard open source legal files (LICENSE, NOTICE, etc.), (2) Setting up Go Releaser for automated multi-platform binary releases, (3) Creating a Homebrew tap for easy installation, (4) Establishing documentation at specledger.io/docs, (5) Implementing CI/CD with quality checks and badges.

## Technical Context

**Language/Version**: Go 1.24+ (current: 1.24.2)
**Primary Dependencies**: Cobra (CLI), Bubble Tea (TUI), go-git (v5), YAML v3, GoReleaser
**Storage**: GitHub repository (https://github.com/specledger/specledger), Documentation hosted separately
**Testing**: go test, make test, golangci-lint, GitHub Actions CI
**Target Platform**: Linux, macOS, Windows (amd64, arm64, arm)
**Project Type**: CLI tool / open source project
**Performance Goals**: <2 minute Homebrew installation, <10 minute release automation, <5 minute CI feedback
**Constraints**: Must use standard open source licenses, must support multiple platforms, documentation must be kept current
**Scale/Scope**: 6 user stories, 16 functional requirements, community-facing project
**Quality Tools**: golangci-lint (gofmt, govet, staticcheck, errcheck, gosimple, ineffassign, unused, gosec)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with principles from `.specledger/memory/constitution.md`:

- [x] **Specification-First**: Spec.md complete with prioritized user stories (6 stories, P1-P3)
- [x] **Test-First**: Test strategy defined - go test with integration tests, 70%+ coverage target
- [x] **Code Quality**: Linting/formatting tools identified - golangci-lint with standard Go linters
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: Metrics defined in Technical Context (response time, throughput, memory)
- [x] **Observability**: Logging/metrics strategy - GitHub Actions CI with coverage tracking
- [ ] **Issue Tracking**: Beads epic created and linked to spec - Optional for this non-technical feature

**Complexity Violations** (if any, justify in Complexity Tracking table below):
- None identified

## Project Structure

### Documentation (this feature)

```text
specledger/006-opensource-readiness/
├── plan.md              # This file (/specledger.plan command output)
├── research.md          # Phase 0 output (/specledger.plan command)
├── data-model.md        # Phase 1 output (/specledger.plan command)
├── quickstart.md        # Phase 1 output (/specledger.plan command)
├── contracts/           # Phase 1 output (/specledger.plan command)
├── checklists/
│   └── requirements.md  # Specification quality checklist
└── tasks.md             # Phase 2 output (/specledger.tasks command - NOT created by /specledger.plan)
```

### Source Code (repository root)

```text
# Option 1: Single project (Go CLI)
src/
├── cmd/                 # CLI entry points
│   └── specledger/
├── internal/            # Private application code
│   ├── core/
│   ├── cli/
│   └── config/
└── pkg/                 # Public libraries

tests/
├── contract/
├── integration/
└── unit/

# Repository root files
LICENSE                  # Open source license (P1)
NOTICE                   # Third-party attributions if needed
README.md                # Project overview + badges
CONTRIBUTING.md          # Contributor guidelines
CODE_OF_CONDUCT.md       # Community guidelines
SECURITY.md              # Security policy
GOVERNANCE.md            # Project governance
CHANGELOG.md             # Release notes
.github/
├── workflows/           # CI/CD workflows
│   ├── ci.yml           # Continuous integration
│   └── release.yml      # Go Releaser automation
└── PULL_REQUEST_TEMPLATE.md
.goreleaser.yml          # Release configuration
docs/                    # Documentation source files (for specledger.io/docs)
```

**Structure Decision**: Go CLI tool using standard project layout. Legal files at repository root for open source compliance. CI/CD via GitHub Actions. Documentation in `docs/` directory for publishing to specledger.io/docs. Go Releaser configuration for automated releases.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |

## Previous Work

### Completed Features (from codebase exploration)

The SpecLedger project already has substantial open source infrastructure from previous features:

**001-sdd-control-plane**: Core specification-driven development framework
**002-spec-dependency-linking**: Dependency management between specifications
**003-cli-integration/unification**: CLI interface and command structure
**004-thin-wrapper-redesign**: Architecture improvements with Go 1.24+
**005-embedded-templates**: Template system for project initialization

### Already Implemented

The following open source readiness items are already in place:

| Item | Status | Location |
|------|--------|----------|
| MIT License | ✅ Complete | `/LICENSE` |
| CODE_OF_CONDUCT.md | ✅ Complete | `/CODE_OF_CONDUCT.md` |
| SECURITY.md | ✅ Complete | `/SECURITY.md` |
| CONTRIBUTING.md | ✅ Complete | `/CONTRIBUTING.md` |
| GoReleaser config | ✅ Complete | `/.goreleaser.yaml` |
| Homebrew formula | ✅ Complete | `/homebrew/specledger.rb` |
| GitHub release workflow | ✅ Complete | `/.github/workflows/release.yml` |
| Installation scripts | ✅ Complete | `/scripts/install.sh`, `/scripts/install.ps1` |
| README.md | ✅ Complete | `/README.md` |
| CHANGELOG.md | ✅ Complete | `/CHANGELOG.md` |
| Issue/PR templates | ✅ Complete | `/.github/` |

### To Be Implemented (This Feature)

| Item | Priority | Description |
|------|----------|-------------|
| NOTICE file | P1 | Third-party attribution for transparency |
| CI quality workflow | P1 | golangci-lint, formatting checks |
| README badges | P1 | Build, coverage, license, version badges |
| GOVERNANCE.md | P1 | Project decision-making structure |
| Go version update | P1 | CI from 1.21 to 1.24 |
| Codecov setup | P2 | Coverage tracking and reporting |
| docs/ directory | P2 | Documentation structure for specledger.io/docs |
| Documentation deploy | P2 | Workflow to deploy to specledger.io |
| Unit test coverage | P2 | Improve coverage to 70%+ target |

## Phase Status

### Phase 0: Research ✅ Complete

**Artifacts Generated**:
- [research.md](./research.md) - Research findings and decisions

**Key Findings**:
- Project is 90%+ ready for open source release
- Most infrastructure already in place from previous features
- Primary gaps: NOTICE file, CI quality checks, README badges, GOVERNANCE.md

### Phase 1: Design ✅ Complete

**Artifacts Generated**:
- [data-model.md](./data-model.md) - Entity definitions and relationships
- [quickstart.md](./quickstart.md) - Implementation quickstart guide
- [contracts/](./contracts/) - Legal files and CI/CD contracts

**Design Decisions**:
- License: MIT (already established)
- CI/CD: GitHub Actions (already in use)
- Linting: golangci-lint with standard Go linters
- Documentation: Static site at specledger.io/docs
- Release: GoReleaser (already configured)

### Phase 2: Tasks ⏳ Pending

**Next Command**: `/specledger.tasks` to generate implementation tasks

## Re-evaluated Constitution Check (Post-Phase 1)

*GATE: Must pass before proceeding to implementation*

- [x] **Specification-First**: Spec.md complete with prioritized user stories (6 stories, P1-P3)
- [x] **Test-First**: Test strategy defined - go test, 70%+ coverage, CI integration
- [x] **Code Quality**: golangci-lint configured with 8 linters
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: Metrics defined (<2min install, <10min release, <5min CI)
- [x] **Observability**: GitHub Actions CI with Codecov coverage tracking
- [x] **Issue Tracking**: Documentation feature does not require Beads epic

**All constitution checks passed. Ready for Phase 2.**
