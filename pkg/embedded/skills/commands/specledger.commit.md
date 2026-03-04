---
description: Commit and push with auth-aware session capture. Use when the user asks to commit via chat.
---

## Purpose

Auth-aware commit workflow that handles session capture gracefully. When the user asks you to commit and push (e.g., "commit giúp tôi", "commit and push for me"), use this command instead of running `git commit && git push` directly.

This does NOT replace Claude's built-in `/commit` command. It only applies when the user asks you to commit via chat.

## Workflow

### Step 1: Stage and verify changes

Run `git status --porcelain` to check for changes.

- If there are unstaged or untracked changes, stage them automatically:
  ```
  git add -A
  ```
- If nothing to commit (working tree clean), tell the user and stop.
- Show a brief summary of what was staged (files changed).

### Step 2: Commit

If `$ARGUMENTS` is provided, use it as the commit message:
```
git commit -m "$ARGUMENTS"
```

If no arguments provided, analyze the staged changes with `git diff --cached` and generate an appropriate commit message. Then commit.

**IMPORTANT**: The commit MUST always proceed. Never skip or block the commit for any reason related to session capture.

### Step 3: Check auth status

Check if session capture should run:

1. Check if `~/.specledger/credentials.json` exists and is valid JSON
2. If no credentials → set `capture_status = "skipped (no auth)"`, go to Step 5
3. If credentials exist, check for project ID in `specledger.yaml` or via `sl` config
4. If no project ID → set `capture_status = "skipped (no project)"`, go to Step 5
5. If both exist → the PostToolUse hook will have already attempted session capture during the commit in Step 2

### Step 4: Note capture result

If auth + project ID were present, the PostToolUse hook (`sl session capture`) ran automatically after the commit.

- If the hook succeeded: `capture_status = "captured"`
- If the hook queued the session: `capture_status = "queued for sync"`
- If the hook had an error: `capture_status = "error (check ~/.specledger/capture-errors.log)"`

### Step 5: Push

Always push regardless of capture status:
```
git push origin <current-branch>
```

If push fails, show the error to the user clearly.

**IMPORTANT**: Push MUST always be attempted. Never skip push due to capture status.

### Step 6: Show summary

Display a summary:
```
Commit: <short-hash> on <branch>
Push: <success/failed>
Session: <capture_status>
```

## Auth Decision Matrix

| Has Credentials | Has Project ID | Session Capture | Error Logging |
|----------------|----------------|-----------------|---------------|
| No             | -              | Skip silently   | None          |
| Yes            | No             | Skip silently   | None          |
| Yes            | Yes            | Attempt         | On failure: local + Supabase |

## Important Notes

- Git commit and push always proceed regardless of session capture
- Zero warnings for unauthenticated users
- Zero warnings for users without a project ID
- Errors only shown when an authenticated user with a project has a real capture failure
- If capture fails, check `~/.specledger/capture-errors.log` for details
