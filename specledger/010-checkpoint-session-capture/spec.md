# Feature Specification: Checkpoint Session Capture

**Feature Branch**: `010-checkpoint-session-capture`
**Created**: 2026-02-11
**Status**: Draft

## Problem Statement

When using AI assistants for software development, there is no persistent record of the conversations that led to code changes. This creates several problems:

1. **No audit trail**: Reviewers cannot see what AI assistance was provided or what reasoning led to each commit
2. **Lost knowledge**: Decisions, trade-offs, and rejected approaches discussed with AI are not captured
3. **No retrospective data**: Teams cannot analyze AI-assisted development patterns or improve their workflows

### What This Feature Does NOT Do

- ❌ **Context continuity for AI** - Loading past sessions to "remember" user preferences is out of scope
- ❌ **Real-time collaboration** - Sessions are captured post-hoc, not streamed
- ❌ **Code analysis** - We capture conversations, not code quality metrics

### Success Looks Like

A developer commits code → The AI conversation (since last commit) is automatically captured → Anyone on the team can later retrieve and review that session to understand the reasoning behind the commit.

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

### User Story 2 - Team Reviews Sessions for Retrospectives (Priority: P2)

As a team lead or developer, I can browse and search past sessions to understand patterns in AI-assisted development, identify recurring issues, and improve team workflows.

**Why this priority**: After capturing sessions (P1), the next value is enabling humans to review and analyze them for process improvement.

**Independent Test**: Can be fully tested by storing multiple sessions, then querying by feature/date/author and verifying results are filterable and readable.

**Acceptance Scenarios**:

1. **Given** multiple sessions exist for a feature, **When** a team member queries by feature branch, **Then** all sessions for that feature are returned with metadata (commit, date, author, size).
2. **Given** sessions from multiple authors, **When** a team lead filters by author, **Then** only that author's sessions are returned.
3. **Given** sessions across a date range, **When** a user filters by date, **Then** only sessions within that range are returned.

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

---

## Implementation Design Decisions

*Added 2026-02-23: Documents technical decisions made during implementation.*

### Approach Selection: Claude Code Hooks

**Options Considered:**

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **Git Hook (post-commit)** | Works with any git workflow, simple | No transcript access, can't differentiate AI commits | Rejected |
| **Claude Code Hook (PostToolUse)** | Direct transcript access, session ID, filters for commits | Only works with Claude Code | **Selected** |
| **File Watcher on Transcript** | Independent of hooks | Complex, race conditions, no clear trigger | Rejected |

**Rationale:** Claude Code hooks provide direct access to the conversation transcript via `transcript_path` in the hook JSON. This is the only approach that gives us the actual conversation content.

### Hook Format Discovery

Initial implementation assumed a `tool_success: bool` field based on expected schema. Testing revealed the actual format:

**Assumed:**
```json
{"tool_success": true, "tool_output": "..."}
```

**Actual (discovered via debug capture):**
```json
{
  "tool_response": {
    "stdout": "...",
    "stderr": "",
    "interrupted": false
  },
  "tool_use_id": "toolu_..."
}
```

**Fix:** Updated `HookInput` struct in `types.go` to match actual format, added `ToolSuccess()` method that checks `!tool_response.interrupted`.

### Trigger Condition: Git Commit Only

**Decision:** Only capture when Bash tool executes `git commit` (excluding `--amend`).

**Implementation:**
```go
var gitCommitPattern = regexp.MustCompile(`^\s*git\s+commit(\s|$)`)
var gitAmendPattern = regexp.MustCompile(`\s+--amend\b`)
```

**Rationale:**
- Commits represent logical checkpoints
- Amends modify history (would create duplicate captures)
- Other git commands (add, push, status) don't represent new work

**Known Limitation:** Excluding `--amend` may miss legitimate work if developers use amend as their primary workflow. This is a deliberate trade-off to avoid duplicate session captures when amending typos or minor fixes. May revisit if this causes issues.

### Delta Capture (Incremental)

**Decision:** Only capture messages since last capture for this session.

**Implementation:**
- Track byte offset in transcript file per session
- Store state in `~/.specledger/session-state.json`
- Read only new JSONL lines from offset

**Rationale:**
- Avoids duplicate storage
- Reduces storage costs
- Each commit gets only relevant context

### Project ID Auto-Lookup

**Decision:** Auto-lookup `project.id` from Supabase when running `sl init`.

**Implementation:**
1. Parse git remote URL (SSH or HTTPS format)
2. Extract `owner/repo`
3. Query Supabase: `GET /projects?repo_owner=eq.{owner}&repo_name=eq.{repo}`
4. Store in `specledger.yaml`

**Rationale:**
- Reduces manual configuration errors
- Git remote uniquely identifies project
- Seamless setup experience

### Storage Architecture

```
Supabase Storage: sessions/{project_id}/{branch}/{commit_hash}.json.gz
PostgreSQL:       sessions table (metadata for querying)
```

**Rationale:**
- Compressed JSON reduces storage costs (~10x compression)
- Metadata separation enables fast queries without downloading content
- Standard patterns for cloud storage

### Offline Queue

**Decision:** Queue failed uploads locally for retry with `sl session sync`.

**Location:** `~/.specledger/session-queue/`

**Rationale:**
- Network failures shouldn't block commits
- Developer workflow must not be interrupted
- Eventual consistency acceptable for session data

---

## Validation & Testing

### Unit Tests (`pkg/cli/session/capture_test.go`)

| Test | What It Validates |
|------|-------------------|
| `TestIsGitCommit` | Correctly identifies git commit commands vs other commands |
| `TestToolInputCommand` | Parses tool_input JSON correctly (object and string formats) |
| `TestToolSuccess` | Correctly interprets tool_response.interrupted field |
| `TestParseHookInput` | Full hook JSON parsing with all fields |

### Manual Validation Steps

**Pre-requisite:** Project initialized with `sl init`, authenticated with `sl auth login`

1. **Verify hook is configured:**
   ```bash
   cat .claude/settings.json | jq '.hooks.PostToolUse'
   # Should show matcher: "Bash" with command: "sl session capture"
   ```

2. **Test capture trigger in Claude Code:**
   ```bash
   # In Claude Code, make a commit
   git commit -m "test"
   # Check session was captured
   sl session list --limit 1
   # Should show the new commit
   ```

3. **Verify delta capture:**
   ```bash
   # Make another commit in same Claude Code session
   git commit --allow-empty -m "second commit"
   sl session list --limit 2
   # Second session should have fewer messages (delta only)
   ```

4. **Verify session content:**
   ```bash
   sl session get <commit-hash> --json | jq '.messages | length'
   # Should return message count > 0
   ```

### What Is NOT Automatically Tested

- E2E test requiring actual Claude Code session (requires manual testing)
- Network failure scenarios (queue/retry)
- Large session handling (>10MB)
