# Feature Specification: Issue Create Fields Enhancement

**Feature Branch**: `597-issue-create-fields`
**Created**: 2026-02-22
**Status**: Draft
**Input**: User description: "Enhance sl issue command with acceptance criteria, definition of done, and other JSONL-supported fields; update specledger.tasks prompt to utilize these fields; improve task blocking tree relations for feature types"

## Clarifications

### Session 2026-02-22

- Q: How should multiple Definition of Done items be provided via the CLI flag? → A: Repeated flags (e.g., `--dod "Item 1" --dod "Item 2"`) using Cobra StringArray pattern
- Q: How should --check-dod and --uncheck-dod match DoD item text? → A: Exact match including case and whitespace (no normalization)
- Q: What should happen when --check-dod is called with text that doesn't match any DoD item? → A: Return error with clear message: "DoD item not found: 'text'"
- Q: Should --notes flag be included in scope alongside the 3 structured fields? → A: Yes, include --notes in US1 acceptance scenarios

### Session 2026-02-24

- Q: Should parent-child relationships be added similar to Beads? → A: Yes, add --parent flag with single parent constraint, update prompts to utilize parent-child relationships
- Q: What should be the maximum parent-child hierarchy depth? → A: Unlimited (no depth restriction, unlike Beads' 3-level limit)
- Q: When displaying children in `sl issue show --tree`, how should they be ordered? → A: Priority then ID (higher priority first, then by creation order)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Issues with Complete Field Set (Priority: P1)

As a developer using SpecLedger, I want to create issues with all available structured fields (`--acceptance-criteria`, `--dod`, `--design`, `--notes`) in a single command so that I don't have to update issues immediately after creation.

**Why this priority**: This is the core functionality gap. Users currently cannot set acceptance criteria, definition of done, design notes, or notes during issue creation, forcing a two-step workflow (create then update). Fixing this enables the full potential of the issue model.

**Independent Test**: Can be fully tested by creating an issue with all 4 new flags and verifying each field is persisted correctly in the JSONL file. Delivers immediate value by eliminating the create-then-update pattern.

**Acceptance Scenarios**:

1. **Given** a user wants to create a fully-specified issue, **When** they run `sl issue create --title "Task" --acceptance-criteria "AC text" --type task`, **Then** the created issue has the acceptance_criteria field populated
2. **Given** a user wants to create an issue with definition of done items, **When** they run `sl issue create --title "Task" --dod "Item 1" --dod "Item 2" --type task`, **Then** the created issue has definition_of_done.items containing both items as unchecked
3. **Given** a user wants to create an issue with design notes, **When** they run `sl issue create --title "Task" --design "Design approach details" --type task`, **Then** the created issue has the design field populated
4. **Given** a user wants to create an issue with implementation notes, **When** they run `sl issue create --title "Task" --notes "Additional context" --type task`, **Then** the created issue has the notes field populated
5. **Given** a user creates an issue with all 4 new fields combined, **When** they run `sl issue create --title "Task" --acceptance-criteria "AC" --dod "DoD1" --design "Design" --notes "Notes" --type task`, **Then** all 4 fields are persisted correctly
6. **Given** a user creates an issue with all new fields, **When** they run `sl issue show` on the created issue, **Then** acceptance_criteria, definition_of_done, design, and notes are displayed in dedicated sections

---

### User Story 2 - Update Definition of Done on Existing Issues (Priority: P2)

As a developer, I want to update the definition of done on existing issues so that I can refine task requirements as understanding evolves.

**Why this priority**: This complements US1 by allowing iterative refinement. Less critical than initial creation but essential for workflow completeness.

**Independent Test**: Can be tested by creating an issue, then updating its DoD items, and verifying the changes are persisted and displayed correctly.

**Acceptance Scenarios**:

1. **Given** an existing issue without DoD, **When** user runs `sl issue update SL-xxxxx --dod "New item 1" --dod "New item 2"`, **Then** the issue's definition_of_done is populated with both items as unchecked
2. **Given** an existing issue with DoD items, **When** user runs `sl issue update SL-xxxxx --dod "Replacement item"`, **Then** the previous DoD items are replaced with the new item
3. **Given** an existing issue, **When** user runs `sl issue update SL-xxxxx --check-dod "Item text"`, **Then** that specific DoD item is marked as checked with a verified_at timestamp
4. **Given** an existing issue with checked DoD item, **When** user runs `sl issue update SL-xxxxx --uncheck-dod "Item text"`, **Then** that DoD item is unchecked and verified_at is cleared
5. **Given** an existing issue, **When** user runs `sl issue update SL-xxxxx --check-dod "Nonexistent item"`, **Then** the command returns an error "DoD item not found: 'Nonexistent item'"

---

### User Story 3 - Tasks Generated with Proper Blocking Relations (Priority: P2)

As a developer running `/specledger.tasks`, I want generated task issues to have correct blocking relationships so that the dependency tree accurately reflects implementation order and parallel work opportunities.

**Why this priority**: Proper blocking relations are essential for the task generation workflow to be useful. Without correct dependencies, teams cannot identify parallelizable work or understand what must be done first.

**Independent Test**: Can be tested by running task generation and verifying the created issues have correct BlockedBy/Blocks relationships using `sl issue show --tree`.

**Acceptance Scenarios**:

1. **Given** a feature with setup tasks and implementation tasks, **When** tasks are generated via `/specledger.tasks`, **Then** implementation tasks are blocked by their prerequisite setup tasks
2. **Given** a user story with models before services before endpoints, **When** tasks are generated, **Then** service tasks are blocked by model tasks, and endpoint tasks are blocked by service tasks
3. **Given** tasks within the same user story phase, **When** tasks are generated, **Then** parallelizable tasks (different files) are NOT blocked by each other
4. **Given** feature-type issues for phases, **When** the epic is created, **Then** phase features are NOT blocked by each other (unless explicitly specified)
5. **Given** a foundational phase with multiple tasks, **When** tasks are generated, **Then** all user story tasks are blocked by the foundational feature issue

---

### User Story 4 - Tasks Prompt Utilizes All Three Structured Fields (Priority: P3)

As a developer running `/specledger.tasks`, I want the command to utilize all three structured field flags (`--acceptance-criteria`, `--dod`, `--design`) when creating issues so that generated tasks have structured, queryable data.

**Why this priority**: This enhances the task generation quality but depends on US1 being implemented first.

**Independent Test**: Can be tested by running `/specledger.tasks` and verifying generated issues use all 3 new flags correctly instead of embedding everything in description.

**Acceptance Scenarios**:

1. **Given** the updated tasks prompt, **When** issues are created, **Then** acceptance criteria is set via `--acceptance-criteria` flag, not embedded in description
2. **Given** the updated tasks prompt, **When** issues are created with DoD items, **Then** DoD is set via multiple `--dod` flags, not embedded in description
3. **Given** the updated tasks prompt, **When** issues are created with design notes, **Then** design is set via `--design` flag, not embedded in description
4. **Given** the updated tasks prompt, **When** issues are created with all 3 fields, **Then** `sl issue show` displays acceptance_criteria, definition_of_done, and design in dedicated sections
5. **Given** the updated tasks prompt, **When** task generation creates feature issues, **Then** design field contains technical approach derived from plan.md

---

### User Story 5 - Implement Prompt Utilizes All Three Structured Fields (Priority: P3)

As a developer running `/specledger.implement`, I want the command to utilize all three structured fields (design, acceptance_criteria, definition_of_done) so that implementation follows the design approach, meets acceptance criteria, and tracks progress through DoD items.

**Why this priority**: This enhances the implementation workflow by providing technical guidance, verification criteria, and progress tracking. Depends on US1 and US2 being implemented first.

**Independent Test**: Can be tested by running `/specledger.implement` and verifying that the agent reads design for approach, verifies against acceptance criteria, and checks off DoD items as work progresses.

**Acceptance Scenarios**:

1. **Given** an issue with design field, **When** the agent begins implementation, **Then** the agent reads the design field to understand the technical approach
2. **Given** an issue with acceptance_criteria, **When** the agent begins implementation, **Then** the agent reads the acceptance_criteria to understand requirements
3. **Given** the agent completes a task, **When** verifying work, **Then** the agent confirms the implementation satisfies the acceptance_criteria before marking complete
4. **Given** the updated implement prompt, **When** a subtask is completed, **Then** the agent uses `sl issue update SL-xxxxx --check-dod "Item text"` to mark relevant DoD items as checked
5. **Given** an issue with multiple DoD items, **When** implementation progresses, **Then** each completed subtask results in the corresponding DoD item being checked
6. **Given** the implement prompt completes a task, **When** all DoD items are checked, **Then** the issue status can be changed to closed via `sl issue close`
7. **Given** the implement prompt, **When** reviewing progress, **Then** `sl issue show` displays which DoD items are checked with verified_at timestamps

---

### User Story 6 - Parent-Child Relationships for Task Hierarchy (Priority: P2)

As a developer using SpecLedger, I want to set parent-child relationships between issues so that I can organize tasks into hierarchies (epic → feature → task) and query child issues efficiently.

**Why this priority**: Parent-child relationships are essential for task organization. Similar to Beads, a task should have only ONE parent to maintain a clean hierarchy. This enables better task breakdown and progress tracking.

**Independent Test**: Can be tested by creating an epic, creating child features with --parent flag, and verifying the hierarchy with `sl issue show --tree`.

**Acceptance Scenarios**:

1. **Given** a user wants to create a child issue, **When** they run `sl issue create --title "Subtask" --parent SL-abc123 --type task`, **Then** the created issue has parentId field set to SL-abc123
2. **Given** an existing issue without a parent, **When** user runs `sl issue update SL-xyz789 --parent SL-abc123`, **Then** the issue's parentId is set to SL-abc123
3. **Given** an issue with an existing parent, **When** user runs `sl issue update SL-xyz789 --parent SL-def456`, **Then** the command returns an error "issue already has a parent, remove existing parent first"
4. **Given** an issue with a parent, **When** user runs `sl issue update SL-xyz789 --parent ""`, **Then** the issue's parentId is cleared
5. **Given** a user tries to set self as parent, **When** user runs `sl issue update SL-xyz789 --parent SL-xyz789`, **Then** the command returns an error "cannot set self as parent"
6. **Given** an issue with parentId set, **When** user runs `sl issue show SL-xyz789`, **Then** the output displays the parent relationship
7. **Given** a parent issue, **When** user runs `sl issue show SL-abc123 --tree`, **Then** child issues are displayed under the parent in tree format
8. **Given** the updated tasks prompt, **When** tasks are generated, **Then** task issues are created with --parent flag pointing to their phase feature issue

---

### Edge Cases

- What happens when `--dod` is provided without `--title`? The existing validation should catch missing required title.
- What happens when `--dod` items contain special characters or newlines? Shell escaping should be handled properly.
- What happens when updating DoD on a closed issue? Allow the update but don't change status.
- What happens when all DoD items are checked via `--check-dod`? The issue should still require explicit close via `sl issue close`.
- What happens when circular dependencies are detected during task generation? Existing cycle detection should prevent creation.
- What happens when `--check-dod` is called with text that doesn't match any DoD item? Return error: "DoD item not found: '<text>'"
- What happens when `--uncheck-dod` is called on an already unchecked item? Return error: "DoD item not found: '<text>'" (same behavior as non-existent)
- What happens when setting parent on an issue that already has a parent? Return error: "issue already has a parent, remove existing parent first"
- What happens when setting self as parent? Return error: "cannot set self as parent"
- What happens when setting parent creates a circular relationship (A→B→A)? Return error: "circular parent-child relationship detected"
- What happens when setting parent to a non-existent issue? Return error: "parent issue not found: <id>"
- What happens when --parent "" is called on an issue without a parent? Silently succeed (idempotent operation)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept `--acceptance-criteria` flag on `sl issue create` to set the acceptance_criteria field
- **FR-002**: System MUST accept repeated `--dod` flags on `sl issue create` (Cobra StringArray pattern) to populate definition_of_done.items array with unchecked items
- **FR-003**: System MUST accept `--design` flag on `sl issue create` to set the design field
- **FR-004**: System MUST accept `--notes` flag on `sl issue create` to set the notes field
- **FR-005**: System MUST accept `--dod` flag on `sl issue update` to replace the entire definition_of_done
- **FR-006**: System MUST accept `--check-dod` flag on `sl issue update` to mark a specific DoD item as checked with timestamp
- **FR-007**: System MUST accept `--uncheck-dod` flag on `sl issue update` to mark a specific DoD item as unchecked
- **FR-008**: DoD item matching for --check-dod and --uncheck-dod MUST be exact (case-sensitive, no whitespace normalization)
- **FR-009**: System MUST return a clear error when --check-dod or --uncheck-dod is called with text that doesn't match any existing DoD item
- **FR-010**: System MUST display acceptance_criteria, definition_of_done, and design in `sl issue show` output in dedicated sections
- **FR-011**: The specledger.tasks prompt MUST be updated to use `--acceptance-criteria`, `--dod`, and `--design` flags instead of embedding in description
- **FR-012**: The specledger.tasks prompt MUST populate design field with technical approach derived from plan.md when creating feature issues
- **FR-013**: The specledger.implement prompt MUST read the design field at task start to understand the technical approach
- **FR-014**: The specledger.implement prompt MUST read acceptance_criteria at task start and verify implementation against it before completion
- **FR-015**: The specledger.implement prompt MUST use `sl issue update --check-dod` when completing subtasks that correspond to DoD items
- **FR-016**: Task generation MUST create proper blocking relationships: setup blocks implementation, models block services, services block endpoints
- **FR-017**: Task generation MUST NOT create false blocking relationships between parallelizable tasks (different files, no dependencies)
- **FR-018**: Feature-type issues for phases MUST have appropriate blocking: foundational phases block user story phases, but phases at the same level should not block each other unless specified
- **FR-019**: System MUST accept `--parent` flag on `sl issue create` to set the parentId field
- **FR-020**: System MUST accept `--parent` flag on `sl issue update` to set or clear the parentId field
- **FR-021**: System MUST enforce single parent constraint - an issue can only have ONE parent; attempting to set a second parent MUST return error "issue already has a parent, remove existing parent first"
- **FR-022**: System MUST prevent setting self as parent - attempting to set an issue as its own parent MUST return error "cannot set self as parent"
- **FR-023**: System MUST prevent circular parent-child relationships (A parent of B, B parent of A)
- **FR-024**: `sl issue show` MUST display parent relationship when parentId is set
- **FR-025**: `sl issue show --tree` MUST display child issues under their parent in tree format
- **FR-026**: The specledger.tasks prompt MUST be updated to use `--parent` flag when creating task issues, setting parent to the phase feature issue
- **FR-027**: The specledger.tasks prompt MUST instruct agents to create proper parent-child hierarchies: epic → feature (phase) → task
- **FR-028**: Parent-child hierarchy depth MUST be unlimited (no maximum depth restriction)
- **FR-029**: Children in `sl issue show --tree` MUST be ordered by priority (higher priority first), then by creation order (ID)

### Key Entities

- **Issue**: Existing model with fields being exposed: acceptance_criteria (string), definition_of_done (struct with items array), design (string), notes (string), parentId (string pointer - NEW)
- **IssueUpdate**: Existing struct needs new fields for DoD operations: DefinitionOfDone replacement, CheckDoDItem, UncheckDoDItem, ParentID (string pointer - NEW)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create issues with all 4 fields (`--acceptance-criteria`, `--dod`, `--design`, `--notes`) in a single command
- **SC-002**: Generated task issues have verifiable blocking relationships using `sl issue show --tree` or `sl issue ready`
- **SC-003**: `sl issue show` displays all 4 fields in organized, readable format with dedicated sections
- **SC-004**: Task generation creates no more false-positive blocking relationships between independent tasks
- **SC-005**: Task generation populates design field with technical approach from plan.md
- **SC-006**: Implementation workflow reads design field for technical guidance
- **SC-007**: Implementation workflow verifies work against acceptance_criteria before completion
- **SC-008**: Implementation workflow automatically checks off DoD items as subtasks complete
- **SC-009**: All existing functionality remains unchanged (backward compatibility)
- **SC-010**: Users can set parent on issue create and update with single parent constraint enforced
- **SC-011**: `sl issue show --tree` displays parent-child hierarchy correctly
- **SC-012**: Task generation creates proper parent-child hierarchies (epic → feature → task)
- **SC-013**: Circular parent-child relationships are prevented with clear error message

### Previous work

### Epic: 591-issue-tracking-upgrade - Issue Tracking System Upgrade

- **Issue Model**: Already supports acceptance_criteria, definition_of_done, design, notes fields in JSONL
- **Issue Update**: Already supports --acceptance-criteria, --notes, --design flags
- **Issue Show**: Already displays definition_of_done, but missing dedicated display for acceptance_criteria and design

### Epic: 595-issue-tree-ready - Issue Tree and Ready Commands

- **Dependency Tree**: Provides tree visualization with `--tree` flag
- **Ready Command**: Shows issues ready to work on (not blocked)
- **Blocking Detection**: Existing IsReady() logic checks all blockers are closed

## Dependencies & Assumptions

### Dependencies

- Issue model fields already exist and are validated
- IssueUpdate struct already has AcceptanceCriteria, DefinitionOfDone, CheckDoDItem, UncheckDoDItem fields

### Assumptions

- Multiple `--dod` flags use Cobra StringArray pattern (repeated flags: `--dod "Item 1" --dod "Item 2"`)
- Shell escaping for complex DoD items is the user's responsibility (standard CLI behavior)
- Existing validation (title required, priority range, etc.) applies to enhanced create
- DoD items are matched by exact text (case-sensitive, no whitespace normalization) for check/uncheck operations
- Parent-child hierarchy has no maximum depth (unlimited nesting allowed)
- Children are displayed ordered by priority (descending), then by creation order (ascending ID)
