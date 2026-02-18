# Feature Specification: Built-In Issue Tracker

**Feature Branch**: `591-issue-tracking-upgrade`
**Created**: 2026-02-18
**Status**: Draft
**Input**: User description: "Beads is slow, need to kill daemon. We don't need beads issue tracking, can be replaced with the sl binary. Usage of beads couple specledger with beads. Solution: Removed the beads + perles usage. Create simple issue tracker within sl CLI. Opensource CLI -> json file artifacts. Use backward comp beads artifact. Don't keep issue list at repo root, keep at spec level so there's less cross branch conflicts."

## Clarifications

### Session 2026-02-18

- Q: How should merge conflicts on issues.jsonl be resolved when merging branches? → A: Auto-merge with dedup (automatically deduplicate by issue ID, keeping both versions' changes where possible)
- Q: What priority scale should issues use? → A: Numeric 0-5 (0 = highest, 5 = lowest; matches Beads format)
- Q: What should happen when issue commands run outside a feature branch context? → A: Fail with error (clear message: "Not on a feature branch. Use --spec flag or checkout a ###-branch.")
- Q: Should issue IDs be scoped per spec or globally unique? → A: Globally unique across the repository using SHA-256 hash of (spec_context + title + created_at) to generate deterministic, collision-resistant IDs
- Q: What length should issue IDs be? → A: 6 characters (first 6 hex characters of SHA-256 digest) balancing brevity with collision resistance; collision probability < 0.01% for up to 100,000 issues
- Q: Should duplicate issues be detected when creating new issues? → A: Yes, `sl issue create` MUST check for semantically similar issues across all specs and warn user before creating; user can override with --force flag
- Q: Should definition of done be checked before closing issues? → A: Yes, definition of done is an optional field in the issue JSONL record. The `sl issue close` command and implement skills/command prompt MUST check this field before allowing closure; user can override with --force flag

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create and Manage Issues (Priority: P1)

As a developer using SpecLedger, I want to create, list, and manage issues directly from the sl CLI so that I can track work without external dependencies or slow daemon processes.

**Why this priority**: Issue creation and management is the core functionality. Without this, no other features matter. This replaces the primary use case of Beads.

**Independent Test**: Can be fully tested by creating an issue with `sl issue create`, listing it with `sl issue list`, updating it with `sl issue update`, and closing it with `sl issue close`. Delivers immediate value as a standalone issue tracker.

**Acceptance Scenarios**:

1. **Given** a spec directory exists at `specledger/010-my-feature/`, **When** I run `sl issue create --title "Add validation" --type task`, **Then** a new issue is created with a globally unique SHA-based ID (e.g., `SL-a3f5d8`) and stored in `specledger/010-my-feature/issues.jsonl`
2. **Given** issues exist in a spec directory, **When** I run `sl issue list`, **Then** all issues for the current spec are displayed in a readable table format
3. **Given** an issue exists with status "open", **When** I run `sl issue close ISSUE-ID`, **Then** the issue status changes to "closed" with a closed_at timestamp

---

### User Story 2 - Migrate Existing Beads Data (Priority: P2)

As a developer with existing Beads issues, I want to migrate my `.beads/issues.jsonl` data to the new format so that I don't lose historical work tracking.

**Why this priority**: Existing users need continuity. Without migration, users would lose all their issue history when upgrading.

**Independent Test**: Can be tested by running `sl issue migrate` on a repository with existing Beads data and verifying the new format files contain all the original data.

**Acceptance Scenarios**:

1. **Given** a `.beads/issues.jsonl` file exists at repository root, **When** I run `sl issue migrate`, **Then** issues are distributed to their respective spec directories based on branch/feature association
2. **Given** migration completes successfully, **When** I compare original Beads data with migrated data, **Then** all issue fields (id, title, description, status, priority, type, timestamps, dependencies) are preserved
3. **Given** an issue cannot be mapped to a spec directory, **When** migration runs, **Then** the issue is placed in a fallback location with a warning message

---

### User Story 3 - Track Dependencies Between Issues (Priority: P3)

As a developer planning complex features, I want to define dependencies between issues so that I can understand the order of work and identify blocking relationships.

**Why this priority**: Dependency tracking adds significant value for complex projects but is not required for basic issue management functionality.

**Independent Test**: Can be tested by creating two issues, linking them with `sl issue link ISSUE-A blocks ISSUE-B`, and verifying the dependency is reflected in both `sl issue list --tree` and JSON exports.

**Acceptance Scenarios**:

1. **Given** two issues exist in the same spec, **When** I run `sl issue link ISSUE-A blocks ISSUE-B`, **Then** ISSUE-B shows ISSUE-A in its `blocked_by` list and ISSUE-A shows ISSUE-B in its `blocks` list
2. **Given** issues with dependencies exist, **When** I run `sl issue list --tree`, **Then** issues are displayed in a tree structure showing parent-child and blocking relationships
3. **Given** a dependency would create a cycle, **When** I attempt to create it, **Then** the operation fails with a clear error message explaining the cycle

---

### User Story 4 - Work Across Multiple Specs (Priority: P3)

As a developer working on multiple features, I want to list and filter issues across all specs so that I can see my complete workload in one view.

**Why this priority**: Cross-spec visibility is valuable for project management but individual spec tracking is the primary use case.

**Independent Test**: Can be tested by creating issues in multiple spec directories and running `sl issue list --all` to see them aggregated.

**Acceptance Scenarios**:

1. **Given** issues exist in multiple spec directories, **When** I run `sl issue list --all`, **Then** all issues from all specs are listed with their spec context
2. **Given** issues across specs, **When** I run `sl issue list --all --status open`, **Then** only open issues from all specs are shown
3. **Given** issues across specs, **When** I run `sl issue list --all --type epic`, **Then** only epic-type issues are shown with their spec prefixes
4. **Given** I want to see issues from a specific spec, **When** I run `sl issue list --spec 010-my-feature`, **Then** only issues from that spec are displayed regardless of current branch context

---

### User Story 5 - Prevent Duplicate Issues (Priority: P2)

As a developer, I want the system to warn me about semantically similar issues when creating new issues so that I can avoid creating duplicates across specs.

**Why this priority**: Duplicate issues create confusion and waste effort; early detection prevents this problem.

**Independent Test**: Can be tested by creating an issue with a title similar to an existing issue and verifying the system warns about potential duplicates.

**Acceptance Scenarios**:

1. **Given** an issue with title "Add user authentication" exists in spec 010, **When** I run `sl issue create --title "Add user auth" --type task` in spec 020, **Then** the system warns about semantically similar issues and lists them
2. **Given** a duplicate warning is shown, **When** I run `sl issue create --title "Add user auth" --force`, **Then** the issue is created without warning
3. **Given** duplicate issues are detected, **When** I review the warnings, **Then** I can decide to link them or create a new issue

---

### User Story 6 - Enforce Definition of Done (Priority: P2)

As a developer, I want the system to check definition of done criteria before allowing me to close an issue so that quality standards are maintained.

**Why this priority**: Enforcing definition of done prevents incomplete work from being marked as done.

**Independent Test**: Can be tested by creating an issue with definition_of_done criteria and attempting to close it without meeting criteria.

**Acceptance Scenarios**:

1. **Given** an issue has a `definition_of_done` field with criteria, **When** I run `sl issue close ISSUE-ID`, **Then** the system checks each criterion and fails if any are not met
2. **Given** definition of done criteria are not met, **When** I run `sl issue close ISSUE-ID --force`, **Then** the issue is closed without checking criteria
3. **Given** all definition of done criteria are met, **When** I run `sl issue close ISSUE-ID`, **Then** the issue closes successfully with a note about which criteria were verified

---

### User Story 7 - Update Implement and Plan Skills/Prompts (Priority: P2)

As a SpecLedger maintainer, I want to update the implement and plan command prompts and skills to be aware of the issue tracking system so that developers can manage issues as part of their implementation workflow.

**Why this priority**: Integration with implement and plan commands ensures issues are checked before closing, enforcing quality standards and preventing incomplete work from being marked as done.

**Independent Test**: Can be tested by updating the implement and plan skill definitions to include issue tracking awareness, then verifying that the commands properly check definition_of_done fields and prevent closure of issues with unmet criteria.

**Acceptance Scenarios**:

1. **Given** the implement skill is updated with issue tracking awareness, **When** a developer attempts to close an issue via the implement command, **Then** the system checks the issue's `definition_of_done` field and prevents closure if criteria are not met
2. **Given** the plan command prompt is updated with issue context, **When** a developer runs `sl plan` in a spec with open issues, **Then** the prompt displays relevant open issues and their status
3. **Given** implement and plan skills are updated, **When** a developer uses these commands, **Then** they receive guidance on managing issues as part of their workflow without requiring external tools

---

### User Story 8 - Remove Beads and Perles Dependencies (Priority: P1)

As a SpecLedger maintainer, I want to remove all dependencies on Beads and Perles from the sl init bootstrap, prerequisite checks, and initialization logic so that SpecLedger operates as a standalone tool without external daemon dependencies.

**Why this priority**: Removing Beads and Perles dependencies is critical to achieving the core goal of eliminating slow daemon processes and external tool coupling. This is a prerequisite for the new issue tracking system to be the primary solution.

**Independent Test**: Can be tested by verifying that `sl init` completes successfully without requiring Beads or Perles to be installed, that prerequisite checks no longer reference these tools, and that all initialization logic uses only built-in functionality.

**Acceptance Scenarios**:

1. **Given** a fresh SpecLedger installation, **When** I run `sl init` in a new repository, **Then** the initialization completes successfully without checking for or requiring Beads or Perles installation
2. **Given** the sl CLI is running, **When** I check prerequisite validation, **Then** no checks reference Beads daemon status or Perles availability
3. **Given** existing bootstrap scripts reference Beads or Perles, **When** I review the codebase, **Then** all references are removed and replaced with native sl CLI functionality
4. **Given** a developer runs any sl command, **When** the command executes, **Then** no background Beads daemon is spawned or required for operation

---

### Edge Cases

- What happens when the spec directory doesn't exist when creating an issue? The command should fail with a helpful error message suggesting to run the spec initialization first.
- How does the system handle concurrent writes to issues.jsonl? File locking prevents corruption; operations queue if lock is held.
- What happens if issues.jsonl becomes corrupted? A `sl issue repair` command attempts to recover valid JSON lines and reports any unrecoverable data.
- How are issue IDs generated to avoid conflicts? IDs use globally unique format `SL-<6-char-hex>` derived from SHA-256 hash of (spec_context + title + created_at), ensuring deterministic and collision-resistant IDs across all specs without requiring a central counter. With 6 hex characters (16.7 million possible values), collision probability is < 0.01% for up to 100,000 issues.
- How are merge conflicts on issues.jsonl resolved? Auto-merge with deduplication by issue ID, preserving both branches' changes where possible.
- What happens if commands run outside a feature branch? Command fails with error: "Not on a feature branch. Use --spec flag or checkout a ###-branch."
- How are duplicate issues detected across different specs? Since issue IDs are globally unique and deterministic, identical issues in different specs will have the same ID. `sl issue create` checks for semantically similar issues using title/description hashing and warns user; `sl issue list --all --check-duplicates` can identify all potential duplicates for manual review.
- What happens if an issue in one spec conflicts with an issue in another spec? Since each spec has its own isolated issues.jsonl file and IDs are deterministically generated, there are no automatic conflicts between specs. Cross-spec issue conflicts must be manually resolved by the developer (e.g., merging duplicate issues, updating dependencies). The `sl issue list --all --check-duplicates` command helps identify such cases.
- What if two issues have identical (spec_context + title + created_at)? This is extremely unlikely due to timestamp precision (nanoseconds), but if it occurs, the system treats them as the same issue. Users should ensure unique titles or timestamps to avoid this edge case.
- What if an issue doesn't have a definition_of_done field? The system skips definition of done checks and allows issue closure without warnings.
- What if definition_of_done field is malformed? The system logs a warning and skips checks, allowing issue closure.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST store issues in JSONL format at `specledger/<spec-dir>/issues.jsonl` (per-spec storage)
- **FR-002**: System MUST support creating issues with fields: id, title, description, status, priority, issue_type, created_at, updated_at, spec_context, and optional definition_of_done
- **FR-003**: System MUST generate globally unique issue IDs using SHA-256 hash of (spec_context + title + created_at), formatted as `SL-<6-char-hex>` (first 6 characters of hex digest)
- **FR-004**: System MUST support issue types: epic, feature, task, bug
- **FR-005**: System MUST support issue statuses: open, in_progress, closed
- **FR-006**: System MUST support issue priority as numeric 0-5 (0 = highest, 5 = lowest)
- **FR-007**: System MUST provide `sl issue create` command with flags for --title, --description, --type, --priority, and --force
- **FR-008**: System MUST provide `sl issue list` command with optional --status, --type, --all, and --spec flags
- **FR-009**: System MUST provide `sl issue update` command to modify existing issue fields
- **FR-010**: System MUST provide `sl issue close` command with --force flag that sets status to closed and records closed_at timestamp
- **FR-011**: System MUST provide `sl issue show <id>` command to display full issue details
- **FR-012**: System MUST maintain backward compatibility with Beads JSONL format for migration purposes
- **FR-013**: System MUST provide `sl issue migrate` command to convert `.beads/issues.jsonl` to per-spec format with globally unique SHA-based IDs
- **FR-014**: System MUST support issue dependencies with fields: blocked_by, blocks
- **FR-015**: System MUST provide `sl issue link <id1> <relationship> <id2>` for dependency management
- **FR-016**: System MUST NOT require any daemon process or background service
- **FR-017**: System MUST complete all operations with file I/O only (no database required)
- **FR-018**: System MUST detect current spec context from git branch name (###-short-name pattern)
- **FR-019**: System MUST fail with error when no spec context detected and no --spec flag provided
- **FR-020**: System MUST auto-merge issues.jsonl conflicts by deduplicating on issue ID, preserving both branches' changes where possible
- **FR-021**: System MUST support `--spec <spec-name>` flag on `sl issue list` to filter issues from a specific spec directory
- **FR-022**: System MUST provide `sl issue list --all --check-duplicates` command to identify semantically similar issues across specs using title and description hashing
- **FR-023**: System MUST check for semantically similar issues when creating new issues and warn user; user can override with --force flag
- **FR-024**: System MUST read optional `definition_of_done` field from issue JSONL record and verify all criteria before allowing issue closure; user can override with --force flag
- **FR-025**: System MUST support definition_of_done field format with checklist items that can be verified programmatically or manually
- **FR-026**: Implement skills and command prompt MUST check definition_of_done field before closing issues and prevent closure if criteria are not met
- **FR-027**: Implement skill MUST be updated to include issue tracking awareness and enforce definition_of_done checks during issue closure
- **FR-028**: Plan command prompt MUST be updated to display relevant open issues and their status for the current spec context
- **FR-029**: `sl init` bootstrap MUST NOT check for or require Beads or Perles installation
- **FR-030**: All prerequisite validation checks MUST be updated to remove references to Beads daemon status and Perles availability
- **FR-031**: All initialization logic MUST use only native sl CLI functionality without spawning or requiring external Beads daemon processes

### Key Entities

- **Issue**: Core tracking unit with globally unique SHA-based ID, title, description, type (epic/feature/task/bug), status (open/in_progress/closed), priority (0-5, where 0=highest), spec_context, timestamps, optional definition_of_done field, and optional dependencies
- **IssueStore**: JSONL file at `specledger/<spec>/issues.jsonl` containing all issues for that spec
- **DefinitionOfDone**: Optional field within issue JSONL record containing checklist criteria for issue closure
- **Dependency**: Relationship between issues with type (blocks/blocked_by/parent-child) and direction

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Issue creation and listing commands complete in under 100ms for files with up to 1000 issues (no daemon overhead)
- **SC-002**: Migration of existing Beads data preserves 100% of issue data without manual intervention
- **SC-003**: Zero daemon processes required - all operations are direct file I/O
- **SC-004**: Issue storage at spec level eliminates cross-branch merge conflicts for typical workflows
- **SC-005**: Users can perform all common issue operations (create, list, update, close) without consulting documentation (intuitive CLI design)
- **SC-006**: New users can start tracking issues within 30 seconds of installing sl CLI
- **SC-007**: Issue IDs are deterministically generated with < 0.01% collision probability for up to 100,000 issues
- **SC-008**: Duplicate detection catches > 90% of semantically similar issues (title/description similarity)
- **SC-009**: Definition of done checks prevent closure of issues that don't meet criteria (100% enforcement with --force override)
- **SC-010**: Implement and plan commands are aware of issue tracking and guide developers through issue management workflows
- **SC-011**: `sl init` completes successfully without Beads or Perles dependencies, reducing setup time by eliminating daemon startup
- **SC-012**: All prerequisite checks pass on systems without Beads or Perles installed, confirming standalone operation

### Previous work

- **010-checkpoint-session-capture**: Established pattern for session capture via hooks; this feature will integrate with session tracking for task completion events
- **009-command-system-enhancements**: Established CLI command patterns in Go/Cobra that this feature will follow

## Dependencies & Assumptions

### Dependencies

- Go 1.24+ and Cobra CLI framework (existing)
- Git for branch detection (existing)
- SHA-256 hashing (standard library)
- String similarity algorithm for duplicate detection (e.g., Levenshtein distance or Jaro-Winkler)
- Implement and plan skill system (existing)

### Assumptions

- Users work primarily within a single spec/feature at a time, making per-spec storage natural
- Issue volume per spec is typically under 1000 issues, making JSONL performant
- Users accept migrating from Beads once to gain simplified architecture
- Standard JSONL line-append pattern is sufficient for concurrent write safety (no high-frequency concurrent writes expected)
- SHA-256 based IDs provide better collision resistance and determinism than sequential counters
- Timestamp precision (nanoseconds) is sufficient to prevent ID collisions for issues with identical spec_context and title
- 6-character hex IDs (16.7 million possible values) provide acceptable collision resistance for typical issue volumes
- Definition of done is optional; if not present in an issue record, no checks are performed
- Users will maintain definition_of_done field in a consistent format for reliable parsing
- Implement skills and command prompt will integrate definition_of_done checks before allowing issue closure
- Implement and plan skills can be updated to include issue tracking awareness without breaking existing functionality
- Beads and Perles are no longer required for SpecLedger operation after this feature is implemented
- Existing codebases using Beads can migrate to the new issue tracking system via `sl issue migrate`

### Out of Scope

- Real-time collaboration features (requires backend service)
- Issue synchronization across multiple machines (can be added later via git-based sync)
- Advanced querying beyond basic filters (can be added later)
- Integration with external issue trackers (Jira, Linear, GitHub Issues)
- Web UI for issue management
- Automatic definition of done criteria generation (must be manually created by team)
- Maintaining Beads/Perles compatibility beyond migration support
