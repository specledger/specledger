# Tasks Index: Skills Registry Integration

Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-90470e`
* **User Stories Source**: `specledger/610-skills-registry/spec.md`
* **Research Inputs**: `specledger/610-skills-registry/research.md`
* **Planning Details**: `specledger/610-skills-registry/plan.md`
* **Data Model**: `specledger/610-skills-registry/data-model.md`
* **Contract Definitions**: `specledger/610-skills-registry/contracts/`
* **Quickstart Scenarios**: `specledger/610-skills-registry/quickstart.md`

## Issue Query Hints

```bash
# All open tasks for this feature
sl issue list --label spec:610-skills-registry --status open

# View the full epic tree
sl issue show SL-90470e

# Filter by phase
sl issue list --label phase:foundational --status open
sl issue list --label phase:us1-us2 --status open

# Filter by user story
sl issue list --label story:US1 --status open
sl issue list --label story:US2 --status open
```

## Tasks and Phases Structure

```
SL-90470e (Epic: Skills Registry Integration)
├── SL-f97f5c (Phase 1: Setup)
│   └── SL-aafde3  T000: Create package + add go-vcr dep
│
├── SL-7bc63d (Phase 2: Foundational) ← blocked by Setup
│   ├── SL-664ae4  T001: Source identifier parser        [parallel]
│   ├── SL-2ce847  T002: HTTP client                     [parallel]
│   ├── SL-3a6930  T003: Lock file read/write            [parallel]
│   └── SL-72c546  T004: Folder hash computation         [parallel]
│
├── SL-b5ac49 (Phase 3: US1+US2 Search+Install P1) ← blocked by Foundational
│   ├── SL-cd4b97  T005: Skill discovery                 ← blocked by T001, T002
│   ├── SL-04d7eb  T006: Skill installation              [parallel with T005]
│   ├── SL-9c864c  T007: Telemetry                       [parallel with T005]
│   ├── SL-9f28f2  T008: sl skill search command         ← blocked by T001, T002
│   └── SL-51a866  T009: sl skill add command            ← blocked by T005, T006, T007
│
├── SL-46548d (Phase 4: US3+US4+US5 Info/List/Remove P2) ← blocked by Foundational
│   ├── SL-279648  T010: sl skill info command           [parallel]
│   ├── SL-745d54  T011: sl skill list command           [parallel]
│   └── SL-c06508  T012: sl skill remove command         [parallel]
│
├── SL-4943f4 (Phase 5: US6 Batch Audit P3) ← blocked by Foundational
│   └── SL-4d17f2  T013: sl skill audit command
│
├── SL-38ebee (Phase 6: US7 Agent Skill Template P2)
│   └── SL-50e427  T014: Create sl-skill embedded SKILL.md
│
└── SL-430f49 (Phase 7: Polish & Integration Tests) ← blocked by all above
    ├── SL-0c84c4  T015: Record VCR cassettes
    └── SL-ae3b59  T016: Write E2E integration tests
```

## Convention Summary

| Type    | Description                  | Parent           | Labels                                 |
| ------- | ---------------------------- | ---------------- | -------------------------------------- |
| epic    | Full feature epic            | _(none)_         | `spec:610-skills-registry`             |
| feature | Implementation phase / story | `--parent SL-90470e` | `phase:[n]`, `story:[US#]`        |
| task    | Implementation task          | `--parent <feature-id>` | `component:[x]`, `requirement:[FR-id]` |

## Dependency Graph

```
Setup ──→ Foundational ──┬──→ US1+US2 (P1) ──────→ Polish
                         ├──→ US3+US4+US5 (P2) ──→ Polish
                         ├──→ US6 (P3) ──────────→ Polish
                         └──→ US7 (P2, no deps) ─→ Polish

Within Foundational (all parallel):
  T001 Source Parser ─┐
  T002 HTTP Client ───┼─→ [both block T005 Discovery, T008 Search]
  T003 Lock File ─────┘ (independent)
  T004 Hash ───────────── (independent)

Within US1+US2:
  T005 Discovery ─────┐
  T006 Installation ──┼─→ T009 sl skill add
  T007 Telemetry ─────┘
  T008 sl skill search (parallel with T005-T007)

Within US3+US4+US5 (all parallel — different files, shared client):
  T010 sl skill info
  T011 sl skill list
  T012 sl skill remove
```

## Implementation Strategy

### MVP (P1): Search + Install

Complete Phases 1-3 (Setup → Foundational → US1+US2) for a working MVP:
- Users can search skills.sh and install skills
- Lock file tracks installations
- Telemetry reports back to ecosystem

### Incremental Delivery

| Increment | Stories | Commands Added |
|-----------|---------|----------------|
| MVP       | US1, US2 | `sl skill search`, `sl skill add` |
| +Management | US3, US4, US5 | `sl skill info`, `sl skill list`, `sl skill remove` |
| +Security | US6 | `sl skill audit` |
| +Agent | US7 | Embedded `sl-skill` template |
| +Quality | — | Integration tests, CLI compliance |

### Parallel Opportunities

1. **Foundational tasks** (T001-T004): All 4 can run in parallel — different files, no shared state
2. **US3+US4+US5 tasks** (T010-T012): All 3 commands can be built in parallel — each owns a distinct function in `skill.go` (`runSkillInfo`, `runSkillList`, `runSkillRemove`). If built by separate agents, each should only modify its own function + Cobra command registration. Merge via git for non-overlapping changes.
3. **US6 and US7**: Independent of each other, can run in parallel after Foundational
4. **US3+US4+US5 and US1+US2**: After Foundational, P2 stories can start in parallel with P1 (if staffed)

## Definition of Done Summary

| Issue ID | Task | DoD Items |
|----------|------|-----------|
| SL-aafde3 | Setup package | doc.go created; go-vcr v4 in go.mod; cassette dir exists; make build passes |
| SL-664ae4 | Source parser | Parse owner/repo, @skill, HTTPS URLs, SSH URLs; error messages suggest format; table-driven tests |
| SL-2ce847 | HTTP client | Search/FetchAudit/FetchSkillContent methods; configurable URLs; VCR cassettes; error handling |
| SL-3a6930 | Lock file | Read/write round-trip; alphabetical sort; fail fast on invalid JSON; empty file handling |
| SL-72c546 | Folder hash | Deterministic SHA-256; skip .git/node_modules; sort by path; include path in hash |
| SL-cd4b97 | Discovery | GitHub Trees API; git clone fallback; auto-retry on 404; YAML frontmatter; temp dir cleanup |
| SL-04d7eb | Installation | Write to all agent paths; create dirs; update lock; RemoveSkill helper |
| SL-9c864c | Telemetry | Fire-and-forget goroutine; 3s timeout; env var/CI gating; specledger-{version} |
| SL-9f28f2 | Search cmd | Compact output; footer hint; --json; --limit; friendly no-results; --help examples |
| SL-51a866 | Add cmd | Parse source; discover; audit display; confirm; install; telemetry; overwrite warn; --yes |
| SL-279648 | Info cmd | Metadata + 3-partner audit; --json; high/critical warning; --help examples |
| SL-745d54 | List cmd | Read lock file; compact list; footer hint; --json; empty state suggestion |
| SL-c06508 | Remove cmd | Delete dirs + lock entry; error if not installed; --json; --help examples |
| SL-4d17f2 | Audit cmd | Batch audit; single skill; 3-partner table; warning summary; --json |
| SL-50e427 | Agent skill template | YAML frontmatter; trigger conditions; 6 subcommands documented; JSON examples; audit guidance |
| SL-0c84c4 | VCR cassettes | 4 cassettes recorded; opt-in recording via env var |
| SL-ae3b59 | E2E tests | Tests for all 6 commands + error scenarios; make lint + make test pass |

## Agent Execution Flow

MCP agents and AI workflows should:

1. **Use `sl issue list --label spec:610-skills-registry --status open`** to find next tasks
2. **Check dependencies** with `sl issue show <id>` before starting
3. **Use this markdown only as a navigational anchor**

> Agents MUST NOT output tasks into this file. They MUST use `sl issue` CLI to record all task and phase structure.

## Status Tracking

Status is tracked in the issue store:

* **Open** → default
* **In Progress** → task being worked on
* **Closed** → complete

Use `sl issue list --label spec:610-skills-registry --status open` to query progress.
