# Data Model: Push-Triggered Scheduler Strategy

**Feature**: 127-specledger-scheduler-push-strategy
**Date**: 2026-03-10

## Entities

### 1. ExecutionLock

Prevents duplicate `sl implement` runs for the same feature.

**Storage**: `.specledger/exec.lock` (JSON file)

| Field       | Type     | Description                              |
|-------------|----------|------------------------------------------|
| pid         | int      | OS process ID of the running sl implement |
| feature     | string   | Feature context (e.g., "127-specledger-scheduler-push-strategy") |
| started_at  | string   | ISO 8601 timestamp of when execution started |

**Validation rules**:
- `pid` must be > 0
- `feature` must be non-empty and match `NNN-feature-name` pattern
- `started_at` must be valid ISO 8601

**State transitions**:
- Created: when `sl implement` starts (background process)
- Removed: when `sl implement` completes (success or failure)
- Stale detection: when PID recorded is no longer running (checked by hook)

### 2. HookExecutionLog

Structured log entry for each push hook invocation.

**Storage**: `.specledger/logs/push-hook.log` (append-only text log)

| Field       | Type     | Description                              |
|-------------|----------|------------------------------------------|
| timestamp   | string   | ISO 8601 timestamp                       |
| branch      | string   | Current git branch name                  |
| feature     | string   | Detected feature name (or "none")        |
| action      | string   | Action taken: "triggered", "skipped", "no-approved-spec", "error" |
| detail      | string   | Additional context (error message, lock info, etc.) |

**Format**: One entry per line, structured as:
```
[2026-03-10T14:30:00Z] branch=127-specledger-scheduler-push-strategy feature=127-specledger-scheduler-push-strategy action=triggered detail="spawned sl implement pid=12345"
```

### 3. SpecStatus (Extended)

Extends the existing spec.md status field with "Approved" state.

**Storage**: `specledger/<feature>/spec.md` (markdown, inline field)

| Value    | Description                                          |
|----------|------------------------------------------------------|
| Draft    | Initial state; spec is being written                 |
| Approved | Spec has passed approval gate; eligible for push-triggered implementation |

**Transition rules**:
- Draft -> Approved: via `sl approve` command, requires spec.md + plan.md + tasks.md present and non-empty
- Approved -> Draft: not supported (manual edit only)

### 4. HookScript

The git hook file content managed by install/uninstall commands.

**Storage**: `.git/hooks/pre-push` (shell script)

**Structure**:
```
[existing hook content, if any]
# BEGIN SPECLEDGER PUSH HOOK
<specledger hook invocation>
# END SPECLEDGER PUSH HOOK
```

**Validation rules**:
- Marker comments must be present for status detection
- Script must be executable (chmod +x)
- SpecLedger block must not contain user modifications (overwritten on reinstall)

## Relationships

```
SpecStatus (spec.md)
  |
  |-- "Approved" triggers -->  HookScript (pre-push)
  |                                |
  |                                |-- checks --> ExecutionLock
  |                                |-- spawns --> sl implement (background)
  |                                |-- writes --> HookExecutionLog
  |
  +-- validated by --> sl approve (requires spec.md + plan.md + tasks.md)
```

## Existing Entities (Referenced, Not Modified)

### Issue (pkg/issues/issue.go)
- Referenced for `sl issue list --all` during previous work extraction
- No schema changes required

### ClaudeSettings / HookMatcher (pkg/cli/hooks/claude.go)
- Existing Claude Code hook infrastructure
- **Not modified** by this feature; git hooks are separate from Claude Code hooks
