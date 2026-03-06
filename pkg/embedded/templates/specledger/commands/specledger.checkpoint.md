---
description: Verify implementation progress, run tests, and log session accomplishments. Updates session log at .specledger/sessions/<spec>-session.md
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Purpose

Verify implementation progress by running tests, checking for uncommitted changes, and logging session accomplishments. Creates a checkpoint in the session log for traceability.

**When to use**: During or after implementation work to capture progress and verify quality.

## Outline

Goal: Verify current implementation state and log accomplishments for the session.

Execution steps:

1. Run `sl spec info --json --paths-only` to get `FEATURE_DIR` and `BRANCH`.

2. Check for uncommitted changes:
   ```bash
   git status --porcelain
   ```
   - If output is non-empty, note pending changes
   - Prompt user to commit if significant work is done

3. Run tests for modified packages:
   ```bash
   # Identify modified Go packages
   git diff --name-only HEAD~1 | grep '\.go$' | xargs -I {} dirname {} | sort -u

   # Run tests for each modified package
   go test ./pkg/cli/commands/... -v
   ```
   - All tests must pass (exit code 0) for checkpoint to succeed
   - If tests fail, report failures and suggest fixes

4. Check task progress:
   - Read `tasks.md` from `FEATURE_DIR`
   - Identify completed, in-progress, and pending tasks
   - Note any blockers or deviations

5. Update session log at `.specledger/sessions/<branch>-session.md`:
   - Create directory if it doesn't exist
   - Append timestamped entry

   ```markdown
   ## Checkpoint: YYYY-MM-DD HH:MM

   ### Completed
   - <Task or item completed>
   - <Task or item completed>

   ### In Progress
   - <Current work item>

   ### Tests
   - Status: PASS/FAIL
   - Packages tested: <list>

   ### Uncommitted Changes
   - <File paths or "None">

   ### Notes
   - <Any observations, decisions, or deviations>
   ```

6. Report checkpoint summary:
   - Tasks completed this session
   - Tasks remaining
   - Test status
   - Commit recommendation (if uncommitted changes)

## Behavior Rules

- Tests must pass for a successful checkpoint
- Warn if session log is missing or stale
- Don't auto-commit - prompt user instead
- Note any deviations from plan.md
- If no progress since last checkpoint, report "no changes"
- Include file paths for uncommitted changes

## Example Usage

```bash
# Checkpoint after completing a task
/specledger.checkpoint

# Checkpoint with notes
/specledger.checkpoint "Completed authentication flow, need to add error handling"

# Checkpoint before handoff
/specledger.checkpoint "Ready for code review"
```

## Session Log Format

Session logs are stored at `.specledger/sessions/<branch>-session.md`:

```markdown
# Session Log: <branch-name>

## Checkpoint: 2026-03-05 14:30

### Completed
- Implemented sl comment list command
- Added --status filter support

### In Progress
- sl comment show command (50% complete)

### Tests
- Status: PASS
- Packages tested: pkg/cli/commands, pkg/cli/comment

### Uncommitted Changes
- pkg/cli/commands/comment.go
- pkg/cli/comment/client.go

### Notes
- Need to add JSON output for show command
- Discovered edge case with empty reply threads

---
```
