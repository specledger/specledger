# Feature Specification: SDD Workflow Streamline

**Feature Branch**: `598-sdd-workflow-streamline`
**Created**: 2026-02-27
**Status**: Draft
**Input**: User description: "we want to streamline the workflow of sdd using specledger, and add/remove/modify ai skills/commands/ or sl cli command"

## Overview

Audit and consolidate the overlapping SDD (Specification-Driven Development) workflow components across three layers:
1. **AI Skills** (`.opencode/skills/`) - 2 skills
2. **AI Commands** (`.opencode/commands/`) - 15 commands
3. **CLI Commands** (`sl` binary) - 11 commands (+1 new: `sl update`)

Goal: Reduce redundancy, clarify responsibilities, streamline the developer experience, and add new workflow capabilities.

**Future Work** (out of scope): Migrate interactive commands (`sl bootstrap`, `sl init`, `sl revise`) to a separate TUI tool.

## Clarifications

### Session 2026-02-27

- Q: How should `sl update` distinguish built-in (updatable) files from custom (preserved) files? → A: Filename matching (compare against list of embedded template names)
- Q: Should spike and checkpoint be implemented as AI commands or CLI commands? → A: AI commands (markdown files in `.opencode/commands/`)

## New Capabilities to Add

1. **Spike** (AI command) - Exploratory research command for time-boxed investigations
2. **Checkpoint** (AI command) - Implementation verification command for tracking progress against specs
3. **Update** (CLI command) - Update AI skills/commands to latest embedded templates

## Current State Inventory

### Skills (2)
| Skill | Purpose |
|-------|---------|
| `sl-issue-tracking` | Issue management with `sl issue` commands |
| `specledger-deps` | Dependency management with `sl deps` commands |

### Commands (15)
| Command | Purpose | Overlap Concern |
|---------|---------|-----------------|
| `add-deps` | Add spec dependencies | Overlaps with `sl deps add` CLI |
| `adopt` | Adopt existing code | Unique |
| `analyze` | Codebase analysis | Similar to `audit` |
| `audit` | Full codebase audit | Similar to `analyze` |
| `checklist` | Create checklists | Unique |
| `clarify` | Clarify spec ambiguities | Unique |
| `constitution` | Project constitution | Unique |
| `help` | Help documentation | Unique |
| `implement` | Implementation guidance | Core workflow |
| `onboard` | Project onboarding | Unique |
| `plan` | Create implementation plan | Core workflow |
| `remove-deps` | Remove spec dependencies | Overlaps with `sl deps remove` CLI |
| `resume` | Resume work session | Unique |
| `specify` | Create feature spec | Core workflow |
| `tasks` | Generate tasks | Core workflow |

### CLI Commands (11 → 12 with `sl update`)
| Command | Purpose | Overlap Concern | Future: Migrate to TUI? |
|---------|---------|-----------------|------------------------|
| `sl bootstrap` | Create new project | Unique | **Yes** - interactive wizard better in TUI |
| `sl init` | Initialize in existing project | Unique | **Yes** - interactive wizard better in TUI |
| `sl deps` | Manage dependencies | Overlaps with `add-deps`/`remove-deps` commands | No |
| `sl graph` | Dependency graph | Unique | No |
| `sl doctor` | System diagnostics | Unique | No |
| `sl playbook` | Run playbooks | Unique | Maybe |
| `sl auth` | Authentication | Unique | No |
| `sl session` | Session capture | Unique | No |
| `sl issue` | Issue tracking | Covered by `sl-issue-tracking` skill | No |
| `sl revise` | Revise comments | Unique | **Yes** - interactive review better in TUI |
| `sl config` | Configuration | Unique | No |
| `sl update` | Update AI skills/commands | New command | No |

> **Note**: TUI migration is out of scope for this spec. Commands marked for TUI migration will be addressed in a future spec when the separate TUI tool is created.

## Identified Overlaps

1. **Dependency Management**: `add-deps`/`remove-deps` commands vs `sl deps add`/`sl deps remove` CLI
2. **Analysis**: `analyze` vs `audit` commands (similar purposes)
3. **Issue Tracking**: `sl-issue-tracking` skill vs `sl issue` CLI (should be complementary, not duplicate)

### Proposed: `sl update` vs `sl init` Overlap Evaluation

| Aspect | `sl init` | `sl init --force` | `sl update` (proposed) |
|--------|-----------|-------------------|-------------------------|
| Purpose | First-time initialization | Re-initialize everything | Update skills/commands only |
| Scope | Full project setup | Full project setup | AI templates only |
| Config preserved | N/A | No (overwrites) | Yes |
| Custom skills/commands | N/A | Overwritten | Preserved (user can opt-in to overwrite) |
| Use case | New project | Reset/corrupted state | Template updates |

**Recommendation**: Add `sl update` as a distinct command. `sl init --force` is destructive reset; `sl update` is selective update.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Audit and Document Workflow Components (Priority: P1)

A developer or maintainer needs a clear map of all workflow components to understand overlaps and gaps.

**Why this priority**: Cannot streamline without understanding current state.

**Independent Test**: Can be fully tested by reviewing documentation that lists all components with their purposes and relationships.

**Acceptance Scenarios**:

1. **Given** the current codebase, **When** audit is complete, **Then** all skills, commands, and CLI commands are documented with their purposes
2. **Given** the component inventory, **When** overlaps are identified, **Then** each overlap is documented with a recommended resolution

---

### User Story 2 - Consolidate Dependency Commands (Priority: P1)

A developer wants a single, clear way to manage spec dependencies without confusion between AI commands and CLI commands.

**Why this priority**: Dependency management has the most obvious overlap and user confusion.

**Independent Test**: Can be fully tested by removing redundant commands and verifying `sl deps` CLI handles all dependency operations.

**Acceptance Scenarios**:

1. **Given** overlapping `add-deps`/`remove-deps` commands and `sl deps` CLI, **When** consolidation is complete, **Then** only one interface exists for dependency management
2. **Given** the consolidated interface, **When** user needs to manage dependencies, **Then** the workflow is clear and documented

---

### User Story 3 - Consolidate Analysis Commands (Priority: P2)

A developer wants clear distinction between codebase analysis and audit commands.

**Why this priority**: Reduces confusion but less critical than dependency management.

**Independent Test**: Can be fully tested by merging or clearly distinguishing `analyze` and `audit` commands.

**Acceptance Scenarios**:

1. **Given** overlapping `analyze` and `audit` commands, **When** consolidation is complete, **Then** each has a distinct, documented purpose or they are merged
2. **Given** the consolidated commands, **When** user needs codebase analysis, **Then** the appropriate command is obvious

---

### User Story 4 - Update Skills to Complement CLI (Priority: P2)

A developer wants AI skills that enhance CLI functionality rather than duplicate it.

**Why this priority**: Improves developer experience by making skills additive.

**Independent Test**: Can be fully tested by reviewing skill content to ensure it references CLI commands rather than duplicating their logic.

**Acceptance Scenarios**:

1. **Given** `sl-issue-tracking` skill and `sl issue` CLI, **When** skill is updated, **Then** skill provides AI context about issue workflow without duplicating CLI functionality
2. **Given** `specledger-deps` skill and `sl deps` CLI, **When** skill is updated, **Then** skill provides AI context about dependency workflow without duplicating CLI functionality

---

### User Story 5 - Document Consolidated Workflow (Priority: P3)

A developer wants clear documentation of the streamlined SDD workflow.

**Why this priority**: Documentation is important but follows the actual consolidation work.

**Independent Test**: Can be fully tested by reviewing updated documentation.

**Acceptance Scenarios**:

1. **Given** consolidated workflow, **When** documentation is updated, **Then** all skills, commands, and CLI commands are documented with their purposes
2. **Given** the documentation, **When** a new developer reads it, **Then** the workflow is clear and unambiguous

---

### User Story 6 - Create Spike for Exploratory Research (Priority: P1)

A developer needs to perform time-boxed exploratory research (spike) on a technical question or approach before committing to implementation. The spike results are captured in a structured format for future reference.

**Why this priority**: Spikes are essential for reducing implementation risk by validating approaches early.

**Independent Test**: Can be fully tested by running the spike command and verifying a research document is created in `[spec]/research/` with date-prefixed filename.

**Acceptance Scenarios**:

1. **Given** an active feature spec, **When** user runs spike command with a topic, **Then** a research file is created at `specledger/[spec-id]/research/yyyy-mm-dd-[spike-name].md`
2. **Given** the spike command, **When** spike completes, **Then** the file contains research question, approach, findings, and recommendations
3. **Given** multiple spikes on same spec, **When** listing research, **Then** all spikes are visible in chronological order

---

### User Story 7 - Checkpoint Implementation Progress (Priority: P1)

A developer wants to verify that current implementation aligns with the spec and track what has changed since the last checkpoint. The checkpoint captures implementation state for review and rollback purposes.

**Why this priority**: Checkpoints provide safety and visibility into implementation progress, enabling course correction before drift becomes costly.

**Independent Test**: Can be fully tested by running checkpoint command and verifying a session file is created in `[spec]/sessions/` with date-prefixed filename containing implementation verification.

**Acceptance Scenarios**:

1. **Given** an active feature with implementation in progress, **When** user runs checkpoint command, **Then** a session file is created at `specledger/[spec-id]/sessions/yyyy-mm-dd-[session-name].md`
2. **Given** the checkpoint command, **When** checkpoint runs, **Then** the file contains: spec compliance status, changed files list, implementation notes, and any deviations found
3. **Given** git changes since last checkpoint, **When** checkpoint runs, **Then** file changes are summarized with diff statistics
4. **Given** multiple checkpoints on same spec, **When** listing sessions, **Then** all checkpoints are visible in chronological order

---

### User Story 8 - Update AI Skills and Commands (Priority: P2)

A developer wants to update their project's AI skills and commands to the latest embedded template versions when a new version of SpecLedger is released, without losing project-specific customizations.

**Why this priority**: Keeps projects up-to-date with workflow improvements, but less urgent than core workflow commands.

**Independent Test**: Can be fully tested by running `sl update` and verifying skills/commands are updated while custom files are preserved.

**Acceptance Scenarios**:

1. **Given** a project with existing skills/commands, **When** user runs `sl update`, **Then** built-in skills and commands are updated to latest embedded versions
2. **Given** a project with custom (non-built-in) skills/commands, **When** user runs `sl update`, **Then** custom files are preserved unchanged
3. **Given** a built-in skill/command that was locally modified, **When** user runs `sl update`, **Then** user is prompted to keep local or use updated template
4. **Given** `sl update --dry-run`, **When** run, **Then** shows what would be updated without making changes
5. **Given** `sl update --list`, **When** run, **Then** shows which embedded templates are available and their versions

---

### User Story 9 - Update CLI README After Streamlining (Priority: P2)

After the CLI command streamlining is complete and stable, the project README.md must be updated to reflect the new simplified command structure and workflow. The README should highlight the streamlined workflow, show clear examples of each command, and guide new users through the improved developer experience.

**Why this priority**: Documentation is essential for adoption. Users discovering SpecLedger via the README need accurate, up-to-date information about the simplified commands. P2 because the commands must be implemented and stable first, but documentation should follow closely.

**Independent Test**: Can be tested by reviewing the updated README.md to verify it documents all streamlined commands with accurate usage examples.

**Acceptance Scenarios**:

1. **Given** the CLI streamlining feature is complete, **When** the README is updated, **Then** it documents all remaining commands with usage examples.
2. **Given** a new user reading the README, **When** they follow the quickstart instructions, **Then** the commands shown match the actual CLI behavior.
3. **Given** the README is updated, **When** a user searches for workflow instructions, **Then** the README provides a clear, step-by-step guide.

---

### User Story 10 - Add CHANGELOG.md for AI Commands/Skills Templates (Priority: P3)

When embedded AI commands (`.opencode/commands/`) and skills (`.opencode/skills/`) templates are updated, a CHANGELOG.md should be included in the embedded templates to track these changes. This allows projects initialized with SpecLedger to understand what changes have been made to their command/skill files over time and decide whether to adopt updates.

**Why this priority**: Provides transparency for template evolution but is not critical to the core workflow experience. P3 because it's a nice-to-have for maintainability but doesn't block user adoption.

**Independent Test**: Can be tested by verifying a CHANGELOG.md exists in the embedded templates and documents changes to command/skill files with version references.

**Acceptance Scenarios**:

1. **Given** the embedded templates directory, **When** a CHANGELOG.md is added, **Then** it lists all AI command and skill files with their initial versions.
2. **Given** a command or skill template is modified, **When** the change is released, **Then** the CHANGELOG.md is updated with the change description, affected files, and version.
3. **Given** a user running `sl bootstrap` or `sl init`, **When** templates are copied to their project, **Then** the CHANGELOG.md is included so they can track template changes.
4. **Given** a user reviewing their project's CHANGELOG.md, **When** they want to update their commands/skills, **Then** the CHANGELOG provides enough information to decide which changes to adopt.

---

### Edge Cases

- What happens to existing projects using removed commands? Document migration path.
- How to handle commands that are removed but still referenced in external docs? Add deprecation notices.
- What happens when spike research directory doesn't exist? Create it automatically.
- What happens when checkpoint finds no changes since last checkpoint? Still create session file noting "no changes".
- What happens when checkpoint detects spec violations? Flag violations prominently in session file.
- What happens when `sl update` finds no updates available? Display "Already up to date" message.
- What happens when `sl update` on a project never initialized? Error with suggestion to run `sl init` first.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST have documented inventory of all skills, commands, and CLI commands with their purposes
- **FR-002**: Dependency management MUST be consolidated to a single interface (prefer `sl deps` CLI)
- **FR-003**: Analysis commands MUST have distinct purposes or be merged
- **FR-004**: Skills MUST complement CLI functionality, not duplicate it
- **FR-005**: Removed/deprecated commands MUST have migration documentation
- **FR-006**: Updated workflow MUST be documented in AGENTS.md or equivalent
- **FR-007**: All remaining commands MUST have clear, non-overlapping purposes
- **FR-008**: System MUST provide spike command to create exploratory research documents in `specledger/[spec-id]/research/yyyy-mm-dd-[name].md`
- **FR-009**: Spike files MUST include research question, approach explored, findings, and recommendations
- **FR-010**: System MUST provide checkpoint command to verify implementation against specs in `specledger/[spec-id]/sessions/yyyy-mm-dd-[name].md`
- **FR-011**: Checkpoint files MUST include spec compliance status, changed files summary, implementation notes, and any deviations
- **FR-012**: Checkpoint MUST detect and summarize git file changes since last checkpoint (or branch creation)
- **FR-013**: Both spike and checkpoint MUST auto-create target directories if they don't exist
- **FR-014**: System MUST provide `sl update` CLI command to update built-in skills and commands to latest embedded template versions
- **FR-015**: `sl update` MUST preserve custom (non-built-in) skills and commands (detected by filename matching against embedded template list)
- **FR-016**: `sl update` MUST prompt for conflict resolution when built-in files were locally modified
- **FR-017**: `sl update --dry-run` MUST show pending updates without making changes
- **FR-018**: `sl update --list` MUST show available embedded templates and versions

### Key Entities

- **Skill**: AI context file that guides agent behavior for specific domains
- **Command**: AI command file that orchestrates multi-step workflows
- **CLI Command**: Binary command that performs concrete operations
- **Spike**: Time-boxed exploratory research document stored in `research/` folder with date prefix, capturing investigation results
- **Checkpoint**: Implementation verification document stored in `sessions/` folder with date prefix, capturing progress against spec
- **Built-in template**: Skills or commands embedded in the `sl` binary, identified by matching filename against the embedded template list

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Total command count reduced by at least 20% (from 15 to 12 or fewer AI commands)
- **SC-002**: Zero overlapping purposes between skills, commands, and CLI
- **SC-003**: Each workflow component has a single, documented responsibility
- **SC-004**: New developers can understand the SDD workflow in under 5 minutes of reading
- **SC-005**: Spike command creates research document in under 2 minutes
- **SC-006**: Checkpoint command captures implementation state in under 30 seconds
- **SC-007**: All spike and checkpoint files follow consistent `yyyy-mm-dd-[name].md` naming convention

### Previous work

Existing infrastructure in `.opencode/skills/` (2 skills) and `.opencode/commands/` (15 commands). CLI commands in `pkg/cli/commands/` (11 commands).

Existing session capture in `sl session capture` command stores in `specledger/[spec-id]/sessions/` but lacks checkpoint verification features.

### Dependencies & Assumptions

**Out of Scope** (future spec):
- TUI tool creation and migration of `sl bootstrap`, `sl init`, `sl revise` to TUI
- Interactive wizard mode for bootstrapping

**Assumptions**:
- CLI commands (`sl deps`, `sl issue`, etc.) are the source of truth for operations
- AI commands should orchestrate workflows and provide context, not duplicate CLI logic
- Skills should provide domain knowledge, not operational instructions
- Spike and checkpoint are AI commands (`.opencode/commands/specledger.spike.md` and `.opencode/commands/specledger.checkpoint.md`)
- Git is available for detecting file changes in checkpoint

**Dependencies**:
- Existing `.opencode/` directory structure
- Current CLI command implementations
- Git for checkpoint change detection
