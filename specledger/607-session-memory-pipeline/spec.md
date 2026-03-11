# Feature Specification: Session-to-Knowledge Memory Pipeline

**Feature Branch**: `607-session-memory-pipeline`
**Created**: 2026-03-11
**Status**: Draft
**Input**: User description: "https://github.com/specledger/specledger/issues/52"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - AI Agent Retrieves Relevant Knowledge Before Starting Work (Priority: P1)

An AI agent begins a new session on a feature. Before writing code, the agent automatically receives relevant knowledge extracted from previous sessions — patterns learned, decisions made, mistakes to avoid, and conventions established. This knowledge helps the agent make better decisions without the user having to manually re-explain past context.

**Why this priority**: This is the core value proposition. Without knowledge retrieval, the system has no user-facing benefit. Everything else (extraction, scoring, storage) exists to serve this moment.

**Independent Test**: Can be tested by pre-populating a knowledge cache file, starting a new session, and verifying the agent receives the cached knowledge as context.

**Acceptance Scenarios**:

1. **Given** cached knowledge entries exist for the current project, **When** an AI agent starts a new session, **Then** relevant knowledge is injected into the agent's context automatically.
2. **Given** no cached knowledge exists, **When** an AI agent starts a session, **Then** the session proceeds normally with no errors or delays.
3. **Given** cached knowledge contains project-specific conventions and cross-project patterns, **When** the agent receives context, **Then** both types of knowledge are included with clear categorization.

---

### User Story 2 - Extract Knowledge from Session Transcripts (Priority: P1)

After a productive session, a user or automated process extracts structured knowledge from the session transcript. The system identifies patterns, decisions, debugging insights, and conventions from the raw conversation and produces structured knowledge entries.

**Why this priority**: Without extraction, there is no knowledge to retrieve. This is the input pipeline that feeds US1.

**Independent Test**: Can be tested by providing a session transcript and verifying the system produces structured knowledge entries with appropriate tags and scores.

**Acceptance Scenarios**:

1. **Given** a completed session transcript, **When** the user runs the knowledge extraction command, **Then** structured entries are produced with titles, descriptions, tags, and relevance scores.
2. **Given** a session transcript with debugging insights, **When** knowledge is extracted, **Then** the debugging pattern is captured with enough detail to be useful in future sessions.
3. **Given** a session with no notable patterns or decisions, **When** knowledge is extracted, **Then** the system reports that no significant knowledge was found (rather than creating low-quality entries).

---

### User Story 3 - Score and Promote High-Value Knowledge (Priority: P2)

Knowledge entries are scored on three axes — how often the pattern recurs (Recurrence), how much it affects outcomes (Impact), and how specific vs. generic it is (Specificity). Entries scoring above a threshold are automatically promoted to the project's persistent knowledge base, while lower-scoring entries remain as candidates for future re-evaluation.

**Why this priority**: Scoring prevents knowledge bloat. Without quality filtering, the knowledge base would grow unchecked and lose signal-to-noise ratio over time.

**Independent Test**: Can be tested by creating entries with known characteristics, running the scoring system, and verifying that high-value entries are promoted while low-value ones are not.

**Acceptance Scenarios**:

1. **Given** a knowledge entry that recurs across 3+ sessions, impacts architecture decisions, and is specific to the project, **When** scored, **Then** the composite score exceeds the promotion threshold (7.0/10).
2. **Given** a generic, one-time observation with low impact, **When** scored, **Then** the composite score falls below the threshold and the entry is not promoted.
3. **Given** the promotion threshold is met, **When** an entry is promoted, **Then** it appears in the project's persistent knowledge file accessible to future sessions.

---

### User Story 4 - View and Manage Knowledge Entries (Priority: P2)

A user wants to review what knowledge the system has extracted, see scoring details, and manually curate entries (promote, demote, or delete). This provides transparency and user control over what the AI agent will learn from.

**Why this priority**: Users need visibility and control over automated knowledge extraction. Without this, the system operates as a black box.

**Independent Test**: Can be tested by populating knowledge entries and verifying the CLI displays them with scores, tags, and management options.

**Acceptance Scenarios**:

1. **Given** knowledge entries exist for the current project, **When** the user runs the knowledge viewing command, **Then** entries are displayed with titles, scores, tags, and source session references.
2. **Given** a low-scoring entry the user considers valuable, **When** the user manually promotes it, **Then** the entry is added to the persistent knowledge base regardless of score.
3. **Given** a promoted entry the user considers outdated, **When** the user removes it, **Then** the entry is removed from the persistent knowledge base.

---

### User Story 5 - Sync Knowledge with Cloud Storage (Priority: P3)

Knowledge entries are indexed in the cloud so they persist across machines and can be shared across team members working on the same project. The cloud serves as the index, while local cached files provide fast access during sessions.

**Why this priority**: Cloud sync is valuable for team collaboration but not essential for single-developer workflows. The core pipeline (extract → score → retrieve) works locally without cloud.

**Independent Test**: Can be tested by creating local knowledge entries, running the sync command, and verifying entries appear in the cloud index, then pulling from another machine.

**Acceptance Scenarios**:

1. **Given** locally promoted knowledge entries exist, **When** the user syncs to the cloud, **Then** entries are indexed in the cloud storage with project association.
2. **Given** knowledge exists in the cloud index, **When** the user pulls knowledge on a new machine, **Then** cached files are generated locally for fast agent access.
3. **Given** no network connectivity, **When** the user works with knowledge, **Then** all operations work locally with sync deferred until connectivity is restored.

---

### Edge Cases

- What happens when two sessions extract contradictory knowledge (e.g., "use pattern A" vs. "don't use pattern A")?
  - The newer entry takes precedence. Both are stored with timestamps, and the knowledge display command shows the conflict for manual resolution.
- What happens when the knowledge cache grows very large (hundreds of entries)?
  - Only entries above the scoring threshold are included in agent context. Lower-scoring entries are archived but not injected.
- What happens when knowledge is extracted from a session that was abandoned or contained errors?
  - The extraction process produces candidates, not promoted knowledge. Manual review or scoring prevents bad knowledge from being promoted.
- What happens when the same pattern is extracted from multiple sessions?
  - The Recurrence score increases, making it more likely to be promoted. Duplicate entries are merged rather than duplicated.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST extract structured knowledge entries from session transcripts, including title, description, tags, and source session reference.
- **FR-002**: Each knowledge entry MUST be scored on three axes: Recurrence (0-10), Impact (0-10), and Specificity (0-10), producing a composite score.
- **FR-003**: Knowledge entries with a composite score of 7.0 or above MUST be automatically promoted to the project's persistent knowledge base.
- **FR-004**: The system MUST provide a command to view all knowledge entries with their scores, tags, and promotion status.
- **FR-005**: Users MUST be able to manually promote, demote, or delete knowledge entries.
- **FR-006**: Promoted knowledge MUST be automatically available to AI agents at session start as injected context.
- **FR-007**: Knowledge entries MUST be stored locally as cached files for offline access and fast retrieval.
- **FR-008**: The system MUST provide a sync mechanism to index knowledge entries in cloud storage for cross-machine and team access.
- **FR-009**: The extraction process MUST handle sessions with no notable patterns gracefully, reporting that no knowledge was found rather than creating low-quality entries.
- **FR-010**: Duplicate knowledge patterns across sessions MUST be merged, increasing the Recurrence score rather than creating separate entries.

### Key Entities

- **Knowledge Entry**: A structured piece of organizational knowledge with title, description, tags, source session, scores (Recurrence, Impact, Specificity), composite score, and promotion status.
- **Knowledge Base**: The persistent collection of promoted knowledge entries for a project, stored as a local file and optionally synced to cloud.
- **Session Transcript**: The raw conversation log from an AI session, serving as the input source for knowledge extraction.
- **Knowledge Cache**: Local file storage for knowledge entries, providing fast access for agent context injection.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: AI agents starting new sessions receive relevant project knowledge within 2 seconds, reducing the need for users to re-explain past context.
- **SC-002**: Knowledge extraction from a typical session (50-200 messages) completes within 30 seconds and produces 0-5 structured entries.
- **SC-003**: The scoring system correctly identifies high-value patterns with at least 80% agreement with manual human evaluation.
- **SC-004**: The knowledge base maintains a signal-to-noise ratio where 90%+ of promoted entries are considered useful by the user.
- **SC-005**: Users can view, promote, demote, and delete knowledge entries in under 10 seconds per operation.
- **SC-006**: Cloud sync completes within 5 seconds for up to 100 knowledge entries.

### Previous work

### Epic: 010-checkpoint-session-capture

- **SL-af6926 - Checkpoint Session Capture**: Built the session capture infrastructure that records AI sessions on commit, providing the raw transcripts that serve as input for knowledge extraction.
- **SL-2076bb - Implement Supabase Storage client**: Established the Supabase storage integration used for cloud-indexed content.
- **SL-552de1 - Implement session metadata PostgREST client**: Created the metadata API client for session records.
- **SL-2b9ae9 - Implement sl session get command**: Added the ability to retrieve past sessions, which this feature builds upon.

### Epic: 602-silent-session-capture

- **SL-482b2c - Silent Session Capture**: Improved session capture to work silently in the background, ensuring transcripts are available without disrupting user workflow.

### Epic: 600-bash-cli-migration

- **SL-1ab18f - US4: sl context update Command**: Created the context update mechanism that this feature extends to inject knowledge into agent context.

## Dependencies

- **Issue #51 (Session lifecycle management)**: Requires session tagging support for associating knowledge entries with their source sessions.
- **Existing 4-layer architecture**: The feature follows the established L0 (hooks), L1 (CLI), L2 (AI commands), L3 (skills) architecture pattern.

## Assumptions

- Session transcripts are already being captured and stored (via the checkpoint session capture feature).
- The three-axis scoring system (Recurrence, Impact, Specificity) with a 7.0 composite threshold provides adequate quality filtering. This threshold may need tuning based on real-world usage.
- Knowledge extraction requires AI processing (LLM) to identify patterns from unstructured conversation transcripts.
- Local-first storage with optional cloud sync is the right architecture — the system must work fully offline.
- The `.specledger/memory/cache/` directory is an appropriate location for local knowledge files, consistent with the existing `.specledger/` structure.
- Knowledge entries are scoped to a project (repository), not globally shared across all projects.
