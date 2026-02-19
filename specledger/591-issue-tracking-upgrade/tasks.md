# Tasks Index: Built-In Issue Tracker

Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-nqa6`
* **User Stories Source**: `specledger/591-issue-tracking-upgrade/spec.md`
* **Research Inputs**: `specledger/591-issue-tracking-upgrade/research.md`
* **Planning Details**: `specledger/591-issue-tracking-upgrade/plan.md`
* **Data Model**: `specledger/591-issue-tracking-upgrade/data-model.md`
* **Quickstart Guide**: `specledger/591-issue-tracking-upgrade/quickstart.md`

## Issue Query Hints

Use the `sl issue` CLI to query and manipulate the issue graph:

```bash
# Find all open tasks for this feature
sl issue list --status open --label "spec:591-issue-tracking-upgrade"

# See all issues across specs
sl issue list --all

# View issue details
sl issue show SL-xxxxxx

# Link dependencies
sl issue link SL-xxxxx blocks SL-yyyyy
```

## Tasks and Phases Structure

This feature follows a 2-level graph structure:

* **Epic**: SL-nqa6 → Built-In Issue Tracker
* **Phases**: Issues of type `feature`, child of the epic
* **Tasks**: Issues of type `task`, children of each feature

### Phase Overview

| Phase | Description | Priority | Dependencies |
|-------|-------------|----------|--------------|
| Setup | Project structure & deps | P1 | None |
| Foundational | Core issue package | P1 | Setup |
| US1: Create/Manage Issues | Core CLI commands | P1 | Foundational |
| US8: Remove Beads/Perles | Cleanup dependencies | P1 | Foundational |
| US2: Migrate Beads | Migration support | P2 | US1, US8 |
| US5: Duplicate Detection | Similarity warnings | P2 | US1 |
| US6: Definition of Done | DoD enforcement | P2 | US1 |
| US7: Update Skills | Update prompts | P2 | US6 |
| US3: Dependencies | Link/unlink issues | P3 | US1 |
| US4: Cross-Spec Work | List across specs | P3 | US1 |
| Polish | Final improvements | P3 | All stories |

## Label Conventions

| Label | Purpose |
|-------|---------|
| `spec:591-issue-tracking-upgrade` | All tasks in this feature |
| `phase:setup`, `phase:us1`, etc. | Phase identification |
| `story:US1`, `story:US2`, etc. | User story traceability |
| `component:cli`, `component:core` | Module mapping |
| `requirement:FR-001` | Functional requirement mapping |

## MVP Scope

**Suggested MVP**: Setup + Foundational + US1 + US8

This delivers:
- Core issue tracking (create, list, show, update, close)
- Removal of Beads/Perles dependencies
- Standalone operation with no external tools

## Execution Strategy

1. **Sequential Start**: Complete Setup → Foundational phases first
2. **Parallel Opportunities**: US1 and US8 can proceed in parallel after Foundational
3. **P1 First**: Complete all P1 stories before moving to P2
4. **P2 Parallel**: US2, US5, US6 can proceed in parallel (US7 waits for US6)
5. **P3 Last**: US3, US4, Polish complete the feature

---

> This file is an index only. All task data lives in the issue store. Use `sl issue` commands to query and update tasks.
