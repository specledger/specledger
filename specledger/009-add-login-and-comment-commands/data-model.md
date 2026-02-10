# Data Model: CLI Authentication and Comment Management

**Feature Branch**: `009-add-login-and-comment-commands`
**Created**: 2026-02-10

## Entities

### Session (Local File)

Represents the authenticated user's session stored on disk.

**Storage**: `~/.specledger/session.json`
**Permissions**: `600` (owner read/write only)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| access_token | string | Yes | JWT access token for API authorization |
| refresh_token | string | Yes | JWT refresh token for obtaining new access tokens |
| expires_at | integer | Yes | Unix timestamp (seconds) when access_token expires |
| user_id | string | Yes | UUID of the authenticated user |

**Validation Rules**:
- All four fields must be present and non-empty
- `expires_at` must be a valid Unix timestamp
- `access_token` must be a valid JWT format (not verified locally)

**State Transitions**:
```
[No Session] --login--> [Active Session]
[Active Session] --logout--> [No Session]
[Active Session] --expired--> [Expired Session]
[Expired Session] --refresh--> [Active Session]
```

**Example**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "v1.MRjvlQ9Mz...",
  "expires_at": 1712345678,
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### Comment (Supabase)

Represents feedback on a specification file, stored in Supabase.

**Storage**: Supabase `comments` table
**Access**: RLS policies based on JWT

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | uuid | Yes | Primary key, auto-generated |
| file_path | string | Yes | Path to the spec file (e.g., "auth.md") |
| content | text | Yes | The comment text |
| type | enum | Yes | Comment type: "critical", "suggestion" |
| status | enum | Yes | Status: "open", "resolved" |
| author_id | uuid | Yes | UUID of comment author |
| author_name | string | Yes | Display name of author |
| created_at | timestamp | Yes | When comment was created |
| resolved_by | uuid | No | UUID of user who resolved (if resolved) |
| resolved_at | timestamp | No | When comment was resolved (if resolved) |
| project_id | uuid | Yes | Project/repo identifier for filtering |

**Validation Rules**:
- `type` must be one of: "critical", "suggestion"
- `status` must be one of: "open", "resolved"
- `resolved_by` and `resolved_at` must both be set or both null

**State Transitions**:
```
[Created] --auto--> status: "open"
[Open] --resolve--> status: "resolved", resolved_by: <user>, resolved_at: <now>
```

**Display Mapping**:
| Type | Icon | Description |
|------|------|-------------|
| critical | ❗ | Blocking issue that must be addressed |
| suggestion | 💡 | Non-blocking improvement idea |
| (resolved) | ✅ | Any resolved comment |

---

## Relationships

```
┌──────────────────────────────────────────────────────────┐
│                      User Session                        │
│  ~/.specledger/session.json                              │
│  ┌─────────────────────────────────────────────────────┐│
│  │ access_token  → Authorization: Bearer <token>       ││
│  │ user_id       → comments.resolved_by (on resolve)   ││
│  └─────────────────────────────────────────────────────┘│
└───────────────────────────┬──────────────────────────────┘
                            │ Authenticates
                            ▼
┌──────────────────────────────────────────────────────────┐
│                   Supabase Backend                        │
│  ┌─────────────────────────────────────────────────────┐│
│  │ comments table                                       ││
│  │ - RLS: Read if in same project                       ││
│  │ - RLS: Update status if authenticated user           ││
│  └─────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────┘
```

---

## API Integration Points

### Read Comments
```
GET /rest/v1/comments
  ?select=*
  &file_path=eq.<file>
  &status=eq.open (optional)

Headers:
  apikey: <SUPABASE_ANON_KEY>
  Authorization: Bearer <access_token>
```

### Update Comment Status
```
PATCH /rest/v1/comments
  ?id=eq.<comment_id>

Headers:
  apikey: <SUPABASE_ANON_KEY>
  Authorization: Bearer <access_token>
  Content-Type: application/json

Body:
  {"status": "resolved", "resolved_by": "<user_id>"}
```

---

## Security Considerations

1. **Session File Permissions**: Must be `600` to prevent other users from reading tokens
2. **Token Storage**: Tokens stored in plaintext (no encryption) - acceptable per design doc
3. **RLS Policies**: All data access controlled by Supabase RLS based on JWT claims
4. **No Secret Storage**: Only public anon key used; actual authorization via JWT
