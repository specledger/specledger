# CLI Contract: `sl comment`

**Pattern**: Data CRUD (D16) | **Layer**: L1 (CLI)

## Commands

### `sl comment list`

List review comments for the current spec.

```
sl comment list [--status open|resolved|all] [--spec <spec-key>] [--json]
```

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--status` | string | `open` | Filter by resolution status |
| `--spec` | string | auto-detect | Override spec key |
| `--json` | bool | false | JSON output |

**Human Output** (default):
```
Review Comments for 598-sdd-workflow-streamline (3 open)

specs/598-sdd-workflow-streamline/spec.md
  #1 [df77879f] (Son Vo, 2026-02-10)
     "Lưu raw HTML/PDF của văn bản gốc vào Supabase Storage"
     → cái này có cần thiết không

  #2 [f846962e] (so0k, 2026-02-17)
     "bd run-ready"
     → non existent command
     └── Reply: "my bad, noticed this..." (so0k, 2026-02-21)

specs/598-sdd-workflow-streamline/tasks.md
  #3 [54181d3b] (TrungDinh, 2026-02-10)
     "Type"
     → need clarify
```

**JSON Output** (`--json`):
```json
{
  "spec_key": "598-sdd-workflow-streamline",
  "change_id": "f364fa2c-...",
  "comments": [
    {
      "parent": {
        "id": "df77879f-...",
        "file_path": "specs/.../spec.md",
        "content": "cái này có cần thiết không",
        "selected_text": "Lưu raw HTML/PDF...",
        "is_resolved": false,
        "author_name": "Son Vo",
        "created_at": "2026-02-10T03:41:27Z"
      },
      "replies": []
    }
  ],
  "total_count": 3
}
```

**Exit Codes**: 0 = success, 1 = error (auth, network, no spec)

---

### `sl comment show <id>`

Show full details of a single comment with thread context.

```
sl comment show <comment-id> [--json]
```

**Human Output**:
```
Comment df77879f (open)
File:     specs/598-sdd-workflow-streamline/spec.md
Author:   Son Vo <vophuochoangson@gmail.com>
Created:  2026-02-10 03:41:27 UTC
Selected: "Lưu raw HTML/PDF của văn bản gốc vào Supabase Storage"

cái này có cần thiết không

Thread (2 replies):
  └── Ngoc Tran (2026-02-18): "this is great"
  └── Ngoc Tran (2026-02-18): "I think we need to adjust more"
```

**Exit Codes**: 0 = success, 1 = not found or error

---

### `sl comment reply <id> --content "text"`

Post a reply to a comment.

```
sl comment reply <comment-id> --content "Addressed in commit abc123" [--json]
```

**Flags**:
| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--content` | string | yes | Reply text |

**Behavior**:
- Creates new `review_comment` with `parent_comment_id` set to target
- Uses authenticated user's name/email from credentials
- `change_id` inherited from parent comment

**JSON Output**: `{"id": "<new-reply-id>", "parent_comment_id": "<parent-id>"}`

**Exit Codes**: 0 = success, 1 = error

---

### `sl comment resolve <id> [--reason "text"]`

Mark a comment as resolved, optionally with a reason.

```
sl comment resolve <comment-id> [--reason "Fixed in PR #42"] [--json]
```

**Flags**:
| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--reason` | string | no | Resolution reason (posted as reply before resolving) |

**Behavior**:
1. If `--reason` provided: post reply with reason content first
2. PATCH comment `is_resolved=true`
3. Cascade: also resolve all replies (PATCH with `id=in.(...)`)

**Exit Codes**: 0 = success, 1 = error

---

## Auth Requirements

All commands require valid Supabase credentials (`~/.specledger/credentials.json`).

**Headers**: `Authorization: Bearer <access_token>` + `apikey: <anon_key>`

**Token refresh**: Automatic on 401/PGRST303 via `auth.ForceRefreshAccessToken()`.

## PostgREST Endpoints Used

| Operation | Method | Endpoint |
|-----------|--------|----------|
| List comments | GET | `/rest/v1/review_comments?change_id=eq.{cid}&is_resolved=eq.{status}&parent_comment_id=is.null` |
| List replies | GET | `/rest/v1/review_comments?change_id=eq.{cid}&parent_comment_id=not.is.null` |
| Show comment | GET | `/rest/v1/review_comments?id=eq.{id}` |
| Create reply | POST | `/rest/v1/review_comments` |
| Resolve comment | PATCH | `/rest/v1/review_comments?id=eq.{id}` body: `{"is_resolved": true}` |
| Cascade resolve | PATCH | `/rest/v1/review_comments?id=in.({ids})` body: `{"is_resolved": true}` |

## Resolution: Project/Spec/Change Chain

Before any comment operation, the client must resolve the query chain:
1. Detect repo owner/name from git remote
2. `GET /projects?repo_owner=eq.{o}&repo_name=eq.{n}` → project_id
3. Detect spec_key from branch (via ContextDetector)
4. `GET /specs?project_id=eq.{pid}&spec_key=eq.{key}` → spec_id
5. `GET /changes?spec_id=eq.{sid}` → change_id (latest open change)
6. Use change_id for all comment queries
