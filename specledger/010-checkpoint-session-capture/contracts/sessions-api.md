# Session API Contracts

**Feature**: 010-checkpoint-session-capture
**Date**: 2026-02-12
**Protocol**: Supabase REST (PostgREST) + Supabase Storage REST API

All endpoints require `Authorization: Bearer {access_token}` and `apikey: {anon_key}` headers.

## Session Metadata (PostgREST)

Base URL: `{SUPABASE_URL}/rest/v1/sessions`

### Create Session Metadata

```
POST /rest/v1/sessions
Content-Type: application/json
Authorization: Bearer {access_token}
apikey: {anon_key}
Prefer: return=representation

{
  "project_id": "uuid",
  "feature_branch": "010-checkpoint-session-capture",
  "commit_hash": "abc123def456...",
  "task_id": null,
  "author_id": "uuid",
  "storage_path": "{project_id}/010-checkpoint-session-capture/abc123def456.json.gz",
  "status": "complete",
  "size_bytes": 245760,
  "raw_size_bytes": 1048576,
  "message_count": 42
}

Response 201:
{
  "id": "uuid",
  "project_id": "uuid",
  "feature_branch": "010-checkpoint-session-capture",
  "commit_hash": "abc123def456...",
  "task_id": null,
  "author_id": "uuid",
  "storage_path": "...",
  "status": "complete",
  "size_bytes": 245760,
  "raw_size_bytes": 1048576,
  "message_count": 42,
  "created_at": "2026-02-12T10:30:00Z"
}
```

### Query Sessions by Feature

```
GET /rest/v1/sessions?project_id=eq.{project_id}&feature_branch=eq.{branch}&order=created_at.desc
Authorization: Bearer {access_token}
apikey: {anon_key}

Response 200:
[
  {
    "id": "uuid",
    "feature_branch": "010-checkpoint-session-capture",
    "commit_hash": "abc123...",
    "task_id": null,
    "status": "complete",
    "message_count": 42,
    "size_bytes": 245760,
    "created_at": "2026-02-12T10:30:00Z"
  }
]
```

### Query Session by Commit Hash

```
GET /rest/v1/sessions?project_id=eq.{project_id}&commit_hash=eq.{hash}&select=*
Authorization: Bearer {access_token}
apikey: {anon_key}

Response 200:
[{ ...session metadata... }]
```

### Query Session by Task ID

```
GET /rest/v1/sessions?project_id=eq.{project_id}&task_id=eq.{task_id}&select=*
Authorization: Bearer {access_token}
apikey: {anon_key}

Response 200:
[{ ...session metadata... }]
```

### Query Sessions by Author and Date Range

```
GET /rest/v1/sessions?project_id=eq.{project_id}&author_id=eq.{uid}&created_at=gte.{start}&created_at=lte.{end}&order=created_at.desc
Authorization: Bearer {access_token}
apikey: {anon_key}

Response 200:
[{ ...session metadata array... }]
```

## Session Content (Supabase Storage)

Base URL: `{SUPABASE_URL}/storage/v1`
Bucket: `sessions`

### Upload Session Content

```
POST /storage/v1/object/sessions/{project_id}/{feature_branch}/{identifier}.json.gz
Content-Type: application/gzip
Authorization: Bearer {access_token}
apikey: {anon_key}

Body: <gzip-compressed JSON bytes>

Response 200:
{
  "Key": "sessions/{project_id}/{feature_branch}/{identifier}.json.gz",
  "Id": "uuid"
}
```

### Download Session Content

```
GET /storage/v1/object/sessions/{project_id}/{feature_branch}/{identifier}.json.gz
Authorization: Bearer {access_token}
apikey: {anon_key}

Response 200:
Content-Type: application/gzip
Body: <gzip-compressed JSON bytes>
```

### Generate Signed URL (Time-Limited Access)

```
POST /storage/v1/object/sign/sessions/{project_id}/{feature_branch}/{identifier}.json.gz
Content-Type: application/json
Authorization: Bearer {access_token}
apikey: {anon_key}

{
  "expiresIn": 3600
}

Response 200:
{
  "signedURL": "https://...supabase.co/storage/v1/object/sign/sessions/...?token=..."
}
```

## CLI Commands (Go/Cobra)

### sl session capture

Triggered by Claude Code hook. Reads hook JSON from stdin.

```
stdin: {"session_id":"...","transcript_path":"/path/to/session.jsonl","cwd":"/path/to/project","hook_event_name":"PostToolUse"}

Exit codes:
  0 - Success (session captured and uploaded or queued)
  0 - No active session or no delta (silently skips)
  1 - Fatal error (logged but does not block commit)
```

### sl session list

List sessions for the current feature branch.

```
sl session list [--feature <branch>] [--commit <hash>] [--task <id>] [--json]

Output (default):
  COMMIT     MESSAGES  SIZE     STATUS    CAPTURED
  abc123...  42        1.0 MB   complete  2026-02-12 10:30:00

Output (--json):
  [{ ...session metadata... }]
```

### sl session get

Retrieve a specific session's content.

```
sl session get <session-id|commit-hash|task-id> [--json] [--raw]

Output (default): formatted conversation
Output (--json): raw session JSON
Output (--raw): raw gzip stream to stdout
```

### sl session sync

Upload any locally queued sessions.

```
sl session sync [--json]

Output:
  Uploaded 2 queued sessions
  1 session still queued (network error)
```

## Error Responses

All Supabase REST errors follow this format:
```json
{
  "statusCode": 401,
  "error": "Unauthorized",
  "message": "JWT expired"
}
```

Storage errors:
```json
{
  "statusCode": 400,
  "error": "Duplicate",
  "message": "The resource already exists"
}
```
