# Quickstart: Checkpoint Session Capture

**Feature**: 010-checkpoint-session-capture

## Prerequisites

1. SpecLedger CLI installed (`sl`)
2. Authenticated: `sl auth login`
3. Claude Code installed and configured
4. Project initialized with SpecLedger

## Setup

### 1. Configure Claude Code Hook

Add to your project's `.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "sl session capture"
          }
        ]
      }
    ]
  }
}
```

This hook fires after every Bash tool use. The `sl session capture` command detects git commits and captures the conversation delta.

### 2. Verify Setup

```bash
# Check auth status
sl auth status

# Verify hook is configured
cat .claude/settings.json | grep "sl session"
```

## Usage

### Automatic Capture (Default Workflow)

1. Work with Claude Code on your feature branch
2. When you commit changes, the hook automatically:
   - Detects the git commit
   - Reads the conversation since the last commit
   - Compresses and uploads it to cloud storage
   - Records metadata in the database
3. No manual action required

### View Past Sessions

```bash
# List sessions for current feature branch
sl session list

# List sessions for a specific branch
sl session list --feature 010-checkpoint-session-capture

# View a session by commit hash
sl session get abc123def456

# View a session by task ID
sl session get SL-42

# JSON output for programmatic use
sl session list --json
```

### Sync Queued Sessions

If sessions failed to upload (e.g., network was down):

```bash
# Upload any locally queued sessions
sl session sync
```

### AI Context Loading

When starting a new AI session on a feature, past sessions are available for context:

```bash
# List all sessions for a feature (AI uses this to load context)
sl session list --feature 010-checkpoint-session-capture --json
```

## Troubleshooting

### Session not captured after commit

1. Verify the Claude Code hook is configured: check `.claude/settings.json`
2. Ensure you're authenticated: `sl auth status`
3. Check if the commit was made via Claude Code's Bash tool (manual terminal commits won't trigger the hook)

### Failed upload

Sessions are cached locally and retried automatically. Run `sl session sync` to force a retry. Check `~/.specledger/session-queue/` for queued sessions.

### Large session warning

Sessions over 10 MB (uncompressed) are rejected. If you hit this limit, commit more frequently to create smaller session segments.
