# Research: Silent Session Capture

**Feature**: 602-silent-session-capture
**Date**: 2026-03-04

## Prior Work

- **010-checkpoint-session-capture**: Original session capture - PostToolUse hook, queue, `sl session sync`, Supabase storage/metadata.
- **598-silent-session-capture**: Initial attempt - stderr suppression only, superseded by this feature.
- No existing error logging infrastructure in the project.

## Research Findings

### R1: Slash Command Format

**Decision**: Create `specledger.commit.md` following the existing command pattern.

**Rationale**: Slash commands are markdown files with YAML frontmatter (`description` field) + instructions for the agent. The agent reads the markdown and executes the steps. Available variable: `$ARGUMENTS` for user input. Stored in `pkg/embedded/skills/commands/` (source) and `.claude/commands/` (distributed).

**Key insight**: The slash command does NOT call `sl session capture` directly. It instructs the agent to run `git commit` + `git push`. The PostToolUse hook triggers `sl session capture` automatically on `git commit`. The command just needs to:
1. Check auth status via `sl auth status` or by checking credentials file
2. Provide proper commit workflow instructions
3. Show capture/error results after commit

### R2: Capture() Silent Skip Changes

**Decision**: Reorder Capture() - move credentials check before project ID. Return silently (nil error) for both no-credentials and no-project-ID cases.

**Rationale**:
- Credentials check is a fast local file read, should be first
- `GetProjectIDFromRemote()` internally calls `auth.GetValidAccessToken()` - wastes effort without credentials
- Returning nil error means `runSessionCapture()` in commands/session.go already does the right thing (no stderr output)

**Changes needed in `pkg/cli/session/capture.go`**:
- Move lines 288-292 (credentials) to after line 224 (tool success check)
- Change to `return result` (nil error) instead of setting `result.Error`
- Remove 3 `fmt.Fprintf(os.Stderr, ...)` lines at 246-248 for project ID
- Change to `return result` (nil error) instead of setting `result.Error`

### R3: Error Logging — Sentry over Supabase

**Decision**: Use Sentry for remote error reporting instead of a Supabase `session_capture_errors` table.

**Rationale**:
- **Storage overload**: Logging errors to Supabase (PostgREST INSERT or Storage upload) adds load to the same system that handles app data. At scale, troubleshooting data competes with production data for quota.
- **Sentry is purpose-built**: Aggregation, deduplication, alerting, trends, source maps, release tracking — all out of the box. A Supabase table would need custom queries for each of these.
- **Industry standard**: Production Go CLIs use Sentry (or similar: Datadog, Bugsnag) for error reporting. Supabase is for app data, not observability.
- **Simpler code**: `sentry.CaptureException()` with context tags vs. manual HTTP POST to PostgREST with auth token management and retry logic.
- **Local log remains**: `~/.specledger/capture-errors.log` (JSONL) for immediate user troubleshooting. Sentry is for team-level visibility.

**Alternatives rejected**:
- Supabase `session_capture_errors` table: Overloads storage, requires custom dashboard, no alerting
- Log only locally: Team can't troubleshoot remotely for multi-user deployments
- Upload full transcripts to Supabase Storage: 50KB-2MB per session, severe storage overload

**Log entry structure (local file)**:
```json
{
  "timestamp": "2026-03-04T10:00:00Z",
  "user_id": "uuid",
  "project_id": "uuid",
  "session_id": "uuid",
  "error_message": "storage upload failed: connection refused",
  "feature_branch": "602-silent-session-capture",
  "commit_hash": "abc123",
  "retry_count": 0
}
```

**Sentry context tags** (sent with each error event):
- `user.id`, `project_id`, `session_id`, `branch`, `commit_hash`, `retry_count`

**Write order**: Local file first (guaranteed), then Sentry (best-effort, non-blocking). Sentry failure never blocks.

### R4: Sentry Go SDK Integration

**Decision**: Use [rockship-06.sentry.io](https://rockship-06.sentry.io) (hosted SaaS) with `github.com/getsentry/sentry-go` SDK. DSN from rockship-06.sentry.io project settings, configured at build time or via environment variable.

**Setup**:
```go
sentry.Init(sentry.ClientOptions{
    Dsn:         sentryDSN,
    Environment: "production",
    Release:     version.Version,
})
defer sentry.Flush(2 * time.Second)
```

**Error reporting**:
```go
sentry.WithScope(func(scope *sentry.Scope) {
    scope.SetUser(sentry.User{ID: userID})
    scope.SetTag("project_id", projectID)
    scope.SetTag("session_id", sessionID)
    scope.SetTag("branch", branch)
    scope.SetTag("commit_hash", commitHash)
    scope.SetExtra("retry_count", retryCount)
    sentry.CaptureException(err)
})
```

**No Supabase migration needed** — eliminates the `session_capture_errors` table entirely.

### R5: Slash Command Integration Strategy

**Decision**: The slash command instructs the agent through a markdown workflow. It does NOT need a new `sl commit` CLI subcommand.

**Flow**:
1. `/specledger.commit` markdown tells agent to check `sl auth status` or check for `~/.specledger/credentials.json`
2. If no auth: agent runs `git commit` + `git push` directly (PostToolUse hook skips silently)
3. If auth: agent runs `git commit` (PostToolUse hook captures session) then `git push`
4. Agent shows summary based on capture result

**Rationale**: This keeps the implementation simple - the slash command is just instructions, the actual capture logic lives in the existing PostToolUse hook (with the silent-skip fix applied).
