# Research: Session-to-Knowledge Memory Pipeline

**Date**: 2026-03-11
**Feature**: 607-session-memory-pipeline

## Prior Work

- **010-checkpoint-session-capture**: Complete session capture pipeline. Sessions are captured on git commit via PostToolUse hook, compressed with gzip, uploaded to Supabase Storage, metadata indexed via PostgREST. Provides `sl session list/get/capture/sync` commands.
- **602-silent-session-capture**: Silent capture improvements — no stderr noise for unauthenticated users, graceful error handling.
- **600-bash-cli-migration (US4)**: `sl context update` command that reads plan.md and injects technology context into CLAUDE.md. This feature extends the same pattern to inject knowledge.

## Findings

### Session Transcript Format

Sessions are stored as `SessionContent` (types.go):
- `Messages[]` with Role (user/assistant), Content (string), Timestamp
- Metadata: session_id, feature_branch, commit_hash, task_id
- Compressed as gzip JSON in Supabase Storage bucket "sessions"

The extraction pipeline reads these messages to identify patterns. Key fields for extraction:
- assistant messages → decisions, explanations, code patterns
- user messages → requirements, corrections, preferences

### Storage Patterns in Codebase

| Feature | Storage | Pattern |
|---------|---------|---------|
| Issues | `specledger/<spec>/issues.jsonl` | JSONL per spec directory |
| Sessions | Supabase Storage + PostgREST | Cloud-first, local queue for offline |
| Agent context | `CLAUDE.md` (root) | Generated markdown, manual additions preserved |

**Decision**: Use JSONL for knowledge entries (consistent with issues pattern). Generate markdown for agent consumption (consistent with context pattern).

### Scoring System Design

**Decision**: Simple average of three axes (Recurrence, Impact, Specificity), each 0-10.

**Rationale**: More complex weighting systems add configuration burden without clear benefit at this stage. The 7.0 threshold is a starting point — can be tuned later.

**Alternatives considered**:
- Weighted average (e.g., Impact × 2) — rejected: adds complexity, unclear which weight is right before real-world data
- Machine learning model — rejected: overkill for MVP, requires training data we don't have yet
- Binary promote/skip — rejected: loses the scoring signal for future analysis

### L2 Command Design

**Decision**: Single `/specledger.memory` command with internal modes (summarize, tag, patterns, synthesize).

**Rationale**: Follows D-MEM-3 from the issue. A single command with modes is simpler to discover than multiple commands.

**Modes**:
- `summarize` — Extract key decisions and patterns from a session
- `tag` — Assign category tags to extracted entries
- `patterns` — Identify recurring patterns across multiple sessions
- `synthesize` — Merge related entries into consolidated knowledge

### L3 Skill Design

**Decision**: `sl-memory` skill that reads `.specledger/memory/knowledge.md` and presents it as agent context.

**Rationale**: Skills are passive — they provide context when loaded. The knowledge file is pre-generated, so no computation needed at skill load time.

### Cloud Sync Architecture

**Decision**: Same pattern as session metadata — Supabase PostgREST for CRUD, local JSONL as cache.

**Table**: `knowledge_entries` with columns:
- id (UUID), project_id, title, description, tags (text[]), source_session_id
- score_recurrence, score_impact, score_specificity, composite_score
- status (candidate|promoted|archived), created_at, updated_at

**Rationale**: Consistent with existing Supabase patterns. PostgREST provides filtering/querying for free.
