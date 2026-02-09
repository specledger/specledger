# Tasks Index: Release Delivery Fix

Beads Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through Beads CLI.

## Feature Tracking

* **Beads Epic ID**: `sl-26y`
* **User Stories Source**: `specledger/007-release-delivery-fix/spec.md`
* **Research Inputs**: `specledger/007-release-delivery-fix/research.md`
* **Planning Details**: `specledger/007-release-delivery-fix/plan.md`
* **Data Model**: `specledger/007-release-delivery-fix/data-model.md` (N/A for this feature)
* **Contract Definitions**: `specledger/007-release-delivery-fix/contracts/` (N/A for this feature)

## Beads Query Hints

Use the `bd` CLI to query and manipulate the issue graph:

```bash
# Find all open tasks for this feature
bd list --label spec:007-release-delivery-fix --status open --limit 20

# Find ready tasks to implement
bd ready --label spec:007-release-delivery-fix --limit 10

# See dependencies for issue
bd dep tree sl-26y

# View issues by component
bd list --label 'component:release' --label 'spec:007-release-delivery-fix' --limit 10

# View issues by phase
bd list --type feature --label 'spec:007-release-delivery-fix'

# View issues by story
bd list --label story:US1 --label 'spec:007-release-delivery-fix'
```

## Tasks and Phases Structure

This feature follows Beads' 2-level graph structure:

* **Epic**: sl-26y → represents the whole feature
* **Phases**: Beads issues of type `feature`, child of the epic
  * Phase = a user story group or technical milestone (setup, foundational, us1, us2, us3, us4, us5, polish)
* **Tasks**: Issues of type `task`, children of each feature issue (phase)

## Convention Summary

| Type    | Description                  | Labels                                 |
| ------- | ---------------------------- | -------------------------------------- |
| epic    | Full feature epic            | `spec:007-release-delivery-fix`        |
| feature | Implementation phase / story | `phase:[name]`, `story:US#`            |
| task    | Implementation task          | `component:[area]`, `fr:FR-###`        |

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Codecov integration for coverage tracking

**Feature**: `sl-0no` - Setup: Codecov Integration

### Tasks

- `sl-7ef` [P0] Add codecov.yml configuration (component:ci, fr:FR-001)
- `sl-0zn` [P0] Update CI workflow for Codecov (component:ci, fr:FR-001, blocks:sl-7ef)

**Checkpoint**: Codecov integration complete - CI reports coverage

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Version flag is required by US3 (Homebrew) for version display

**⚠️ CRITICAL**: User Story 3 (Homebrew) cannot complete without this phase

**Feature**: `sl-3tj` - Foundational: Add Version Flag to CLI

### Tasks

- `sl-y7x` [P0] Add version variable to main.go (component:cli, fr:FR-011)

**Checkpoint**: Version flag implemented - `sl --version` displays version info

---

## Phase 3: User Story 1 - Binary Download from GitHub Releases (Priority: P1) 🎯 MVP

**Goal**: Users can download pre-built macOS binaries (amd64, arm64) from GitHub Releases

**Independent Test**: Download binary from release, extract, run `sl` command

**Feature**: `sl-u1u` - US1: Binary Download from GitHub Releases

### Tasks

- `sl-vhp` [P0] Simplify GoReleaser builds to macOS only (component:goreleaser, fr:FR-001, fr:FR-012, fr:FR-015)
- `sl-a6k` [P1] Verify GoReleaser dry-run (component:goreleaser, fr:FR-002, fr:FR-003, blocks:sl-vhp)

**Checkpoint**: macOS binaries (darwin_amd64, darwin_arm64) build and attach to releases

---

## Phase 4: User Story 2 - Shell Script Installation (Priority: P1)

**Goal**: Install tool with single curl command - detects Intel/Apple Silicon, verifies checksums

**Independent Test**: Run install script, verify `sl` command available

**Feature**: `sl-5t1` - US2: Shell Script Installation

### Tasks

- `sl-m82` [P0] Fix install script architecture detection (component:install, fr:FR-004, fr:FR-005, fr:FR-006, fr:FR-007, fr:FR-009)

**Checkpoint**: Install script works on macOS (Intel and Apple Silicon) with checksum verification

---

## Phase 5: User Story 3 - Homebrew Installation (Priority: P1)

**Goal**: macOS users can install via Homebrew tap

**Independent Test**: Tap repository, run `brew install specledger`, verify `sl` works

**Feature**: `sl-znp` - US3: Homebrew Installation

### Tasks

- `sl-5ye` [P1] Create Homebrew tap repository (component:homebrew, fr:FR-008)
- `sl-o1d` [P1] Enable Homebrew formula uploads (component:homebrew, fr:FR-009, blocks:sl-5ye)

**Checkpoint**: Homebrew tap functional, users can `brew install specledger`

---

## Phase 6: User Story 4 - Go Install (Priority: P2)

**Goal**: Go developers can install using `go install` command

**Independent Test**: Run `go install`, verify binary accessible

**Feature**: `sl-6u3` - US4: Go Install

### Tasks

- `SL-gry` [P2] Verify go install works correctly (component:cli, fr:FR-010, blocks:sl-y7x)
  - Package path: `github.com/specledger/specledger/cmd@latest`
  - Note: Binary is named `sl` by GoReleaser, but package path is `cmd/` not `cmd/sl/`

**Checkpoint**: `go install` works, `sl --version` displays correct version

---

## Phase 7: User Story 5 - Release Automation (Priority: P1)

**Goal**: Pushing git tag automatically builds and publishes all release artifacts

**Independent Test**: Push tag, verify GitHub Actions creates release with all artifacts

**Feature**: `sl-4cj` - US5: Release Automation

### Tasks

- `sl-35x` [P0] Test GitHub Actions release workflow (component:release, fr:FR-003, blocks:sl-vhp, blocks:sl-m82)

**Checkpoint**: Release automation functional - tag push creates complete release

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates and manual testing

**Feature**: `sl-1u1` - Polish: Documentation and Verification

### Tasks

- `sl-ptc` [P2] Update README with installation instructions (component:docs)
- `sl-f4q` [P2] Run manual installation testing (component:testing)

**Checkpoint**: Documentation complete, all installation methods verified

---

## Dependencies & Execution Order

### Phase Dependencies

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              EPIC: sl-26y                                    │
│                              Release Delivery Fix                           │
└─────────────────────────────────────────────────────────────────────────────┘
         │
         ├──┌─────────────────────────────────────────────────────┐
         │ │ Phase 1: Setup (sl-0no)                              │
         │ │   • Add codecov.yml (sl-7ef)                         │
         │ │   • Update CI workflow (sl-0zn) ──> sl-7ef           │
         │ └─────────────────────────────────────────────────────┘
         │
         ├──┌─────────────────────────────────────────────────────┐
         │ │ Phase 2: Foundational (sl-3tj) ⚠️ BLOCKS US3        │
         │ │   • Add version flag (sl-y7x)                        │
         │ └─────────────────────────────────────────────────────┘
         │
         ├──┌─────────────────────────────────────────────────────┐
         │ │ Phase 3: US1 Binary Download (sl-u1u) 🎯 MVP        │
         │ │   • Simplify GoReleaser (sl-vhp)                    │
         │ │   • Verify dry-run (sl-a6k) ──> sl-vhp              │
         │ └─────────────────────────────────────────────────────┘
         │
         ├──┌─────────────────────────────────────────────────────┐
         │ │ Phase 4: US2 Shell Script (sl-5t1)                  │
         │ │   • Fix install script (sl-m82)                     │
         │ └─────────────────────────────────────────────────────┘
         │
         ├──┌─────────────────────────────────────────────────────┐
         │ │ Phase 5: US3 Homebrew (sl-znp)                      │
         │ │   • Create tap repo (sl-5ye)                        │
         │ │   • Enable uploads (sl-o1d) ──> sl-5ye              │
         │ │ ⚠️ Requires: Phase 2 (version flag)                │
         │ └─────────────────────────────────────────────────────┘
         │
         ├──┌─────────────────────────────────────────────────────┐
         │ │ Phase 6: US4 Go Install (sl-6u3)                     │
         │ │   • Verify go install (sl-gry) ──> sl-y7x           │
         │ │ ⚠️ Requires: Phase 2 (version flag)                │
         │ └─────────────────────────────────────────────────────┘
         │
         ├──┌─────────────────────────────────────────────────────┐
         │ │ Phase 7: US5 Release Automation (sl-4cj)            │
         │ │   • Test release workflow (sl-35x)                  │
         │ │ ⚠️ Requires: Phase 1 (GoReleaser), Phase 2 (script)│
         │ └─────────────────────────────────────────────────────┘
         │
         └──┌─────────────────────────────────────────────────────┐
           │ Phase 8: Polish (sl-1u1)                            │
           │   • Update README (sl-ptc)                          │
           │   • Manual testing (sl-f4q)                         │
           └─────────────────────────────────────────────────────┘
```

### Critical Path

The minimal path to a working release system:

1. **Setup Phase** (sl-0no) → Codecov integration
2. **Foundational Phase** (sl-3tj) → Version flag (required for Homebrew)
3. **US1** (sl-u1u) → GoReleaser builds macOS binaries
4. **US2** (sl-5t1) → Install script works
5. **US5** (sl-4cj) → Release automation end-to-end

### Parallel Opportunities

- **US3 (Homebrew)** and **US4 (Go Install)** can proceed in parallel after Foundational phase
- **US1, US2, US5** share GoReleaser configuration changes - should be sequential
- **Polish phase** can overlap with later user stories

---

## Implementation Strategy

### MVP Scope (Minimum Viable Product)

**MVP = User Story 1 (Binary Download)**

With just US1 complete:
- Users can download binaries from GitHub Releases
- GoReleaser automation works for macOS
- Foundation for other installation methods

### Incremental Delivery

1. **First Release** (US1 only): Binary download + GoReleaser automation
2. **Second Release** (+US2): Add one-line install script
3. **Third Release** (+US3): Add Homebrew support
4. **Fourth Release** (+US5): Full end-to-end release automation validation
5. **Future Release** (+US4): Go install method (lower priority)

### Story Testability

Each user story is independently testable:

- **US1**: Download binary, extract, run → verifiable without other stories
- **US2**: Run install script → verifiable without other stories
- **US3**: `brew tap` and `brew install` → verifiable without other stories
- **US4**: `go install` command → verifiable without other stories
- **US5**: Push git tag → verifiable end-to-end automation

---

## Quick Queries for Execution

```bash
# Show all tasks for this feature
bd list --label spec:007-release-delivery-fix

# Show ready tasks (no blocking dependencies)
bd ready --label spec:007-release-delivery-fix

# Show tasks by priority
bd list --label spec:007-release-delivery-fix --sort priority

# Show dependency tree for epic
bd dep tree sl-26y

# Show blocked tasks
bd blocked --label spec:007-release-delivery-fix

# Show feature phases
bd list --type feature --label spec:007-release-delivery-fix
```

---

> **Note**: This file is intentionally light and index-only. Implementation data lives in Beads. Update this file only to point humans and agents to canonical query paths and feature references.
