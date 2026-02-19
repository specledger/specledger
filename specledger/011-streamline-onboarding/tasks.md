# Tasks Index: Streamlined Onboarding Experience

Beads Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through Beads CLI.

## Feature Tracking

* **Beads Epic ID**: `SL-m0h4`
* **User Stories Source**: `specledger/011-streamline-onboarding/spec.md`
* **Research Inputs**: `specledger/011-streamline-onboarding/research.md`
* **Planning Details**: `specledger/011-streamline-onboarding/plan.md`
* **Data Model**: `specledger/011-streamline-onboarding/data-model.md`
* **Contract Definitions**: `specledger/011-streamline-onboarding/contracts/`

## Beads Query Hints

Use the `bd` CLI to query and manipulate the issue graph:

```bash
# Find all open tasks for this feature
bd list --label spec:011-streamline-onboarding --status open --limit 10

# Find ready tasks to implement (no unresolved blockers)
bd ready --label spec:011-streamline-onboarding --limit 5

# See dependency tree for the epic
bd dep tree --reverse SL-m0h4

# View issues by component
bd list --label 'component:tui' --label 'spec:011-streamline-onboarding' --limit 5
bd list --label 'component:bootstrap' --label 'spec:011-streamline-onboarding' --limit 5
bd list --label 'component:launcher' --label 'spec:011-streamline-onboarding' --limit 5
bd list --label 'component:embedded' --label 'spec:011-streamline-onboarding' --limit 5

# View issues by user story
bd list --label 'story:US1' --label 'spec:011-streamline-onboarding'
bd list --label 'story:US2' --label 'spec:011-streamline-onboarding'
bd list --label 'story:US3' --label 'spec:011-streamline-onboarding'
bd list --label 'story:US4' --label 'spec:011-streamline-onboarding'

# Show all phases
bd list --type feature --label 'spec:011-streamline-onboarding'
```

## Tasks and Phases Structure

This feature follows Beads' 2-level graph structure:

* **Epic**: SL-m0h4 → Streamlined Onboarding Experience
* **Phases**: Beads issues of type `feature`, child of the epic
  * Phase 1: Setup (SL-4603)
  * Phase 2: Foundational (SL-499u)
  * Phase 3: US1 - Unified Setup for New Projects (SL-sz26)
  * Phase 4: US2 - Unified Setup for Existing Repos (SL-0uxo)
  * Phase 5: US3 - Guided First Feature Workflow (SL-3hpe)
  * Phase 6: US4 - Agent Preference Persistence (SL-u3pz)
  * Phase 7: Polish and Cross-Cutting Concerns (SL-0fsy)
* **Tasks**: Issues of type `task`, children of each feature issue (phase)

## Convention Summary

| Type    | Description                  | Labels                                           |
| ------- | ---------------------------- | ------------------------------------------------ |
| epic    | Full feature epic            | `spec:011-streamline-onboarding`                 |
| feature | Implementation phase / story | `phase:[name]`, `story:[US#]`                    |
| task    | Implementation task          | `component:[x]`, `requirement:[fr-id]`           |

## Phase 1: Setup — Shared Infrastructure (`SL-4603`)

**Purpose**: Create new packages and helpers needed by all user stories

| Task ID  | Title                                        | Component   | Priority |
| -------- | -------------------------------------------- | ----------- | -------- |
| SL-mnbf  | T001: Create agent launcher package          | launcher    | P1       |
| SL-th5w  | T002: Add constitution detection/write helpers| bootstrap   | P1       |
| SL-qbs7  | T003: Add Agent Preferences to constitution template | embedded | P1   |

```bash
bd list --label 'phase:setup' --label 'spec:011-streamline-onboarding'
```

---

## Phase 2: Foundational — Bootstrap Wiring (`SL-499u`)

**Purpose**: Wire launcher and constitution helpers into the bootstrap flow (blocks all user stories)

**Blocked by**: Phase 1 (SL-4603)

| Task ID  | Title                                        | Component   | Priority |
| -------- | -------------------------------------------- | ----------- | -------- |
| SL-9gov  | T004: Wire agent launch into bootstrap flow  | bootstrap   | P1       |
| SL-pxbe  | T005: Wire constitution check into sl init   | bootstrap   | P1       |

```bash
bd list --label 'phase:foundational' --label 'spec:011-streamline-onboarding'
```

**Checkpoint**: Foundation ready — user story implementation can begin

---

## Phase 3: US1 — Unified Setup for New Projects (`SL-sz26`) 🎯 MVP

**Goal**: `sl new` presents constitution + agent preference steps, writes populated constitution, launches agent

**Independent Test**: Run `sl new`, verify TUI shows all 7 steps, creates populated constitution, launches agent

**Blocked by**: Phase 2 (SL-499u)

| Task ID  | Title                                        | Component   | Priority |
| -------- | -------------------------------------------- | ----------- | -------- |
| SL-r40c  | T006: Add constitution principles TUI step   | tui         | P1       |
| SL-i1y4  | T007: Add agent preference TUI step          | tui         | P1       |
| SL-kteh  | T008: Write constitution and launch agent    | bootstrap   | P1       |

```bash
bd list --label 'story:US1' --label 'spec:011-streamline-onboarding'
```

**Checkpoint**: `sl new` end-to-end flow working with constitution + agent launch

---

## Phase 4: US2 — Unified Setup for Existing Repos (`SL-0uxo`)

**Goal**: `sl init` presents interactive TUI for missing config, detects existing constitution, launches agent

**Independent Test**: Run `sl init` in existing repo, verify TUI shows only missing config, agent launches

**Blocked by**: Phase 2 (SL-499u)

| Task ID  | Title                                        | Component   | Priority |
| -------- | -------------------------------------------- | ----------- | -------- |
| SL-7p8n  | T009: Create interactive TUI for sl init     | tui         | P1       |
| SL-m9fj  | T010: Wire sl init to use TUI and launch agent | bootstrap | P1       |

```bash
bd list --label 'story:US2' --label 'spec:011-streamline-onboarding'
```

**Checkpoint**: `sl init` end-to-end flow working with interactive prompts + agent launch

---

## Phase 5: US3 — Guided First Feature Workflow (`SL-3hpe`)

**Goal**: `/specledger.onboard` command guides user through specify → clarify → plan → tasks → review → implement

**Independent Test**: Launch agent, run `/specledger.onboard`, verify workflow pauses at task review

**Blocked by**: Phase 3 (SL-sz26) and Phase 4 (SL-0uxo)

| Task ID  | Title                                        | Component   | Priority |
| -------- | -------------------------------------------- | ----------- | -------- |
| SL-4ras  | T011: Create /specledger.onboard command     | embedded    | P2       |
| SL-j4er  | T012: Register onboard in help command       | embedded    | P2       |

```bash
bd list --label 'story:US3' --label 'spec:011-streamline-onboarding'
```

**Checkpoint**: Guided onboarding workflow fully functional

---

## Phase 6: US4 — Agent Preference Persistence (`SL-u3pz`)

**Goal**: Returning users see their previously selected agent as the default in TUI

**Independent Test**: Set agent preference, run `sl init` again, verify default is preserved

**Blocked by**: Phase 4 (SL-0uxo)

| Task ID  | Title                                        | Component     | Priority |
| -------- | -------------------------------------------- | ------------- | -------- |
| SL-25b9  | T013: Use existing agent preference as default | bootstrap/tui | P3     |

```bash
bd list --label 'story:US4' --label 'spec:011-streamline-onboarding'
```

---

## Phase 7: Polish & Cross-Cutting Concerns (`SL-0fsy`)

**Purpose**: Integration tests, edge case handling, and final validation

**Blocked by**: All user story phases (SL-sz26, SL-0uxo, SL-3hpe, SL-u3pz)

| Task ID  | Title                                        | Component   | Priority |
| -------- | -------------------------------------------- | ----------- | -------- |
| SL-8x0d  | T014: Add integration tests for onboarding flow | testing   | P2       |
| SL-dady  | T015: Handle edge cases for agent launch/constitution | bootstrap/tui | P2 |

```bash
bd list --label 'phase:polish' --label 'spec:011-streamline-onboarding'
```

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)          → no deps, start immediately
Phase 2 (Foundational)   → blocked by Phase 1
Phase 3 (US1)            → blocked by Phase 2
Phase 4 (US2)            → blocked by Phase 2
  ↳ Phase 3 and 4 can run in PARALLEL after Phase 2
Phase 5 (US3)            → blocked by Phase 3 + Phase 4
Phase 6 (US4)            → blocked by Phase 4
Phase 7 (Polish)         → blocked by all user story phases
```

### Parallel Execution Opportunities

| Parallel Group | Tasks                              | Condition                |
| -------------- | ---------------------------------- | ------------------------ |
| Group A        | T001, T002, T003 (Setup)           | Start immediately        |
| Group B        | T004, T005 (Foundational)          | After Setup complete     |
| Group C        | T006+T007+T008 ∥ T009+T010        | US1 and US2 in parallel  |
| Group D        | T011+T012 ∥ T013                   | US3 and US4 in parallel  |
| Group E        | T014, T015 (Polish)                | After all stories done   |

### Implementation Strategy

**MVP Scope**: Phase 1 + Phase 2 + Phase 3 (US1: `sl new` with constitution + agent launch)
- Delivers the core value: new project onboarding with constitution creation and AI agent launch
- 8 tasks total (T001-T008)
- Independently testable and deliverable

**Incremental Delivery**:
1. **MVP**: US1 (`sl new` flow) — constitution + agent preference + agent launch
2. **Increment 2**: US2 (`sl init` flow) — interactive prompts for existing repos
3. **Increment 3**: US3 (guided workflow) — `/specledger.onboard` command
4. **Increment 4**: US4 (persistence) — remember agent preference
5. **Final**: Polish — integration tests + edge cases

---

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
bd dep tree --reverse SL-m0h4

# Get all tasks by label
bd list --label spec:011-streamline-onboarding --label story:US1

# Add a new task
bd create "New task title" -t task --parent SL-sz26 --label spec:011-streamline-onboarding --label component:tui

# Update notes on a task
bd update SL-xyz123 --notes "Re-use helper functions from existing module"

# Add a comment to an issue
bd comments add SL-xyz123 "Research finding..."

# Mark task as completed with context
bd close SL-xyz123 --reason "Completed implementation"
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
