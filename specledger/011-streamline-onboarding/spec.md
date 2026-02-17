# Feature Specification: Streamlined Onboarding Experience

**Feature Branch**: `011-streamline-onboarding`
**Created**: 2026-02-17
**Status**: Draft
**Input**: User description: "Simplify sl commands so onboarding is easier. After sl init or sl new, start coding agent session immediately. sl init should ask the same questions as sl new if not answered. AI coding agent should guide user through specify → clarify → plan → implement workflow. Project config is defined in CONSTITUTION.md — sl new always creates constitution; sl init checks for existing constitution and if missing, launches explore agent to analyze codebase and propose guiding principles via ask-user-question."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unified Interactive Setup for New Projects (Priority: P1)

A new user runs `sl new` to create a project. They answer setup questions (project name, directory, short code, playbook), then are guided through creating their project constitution — defining core principles, coding standards, and guiding values for the project. Finally, they are asked which AI coding agent they prefer (e.g., Claude Code). After confirming, SpecLedger creates the project with a populated constitution and immediately launches their chosen coding agent with an onboarding prompt that guides them through the SpecLedger workflow.

**Why this priority**: This is the primary onboarding path for brand-new users. If new users can go from zero to a guided AI session in a single command — with a meaningful project constitution already in place — adoption friction drops dramatically. The constitution ensures all subsequent AI-assisted work (specify, plan, implement) is grounded in explicit project values from day one.

**Independent Test**: Can be fully tested by running `sl new`, answering all prompts including constitution principles, and verifying the project is created with a populated CONSTITUTION.md and the coding agent launches with a guided onboarding message.

**Acceptance Scenarios**:

1. **Given** a user with no existing project, **When** they run `sl new` and complete all prompts including constitution and agent preference, **Then** the project is created with a populated CONSTITUTION.md and their preferred coding agent launches automatically with a guided onboarding prompt.
2. **Given** a user creating a new project, **When** they reach the constitution step, **Then** they are presented with suggested guiding principles (e.g., specification-first, test-first, code quality) and can accept, modify, or add their own via interactive prompts.
3. **Given** a user running `sl new` who selects "None" for coding agent, **When** project creation completes, **Then** the project is created with a populated constitution and SpecLedger displays a message suggesting they run their coding agent manually with a provided command.
4. **Given** a user running `sl new` in CI mode (`--ci`), **When** project creation completes, **Then** no coding agent is launched and constitution creation uses provided flags or defaults (non-interactive environments skip interactive prompts).

---

### User Story 2 - Unified Interactive Setup for Existing Repositories (Priority: P1)

A user runs `sl init` in an existing repository that has not been configured with SpecLedger yet. Instead of silently using defaults, `sl init` detects that setup questions have not been answered and presents the same interactive prompts as `sl new` (minus project name and directory, since those are already determined by the current repository). For the constitution step, `sl init` first checks whether a CONSTITUTION.md already exists in the project. If one exists, it is used as-is. If no constitution exists, the system launches an exploration step that analyzes the existing codebase — identifying languages, frameworks, patterns, and conventions — and proposes tailored guiding principles for the constitution. The user is then presented with these proposed principles and can accept, modify, or add their own. After setup, the user is offered the option to launch their preferred coding agent with the onboarding guide.

**Why this priority**: `sl init` is the primary path for users adopting SpecLedger in existing projects. Without interactive prompts, users miss important configuration and don't get a guided introduction. The codebase-aware constitution proposal is critical here because existing projects already have implicit conventions that should be captured, not ignored. This has equal importance with User Story 1 since both are entry points.

**Independent Test**: Can be fully tested by running `sl init` in an uninitialized git repository with existing code, verifying it analyzes the codebase, proposes constitution principles reflecting the detected technologies and patterns, prompts for user confirmation, then launches the agent with a guided message.

**Acceptance Scenarios**:

1. **Given** a user in an existing git repository without SpecLedger and without a CONSTITUTION.md, **When** they run `sl init`, **Then** the system analyzes the codebase to detect languages, frameworks, and conventions, proposes tailored guiding principles, and presents them for user review and customization.
2. **Given** a user in an existing git repository that already has a CONSTITUTION.md, **When** they run `sl init`, **Then** the existing constitution is preserved and used, and the user is not prompted to create a new one.
3. **Given** a user in a repository already initialized with SpecLedger, **When** they run `sl init`, **Then** they are informed the project is already initialized and offered to re-run setup with `--force`, without re-asking already-answered questions.
4. **Given** a user who provides `--short-code` and `--playbook` flags, **When** they run `sl init`, **Then** those prompts are skipped and only the unanswered questions (e.g., constitution, agent preference) are asked.
5. **Given** the codebase analysis detects a Go project with Cobra CLI and Bubble Tea TUI, **When** the system proposes constitution principles, **Then** the proposed principles reflect these technologies (e.g., "CLI-First Interface", "Interactive TUI Experience") rather than generic defaults.

---

### User Story 3 - Quick Start with `sl start` (Priority: P1)

A user runs `sl start` in an existing project directory. The command automatically detects the project structure, reads existing project metadata (if available), analyzes the codebase to understand its technologies and conventions, and automatically generates or updates a CONSTITUTION.md file with tailored guiding principles. Once the constitution is in place, the command immediately launches the user's preferred AI coding agent (e.g., Claude Code) with an onboarding prompt that guides them through the SpecLedger workflow. This provides the fastest path to productive AI-assisted development without requiring multiple interactive prompts.

**Why this priority**: `sl start` is the ultimate convenience command for users who want to jump directly into AI-assisted development. It combines project analysis, constitution generation, and agent launch into a single command, eliminating friction for users who already have a project structure in place. This is equally important as Stories 1 and 2 because it serves users who want minimal setup overhead.

**Independent Test**: Can be fully tested by running `sl start` in a directory with existing code (no prior SpecLedger initialization), verifying it analyzes the codebase, generates a CONSTITUTION.md reflecting detected technologies, and launches the coding agent with the guided onboarding prompt.

**Acceptance Scenarios**:

1. **Given** a user in a directory with existing code and no SpecLedger initialization, **When** they run `sl start`, **Then** the system analyzes the codebase, generates a CONSTITUTION.md with tailored principles, and launches the preferred coding agent with the onboarding prompt.
2. **Given** a user in a SpecLedger-initialized project with an existing CONSTITUTION.md, **When** they run `sl start`, **Then** the existing constitution is preserved and the coding agent launches immediately with the onboarding prompt.
3. **Given** a user who runs `sl start` with `--agent claude-code` flag, **When** the command executes, **Then** the specified agent is launched regardless of saved preference.
4. **Given** a user who runs `sl start` with `--no-agent` flag, **When** the command executes, **Then** the constitution is generated/updated but no coding agent is launched.
5. **Given** the codebase analysis detects a TypeScript/React project with Jest testing, **When** the system generates the constitution, **Then** the principles reflect these technologies (e.g., "Component-Driven Development", "Test-First Approach").
6. **Given** a user runs `sl start` in a directory without a git repository, **When** the command executes, **Then** it offers to initialize git and proceeds with codebase analysis on the existing files.

---

### User Story 4 - Guided First Feature Workflow (Priority: P2)

After the coding agent launches from `sl new`, `sl init`, or `sl start`, it presents a clear guided walkthrough of the SpecLedger workflow. The agent first confirms the project constitution is in place (it was created during setup), then leads the user through creating their first feature specification, clarifying it, planning the implementation, generating tasks, reviewing tasks, and only then running implementation. The constitution serves as the foundation that informs every subsequent step.

**Why this priority**: The guided workflow converts a first-time user into a productive user. Without it, users launch the agent but don't know what to do next. This story depends on Stories 1, 2, and 3 being in place but is the key differentiator for user retention.

**Independent Test**: Can be tested by launching a coding agent with the onboarding prompt in any initialized SpecLedger project (with constitution already created) and verifying it walks through the full workflow sequence with appropriate pauses for user input.

**Acceptance Scenarios**:

1. **Given** a coding agent launched via SpecLedger onboarding, **When** the session starts, **Then** the agent confirms the constitution is in place, explains the SpecLedger workflow, and prompts the user to describe their first feature.
2. **Given** a user who has described a feature, **When** the agent runs `/specledger.specify`, **Then** it proceeds to `/specledger.clarify` and explains why clarification matters before continuing.
3. **Given** a user who has completed specify and clarify, **When** the agent runs `/specledger.plan` and `/specledger.tasks`, **Then** it pauses and asks the user to review all generated tasks before proceeding.
4. **Given** a user who has reviewed and approved the tasks, **When** the user confirms, **Then** the agent runs `/specledger.implement` to begin executing the tasks.
5. **Given** a user who wants to modify tasks after review, **When** they provide feedback, **Then** the agent updates the tasks before proceeding to implementation.

---

### User Story 5 - Agent Preference Persistence (Priority: P3)

When a user selects their preferred coding agent during onboarding, that preference is saved so subsequent commands and sessions can use it. If the user later wants to change their preference, they can update it without re-running full setup.

**Why this priority**: Persistence avoids repetitive prompting and enables future features that auto-launch the agent. Lower priority because the core onboarding works even if the preference is asked each time.

**Independent Test**: Can be tested by running `sl new`, selecting an agent, then verifying the preference is stored and used on subsequent `sl init --force` runs without re-prompting.

**Acceptance Scenarios**:

1. **Given** a user who selected "Claude Code" during onboarding, **When** they run `sl init --force` later, **Then** the agent preference prompt shows "Claude Code" as the default.
2. **Given** a user who wants to change their agent preference, **When** they update the project configuration, **Then** future sessions use the new preference.

---

### User Story 6 - Standardized Commits with `specledger.commit` (Priority: P2)

When a user completes a feature implementation and wants to commit their changes, they can use the `/specledger.commit` command to create a standardized, well-formatted commit message. The command automatically generates a commit message that includes:

- A clear summary line (max 72 characters)
- A detailed description of changes made
- Reference to the feature specification (if available)
- Automatic coauthor attribution with SpecLedger fingerprint

The coauthor line includes the SpecLedger AI assistant as a coauthor with the standardized fingerprint: `SpecLedger <specledger@noreply.github.com>`. This ensures that AI-assisted work is properly attributed while maintaining a clear audit trail of human-AI collaboration.

**Why this priority**: Standardized commits improve project history readability and ensure consistent attribution of AI-assisted work. This is P2 because it enhances workflow quality but is not critical to the core onboarding experience.

**Independent Test**: Can be tested by running `/specledger.commit` after completing a feature implementation, verifying the generated commit message includes proper formatting, feature reference, and coauthor attribution with the correct fingerprint.

**Acceptance Scenarios**:

1. **Given** a user who has completed feature implementation, **When** they run `/specledger.commit`, **Then** the system generates a commit message with a clear summary, detailed description, and coauthor line.
2. **Given** a feature specification exists for the completed work, **When** `/specledger.commit` is executed, **Then** the commit message includes a reference to the specification file (e.g., "Implements spec.md").
3. **Given** a user running `/specledger.commit`, **When** the commit message is generated, **Then** the coauthor line includes the fingerprint `SpecLedger <specledger@noreply.github.com>`.
4. **Given** a user who wants to customize the commit message, **When** they provide feedback on the generated message, **Then** the agent updates it before committing.
5. **Given** a user running `/specledger.commit` in a repository with no staged changes, **When** the command executes, **Then** the system prompts the user to stage changes first or offers to stage all changes.

---

### Edge Cases

- What happens when the preferred coding agent is not installed on the user's system?
  - SpecLedger checks availability before attempting launch, displays a helpful error with installation instructions, and completes the project setup successfully.
- What happens when `sl init` or `sl start` is run in a directory that is not a git repository?
  - The interactive prompts still work. Git initialization is offered as an optional step.
- What happens when the user cancels during the interactive prompts (Ctrl+C)?
  - Any partial setup is rolled back cleanly, or clearly communicated so the user can resume.
- What happens if the coding agent exits immediately after launch?
  - The project setup is still complete. The user can manually re-launch the agent at any time.
- What happens when `sl start` is run in a directory with no code files?
  - The system generates a default constitution based on the project structure and offers to create a starter template.
- What happens when `/specledger.commit` is run without a git repository?
  - The system displays an error message indicating that git is required and suggests initializing a repository first.
- What happens when `/specledger.commit` is run with no staged changes?
  - The system prompts the user to stage changes or offers to automatically stage all modified files.

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
- **FR-012**: `sl start` MUST automatically analyze the project codebase to detect languages, frameworks, and conventions.
- **FR-013**: `sl start` MUST generate or update a CONSTITUTION.md file with tailored guiding principles based on codebase analysis.
- **FR-014**: `sl start` MUST preserve any existing CONSTITUTION.md file and not overwrite it without explicit user consent.
- **FR-015**: `sl start` MUST launch the preferred coding agent (or specified via `--agent` flag) immediately after constitution is ready.
- **FR-016**: `sl start` MUST support `--agent <agent-name>` flag to override the saved agent preference.
- **FR-017**: `sl start` MUST support `--no-agent` flag to skip agent launch and only generate/update the constitution.
- **FR-018**: `sl start` MUST work in directories without prior SpecLedger initialization.
- **FR-019**: `sl start` MUST offer to initialize git if the directory is not a git repository.
- **FR-020**: `/specledger.commit` MUST generate a standardized commit message with summary, description, and coauthor attribution.
- **FR-021**: `/specledger.commit` MUST include the coauthor fingerprint `SpecLedger <specledger@noreply.github.com>` in the commit message.
- **FR-022**: `/specledger.commit` MUST reference the feature specification file if one exists in the project.
- **FR-023**: `/specledger.commit` MUST verify that changes are staged before attempting to commit.
- **FR-024**: `/specledger.commit` MUST allow the user to review and customize the generated commit message before finalizing.
- **FR-025**: `/specledger.commit` MUST work only in git repositories and display an appropriate error message if git is not initialized.

### Key Entities

- **Agent Preference**: The user's chosen AI coding agent (e.g., Claude Code, None). Stored per-project in project configuration alongside existing metadata like short code and playbook.
- **Onboarding Prompt**: A structured message passed to the coding agent at launch time that guides the user through the SpecLedger workflow steps in order.
- **Codebase Analysis**: Automated detection of project structure, languages, frameworks, dependencies, and conventions to inform constitution generation.
- **Constitution Generation**: Automatic creation of CONSTITUTION.md with tailored guiding principles based on codebase analysis and project metadata.
- **Standardized Commit Message**: A formatted commit message that includes summary, description, specification reference, and coauthor attribution with the SpecLedger fingerprint.
- **Coauthor Fingerprint**: The standardized identifier `SpecLedger <specledger@noreply.github.com>` used to attribute AI-assisted work in git commits.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new user can go from zero to an active, guided AI coding session in a single command (`sl new`) with no more than 5 interactive prompts.
- **SC-002**: A user adopting SpecLedger in an existing repository can complete setup and launch a guided session using only `sl init` — no additional commands required.
- **SC-003**: A user with an existing project can launch a guided AI session with a single command (`sl start`) with zero interactive prompts.
- **SC-004**: 100% of onboarding sessions that launch a coding agent include the guided workflow prompt (specify → clarify → plan → tasks review → implement).
- **SC-005**: The guided workflow pauses for user review before implementation in every session — zero cases of auto-running implementation without explicit user approval.
- **SC-006**: Users who have already configured their project are never re-prompted for settings they already provided, unless using `--force`.
- **SC-007**: `sl start` generates a CONSTITUTION.md that reflects the detected project technologies and conventions in 100% of cases.
- **SC-008**: 100% of commits generated via `/specledger.commit` include the coauthor fingerprint `SpecLedger <specledger@noreply.github.com>`.
- **SC-009**: Commit messages generated via `/specledger.commit` follow a consistent format with summary (≤72 chars), description, and specification reference.
- **SC-010**: Users can review and customize generated commit messages before finalizing in 100% of cases.

### Previous work

- **SL-31n** - "Command System Enhancements" (CLOSED): Added `/specledger.help`, enhanced all commands with Purpose sections for better discoverability. This feature builds on that foundation by making the initial onboarding guided rather than relying on users discovering commands.
- **SL-6t7** - "Enhanced Purpose Sections" (CLOSED): Added Purpose and "When to use" sections to all commands. Informed the design of the guided workflow sequence.
- **SL-9c7** - "Open Source Readiness" (OPEN, P1): Includes contributor onboarding and documentation improvements. Streamlined onboarding directly supports this initiative.
- **Feature 009-command-system-enhancements**: Established the command workflow organization (Core Workflow, Analysis & Validation, Setup & Configuration, Collaboration) that this feature leverages for the guided onboarding sequence.

## Assumptions

- The initial supported coding agent is Claude Code, as it is the only agent with deep SpecLedger integration (embedded commands and skills). Support for other agents (Cursor, Windsurf, etc.) may be added in future iterations.
- The onboarding prompt is a text-based message that can be passed as an initial prompt/argument when launching the coding agent.
- Users running `sl new`, `sl init`, or `sl start` in a terminal have an interactive TTY available unless `--ci` or `--no-agent` is specified.
- The existing Bubble Tea TUI used by `sl new` can be extended with additional prompts without a major rewrite.
- Project metadata (specledger.yaml) is the appropriate place to persist agent preference alongside existing project configuration.
- Codebase analysis can be performed using language detection libraries and dependency file parsing (package.json, go.mod, requirements.txt, etc.).
- The coauthor fingerprint `SpecLedger <specledger@noreply.github.com>` is the standardized identifier for SpecLedger AI assistance in git commits.
- Git commit coauthor lines follow the GitHub standard format: `Co-authored-by: Name <email@example.com>`.
