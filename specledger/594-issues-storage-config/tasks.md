# Tasks Index: Issues Storage Configuration

Task graph index for feature 594-issues-storage-config implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-775ab4`
* **User Stories Source**: `specledger/594-issues-storage-config/spec.md`
* **Research Inputs**: `specledger/594-issues-storage-config/research.md`
* **Planning Details**: `specledger/594-issues-storage-config/plan.md`
* **Data Model**: `specledger/594-issues-storage-config/data-model.md`
* **Quickstart Tests**: `specledger/594-issues-storage-config/quickstart.md`

## Query Hints

```bash
# Find all open tasks for this feature
sl issue list --label spec:594-issues-storage-config --status open

# Find ready tasks (no blocking dependencies)
sl issue list --label spec:594-issues-storage-config --status open --blocked false

# View task dependencies
sl issue show SL-734037

# List by story
sl issue list --label story:US1
sl issue list --label story:US2
sl issue list --label story:US3
```

## Tasks Summary

| ID | Title | Story | Priority | Status |
|----|-------|-------|----------|--------|
| SL-734037 | Rename lock file to issues.jsonl.lock | US1 | 1 | open |
| SL-096cb1 | Load artifact_path from specledger.yaml | US2 | 1 | open |
| SL-648ba0 | Add issues.jsonl.lock to .gitignore | US3 | 2 | open |
| SL-3a3b40 | Run quickstart.md validation tests | Polish | 2 | open |

## Dependency Graph

```
SL-734037 (US1: Lock naming) ──┬──► SL-648ba0 (US3: Gitignore)
                                │
                                └──► SL-3a3b40 (Polish)
                                               ▲
SL-096cb1 (US2: Artifact path) ────────────────┘
```

## Execution Order

### Phase 1: Core Changes (Parallel)

These tasks can run in parallel as they modify different files:

- **SL-734037** [US1] Rename lock file to issues.jsonl.lock
  - File: `pkg/issues/store.go`
  - No dependencies

- **SL-096cb1** [US2] Load artifact_path from specledger.yaml
  - File: `pkg/cli/commands/issue.go`
  - No dependencies

### Phase 2: Configuration (Depends on US1)

- **SL-648ba0** [US3] Add issues.jsonl.lock to .gitignore
  - File: `.gitignore`
  - Depends on: SL-734037

### Phase 3: Validation (Depends on all)

- **SL-3a3b40** [Polish] Run quickstart.md validation tests
  - Depends on: SL-734037, SL-096cb1, SL-648ba0

## MVP Scope

**Minimum Viable Product**: Complete US1 + US2 (SL-734037 + SL-096cb1)

This delivers the core value:
- Lock files named correctly for gitignore compatibility
- Custom artifact_path support for project flexibility

## Definition of Done Summary

| Issue ID | DoD Items |
|----------|-----------|
| SL-734037 | - Lock file created as issues.jsonl.lock<br>- Cross-spec operations use new naming<br>- Existing functionality unchanged |
| SL-096cb1 | - Issues stored in custom artifact_path<br>- Default path works when no config<br>- All CLI commands respect artifact_path<br>- List --all searches in correct path |
| SL-648ba0 | - .gitignore contains pattern<br>- git status doesn't show lock files<br>- Lock files ignored by git |
| SL-3a3b40 | - All 6 quickstart tests pass |

## Parallel Execution Notes

Agents can work on SL-734037 and SL-096cb1 simultaneously:
- Different files (store.go vs issue.go)
- No shared dependencies
- Both are P1 priority

After both complete, SL-648ba0 can start.
After all three complete, SL-3a3b40 can validate.
