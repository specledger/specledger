# Feature Specification: Improve SpecLedger Command Prompts

**Feature Branch**: `592-prompt-updates`
**Created**: 2026-02-20
**Status**: Draft
**Input**: User description: "update specledger prompts (both in .claude and embedded) - specledger.specify: utilize dependency (sl deps) if referred by user, specledger.tasks: fix sl issue link/create errors, utilize definition of done, make issues more descriptive, specledger.implement: check definition of done and acceptance criteria"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reference Dependencies During Specification (Priority: P1)

As a developer creating a feature specification, when I mention external specifications, APIs, or dependencies in my feature description, the `/specledger.specify` command should automatically recognize these references and either load context from `sl deps` or prompt me to add them as dependencies.

**Why this priority**: This is the starting point of the workflow. If dependencies aren't properly recognized during specification, downstream tasks will lack necessary context.

**Independent Test**: Can be fully tested by creating a spec that references an external API (e.g., "integrate with Stripe payment API") and verifying that the system either loads existing Stripe deps context or prompts to add them.

**Acceptance Scenarios**:

1. **Given** a user creates a spec mentioning "integrate with company-design-system", **When** the design system dependency exists via `sl deps`, **Then** the spec generation should load and reference relevant context from that dependency
2. **Given** a user creates a spec mentioning an unknown dependency, **When** no matching dep exists, **Then** the system should prompt "I noticed you mentioned 'X'. Would you like to add this as a dependency using 'sl deps add'?"
3. **Given** a user explicitly references a dependency alias like "using deps:api-contracts", **When** the dependency exists, **Then** load the content and include relevant context in the spec

---

### User Story 2 - Generate Descriptive, Complete Issues (Priority: P1)

As a developer running `/specledger.tasks`, I want the generated issues to be descriptive, complete, and concise so that any developer can pick up a task and understand what needs to be done without additional context gathering.

**Why this priority**: Task quality directly impacts implementation efficiency. Poorly described tasks lead to context-switching and rework.

**Independent Test**: Can be tested by generating tasks from a completed plan and verifying that each issue: (a) has a clear problem statement, (b) describes inputs/outputs, (c) references relevant files, (d) has testable acceptance criteria.

**Acceptance Scenarios**:

1. **Given** a plan.md with clear components, **When** tasks are generated, **Then** each issue includes a concise title (under 80 chars), a problem statement explaining WHY, implementation details explaining HOW/WHERE, and acceptance criteria for WHAT success looks like
2. **Given** tasks are being created, **When** the `sl issue create` command is called, **Then** it should succeed without errors (handle edge cases like special characters in descriptions)
3. **Given** tasks need dependencies linked, **When** the `sl issue link` command is called, **Then** it should properly establish the relationship without errors

---

### User Story 3 - Utilize Definition of Done During Implementation (Priority: P1)

As a developer implementing tasks with `/specledger.implement`, I want the system to check each task's Definition of Done (DoD) and acceptance criteria before marking it complete, ensuring quality standards are met.

**Why this priority**: Without proper DoD verification, tasks may be marked complete prematurely, leading to technical debt and incomplete features.

**Independent Test**: Can be tested by implementing a task that has DoD items, then verifying the system checks each DoD item before allowing closure.

**Acceptance Scenarios**:

1. **Given** a task with DoD items, **When** implementation is complete, **Then** the system should display the DoD checklist and verify each item before allowing the issue to be closed
2. **Given** a task with acceptance criteria, **When** implementation is complete, **Then** the system should verify the implementation meets each acceptance criterion
3. **Given** a task where DoD items are not checked, **When** attempting to close the issue, **Then** the system should warn about incomplete DoD items and require explicit confirmation or `--force` to proceed

---

### Edge Cases

- What happens when `sl deps` references a dependency that no longer exists in the cache?
- What happens when `sl issue create` fails due to file system errors (e.g., permissions)?
- What happens when a dependency reference is ambiguous (multiple matching aliases)?
- What happens when DoD items contain special characters that break parsing?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `/specledger.specify` command MUST detect dependency references in user-provided feature descriptions
- **FR-002**: The `/specledger.specify` command MUST load existing dependency content from `sl deps list` cache when referenced dependencies exist
- **FR-003**: The `/specledger.specify` command MUST prompt users to add missing dependencies when external specifications are mentioned
- **FR-004**: The `/specledger.tasks` command MUST generate issues with structured content: title, problem statement (WHY), implementation details (HOW/WHERE), and acceptance criteria (WHAT)
- **FR-005**: The `/specledger.tasks` command MUST handle `sl issue create` and `sl issue link` errors gracefully with clear error messages
- **FR-006**: The `/specledger.tasks` command MUST include definition_of_done items in each generated task based on acceptance criteria from spec.md
- **FR-007**: The `/specledger.implement` command MUST verify Definition of Done items before closing issues
- **FR-008**: The `/specledger.implement` command MUST verify acceptance criteria are met before task completion
- **FR-009**: The `/specledger.implement` command MUST require explicit confirmation when closing issues with incomplete DoD items
- **FR-010**: Both `.claude/commands/` and `pkg/embedded/skills/commands/` prompt files MUST be updated with the same changes

### Key Entities

- **Dependency Reference**: A mention of an external specification, API, or system within a feature description that should be resolved via `sl deps`
- **Issue Content Structure**: A structured format containing: title, problem statement, implementation details, acceptance criteria, and definition_of_done items
- **Definition of Done (DoD)**: A checklist of verifiable items that must be satisfied before a task is considered complete

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can reference dependencies in spec descriptions and have relevant context automatically loaded 95% of the time when dependencies exist
- **SC-002**: Generated issues have all required fields (title, problem statement, implementation details, acceptance criteria) 100% of the time
- **SC-003**: `sl issue create` and `sl issue link` commands succeed on first attempt 99% of the time (error handling for edge cases)
- **SC-004**: Tasks are never closed without DoD verification unless explicitly forced
- **SC-005**: Developers can understand what a task requires without additional context gathering in 90% of cases

### Previous work

- **591-issue-tracking-upgrade**: Built-in issue tracking system with `sl issue` commands
- **008-fix-sl-deps**: SpecLedger dependencies management system

### Epic: 592 - Improve SpecLedger Command Prompts

- **Dependency-aware specification**: Recognize and load external dependencies during spec creation
- **Quality task generation**: Generate descriptive, complete, concise issues with proper error handling
- **DoD-enforced implementation**: Verify definition of done and acceptance criteria during implementation
