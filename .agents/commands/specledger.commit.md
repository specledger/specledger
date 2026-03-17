---
description: "TRIGGER: When the user asks to commit, push, commit and push, or save changes to git/github in ANY language (e.g. 'commit and push', 'commit giúp tôi', 'push to github', 'commit for me', 'save and push'). This skill MUST be invoked instead of manual git commands."
---

## Purpose

Auth-aware commit workflow. When the user asks you to commit and push (e.g., "commit giúp tôi", "commit and push for me", "commit and push to github", "push to github"), follow this workflow.

**IMPORTANT**: This skill MUST be used for ALL commit/push requests. Do NOT run manual git commit/push commands — always invoke this skill instead.

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

### Step 3: Capture session (inline)

After the commit succeeds, run session capture directly by piping hook-compatible JSON to `sl session capture`. This does NOT rely on PostToolUse hooks.

```bash
CWD_WIN=$(cygpath -m "$(pwd)" 2>/dev/null || pwd) && TRANSCRIPT=$(ls -t ~/.claude/projects/*/*.jsonl 2>/dev/null | head -1) && TRANSCRIPT_WIN=$(cygpath -m "$TRANSCRIPT" 2>/dev/null || echo "$TRANSCRIPT") && SESSION_ID=$(basename "$TRANSCRIPT" .jsonl) && echo '{"session_id":"'"$SESSION_ID"'","transcript_path":"'"$TRANSCRIPT_WIN"'","cwd":"'"$CWD_WIN"'","hook_event_name":"PostToolUse","tool_name":"Bash","tool_input":{"command":"git commit"},"tool_response":{"stdout":"ok","stderr":"","interrupted":false},"tool_use_id":"inline-capture"}' | sl session capture; echo "CAPTURE_EXIT=$?"
```

Set `capture_status` based on the output:
- If output contains "Session captured:" → `"captured"`
- If output contains "Session queued" → `"queued"` (will upload later)
- If no output and CAPTURE_EXIT=0 → `"skipped"` (no auth or no project)
- If CAPTURE_EXIT is non-zero → `"failed"`

**IMPORTANT**: Capture MUST NOT block the workflow. If capture fails for any reason, proceed to push.

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
- Session capture is done INLINE (piped to sl session capture), not via PostToolUse hooks
- Project config is at `specledger/specledger.yaml` (inside specledger/ subfolder), NOT at root
- If capture fails, check `~/.specledger/capture-errors.log` for details
