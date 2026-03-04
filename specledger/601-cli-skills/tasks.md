# Tasks: New CLI Commands and Skills

**Epic**: SL-3a436b - New CLI Commands and Skills
**Feature**: 601-cli-skills
**Created**: 2026-03-03

## Overview

Add new CLI commands and skills for comment management, research spikes, and implementation checkpoints. This is **Stream 3** of the 3-stream SDD alignment effort.

**Scope**: Extract `sl comment` CLI from `sl revise`, create `sl-comment` skill, add `spike` and `checkpoint` AI commands.

## Issue Hierarchy

```
SL-3a436b (Epic: New CLI Commands and Skills)
├── SL-d74775 (US1: sl comment list Command) [P1]
│   ├── SL-b31f68 Extract comment types from revise package
│   ├── SL-f6e983 Extract comment client from revise package
│   ├── SL-a801e5 Implement sl comment list subcommand
│   ├── SL-035a5d Update revise imports to use comment package
│   └── SL-c2922e Implement sl comment show subcommand
├── SL-5c4c49 (US2: sl comment show Command) [P1]
│   └── SL-c2922e Implement sl comment show subcommand (shared with US1)
├── SL-90a825 (US3: sl comment reply/resolve Commands) [P1]
│   ├── SL-d2f47b Add CreateReply method to comment client
│   ├── SL-34e30 Implement sl comment reply subcommand
│   └── SL-638e23 Implement sl comment resolve subcommand
├── SL-c94d4d (US4: sl-comment Skill) [P2]
│   └── SL-522810 Create sl-comment skill markdown file
├── SL-edd2da (US5: spike AI Command) [P2]
│   └── SL-145f45 Create spike AI command template
├── SL-e0ef21 (US6: checkpoint AI Command) [P2]
│   └── SL-f21e96 Create checkpoint AI command template
└── Integration
    └── SL-827e41 Update clarify command to use sl comment list
```

## Execution Phases

### Phase 1: US1 - sl comment list Command (P1)

**Feature**: SL-d74775
**Depends on**: None (foundational)

| Task ID | Title | Status | Dependencies |
|---------|-------|--------|--------------|
| SL-b31f68 | Extract comment types from revise package | open | - |
| SL-f6e983 | Extract comment client from revise package | open | SL-b31f68 |
| SL-a801e5 | Implement sl comment list subcommand | open | SL-f6e983 |
| SL-035a5d | Update revise imports to use comment package | open | SL-a801e5 |
| SL-c2922e | Implement sl comment show subcommand | open | SL-a801e5 |

**Independent Test**: `sl comment list --json --status open` returns valid JSON array

### Phase 2: US2 + US3 - Comment Show/Reply/Resolve (P1)

**Features**: SL-5c4c49, SL-90a825
**Depends on**: Phase 1

| Task ID | Title | Status | Dependencies |
|---------|-------|--------|--------------|
| SL-c2922e | Implement sl comment show subcommand | open | Phase 1 |
| SL-d2f47b | Add CreateReply method to comment client | open | SL-f6e983 |
| SL-34e30 | Implement sl comment reply subcommand | open | SL-d2f47b |
| SL-638e23 | Implement sl comment resolve subcommand | open | SL-f6e983 |

**Independent Tests**:
- US2: `sl comment show <id> --json` returns full content with replies
- US3: `sl comment reply <id> "msg"` posts reply successfully

### Phase 3: US4 - sl-comment Skill (P2)

**Feature**: SL-c94d4d
**Depends on**: Phase 1, Phase 2

| Task ID | Title | Status | Dependencies |
|---------|-------|--------|--------------|
| SL-522810 | Create sl-comment skill markdown file | open | SL-c2922e, SL-34e30, SL-638e23 |

**Independent Test**: Skill loads and provides actionable guidance

### Phase 4: US5 + US6 - Spike/Checkpoint Commands (P2)

**Features**: SL-edd2da, SL-e0ef21
**Depends on**: None (parallel with Phase 1)

| Task ID | Title | Status | Dependencies |
|---------|-------|--------|--------------|
| SL-145f45 | Create spike AI command template | open | - |
| SL-f21e96 | Create checkpoint AI command template | open | - |

**Independent Tests**:
- US5: `/specledger.spike "topic"` creates research file
- US6: `/specledger.checkpoint` updates session log

### Phase 5: Integration (P2)

**Depends on**: Phase 1

| Task ID | Title | Status | Dependencies |
|---------|-------|--------|--------------|
| SL-827e41 | Update clarify command to use sl comment list | open | SL-a801e5 |

**Independent Test**: `/specledger.clarify` uses `sl comment list`

**Performance Verification**: Verify all `sl comment` commands meet <2s P95 target (NFR-001)

## Dependency Graph

```
Phase 1 (US1) ── SL-b31f68 ──► SL-f6e983 ──► SL-a801e5 ──┬──► SL-035a5d
                                              │              │
                                              │              └──► SL-827e41 (Integration)
                                              │
                                              ├──► SL-d2f47b ──► SL-34e30 (US3 reply)
                                              │
                                              └──► SL-638e23 (US3 resolve)
                                                        │
                                                        ▼
                               SL-c2922e (US2 show) ──► SL-522810 (US4 skill)

Phase 4 (US5/US6) ── SL-145f45 (spike) [parallel]
                    └── SL-f21e96 (checkpoint) [parallel]
```

## Query Commands

```bash
# View all issues for this spec
sl issue list --label "spec:601-cli-skills"

# View open issues
sl issue list --status open --label "spec:601-cli-skills"

# View by phase
sl issue list --label "phase:us1"
sl issue list --label "phase:us2"
sl issue list --label "phase:us3"

# View ready-to-work (no blockers)
sl issue ready

# View epic tree
sl issue show SL-3a436b
```

## Definition of Done Summary

| Issue ID | DoD Items |
|----------|-----------|
| SL-b31f68 | ReviewComment moved, ThreadReply moved, ReplyMap moved, Revise imports updated, go build passes |
| SL-f6e983 | FetchComments moved, PostgREST chain moved, Auth retry preserved, sl revise still works |
| SL-a801e5 | --json valid JSON, --status open works, --status resolved works, Auth failure exits 1, Compact output 120-char truncation, Counts shown, Footer hint for drill-down, ~500 tokens for 25 comments |
| SL-035a5d | Imports updated in types.go, Imports updated in client.go, sl revise works, go test passes |
| SL-c2922e | Single ID shows comment, Multiple IDs work, --json complete (no truncation), Non-existent ID error, Replies chronological, ~200 tokens for 1 comment + 3 replies |
| SL-d2f47b | POST to correct endpoint, Returns ThreadReply with ID, Auth retry works, 404 parent handled |
| SL-34e30 | Message posted to thread, --json includes reply_id, --json includes timestamp, Parent not found error, Minimal output (~30 tokens) |
| SL-638e23 | Single ID resolves, Multiple IDs work, Cascade to replies, --json shows IDs, Minimal confirmation output |
| SL-522810 | When to use documented, Decision criteria for list vs show, JSON parsing examples, Reply/resolve workflow |
| SL-145f45 | Findings section, Decisions section, Recommendations section, Unique filename (yyyy-mm-dd-<topic>.md) |
| SL-f21e96 | In-progress tasks checked, Tests verified (go test ./... exit 0), Uncommitted changes noted, Summary shows accomplishments |
| SL-827e41 | sl revise --summary replaced, sl comment list used, Reply/resolve instructions updated |

## Success Criteria

- [ ] SC-001: 4 new `sl comment` subcommands available
- [ ] SC-002: `sl comment list --json` produces valid JSON
- [ ] SC-003: `/specledger.clarify` uses `sl comment list`
- [ ] SC-004: `sl-comment` skill provides actionable guidance
- [ ] SC-005: `/specledger.spike` creates research files
- [ ] SC-006: `/specledger.checkpoint` tracks progress

## MVP Scope

**MVP = US1 + US2 + US3** (all P1 stories)

This delivers:
- `sl comment list` - Replace `sl revise --summary`
- `sl comment show` - Full comment details
- `sl comment reply` - Post replies
- `sl comment resolve` - Mark resolved
- `/specledger.clarify` updated to use new CLI

**Post-MVP = US4 + US5 + US6** (P2 stories)

These add:
- `sl-comment` skill for agent guidance
- `/specledger.spike` for research
- `/specledger.checkpoint` for progress tracking

## References

- [spec.md](spec.md) - Feature specification
- [plan.md](plan.md) - Implementation plan
- [research.md](research.md) - Extraction patterns, skill structure
