# Implementation Plan: Session-to-Knowledge Memory Pipeline

**Branch**: `607-session-memory-pipeline` | **Date**: 2026-03-11 | **Spec**: [spec.md](./spec.md)

## Summary

Build a pipeline that extracts structured knowledge from AI session transcripts, scores entries on three axes (Recurrence, Impact, Specificity), promotes high-value entries to a persistent knowledge base, and injects them into agent context at session start.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), Supabase (storage/PostgREST), existing session capture (pkg/cli/session)
**Storage**: Local JSONL + cached markdown (`.specledger/memory/`), Supabase PostgREST for cloud index
**Project Type**: CLI tool with AI command integration
**Target Platform**: Cross-platform (macOS, Linux, Windows)

## Constitution Check

Constitution is uninitialized (template placeholders only). No gates to evaluate.

## Phase 0: Research Summary

### Architecture Decision: Four-Layer Integration

The feature spans all four architecture layers:

| Layer | Component | Role |
|-------|-----------|------|
| L0 (Hooks) | Future SessionStart hook | Auto-inject knowledge at session start (deferred — pending Claude Code support) |
| L1 (CLI) | `sl memory` commands | show, pull, extract, promote, demote, delete, sync |
| L2 (AI Commands) | `/specledger.memory` | AI-powered extraction with summarize, tag, patterns, synthesize modes |
| L3 (Skills) | `sl-memory` skill | Passive injection of cached knowledge into agent context |

### Storage Architecture

**Local-first design** — all operations work offline:

- **Knowledge entries**: `.specledger/memory/cache/entries.jsonl` — one JSON line per entry
- **Promoted knowledge**: `.specledger/memory/knowledge.md` — human-readable markdown, auto-generated from promoted entries
- **Cache directory**: `.specledger/memory/cache/` — gitignored, contains working data

**Cloud sync** (optional):
- **Supabase table**: `knowledge_entries` — indexed entries with project_id, scores, promotion status
- **Pattern**: Same as existing session metadata — PostgREST for queries, local cache for speed

### Scoring System

Three-axis scoring (each 0-10):
- **Recurrence**: How often the pattern appears across sessions (auto-incremented on duplicate detection)
- **Impact**: How much the pattern affects outcomes (AI-assessed during extraction)
- **Specificity**: How project-specific vs. generic the pattern is (AI-assessed)

**Composite score**: Average of three axes. Threshold for auto-promotion: **7.0**

### Key Design Decisions

1. **JSONL for entries** (not markdown): Entries need structured fields (scores, tags, status) that are hard to parse from markdown. JSONL is consistent with issues.jsonl pattern.
2. **Markdown for promoted knowledge**: The agent-facing output is a generated markdown file, easy to inject into context and read.
3. **AI extraction via L2 command**: Extraction requires LLM processing to identify patterns from unstructured transcripts. The `/specledger.memory` command handles this.
4. **L3 skill for injection**: A passive skill file that includes the knowledge.md content, loaded by the agent automatically.
5. **Existing session infrastructure reused**: Sessions are already captured and stored. The pipeline reads from existing `sl session get` data.

### Previous Work

- **010-checkpoint-session-capture**: Built the session capture pipeline (capture.go, storage.go, metadata.go, types.go). Provides the raw transcripts this feature processes.
- **602-silent-session-capture**: Improved capture to work silently. Ensures transcripts are available without user disruption.
- **600-bash-cli-migration (US4)**: Created `sl context update` command. This feature extends context injection to include knowledge.

### Dependencies

- **Issue #51**: Session lifecycle management with tags column. If not available, extraction works without session tags (tags are optional metadata).

## Phase 1: Design

### New Files

| File | Purpose |
|------|---------|
| `pkg/cli/memory/types.go` | KnowledgeEntry struct, scoring types, constants |
| `pkg/cli/memory/store.go` | JSONL read/write, merge/dedup, promotion logic |
| `pkg/cli/memory/scorer.go` | Composite score calculation, threshold checking |
| `pkg/cli/memory/renderer.go` | Generate knowledge.md from promoted entries |
| `pkg/cli/memory/sync.go` | Supabase sync client (push/pull entries) |
| `pkg/cli/commands/memory.go` | CLI commands: show, pull, promote, demote, delete, sync |
| `.claude/commands/specledger.memory.md` | L2 AI command for extraction |
| `.claude/skills/sl-memory/skill.md` | L3 skill for knowledge injection |

### Modified Files

| File | Change |
|------|--------|
| `cmd/sl/main.go` | Register `commands.VarMemoryCmd` |
| `.gitignore` | Add `.specledger/memory/cache/` |

### Data Model

See [data-model.md](./data-model.md) for full entity definitions.

### CLI Command Structure

```
sl memory
├── show          # Display knowledge entries with scores and status
├── pull          # Download knowledge from cloud to local cache
├── promote <id>  # Manually promote an entry to knowledge base
├── demote <id>   # Remove an entry from knowledge base
├── delete <id>   # Delete an entry entirely
└── sync          # Push local promoted entries to cloud
```

### Verification Strategy

- Unit tests for scoring logic (composite calculation, threshold)
- Unit tests for JSONL store (CRUD, merge, dedup)
- Unit tests for markdown renderer (promoted → knowledge.md)
- Integration test: extract → score → promote → render pipeline
- Manual test: inject knowledge into agent context via skill

## Phase 2: Work Breakdown

### Foundational (must complete before any user story)
- **T001**: Create `pkg/cli/memory/` package with types.go (KnowledgeEntry, Score, etc.)
- **T002**: Implement JSONL store (read/write/update/delete entries)
- **T003**: Implement scorer (composite calculation, threshold check)

### US1 + US6 — Knowledge Retrieval (P1)
- **T004**: Implement renderer (generate knowledge.md from promoted entries)
- **T005**: Create `.claude/skills/sl-memory/skill.md` for context injection
- **T006**: Add `.specledger/memory/cache/` to .gitignore

### US2 — Knowledge Extraction (P1)
- **T007**: Create `/specledger.memory` L2 AI command for extraction
- **T008**: Create `sl memory show` CLI command

### US3 — Scoring and Promotion (P2)
- **T009**: Implement merge/dedup logic in store (Recurrence tracking)
- **T010**: Create `sl memory promote` and `sl memory demote` commands

### US4 — View and Manage (P2)
- **T011**: Create `sl memory delete` command
- **T012**: Register VarMemoryCmd in main.go

### US5 — Cloud Sync (P3)
- **T013**: Implement Supabase sync client (push/pull)
- **T014**: Create `sl memory pull` and `sl memory sync` commands
- **T015**: Create Supabase migration for knowledge_entries table

## Success Criteria

- [ ] SC-001: Agent context includes promoted knowledge entries at session start
- [ ] SC-002: `sl memory show` displays entries with scores, tags, and status
- [ ] SC-003: Scoring correctly calculates composite from three axes
- [ ] SC-004: Auto-promotion at threshold 7.0 works
- [ ] SC-005: Knowledge.md is generated from promoted entries
- [ ] SC-006: Cloud sync pushes/pulls entries via Supabase
