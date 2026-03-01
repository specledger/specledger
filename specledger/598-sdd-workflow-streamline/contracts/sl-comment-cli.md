# CLI Contract: `sl comment`

**Pattern**: Data CRUD (D16) | **Layer**: L1 (CLI)

## Design Principle: Token-Efficient Output

All `sl` commands consumed by AI agents MUST minimize context window usage:

1. **Default output is compact** — list commands produce scannable overviews, not full dumps
2. **Preview fields are truncated** — `content_preview` (120 chars), `selected_text_preview` (80 chars)
3. **Counts replace nested data** — `reply_count: 2` instead of full `replies[]` array
4. **Detail on demand** — `sl comment show <id>` provides full content for specific comments
5. **Footer hints** — every truncated output includes a follow-up instruction for the agent

This pattern applies to ALL `sl` commands, not just `sl comment`.

---

## Commands

### `sl comment list`

List review comments for the current spec. **Default output is a compact overview.**

```
sl comment list [--status open|resolved|all] [--spec <spec-key>] [--json]
```

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--status` | string | `open` | Filter by resolution status |
| `--spec` | string | auto-detect | Override spec key |
| `--json` | bool | false | JSON output (compact by default) |

**Human Output** (default) — compact table, space-separated:
```
ID        FILE                 AUTHOR     SELECTED                 FEEDBACK                         REPLIES
df77879f  spec.md              Son Vo     "Lưu raw HTML/PDF..."    cái này có cần thiết không       0
f846962e  spec.md              so0k       "bd run-ready"           non existent command             1
54181d3b  tasks.md             TrungDinh  "Type"                   need clarify                     0

3 open comment(s) across 2 artifact(s) for 598-sdd-workflow-streamline
→ Use 'sl comment show <id>' for full details and thread context
```

**JSON Output** (`--json`) — **compact: previews + counts, no nested threads**:
```json
{
  "spec_key": "598-sdd-workflow-streamline",
  "change_id": "f364fa2c-...",
  "comments": [
    {
      "id": "df77879f-...",
      "file_path": "spec.md",
      "content_preview": "cái này có cần thiết không",
      "selected_text_preview": "Lưu raw HTML/PDF...",
      "author_name": "Son Vo",
      "reply_count": 0,
      "is_resolved": false,
      "created_at": "2026-02-10T03:41:27Z"
    },
    {
      "id": "f846962e-...",
      "file_path": "spec.md",
      "content_preview": "non existent command",
      "selected_text_preview": "bd run-ready",
      "author_name": "so0k",
      "reply_count": 1,
      "is_resolved": false,
      "created_at": "2026-02-17T12:36:24Z"
    }
  ],
  "total_count": 3,
  "hint": "Use 'sl comment show <id> --json' for full content and thread replies"
}
```

**Truncation rules**:
- `content_preview`: first 120 chars, append `...` if truncated
- `selected_text_preview`: first 80 chars, append `...` if truncated
- Newlines replaced with spaces in both preview fields

**Token budget**: ~20 tokens per comment (ID + file + author + previews + count). A spec with 25 comments ≈ 500 tokens for the full list. Compare: verbose dump of same 25 comments with full threads ≈ 3000-5000 tokens.

**Exit Codes**: 0 = success, 1 = error (auth, network, no spec)

---

### `sl comment show <id> [<id2> ...]`

Show **full details** of one or more comments with complete thread context. This is the "drill down" command — agents call it after scanning the `list` overview.

```
sl comment show <comment-id> [<comment-id-2> ...] [--json]
```

**Accepts multiple IDs** — agents can batch-fetch related comments in a single call.

**Human Output** (single comment):
```
Comment df77879f (open)
File:     spec.md
Author:   Son Vo <vophuochoangson@gmail.com>
Created:  2026-02-10 03:41:27 UTC
Selected: "Lưu raw HTML/PDF của văn bản gốc vào Supabase Storage"

cái này có cần thiết không

Thread (2 replies):
  └── Ngoc Tran (2026-02-18): "this is great"
  └── Ngoc Tran (2026-02-18): "I think we need to adjust more"

→ Use 'sl comment reply df77879f --content "..."' to respond
→ Use 'sl comment resolve df77879f --reason "..."' to close
```

**JSON Output** (`--json`) — **full content, full thread**:
```json
{
  "comments": [
    {
      "id": "df77879f-...",
      "file_path": "spec.md",
      "content": "cái này có cần thiết không",
      "selected_text": "Lưu raw HTML/PDF của văn bản gốc vào Supabase Storage",
      "line": null,
      "start_line": null,
      "is_resolved": false,
      "author_name": "Son Vo",
      "author_email": "vophuochoangson@gmail.com",
      "is_ai_generated": null,
      "created_at": "2026-02-10T03:41:27Z",
      "replies": [
        {
          "id": "2878bfff-...",
          "content": "this is great",
          "author_name": "Ngoc Tran",
          "created_at": "2026-02-18T23:16:58Z"
        },
        {
          "id": "04837828-...",
          "content": "I think we need to adjust more",
          "author_name": "Ngoc Tran",
          "created_at": "2026-02-18T23:20:38Z"
        }
      ]
    }
  ]
}
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

**Human Output**: `Reply posted to comment df77879f`
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

**Human Output**: `Resolved comment df77879f (+ 2 replies)`
**JSON Output**: `{"resolved": "df77879f", "cascade_count": 2}`

**Exit Codes**: 0 = success, 1 = error

---

## Agent Workflow: Two-Level Retrieval

The intended agent workflow for processing review comments:

```
Step 1: SCAN — get compact overview
  $ sl comment list --status open --json
  → Agent receives ~20 tokens/comment overview
  → Agent identifies thematic clusters from file_path + content_preview
  → Agent picks which comments to process

Step 2: DRILL — get full context for selected comments
  $ sl comment show <id1> <id2> <id3> --json
  → Agent receives full content + thread replies for specific comments
  → Agent has enough context to reason about the feedback

Step 3: ACT — reply and resolve
  $ sl comment reply <id> --content "Addressed: ..."
  $ sl comment resolve <id> --reason "Fixed in ..."
```

This pattern keeps the agent's context budget proportional to the comments it actually processes, not the total comment count on the spec.

---

## Auth Requirements

All commands require valid Supabase credentials (`~/.specledger/credentials.json`).

**Headers**: `Authorization: Bearer <access_token>` + `apikey: <anon_key>`

**Token refresh**: Automatic on 401/PGRST303 via `auth.ForceRefreshAccessToken()`.

## PostgREST Endpoints Used

| Operation | Method | Endpoint |
|-----------|--------|----------|
| List comments | GET | `/rest/v1/review_comments?change_id=eq.{cid}&is_resolved=eq.{status}&parent_comment_id=is.null&select=id,file_path,content,selected_text,author_name,created_at,is_resolved` |
| Count replies | GET | `/rest/v1/review_comments?change_id=eq.{cid}&parent_comment_id=not.is.null&select=id,parent_comment_id` |
| Show comment | GET | `/rest/v1/review_comments?id=eq.{id}&select=*` |
| Show replies | GET | `/rest/v1/review_comments?parent_comment_id=eq.{id}&select=id,content,author_name,created_at&order=created_at.asc` |
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
