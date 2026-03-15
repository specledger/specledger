# Tasks Index: Gitattributes Merge with Linguist-Generated Markers

Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-7bb372`
* **User Stories Source**: `specledger/609-gitattributes-merge/spec.md`
* **Research Inputs**: `specledger/609-gitattributes-merge/research.md`
* **Planning Details**: `specledger/609-gitattributes-merge/plan.md`
* **Data Model**: `specledger/609-gitattributes-merge/data-model.md`

## Issue Query Hints

```bash
# Find all open tasks for this feature
sl issue list --label spec:609-gitattributes-merge --status open

# See all issues across specs
sl issue list --all --status open

# View issue details
sl issue show SL-7bb372

# Link dependencies
sl issue link [from-id] blocks [to-id]
```

## Tasks and Phases Structure

* **Epic**: SL-7bb372 → Gitattributes Merge with Linguist-Generated Markers
* **Phases**: Issues of type `feature`, child of the epic
  * Phase 1: Setup (template + manifest + types)
  * Phase 2: Foundational (merge function + tests)
  * Phase 3: US1+US2 (wire merge into copy flow)
  * Phase 4: US3 (idempotency verification)
  * Phase 5: Integration Tests (sl init + sl doctor against real binary)
  * Phase 6: Polish (build, test, manual e2e)
* **Tasks**: Issues of type `task`, children of each feature issue (phase)

## Phase 1: Setup (Shared Infrastructure)

**Feature**: `SL-5217cc` — Setup Phase

| Issue | Title | Priority |
|-------|-------|----------|
| SL-3da959 | Populate .gitattributes template with linguist-generated patterns | 1 |
| SL-e5422b | Add Mergeable field to Playbook struct and FilesMerged to CopyResult | 1 |
| SL-b679d1 | Add mergeable list to manifest.yaml | 1 |

All 3 tasks are parallelizable (different files, no dependencies).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Feature**: `SL-395149` — Foundational Phase
**Blocked by**: Setup (SL-5217cc)

| Issue | Title | Priority | Blocked By |
|-------|-------|----------|------------|
| SL-03de35 | Implement MergeSentinelSection() pure function | 1 | — |
| SL-9f170e | Add table-driven tests for MergeSentinelSection | 1 | SL-03de35 |

**Checkpoint**: Merge function implemented and tested — user story work can begin.

---

## Phase 3: US1+US2 — First-time init + Re-init merge (Priority: P1)

**Feature**: `SL-661f90` — US1+US2: Wire merge into copy flow
**Blocked by**: Foundational (SL-395149)

**Goal**: `sl init` creates/merges .gitattributes with sentinel block. Works for both new and existing files. Also works via `sl doctor --template`.

| Issue | Title | Priority | Blocked By |
|-------|-------|----------|------------|
| SL-ca79fc | Wire mergeableMap into CopyPlaybooks and copyStructureItem | 1 | — |
| SL-a3e143 | Implement mergeFile function in copy.go | 1 | SL-ca79fc |
| SL-ec3dd1 | Update ApplyToProject reporting for merged files | 2 | SL-a3e143 |

**Checkpoint**: `sl init` and `sl doctor --template` both merge .gitattributes correctly.

---

## Phase 4: US3 — Idempotent Re-runs (Priority: P2)

**Feature**: `SL-a46126` — US3: Idempotent Re-runs Verification
**Blocked by**: Foundational (SL-395149)

| Issue | Title | Priority |
|-------|-------|----------|
| SL-8f01ea | Add idempotency test case to merge_test.go | 2 |

Can run in parallel with Phase 3.

---

## Phase 5: Integration Tests (sl init + sl doctor against real binary)

**Feature**: `SL-a874fd` — Integration Tests: sl init and sl doctor gitattributes merge
**Blocked by**: US1+US2 (SL-661f90)

**Strategy**: Build `sl` binary via `go build`, set up isolated fixture directories with `t.TempDir()`, run binary via `os/exec.Command()`. All tests parallel-safe. Follows existing patterns from `tests/integration/bootstrap_test.go`.

| Issue | Title | Priority | Story |
|-------|-------|----------|-------|
| SL-2e63b3 | Test: sl init creates .gitattributes in new project | 1 | US1 |
| SL-be6192 | Test: sl init merges into existing .gitattributes | 1 | US2 |
| SL-46cccf | Test: sl init updates existing sentinel block | 1 | US2 |
| SL-816815 | Test: sl init is idempotent | 1 | US3 |
| SL-d32abf | Test: sl init --force merges (not overwrites) | 1 | FR-010 |
| SL-f66127 | Test: sl init handles malformed sentinel | 1 | FR-011 |
| SL-eff94a | Test: sl doctor --template merges .gitattributes | 1 | — |
| SL-b41344 | Test: sl doctor --template is idempotent | 1 | US3 |

All 8 tests are parallelizable (each uses its own `t.TempDir()`). All go in `tests/integration/gitattributes_test.go`.

**Checkpoint**: All integration tests pass with `go test ./tests/integration/ -run TestGitattributes -v`

---

## Phase 6: Polish & Cross-Cutting Concerns

**Feature**: `SL-b12161` — Polish and Cross-Cutting Concerns
**Blocked by**: US1+US2 (SL-661f90), US3 (SL-a46126), and Integration Tests (SL-a874fd)

| Issue | Title | Priority | Blocked By |
|-------|-------|----------|------------|
| SL-802088 | Run go build and full test suite | 2 | — |
| SL-a29fe5 | Manual end-to-end testing per quickstart.md | 2 | SL-802088 |

---

## Dependencies & Execution Order

```
Setup (SL-5217cc) → Foundational (SL-395149) ──┬──→ US1+US2 (SL-661f90) ──→ Integration Tests (SL-a874fd) ──→ Polish (SL-b12161)
  ├─ SL-3da959       ├─ SL-03de35               │    ├─ SL-ca79fc              ├─ SL-2e63b3 (init creates)      ├─ SL-802088 (build)
  ├─ SL-e5422b       └─ SL-9f170e               │    ├─ SL-a3e143              ├─ SL-be6192 (init merges)       └─ SL-a29fe5 (e2e)
  └─ SL-b679d1                                   │    └─ SL-ec3dd1              ├─ SL-46cccf (init updates)
                                                  │                              ├─ SL-816815 (init idempotent)
                                                  └──→ US3 (SL-a46126) ─────────┤├─ SL-d32abf (init force)
                                                       └─ SL-8f01ea             ├─ SL-f66127 (malformed)
                                                                                 ├─ SL-eff94a (doctor merges)
                                                                                 └─ SL-b41344 (doctor idempotent)
```

## Implementation Strategy

- **MVP**: Phase 1 + Phase 2 + Phase 3 = `sl init` creates/merges .gitattributes with sentinel block
- **Full delivery**: + Phase 4 + Phase 5 (integration tests) + Phase 6 (polish)
- **Total tasks**: 19 (11 implementation + 8 integration tests)
- **Parallel opportunities**: Setup tasks (3 parallel), US1+US2 and US3 (parallel after Foundational), all 8 integration tests (parallel)

## Definition of Done Summary

| Issue ID | DoD Items |
|----------|-----------|
| SL-3da959 | - .gitattributes template contains issues.jsonl pattern<br>- .gitattributes template contains tasks.md pattern<br>- No other files marked as linguist-generated |
| SL-e5422b | - Mergeable field added to Playbook struct<br>- FilesMerged field added to CopyResult struct |
| SL-b679d1 | - mergeable list added to manifest.yaml<br>- .gitattributes listed as mergeable |
| SL-03de35 | - merge.go created with MergeSentinelSection function<br>- Handles all 4 sentinel states<br>- Idempotent |
| SL-9f170e | - merge_test.go with table-driven tests<br>- All edge cases covered<br>- All tests pass |
| SL-ca79fc | - mergeableMap built and passed through<br>- Mergeable check before protected check |
| SL-a3e143 | - mergeFile function created<br>- Reads embedded + existing, merges, writes<br>- Respects DryRun |
| SL-ec3dd1 | - FilesMerged reported in output<br>- No output when 0 |
| SL-8f01ea | - Idempotency test case added and passes |
| SL-802088 | - go build succeeds<br>- All tests pass<br>- No regressions |
| SL-a29fe5 | - All 4 manual scenarios pass per quickstart.md |
| SL-2e63b3 | - Test passes: .gitattributes created with sentinel block |
| SL-be6192 | - Test passes: user content preserved, sentinel appended |
| SL-46cccf | - Test passes: sentinel replaced, no duplication |
| SL-816815 | - Test passes: two runs produce identical output |
| SL-d32abf | - Test passes: --force preserves user content |
| SL-f66127 | - Test passes: malformed sentinel auto-fixed |
| SL-eff94a | - Test passes: sl doctor --template merges correctly |
| SL-b41344 | - Test passes: sl doctor --template idempotent |
