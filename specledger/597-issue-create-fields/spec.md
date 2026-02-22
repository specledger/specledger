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

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Issues with Complete Field Set (Priority: P1)

As a developer using SpecLedger, I want to create issues with all available fields (acceptance criteria, definition of done, design notes) in a single command so that I don't have to update issues immediately after creation.

**Why this priority**: This is the core functionality gap. Users currently cannot set acceptance criteria, definition of done, design notes, or notes during issue creation, forcing a two-step workflow (create then update). Fixing this enables the full potential of the issue model.

**Independent Test**: Can be fully tested by creating an issue with all new flags and verifying each field is persisted correctly in the JSONL file. Delivers immediate value by eliminating the create-then-update pattern.

**Acceptance Scenarios**:

1. **Given** a user wants to create a fully-specified issue, **When** they run `sl issue create --title "Task" --acceptance-criteria "AC text" --type task`, **Then** the created issue has the acceptance_criteria field populated
2. **Given** a user wants to create an issue with definition of done items, **When** they run `sl issue create --title "Task" --dod "Item 1" --dod "Item 2" --type task`, **Then** the created issue has definition_of_done.items containing both items as unchecked
3. **Given** a user wants to create an issue with design notes, **When** they run `sl issue create --title "Task" --design "Design approach details" --type task`, **Then** the created issue has the design field populated
4. **Given** a user wants to create an issue with notes, **When** they run `sl issue create --title "Task" --notes "Implementation notes" --type task`, **Then** the created issue has the notes field populated
5. **Given** a user creates an issue with all new fields combined, **When** they run `sl issue show` on the created issue, **Then** all fields are displayed correctly

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

### User Story 4 - Tasks Prompt Utilizes Issue Fields (Priority: P3)

As a developer running `/specledger.tasks`, I want the command to utilize the new CLI flags (acceptance-criteria, definition-of-done, design) when creating issues so that generated tasks have structured, queryable data.

**Why this priority**: This enhances the task generation quality but depends on US1 being implemented first.

**Independent Test**: Can be tested by running `/specledger.tasks` and verifying generated issues use the new flags correctly instead of embedding everything in description.

**Acceptance Scenarios**:

1. **Given** the updated tasks prompt, **When** issues are created, **Then** acceptance criteria is set via `--acceptance-criteria` flag, not embedded in description
2. **Given** the updated tasks prompt, **When** issues are created with DoD items, **Then** DoD is set via multiple `--dod` flags, not embedded in description
3. **Given** the updated tasks prompt, **When** issues are created with design notes, **Then** design is set via `--design` flag
4. **Given** the updated tasks prompt, **When** `sl issue show` displays the issue, **Then** acceptance criteria, DoD items, and design are shown in dedicated sections

---

### User Story 5 - Implement Prompt Utilizes DoD and Acceptance Criteria (Priority: P3)

As a developer running `/specledger.implement`, I want the command to check off Definition of Done items as implementation progresses and verify work against acceptance criteria so that issues accurately reflect completion status.

**Why this priority**: This enhances the implementation workflow by automatically tracking progress through DoD items and ensuring work meets defined acceptance criteria. Depends on US1 and US2 being implemented first.

**Independent Test**: Can be tested by running `/specledger.implement` and verifying that as each task phase completes, corresponding DoD items are marked as checked and acceptance criteria is verified.

**Acceptance Scenarios**:

1. **Given** the updated implement prompt, **When** a task is completed, **Then** the agent uses `sl issue update SL-xxxxx --check-dod "Item text"` to mark relevant DoD items as checked
2. **Given** an issue with multiple DoD items, **When** implementation progresses, **Then** each completed subtask results in the corresponding DoD item being checked
3. **Given** an issue with acceptance_criteria, **When** the agent begins implementation, **Then** the agent reads the acceptance_criteria to understand requirements
4. **Given** the agent completes a task, **When** verifying work, **Then** the agent confirms the implementation satisfies the acceptance_criteria before marking complete
5. **Given** the implement prompt completes a task, **When** all DoD items are checked, **Then** the issue status can be changed to closed via `sl issue close`
6. **Given** the implement prompt, **When** reviewing progress, **Then** `sl issue show` displays which DoD items are checked with verified_at timestamps

---

### Edge Cases

- What happens when `--dod` is provided without `--title`? The existing validation should catch missing required title.
- What happens when `--dod` items contain special characters or newlines? Shell escaping should be handled properly.
- What happens when updating DoD on a closed issue? Allow the update but don't change status.
- What happens when all DoD items are checked via `--check-dod`? The issue should still require explicit close via `sl issue close`.
- What happens when circular dependencies are detected during task generation? Existing cycle detection should prevent creation.
- What happens when `--check-dod` is called with text that doesn't match any DoD item? Return error: "DoD item not found: '<text>'"
- What happens when `--uncheck-dod` is called on an already unchecked item? Return error: "DoD item not found: '<text>'" (same behavior as non-existent)

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
- **FR-011**: The specledger.tasks prompt MUST be updated to instruct using new CLI flags instead of embedding in description
- **FR-012**: The specledger.implement prompt MUST be updated to read acceptance_criteria at task start and verify implementation against it before completion
- **FR-013**: The specledger.implement prompt MUST be updated to use `sl issue update --check-dod` when completing subtasks that correspond to DoD items
- **FR-014**: Task generation MUST create proper blocking relationships: setup blocks implementation, models block services, services block endpoints
- **FR-015**: Task generation MUST NOT create false blocking relationships between parallelizable tasks (different files, no dependencies)
- **FR-016**: Feature-type issues for phases MUST have appropriate blocking: foundational phases block user story phases, but phases at the same level should not block each other unless specified

### Key Entities

- **Issue**: Existing model with fields being exposed: acceptance_criteria (string), definition_of_done (struct with items array), design (string), notes (string)
- **IssueUpdate**: Existing struct needs new fields for DoD operations: DefinitionOfDone replacement, CheckDoDItem, UncheckDoDItem

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create fully-specified issues with all fields in a single command (no create-then-update needed)
- **SC-002**: Generated task issues have verifiable blocking relationships using `sl issue show --tree` or `sl issue ready`
- **SC-003**: `sl issue show` displays all fields in organized, readable format with dedicated sections
- **SC-004**: Task generation creates no more false-positive blocking relationships between independent tasks
- **SC-005**: Implementation workflow automatically checks off DoD items as subtasks complete
- **SC-006**: All existing functionality remains unchanged (backward compatibility)

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
