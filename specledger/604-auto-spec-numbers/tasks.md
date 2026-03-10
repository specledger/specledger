# Tasks Index: Auto-Generate Spec Numbers

Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-09129b`
* **User Stories Source**: `specledger/604-auto-spec-numbers/spec.md`
* **Research Inputs**: `specledger/604-auto-spec-numbers/research.md`
* **Planning Details**: `specledger/604-auto-spec-numbers/plan.md`

## Issue Query Hints

```bash
# All issues for this feature
sl issue list --label spec:604-auto-spec-numbers

# Open tasks only
sl issue list --label spec:604-auto-spec-numbers --status open

# View issue details
sl issue show SL-09129b

# Filter by user story
sl issue list --label story:US1
sl issue list --label story:US2
sl issue list --label story:US3
```

## Tasks and Phases Structure

```
Epic: SL-09129b (Auto-generate spec numbers)
│
├─ Phase 1 (Setup): SL-b436da
│  ├─ T001: SL-58dfed  Add GetNextAvailableNum (FR-002)     [US1]
│  ├─ T002: SL-a2e270  Make --number flag optional (FR-001)  [US1] ← blocked by T001
│  ├─ T003: SL-472191  Add collision suggestion (FR-004)     [US2]
│  └─ T004: SL-b56f9f  Add FEATURE_ID to JSON output (FR-005) [US3]
│
├─ Phase 2 (US1): SL-7fe983  ← blocked by Setup
│  └─ All setup tasks cover US1 implementation
│
├─ Phase 3 (US2): SL-45e767  ← blocked by Setup
│  └─ All setup tasks cover US2 implementation
│
├─ Phase 4 (US3): SL-b9b142  ← blocked by Setup
│  └─ T005: SL-4274ff  Update specledger.specify skill docs (FR-007)
│
└─ Polish:
   └─ T006: SL-f55640  Unit tests for collision and auto-increment
```

## Convention Summary

| Type    | Description                  | Labels                                 |
| ------- | ---------------------------- | -------------------------------------- |
| epic    | Full feature epic            | `spec:604-auto-spec-numbers`           |
| feature | Implementation phase / story | `phase:setup`, `story:US1`             |
| task    | Implementation task          | `component:cli`, `requirement:FR-001`  |

## Definition of Done Summary

| Issue ID   | DoD Items |
|------------|-----------|
| SL-58dfed  | - GetNextAvailableNum implemented<br>- Handles empty dir (returns 001)<br>- Skips colliding numbers<br>- Best-effort remote check |
| SL-a2e270  | - Remove --number required error<br>- Auto-generate when empty<br>- Print to stderr<br>- Manual override works |
| SL-472191  | - Error shows collision detail<br>- Suggests next available number |
| SL-b56f9f  | - FeatureID field in struct<br>- JSON output includes FEATURE_ID |
| SL-4274ff  | - Step 2 simplified<br>- Manual scan removed<br>- FEATURE_ID documented |
| SL-f55640  | - 8 unit tests pass<br>- Covers empty, existing, collision, parse |

## Implementation Strategy

**MVP**: User Story 1 (P1) — auto-generate numbers without `--number` flag. This single change eliminates 90% of the friction.

**Incremental delivery**:
1. Setup phase delivers all core code changes (T001-T004)
2. US1/US2/US3 are effectively complete after Setup since all code changes are in the same files
3. US3 adds documentation update
4. Polish adds test coverage

**Parallel opportunities**: T001 and T003/T004 can be implemented in parallel (different functions). US2 and US3 are independent of each other.

---

> This file is intentionally light and index-only. Implementation data lives in the issue store.
