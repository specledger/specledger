# Contract: Session Capture Error Logging

## Overview

Error logging uses two channels:
1. **Local file** (`~/.specledger/capture-errors.log`) — always written, user can inspect directly
2. **Sentry** — remote error aggregation, team troubleshooting, alerting

## Local Log File

**Path**: `~/.specledger/capture-errors.log`
**Format**: JSONL (append-only, one JSON object per line)

```json
{"timestamp":"2026-03-04T10:00:00Z","user_id":"uuid","project_id":"uuid","session_id":"uuid","error_message":"storage upload failed: connection refused","feature_branch":"602-silent-session-capture","commit_hash":"abc123","retry_count":0}
```

## Sentry Integration

**Service**: [rockship-06.sentry.io](https://rockship-06.sentry.io) (hosted SaaS)
**SDK**: `github.com/getsentry/sentry-go`
**DSN**: From rockship-06.sentry.io project settings. Configured via `SENTRY_DSN` environment variable or embedded at build time.

### Error Event Structure

```go
sentry.WithScope(func(scope *sentry.Scope) {
    scope.SetUser(sentry.User{ID: userID})
    scope.SetTag("project_id", projectID)
    scope.SetTag("session_id", sessionID)
    scope.SetTag("branch", featureBranch)
    scope.SetTag("commit_hash", commitHash)
    scope.SetExtra("retry_count", retryCount)
    sentry.CaptureException(err)
})
```

### Context Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `user.id` | Sentry user ID | `550e8400-e29b-41d4-a716-446655440000` |
| `project_id` | SpecLedger project ID | `a1b2c3d4-...` |
| `session_id` | Claude Code session ID | `5430aee0-...` |
| `branch` | Git branch at time of error | `602-silent-session-capture` |
| `commit_hash` | Git commit hash | `65488f5` |
| `retry_count` | Number of retry attempts | `0` (first failure), `1`, `2`, ... |

### Error Handling

- If Sentry send fails: local log already has the error (written first). Never block the workflow.
- Sentry flush on program exit: `sentry.Flush(2 * time.Second)` to ensure buffered events are sent.
- All error reporting is non-blocking and fire-and-forget.

## Auth Decision Matrix

| Has Credentials | Has Project ID | Session Capture | Error Logging |
|----------------|----------------|-----------------|---------------|
| No | - | Skip silently | None |
| Yes | No | Skip silently | None |
| Yes | Yes | Attempt | On failure: local file + Sentry |
