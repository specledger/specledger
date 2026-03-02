# Feature Specification: Silent Session Capture

**Feature Branch**: `598-silent-session-capture`
**Created**: 2026-03-02
**Status**: Draft
**Input**: User description: "Fix session capture to silently skip when user has no access token or no project ID, instead of spamming warning logs on every commit. Only log errors when user is authenticated AND project is synced but upload actually fails."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unauthenticated User Commits Without Warnings (Priority: P1)

As a developer who uses the SpecLedger CLI locally but has not logged in (`sl auth login`), when I make git commits while working with Claude Code, the session capture hook runs silently without producing any warning messages. I should not be aware that session capture exists unless I opt in by authenticating.

**Why this priority**: This is the most common case causing user confusion. Users who haven't authenticated don't know about session capture and interpret the warnings as bugs in their workflow.

**Independent Test**: Can be fully tested by removing credentials file, making a commit via Claude Code, and verifying zero output to stderr from the session capture hook.

**Acceptance Scenarios**:

1. **Given** a user has no credentials file, **When** they make a git commit via Claude Code, **Then** the session capture hook exits silently with exit code 0 and produces no output to stderr.
2. **Given** a user has a malformed or empty credentials file, **When** they make a git commit via Claude Code, **Then** the session capture hook exits silently with exit code 0 and produces no output to stderr.
3. **Given** a user has valid credentials structure but the access token and refresh token fields are empty, **When** they make a git commit via Claude Code, **Then** the session capture hook exits silently with exit code 0 and produces no output to stderr.

---

### User Story 2 - Authenticated User Without Synced Project Commits Silently (Priority: P1)

As a developer who has logged in to SpecLedger but has not synced my current project (no project ID in specledger.yaml and project not registered on the platform), when I make git commits, the session capture hook runs silently. No warnings about missing project ID should appear.

**Why this priority**: Equally important as US1 - authenticated users working on unregistered projects should not be spammed with "project.id not set" warnings on every commit.

**Independent Test**: Can be fully tested by having valid credentials but no specledger.yaml (or one without project.id), making a commit, and verifying zero output to stderr.

**Acceptance Scenarios**:

1. **Given** a user has valid credentials but no `specledger.yaml` exists in the project, **When** they make a git commit via Claude Code, **Then** the session capture hook exits silently with exit code 0 and produces no output to stderr.
2. **Given** a user has valid credentials and `specledger.yaml` exists but has no `project.id` field, and the project is not registered on the platform (git remote lookup fails), **When** they make a git commit, **Then** the session capture hook exits silently with exit code 0 and produces no output to stderr.

---

### User Story 3 - Authenticated User With Synced Project Gets Error Logging on Upload Failure (Priority: P2)

As a developer who has logged in and has a synced project (project ID exists), when session upload fails due to network issues or server errors, the system should log a meaningful error so I can troubleshoot. This is the only scenario where errors should be visible.

**Why this priority**: Once a user has opted in (authenticated + project synced), upload failures represent real problems that should be surfaced for troubleshooting.

**Independent Test**: Can be tested by having valid credentials and project ID, simulating a network failure during upload, and verifying that an error message appears on stderr.

**Acceptance Scenarios**:

1. **Given** a user has valid credentials and a valid project ID, **When** a git commit triggers session capture but storage upload fails, **Then** the session is queued for later upload and a message is printed to stderr indicating the session was queued.
2. **Given** a user has valid credentials and a valid project ID, **When** a git commit triggers session capture but metadata creation fails after successful storage upload, **Then** the session is queued and a message is printed to stderr.
3. **Given** a user has valid credentials and a valid project ID, **When** token refresh fails during session capture, **Then** the session is queued for later upload and a message is printed to stderr.

---

### User Story 4 - Successful Session Capture Continues Working (Priority: P2)

As a developer with full setup (authenticated + synced project), the existing happy-path behavior of session capture remains unchanged: sessions are captured, uploaded, and confirmed on stderr.

**Why this priority**: Ensuring no regression in the working path is critical.

**Independent Test**: Can be tested by making a commit with full setup and verifying the session captured confirmation message appears.

**Acceptance Scenarios**:

1. **Given** a user has valid credentials and a valid project ID, **When** a git commit triggers session capture and upload succeeds, **Then** a success message is printed to stderr with session ID, message count, and size.

---

### Edge Cases

- What happens when credentials file exists but is not valid JSON? Silent skip (treat as no credentials).
- What happens when the user authenticates mid-session (credentials appear between commits)? Next commit should attempt full capture.
- What happens when project ID is removed from specledger.yaml after a previous successful capture? Silent skip on next commit.
- What happens when the credentials file has valid structure but expired tokens and refresh also fails? Session should be queued (user has opted in, this is a real failure).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST check for local credentials existence before attempting any session capture operations (project ID lookup, upload, etc.).
- **FR-002**: System MUST silently exit (exit code 0, no stderr output) when no valid credentials are found locally.
- **FR-003**: System MUST silently exit (exit code 0, no stderr output) when credentials exist but no project ID can be resolved (neither from specledger.yaml nor git remote lookup).
- **FR-004**: System MUST log errors to stderr only when the user has valid credentials AND a resolvable project ID AND an actual upload/metadata operation fails.
- **FR-005**: System MUST preserve the existing behavior for queuing sessions when upload fails (authenticated user with project ID).
- **FR-006**: System MUST preserve the existing happy-path behavior (successful capture prints confirmation to stderr).
- **FR-007**: System MUST never produce a non-zero exit code from the session capture hook (to avoid blocking git commits).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero warning messages appear on stderr when an unauthenticated user makes commits via Claude Code.
- **SC-002**: Zero warning messages appear on stderr when an authenticated user without a synced project makes commits via Claude Code.
- **SC-003**: Error messages appear on stderr only when an authenticated user with a synced project experiences an upload failure.
- **SC-004**: All existing session capture tests continue to pass without modification (no regression).
- **SC-005**: The session capture hook completes within the same time budget as before (no added latency for the silent-skip paths).

### Previous work

- **010-checkpoint-session-capture**: Original implementation of the session capture feature, including hook mechanism, transcript delta computation, Supabase storage/metadata upload, and local queuing for offline captures.
