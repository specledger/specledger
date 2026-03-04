# Contract: Session Capture Error Logging API

## POST /rest/v1/session_capture_errors

Log a session capture error to Supabase.

### Request

```
POST https://<supabase-url>/rest/v1/session_capture_errors
Content-Type: application/json
Authorization: Bearer <access_token>
apikey: <anon_key>
Prefer: return=representation
```

**Body**:
```json
{
  "user_id": "string (required)",
  "project_id": "string (required)",
  "session_id": "string (nullable)",
  "error_message": "string (required)",
  "feature_branch": "string (nullable)",
  "commit_hash": "string (nullable)",
  "retry_count": 0
}
```

### Response

**201 Created**:
```json
[{
  "id": "uuid",
  "user_id": "...",
  "project_id": "...",
  "session_id": "...",
  "error_message": "...",
  "feature_branch": "...",
  "commit_hash": "...",
  "retry_count": 0,
  "created_at": "2026-03-04T10:00:00Z"
}]
```

**401 Unauthorized**: Token expired → refresh and retry once.

### Error Handling

- If POST fails: fall back to local log only. Never block the workflow.
- If 401: attempt token refresh via `auth.ForceRefreshAccessToken()`, retry once.
- All other errors: log locally, continue.

## GET /rest/v1/session_capture_errors (for troubleshooting)

Query error logs by user or project.

```
GET /rest/v1/session_capture_errors?user_id=eq.<user_id>&order=created_at.desc&limit=50
GET /rest/v1/session_capture_errors?project_id=eq.<project_id>&order=created_at.desc&limit=50
```
