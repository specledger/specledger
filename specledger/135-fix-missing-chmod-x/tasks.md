# Tasks Index: Fix Executable Permissions for Template Scripts

Beads Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through Beads CLI.

## Feature Tracking

* **Beads Epic ID**: `SL-slwk`
* **User Stories Source**: `specledger/135-fix-missing-chmod-x/spec.md`
* **Planning Details**: `specledger/135-fix-missing-chmod-x/plan.md`
* **Data Model**: N/A (bug fix, no data model)
* **Contract Definitions**: N/A (CLI tool, no API contracts)

## Beads Query Hints

Use the `bd` CLI to query and manipulate the issue graph:

```bash
# Find all open tasks for this feature
bd list --label spec:135-fix-missing-chmod-x --status open --limit 10

# Find ready tasks to implement
bd ready --label spec:135-fix-missing-chmod-x --limit 5

# See full dependency tree
bd dep tree SL-slwk

# View tasks by user story
bd list --label spec:135-fix-missing-chmod-x --label story:US1
bd list --label spec:135-fix-missing-chmod-x --label story:US2

# View by phase
bd list --label spec:135-fix-missing-chmod-x --label phase:us1
bd list --label spec:135-fix-missing-chmod-x --label phase:polish
```

## Tasks and Phases Structure

This feature follows Beads' 2-level graph structure:

* **Epic**: SL-slwk → Fix Executable Permissions for Template Scripts
* **Phases**: Beads issues of type `feature`, child of the epic
  * US1 (P1): Core bug fix - Scripts work immediately after bootstrap
  * US2 (P2): Existing projects can be fixed
  * Polish (P3): Documentation and cleanup
* **Tasks**: Issues of type `task`, children of each feature issue

## Dependency Graph

```
SL-slwk: Fix Executable Permissions for Template Scripts [P1]
├── SL-slwk.1: US1: Scripts Work Immediately After Bootstrap [P1]
│   ├── SL-slwk.1.1: Create isExecutableFile helper function [P1]
│   ├── SL-slwk.1.2: Fix copyEmbeddedFile [P1] ← depends on SL-slwk.1.1
│   ├── SL-slwk.1.3: Fix applyEmbeddedSkills [P1] ← depends on SL-slwk.1.1
│   └── SL-slwk.1.4: Add unit tests [P2] ← depends on SL-slwk.1.1, SL-slwk.1.2
├── SL-slwk.2: US2: Existing Projects Can Be Fixed [P2]
│   └── SL-slwk.2.1: Integration test [P2] ← depends on SL-slwk.1.2, SL-slwk.1.3
└── SL-slwk.3: Polish: Documentation and Cleanup [P3]
    └── SL-slwk.3.1: Update CLAUDE.md [P3] ← depends on SL-slwk.1.2, SL-slwk.1.3
```

## Phase Summary

### Phase 1: US1 - Scripts Work Immediately After Bootstrap (P1) 🎯 MVP

**Goal**: Fix the core bug so that `sl init` and `sl new` produce executable scripts.

**Independent Test**: Run `sl init my-project` and immediately execute `.specledger/scripts/bash/create-new-feature.sh --help` without permission errors.

**Tasks**:
- `SL-slwk.1.1`: Create isExecutableFile helper function
- `SL-slwk.1.2`: Fix copyEmbeddedFile to set executable permissions
- `SL-slwk.1.3`: Fix applyEmbeddedSkills to set executable permissions
- `SL-slwk.1.4`: Add unit tests for executable detection

**Parallel Opportunities**: Tasks 1.2 and 1.3 can run in parallel after 1.1 is complete.

---

### Phase 2: US2 - Existing Projects Can Be Fixed (P2)

**Goal**: Ensure `sl init --force` fixes permissions on existing projects.

**Independent Test**: Run `sl init --force` on a project with non-executable scripts and verify they become executable.

**Tasks**:
- `SL-slwk.2.1`: Integration test for sl init --force

---

### Phase 3: Polish - Documentation and Cleanup (P3)

**Goal**: Update documentation with fix details.

**Tasks**:
- `SL-slwk.3.1`: Update CLAUDE.md with fix details

---

## MVP Scope

**Recommended MVP**: Complete US1 (Phase 1) only. This delivers the core bug fix:
- Scripts are executable immediately after `sl init` or `sl new`
- No manual `chmod +x` required

US2 and Polish can follow as enhancements.

## Implementation Strategy

1. **Start with SL-slwk.1.1** - Create the helper function (no dependencies)
2. **Then parallel**: SL-slwk.1.2 and SL-slwk.1.3 - Fix both copy functions
3. **Then**: SL-slwk.1.4 - Add tests
4. **Finally**: SL-slwk.2.1 and SL-slwk.3.1 - Integration tests and docs

## Label Conventions

| Label | Purpose |
|-------|---------|
| `spec:135-fix-missing-chmod-x` | All tasks in this feature |
| `phase:us1`, `phase:us2`, `phase:polish` | Phase identification |
| `story:US1`, `story:US2` | User story traceability |
| `component:cli`, `component:docs` | Component mapping |
| `requirement:FR-001` etc. | Functional requirement traceability |
| `test:unit`, `test:integration` | Test type identification |

---

> This file is intentionally light and index-only. Implementation data lives in Beads.
