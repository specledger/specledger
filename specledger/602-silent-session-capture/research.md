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

### R3: Error Logging Infrastructure

**Decision**: Create a new `CaptureErrorLogger` that writes to both local file and Supabase.

**Rationale**:
- No error logging exists in the project currently
- Supabase PostgREST pattern already established in `metadata.go` and `storage.go`
- Need a new Supabase table: `session_capture_errors`
- Local log at `~/.specledger/capture-errors.log` (append-only, structured JSON lines)

**Log entry structure**:
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

**Write order**: Local file first (guaranteed), then Supabase (best-effort). Supabase failure never blocks.

**Alternatives considered**:
- Log only to Supabase: Rejected - not available when offline or when Supabase itself fails
- Log only locally: Rejected - team can't troubleshoot remotely
- Use Go's `log` package: Rejected - need structured JSON for queryability

### R4: Supabase Error Log Table

**Decision**: New table `session_capture_errors` on Supabase with RLS (Row Level Security) by user.

**Required columns**:
- `id` (UUID, primary key)
- `user_id` (text, indexed)
- `project_id` (text, indexed)
- `session_id` (text, nullable)
- `error_message` (text)
- `feature_branch` (text, nullable)
- `commit_hash` (text, nullable)
- `retry_count` (integer, default 0)
- `created_at` (timestamptz, default now())

**RLS policy**: Users can read/write their own errors. Project admins can read all errors for their projects.

**Note**: Table creation is a Supabase migration, not a Go code change. The Go code just POSTs to `/rest/v1/session_capture_errors`.

### R5: Slash Command Integration Strategy

**Decision**: The slash command instructs the agent through a markdown workflow. It does NOT need a new `sl commit` CLI subcommand.

**Flow**:
1. `/specledger.commit` markdown tells agent to check `sl auth status` or check for `~/.specledger/credentials.json`
2. If no auth: agent runs `git commit` + `git push` directly (PostToolUse hook skips silently)
3. If auth: agent runs `git commit` (PostToolUse hook captures session) then `git push`
4. Agent shows summary based on capture result

**Rationale**: This keeps the implementation simple - the slash command is just instructions, the actual capture logic lives in the existing PostToolUse hook (with the silent-skip fix applied).
