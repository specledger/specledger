# Research: Checkpoint Session Capture

**Feature**: 010-checkpoint-session-capture
**Date**: 2026-02-12
**Status**: Complete

## Prior Work

### Related Features
- **008-cli-auth**: Browser-based OAuth + credential storage at `~/.specledger/credentials.json`. Established Supabase constants (`SupabaseURL`, `SupabaseAnonKey`), `GetValidAccessToken()` pattern with auto-refresh, and HTTP client patterns using `net/http`.
- **009-add-login-and-comment-commands**: Supabase REST API integration for comments. Established slash-command auth via token paste with `~/.specledger/session.json`.
- **Beads Issue Tracker**: Provides `.beads/issues.jsonl` task execution framework. Relevant for per-task session capture (US-3).

### Task Overlap
- No existing beads tasks overlap with this feature's scope.
- Auth infrastructure (token management, Supabase client) can be reused directly.

## Research Findings

### R1: How to Capture Claude Code Conversation Content

**Decision**: Use Claude Code hooks (`Stop` event) to trigger session capture, reading the conversation from `transcript_path`.

**Rationale**: Claude Code provides a hooks system where shell commands execute on lifecycle events. The `Stop` hook fires after each Claude response and provides `transcript_path` (the JSONL file containing the full conversation) and `session_id` via JSON on stdin. This is the only reliable programmatic mechanism to access conversation data.

**Key Details**:
- **Hook Events Available**: `SessionStart`, `UserPromptSubmit`, `PreToolUse`, `PostToolUse`, `Stop`, `SessionEnd`, `PreCompact`, etc.
- **Hook Input (JSON via stdin)**: `session_id`, `transcript_path`, `cwd`, `hook_event_name`
- **Conversation Format**: JSONL where each line is `{"type":"user"|"assistant", "message":{"role":"...", "content":"..."}, "timestamp":"ISO8601"}`
- **No native git commit hooks exist** in Claude Code (feature request open: #4834)
- The hook script will be a Go binary (`sl session capture`) or shell script that reads stdin JSON, extracts `transcript_path`, and processes the delta

**Alternatives Considered**:
1. **Git post-commit hook**: Cannot access Claude Code session state. Would need a separate mechanism to know which transcript to capture.
2. **Manual export/upload**: Poor UX, violates FR-006 (async, automatic capture).
3. **File watcher on `~/.claude/projects/`**: Fragile, race conditions, no clear boundary between commits.

**Capture Strategy**:
- Use `PostToolUse` hook with `Bash` matcher to detect `git commit` commands
- When a commit is detected, read the `transcript_path` JSONL, compute the delta since the last capture checkpoint, upload to Supabase Storage, and record metadata in the database
- Store a local marker (last captured line number or offset) to compute deltas efficiently

### R2: Cloud Storage for Sessions

**Decision**: Use Supabase Storage (S3-compatible) via REST API with `net/http`.

**Rationale**: The project already uses Supabase for auth and database. Supabase Storage provides S3-compatible object storage accessible via REST API. Using the REST API directly (not a Go SDK) maintains consistency with the existing HTTP client pattern in `pkg/cli/auth/client.go`.

**Key Details**:
- **Supabase Storage REST API**: `{SUPABASE_URL}/storage/v1/object/{bucket}/{path}`
- **Upload**: `POST /storage/v1/object/{bucket}/{path}` with Bearer token and file body
- **Download**: `GET /storage/v1/object/{bucket}/{path}` with Bearer token
- **Signed URLs**: `POST /storage/v1/object/sign/{bucket}/{path}` for time-limited access
- **Bucket**: `sessions` (created via Supabase dashboard or migration)
- **Key pattern**: `{project_id}/{feature_branch}/{commit_hash}.json.gz`

**Alternatives Considered**:
1. **supabase-community/storage-go SDK**: Adds dependency, project uses raw `net/http` for all Supabase calls. SDK pattern would diverge.
2. **Direct S3 SDK (aws-sdk-go-v2)**: Over-engineered, adds heavy dependency for a single bucket.
3. **Local file storage with sync**: Defeats cloud accessibility requirement (FR-012).

### R3: Session Delta Computation

**Decision**: Track last-captured line offset per session; compute delta as new JSONL lines since last capture.

**Rationale**: Claude Code conversations are append-only JSONL files. Each commit checkpoint captures only new lines since the last checkpoint. A simple line-offset tracker (stored locally at `~/.specledger/session-state.json`) provides efficient delta computation without parsing the entire file.

**Key Details**:
- Store `{session_id: {last_offset: N, last_commit: "hash"}}` locally
- On capture: read lines from offset N to EOF, that's the delta
- On first capture (no prior offset): capture entire conversation up to that point
- On amend: new commit hash → new capture (old session retained per spec edge case)

**Alternatives Considered**:
1. **Timestamp-based windowing**: Unreliable if timestamps drift or messages lack timestamps.
2. **Content hashing / diffing**: Over-complex for append-only JSONL.
3. **Full conversation per commit**: Wasteful for large sessions; violates "segment/delta" clarification.

### R4: Supabase Database Schema for Session Metadata

**Decision**: Create a `sessions` table in Supabase with RLS policies scoped to project membership.

**Rationale**: Session metadata (not content) lives in the database for fast querying. Content lives in Storage. This separation keeps the database lightweight while supporting FR-011 (query by feature) and FR-013 (query by commit, task, author, date range).

**Key Details**:
- Table: `sessions` with columns: `id`, `project_id`, `feature_branch`, `commit_hash`, `task_id`, `author_id`, `storage_path`, `status`, `size_bytes`, `message_count`, `created_at`
- RLS: `SELECT` allowed if user is a member of the project
- Index on: `(project_id, feature_branch)`, `(project_id, commit_hash)`, `(project_id, task_id)`
- Content stored at: `sessions/{project_id}/{feature_branch}/{commit_hash}.json.gz`

### R5: Offline Caching & Retry

**Decision**: Queue failed uploads in `~/.specledger/session-queue/` with exponential backoff retry.

**Rationale**: Session capture must not block commits (FR-006). Failed uploads are cached locally as files and retried on subsequent CLI invocations. This pattern is simple, requires no background daemon, and handles temporary network issues.

**Key Details**:
- On upload failure: save compressed session to `~/.specledger/session-queue/{uuid}.json.gz` + `{uuid}.meta.json`
- On any subsequent `sl` command (or dedicated `sl session sync`): attempt to upload queued sessions
- Retry with exponential backoff: 1s, 2s, 4s (3 attempts per invocation)
- After persistent failure (>3 invocations): warn user, keep queued

### R6: Compression

**Decision**: gzip compression for session content before upload.

**Rationale**: JSON conversation data compresses 70-80% with gzip. A 10MB session becomes ~2-3MB. Go's `compress/gzip` is stdlib — no dependency needed. Supabase Storage handles any content type.

**Key Details**:
- Compress with `compress/gzip` before upload
- Store as `.json.gz` in storage
- Decompress on retrieval
- Content-Type: `application/gzip`
