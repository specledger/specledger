# Feature Specification: Streamlined Onboarding Experience

**Feature Branch**: `011-streamline-onboarding`
**Created**: 2026-02-17
**Status**: Draft
**Input**: User description: "Simplify sl commands so onboarding is easier. After sl init or sl new, start coding agent session immediately. sl init should ask the same questions as sl new if not answered. AI coding agent should guide user through specify → clarify → plan → implement workflow. Project config is defined in CONSTITUTION.md — sl new always creates constitution; sl init checks for existing constitution and if missing, launches explore agent to analyze codebase and propose guiding principles via ask-user-question."

## Clarifications

### Session 2026-02-18

- Q: What is the relationship between CONSTITUTION.md and specledger.yaml? → A: Coexist — specledger.yaml keeps project metadata (id, name, short_code, playbook). Constitution handles principles, preferences (agent), and coding standards. Both files serve distinct purposes.
- Q: How should codebase analysis run during sl init? → A: Delegate to AI agent — sl init launches the coding agent first, then the agent runs /specledger.audit to deeply analyze the codebase and propose constitution principles.
- Q: What is the canonical file path for the populated constitution? → A: `.specledger/memory/constitution.md` — consistent with existing template structure, internal to SpecLedger tooling.
- Q: Should this feature also address command consolidation beyond sl new/init? → A: Originally scoped to sl new/init only. Expanded to also include `sl start` (zero-prompt convenience command) and `/specledger.commit` (standardized commit messages). Other existing commands stay as-is.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unified Interactive Setup for New Projects (Priority: P1)

A new user runs `sl new` to create a project. They answer setup questions (project name, directory, short code, playbook), then are guided through creating their project constitution — defining core principles, coding standards, and guiding values for the project. Finally, they are asked which AI coding agent they prefer (e.g., Claude Code). After confirming, SpecLedger creates the project with a populated constitution and immediately launches their chosen coding agent with an onboarding prompt that guides them through the SpecLedger workflow.

**Why this priority**: This is the primary onboarding path for brand-new users. If new users can go from zero to a guided AI session in a single command — with a meaningful project constitution already in place — adoption friction drops dramatically. The constitution ensures all subsequent AI-assisted work (specify, plan, implement) is grounded in explicit project values from day one.

**Independent Test**: Can be fully tested by running `sl new`, answering all prompts including constitution principles, and verifying the project is created with a populated `.specledger/memory/constitution.md` and the coding agent launches with a guided onboarding message.

**Acceptance Scenarios**:

1. **Given** a user with no existing project, **When** they run `sl new` and complete all prompts including constitution and agent preference, **Then** the project is created with a populated `.specledger/memory/constitution.md` and their preferred coding agent launches automatically with a guided onboarding prompt.
2. **Given** a user creating a new project, **When** they reach the constitution step, **Then** they are presented with suggested guiding principles (e.g., specification-first, test-first, code quality) and can accept, modify, or add their own via interactive prompts.
3. **Given** a user running `sl new` who selects "None" for coding agent, **When** project creation completes, **Then** the project is created with a populated constitution and SpecLedger displays a message suggesting they run their coding agent manually with a provided command.
4. **Given** a user running `sl new` in CI mode (`--ci`), **When** project creation completes, **Then** no coding agent is launched and constitution creation uses provided flags or defaults (non-interactive environments skip interactive prompts).

---

### User Story 2 - Unified Interactive Setup for Existing Repositories (Priority: P1)

A user runs `sl init` in an existing repository that has not been configured with SpecLedger yet. Instead of silently using defaults, `sl init` detects that setup questions have not been answered and presents the same interactive prompts as `sl new` (minus project name and directory, since those are already determined by the current repository). For the constitution step, `sl init` first checks whether a populated constitution already exists at `.specledger/memory/constitution.md`. If one exists, it is used as-is. If no constitution exists, `sl init` completes the basic setup (short code, playbook, agent preference) and then launches the selected AI coding agent. The agent's onboarding prompt instructs it to first run `/specledger.audit` to deeply analyze the codebase — identifying languages, frameworks, patterns, and conventions — and then use the findings to propose tailored guiding principles for the constitution via `/specledger.constitution`. The user reviews and customizes the proposed principles interactively within the agent session.

**Why this priority**: `sl init` is the primary path for users adopting SpecLedger in existing projects. Without interactive prompts, users miss important configuration and don't get a guided introduction. Delegating codebase analysis to the AI agent provides richer, deeper understanding of the existing codebase than built-in heuristics could offer, and existing projects already have implicit conventions that should be captured, not ignored. This has equal importance with User Story 1 since both are entry points.

**Independent Test**: Can be fully tested by running `sl init` in an uninitialized git repository with existing code, verifying it completes basic setup, launches the AI agent, and the agent runs audit + constitution creation with principles reflecting the detected technologies and patterns.

**Acceptance Scenarios**:

1. **Given** a user in an existing git repository without SpecLedger and without a populated constitution, **When** they run `sl init`, **Then** basic setup completes (short code, playbook, agent preference), the AI agent launches, and the agent's first guided step is to run `/specledger.audit` followed by `/specledger.constitution` to propose tailored principles for user review.
2. **Given** a user in an existing git repository that already has a populated `.specledger/memory/constitution.md`, **When** they run `sl init`, **Then** the existing constitution is preserved and used, and the agent's onboarding skips the audit + constitution step.
3. **Given** a user in a repository already initialized with SpecLedger, **When** they run `sl init`, **Then** they are informed the project is already initialized and offered to re-run setup with `--force`, without re-asking already-answered questions.
4. **Given** a user who provides `--short-code` and `--playbook` flags, **When** they run `sl init`, **Then** those prompts are skipped and only the unanswered questions (e.g., agent preference) are asked.
5. **Given** the AI agent's audit detects a Go project with Cobra CLI and Bubble Tea TUI, **When** the agent proposes constitution principles, **Then** the proposed principles reflect these technologies (e.g., "CLI-First Interface", "Interactive TUI Experience") rather than generic defaults.

---

### User Story 3 - Quick Start with `sl start` (Priority: P1)

A user runs `sl start` in an existing project directory. The command automatically detects the project structure, reads existing project metadata (if available), and checks for an existing populated constitution at `.specledger/memory/constitution.md`. If no constitution exists (or only an unfilled template is present), `sl start` launches the preferred AI coding agent, which runs `/specledger.audit` to analyze the codebase — detecting languages, frameworks, patterns, and conventions — then runs `/specledger.constitution` to propose tailored guiding principles for user review. Once the constitution is in place, the agent continues with the guided onboarding workflow. This provides the fastest path to productive AI-assisted development without requiring multiple interactive prompts.

**Why this priority**: `sl start` is the ultimate convenience command for users who want to jump directly into AI-assisted development. It combines project analysis, constitution generation, and agent launch into a single command, eliminating friction for users who already have a project structure in place. This is equally important as Stories 1 and 2 because it serves users who want minimal setup overhead.

**Independent Test**: Can be fully tested by running `sl start` in a directory with existing code (no prior SpecLedger initialization), verifying it launches the agent, the agent analyzes the codebase, generates a populated `.specledger/memory/constitution.md` reflecting detected technologies, and proceeds with the guided onboarding prompt.

**Acceptance Scenarios**:

1. **Given** a user in a directory with existing code and no SpecLedger initialization, **When** they run `sl start`, **Then** the system initializes SpecLedger, launches the AI agent, the agent runs `/specledger.audit` and `/specledger.constitution` to create a constitution with tailored principles, and proceeds with the onboarding prompt.
2. **Given** a user in a SpecLedger-initialized project with an existing populated `.specledger/memory/constitution.md`, **When** they run `sl start`, **Then** the existing constitution is preserved and the coding agent launches immediately with the onboarding prompt (skipping audit + constitution).
3. **Given** a user who runs `sl start` with `--agent claude-code` flag, **When** the command executes, **Then** the specified agent is launched regardless of saved preference.
4. **Given** a user who runs `sl start` with `--no-agent` flag, **When** the command executes, **Then** SpecLedger setup completes but no coding agent is launched.
5. **Given** the AI agent's audit detects a TypeScript/React project with Jest testing, **When** the agent proposes constitution principles, **Then** the principles reflect these technologies (e.g., "Component-Driven Development", "Test-First Approach") rather than generic defaults.
6. **Given** a user runs `sl start` in a directory without a git repository, **When** the command executes, **Then** it offers to initialize git and proceeds with codebase analysis on the existing files.

---

### User Story 4 - Guided First Feature Workflow (Priority: P2)

After the coding agent launches from `sl new`, `sl init`, or `sl start`, it presents a clear guided walkthrough of the SpecLedger workflow. For `sl init` and `sl start` projects without a constitution, the agent's first step is to run `/specledger.audit` then `/specledger.constitution` to create one collaboratively with the user. For `sl new` projects (where the constitution was created during TUI setup), the agent confirms it is in place. In both cases, the agent then leads the user through creating their first feature specification, clarifying it, planning the implementation, generating tasks, reviewing tasks, and only then running implementation. The constitution serves as the foundation that informs every subsequent step.

**Why this priority**: The guided workflow converts a first-time user into a productive user. Without it, users launch the agent but don't know what to do next. This story depends on Stories 1, 2, and 3 being in place but is the key differentiator for user retention.

**Independent Test**: Can be tested by launching a coding agent with the onboarding prompt in any initialized SpecLedger project (with constitution already created) and verifying it walks through the full workflow sequence with appropriate pauses for user input.

**Acceptance Scenarios**:

1. **Given** a coding agent launched via SpecLedger onboarding from `sl init` (no existing constitution), **When** the session starts, **Then** the agent runs `/specledger.audit` and `/specledger.constitution` first, then explains the SpecLedger workflow and prompts the user to describe their first feature.
2. **Given** a coding agent launched via SpecLedger onboarding from `sl new` (constitution already created), **When** the session starts, **Then** the agent confirms the constitution is in place, explains the SpecLedger workflow, and prompts the user to describe their first feature.
3. **Given** a user who has described a feature, **When** the agent runs `/specledger.specify`, **Then** it proceeds to `/specledger.clarify` and explains why clarification matters before continuing.
4. **Given** a user who has completed specify and clarify, **When** the agent runs `/specledger.plan` and `/specledger.tasks`, **Then** it pauses and asks the user to review all generated tasks before proceeding.
5. **Given** a user who has reviewed and approved the tasks, **When** the user confirms, **Then** the agent runs `/specledger.implement` to begin executing the tasks.
6. **Given** a user who wants to modify tasks after review, **When** they provide feedback, **Then** the agent updates the tasks before proceeding to implementation.

---

### User Story 5 - Agent Preference Persistence in Constitution (Priority: P3)

When a user selects their preferred coding agent during onboarding, that preference is saved as part of the project constitution so subsequent commands and sessions can use it. If the user later wants to change their preference, they can update the constitution without re-running full setup.

**Why this priority**: Persistence avoids repetitive prompting and enables future features that auto-launch the agent. Lower priority because the core onboarding works even if the preference is asked each time. Storing it in the constitution keeps all project governance in one place.

**Independent Test**: Can be tested by running `sl new`, selecting an agent, then verifying the preference is recorded in `.specledger/memory/constitution.md` and used on subsequent `sl init --force` runs without re-prompting.

**Acceptance Scenarios**:

1. **Given** a user who selected "Claude Code" during onboarding, **When** they run `sl init --force` later, **Then** the agent preference is read from the existing constitution and shown as the default.
2. **Given** a user who wants to change their agent preference, **When** they update the constitution, **Then** future sessions use the new preference.

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
- What happens when `sl init` or `sl start` launches the AI agent for codebase analysis on an empty or nearly empty repository?
  - The agent's `/specledger.audit` detects minimal code and `/specledger.constitution` falls back to offering generic default principles (same as `sl new`) rather than attempting to infer conventions from insufficient code.
- What happens when an existing `.specledger/memory/constitution.md` is found but is still in template/placeholder form (unfilled)?
  - The system treats an unfilled template the same as "no constitution" and proceeds with the interactive constitution creation flow, using codebase analysis to propose principles.
- What happens when the codebase analysis detects conflicting patterns (e.g., mixed languages, inconsistent conventions)?
  - The system presents all detected patterns transparently and lets the user decide which conventions to enshrine as principles.
- What happens when `/specledger.commit` is run without a git repository?
  - The system displays an error message indicating that git is required and suggests initializing a repository first.
- What happens when `/specledger.commit` is run with no staged changes?
  - The system prompts the user to stage changes or offers to automatically stage all modified files.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `sl init` MUST present the same interactive setup prompts as `sl new` when required configuration is missing (short code, playbook, constitution, agent preference).
- **FR-002**: `sl init` MUST skip prompts for configuration values that have already been set or are provided via command-line flags.
- **FR-003**: Both `sl new` and `sl init` MUST include an agent preference prompt asking which AI coding agent to launch after setup.
- **FR-004**: The agent preference prompt MUST offer at least: Claude Code, None (manual launch). Additional agents may be added later.
- **FR-005**: After successful project setup, the system MUST automatically launch the selected coding agent if one was chosen.
- **FR-006**: The launched coding agent MUST receive an onboarding prompt that guides the user through: specify → clarify → plan → tasks (review) → implement.
- **FR-007**: The onboarding prompt MUST instruct the agent to pause and wait for user approval after task generation and before running implementation.
- **FR-008**: The system MUST verify the selected coding agent is available on the user's system before attempting to launch it.
- **FR-009**: The agent preference MUST be persisted in the project constitution so it does not need to be re-entered.
- **FR-010**: The system MUST NOT launch a coding agent in non-interactive environments (CI mode, piped input).
- **FR-011**: `sl init` in an already-initialized repository MUST inform the user and suggest `--force` to re-run setup.
- **FR-012**: `sl new` MUST always create a project constitution as part of the setup flow, presenting suggested principles for the user to accept, modify, or extend.
- **FR-013**: `sl init` and `sl start` MUST check whether a populated `.specledger/memory/constitution.md` already exists in the project. If one exists, it MUST be preserved and used, and the agent's onboarding MUST skip the audit + constitution step.
- **FR-014**: `sl init` and `sl start` MUST detect an unfilled/template constitution (containing only placeholder tokens) and treat it the same as "no constitution."
- **FR-015**: When no constitution exists during `sl init` or `sl start`, the onboarding prompt MUST instruct the launched AI agent to first run `/specledger.audit` to analyze the codebase, then run `/specledger.constitution` to propose tailored guiding principles based on the audit findings.
- **FR-016**: Proposed constitution principles MUST be presented to the user for review, with the ability to accept, modify, reject individual principles, or add new ones.
- **FR-017**: The project constitution MUST be stored at `.specledger/memory/constitution.md` and serve as the authoritative source for guiding principles, coding standards, and preferences (including agent preference). Project metadata (id, name, short_code, playbook) remains in `specledger.yaml`.
- **FR-018**: `sl start` MUST work in directories without prior SpecLedger initialization, performing automatic setup before agent launch.
- **FR-019**: `sl start` MUST offer to initialize git if the directory is not a git repository.
- **FR-020**: `sl start` MUST support `--agent <agent-name>` flag to override the saved agent preference.
- **FR-021**: `sl start` MUST support `--no-agent` flag to skip agent launch.
- **FR-022**: `/specledger.commit` MUST generate a standardized commit message with summary, description, and coauthor attribution.
- **FR-023**: `/specledger.commit` MUST include the coauthor fingerprint `SpecLedger <specledger@noreply.github.com>` in the commit message.
- **FR-024**: `/specledger.commit` MUST reference the feature specification file if one exists in the project.
- **FR-025**: `/specledger.commit` MUST verify that changes are staged before attempting to commit.
- **FR-026**: `/specledger.commit` MUST allow the user to review and customize the generated commit message before finalizing.
- **FR-027**: `/specledger.commit` MUST work only in git repositories and display an appropriate error message if git is not initialized.
- **FR-028**: This feature's scope covers `sl new`, `sl init`, and `sl start` commands plus the `/specledger.commit` embedded command. Other existing commands (doctor, deps, session, playbook, auth, graph) remain unchanged.

### Key Entities

- **Constitution**: The project's foundational governance document (`.specledger/memory/constitution.md`) containing core principles, coding standards, guiding values, and project preferences (including agent preference). Created during onboarding and used by all subsequent SpecLedger commands (specify, plan, analyze) to validate compliance. Coexists with `specledger.yaml`, which handles project metadata separately.
- **Agent Preference**: The user's chosen AI coding agent (e.g., Claude Code, None). Stored as part of the project constitution alongside core principles and other project configuration.
- **Onboarding Prompt**: A structured message passed to the coding agent at launch time that guides the user through the SpecLedger workflow steps in order.
- **Codebase Analysis**: An AI-agent-driven exploration of the existing repository (during `sl init` and `sl start` onboarding) using `/specledger.audit`. The agent detects languages, frameworks, patterns, and conventions to inform constitution principle proposals via `/specledger.constitution`.
- **Standardized Commit Message**: A formatted commit message that includes summary, description, specification reference, and coauthor attribution with the SpecLedger fingerprint.
- **Coauthor Fingerprint**: The standardized identifier `SpecLedger <specledger@noreply.github.com>` used to attribute AI-assisted work in git commits.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new user can go from zero to an active, guided AI coding session in a single command (`sl new`) — including a populated project constitution — with no more than 7 interactive steps.
- **SC-002**: A user adopting SpecLedger in an existing repository can complete setup (including constitution creation with codebase-aware proposals) and launch a guided session using only `sl init` — no additional commands required.
- **SC-003**: A user with an existing project can launch a guided AI session with a single command (`sl start`) with zero interactive prompts.
- **SC-004**: 100% of onboarding sessions that launch a coding agent include the guided workflow prompt (specify → clarify → plan → tasks review → implement).
- **SC-005**: The guided workflow pauses for user review before implementation in every session — zero cases of auto-running implementation without explicit user approval.
- **SC-006**: Users who have already configured their project are never re-prompted for settings they already provided, unless using `--force`.
- **SC-007**: 100% of projects created via `sl new` have a populated `.specledger/memory/constitution.md` upon completion — no template placeholders remain.
- **SC-008**: For `sl init` and `sl start` on existing codebases, proposed constitution principles reflect at least 2 specific characteristics detected from the actual codebase (e.g., detected language, framework, or convention).
- **SC-009**: 100% of commits generated via `/specledger.commit` include the coauthor fingerprint `SpecLedger <specledger@noreply.github.com>`.
- **SC-010**: Commit messages generated via `/specledger.commit` follow a consistent format with summary (≤72 chars), description, and specification reference.
- **SC-011**: Users can review and customize generated commit messages before finalizing in 100% of cases.

### Previous work

- **SL-31n** - "Command System Enhancements" (CLOSED): Added `/specledger.help`, enhanced all commands with Purpose sections for better discoverability. This feature builds on that foundation by making the initial onboarding guided rather than relying on users discovering commands.
- **SL-6t7** - "Enhanced Purpose Sections" (CLOSED): Added Purpose and "When to use" sections to all commands. Informed the design of the guided workflow sequence.
- **SL-9c7** - "Open Source Readiness" (OPEN, P1): Includes contributor onboarding and documentation improvements. Streamlined onboarding directly supports this initiative.
- **Feature 009-command-system-enhancements**: Established the command workflow organization (Core Workflow, Analysis & Validation, Setup & Configuration, Collaboration) that this feature leverages for the guided onboarding sequence.

## Assumptions

- The initial supported coding agent is Claude Code, as it is the only agent with deep SpecLedger integration (embedded commands and skills). Support for other agents (Cursor, Windsurf, etc.) may be added in future iterations.
- The onboarding prompt is a text-based message that can be passed as an initial prompt/argument when launching the coding agent.
- Users running `sl new`, `sl init`, or `sl start` in a terminal have an interactive TTY available unless `--ci` or `--no-agent` is specified.
- The existing Bubble Tea TUI used by `sl new` can be extended with additional prompts (including constitution creation) without a major rewrite.
- The constitution (at `.specledger/memory/constitution.md`) and `specledger.yaml` coexist with distinct responsibilities: specledger.yaml stores project metadata (id, name, short_code, version, playbook), while the constitution stores guiding principles, coding standards, and preferences (including agent preference).
- For `sl init` and `sl start`, codebase analysis is delegated to the AI coding agent, which runs `/specledger.audit` followed by `/specledger.constitution`. This provides richer analysis than built-in heuristics but requires the AI agent to be installed.
- An unfilled constitution template (containing `[PLACEHOLDER]` tokens) is distinguishable from a populated constitution and is treated as "no constitution present."
- The constitution creation step during `sl new` uses sensible default principle suggestions (specification-first, test-first, code quality, etc.) that the user can customize, keeping the onboarding flow fast for users who accept defaults.
- The coauthor fingerprint `SpecLedger <specledger@noreply.github.com>` is the standardized identifier for SpecLedger AI assistance in git commits.
- Git commit coauthor lines follow the GitHub standard format: `Co-authored-by: Name <email@example.com>`.
