# Tasks Index: Session-to-Knowledge Memory Pipeline

Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-189675`
* **User Stories Source**: `specledger/607-session-memory-pipeline/spec.md`
* **Research Inputs**: `specledger/607-session-memory-pipeline/research.md`
* **Planning Details**: `specledger/607-session-memory-pipeline/plan.md`
* **Data Model**: `specledger/607-session-memory-pipeline/data-model.md`

## Issue Query Hints

Use the `sl issue` CLI to query and manipulate the issue graph:

```bash
# Find all open tasks for this feature
sl issue list --label "spec:607-session-memory-pipeline" --status open

# Find ready tasks (unblocked)
sl issue ready

# See all issues across specs
sl issue list --all --status open

# View issue details
sl issue show <issue-id>
```

## Tasks and Phases Structure

* **Epic**: SL-189675 → Session-to-Knowledge Memory Pipeline
* **Phases**: Issues of type `feature`, child of the epic
  * Foundational: SL-01f95e (Core Memory Package — blocks all user stories)
  * US1: SL-d69ec6 (Knowledge Retrieval — P1)
  * US2: SL-5ad579 (Knowledge Extraction — P1)
  * US3: SL-c4622c (Scoring and Promotion — P2)
  * US4: SL-13bccd (View and Manage — P2)
  * US5: SL-22d38c (Cloud Sync — P3)

## Convention Summary

| Type    | Description                  | Labels                                 |
| ------- | ---------------------------- | -------------------------------------- |
| epic    | Full feature epic            | `spec:607-session-memory-pipeline`     |
| feature | Implementation phase / story | `phase:<name>`, `story:US#`            |
| task    | Implementation task          | `component:<x>`, `requirement:FR-###`  |

---

## Phase 1: Foundational — Core Memory Package (SL-01f95e)

**Purpose**: Core infrastructure that MUST complete before ANY user story

| Task ID    | Title                              | Priority | Depends On |
|------------|------------------------------------|----------|------------|
| SL-4e2bd2  | Define KnowledgeEntry types        | P0       | —          |
| SL-47cd52  | Implement JSONL knowledge store    | P0       | SL-4e2bd2  |
| SL-bdc4d9  | Implement three-axis scorer        | P0       | SL-4e2bd2  |

**Parallel**: SL-47cd52 and SL-bdc4d9 can run in parallel after SL-4e2bd2.

**Checkpoint**: Foundation ready — user story implementation can begin.

---

## Phase 2: US1 — Knowledge Retrieval (P1) (SL-d69ec6)

**Goal**: AI agents receive relevant knowledge at session start.

**Independent Test**: Pre-populate knowledge.md, start session, verify agent context includes it.

| Task ID    | Title                              | Priority | Depends On |
|------------|------------------------------------|----------|------------|
| SL-abdc50  | Implement knowledge.md renderer    | P1       | SL-47cd52  |
| SL-940684  | Create sl-memory skill             | P1       | — (parallel) |
| SL-2f6c87  | Add memory/cache/ to .gitignore    | P1       | — (parallel) |

**Parallel**: SL-940684 and SL-2f6c87 are independent of renderer.

**Checkpoint**: Promoted entries render to knowledge.md; skill loads it into agent context.

---

## Phase 3: US2 — Knowledge Extraction (P1) (SL-5ad579)

**Goal**: Extract structured knowledge entries from session transcripts via AI.

**Independent Test**: Provide transcript, verify structured entries with titles, tags, scores.

| Task ID    | Title                              | Priority | Depends On |
|------------|------------------------------------|----------|------------|
| SL-c3ad17  | Create /specledger.memory L2 cmd   | P1       | SL-47cd52  |
| SL-0ac06f  | Implement sl memory show command   | P1       | SL-47cd52  |

**Parallel**: Both tasks can run in parallel (different files).

**Checkpoint**: Extraction produces entries; `sl memory show` displays them.

---

## Phase 4: US3 — Scoring and Promotion (P2) (SL-c4622c)

**Goal**: Auto-promote entries with composite >= 7.0; manual promote/demote.

**Independent Test**: Create entries with known scores, verify promotion logic.

| Task ID    | Title                              | Priority | Depends On           |
|------------|------------------------------------|----------|----------------------|
| SL-0b4747  | Implement merge/dedup logic        | P2       | SL-47cd52            |
| SL-8109ab  | Implement promote/demote commands  | P2       | SL-bdc4d9, SL-abdc50 |

**Sequential**: SL-8109ab depends on scorer and renderer.

**Checkpoint**: Scoring, auto-promotion, and manual promote/demote work end-to-end.

---

## Phase 5: US4 — View and Manage (P2) (SL-13bccd)

**Goal**: Users can delete entries and access all memory commands from CLI.

**Independent Test**: Populate entries, verify delete and full CLI registration.

| Task ID    | Title                              | Priority | Depends On                          |
|------------|------------------------------------|----------|-------------------------------------|
| SL-b49e90  | Implement sl memory delete command | P2       | (foundational)                      |
| SL-fb23eb  | Register VarMemoryCmd in main.go   | P2       | SL-0ac06f, SL-8109ab, SL-b49e90    |

**Sequential**: Registration must wait until all subcommands exist.

**Checkpoint**: `sl memory` fully registered with show, promote, demote, delete.

---

## Phase 6: US5 — Cloud Sync (P3) (SL-22d38c)

**Goal**: Knowledge entries sync with Supabase for cross-machine/team access.

**Independent Test**: Create local entries, sync to cloud, pull on fresh machine.

| Task ID    | Title                              | Priority | Depends On |
|------------|------------------------------------|----------|------------|
| SL-c86ee2  | Implement Supabase sync client     | P3       | SL-47cd52  |
| SL-d012d2  | Implement pull/sync CLI commands   | P3       | SL-c86ee2  |
| SL-bde229  | Create Supabase migration          | P3       | — (parallel) |

**Parallel**: SL-bde229 (migration) can run in parallel with SL-c86ee2.

**Checkpoint**: Full cloud round-trip: push, pull, offline resilience.

---

## Dependencies & Execution Order

### Phase Dependencies

```
Foundational (SL-01f95e) ──blocks──▶ US1 (SL-d69ec6) [P1]
                          ──blocks──▶ US2 (SL-5ad579) [P1]
                          ──blocks──▶ US3 (SL-c4622c) [P2]
                          ──blocks──▶ US4 (SL-13bccd) [P2]
                          ──blocks──▶ US5 (SL-22d38c) [P3]
```

### User Story Independence

- **US1 + US2** (P1): Can run in parallel after Foundational
- **US3 + US4** (P2): Can start after Foundational; US4 registration depends on US2 show + US3 promote/demote
- **US5** (P3): Independent of US1-US4 after Foundational

### MVP Scope

**Suggested MVP**: Foundational + US1 + US2 (P1 stories only)

This delivers the core value: extract knowledge from sessions and inject it into agent context. Scoring, management, and cloud sync can follow incrementally.

---

## Definition of Done Summary

| Issue ID   | DoD Items |
|------------|-----------|
| SL-4e2bd2  | - KnowledgeEntry struct defined<br>- Score struct defined<br>- EntryStatus constants defined<br>- Validation functions implemented |
| SL-47cd52  | - JSONL read/write operations work<br>- CRUD operations implemented<br>- File locking for concurrent access<br>- entries.jsonl created in correct path |
| SL-bdc4d9  | - Composite score calculated correctly<br>- Threshold check at 7.0 works<br>- Each axis validated 0-10 range |
| SL-abdc50  | - knowledge.md generated from promoted entries<br>- Entries grouped by tag category<br>- Source and score metadata included<br>- Empty state handled |
| SL-940684  | - skill.md reads knowledge.md<br>- Skill loads into agent context |
| SL-2f6c87  | - .gitignore updated with memory/cache/ pattern |
| SL-c3ad17  | - L2 command with summarize, tag, patterns, synthesize modes<br>- Structured entries output with scores and tags |
| SL-0ac06f  | - show subcommand registered<br>- Entries displayed with title, score, status, tags<br>- Sorted by composite score<br>- Empty state handled |
| SL-0b4747  | - FindSimilar method implemented<br>- Merge method combines entries<br>- Recurrence score and count incremented |
| SL-8109ab  | - promote subcommand implemented<br>- demote subcommand implemented<br>- knowledge.md regenerated on change<br>- Auto-promotion at 7.0 integrated |
| SL-b49e90  | - delete subcommand implemented<br>- Entry removed from store<br>- knowledge.md regenerated if needed<br>- Confirmation prompt shown |
| SL-fb23eb  | - VarMemoryCmd registered in main.go<br>- sl memory --help shows all subcommands |
| SL-c86ee2  | - Push method uploads to cloud<br>- Pull method downloads to local<br>- Offline error handling |
| SL-d012d2  | - pull subcommand implemented<br>- sync subcommand implemented<br>- Success/failure counts reported |
| SL-bde229  | - Migration SQL file created<br>- All columns mapped<br>- RLS policies added |

---

> This file is intentionally light and index-only. Implementation data lives in the issue store. Update this file only to point humans and agents to canonical query paths and feature references.
