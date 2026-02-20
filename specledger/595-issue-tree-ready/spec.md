# Feature Specification: Issue Tree View and Ready Command

**Feature Branch**: `595-issue-tree-ready`
**Created**: 2026-02-20
**Status**: Draft
**Input**: User description: "sl issue tree doesnt show the view as a tree, need to have sl issue ready to list down the issue in current spec that is read (not blocked by dependency)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Issue Dependencies as Tree (Priority: P1)

As a developer working on a spec, I want to see my issues displayed as a hierarchical tree so that I can quickly understand the dependency relationships between issues and identify which issues block others.

**Why this priority**: Understanding dependencies is critical for planning work order and identifying blockers. This is the core visualization improvement.

**Independent Test**: Can be fully tested by creating issues with dependencies using `sl issue link` and then running `sl issue list --tree` to verify the hierarchical display shows parent-child relationships correctly.

**Acceptance Scenarios**:

1. **Given** issues exist with dependency relationships, **When** user runs `sl issue list --tree`, **Then** issues are displayed in a hierarchical tree format showing parent-child relationships with visual indentation and connecting lines
2. **Given** issues have no dependencies, **When** user runs `sl issue list --tree`, **Then** all issues are displayed as a flat list under a root node
3. **Given** an issue blocks multiple other issues, **When** tree view is displayed, **Then** all blocked issues appear as children of the blocking issue
4. **Given** the `--all` flag is combined with `--tree`, **When** user runs `sl issue list --all --tree`, **Then** all issues across all specs are displayed in tree format grouped by spec

---

### User Story 2 - List Ready-to-Work Issues (Priority: P1)

As a developer starting work, I want to see a list of issues that are ready to work on (not blocked by dependencies) so that I can quickly pick up the next available task without manually checking each issue's dependencies.

**Why this priority**: This is the most common workflow - developers need to know what they can work on right now. It directly improves productivity.

**Independent Test**: Can be fully tested by creating issues with various blocking relationships and verifying that `sl issue ready` only shows issues where all blocking issues are closed.

**Acceptance Scenarios**:

1. **Given** issues exist with no blocking dependencies, **When** user runs `sl issue ready`, **Then** all open issues are listed
2. **Given** an open issue is blocked by an open issue, **When** user runs `sl issue ready`, **Then** the blocked issue is not shown in the ready list
3. **Given** an open issue was blocked but the blocking issue is now closed, **When** user runs `sl issue ready`, **Then** the previously blocked issue now appears in the ready list
4. **Given** an issue has multiple blockers and at least one is still open, **When** user runs `sl issue ready`, **Then** the issue is not shown
5. **Given** user runs `sl issue ready --all`, **When** command executes, **Then** ready issues from all specs are listed

---

### User Story 3 - View Single Issue Dependency Tree (Priority: P2)

As a developer investigating a specific issue, I want to see its dependency tree so that I can understand what it blocks and what blocks it.

**Why this priority**: This is valuable for focused investigation but less common than the general tree view or ready list.

**Independent Test**: Can be fully tested by running `sl issue show <id> --tree` and verifying the tree shows the issue with all its dependencies and dependents.

**Acceptance Scenarios**:

1. **Given** an issue with blocking relationships, **When** user runs `sl issue show <id> --tree`, **Then** a centered tree is displayed showing what this issue blocks and what blocks it
2. **Given** an issue with no dependencies, **When** user runs `sl issue show <id> --tree`, **Then** the issue is shown as a standalone node

---

### Edge Cases

- What happens when there are cyclic dependencies? Display warning and show cycle in tree
- How does the system handle issues with broken dependency references? Show warning indicator on the affected issue
- What happens with deeply nested trees (>10 levels)? Truncate with indication of hidden levels
- What about issues that block issues in other specs? Show cross-spec reference in tree

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display issues in a hierarchical tree format when `--tree` flag is passed to `sl issue list`
- **FR-002**: Tree view MUST show visual indentation and connecting characters to indicate parent-child relationships
- **FR-003**: Tree view MUST include issue ID, truncated title, and status for each issue
- **FR-004**: System MUST provide a new `sl issue ready` command that lists issues not blocked by dependencies
- **FR-005**: `sl issue ready` MUST only show issues with status `open` or `in_progress` where all blocking dependencies are closed
- **FR-006**: `sl issue ready` MUST support the `--all` flag to list ready issues across all specs
- **FR-007**: `sl issue ready` MUST support the `--json` flag for machine-readable output
- **FR-008**: System MUST enhance `sl issue show --tree` to display the issue's dependency context (what it blocks, what blocks it)
- **FR-009**: Tree view MUST detect and warn about cyclic dependencies without crashing
- **FR-010**: System MUST respect existing filters (status, type, priority, label, spec) when combined with `--tree`

### Key Entities

- **Issue**: Existing entity with ID, title, status, dependencies (blocks/blocked_by relationships)
- **Dependency Tree**: Virtual structure computed at runtime from issue relationships
- **Ready State**: Computed property - an issue is "ready" when it is open/in_progress AND all issues that block it are closed

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can identify the next available task in under 5 seconds using `sl issue ready`
- **SC-002**: Tree view renders correctly for dependency chains up to 10 levels deep
- **SC-003**: 100% of open issues with closed blockers appear in `sl issue ready` output
- **SC-004**: Zero false positives in ready list (no blocked issues shown as ready)
- **SC-005**: Tree view command completes in under 2 seconds for specs with up to 100 issues

### Previous work

- **591-issue-tracking-upgrade**: Built-in issue tracking system with JSONL storage, dependencies, and CLI commands
- **594-issues-storage-config**: Issues storage configuration with configurable artifact paths
