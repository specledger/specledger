# Feature Specification: Silent Session Capture

**Feature Branch**: `602-silent-session-capture`
**Created**: 2026-03-04
**Status**: Draft
**Input**: User description: "Create /specledger.commit slash command for auth-aware commit workflow. When user asks agent to commit (e.g. 'commit giúp tôi'), agent runs this command. No auth → still push, skip capture silently. No project ID → still push, skip capture. Session upload fails → queue for sl session sync retry + log error to Supabase for troubleshooting."

## Problem Statement (V2)
Currently the AI agent (Claude Code) auto-commits and pushes code directly when the user asks "commit push giúp tôi". This causes three problems:

1. **Session capture logs are invisible**: The PostToolUse hook runs `sl session capture` after the commit, but the agent pushes immediately after, so warning/error logs from session capture are not visible to the user.
2. **Spam warnings for unauthenticated users**: If the user hasn't logged in or hasn't synced the project, session capture spams stderr warnings on every commit, confusing users who don't know about the feature.
3. **No troubleshooting capability**: When session capture fails for authenticated users, there is no way to review error logs. Errors are printed to stderr and lost. There is no persistent error log for debugging.

The solution is a `/specledger.commit` slash command that the agent uses when the user asks to commit. It does NOT replace Claude's built-in `/commit` - it is a separate command that handles the commit → capture → push workflow with proper auth checking and error logging.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - /specledger.commit Slash Command (Priority: P1)

As a developer working with Claude Code, I can use `/specledger.commit` to commit and push with auth-aware session capture. The command follows a controlled workflow: stage → commit → check auth → capture session (if applicable) → push. Git commit and push always proceed regardless of capture status. Session capture only runs when auth + project ID are present.

**Why this priority**: This is the core command. It provides the controlled commit flow with proper auth checking.

**Independent Test**: Can be tested by invoking `/specledger.commit` in all auth states and verifying correct behavior.

**Acceptance Scenarios**:

1. **Given** a user has staged changes and is authenticated with a synced project, **When** they invoke `/specledger.commit`, **Then** the system commits, captures the session, pushes to remote, and shows a success summary.
2. **Given** a user has staged changes but is NOT authenticated (no credentials file), **When** they invoke `/specledger.commit`, **Then** the system commits and pushes to remote normally. Session capture is silently skipped with no warnings.
3. **Given** a user has staged changes and is authenticated but has NO project ID, **When** they invoke `/specledger.commit`, **Then** the system commits and pushes to remote normally. Session capture is silently skipped with no warnings.
4. **Given** a user has NO staged changes, **When** they invoke `/specledger.commit`, **Then** the system prompts the user to stage files first and does not commit.

---

### User Story 2 - Agent Uses /specledger.commit When User Asks to Commit (Priority: P1)

As a developer, when I ask the agent in natural language to commit and push (e.g., "commit push giúp tôi", "commit and push for me"), the agent invokes `/specledger.commit` instead of running `git commit && git push` directly. Claude's built-in `/commit` still works as normal for users who use it directly - this only applies when the user chats with the agent to request a commit.

**Why this priority**: This is the most common way users trigger commits through the agent. Without this, the agent bypasses all auth/capture logic.

**Independent Test**: Can be tested by asking the agent to commit/push in natural language and verifying it invokes /specledger.commit.

**Acceptance Scenarios**:

1. **Given** a user asks the agent "commit and push for me" and is authenticated with project ID, **When** the agent processes the request, **Then** it invokes /specledger.commit (commit → capture → push).
2. **Given** a user asks "commit giúp tôi" and is NOT authenticated, **When** the agent processes the request, **Then** it invokes /specledger.commit which commits and pushes normally without session capture warnings.
3. **Given** a user types Claude's built-in `/commit` directly, **Then** it works as normal (unaffected by this feature).

---

### User Story 3 - Session Capture Failure: Queue + Error Log for Troubleshooting (Priority: P2)

As an authenticated developer with a synced project, when session capture upload fails (network error, server error, token expired), three things happen:
1. The session is queued locally for later retry via `sl session sync`
2. The error is logged locally (to a persistent log file) so the user can review it
3. The error is also logged to Supabase so the team can troubleshoot remotely

Errors are always logged to **both** local and Supabase. If Supabase logging fails, the local log still captures the error. This ensures there is always a persistent record of what went wrong.

**Why this priority**: Authenticated users with synced projects have opted in. Failures need to be visible AND debuggable - not just shown once in stderr and lost.

**Independent Test**: Can be tested by simulating upload failure with valid credentials and project ID, then checking both the local log file and error logs on Supabase.

**Acceptance Scenarios**:

1. **Given** a user is authenticated with a valid project ID, **When** session capture upload fails, **Then** the session is queued locally, the error is written to a local log file, AND the error is logged to Supabase (with user ID, project ID, error message, timestamp, session ID).
2. **Given** a queued session exists, **When** the user runs `sl session sync`, **Then** the system retries uploading queued sessions. If retry fails, the error is appended to both local and Supabase logs.
3. **Given** a user is authenticated with a valid project ID, **When** token refresh fails during capture, **Then** the session is queued, error logged to both local and Supabase.
4. **Given** a team admin wants to troubleshoot session capture failures, **When** they query error logs on Supabase, **Then** they can see errors filtered by user/project to identify patterns and root causes.
5. **Given** a user wants to troubleshoot locally, **When** they check the local error log file, **Then** they can see all capture errors with timestamps, error messages, and session IDs.

---

### User Story 4 - Silent Capture in PostToolUse Hook (Priority: P2)

As a developer, the existing PostToolUse hook (`sl session capture`) should also be updated to silently skip when no credentials or no project ID exist. This handles edge cases where commits happen outside the /specledger.commit flow (e.g., direct git commit in terminal).

**Why this priority**: Defense-in-depth. Even if a commit bypasses /specledger.commit, the hook should not spam warnings.

**Independent Test**: Can be tested by making a direct `git commit` and verifying no warnings appear for unauthenticated users.

**Acceptance Scenarios**:

1. **Given** a user has no credentials, **When** a git commit triggers the PostToolUse hook, **Then** session capture exits silently with exit code 0 and no stderr output.
2. **Given** a user has credentials but no project ID, **When** a git commit triggers the PostToolUse hook, **Then** session capture exits silently with exit code 0 and no stderr output.
3. **Given** a user has credentials and project ID, **When** a git commit triggers the hook and capture fails, **Then** the session is queued and error is logged to Supabase.

---

### Edge Cases

- What happens when credentials file exists but is not valid JSON? Silent skip (treat as no credentials).
- What happens when push succeeds but session capture fails? Push already done. Session is queued. Error logged.
- What happens when user is offline? Commit succeeds locally. Push fails with clear error. Session is queued.
- What happens when user provides a commit message via argument? Use the provided message directly.
- What happens when the user authenticates mid-session? Next commit picks up the new credentials.
- What happens when Supabase error logging itself fails? Local log file already has the error (written first). Do not block the workflow.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a `/specledger.commit` slash command that handles commit → capture → push workflow. This does NOT replace Claude's built-in `/commit`.
- **FR-002**: The slash command MUST always commit and push regardless of session capture status. Commit and push are never blocked by capture failures.
- **FR-003**: The slash command MUST check for local credentials (`~/.specledger/credentials.json`) before attempting session capture.
- **FR-004**: The slash command MUST silently skip session capture (no warnings, no errors) when no valid credentials are found locally.
- **FR-005**: The slash command MUST silently skip session capture when credentials exist but no project ID can be resolved.
- **FR-006**: When session capture fails for an authenticated user with a valid project ID, the system MUST: (a) display the error to the user, (b) write the error to a local log file, AND (c) log the error to Supabase. All three happen on every failure.
- **FR-007**: Failed session captures MUST be queued locally for retry via `sl session sync`. When `sl session sync` retries and fails again, the retry error MUST also be logged to both local and Supabase.
- **FR-008**: The agent MUST invoke /specledger.commit when the user asks to commit/push via chat (e.g., "commit giúp tôi"). Claude's built-in /commit is unaffected.
- **FR-009**: The existing PostToolUse hook MUST be updated to silently skip when no credentials or no project ID exist.
- **FR-010**: System MUST never produce a non-zero exit code from the session capture hook.
- **FR-011**: The slash command MUST accept an optional commit message as an argument.
- **FR-012**: Error logs MUST include: user ID, project ID, error message, timestamp, and session ID for correlation. This applies to BOTH local log file and Supabase log.
- **FR-013**: Error logging MUST always write to local log file first (guaranteed), then attempt Supabase. If Supabase logging fails, the local log still has the full record. Logging failures MUST never block the commit/push workflow.
- **FR-014**: The local error log file location MUST be discoverable (e.g., `~/.specledger/capture-errors.log`) so users can check it for troubleshooting.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero warning messages appear when an unauthenticated user commits via /specledger.commit or the agent.
- **SC-002**: Zero warning messages appear when an authenticated user without a synced project commits.
- **SC-003**: Git commit and push always succeed regardless of session capture status (capture never blocks commit/push).
- **SC-004**: Clear error messages appear only when an authenticated user with a synced project experiences a real capture failure.
- **SC-005**: All capture errors for authenticated users are persisted to BOTH local log file AND Supabase (queryable by user/project).
- **SC-006**: `sl session sync` successfully retries and uploads previously queued sessions.
- **SC-007**: All existing session capture tests continue to pass (no regression).

### Previous work

- **010-checkpoint-session-capture**: Original session capture implementation with PostToolUse hook mechanism, queue system, and `sl session sync`.
- **598-silent-session-capture**: Initial attempt to fix silent capture (stderr suppression only - superseded by this feature which adds the slash command + error logging approach).
