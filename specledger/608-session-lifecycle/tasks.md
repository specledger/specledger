# Tasks Index: Session Lifecycle Management

Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-455424`
* **User Stories Source**: `specledger/608-session-lifecycle/spec.md`
* **Research Inputs**: N/A (no plan phase executed)
* **Planning Details**: N/A (no plan phase executed)
* **Data Model**: N/A
* **Contract Definitions**: N/A

## Issue Query Hints

```bash
# Find all open tasks for this feature
sl issue list --label spec:608-session-lifecycle --status open

# See all issues across specs
sl issue list --all --status open

# View issue details
sl issue show <issue-id>

# Find ready tasks (no unresolved blockers)
sl issue ready

# Link dependencies
sl issue link <from-id> blocks <to-id>
```

## Tasks and Phases Structure

* **Epic**: SL-455424 → Session Lifecycle Management
* **Phases** (feature issues, children of epic):
  * SL-38f997 — Foundational: Extend session types and clients
  * SL-927562 — US1: Prune Old Sessions (P1)
  * SL-0188f1 — US2: Configure Session TTL (P1)
  * SL-9b9029 — US3: Tag Sessions for Organization (P2)
  * SL-a85eb8 — US4: View Session Usage Statistics (P2)
  * SL-0fba88 — US5: Query Sessions Across Projects (P3)
  * SL-7c4228 — Polish: Cross-cutting concerns and hardening

## Phase Details

### Phase 1: Foundational (SL-38f997) — BLOCKS all user stories

| Task ID | Title | Labels |
|---------|-------|--------|
| SL-6f5b92 | Add Tags field to SessionMetadata type | FR-006 |
| SL-7a99f2 | Add Delete method to storage client | FR-003 |
| SL-9c3287 | Add Delete method to metadata client | FR-003 |
| SL-561ebe | Add TTL field to project config schema | FR-004 |
| SL-4645c9 | Add Supabase migration for tags column | FR-006 |

All 5 tasks are parallelizable (different files, no shared data).

### Phase 2: US1 — Prune Old Sessions (SL-927562) [P1 MVP]

| Task ID | Title | Depends On |
|---------|-------|------------|
| SL-55bf84 | Implement session prune logic | Foundational |
| SL-5f3470 | Add sl session prune CLI command | SL-55bf84 |

### Phase 3: US2 — Configure Session TTL (SL-0188f1) [P1]

| Task ID | Title | Depends On |
|---------|-------|------------|
| SL-422f34 | Add sl session config TTL CLI command | Foundational |

Integrates with US1 prune --expired mode.

### Phase 4: US3 — Tag Sessions (SL-9b9029) [P2]

| Task ID | Title | Depends On |
|---------|-------|------------|
| SL-22c7ab | Auto-tag sessions from branch name | Foundational |
| SL-649e59 | Add --tag flag to capture command | SL-22c7ab |
| SL-b43be5 | Add --tag filter to list command | SL-22c7ab |

SL-649e59 and SL-b43be5 can run in parallel after SL-22c7ab.

### Phase 5: US4 — Session Statistics (SL-a85eb8) [P2]

| Task ID | Title | Depends On |
|---------|-------|------------|
| SL-fd64d4 | Implement sl session stats command | Foundational |

### Phase 6: US5 — Cross-Project Queries (SL-0fba88) [P3]

| Task ID | Title | Depends On |
|---------|-------|------------|
| SL-51ce2e | Add --all-projects flag to session list | Foundational |

### Phase 7: Polish (SL-7c4228) — Blocked by ALL user stories

| Task ID | Title | Labels |
|---------|-------|--------|
| SL-68621e | Offline queue TTL-aware discard | FR-013 |

## Dependency Graph

```
Foundational (SL-38f997)
├──→ US1: Prune (SL-927562)
│    └─ Prune Logic → Prune CLI
├──→ US2: TTL (SL-0188f1) [parallel to US1]
│    └─ TTL CLI integration
├──→ US3: Tags (SL-9b9029) [parallel to US1/US2]
│    └─ Auto-tag → Tag flag [P]
│                → Tag filter [P]
├──→ US4: Stats (SL-a85eb8) [parallel to US1-US3]
│    └─ Stats command
├──→ US5: Cross-project (SL-0fba88) [parallel to US1-US4]
│    └─ All-projects flag
└──→ Polish (SL-7c4228) [blocked by all above]
     └─ Queue TTL discard
```

## Definition of Done Summary

| Issue ID | DoD Items |
|----------|-----------|
| SL-6f5b92 | - Tags field added to SessionMetadata<br>- JSON tag includes omitempty |
| SL-7a99f2 | - Delete method on storage client<br>- Authenticated HTTP DELETE<br>- Error on failure |
| SL-9c3287 | - Delete method on metadata client<br>- Authenticated HTTP DELETE<br>- Error on failure |
| SL-561ebe | - TTL field in config schema<br>- Default 30 days<br>- YAML round-trip |
| SL-4645c9 | - Migration file created<br>- Tags text[] column<br>- GIN index<br>- Backward compatible |
| SL-55bf84 | - PruneSessions function<br>- Deletes storage + metadata<br>- Dry-run mode<br>- Partial failure handling<br>- Auth check |
| SL-5f3470 | - prune subcommand<br>- --days flag (default 30)<br>- --dry-run flag<br>- --expired flag uses TTL<br>- Auth error message<br>- Output counts |
| SL-422f34 | - TTL from config when --expired<br>- --days overrides TTL<br>- Default 30 days |
| SL-22c7ab | - Tag extraction from branch<br>- Tags in SessionMetadata<br>- Deduplication |
| SL-649e59 | - --tag flag on capture<br>- Merged with auto-tags<br>- Deduplication |
| SL-b43be5 | - --tag filter on list<br>- PostgREST tag query<br>- Multi-tag matching |
| SL-fd64d4 | - stats subcommand<br>- Total count/size<br>- Per-branch distribution<br>- Message averages<br>- Date range<br>- Empty state<br>- Missing metadata warnings |
| SL-51ce2e | - --all-projects flag<br>- No project_id filter<br>- Project identifier in results<br>- Auth required |
| SL-68621e | - TTL check in ProcessQueue<br>- Expired entries discarded<br>- Non-expired processed |

## Implementation Strategy

### MVP (Recommended)
**US1: Prune Old Sessions** — Delivers the most pressing operational value (storage cleanup). Requires only the foundational phase + 2 tasks.

### Incremental Delivery
1. **Foundational** → All 5 tasks in parallel
2. **US1 + US2** → Prune + TTL (both P1, can run in parallel)
3. **US3 + US4** → Tags + Stats (both P2, can run in parallel)
4. **US5** → Cross-project queries (P3)
5. **Polish** → Queue TTL discard

### Parallel Opportunities
- All foundational tasks (5 tasks, different files)
- US1 and US2 (independent P1 stories)
- US3 and US4 (independent P2 stories)
- Within US3: tag flag and tag filter (after auto-tag completes)

---

> This file is intentionally light and index-only. Implementation data lives in the issue store. Update this file only to point humans and agents to canonical query paths and feature references.
