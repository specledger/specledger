# Data Model: Checkpoint Session Capture

**Feature**: 010-checkpoint-session-capture
**Date**: 2026-02-12

## Entities

### Session

A captured AI conversation segment linked to a checkpoint (commit) or task.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, auto-generated | Unique session identifier |
| `project_id` | UUID | FK → projects.id, NOT NULL | Project this session belongs to |
| `feature_branch` | TEXT | NOT NULL | Feature branch name (e.g., `010-checkpoint-session-capture`) |
| `commit_hash` | TEXT | nullable | Git commit hash (for checkpoint sessions) |
| `task_id` | TEXT | nullable | Beads task ID (for task sessions, e.g., `SL-42`) |
| `author_id` | UUID | FK → auth.users.id, NOT NULL | User who created the session |
| `storage_path` | TEXT | NOT NULL, UNIQUE | Path in Supabase Storage bucket |
| `status` | TEXT | NOT NULL, CHECK(IN 'complete','rejected','abandoned') | Session completion status |
| `size_bytes` | BIGINT | NOT NULL | Compressed size in bytes |
| `raw_size_bytes` | BIGINT | NOT NULL | Uncompressed size in bytes |
| `message_count` | INTEGER | NOT NULL | Number of messages in the session segment |
| `created_at` | TIMESTAMPTZ | NOT NULL, DEFAULT now() | When the session was captured |

**Constraints**:
- CHECK: at least one of `commit_hash` or `task_id` must be non-null
- UNIQUE: `(project_id, commit_hash)` where `commit_hash IS NOT NULL`
- Size limit: `raw_size_bytes <= 10485760` (10 MB)

**Indexes**:
- `idx_sessions_project_feature` ON `(project_id, feature_branch)`
- `idx_sessions_project_commit` ON `(project_id, commit_hash)` WHERE `commit_hash IS NOT NULL`
- `idx_sessions_project_task` ON `(project_id, task_id)` WHERE `task_id IS NOT NULL`
- `idx_sessions_author` ON `(author_id, created_at DESC)`

### Session Content (Supabase Storage Object)

The actual conversation data stored in Supabase Storage as gzip-compressed JSON.

**Storage Path Pattern**: `{project_id}/{feature_branch}/{commit_hash_or_task_id}.json.gz`

**Content Schema** (JSON):
```json
{
  "version": "1.0",
  "session_id": "uuid",
  "feature_branch": "010-checkpoint-session-capture",
  "commit_hash": "abc123...",
  "task_id": null,
  "author": "user@example.com",
  "captured_at": "2026-02-12T10:30:00Z",
  "messages": [
    {
      "role": "user",
      "content": "Please implement the session capture feature",
      "timestamp": "2026-02-12T10:15:00Z"
    },
    {
      "role": "assistant",
      "content": "I'll start by creating the data model...",
      "timestamp": "2026-02-12T10:15:05Z"
    }
  ]
}
```

### Session State (Local)

Local tracking state for delta computation. Stored at `~/.specledger/session-state.json`.

```json
{
  "sessions": {
    "<session_id>": {
      "last_offset": 1542,
      "last_commit": "abc123def456...",
      "transcript_path": "/Users/.../.claude/projects/.../session.jsonl"
    }
  }
}
```

### Session Queue (Local)

Failed uploads queued for retry. Stored at `~/.specledger/session-queue/`.

**Queue Entry**: `{uuid}.json.gz` (compressed session content) + `{uuid}.meta.json`

**Meta Schema**:
```json
{
  "session_id": "uuid",
  "project_id": "uuid",
  "feature_branch": "010-checkpoint-session-capture",
  "commit_hash": "abc123...",
  "task_id": null,
  "author_id": "uuid",
  "status": "complete",
  "created_at": "2026-02-12T10:30:00Z",
  "retry_count": 2,
  "last_retry": "2026-02-12T10:31:00Z"
}
```

## Relationships

```
┌──────────────┐       ┌──────────────────┐
│   projects   │──1:N──│    sessions      │
│              │       │ (metadata table)  │
└──────────────┘       └────────┬─────────┘
                                │
                                │ storage_path
                                ▼
                       ┌──────────────────┐
                       │ Supabase Storage │
                       │  sessions bucket │
                       │  (content blobs) │
                       └──────────────────┘

┌──────────────┐
│  auth.users  │──1:N──→ sessions.author_id
└──────────────┘
```

## State Transitions

### Session Status
```
[capture triggered] → complete    (normal commit/task completion)
[capture triggered] → rejected    (user rejected task changes)
[capture triggered] → abandoned   (task abandoned mid-session)
```

Session status is immutable after creation — no transitions between states.

## Validation Rules

1. `commit_hash` must be a valid 40-character hex string when provided
2. `task_id` must match the beads ID pattern (e.g., `SL-xxx`) when provided
3. `feature_branch` must match the current git branch at capture time
4. `size_bytes` must be > 0
5. `message_count` must be > 0 (empty sessions are not stored)
6. `raw_size_bytes` must be <= 10,485,760 (10 MB)

## Row-Level Security (RLS)

```sql
-- Sessions are readable by any authenticated member of the project
CREATE POLICY "project_members_can_read_sessions"
  ON sessions FOR SELECT
  USING (
    project_id IN (
      SELECT project_id FROM project_members
      WHERE user_id = auth.uid()
    )
  );

-- Sessions are insertable by authenticated users who are project members
CREATE POLICY "project_members_can_insert_sessions"
  ON sessions FOR INSERT
  WITH CHECK (
    author_id = auth.uid()
    AND project_id IN (
      SELECT project_id FROM project_members
      WHERE user_id = auth.uid()
    )
  );
```

## Supabase Storage Policies

```sql
-- Bucket: sessions (private)

-- Upload: authenticated project members
CREATE POLICY "project_members_can_upload_sessions"
  ON storage.objects FOR INSERT
  WITH CHECK (
    bucket_id = 'sessions'
    AND auth.role() = 'authenticated'
    AND (storage.foldername(name))[1] IN (
      SELECT project_id::text FROM project_members
      WHERE user_id = auth.uid()
    )
  );

-- Download: authenticated project members
CREATE POLICY "project_members_can_download_sessions"
  ON storage.objects FOR SELECT
  USING (
    bucket_id = 'sessions'
    AND auth.role() = 'authenticated'
    AND (storage.foldername(name))[1] IN (
      SELECT project_id::text FROM project_members
      WHERE user_id = auth.uid()
    )
  );
```
