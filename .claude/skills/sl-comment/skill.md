# sl-comment Skill

**When to Load**: Triggered when AI commands mention "review comments", "comment", or reference `sl comment` CLI.

## Overview

The `sl comment` CLI provides review comment management for SpecLedger projects. Comments are stored in Supabase and accessed via the PostgREST API.

## Subcommands

| Command | Purpose | Output Mode |
|---------|---------|-------------|
| `sl comment list` | List all comments (compact) | Truncated previews, reply counts |
| `sl comment show <id>` | Full comment details | Complete content, all replies |
| `sl comment reply <id> "msg"` | Reply to a comment | Minimal confirmation |
| `sl comment resolve <id> --reason "text"` | Mark comment resolved (reason required, posted as reply) | Minimal confirmation |

## Decision Criteria

### When to Use list vs show

**Use `sl comment list` when:**
- Scanning for open issues across all files
- Getting an overview of review feedback
- Initial context gathering (compact mode)

**Use `sl comment show` when:**
- Understanding a specific comment's full context
- Reading thread replies
- Before replying or resolving

### When to Reply vs Resolve

**Reply (`sl comment reply`):**
- Providing information without closing the issue
- Asking clarifying questions
- Explaining implementation approach

**Resolve (`sl comment resolve --reason`):**
- Comment has been addressed in code — `--reason` is required and auto-posts a reply
- Issue is no longer relevant
- Duplicate of another comment

## JSON Parsing Examples

### List Comments (Compact)

```bash
sl comment list --status open --json
```

```json
[
  {
    "id": "abc123",
    "file_path": "spec.md",
    "line": 42,
    "content_preview": "This requirement is ambiguous...",
    "author": "alice",
    "reply_count": 2
  }
]
```

**Note**: `content_preview` is truncated to 120 characters. Use `sl comment show` for full content.

### Show Comment (Full)

```bash
sl comment show abc123 --json
```

```json
{
  "id": "abc123",
  "file_path": "spec.md",
  "line": 42,
  "content": "This requirement is ambiguous - does it apply to all users or just admins?",
  "selected_text": "all users",
  "author": "alice",
  "status": "open",
  "created_at": "2026-03-05T10:30:00Z",
  "replies": [
    {
      "id": "reply1",
      "author": "bob",
      "content": "Good point - we should clarify this in the spec.",
      "created_at": "2026-03-05T11:00:00Z"
    }
  ]
}
```

## Workflow Patterns

### Pattern 1: Address Review Feedback

```bash
# 1. List all open comments
sl comment list --status open --json

# 2. For each comment, get full details
sl comment show <id> --json

# 3. After addressing in code, resolve with reason (auto-posts reply)
sl comment resolve <id> --reason "Fixed in commit abc123. Added role check."
```

### Pattern 2: Batch Processing

```bash
# List all open comments, then resolve multiple
sl comment list --status open --json
# ... after addressing all ...
sl comment resolve id1 id2 id3 --reason "Batch resolved: all addressed in latest revision"
```

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| Exit code 1 (silent) | No auth token | Run `sl auth login` |
| "comment not found" | Invalid ID | Verify ID from list output |
| Network error | Supabase unavailable | Retry or check connectivity |

## Token Efficiency (D21)

The CLI follows the token-efficient output pattern:

- **list**: Compact mode by default (~500 tokens for 25 comments)
- **show**: Full detail justified by explicit drill-down request (~200 tokens per comment)
- **reply/resolve**: Minimal output (~30 tokens)

Use `--json` flag for programmatic parsing by AI agents.
