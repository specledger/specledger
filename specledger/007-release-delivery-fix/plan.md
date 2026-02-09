# Implementation Plan: Release Delivery Fix

**Branch**: `007-release-delivery-fix` | **Date**: 2025-02-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specledger/007-release-delivery-fix/spec.md`

## Summary

Fix the release delivery system to ensure macOS users can install SpecLedger via binary download, shell script, Homebrew, and `go install`. The feature focuses on macOS (darwin) for amd64 and arm64 architectures as the primary development target, with Linux and Windows support deferred to future iterations.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: GoReleaser v2, GitHub Actions, Homebrew
**Storage**: N/A (release artifacts stored in GitHub Releases)
**Testing**: Go built-in testing, manual installation testing
**Target Platform**: macOS (darwin) - amd64 (Intel) and arm64 (Apple Silicon)
**Project Type**: Single CLI project (Go)
**Performance Goals**: Release completes in <5 minutes, install script completes in <1 minute
**Constraints**:
  - Public repository (no Codecov token needed)
  - CGO_ENABLED=0 for portability
  - GoReleaser v2 syntax compliance
**Scale/Scope**:
  - 2 platform builds (darwin_amd64, darwin_arm64)
  - 4 installation methods (binary download, script, Homebrew, go install)
  - Single binary (`sl`)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with principles from `.specledger/memory/constitution.md`:

- [x] **Specification-First**: Spec.md complete with 5 prioritized user stories (P1: Binary Download, Shell Script, Homebrew, Release Automation; P2: Go Install)
- [x] **Test-First**: Test strategy defined - manual testing of installation methods, verification of release artifacts
- [x] **Code Quality**: golangci-lint configured, GoReleaser v2 for release automation
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios for all installation methods
- [x] **Performance**: Metrics defined - release <5 minutes, install <1 minute, checksums 100% verified
- [x] **Observability**: GitHub Actions provides build logs, install script provides verbose output
- [x] **Issue Tracking**: Beads tasks from 006 epic referenced (SL-m1x, SL-4tk for dry-run verification)

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/007-release-delivery-fix/
├── spec.md              # Feature specification
├── plan.md              # This file (/specledger.plan command output)
├── research.md          # Phase 0 output - findings from investigation
├── data-model.md        # Phase 1 output - N/A (no data model for this feature)
├── quickstart.md        # Phase 1 output - installation testing guide
├── contracts/           # Phase 1 output - N/A (no API contracts for this feature)
└── tasks.md             # Phase 2 output (/specledger.tasks command)
```

### Source Code (repository root)

```text
# Single Go project structure
cmd/
├── main.go              # CLI entry point

pkg/
├── ...                  # Internal packages

scripts/
├── install.sh           # Installation script (to be fixed)

.github/
├── workflows/
│   ├── ci.yml           # CI workflow (updated with Codecov)
│   └── release.yml      # Release workflow (GoReleaser automation)

.goreleaser.yaml         # GoReleaser v2 configuration (to be fixed)
codecov.yml              # Codecov configuration (added)
Makefile                 # Build targets
README.md                # Installation instructions (to be updated)
```

**Structure Decision**: Single Go project for CLI tool. GoReleaser handles cross-compilation for darwin_amd64 and darwin_arm64 from single codebase.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |

---

## Phase 0: Outline & Research

### Previous Work Analysis

From Beads issue tracker (spec 006-opensource-readiness):

**Completed Tasks**:
- SL-hlt: Setup Codecov integration (closed)
- SL-rfk: Verify GoReleaser configuration (closed - but had deprecation warnings)

**Open Tasks**:
- SL-m1x: Dry-run release verification (open)
- SL-4tk: Test release process dry-run (open)

**Related Issues** (from recent work on branch 007-build-delivery-setup):
- Fixed GoReleaser v2 deprecation warnings (`archives.format`, `brews` → `homebrew_casks`)
- Fixed broken `windows_arm_7` target by splitting builds
- Fixed archive naming from `sl_*` to `specledger_*`
- Added Codecov configuration

### Current State Investigation

**Issues Identified**:

1. **GoReleaser Configuration**: Recently fixed for v2 syntax but needs verification for macOS-only builds
2. **Install Script**: Binary extraction path assumes old directory structure - needs to match archive naming
3. **Homebrew Tap Repository**: `specledger/homebrew-specledger` may not exist yet (`skip_upload: auto`)
4. **Version Variable**: Binary needs `--version` flag to display version information

### Unknowns to Research

1. **Homebrew Tap Setup**: Does `specledger/homebrew-specledger` repository exist?
2. **Install Script Binary Path**: Current script looks for `$extract_dir/sl` but archive extracts to nested directory
3. **Architecture Detection**: Script defaults to `amd64` - needs to detect arm64 on Apple Silicon

### Research Tasks

Spawn research agents for:

1. **Homebrew Tap Repository**: Verify if `github.com/specledger/homebrew-specledger` exists
2. **Archive Structure**: Confirm GoReleaser archive extraction format
3. **Architecture Detection**: Best practices for detecting Intel vs Apple Silicon on macOS

---

## Phase 1: Design & Contracts

### Data Model

*No data model required for this feature* - release delivery is a DevOps/infrastructure concern with no persistent data entities.

### API Contracts

*No API contracts required* - this feature involves:
- Installation scripts (bash)
- GoReleaser configuration (YAML)
- GitHub Actions workflows (YAML)

### Quickstart Guide

Will be generated in `quickstart.md` covering:
1. How to test the install script locally
2. How to run a GoReleaser dry-run
3. How to verify Homebrew formula
4. How to test `go install` method

---

## Design Decisions

### 1. Platform Scope

**Decision**: Focus on macOS only (darwin_amd64, darwin_arm64)

**Rationale**: User requested "for now focus first on mac development" - reduces testing scope and ensures primary platform works correctly before expanding to Linux/Windows.

### 2. Archive Naming

**Decision**: Use `specledger_VERSION_OS_ARCH.tar.gz` format

**Rationale**: Matches install script expectations, clearly identifies binary contents.

### 3. Install Script Improvements

**Decision**:
- Detect arm64 automatically on Apple Silicon
- Fix binary extraction path
- Add checksum verification

**Rationale**: Current script has hardcoded amd64 and incorrect extraction path.

### 4. Homebrew Tap

**Decision**: Create `github.com/specledger/homebrew-specledger` repository

**Rationale**: Required for `brew install specledger` to work. Use `skip_upload: true` initially, remove when repository exists.

---

## Implementation Phases

### Phase 1.1: Fix GoReleaser Configuration

1. Simplify builds to macOS only (darwin_amd64, darwin_arm64)
2. Remove Windows and Linux builds temporarily
3. Ensure archive naming matches `specledger_VERSION_OS_ARCH.tar.gz`
4. Set `skip_upload: true` for brews until tap repository exists

### Phase 1.2: Fix Install Script

1. Add architecture detection for arm64 (Apple Silicon)
2. Fix binary extraction path
3. Add checksum verification download and check
4. Test on both Intel and Apple Silicon Macs

### Phase 1.3: Add Version Flag

1. Add `--version` flag to CLI
2. Embed version info via ldflags in GoReleaser
3. Test `sl --version` displays correct output

### Phase 1.4: Create Homebrew Tap Repository

1. Create `github.com/specledger/homebrew-specledger` repository
2. Initialize with README
3. Remove `skip_upload: true` from GoReleaser config
4. Test `brew tap` and `brew install`

### Phase 1.5: Update README

1. Update installation instructions
2. Add architecture-specific notes
3. Add troubleshooting section

---

## Testing Strategy

### Manual Testing Checklist

- [ ] Install script works on Intel Mac (amd64)
- [ ] Install script works on Apple Silicon Mac (arm64)
- [ ] Install script verifies checksums
- [ ] Binary download from GitHub Releases works
- [ ] `go install` method works
- [ ] `sl --version` displays correct version
- [ ] Homebrew install works (after tap creation)
- [ ] Upgrade path works (installing over existing version)

### GoReleaser Dry-Run

```bash
goreleaser release --snapshot --clean
```

Verify:
- [ ] No deprecation warnings
- [ ] Builds complete for darwin_amd64 and darwin_arm64
- [ ] Archive names correct
- [ ] Checksums.txt generated
