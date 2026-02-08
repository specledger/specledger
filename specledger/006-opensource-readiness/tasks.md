# Tasks Index: Open Source Readiness

Beads Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through Beads CLI.

## Feature Tracking

* **Beads Epic ID**: `SL-9c7`
* **User Stories Source**: `specledger/006-opensource-readiness/spec.md`
* **Research Inputs**: `specledger/006-opensource-readiness/research.md`
* **Planning Details**: `specledger/006-opensource-readiness/plan.md`
* **Data Model**: `specledger/006-opensource-readiness/data-model.md`
* **Contract Definitions**: `specledger/006-opensource-readiness/contracts/`

## Beads Query Hints

Use the `bd` CLI to query and manipulate the issue graph:

```bash
# Find all open tasks for this feature
bd list --label spec:006-opensource-readiness --status open

# Find ready tasks to implement
bd ready --label spec:006-opensource-readiness --limit 5

# See dependencies for issue
bd dep tree SL-9c7

# View issues by component
bd list --label 'component:legal' --label 'spec:006-opensource-readiness'
bd list --label 'component:ci' --label 'spec:006-opensource-readiness'
bd list --label 'component:docs' --label 'spec:006-opensource-readiness'
bd list --label 'component:release' --label 'spec:006-opensource-readiness'

# View issues by phase
bd list --label 'phase:setup' --label 'spec:006-opensource-readiness'
bd list --label 'phase:US1' --label 'spec:006-opensource-readiness'
bd list --label 'phase:US4' --label 'spec:006-opensource-readiness'

# Show all phases
bd list --type feature --label 'spec:006-opensource-readiness'
```

## Tasks and Phases Structure

This feature follows Beads' 2-level graph structure:

* **Epic**: SL-9c7 → represents the whole feature (Open Source Readiness)
* **Phases**: Beads issues of type `feature`, child of the epic
  * Phase = a user story group or technical milestone (setup, US1-US6)
* **Tasks**: Issues of type `task`, children of each feature issue (phase)

## Convention Summary

| Type    | Description                  | Labels                                 |
| ------- | ---------------------------- | -------------------------------------- |
| epic    | Full feature epic            | `spec:006-opensource-readiness`        |
| feature | Implementation phase / story | `phase:[name]`, `story:US[1-6]`        |
| task    | Implementation task          | `component:[area]`, `fr:FR-[###]`      |

## Agent Execution Flow

MCP agents and AI workflows should:

1. **Assume `bd init` already done** by `specify init`
2. **Use `bd create`** to directly generate Beads issues
3. **Set metadata and dependencies** in the graph, not markdown
4. **Use this markdown only as a navigational anchor**

> Agents MUST NOT output tasks into this file. They MUST use Beads CLI to record all task and phase structure.

## Example Queries for Agents

```bash
# Get all tasks in tree structure for the feature
bd dep tree --reverse SL-9c7

# Get all tasks by user story
bd list --label spec:006-opensource-readiness --label story:US1
bd list --label spec:006-opensource-readiness --label story:US4

# Get ready tasks (no unresolved dependencies)
bd ready --label spec:006-opensource-readiness

# Add a new task to a phase
bd create "New task title" -t task --parent SL-[phase-id] --label spec:006-opensource-readiness --label component:[area]

# Update task status
bd update SL-[task-id] --status closed --notes "Completed with details"
```

## Status Tracking

Status is tracked only in Beads:

* **Open** → default
* **In Progress** → task being worked on
* **Blocked** → dependency unresolved
* **Closed** → complete

Use `bd ready`, `bd blocked`, `bd stats` with appropriate filters to query progress.

---

> This file is intentionally light and index-only. Implementation data lives in Beads. Update this file only to point humans and agents to canonical query paths and feature references.


## Phase 1: Setup and Infrastructure (SL-2ww)

**Purpose**: Shared infrastructure setup for all open source readiness work

**Beads Feature**: `SL-2ww`

**Tasks**:
- `SL-fup` Add golangci-lint configuration [P0]
- `SL-1gj` Create CI quality workflow [P0]
- `SL-j89` Add README badges [P1]

---

## Phase 2: US1 - License Compliance (SL-1pv) 🎯 MVP

**Purpose**: Open Source License Compliance (P1)

**Goal**: Ensure project has proper open source licensing and legal documentation

**Independent Test**: Review presence and completeness of required legal files (LICENSE, NOTICE, etc.) and verify they contain appropriate content

**Beads Feature**: `SL-1pv`

**Tasks**:
- `SL-6o1` Create NOTICE file [P0] - FR-004
- `SL-7x9` Create GOVERNANCE.md [P0] - FR-010
- `SL-eke` Verify existing legal files [P1] - FR-001, FR-003, FR-006

**Checkpoint**: All legal files present and verified - project is legally compliant for open source release

---

## Phase 3: US4 - Release and Distribution (SL-2nv)

**Purpose**: Release and Distribution (P1)

**Goal**: Easy installation and updates using standard package managers

**Independent Test**: Install the project using Homebrew and verify installation works correctly

**Beads Feature**: `SL-2nv`

**Dependencies**: Depends on US1 completion

**Tasks**:
- `SL-rfk` Verify GoReleaser configuration [P1] - FR-013
- `SL-m1x` Dry-run release verification [P2] - FR-013

**Checkpoint**: Release automation verified and ready for production use

---

## Phase 4: US5 - CI and Quality (SL-lcn)

**Purpose**: Continuous Integration and Quality (P2)

**Goal**: Automated testing and quality checks for code quality and contributor feedback

**Independent Test**: Submit changes and verify automated checks run and provide appropriate feedback

**Beads Feature**: `SL-lcn`

**Dependencies**: Depends on Setup phase completion

**Tasks**:
- `SL-hlt` Setup Codecov integration [P1] - FR-005
- `SL-vl9` Verify local quality tools [P2] - FR-005

**Checkpoint**: CI/CD pipeline fully functional with quality checks and coverage tracking

---

## Phase 5: US6 - Documentation and Branding (SL-604)

**Purpose**: Documentation and Branding (P2)

**Goal**: Comprehensive and up-to-date documentation at memorable domain

**Independent Test**: Navigate to main website and documentation site and verify all links work and content is current

**Beads Feature**: `SL-604`

**Tasks**:
- `SL-48g` Create docs directory structure [P1] - FR-011
- `SL-48v` Create documentation deployment workflow [P2] - FR-016

**Checkpoint**: Documentation structure ready for deployment to specledger.io/docs

---

## Phase 6: US2 - Contributor Onboarding (SL-8gh)

**Purpose**: Contributor Onboarding (P2)

**Goal**: Clear documentation for setting up, building, and contributing

**Independent Test**: Follow the documented setup and contribution instructions from scratch on a clean system

**Beads Feature**: `SL-8gh`

**Dependencies**: Depends on US6 (Documentation structure)

**Tasks**:
- `SL-odm` Populate contributor documentation [P1] - FR-005, FR-016
- `SL-2sa` Verify README getting started section [P2] - FR-002

**Checkpoint**: New contributors can onboard independently using documentation

---

## Phase 7: US3 - Project Governance (SL-7r2)

**Purpose**: Project Governance and Maintenance (P3)

**Goal**: Understanding of how project is governed and maintained

**Independent Test**: Review governance documentation and verify it outlines decision-making processes and maintenance policies

**Beads Feature**: `SL-7r2`

**Dependencies**: Depends on US1 (GOVERNANCE.md) and US6 (Documentation structure)

**Tasks**:
- `SL-9mp` Populate governance documentation [P2] - FR-010
- `SL-a5j` Verify SECURITY.md completeness [P2] - FR-006

**Checkpoint**: Governance documentation complete and accessible to community

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **US1 - License Compliance (Phase 2)**: No dependencies on other phases 🎯 MVP
- **US4 - Release (Phase 3)**: Depends on US1 completion
- **US5 - CI/Quality (Phase 4)**: Depends on Setup completion
- **US6 - Documentation (Phase 5)**: No dependencies on other phases (can run in parallel with US2, US3)
- **US2 - Contributor Onboarding (Phase 6)**: Depends on US6 (Documentation structure)
- **US3 - Governance (Phase 7)**: Depends on US1 (GOVERNANCE.md) and US6 (Documentation structure)

### MVP Scope

**Suggested MVP**: Phase 1 (Setup) + Phase 2 (US1 - License Compliance)

This delivers:
- Quality infrastructure (golangci-lint, CI workflow)
- Legal compliance (NOTICE, GOVERNANCE.md)
- README badges
- All P1 items for open source readiness

### Parallel Execution Opportunities

- **Phase 1 + Phase 2 (US1)**: Can run in parallel (Setup and License Compliance are independent)
- **Phase 5 (US6) + Phase 4 (US5)**: Documentation and CI/Quality can proceed in parallel after Setup
- **After US6**: US2 and US3 can both work with documentation structure in parallel

### Story Testability

All user stories are independently testable:
- **US1**: Verify legal files exist and contain correct content
- **US4**: Run Homebrew installation and verify binary works
- **US5**: Submit PR and verify CI checks run
- **US6**: Navigate to specledger.io/docs and verify content
- **US2**: Follow setup docs on clean system
- **US3**: Review governance documentation

## Implementation Strategy

### Incremental Delivery

1. **MVP (P0-P1 items)**: Setup + US1 - Foundation for open source release
2. **Release Readiness (P1)**: Add US4 - Verify distribution works
3. **Quality Infrastructure (P1-P2)**: Add US5 - Complete CI/CD pipeline
4. **Documentation Complete (P2)**: Add US6 + US2 - Full contributor onboarding
5. **Governance Complete (P2-P3)**: Add US3 - Community governance transparency

### Total Tasks

- **Epic**: 1 (SL-9c7)
- **Features (Phases)**: 7
- **Tasks**: 18
- **P0 Tasks**: 6 (critical path)
- **P1 Tasks**: 8 (high priority)
- **P2 Tasks**: 4 (normal priority)

## Quick Reference

```bash
# View all tasks for this feature
bd list --label spec:006-opensource-readiness

# View ready to implement (no blocking dependencies)
bd ready --label spec:006-opensource-readiness

# View by priority
bd list --label spec:006-opensource-readiness --priority 0  # Critical
bd list --label spec:006-opensource-readiness --priority 1  # High
bd list --label spec:006-opensource-readiness --priority 2  # Normal

# View MVP scope (P0-P1)
bd list --label spec:006-opensource-readiness --max-priority 1

# Dependency tree for epic
bd dep tree SL-9c7
```
