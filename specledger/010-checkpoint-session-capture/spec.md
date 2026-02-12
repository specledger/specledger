# Feature Specification: Checkpoint Session Capture

**Feature Branch**: `010-checkpoint-session-capture`
**Created**: 2026-02-11
**Status**: Draft
**Input**: User description: "Capture AI chat sessions during checkpoint creation and beads task execution, store them to S3, and associate session links with checkpoints in the database."

## Clarifications

### Session 2026-02-11

- Q: What portion of the AI conversation should be captured per checkpoint? → A: Segment (delta) only — each checkpoint captures only the conversation since the last checkpoint.
- Q: Are checkpoints already tracked as entities in the system, or does this feature need to define them? → A: Git commits serve as checkpoints — each git commit on the feature branch IS the checkpoint, identified by commit hash. No separate checkpoint entity needed.
- Q: Who within a project should be able to view stored sessions? → A: All project members — any authenticated member of the project can view any session in that project.
- Q: What format should the captured session content use? → A: Structured with messages — preserving individual messages with role (user/AI), timestamp, and content for filtering, search, and rich rendering.
- Q: (User clarification) Sessions are too large for commit messages and must be stored in cloud storage. The primary retrieval use case is the AI itself — loading past sessions to recall user decisions, user-provided system context, and preferences to improve continuity across sessions. → A: Sessions are stored in cloud storage (not commit messages). AI is a first-class consumer that queries past sessions for context continuity.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Capture Session on Commit (Priority: P1)

As a developer working with AI on a feature, when I commit changes (which serves as a checkpoint), the system automatically captures the AI conversation segment since the last commit and stores it in a structured format so I can reference it later.

This is the core workflow: the developer uses Claude Code to work on changes, commits the result (creating a checkpoint), and the conversation segment for that work is preserved as an artifact linked to the commit. This provides traceability — anyone reviewing the commit can see the reasoning, decisions, and trade-offs discussed during its creation.

**Why this priority**: This is the primary use case — every AI-assisted commit should have its session segment recorded. Without this, there is no session data to store or retrieve.

**Independent Test**: Can be fully tested by creating a commit via Claude Code and verifying that the conversation segment is captured in structured format, stored, and linked to the commit hash.

**Acceptance Scenarios**:

1. **Given** a developer is working with Claude Code on a feature, **When** the developer commits the changes, **Then** the system captures the conversation segment since the last commit in structured format (preserving message roles, timestamps, and content) and stores it.
2. **Given** a session has been captured and stored, **When** the session metadata is recorded in the database, **Then** it is associated with the git commit hash that triggered the capture.
3. **Given** a stored session, **When** any authenticated project member accesses the session link, **Then** the structured conversation content is returned.
4. **Given** a commit occurs outside of an AI session (manual commit), **When** no active AI conversation exists, **Then** no session is captured and the commit proceeds without a session link.

---

### User Story 2 - AI Retrieves Past Sessions for Context Continuity (Priority: P2)

As an AI assistant working on a feature across multiple sessions, I can retrieve relevant past sessions to recall what the user decided, what additional context the user provided about the system, and what approaches were accepted or rejected — so I don't repeat questions or contradict earlier decisions.

Sessions are too large to embed in commit messages or local files. They must be stored in cloud storage and be queryable by the AI. When the AI starts a new task or session, it can look up past sessions for the same feature (or related tasks) to load prior context. This is the primary value driver — sessions become the AI's institutional memory for the project.

**Why this priority**: Without AI-consumable retrieval, stored sessions are just a passive archive. Making sessions queryable by the AI transforms them into an active knowledge base that improves AI quality over time.

**Independent Test**: Can be fully tested by storing sessions across multiple commits/tasks, then having the AI query for sessions related to a specific feature or task and verifying it receives relevant prior context.

**Acceptance Scenarios**:

1. **Given** multiple sessions exist for a feature, **When** the AI begins work on a new task in that feature, **Then** the AI can query and retrieve past sessions for the same feature to understand prior decisions and user-provided context.
2. **Given** a session where the user selected a specific approach (e.g., "use approach A, not B"), **When** the AI retrieves that session later, **Then** the user's decision is identifiable from the structured message data (user messages containing selections/preferences).
3. **Given** a session where the user provided additional system context (e.g., architecture details, constraints, domain knowledge), **When** the AI retrieves that session, **Then** the context is available to inform the AI's approach in the current session.
4. **Given** no prior sessions exist for a feature, **When** the AI queries for past context, **Then** the query returns empty results without error.

---

### User Story 3 - Capture Session During Beads Task Execution (Priority: P2)

As a developer using AI to execute tasks defined in beads, when the AI completes a task, the system captures the session for that task — including the AI's thinking, approach, and whether I accepted or rejected the proposed changes.

Each beads task execution is a discrete unit of work with its own conversation. Capturing these sessions provides a detailed audit trail of how each task was implemented and what decisions were made.

**Why this priority**: Beads task execution is a structured workflow where sessions are naturally bounded per task, making capture straightforward. This extends the P1 capability to the task-driven workflow.

**Independent Test**: Can be fully tested by executing a beads task via AI, completing it, and verifying the session is captured and linked to the specific task.

**Acceptance Scenarios**:

1. **Given** the AI is executing a beads task, **When** the task is completed (accepted by the user), **Then** the session for that task is captured and stored.
2. **Given** a beads task session has been stored, **When** the task record is updated in the database, **Then** it contains a link to the stored session.
3. **Given** the AI is executing a beads task, **When** the user rejects or abandons the task mid-session, **Then** the partial session is still captured and marked as "rejected" or "abandoned".

---

### User Story 4 - Retrieve and View Past Sessions (Priority: P3)

As a developer or team lead reviewing feature history, I can retrieve and read stored sessions associated with checkpoints or tasks to understand the reasoning behind past decisions.

**Why this priority**: Human retrieval is essential for review and audit, but depends on sessions being captured and stored first (P1) and is secondary to AI retrieval (P2).

**Independent Test**: Can be fully tested by storing a session and then retrieving it via its link, verifying the content is complete and readable.

**Acceptance Scenarios**:

1. **Given** a commit with an associated session, **When** an authenticated project member views the commit details, **Then** they can retrieve and read the structured session content.
2. **Given** multiple commits each with sessions, **When** a user views the feature commit history, **Then** each commit indicates whether it has an associated session and provides a link.
3. **Given** a session stored for a beads task, **When** a user views the task details, **Then** the session is accessible from the task record.

---

### User Story 5 - Authorized Session Access (Priority: P3)

As a project owner, I need sessions to be accessible only to authorized team members so that conversation data (which may contain proprietary reasoning or sensitive context) is not exposed to unauthorized users.

**Why this priority**: Authorization is a cross-cutting concern that secures the stored data. It builds on the existing Supabase authentication and row-level security already in place for the project.

**Independent Test**: Can be fully tested by storing a session as one user and attempting to access it as both an authorized and unauthorized user.

**Acceptance Scenarios**:

1. **Given** a stored session for a project, **When** an authenticated team member of that project requests the session, **Then** the session content is returned.
2. **Given** a stored session for a project, **When** an unauthenticated user or a user outside the project requests the session, **Then** access is denied.

---

### Edge Cases

- What happens when the AI session is extremely large (e.g., hours-long conversation with extensive code output)? The system should handle sessions up to a reasonable limit (e.g., 10 MB of text) and warn if a session exceeds this threshold.
- What happens when storage is temporarily unavailable during a commit? The session capture should not block the commit. The system should retry or queue the upload and warn the user if capture fails.
- What happens if a commit is amended (git commit --amend)? Since the commit hash changes, a new session should be captured for the new commit hash. The original session (linked to the old hash) becomes orphaned and should be retained for audit purposes.
- What happens when a user's authentication token expires during upload? The system should refresh the token transparently or prompt re-authentication without losing the captured session data.
- How are sessions handled for offline commits (no network connectivity)? Sessions should be cached locally and uploaded when connectivity is restored.
- What happens when a feature has dozens of prior sessions and the AI needs context? The system should return session metadata (summaries, decisions, key context) without requiring the AI to download every full session. Metadata-level queries should be lightweight; full content retrieval happens only for the most relevant sessions.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST automatically capture the AI conversation segment since the last commit in structured format (preserving individual messages with role, timestamp, and content) when a user commits changes.
- **FR-002**: System MUST capture the AI session for each beads task execution in structured format, including the AI's reasoning, proposed approach, and user's accept/reject decision.
- **FR-003**: System MUST store captured sessions in cloud storage and generate a retrievable link for each session.
- **FR-004**: System MUST associate the session link with the corresponding git commit hash (for checkpoint sessions) or task identifier (for beads task sessions) in the database.
- **FR-005**: System MUST enforce authorization so that any authenticated member of the project can access any stored session within that project. Users outside the project MUST be denied access.
- **FR-006**: Session capture MUST NOT block or delay the commit or task completion workflow — capture should happen asynchronously after the primary action succeeds.
- **FR-007**: System MUST handle capture failures gracefully by caching the session locally and retrying upload, notifying the user of any persistent failures.
- **FR-008**: System MUST support sessions of up to 10 MB in text size.
- **FR-009**: System MUST record session metadata including: author, timestamp, associated checkpoint/task ID, session status (complete, rejected, abandoned), and session size.
- **FR-010**: System MUST provide a way for users to retrieve the full session content given a session link or commit hash/task identifier.
- **FR-011**: System MUST support querying sessions by feature (all sessions for a given feature branch) and by task identifier, so the AI can load relevant prior context when starting new work.
- **FR-012**: System MUST store sessions in cloud storage (not in commit messages or local-only files), ensuring sessions are accessible from any machine and any future AI session with proper authentication.
- **FR-013**: Session metadata in the database MUST be queryable by feature name, commit hash, task ID, author, and date range to support both AI context loading and human browsing.

### Key Entities

- **Session**: A captured AI conversation segment in structured format (individual messages with role, timestamp, and content). Key attributes: unique identifier, author (user who created it), structured content (messages array), status (complete/rejected/abandoned), size, creation timestamp, storage link.
- **Checkpoint (Git Commit)**: A git commit on the feature branch that represents a unit of completed work. Not a separate database entity — identified by commit hash. A commit may or may not have an associated session depending on whether AI was involved.
- **Task (Beads)**: A discrete unit of implementation work. Key attributes: unique identifier (e.g., SL-xxx), associated feature, optional session link, execution status.
- **Session Metadata**: Queryable record stored in the database linking a session's storage location to its context. Key attributes: session ID, storage link, commit hash or task ID, feature name/branch, project ID, author ID, creation timestamp. Must be queryable by feature, task, commit hash, author, and date range to support AI context loading and human browsing.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of AI-assisted checkpoint commits automatically trigger session capture without user intervention.
- **SC-002**: 100% of beads task executions with AI involvement result in a captured and linked session.
- **SC-003**: Session capture adds no more than 5 seconds of delay to the commit or task completion workflow (asynchronous upload).
- **SC-004**: Stored sessions are retrievable by authorized users within 3 seconds.
- **SC-005**: Unauthorized access attempts to session data are blocked 100% of the time.
- **SC-006**: Failed session uploads are retried and succeed within 3 attempts for transient failures, with 95% eventual success rate.
- **SC-007**: Sessions up to 10 MB are stored and retrieved without errors.
- **SC-008**: AI can query and load all prior sessions for a feature in under 5 seconds, enabling context-aware continuation of work across sessions.

## Assumptions

- The project already uses Supabase for authentication and database (established in 008-cli-auth and 009-command-system-enhancements).
- Users are authenticated via the existing Supabase JWT-based auth flow before session capture occurs.
- Claude Code exposes or can be hooked into to extract the current conversation text at commit time. The mechanism for extracting conversation text will be researched during planning.
- Network connectivity is available for most session uploads; offline caching handles temporary disconnections.
- Session data is structured text (individual messages with role, timestamp, and content), not binary artifacts like screenshots or images.

### Previous work

- **008-cli-auth**: Established browser-based OAuth authentication for the CLI and credential storage at `~/.specledger/credentials.json`.
- **009-add-login-and-comment-commands**: Established slash-command authentication via token paste with session storage at `~/.specledger/session.json`, and Supabase REST API integration for comments.
- **Beads Issue Tracker**: Provides the task execution framework (`.beads/issues.jsonl`) where per-task sessions will be captured.
