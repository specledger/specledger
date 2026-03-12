# Tasks Index: Improve Issue Parent-Child Linking

Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-50b063`
* **User Stories Source**: `specledger/606-issue-parent-linking/spec.md`
* **Research Inputs**: `specledger/606-issue-parent-linking/research.md`
* **Planning Details**: `specledger/606-issue-parent-linking/plan.md`

## Issue Query Hints

```bash
# Find all open tasks for this feature
sl issue list --label spec:606-issue-parent-linking --status open

# See all issues across specs
sl issue list --all --status open

# View issue details
sl issue show SL-50b063
```

## Tasks and Phases Structure

* **Epic**: SL-50b063 → Improve Issue Parent-Child Linking
* **Phases**:
  * **US1** (SL-65d808): Link Parent via Issue Link Command [P1]
    * SL-cf175b: Add parent case to link command switch [FR-001, FR-002, FR-003]
  * **US4** (SL-0fb4cd): Update AI Skill Instructions for --parent [P1]
    * SL-a1dbd4: Emphasize --parent as mandatory in specledger.tasks skill [FR-007]
  * **US2** (SL-d0aead): Warn About Orphaned Issues [P2]
    * SL-34d6e6: Add Orphaned field to ListFilter struct [FR-004]
    * SL-1166a6: Add orphan filtering logic to store List method [FR-004]
    * SL-a71d70: Add --orphaned flag to issue list command [FR-004]
  * **US3** (SL-133ce2): Bulk Reparent Issues [P3]
    * SL-335f39: Add reparent subcommand to issue command [FR-005, FR-006]

## Dependencies & Execution Order

```
SL-cf175b (parent link type)     ── no deps, ready immediately
SL-a1dbd4 (skill instructions)  ── no deps, ready immediately [parallel with US1]

SL-34d6e6 (ListFilter field) ──→ SL-1166a6 (store logic) ──→ SL-a71d70 (CLI flag)

SL-335f39 (reparent command)    ── no deps, ready immediately [parallel with all]
```

**Parallel opportunities**:
- US1 (SL-cf175b), US4 (SL-a1dbd4), and US3 (SL-335f39) can all run in parallel — different files, no shared state
- US2 tasks are sequential (same data flow: model → store → CLI)

## Definition of Done Summary

| Issue ID   | DoD Items |
|------------|-----------|
| SL-cf175b  | - parent case added to link command switch<br>- Routes through store.Update with ParentID<br>- Reuses existing cycle detection<br>- Error on nonexistent parent ID<br>- Error on self-parent |
| SL-a1dbd4  | - Bold/CRITICAL instruction added for --parent requirement<br>- All feature creation examples include --parent<br>- All task creation examples include --parent<br>- Post-creation validation reminder added |
| SL-34d6e6  | - Orphaned bool added to ListFilter |
| SL-1166a6  | - Orphan filter logic added to List method<br>- Epics excluded from orphan results<br>- Filter composes correctly with other filters |
| SL-a71d70  | - --orphaned flag registered on list command<br>- Filter passed to store.List correctly<br>- Success message when no orphans found |
| SL-335f39  | - reparent subcommand registered<br>- Parent existence validated upfront<br>- Each child updated with ParentID<br>- Continue-on-error for invalid child IDs<br>- Summary of successes and failures printed |

## Implementation Strategy

### MVP Scope

**MVP = US1 + US4** (both P1): Add `parent` link type and fix AI skill instructions. This addresses both the immediate usability gap and the root cause (agents not using `--parent`).

### Incremental Delivery

1. **MVP**: US1 + US4 — link parent type + skill instructions (parallel, immediate)
2. **P2**: US2 — orphan detection for post-hoc verification
3. **P3**: US3 — bulk reparent for large-scale remediation

### Story Testability

- **US1**: Create epic + task, run `sl issue link <task> parent <epic>`, verify with `sl issue show`
- **US2**: Create mix of issues with/without parents, run `sl issue list --orphaned`, verify only orphans shown
- **US3**: Create orphaned tasks, run `sl issue reparent <parent> <t1> <t2> <t3>`, verify all reparented
- **US4**: Review `.claude/commands/specledger.tasks.md` for explicit `--parent` guidance

---

> This file is intentionally light and index-only. Implementation data lives in the issue store.
