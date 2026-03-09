# Research: New CLI Commands and Skills

**Feature**: 601-cli-skills | **Date**: 2026-03-03

## Prior Work

### 136-revise-comments (Source for Extraction)

The `sl revise` command already implements comment fetching, display, and resolution. Key files:

| File | Lines | What to Extract |
|------|-------|-----------------|
| `pkg/cli/revise/types.go` | ~100 | `ReviewComment`, `ThreadReply`, `ReplyMap`, `ProcessedComment` |
| `pkg/cli/revise/client.go` | ~200 | PostgREST client, auth retry, comment CRUD |
| `pkg/cli/commands/revise.go` | ~1100 | Interactive flow, `--summary` mode |

### 598-sdd-workflow-streamline (Architecture Decisions)

| Decision | ID | Description |
|----------|----|----|
| Comment Management | D4 | Extract `sl comment` CLI + `sl-comment` skill |
| Spike Command | D13 | Time-boxed exploratory research |
| Checkpoint Command | D14 | Implementation verification + session log |
| Token-Efficient Output | D21 | Compact list, full show, follow-up hints |

### 599-alignment (Command Updates)

- `/specledger.clarify` currently uses `sl revise --summary`
- Will be updated to use `sl comment list --status open --json`

---

## Extraction Analysis

### Types to Extract (from pkg/cli/revise/types.go)

```go
// ReviewComment - a review comment from Supabase
type ReviewComment struct {
    ID           string  `json:"id"`
    ChangeID     string  `json:"change_id"`
    FilePath     string  `json:"file_path"`
    Line         *int    `json:"line"`
    StartLine    *int    `json:"start_line"`
    Content      string  `json:"content"`
    SelectedText string  `json:"selected_text"`
    AuthorName   string  `json:"author_name"`
    AuthorEmail  string  `json:"author_email"`
    Status       string  `json:"status"`
    CreatedAt    string  `json:"created_at"`
}

// ThreadReply - a reply in a comment thread
type ThreadReply struct {
    ID         string `json:"id"`
    ParentID   string `json:"parent_id"`
    Content    string `json:"content"`
    AuthorName string `json:"author_name"`
    CreatedAt  string `json:"created_at"`
}

// ReplyMap - maps parent comment ID to its replies
type ReplyMap map[string][]ReviewComment

// ProcessedComment - comment with optional guidance (stays in revise)
type ProcessedComment struct {
    Comment  ReviewComment
    Guidance string
    Index    int
}
```

**Decision**: Move `ReviewComment`, `ThreadReply`, `ReplyMap` to `pkg/cli/comment/types.go`. Keep `ProcessedComment` in revise (revise-specific).

### Client Methods to Extract (from pkg/cli/revise/client.go)

```go
// Already implemented - move to comment package
type ReviseClient struct {
    accessToken string
    baseURL     string
}

func NewReviseClient(accessToken string) *ReviseClient
func (c *ReviseClient) GetProject(owner, name string) (*Project, error)
func (c *ReviseClient) GetSpec(projectID, specKey string) (*Spec, error)
func (c *ReviseClient) GetChange(specID string) (*Change, error)
func (c *ReviseClient) FetchComments(changeID string) ([]ReviewComment, error)
func (c *ReviseClient) FetchReplies(changeID string) ([]ThreadReply, error)
func (c *ReviseClient) ResolveComment(commentID string) error
func (c *ReviseClient) ResolveCommentWithReplies(commentID string, replyIDs []string) error
```

**New methods needed**:
```go
// NEW - for sl comment reply
func (c *CommentClient) CreateReply(commentID, content string) (*ThreadReply, error)
```

### Summary Mode (from pkg/cli/commands/revise.go)

The `--summary` flag implementation in `runSummary()` provides the pattern for `sl comment list`:

```go
// Current format (sl revise --summary):
// file_path:line  "selected_text"  (author)  [N replies]

// Target format (sl comment list):
// Same compact format, plus --json mode
```

---

## CLI Design

### sl comment list

**Purpose**: List comments with token-efficient output (D21)

**Flags**:
- `--json` - JSON output for agent consumption
- `--status` - Filter: open (default), resolved, all
- `--branch` - Target specific branch (default: current)

**Output Formats**:

```
# Compact (default - human readable)
file_path:line  "selected_text"  (author)  [N replies]

# JSON (--json flag)
[
  {
    "id": "uuid",
    "file_path": "spec.md",
    "line": 42,
    "content": "Full content...",
    "selected_text": "excerpt",
    "author": "reviewer",
    "status": "open",
    "reply_count": 2
  }
]
```

**Exit Codes**:
- 0: Success (including empty list)
- 1: Auth failure (silent, for agent integration)

### sl comment show

**Purpose**: Full comment details with thread replies

**Flags**:
- `--json` - JSON output

**Output Format**:

```json
{
  "id": "uuid",
  "file_path": "spec.md",
  "line": 42,
  "content": "Full content...",
  "selected_text": "excerpt",
  "author": "reviewer",
  "status": "open",
  "created_at": "2026-03-03T10:00:00Z",
  "thread": [
    {
      "id": "reply-uuid",
      "author": "developer",
      "content": "Reply content...",
      "created_at": "2026-03-03T11:00:00Z"
    }
  ]
}
```

### sl comment reply

**Purpose**: Post a reply to a comment thread

**Usage**: `sl comment reply <comment-id> "message"`

**Flags**:
- `--json` - JSON output with reply_id and timestamp

### sl comment resolve

**Purpose**: Mark comments as resolved (with cascade)

**Usage**: `sl comment resolve <comment-id> [<comment-id>...]`

**Behavior**:
- Marks parent comment as resolved
- Cascades to all thread replies
- Supports multiple IDs in one command

---

## Skill Design

### sl-comment Skill Structure

Follows the pattern from `pkg/embedded/skills/specledger-deps/SKILL.md`:

```markdown
# sl-comment Skill

## When to Use

Use this skill when:
- AI commands reference "review comments" or need to process feedback
- You need to list, show, reply to, or resolve review comments
- The `/specledger.clarify` command is invoked

## Key Concepts

### Comment Workflow

1. **List** → Get overview of unresolved comments
2. **Show** → Drill down into specific comment with full context
3. **Reply** → Post response to comment thread
4. **Resolve** → Mark comment as addressed

### Decision Criteria

| Use Case | Command |
|----------|---------|
| Get all open comments | `sl comment list --status open` |
| Get comment details | `sl comment show <id>` |
| Post a reply | `sl comment reply <id> "message"` |
| Mark as resolved | `sl comment resolve <id>` |

### JSON Parsing

```bash
# Get comment IDs
sl comment list --json | jq '.[].id'

# Get comments for specific file
sl comment list --json | jq '.[] | select(.file_path == "spec.md")'
```
```

---

## AI Command Design

### /specledger.spike

**Purpose**: Time-boxed exploratory research

**Template Structure**:
```markdown
---
allowed-tools:
  - Glob
  - Grep
  - Read
  - WebSearch
  - WebFetch
---

# Spike: Research {topic}

**Timebox**: {duration} (default: 30m)

## Objective
[Research goal extracted from topic]

## Approach
1. Search for relevant documentation
2. Analyze existing patterns in codebase
3. Document findings

## Output
Create file at: `specledger/{spec}/research/{date}-{slug}.md`

## Required Sections
- Findings
- Decisions
- Recommendations
```

### /specledger.checkpoint

**Purpose**: Implementation verification + session log

**Template Structure**:
```markdown
# Checkpoint: Progress Verification

## Tasks Check
1. Run `sl issue list --status in_progress`
2. For each in-progress task:
   - Verify tests pass (if applicable)
   - Note completion status

## Git Status
- Check for uncommitted changes
- Prompt to commit if changes found

## Session Summary
Update session log with:
- Tasks completed this session
- Tests status
- Next steps
```

---

## Recommendations

### High Priority

1. **Extract comment package first** - Enables all other work
2. **Implement sl comment list** - Replaces `sl revise --summary`
3. **Update clarify command** - Critical for 599-alignment

### Medium Priority

4. **Create sl-comment skill** - Enhances agent capability
5. **Implement remaining subcommands** - reply, resolve

### Low Priority

6. **Create spike/checkpoint commands** - P2, can be done in parallel

---

## Open Questions

None - all decisions documented in 598 (D4, D13, D14, D21).
