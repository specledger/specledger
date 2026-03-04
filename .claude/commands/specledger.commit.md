---
description: Commit and push with auth-aware session capture. Use when the user asks to commit via chat.
---

## Purpose

Auth-aware commit workflow. When the user asks you to commit and push (e.g., "commit giúp tôi", "commit and push for me"), follow this workflow.

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

**IMPORTANT**: The commit MUST always proceed. Never skip or block the commit for any reason.

### Step 3: Check auth and capture status

After the commit, check session capture status:

1. **Check credentials**: `test -f ~/.specledger/credentials.json && echo "OK" || echo "NONE"`
2. **Check project ID**: `cat specledger/specledger.yaml 2>/dev/null | grep "id:"` (note: the file is at `specledger/specledger.yaml`, NOT `specledger.yaml` in root)
3. **Check hook config**: Verify PostToolUse hook exists in `~/.claude/settings.json`
4. **Check capture errors**: `cat ~/.specledger/capture-errors.log 2>/dev/null | tail -1`

Set `capture_status` based on results:
- No credentials → `"skipped (no auth)"`
- No project ID → `"skipped (no project)"`
- Credentials + project ID + no errors → `"captured"` (hook ran automatically after commit)
- Credentials + project ID + new error in log → `"failed"` (show the error)

### Step 4: Push

Always push after commit:
```
git push origin <current-branch>
```

If push fails, show the error to the user clearly.

**IMPORTANT**: Push MUST always be attempted. Never skip push due to capture status.

### Step 5: Show summary

Display a summary:
```
Commit: <short-hash> on <branch>
Push: <success/failed>
Session: <capture_status>
```

If capture failed, also show the last error from `~/.specledger/capture-errors.log`.

## Important Notes

- Git commit and push always proceed — never skip or block for any reason
- Project config is at `specledger/specledger.yaml` (inside specledger/ subfolder), NOT at root
- Hook is configured in `~/.claude/settings.json` (global), NOT in project settings
- Only ONE hook should exist — multiple hooks cause stdin conflicts
- If capture fails, check `~/.specledger/capture-errors.log` for details
