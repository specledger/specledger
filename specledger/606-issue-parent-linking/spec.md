# Feature Specification: Improve Issue Parent-Child Linking

**Feature Branch**: `606-issue-parent-linking`
**Created**: 2026-03-10
**Status**: Draft
**Input**: User description: "https://github.com/specledger/specledger/issues/58"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Link Parent via Issue Link Command (Priority: P1)

A user or AI agent has already created a batch of issues but forgot to set `--parent` during creation. They want to establish parent-child relationships after the fact using the familiar `sl issue link` command syntax, rather than running `sl issue update --parent` for each one.

**Why this priority**: This is the most impactful gap. The `sl issue link` command already supports `blocks`/`blocked_by` link types, but not `parent`. Adding `parent` as a link type creates a consistent, intuitive interface for all relationship types and matches how AI agents naturally try to establish hierarchies.

**Independent Test**: Can be fully tested by creating an epic and a task, then running `sl issue link <task> parent <epic>` and verifying the parent-child relationship is established.

**Acceptance Scenarios**:

1. **Given** a task issue and an epic issue exist, **When** the user runs `sl issue link <task-id> parent <epic-id>`, **Then** the task's parent is set to the epic and the relationship is visible in `sl issue show`.
2. **Given** a task already has a parent, **When** the user runs `sl issue link <task-id> parent <new-parent-id>`, **Then** the task's parent is updated to the new parent.
3. **Given** a valid task ID, **When** the user runs `sl issue link <task-id> parent <nonexistent-id>`, **Then** an error is displayed indicating the parent issue was not found.

---

### User Story 2 - Warn About Orphaned Issues (Priority: P2)

After generating a batch of issues, a user wants to quickly identify which tasks and features are missing parent relationships so they can fix them. The system should surface orphaned issues — tasks or features without a parent — as warnings.

**Why this priority**: During large task generation sessions (e.g., 39 issues), it's easy to miss parent assignments. Proactive warnings help users catch and fix hierarchy gaps before they cause confusion in tree views.

**Independent Test**: Can be fully tested by creating several issues (some with parents, some without), then running a command that surfaces the orphaned ones.

**Acceptance Scenarios**:

1. **Given** a spec has 10 issues where 3 tasks lack parents, **When** the user runs `sl issue list --orphaned`, **Then** only the 3 orphaned tasks are displayed.
2. **Given** all tasks and features have parents, **When** the user runs `sl issue list --orphaned`, **Then** no issues are listed and a success message is shown.
3. **Given** an epic issue has no parent, **When** the user runs `sl issue list --orphaned`, **Then** the epic is NOT shown (epics are root-level by nature).

---

### User Story 3 - Bulk Reparent Issues (Priority: P3)

A user has multiple orphaned issues that all belong under the same parent. They want to assign a parent to multiple issues in a single command rather than running individual update commands for each one.

**Why this priority**: This is a convenience feature that saves significant time when fixing hierarchy issues after bulk creation (e.g., the observed case of 29 separate update commands).

**Independent Test**: Can be fully tested by creating multiple orphaned tasks, running the bulk reparent command with a target parent, and verifying all specified issues now have the correct parent.

**Acceptance Scenarios**:

1. **Given** 5 orphaned task issues exist, **When** the user runs `sl issue reparent <parent-id> <task1> <task2> <task3> <task4> <task5>`, **Then** all 5 tasks have their parent set to the specified parent.
2. **Given** a mix of valid and invalid issue IDs, **When** the user runs the reparent command, **Then** valid issues are reparented and errors are reported for invalid IDs without stopping the operation.
3. **Given** some tasks already have parents, **When** the user runs the reparent command, **Then** existing parents are overwritten with the new parent.

---

### User Story 4 - AI Agent Skill Instructions Updated (Priority: P1)

AI agents generating issues via `/specledger.tasks` consistently set the `--parent` flag when creating features and tasks. The skill instructions clearly emphasize parent-child hierarchy as a required part of issue creation.

**Why this priority**: Prevention is better than cure. The root cause of the observed problem was that AI agents weren't instructed strongly enough to use `--parent`. Updated instructions prevent the issue from recurring.

**Independent Test**: Can be verified by reviewing the skill instructions and confirming they contain explicit guidance about using `--parent` for all non-epic issues.

**Acceptance Scenarios**:

1. **Given** an AI agent runs `/specledger.tasks`, **When** it creates feature issues, **Then** each feature issue includes `--parent <epic-id>`.
2. **Given** an AI agent runs `/specledger.tasks`, **When** it creates task issues, **Then** each task issue includes `--parent <feature-id>`.
3. **Given** the skill instructions are loaded, **When** an AI reads the task creation section, **Then** the instructions explicitly state that `--parent` is required for all non-epic issues.

---

### Edge Cases

- What happens when a user tries to create a circular parent relationship (A is parent of B, B is parent of A)?
  - The system should detect and reject circular parent chains with a clear error message.
- What happens when a parent issue is closed but its children are still open?
  - Children remain valid; closing a parent does not cascade to children.
- What happens when reparenting an issue that has blocking dependencies on its current siblings?
  - Dependencies are independent of parent-child hierarchy; reparenting should not affect `blocks`/`blocked_by` relationships.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `sl issue link` command MUST support `parent` as a link type, setting the parent-child relationship between two issues.
- **FR-002**: The system MUST validate that the target parent issue exists before establishing the relationship.
- **FR-003**: The system MUST detect and reject circular parent-child relationships.
- **FR-004**: The `sl issue list` command MUST support an `--orphaned` flag to display non-epic issues that lack a parent within the current spec context.
- **FR-005**: A bulk reparent command MUST allow setting the same parent for multiple issues in a single operation.
- **FR-006**: The bulk reparent command MUST continue processing remaining issues when individual items fail, reporting errors without halting.
- **FR-007**: AI agent skill instructions for task generation MUST explicitly require `--parent` for all non-epic issue creation.

### Key Entities

- **Issue**: An existing entity with `parentId` field that establishes hierarchy. Types include epic, feature, and task.
- **Link Type**: The relationship type used in `sl issue link`. Currently supports `blocks`/`blocked_by`. This feature adds `parent` as a new link type.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can establish parent-child relationships using `sl issue link` in a single command, reducing the steps needed compared to `sl issue update --parent`.
- **SC-002**: Orphaned issues (non-epic issues without parents) are identifiable with a single command, achieving 100% detection accuracy.
- **SC-003**: Bulk reparenting of N issues completes in a single command invocation instead of N separate commands.
- **SC-004**: AI-generated task batches have 100% parent-child coverage (zero orphaned tasks/features) when following updated skill instructions.

### Previous work

### Epic: 591-issue-tracking-upgrade

- **SL-963e8c - Define Issue entity and types**: Defined the Issue entity with `parentId` field and type hierarchy (epic, feature, task).
- **SL-17a714 - Implement sl issue create command**: Implemented the create command including `--parent` flag support.

### Epic: SL-00d561 - Issue Tree View and Ready Command (595-issue-tree-ready)

- **SL-939c5d - US1: View Issue Dependencies as Tree**: Implemented tree view that depends on proper parent-child relationships.
- **SL-40f7df - US2: List Ready-to-Work Issues**: Implemented ready state computation that traverses the issue hierarchy.

## Assumptions

- The `sl issue link` command's existing architecture can be extended to support a new `parent` link type without significant refactoring.
- Circular parent detection only needs to check direct chains (A→B→C→A), not complex graph cycles, since parent-child is a strict tree structure.
- The `--orphaned` flag operates within the current spec context by default (consistent with other `sl issue list` flags).
- AI skill instruction updates are effective — agents follow updated prompts reliably.
