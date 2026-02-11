# Feature Specification: Conditional Prerequisites Checking

**Feature Branch**: `508-prerequisites-check`
**Created**: 2026-02-11
**Status**: Draft
**Input**: User description: "use this current branch 508, simplify preqreuisites checking, check whether beads and perlsed needed, if not no need to check"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Early Stage Workflow (Priority: P1)

A developer starting a new feature only needs to run `/specledger.specify` and `/specledger.clarify` to create their specification. They should not be blocked by installing tools (beads, perles) that are only needed later in the workflow.

**Why this priority**: This is the foundational workflow. If users cannot even start specifying features without installing unnecessary tools, they may abandon the workflow entirely. Making early commands work with minimal prerequisites reduces friction and enables rapid iteration.

**Independent Test**: A new user can run `/specledger.specify "test feature"` and complete the specification phase without having beads or perles installed. The system only reports missing tools when attempting commands that actually require them.

**Acceptance Scenarios**:

1. **Given** a developer has only mise installed, **When** they run `/specledger.specify`, **Then** the command completes successfully without checking for beads or perles
2. **Given** a developer has only mise installed, **When** they run `/specledger.clarify`, **Then** the command completes successfully without checking for beads or perles
3. **Given** a developer attempts to run `/specledger.plan` without beads installed, **Then** the system provides a clear message that beads is required for this command
4. **Given** a developer has mise installed, **When** they run the `doctor` command, **Then** only mise is checked by default, with options to check other tools

---

### User Story 2 - Command-Specific Tool Validation (Priority: P2)

A developer wants to run commands that require specific tools (like `plan` or `tasks` which need beads). When they attempt these commands, the system validates only the tools needed for that specific command.

**Why this priority**: This prevents false errors and provides accurate, actionable guidance. Users only see relevant errors, reducing confusion and support burden.

**Independent Test**: A user without beads installed can run `specify` and `clarify` commands successfully, but receives helpful installation instructions when running `plan` or `tasks` commands that require beads.

**Acceptance Scenarios**:

1. **Given** a developer has mise installed but not beads, **When** they run `/specledger.plan`, **Then** the system exits with a clear message: "plan.md requires beads to query previous work. Install with: mise install ubi:steveyegge/beads@0.28.0"
2. **Given** a developer has mise installed but not beads, **When** they run `/specledger.tasks`, **Then** the system exits with a clear message: "tasks.md requires beads for issue tracking. Install with: mise install ubi:steveyegge/beads@0.28.0"
3. **Given** a developer has beads installed, **When** they run `/specledger.plan`, **Then** the command proceeds without tool-related errors
4. **Given** a developer runs `doctor --check-all`, **When** tools are missing, **Then** all tools are checked and reported with appropriate installation commands

---

### User Story 3 - Perles Tool Removal (Priority: P3)

The perles tool, currently listed as a core requirement, is not actually used by any SpecLedger commands. Removing it simplifies the prerequisites.

**Why this priority**: This is a cleanup task that reduces confusion. Perles appears in the code but has no actual usage, making it a source of confusion for new users.

**Independent Test**: After removal, the `doctor` command no longer checks for perles, and the prerequisites list in documentation no longer includes it.

**Acceptance Scenarios**:

1. **Given** a developer runs the `doctor` command, **When** perles is not installed, **Then** no error is reported for perles
2. **Given** a developer reads the prerequisites documentation, **When** they view the core tools list, **Then** perles is not mentioned
3. **Given** the code previously referenced perles, **When** the code is updated, **Then** all perles references are removed from core tool lists

---

### Edge Cases

- What happens when a user has an outdated version of beads that doesn't support required flags?
- How does the system handle when multiple commands are chained together with different tool requirements?
- What happens when a user installs a tool mid-session after seeing an error?
- How does CI/CD environment handle tool checks when running commands programmatically?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST only validate tools that are actually required for the command being executed
- **FR-002**: Commands that require beads (plan, tasks, analyze, resume, implement) MUST check for beads availability before execution
- **FR-003**: Commands that do NOT require beads (specify, clarify) MUST NOT fail due to missing beads
- **FR-004**: System MUST remove perles from all core tool requirements
- **FR-005**: When a required tool is missing, system MUST provide clear installation instructions specific to that tool
- **FR-006**: The `doctor` command MUST support a `--check-all` flag to validate all optional tools
- **FR-007**: Tool validation errors MUST clearly indicate which command requires which tool
- **FR-008**: mise MUST remain the only truly required tool for basic SpecLedger functionality
- **FR-009**: System MUST distinguish between "core required" tools and "command-specific" tools in messaging
- **FR-010**: Framework tools (specify, openspec) MUST remain optional and not be checked unless explicitly requested

### Key Entities

- **Command**: A SpecLedger workflow command (specify, clarify, plan, tasks, implement, analyze, resume, etc.)
- **Tool**: An external executable required for certain operations (mise, beads)
- **Tool Category**: Classification of tools as "core required" or "command-specific optional"
- **Prerequisite Check**: Validation process that occurs before command execution

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can run `/specledger.specify` and `/specledger.clarify` with only mise installed (0 additional tools)
- **SC-002**: Developers receive clear, actionable error messages when running commands without required tools (100% of cases)
- **SC-003**: Time to first successful spec creation is reduced by 50% (no need to install unused tools)
- **SC-004**: Perles is completely removed from all tool lists and documentation
- **SC-005**: The `doctor` command completes successfully with only mise installed in under 2 seconds
- **SC-006**: User confusion about prerequisites is reduced, measured by 80% fewer support requests related to "why do I need X tool?"

### Previous work

### Epic: 009-command-system-enhancements - Command System Enhancements

- **CLI Authentication**: Added GitHub OAuth integration for CLI commands
- **Beads Issue Tracking**: Integrated beads (bd) for task management and dependency tracking
- **Perles Tool**: Perles was added as a core tool but appears to have no actual usage in the codebase

### Related Features

- **008-cli-auth**: Authentication workflow that doesn't require beads
- **007-release-delivery-fix**: Release processes that may use different tool sets

## Dependencies & Assumptions

### Dependencies

- Existing mise installation workflow remains unchanged
- Beads tool availability and installation methods (mise install ubi:steveyegge/beads@0.28.0)
- Command invocation patterns in the CLI framework

### Assumptions

- Users will install tools when they need them for specific commands
- The perles tool has no actual usage and can be safely removed
- mise is the only tool that must be pre-installed for all workflows
- Commands have well-defined tool requirements that don't change dynamically
