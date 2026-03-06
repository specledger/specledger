# Data Model: Silent Session Capture

## Local Error Log File

**Path**: `~/.specledger/capture-errors.log`
**Format**: JSONL (one JSON object per line, append-only)

Each line:
```json
{"timestamp":"2026-03-04T10:00:00Z","user_id":"uuid","project_id":"uuid","session_id":"uuid","error_message":"...","feature_branch":"...","commit_hash":"...","retry_count":0}
```

## Sentry Error Events (remote)

Errors are sent to Sentry via `sentry-go` SDK. No Supabase table needed.

**Sentry context per event**:

| Field | Sentry Mapping | Description |
|-------|---------------|-------------|
| user_id | `User.ID` | User who experienced the error |
| project_id | Tag: `project_id` | Project the capture was for |
| session_id | Tag: `session_id` | Session ID for correlation |
| error_message | Exception message | Human-readable error description |
| feature_branch | Tag: `branch` | Git branch at time of error |
| commit_hash | Tag: `commit_hash` | Git commit that triggered the capture |
| retry_count | Extra: `retry_count` | Number of retry attempts (0 = first failure) |

**Why Sentry over Supabase table**: Sentry provides aggregation, deduplication, alerting, and trend analysis out of the box. A Supabase table would require building all of this manually and adds storage load to the production database.

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
