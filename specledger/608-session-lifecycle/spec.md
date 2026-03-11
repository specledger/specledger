# Feature Specification: Session Lifecycle Management

**Feature Branch**: `608-session-lifecycle`
**Created**: 2026-03-11
**Status**: Draft
**Input**: User description: "https://github.com/specledger/specledger/issues/51"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Prune Old Sessions (Priority: P1)

A developer's project has accumulated many session recordings over weeks of work. Storage costs are growing, and old sessions are no longer useful. The developer wants to remove sessions older than a threshold to reclaim storage and keep their session list clean.

**Why this priority**: Storage management is the most pressing operational need. Without pruning, session data grows indefinitely, incurring cloud storage costs and cluttering the session list.

**Independent Test**: Can be tested by creating sessions with backdated timestamps, running the prune command, and verifying old sessions are removed while recent ones are preserved.

**Acceptance Scenarios**:

1. **Given** sessions exist spanning the last 60 days, **When** the user runs the prune command with a 30-day threshold, **Then** sessions older than 30 days are deleted from both cloud storage and metadata index.
2. **Given** the user wants to preview what would be deleted, **When** the user runs prune with a dry-run option, **Then** the system lists sessions that would be deleted without actually removing them.
3. **Given** the user is not authenticated, **When** they attempt to prune, **Then** the system reports that authentication is required and suggests logging in.
4. **Given** no sessions exceed the age threshold, **When** the user runs prune, **Then** the system reports that no sessions need pruning.

---

### User Story 2 - Configure Session TTL (Priority: P1)

A project maintainer wants to set a default retention period for sessions so that expired sessions can be pruned automatically based on project policy. This setting should be part of the project configuration and apply consistently across team members.

**Why this priority**: TTL configuration powers the prune functionality and establishes organizational policy for session retention.

**Independent Test**: Can be tested by setting a TTL value in project configuration, then verifying the prune command respects it when using the "expired" mode.

**Acceptance Scenarios**:

1. **Given** no TTL is configured, **When** the system checks session retention policy, **Then** it uses a default of 30 days.
2. **Given** a TTL of 14 days is configured in the project settings, **When** the user runs prune in expired mode, **Then** sessions older than 14 days are targeted for deletion.
3. **Given** a TTL is configured, **When** the user runs prune with an explicit age threshold, **Then** the explicit threshold overrides the configured TTL.

---

### User Story 3 - Tag Sessions for Organization (Priority: P2)

A developer wants to categorize sessions with tags to make them easier to find later. Tags should be automatically derived from branch naming patterns (e.g., "feature/auth" → tag "auth") and also assignable manually. Sessions can then be filtered by tag.

**Why this priority**: Tagging improves discoverability and organization of sessions but is not required for core session management operations.

**Independent Test**: Can be tested by capturing sessions with tags, then filtering the session list by tag to verify only matching sessions appear.

**Acceptance Scenarios**:

1. **Given** a session is captured on branch `feature/user-auth`, **When** the session is stored, **Then** it is automatically tagged with "user-auth" derived from the branch name.
2. **Given** a user wants to add a custom tag, **When** they capture a session with a tag flag, **Then** the tag is stored alongside auto-generated tags.
3. **Given** tagged sessions exist, **When** the user lists sessions filtered by a specific tag, **Then** only sessions with that tag are displayed.
4. **Given** a session has multiple tags, **When** the user filters by any one of those tags, **Then** the session appears in the results.

---

### User Story 4 - View Session Usage Statistics (Priority: P2)

A developer or team lead wants to see aggregate metrics about session usage — total count, storage consumed, distribution by branch, and messaging patterns. This helps understand usage patterns and justify storage costs.

**Why this priority**: Statistics provide visibility but are not essential for day-to-day operations. They help with capacity planning and cost justification.

**Independent Test**: Can be tested by populating sessions with known data, running the stats command, and verifying computed metrics match expected values.

**Acceptance Scenarios**:

1. **Given** sessions exist for the current project, **When** the user runs the stats command, **Then** aggregate metrics are displayed including total count, total storage size, and date range.
2. **Given** sessions span multiple branches, **When** the user views statistics, **Then** per-branch distribution is shown (session count and storage per branch).
3. **Given** no sessions exist, **When** the user runs stats, **Then** the system reports that no sessions are available.

---

### User Story 5 - Query Sessions Across Projects (Priority: P3)

A developer working across multiple repositories wants to list sessions from all their projects in one view. This helps them find context from past sessions regardless of which project they were working on.

**Why this priority**: Cross-project queries are a convenience feature. Most developers work within a single project at a time, and this addresses an occasional need.

**Independent Test**: Can be tested by having sessions in multiple projects, running the list command with the all-projects flag, and verifying sessions from all projects appear.

**Acceptance Scenarios**:

1. **Given** sessions exist across multiple projects, **When** the user lists sessions with an all-projects flag, **Then** sessions from all accessible projects are returned.
2. **Given** the user has no authentication, **When** they attempt cross-project queries, **Then** the system reports that authentication is required.

---

### Edge Cases

- What happens when pruning encounters a network error mid-deletion?
  - The prune operation continues with remaining sessions, reports partial results with a count of successful and failed deletions.
- What happens when a session has both auto-generated and manually assigned tags that overlap?
  - Duplicate tags are deduplicated — each unique tag appears only once.
- What happens when the TTL is set to 0?
  - A TTL of 0 means no automatic expiry — sessions are retained indefinitely until manually pruned.
- What happens when pruning is run while offline?
  - The system reports that pruning requires network access (since it removes cloud storage) and suggests trying again when connected.
- What happens when the stats command encounters sessions with missing metadata?
  - Missing fields are excluded from calculations with a warning, and available data is still reported.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a command to delete sessions older than a specified age threshold.
- **FR-002**: The prune command MUST support a dry-run mode that lists sessions that would be deleted without removing them.
- **FR-003**: Pruning MUST remove both the cloud storage object and the metadata index entry for each deleted session.
- **FR-004**: The system MUST support a configurable session TTL (time-to-live) in the project settings, defaulting to 30 days.
- **FR-005**: Pruning in "expired" mode MUST use the configured TTL to determine which sessions to delete.
- **FR-006**: Sessions MUST support a tags field containing zero or more text labels.
- **FR-007**: Tags MUST be auto-populated from branch naming patterns when a session is captured (e.g., extracting meaningful segments from branch names).
- **FR-008**: Users MUST be able to add custom tags to sessions via a command-line flag during capture.
- **FR-009**: The session list command MUST support filtering by tag.
- **FR-010**: The system MUST provide a stats command showing aggregate session metrics: total count, total storage size, per-branch distribution, message count averages, and temporal range.
- **FR-011**: The session list command MUST support an all-projects flag to query sessions across multiple projects.
- **FR-012**: Pruning MUST require user authentication since it modifies cloud storage.
- **FR-013**: Offline queue entries MUST respect the configured TTL — expired queued sessions should be discarded rather than uploaded.

### Key Entities

- **Session**: An existing entity representing a captured AI conversation transcript. Extended with a tags field (list of text labels) for categorization.
- **Session TTL Configuration**: A project-level setting defining the maximum age for session retention, stored in the project configuration file.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can prune expired sessions in under 10 seconds for up to 100 sessions, with clear feedback on what was deleted.
- **SC-002**: The dry-run mode accurately reports which sessions would be deleted without making any changes.
- **SC-003**: TTL configuration takes effect immediately — pruning in expired mode deletes all sessions exceeding the configured TTL.
- **SC-004**: Users can filter sessions by tag and see results in under 2 seconds for up to 500 sessions.
- **SC-005**: The stats command produces accurate aggregate metrics in under 5 seconds for up to 1000 sessions.
- **SC-006**: Cross-project session listing returns results from all accessible projects in a single view.

### Previous work

### Epic: 010-checkpoint-session-capture

- **SL-af6926 - Checkpoint Session Capture**: Built the foundational session capture pipeline — types, storage, metadata, compression, and CLI commands (`sl session capture/list/get`).
- **SL-2076bb - Implement Supabase Storage client**: Created the cloud storage integration for session files.
- **SL-552de1 - Implement session metadata PostgREST client**: Built the metadata API client for querying and creating session records.
- **SL-737069 - Implement offline session queue**: Added offline queue for capturing sessions without connectivity.

### Epic: 602-silent-session-capture

- **SL-482b2c - Silent Session Capture**: Improved capture to work silently without stderr noise for unauthenticated users.

## Dependencies

- **010-checkpoint-session-capture**: The existing session capture infrastructure (types, storage client, metadata client) must be in place. This feature extends those components.
- **Authentication system**: Pruning and cross-project queries require a working authentication flow (`sl auth login`).

## Assumptions

- The existing sessions database table can be altered to add a `tags` column without breaking existing functionality.
- Branch naming patterns follow common conventions (e.g., `feature/xxx`, `fix/xxx`, `608-session-lifecycle`) where meaningful segments can be extracted for auto-tagging.
- The default TTL of 30 days is appropriate for most use cases; this can be overridden per project.
- Session pruning operates on cloud data; local offline queue entries are handled separately by discarding expired entries before upload.
- The stats command computes metrics from the cloud index (metadata), not by downloading and analyzing session content.
