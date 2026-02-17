# Feature Specification: Streamlined Onboarding Experience

**Feature Branch**: `011-streamline-onboarding`
**Created**: 2026-02-17
**Status**: Draft
**Input**: User description: "Simplify sl commands so onboarding is easier. After sl init or sl new, start coding agent session immediately. sl init should ask the same questions as sl new if not answered. AI coding agent should guide user through specify → clarify → plan → implement workflow."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unified Interactive Setup for New Projects (Priority: P1)

A new user runs `sl new` to create a project. They answer a few setup questions (project name, directory, short code, playbook), then are asked which AI coding agent they prefer (e.g., Claude Code, Cursor, Windsurf). After confirming, SpecLedger creates the project and immediately launches their chosen coding agent with an onboarding prompt that guides them through the SpecLedger workflow.

**Why this priority**: This is the primary onboarding path for brand-new users. If new users can go from zero to a guided AI session in a single command, adoption friction drops dramatically. This story delivers the full end-to-end value of the feature.

**Independent Test**: Can be fully tested by running `sl new`, answering all prompts, and verifying the coding agent launches with a guided onboarding message. Delivers immediate value: user is inside an AI session ready to create their first feature.

**Acceptance Scenarios**:

1. **Given** a user with no existing project, **When** they run `sl new` and complete all prompts including agent preference, **Then** the project is created and their preferred coding agent launches automatically with a guided onboarding prompt.
2. **Given** a user running `sl new` who selects "None" for coding agent, **When** project creation completes, **Then** the project is created normally and SpecLedger displays a message suggesting they run their coding agent manually with a provided command.
3. **Given** a user running `sl new` in CI mode (`--ci`), **When** project creation completes, **Then** no coding agent is launched (non-interactive environments skip agent launch).

---

### User Story 2 - Unified Interactive Setup for Existing Repositories (Priority: P1)

A user runs `sl init` in an existing repository that has not been configured with SpecLedger yet. Instead of silently using defaults, `sl init` detects that setup questions have not been answered and presents the same interactive prompts as `sl new` (minus project name and directory, since those are already determined by the current repository). After setup, the user is offered the option to launch their preferred coding agent with the onboarding guide.

**Why this priority**: `sl init` is the primary path for users adopting SpecLedger in existing projects. Without interactive prompts, users miss important configuration and don't get a guided introduction. This has equal importance with User Story 1 since both are entry points.

**Independent Test**: Can be fully tested by running `sl init` in an uninitialized git repository, verifying it prompts for short code, playbook, and agent preference, then confirming the agent launches with a guided message.

**Acceptance Scenarios**:

1. **Given** a user in an existing git repository without SpecLedger, **When** they run `sl init`, **Then** they are presented with interactive prompts for short code, playbook selection, and agent preference.
2. **Given** a user in a repository already initialized with SpecLedger, **When** they run `sl init`, **Then** they are informed the project is already initialized and offered to re-run setup with `--force`, without re-asking already-answered questions.
3. **Given** a user who provides `--short-code` and `--playbook` flags, **When** they run `sl init`, **Then** those prompts are skipped and only the unanswered questions (e.g., agent preference) are asked.

---

### User Story 3 - Guided First Feature Workflow (Priority: P2)

After the coding agent launches from `sl new` or `sl init`, it presents a clear guided walkthrough of the SpecLedger workflow. The agent explains each step and leads the user through creating their first feature specification, clarifying it, planning the implementation, generating tasks, reviewing tasks, and only then running implementation.

**Why this priority**: The guided workflow converts a first-time user into a productive user. Without it, users launch the agent but don't know what to do next. This story depends on Stories 1 and 2 being in place but is the key differentiator for user retention.

**Independent Test**: Can be tested by launching a coding agent with the onboarding prompt in any initialized SpecLedger project and verifying it walks through the full workflow sequence with appropriate pauses for user input.

**Acceptance Scenarios**:

1. **Given** a coding agent launched via SpecLedger onboarding, **When** the session starts, **Then** the agent explains the SpecLedger workflow and prompts the user to describe their first feature.
2. **Given** a user who has described a feature, **When** the agent runs `/specledger.specify`, **Then** it proceeds to `/specledger.clarify` and explains why clarification matters before continuing.
3. **Given** a user who has completed specify and clarify, **When** the agent runs `/specledger.plan` and `/specledger.tasks`, **Then** it pauses and asks the user to review all generated tasks before proceeding.
4. **Given** a user who has reviewed and approved the tasks, **When** the user confirms, **Then** the agent runs `/specledger.implement` to begin executing the tasks.
5. **Given** a user who wants to modify tasks after review, **When** they provide feedback, **Then** the agent updates the tasks before proceeding to implementation.

---

### User Story 4 - Agent Preference Persistence (Priority: P3)

When a user selects their preferred coding agent during onboarding, that preference is saved so subsequent commands and sessions can use it. If the user later wants to change their preference, they can update it without re-running full setup.

**Why this priority**: Persistence avoids repetitive prompting and enables future features that auto-launch the agent. Lower priority because the core onboarding works even if the preference is asked each time.

**Independent Test**: Can be tested by running `sl new`, selecting an agent, then verifying the preference is stored and used on subsequent `sl init --force` runs without re-prompting.

**Acceptance Scenarios**:

1. **Given** a user who selected "Claude Code" during onboarding, **When** they run `sl init --force` later, **Then** the agent preference prompt shows "Claude Code" as the default.
2. **Given** a user who wants to change their agent preference, **When** they update the project configuration, **Then** future sessions use the new preference.

---

### Edge Cases

- What happens when the preferred coding agent is not installed on the user's system?
  - SpecLedger checks availability before attempting launch, displays a helpful error with installation instructions, and completes the project setup successfully.
- What happens when `sl init` is run in a directory that is not a git repository?
  - The interactive prompts still work. Git initialization is offered as an optional step.
- What happens when the user cancels during the interactive prompts (Ctrl+C)?
  - Any partial setup is rolled back cleanly, or clearly communicated so the user can resume.
- What happens if the coding agent exits immediately after launch?
  - The project setup is still complete. The user can manually re-launch the agent at any time.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `sl init` MUST present the same interactive setup prompts as `sl new` when required configuration is missing (short code, playbook, agent preference).
- **FR-002**: `sl init` MUST skip prompts for configuration values that have already been set or are provided via command-line flags.
- **FR-003**: Both `sl new` and `sl init` MUST include an agent preference prompt asking which AI coding agent to launch after setup.
- **FR-004**: The agent preference prompt MUST offer at least: Claude Code, None (manual launch). Additional agents may be added later.
- **FR-005**: After successful project setup, the system MUST automatically launch the selected coding agent if one was chosen.
- **FR-006**: The launched coding agent MUST receive an onboarding prompt that guides the user through: specify → clarify → plan → tasks (review) → implement.
- **FR-007**: The onboarding prompt MUST instruct the agent to pause and wait for user approval after task generation and before running implementation.
- **FR-008**: The system MUST verify the selected coding agent is available on the user's system before attempting to launch it.
- **FR-009**: The agent preference MUST be persisted in the project configuration so it does not need to be re-entered.
- **FR-010**: The system MUST NOT launch a coding agent in non-interactive environments (CI mode, piped input).
- **FR-011**: `sl init` in an already-initialized repository MUST inform the user and suggest `--force` to re-run setup.

### Key Entities

- **Agent Preference**: The user's chosen AI coding agent (e.g., Claude Code, None). Stored per-project in project configuration alongside existing metadata like short code and playbook.
- **Onboarding Prompt**: A structured message passed to the coding agent at launch time that guides the user through the SpecLedger workflow steps in order.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new user can go from zero to an active, guided AI coding session in a single command (`sl new`) with no more than 5 interactive prompts.
- **SC-002**: A user adopting SpecLedger in an existing repository can complete setup and launch a guided session using only `sl init` — no additional commands required.
- **SC-003**: 100% of onboarding sessions that launch a coding agent include the guided workflow prompt (specify → clarify → plan → tasks review → implement).
- **SC-004**: The guided workflow pauses for user review before implementation in every session — zero cases of auto-running implementation without explicit user approval.
- **SC-005**: Users who have already configured their project are never re-prompted for settings they already provided, unless using `--force`.

### Previous work

- **SL-31n** - "Command System Enhancements" (CLOSED): Added `/specledger.help`, enhanced all commands with Purpose sections for better discoverability. This feature builds on that foundation by making the initial onboarding guided rather than relying on users discovering commands.
- **SL-6t7** - "Enhanced Purpose Sections" (CLOSED): Added Purpose and "When to use" sections to all commands. Informed the design of the guided workflow sequence.
- **SL-9c7** - "Open Source Readiness" (OPEN, P1): Includes contributor onboarding and documentation improvements. Streamlined onboarding directly supports this initiative.
- **Feature 009-command-system-enhancements**: Established the command workflow organization (Core Workflow, Analysis & Validation, Setup & Configuration, Collaboration) that this feature leverages for the guided onboarding sequence.

## Assumptions

- The initial supported coding agent is Claude Code, as it is the only agent with deep SpecLedger integration (embedded commands and skills). Support for other agents (Cursor, Windsurf, etc.) may be added in future iterations.
- The onboarding prompt is a text-based message that can be passed as an initial prompt/argument when launching the coding agent.
- Users running `sl new` or `sl init` in a terminal have an interactive TTY available unless `--ci` is specified.
- The existing Bubble Tea TUI used by `sl new` can be extended with additional prompts without a major rewrite.
- Project metadata (specledger.yaml) is the appropriate place to persist agent preference alongside existing project configuration.
