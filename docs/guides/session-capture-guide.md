# Session Capture Guide

> Capture and retrieve AI conversation sessions linked to commits and tasks.

## Overview

Session capture automatically records AI conversations when you commit code while working with Claude Code. These sessions provide:

- **Audit trail** - Track what AI assistance and reasoning led to each commit
- **Retrospectives** - Review past sessions for team retrospectives and process improvement
- **Data mining** - Analyze captured sessions for patterns, insights, and documentation

## Prerequisites

1. **SpecLedger CLI installed**: `sl --version`
2. **Project initialized**: `specledger.yaml` must have `project.id` set
3. **Authenticated**: Run `sl auth login`

## Quick Start

```bash
# 1. Login (one-time)
sl auth login

# 2. Check your sessions
sl session list

# 3. View a specific session
sl session get <commit-hash>

# 4. Upload any queued sessions (if offline captures exist)
sl session sync
```

## Commands

### `sl session list`

List sessions for a feature branch.

```bash
# List sessions for current branch
sl session list

# List sessions for a specific branch
sl session list --feature main

# Filter by commit hash (partial match works)
sl session list --commit abc123

# Filter by task ID
sl session list --task SL-42

# Limit results
sl session list --limit 10

# JSON output (for scripts/AI)
sl session list --json
```

**Example output:**
```
COMMIT   MESSAGES  SIZE     STATUS    CAPTURED
abc1234  45        12.5 KB  complete  2026-02-23 10:30
def5678  23        8.2 KB   complete  2026-02-22 14:15
```

### `sl session get`

Retrieve and display a session's conversation content.

```bash
# Get by partial commit hash
sl session get abc1234

# Get by task ID
sl session get SL-42

# Get by full session UUID
sl session get 550e8400-e29b-41d4-a716-446655440000

# Output as JSON (for AI processing)
sl session get abc1234 --json

# Output raw gzip stream (for piping)
sl session get abc1234 --raw > session.json.gz
```

**Example output:**
```
Session: 550e8400-e29b-41d4-a716-446655440000
Branch:  feature/add-auth
Commit:  abc1234def5678...
Author:  user@example.com
Date:    2026-02-23 10:30:45
Messages: 45
------------------------------------------------------------

[USER] 10:25:30
Can you help me implement user authentication?

[ASSISTANT] 10:25:35
I'll help you implement user authentication. Let me first check...
```

### `sl session sync`

Upload sessions that were queued due to network failures.

```bash
# Upload all queued sessions
sl session sync

# Check queue status without uploading
sl session sync --status

# JSON output
sl session sync --json
```

**Example output (--status):**
```
2 session(s) queued for upload:
  550e8400  abc1234  (retries: 0)
  661f9511  def5678  (retries: 1)
```

**Example output (after sync):**
```
Uploaded 2 session(s)
```

### `sl session capture`

> **Internal command** - Called automatically by Claude Code hooks.

This command is triggered by the PostToolUse hook when you run `git commit`. You don't need to call it manually.

## How It Works

### When Does Capture Trigger?

Session capture **only triggers** when ALL of these conditions are met:

| Condition | Description |
|-----------|-------------|
| **1. Bash tool used** | Hook only fires on Bash commands |
| **2. Command is git commit** | Must contain `git commit` (not `git add`, `git push`, etc.) |
| **3. Commit succeeded** | `tool_success: true` in hook input |
| **4. Project has ID** | `specledger.yaml` must have `project.id` set |
| **5. Transcript exists** | Claude Code must be saving conversation to JSONL |

**Examples of what triggers capture:**
```bash
git commit -m "Add feature"           # ✓ Triggers
git commit --amend                    # ✓ Triggers
git commit -a -m "Fix bug"            # ✓ Triggers
```

**Examples of what does NOT trigger capture:**
```bash
git add .                             # ✗ Not a commit
git push origin main                  # ✗ Not a commit
git status                            # ✗ Not a commit
echo "git commit"                     # ✗ Not actually running git commit
```

### What Gets Captured?

Each session capture includes:

| Field | Description |
|-------|-------------|
| `session_id` | Unique UUID for this capture |
| `project_id` | Project identifier from specledger.yaml |
| `feature_branch` | Current git branch name |
| `commit_hash` | Full 40-character commit hash |
| `author` | User email from credentials |
| `messages` | Array of conversation messages (delta since last capture) |
| `captured_at` | Timestamp of capture |

### Delta Capture (Incremental)

Sessions capture **only new messages** since the last capture for that session:

```
Session starts (offset = 0)
├── Message 1: "Help me add a feature"
├── Message 2: "Sure, let me check..."
├── Message 3: "Here's the implementation..."
│
├── git commit -m "Add feature"  ← Capture messages 1-3, offset = 3
│
├── Message 4: "Now add tests"
├── Message 5: "I'll create test cases..."
│
└── git commit -m "Add tests"    ← Capture messages 4-5 only (delta)
```

Offset state is stored at: `~/.specledger/session-state.json`

### Automatic Capture Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  1. You work with Claude Code                                   │
│     └── Conversation is recorded in transcript                  │
│                                                                 │
│  2. You run: git commit -m "Add feature X"                      │
│     └── PostToolUse hook triggers: sl session capture           │
│                                                                 │
│  3. Session capture:                                            │
│     ├── Detects git commit command                              │
│     ├── Reads conversation since last capture (delta)          │
│     ├── Compresses with gzip                                    │
│     ├── Uploads to Supabase Storage                             │
│     └── Records metadata in database                            │
│                                                                 │
│  4. If network fails:                                           │
│     └── Session queued locally for later upload                 │
│                                                                 │
│  5. Next time online:                                           │
│     └── Run: sl session sync                                    │
└─────────────────────────────────────────────────────────────────┘
```

### Storage Structure

- **Content**: Supabase Storage bucket `sessions`
- **Path**: `{project_id}/{branch}/{commit_hash}.json.gz`
- **Metadata**: Supabase PostgreSQL `sessions` table

## Configuration

### Enable Session Capture

Session capture is enabled via Claude Code hooks. Check `.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [{
      "matcher": "Bash",
      "hooks": [{
        "type": "command",
        "command": "sl session capture"
      }]
    }]
  }
}
```

### Project ID Setup

Ensure `specledger.yaml` has the project ID:

```yaml
project:
  id: "your-project-uuid-from-supabase"
```

To get your project ID, check the Supabase dashboard or run `sl init`.

## Testing Session Capture

### Test Mode (Recommended)

Use `--test-mode` to verify your setup without making a real commit.

> **Note:** Run this command from within Claude Code (not a regular terminal) because it needs access to the active transcript path.

```bash
sl session capture --test-mode
```

**Expected output if configured correctly:**
```
Running in test mode...
✓ Git repository detected
✓ Project ID found: abc123-def456-...
✓ Authenticated as: user@example.com
✓ Claude Code sessions directory found
✓ Found transcript: /home/user/.claude/sessions/xxx/transcript.jsonl
✓ Session ID: xxx-yyy-zzz

📝 Simulating git commit hook...

⚠️  Test mode simulates the capture flow but won't create a real session.
To capture a real session, make a git commit while using Claude Code.
✅ Session capture system is configured correctly!

Next steps:
  1. Work on your code with Claude Code
  2. Make a git commit
  3. Session will be automatically captured

To view sessions: sl session list
```

### Verify After Real Commit

After making a commit, check if session was captured:

```bash
# List sessions - should show your recent commit
sl session list

# Expected output:
# COMMIT   MESSAGES  SIZE     STATUS    CAPTURED
# abc1234  45        12.5 KB  complete  2026-02-23 10:30
```

## Troubleshooting

### "authentication required"

```bash
# Re-authenticate
sl auth login
```

### "project not configured"

Ensure `specledger.yaml` exists with `project.id`:

```bash
# Check if file exists
cat specledger.yaml

# Or initialize project
sl init
```

### "Session capture skipped"

This warning appears during commits when:
- `project.id` is not set in `specledger.yaml`
- Not in a git repository
- Transcript path not available

**Fix**: Ensure your project is properly initialized with `sl init`.

### Sessions not appearing in list

1. Check you're on the right branch: `git branch --show-current`
2. Check authentication: `sl auth login`
3. Check for queued sessions: `sl session sync --status`

### Queued sessions failing to upload

```bash
# Check queue status
sl session sync --status

# View detailed errors
sl session sync --json

# If max retries reached, sessions are skipped
# Check ~/.specledger/session-queue/ for orphaned files
```

## Best Practices

1. **Commit often** - Each commit captures the relevant conversation delta
2. **Use descriptive commit messages** - Sessions are linked to commits
3. **Run sync when back online** - `sl session sync` uploads queued sessions
4. **Use JSON output for AI** - `sl session list --json` for programmatic access

## Programmatic Access

Sessions can be exported for analysis and tooling:

```bash
# Get recent sessions as JSON
sl session list --feature main --limit 5 --json

# Load specific session content
sl session get abc1234 --json
```

Use JSON output to integrate with reporting tools, build dashboards, or perform analysis on AI-assisted development patterns.
