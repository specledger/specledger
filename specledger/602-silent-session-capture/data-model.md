# Data Model: Silent Session Capture

## New Entity: Capture Error Log

### session_capture_errors (Supabase table)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK, auto-generated | Unique error record ID |
| user_id | text | NOT NULL, indexed | User who experienced the error |
| project_id | text | NOT NULL, indexed | Project the capture was for |
| session_id | text | nullable | Session ID for correlation with queued sessions |
| error_message | text | NOT NULL | Human-readable error description |
| feature_branch | text | nullable | Git branch at time of error |
| commit_hash | text | nullable | Git commit that triggered the capture |
| retry_count | integer | default 0 | Number of retry attempts (0 = first failure) |
| created_at | timestamptz | default now() | When the error occurred |

**RLS**: Users can read/write own errors. Project admins can read all project errors.

### Local Error Log File

**Path**: `~/.specledger/capture-errors.log`
**Format**: JSONL (one JSON object per line, append-only)

Each line:
```json
{"timestamp":"2026-03-04T10:00:00Z","user_id":"uuid","project_id":"uuid","session_id":"uuid","error_message":"...","feature_branch":"...","commit_hash":"...","retry_count":0}
```

## Existing Entities (unchanged)

### Credentials (`~/.specledger/credentials.json`)
- Used for auth check (read-only in this feature)
- Fields: access_token, refresh_token, expires_in, created_at, user_email, user_id

### Session Queue (`~/.specledger/session-queue/`)
- Existing queue mechanism for failed uploads
- QueueEntry contains: session_id, project_id, feature_branch, commit_hash, author_id, status, retry_count
- No changes needed to queue structure

### specledger.yaml (`specledger/specledger.yaml`)
- Contains project.id field
- Used for project ID lookup (read-only in this feature)
