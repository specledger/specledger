# Tasks Index: Built-In Issue Tracker

Beads Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through Beads CLI.

## Feature Tracking

* **Beads Epic ID**: `SL-nqa6`
* **User Stories Source**: `specledger/591-issue-tracking-upgrade/spec.md`
* **Research Inputs**: `specledger/591-issue-tracking-upgrade/research.md`
* **Planning Details**: `specledger/591-issue-tracking-upgrade/plan.md`
* **Data Model**: `specledger/591-issue-tracking-upgrade/data-model.md`
* **Quickstart Guide**: `specledger/591-issue-tracking-upgrade/quickstart.md`

## Beads Query Hints

Use the `bd` CLI to query and manipulate the issue graph:

```bash
# Find all open tasks for this feature
bd list --label spec:591-issue-tracking-upgrade --status open --limit 10

# Find ready tasks to implement
bd ready --label spec:591-issue-tracking-upgrade --limit 5

# See dependency tree for the epic
bd dep tree SL-nqa6

# View issues by phase
bd list --type feature --label 'spec:591-issue-tracking-upgrade'

# View issues by user story
bd list --label 'story:US1' --label 'spec:591-issue-tracking-upgrade'
```

## Tasks and Phases Structure

This feature follows Beads' 2-level graph structure:

* **Epic**: SL-nqa6 → Built-In Issue Tracker
* **Phases**: Beads issues of type `feature`, child of the epic
* **Tasks**: Issues of type `task`, children of each feature

### Phase Overview

| Phase | Issue ID | Description | Priority | Dependencies |
|-------|----------|-------------|----------|--------------|
| Setup | SL-iru3 | Project structure & deps | P1 | None |
| Foundational | SL-xajl | Core issue package | P1 | Setup |
| US1: Create/Manage Issues | SL-wyrp | Core CLI commands | P1 | Foundational |
| US8: Remove Beads/Perles | SL-5dza | Cleanup dependencies | P1 | Foundational |
| US2: Migrate Beads | SL-kv71 | Migration support | P2 | US1, US8 |
| US5: Duplicate Detection | SL-znei | Similarity warnings | P2 | US1 |
| US6: Definition of Done | SL-mghx | DoD enforcement | P2 | US1 |
| US7: Update Skills | SL-rq4c | Update prompts | P2 | US6 |
| US3: Dependencies | SL-0s5l | Link/unlink issues | P3 | US1 |
| US4: Cross-Spec Work | SL-n7zt | List across specs | P3 | US1 |
| Polish | SL-jtxv | Final improvements | P3 | All stories |

## Task Summary by Phase

### Phase 1: Setup (SL-iru3)
- SL-qcuq: Create pkg/issues directory structure
- SL-bf83: Add external dependencies to go.mod
- SL-fhsl: Create tests/issues directory structure

### Phase 2: Foundational (SL-xajl)
- SL-5yrs: Define Issue entity and types
- SL-hl62: Implement SHA-256 ID generation
- SL-njbb: Implement JSONL file store operations
- SL-y8gl: Implement spec context detection

### Phase 3: US1 - Create and Manage Issues (SL-wyrp) 🎯 MVP
- SL-yyux: Implement sl issue create command
- SL-3yzw: Implement sl issue list command
- SL-52z0: Implement sl issue show command
- SL-iu7i: Implement sl issue update command
- SL-7rzi: Implement sl issue close command
- SL-22i2: Register issue command in root CLI

### Phase 4: US8 - Remove Beads/Perles (SL-5dza)
- SL-ny61: Remove bd and perles from CoreTools
- SL-5w0k: Remove beads setup from init.sh
- SL-t70u: Delete setup-beads.sh script
- SL-idz9: Remove beads/perles from embedded mise.toml
- SL-wss2: Update bootstrap success messages

### Phase 5: US2 - Migrate Beads Data (SL-kv71)
- SL-hsga: Create Beads migrator struct and parsing
- SL-deko: Implement migration execution logic
- SL-ap7t: Implement migration cleanup logic
- SL-ziwd: Implement sl issue migrate command

### Phase 6: US5 - Prevent Duplicates (SL-znei)
- SL-e64t: Implement string similarity algorithm
- SL-itpo: Add duplicate detection to issue create
- SL-fm4t: Implement check-duplicates list command

### Phase 7: US6 - Definition of Done (SL-mghx)
- SL-6aal: Define DefinitionOfDone types
- SL-19ui: Implement DoD validation on close
- SL-mbtv: Add DoD update commands

### Phase 8: US7 - Update Skills/Prompts (SL-rq4c)
- SL-seqq: Update specledger.implement skill
- SL-yvde: Update specledger.tasks skill
- SL-rsow: Update embedded template skills

### Phase 9: US3 - Track Dependencies (SL-0s5l)
- SL-rawv: Implement dependency management in store
- SL-ab2s: Implement sl issue link/unlink commands
- SL-whww: Add --tree flag to issue list/show

### Phase 10: US4 - Cross-Spec Work (SL-n7zt)
- SL-oxpn: Implement cross-spec list functionality
- SL-tjb4: Add --spec flag to issue list

### Phase 11: Polish (SL-jtxv)
- SL-8pfn: Implement sl issue repair command
- SL-8luu: Add --json output to all commands
- SL-jvfi: Write quickstart.md validation
- SL-kba7: Integration test for CLI commands

## MVP Scope

**Suggested MVP**: Phase 1 + Phase 2 + Phase 3 (US1) + Phase 4 (US8)

This delivers:
- Core issue tracking (create, list, show, update, close)
- Removal of Beads/Perles dependencies
- Standalone operation with no external tools

Total MVP tasks: 18 tasks

## Execution Strategy

1. **Sequential Start**: Complete Setup → Foundational phases first
2. **Parallel Opportunities**: US1 and US8 can proceed in parallel after Foundational
3. **P1 First**: Complete all P1 stories before moving to P2
4. **P2 Parallel**: US2, US5, US6 can proceed in parallel (US7 waits for US6)
5. **P3 Last**: US3, US4, Polish complete the feature

## Label Conventions

| Label | Purpose |
|-------|---------|
| `spec:591-issue-tracking-upgrade` | All tasks in this feature |
| `phase:setup`, `phase:us1`, etc. | Phase identification |
| `story:US1`, `story:US2`, etc. | User story traceability |
| `component:cli`, `component:core` | Module mapping |
| `requirement:FR-001` | Functional requirement mapping |

---

> This file is an index only. All task data lives in Beads. Use `bd` commands to query and update tasks.
