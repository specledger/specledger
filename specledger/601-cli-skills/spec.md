# Feature Specification: New CLI Commands and Skills

**Feature Branch**: `601-cli-skills`
**Created**: 2026-03-03
**Status**: Draft
**Input**: User description: "create new specs for stream 3"

## Overview

Add new CLI commands and skills for comment management, research spikes, and implementation checkpoints. This is **Stream 3** of the 3-stream SDD alignment effort.

| Stream | Focus | Feature | Status |
|--------|-------|---------|--------|
| 1 | AI command consolidation | 599-alignment | Complete |
| 2 | Bash script → Go CLI migration | 600-bash-cli-migration | Planned |
| **3** | **New CLI + skills** | **601-cli-skills** | **This spec** |

**Stream 3 Scope** (this spec):
- Extract `sl comment` CLI from `sl revise` for granular comment management
- Create `sl-comment` skill for agent guidance
- Add `spike` AI command for exploratory research
- Add `checkpoint` AI command for implementation verification

**Components**:

| Component | Type | Purpose |
|-----------|------|---------|
| `sl comment` | CLI (L1) | Review comment CRUD operations |
| `sl-comment` | Skill (L3) | Teaches agent when/how to use `sl comment` |
| `spike` | AI Command (L2) | Time-boxed exploratory research |
| `checkpoint` | AI Command (L2) | Implementation verification + session log |

## User Scenarios & Testing *(mandatory)*

### User Story 1 - sl comment list Command (Priority: P1)

As an AI agent processing review feedback, I need `sl comment list` to get all unresolved comments so that I can understand what needs to be addressed without launching an interactive session.

**Why this priority**: Replaces `sl revise --summary` which is currently used by `/specledger.clarify`.

**Independent Test**: Run `sl comment list --json --status open` and verify JSON output with comment details.

**Acceptance Scenarios**:

1. **Given** a feature branch with unresolved comments, **When** running `sl comment list --json`, **Then** output is valid JSON array with id, file_path, line, content_preview (truncated to 120 chars), author, reply_count fields
2. **Given** `--status open` flag, **When** run, **Then** only unresolved comments are returned
3. **Given** `--status resolved` flag, **When** run, **Then** only resolved comments are returned
4. **Given** `--status all` flag, **When** run, **Then** all comments are returned
5. **Given** no auth token, **When** run, **Then** exits silently with code 1 (for agent integration)
6. **Given** compact mode (default), **When** run with 25 comments, **Then** output stays under ~500 tokens with truncated previews and reply counts instead of nested arrays (progressive disclosure for agent context efficiency)
7. **Given** compact mode output, **When** parsed by agent, **Then** footer hint guides drill-down: `"hint": "Use 'sl comment show <id> --json' for full content"`

---

### User Story 2 - sl comment show Command (Priority: P1)

As an AI agent addressing a specific comment, I need `sl comment show` to get full comment details including thread replies so that I can understand the complete context.

**Why this priority**: Required for agents to drill down into specific comments after listing.

**Independent Test**: Run `sl comment show <id> --json` and verify full comment content with thread replies.

**Acceptance Scenarios**:

1. **Given** a comment ID, **When** running `sl comment show <id> --json`, **Then** output includes full content (no truncation), selected_text, all thread replies — this is the drill-down command for progressive disclosure after `list` scan
2. **Given** multiple comment IDs, **When** running `sl comment show <id1> <id2>`, **Then** all comments are shown in sequence
3. **Given** a non-existent comment ID, **When** run, **Then** error with "comment not found" message
4. **Given** a comment with thread replies, **When** shown, **Then** replies are included in chronological order
5. **Given** a comment with 3 thread replies, **When** shown in JSON, **Then** output stays under ~200 tokens (full detail justified by explicit drill-down request)

---

### User Story 3 - sl comment reply/resolve Commands (Priority: P1)

As a developer addressing review feedback, I need `sl comment reply` and `sl comment resolve` to respond to and close comments so that reviewers can track progress.

**Why this priority**: Core CRUD operations needed for comment workflow.

**Independent Test**: Run `sl comment reply <id> "Addressed in commit abc123"` and verify reply is posted.

**Acceptance Scenarios**:

1. **Given** a comment ID and message, **When** running `sl comment reply <id> "message"`, **Then** reply is posted to the comment thread with minimal output (~30 tokens) for agent context efficiency
2. **Given** `--json` flag on reply, **When** run, **Then** output includes reply_id and timestamp
3. **Given** a comment ID, **When** running `sl comment resolve <id>`, **Then** comment is marked as resolved with minimal confirmation output
4. **Given** multiple comment IDs to resolve, **When** running `sl comment resolve <id1> <id2>`, **Then** all are resolved
5. **Given** resolving a parent comment, **When** run, **Then** all thread replies are also resolved (cascade)

---

### User Story 4 - sl-comment Skill (Priority: P2)

As an AI agent, I need the `sl-comment` skill to understand when and how to use `sl comment` commands so that I can manage review feedback efficiently.

**Why this priority**: Enhances agent capability but commands work without it.

**Independent Test**: Load skill and verify it documents all `sl comment` subcommands with usage patterns.

**Acceptance Scenarios**:

1. **Given** the skill is loaded, **When** agent needs to process comments, **Then** skill provides decision criteria for which subcommand to use
2. **Given** the skill content, **When** reviewed, **Then** it includes JSON parsing examples for `--json` output
3. **Given** the skill content, **When** reviewed, **Then** it documents when to use list vs show (compact vs full)
4. **Given** the skill content, **When** reviewed, **Then** it explains reply/resolve workflow for agents

---

### User Story 5 - spike AI Command (Priority: P2)

As a developer exploring a new technology, I need the `spike` command to run research so that findings are captured in a structured format.

**Why this priority**: Enables exploratory work but not required for core workflow.

**Independent Test**: Run `/specledger.spike "research auth patterns"` and verify research file created.

**Acceptance Scenarios**:

1. **Given** a research topic, **When** running `/specledger.spike "topic"`, **Then** research file is created at `specledger/<spec>/research/yyyy-mm-dd-topic.md`
2. **Given** spike output, **When** complete, **Then** file includes Findings, Decisions, and Recommendations sections
4. **Given** existing research files, **When** creating new spike, **Then** unique filename is generated (no overwrite)

---

### User Story 6 - checkpoint AI Command (Priority: P2)

As a developer implementing a feature, I need the `checkpoint` command to verify my progress and log the session so that I can track what's been done.

**Why this priority**: Supports implementation tracking but not required for core workflow.

**Independent Test**: Run `/specledger.checkpoint` and verify session log updated with progress.

**Acceptance Scenarios**:

1. **Given** in-progress tasks, **When** running `/specledger.checkpoint`, **Then** session log is updated with completed items
2. **Given** the command, **When** run, **Then** it verifies tests pass for completed work (runs `go test ./...` for modified packages, requires exit 0)
3. **Given** uncommitted changes, **When** run, **Then** it prompts to commit or notes pending changes
4. **Given** checkpoint output, **When** complete, **Then** summary shows what was accomplished this session

---

### Edge Cases

- What if Supabase API is unavailable? → `sl comment` commands return network error with retry hint
- What if comment was deleted by another user? → `sl comment show` returns "comment not found"
- What if checkpoint finds no progress? → Reports "no changes since last checkpoint"
- What if multiple agents reply to same comment? → Both replies are preserved in thread order

## Requirements *(mandatory)*

### Token-Efficient Output Pattern (D21)

CLI commands optimized for AI agent consumption minimize token usage:

**Compact Mode (default for `sl comment list`)**:
- Previews truncated to 80 characters with "..." ellipsis
- Shows count instead of full data (e.g., "3 replies" instead of reply content)
- One comment per line in table format
- Example: `abc123 | src/main.go:42 | Fix error handling... | alice | 3 replies`

**Full Mode (`--json` flag or `sl comment show`)**:
- Complete data with no truncation
- All thread replies included
- Structured JSON for programmatic parsing

### Functional Requirements

- **FR-001**: `sl comment list` MUST output JSON array with comment details when `--json` flag is set
- **FR-002**: `sl comment list` MUST support `--status` filter (open/resolved/all)
- **FR-003**: `sl comment list` MUST exit silently with code 1 on auth failure (for agent integration)
- **FR-004**: `sl comment show` MUST include full comment content and all thread replies
- **FR-005**: `sl comment show` MUST accept multiple IDs for batch retrieval
- **FR-006**: `sl comment reply` MUST post a reply to an existing comment thread
- **FR-007**: `sl comment resolve` MUST mark comments as resolved
- **FR-008**: `sl comment resolve` MUST cascade to thread replies when parent is resolved
- **FR-009**: All `sl comment` subcommands defined in this spec (list, show, reply, resolve) MUST support `--json` output flag
- **FR-010**: `sl-comment` skill MUST document when to use each subcommand
- **FR-011**: `/specledger.spike` MUST create timestamped research files at `specledger/<spec>/research/yyyy-mm-dd-<topic>.md`
- **FR-012**: `/specledger.checkpoint` MUST update session logs with progress (session log format: `.specledger/sessions/<spec>-session.md`)

### Key Entities

- **ReviewComment**: A review comment from Supabase
  - id, file_path, line, content, selected_text, author, status, created_at
- **ThreadReply**: A reply in a comment thread
  - id, parent_id, content, author, created_at
- **SpikeReport**: Research spike output
  - topic, findings, decisions, recommendations, created_at
- **CheckpointLog**: Session progress record
  - tasks_completed, tests_status, uncommitted_changes, timestamp
  - Location: `.specledger/sessions/<spec>-session.md`
  - Format: Markdown with timestamped entries

### Non-Functional Requirements

- **NFR-001**: All `sl comment` commands MUST complete in <2s (measured as P95 CLI execution time excluding network RTT to Supabase)
- **NFR-002**: Token-efficient output MUST be enforced for compact mode (80 char truncation, counts over full data)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 4 new `sl comment` subcommands available (list, show, reply, resolve) ✅ DONE (CLI implemented)
- **SC-002**: `sl comment list --json` produces valid JSON parseable by agents ✅ DONE
- **SC-003**: `/specledger.clarify` successfully uses `sl comment list` instead of `sl revise --summary` ✅ DONE
- **SC-004**: `sl-comment` skill loads and provides actionable guidance ✅ DONE
- **SC-005**: `/specledger.spike` creates structured research files ✅ DONE
- **SC-006**: `/specledger.checkpoint` tracks session progress ✅ DONE

### Implementation Notes

**Completed:**
- CLI: `sl comment list/show/reply/resolve` implemented in `pkg/cli/commands/comment.go`
- Skill: `sl-comment` skill deployed to `.claude/skills/sl-comment/`
- AI Command: `/specledger.spike` created at `.claude/commands/specledger.spike.md`
- AI Command: `/specledger.checkpoint` created at `.claude/commands/specledger.checkpoint.md`
- Updated: `specledger.clarify.md` to use `sl comment list --status open --json`

### Previous work

### Epic: 598 - SDD Workflow Streamline

- **D4**: Comment Management decision - extract `sl comment` CLI + `sl-comment` skill
- **D13**: Spike command for exploratory research
- **D14**: Checkpoint command for implementation verification
- **D21**: Token-efficient output pattern for CLI commands

### Epic: 136 - Revise Comments

- PostgREST client with auth retry (`pkg/cli/revise/client.go`) - to be extracted
- Comment types (`pkg/cli/revise/types.go`) - to be extracted
- Comment resolution with cascade - already implemented

## Dependencies & Assumptions

### Dependencies

- **Supabase PostgREST**: For review_comments table (already in use)
- **599-alignment**: `/specledger.clarify` updated to use `sl comment list`
- **136-revise-comments**: Existing comment infrastructure to extract

### Assumptions

- PostgREST API structure remains unchanged
- Auth flow (`sl auth login`) is already working
- Agent launcher pattern is stable

## Out of Scope

- **Stream 1**: AI command consolidation (599-alignment)
- **Stream 2**: Bash script migration (600-bash-cli-migration)
- TUI features for comment management
- Bulk comment import/export
