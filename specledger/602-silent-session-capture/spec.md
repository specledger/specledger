# Feature Specification: Silent Session Capture

**Feature Branch**: `602-silent-session-capture`
**Created**: 2026-03-04
**Status**: Draft
**Input**: User description: "Create a /specledger.commit slash command to control the commit/push workflow with auth-aware session capture. Agent must not auto-commit/push blindly. Check auth locally first - if no auth, silently skip capture without logging errors. If auth exists but capture/upload fails, log errors. Handle both user-invoked /commit and agent-driven commit scenarios."

## Problem Statement

Currently the AI agent (Claude Code) auto-commits and pushes code directly when the user asks "commit push giúp tôi". This causes two problems:

1. **Session capture logs are invisible**: The PostToolUse hook runs `sl session capture` after the commit, but the agent pushes immediately after, so warning/error logs from session capture are not visible to the user.
2. **Spam warnings for unauthenticated users**: If the user hasn't logged in or hasn't synced the project, session capture spams stderr warnings on every commit, confusing users who don't know about the feature.

The solution is a `/specledger.commit` slash command that gives the system control over the commit → capture → push workflow, with proper auth checking at each step.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - User Invokes /specledger.commit (Priority: P1)

As a developer working with Claude Code, I type `/commit` (or ask the agent to commit) and the system handles the full workflow: stage → commit → session capture → push. Before attempting session capture or push, it checks if I'm authenticated. If I'm not authenticated, it commits locally without capture and without errors. If I am authenticated, it captures the session and pushes.

**Why this priority**: This is the primary use case. Every commit through the agent should go through this controlled flow.

**Independent Test**: Can be tested by invoking `/specledger.commit` with and without credentials, verifying the correct behavior in each case.

**Acceptance Scenarios**:

1. **Given** a user has staged changes and is authenticated with a synced project, **When** they invoke `/specledger.commit`, **Then** the system commits, captures the session, pushes to remote, and shows a success summary.
2. **Given** a user has staged changes but is NOT authenticated, **When** they invoke `/specledger.commit`, **Then** the system commits locally, silently skips session capture (no warnings), and pushes to remote without errors.
3. **Given** a user has staged changes and is authenticated but has NO project ID, **When** they invoke `/specledger.commit`, **Then** the system commits locally, silently skips session capture (no warnings), and pushes to remote.
4. **Given** a user has NO staged changes, **When** they invoke `/specledger.commit`, **Then** the system prompts the user to stage files first and does not commit.

---

### User Story 2 - Agent Handles "commit push giúp tôi" Requests (Priority: P1)

As a developer, when I ask the agent in natural language to commit and push (e.g., "commit push giúp tôi", "commit and push for me"), the agent must follow the same controlled workflow as `/specledger.commit` instead of running `git commit && git push` directly. This ensures auth checking and session capture happen properly.

**Why this priority**: Equally critical as US1 - this is the most common way users trigger commits through the agent.

**Independent Test**: Can be tested by asking the agent to commit/push in natural language and verifying it follows the controlled workflow.

**Acceptance Scenarios**:

1. **Given** a user asks the agent "commit and push for me" and is authenticated with project ID, **When** the agent processes the request, **Then** it follows the /specledger.commit workflow (commit → capture → push) with proper auth checking.
2. **Given** a user asks the agent "commit giúp tôi" and is NOT authenticated, **When** the agent processes the request, **Then** it commits and pushes without session capture warnings.

---

### User Story 3 - Auth Check with Error Logging on Real Failures (Priority: P2)

As an authenticated developer with a synced project, when session capture or push fails due to network/server issues, I should see clear error messages so I can troubleshoot. This is the ONLY scenario where errors should be visible.

**Why this priority**: Once a user has opted in (authenticated + synced project), failures are real problems worth surfacing.

**Independent Test**: Can be tested by simulating network failure during capture/push with valid credentials and project ID.

**Acceptance Scenarios**:

1. **Given** a user is authenticated with a valid project ID, **When** session capture upload fails, **Then** the session is queued locally and an error message is displayed explaining the failure and that it will retry.
2. **Given** a user is authenticated with a valid project ID, **When** git push fails, **Then** an error message is displayed with the failure reason.
3. **Given** a user is authenticated with a valid project ID, **When** token refresh fails, **Then** the session is queued and an error message is displayed.

---

### User Story 4 - Silent Capture in PostToolUse Hook (Priority: P2)

As a developer, the existing PostToolUse hook (`sl session capture`) should also be updated to silently skip when no credentials or no project ID exist, matching the behavior of /specledger.commit. This handles edge cases where commits happen outside the slash command.

**Why this priority**: Defense-in-depth - even if a commit bypasses /specledger.commit, the hook should not spam warnings.

**Independent Test**: Can be tested by making a direct `git commit` (not through /specledger.commit) and verifying no warnings appear for unauthenticated users.

**Acceptance Scenarios**:

1. **Given** a user has no credentials, **When** a git commit triggers the PostToolUse hook, **Then** session capture exits silently with exit code 0 and no stderr output.
2. **Given** a user has credentials but no project ID, **When** a git commit triggers the PostToolUse hook, **Then** session capture exits silently with exit code 0 and no stderr output.

---

### Edge Cases

- What happens when credentials file exists but is not valid JSON? Silent skip (treat as no credentials).
- What happens when push succeeds but session capture fails? Session is queued locally, user is informed.
- What happens when user is offline? Commit succeeds locally, push fails with clear error, session is queued.
- What happens when user provides a commit message via /specledger.commit argument? Use the provided message directly.
- What happens when the user authenticates mid-session? Next commit through /specledger.commit picks up the new credentials.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a `/specledger.commit` slash command that orchestrates the commit → capture → push workflow.
- **FR-002**: The slash command MUST check for local credentials existence before attempting session capture.
- **FR-003**: The slash command MUST silently skip session capture (no warnings, no errors) when no valid credentials are found locally.
- **FR-004**: The slash command MUST silently skip session capture when credentials exist but no project ID can be resolved.
- **FR-005**: The slash command MUST log errors only when the user has valid credentials AND a resolvable project ID AND an actual capture/upload operation fails.
- **FR-006**: The slash command MUST handle the push step separately from session capture, so push works even if capture fails.
- **FR-007**: The agent MUST use the /specledger.commit workflow when the user asks to commit/push in natural language.
- **FR-008**: The existing PostToolUse hook MUST be updated to silently skip when no credentials or no project ID exist.
- **FR-009**: System MUST never produce a non-zero exit code from the session capture hook (to avoid blocking git commits).
- **FR-010**: The slash command MUST accept an optional commit message as an argument.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero warning messages appear when an unauthenticated user commits via /specledger.commit or the agent.
- **SC-002**: Zero warning messages appear when an authenticated user without a synced project commits.
- **SC-003**: Clear error messages appear only when an authenticated user with a synced project experiences a real capture/push failure.
- **SC-004**: The /specledger.commit command completes the full workflow (commit → capture → push) in a single invocation.
- **SC-005**: All existing session capture tests continue to pass (no regression).

### Previous work

- **010-checkpoint-session-capture**: Original session capture implementation with PostToolUse hook mechanism.
- **598-silent-session-capture**: Initial attempt to fix silent capture (stderr suppression only - superseded by this feature which adds the slash command approach).
