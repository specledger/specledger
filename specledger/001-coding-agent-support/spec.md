# Feature Specification: Multi-Coding Agent Support

**Feature Branch**: `001-coding-agent-support`
**Created**: 2026-03-07
**Status**: Draft
**Input**: User description: "multi coding agent support (claude, opencode, codex, etc), add sl code to launch the coding agent, utilize sl config to set it in global, project level, make sure the launch config can pass the arguments for each coding agent correctly e.g: --dangerously-skip-permission, etc, sl new and init can select multiple code editor and use symlink from .agent/commands and .agent/skills to the respected coding agent configuration"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Launch Coding Agent with `sl code` (Priority: P1)

As a developer, I want to launch my preferred AI coding agent from the command line using `sl code` so that I can quickly start an AI-assisted coding session with my project's configured settings.

**Why this priority**: This is the core value proposition - developers need a quick, unified way to launch any coding agent with proper configuration. Without this, the feature has no user-facing value.

**Independent Test**: Can be fully tested by running `sl code` in a project directory and verifying the agent launches with correct arguments and environment variables.

**Acceptance Scenarios**:

1. **Given** a project with agent configuration, **When** I run `sl code`, **Then** the configured default agent launches in the project directory
2. **Given** a project without agent configuration, **When** I run `sl code`, **Then** the default agent (Claude Code) launches with default settings
3. **Given** I want to launch a specific agent, **When** I run `sl code opencode`, **Then** OpenCode launches instead of the default
4. **Given** a project with `agent.arguments="--dangerously-skip-permissions"`, **When** I run `sl code`, **Then** the agent launches with those arguments passed directly
5. **Given** global config specifies agent.model=opus, **When** I run `sl code`, **Then** the agent launches with the correct model environment variable

---

### User Story 2 - Configure Agent Settings via `sl config` (Priority: P1)

As a developer, I want to configure coding agent settings (model, flags, environment variables) at both global and project levels using `sl config` so that my preferences are respected across all projects or per-project as needed.

**Why this priority**: Configuration is essential for the launch command to work correctly. Without proper configuration, the agent won't launch with the right settings.

**Independent Test**: Can be fully tested by running `sl config set` commands and verifying the values are stored and retrieved correctly.

**Acceptance Scenarios**:

1. **Given** I want to set a global default agent, **When** I run `sl config set --global agent.default claude`, **Then** the global config stores this preference
2. **Given** I want to set project-specific agent settings, **When** I run `sl config set agent.model sonnet`, **Then** the project-level config stores this preference
3. **Given** both global and project config exist, **When** I run `sl code`, **Then** project settings override global settings
4. **Given** I want to pass custom arguments to the agent, **When** I run `sl config set agent.arguments "--dangerously-skip-permissions --verbose"`, **Then** those arguments are passed directly to the agent on launch

---

### User Story 3 - Select Multiple Agents During Project Setup (Priority: P2)

As a developer setting up a new project, I want to select multiple coding agents (e.g., Claude Code and OpenCode) during `sl new` or `sl init` so that my project is configured to work with any of my preferred agents.

**Why this priority**: Multi-agent selection enhances the setup experience but the core launch functionality can work with a single agent first.

**Independent Test**: Can be fully tested by running `sl new` or `sl init` and verifying multiple agents can be selected and configuration directories are created.

**Acceptance Scenarios**:

1. **Given** I am creating a new project, **When** I reach the agent selection step, **Then** I can select multiple agents (not just one)
2. **Given** I select Claude Code and OpenCode, **When** project setup completes, **Then** both `.claude/` and `.opencode/` directories are created
3. **Given** I select multiple agents, **When** project setup completes, **Then** symlink structure is created for shared commands/skills
4. **Given** I have multiple agents configured, **When** I run `sl code opencode`, **Then** OpenCode launches instead of the default

---

### User Story 4 - Shared Commands and Skills via Symlinks (Priority: P2)

As a developer using multiple coding agents, I want commands and skills to be shared across agents via symlinks so that I don't have to maintain duplicate configuration files.

**Why this priority**: Symlink sharing reduces maintenance burden but can be implemented after basic multi-agent support is working.

**Independent Test**: Can be fully tested by creating a project with multiple agents and verifying symlink structure is correct.

**Acceptance Scenarios**:

1. **Given** I select multiple agents during setup, **When** setup completes, **Then** `.agent/commands` and `.agent/skills` directories are created
2. **Given** `.agent/commands` exists, **When** setup completes for Claude, **Then** `.claude/commands` is a symlink to `../.agent/commands`
3. **Given** `.agent/skills` exists, **When** setup completes for OpenCode, **Then** `.opencode/skills` is a symlink to `../.agent/skills`
4. **Given** I add a new command to `.agent/commands`, **Then** all linked agents can access it immediately

---

### User Story 5 - Pass Custom Arguments to Agents (Priority: P3)

As a developer, I want to pass arbitrary arguments to my coding agent via `agent.arguments` config so that I can customize agent behavior without needing individual config keys for each flag.

**Why this priority**: Generic argument passing provides flexibility but can be refined after core functionality is working.

**Independent Test**: Can be fully tested by setting `agent.arguments` and verifying the arguments are passed to the launched agent.

**Acceptance Scenarios**:

1. **Given** I set `agent.arguments="--dangerously-skip-permissions"`, **When** I run `sl code`, **Then** Claude launches with that flag
2. **Given** I set `agent.arguments="--model gpt-4"`, **When** I run `sl code codex`, **Then** Codex receives the model flag
3. **Given** I set `agent.arguments` with multiple flags, **When** I run `sl code`, **Then** all arguments are passed as-is to the agent
4. **Given** an agent is configured with custom env vars, **When** I run `sl code`, **Then** those env vars are injected into the agent process

---

### Edge Cases

- What happens when the selected agent binary is not installed or not in PATH?
- How does the system handle conflicting configuration between global and project settings?
- What happens when a user tries to launch an agent that wasn't configured during setup?
- How are symlinks handled on Windows (which has different symlink support)?
- What happens when `.agent/` directory already exists with different structure?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide an `sl code [<agent>]` command to launch a coding agent (optional positional argument for agent selection)
- **FR-002**: System MUST support at minimum the following agents: Claude Code, OpenCode, Codex
- **FR-003**: System MUST allow agent configuration via `sl config set` at global and project levels
- **FR-004**: System MUST support `agent.arguments` config key for passing arbitrary arguments directly to the agent
- **FR-005**: System MUST allow selection of multiple agents during `sl new` and `sl init`
- **FR-006**: System MUST create symlink structure from `.agent/commands` and `.agent/skills` to each agent's config directory
- **FR-007**: System MUST use `agent.default` config key to determine which agent to launch when no positional argument is provided
- **FR-008**: System MUST merge configuration from global and project levels with project taking precedence
- **FR-009**: System MUST pass environment variables from config to the launched agent process
- **FR-010**: System MUST detect and report when a selected agent is not installed
- **FR-011**: System MUST provide install instructions when a selected agent is unavailable
- **FR-012**: System MUST support new config keys: `agent.default`, `agent.arguments`, `agent.env`

### Key Entities

- **Agent Definition**: Represents a coding agent with its name, CLI command, and install instructions
- **Agent Configuration**: User preferences including default agent, custom arguments string, and environment variables
- **Launch Profile**: A combination of agent selection and configuration settings used to launch an agent
- **Shared Configuration Directory**: The `.agent/` directory containing shared commands and skills symlinked to agent-specific directories

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can launch any supported coding agent in under 2 seconds using `sl code` or `sl code <agent>`
- **SC-002**: Configuration changes via `sl config` are reflected in agent launch within 1 command execution
- **SC-003**: 100% of arguments in `agent.arguments` are passed directly to the launched agent
- **SC-004**: Users can set up a project with 3+ agents in a single `sl new` or `sl init` session
- **SC-005**: Symlink-based sharing works across all supported platforms (macOS, Linux)
- **SC-006**: Users receive clear error messages and install instructions when an agent is unavailable

### Previous work

No direct previous work exists for multi-agent support. Related infrastructure:

- **launcher/launcher.go**: Existing agent launcher for Claude Code
- **config/schema.go**: Configuration schema with agent.* keys
- **tui/sl_new.go, tui/sl_init.go**: TUI for agent selection (single agent only)

### Dependencies & Assumptions

**Dependencies**:
- Existing `sl config` command infrastructure
- Existing launcher package for agent process management
- Existing TUI framework for multi-select support

**Assumptions**:
- Users have their preferred coding agents installed separately
- Symlinks are supported on target platforms (macOS, Linux; Windows may need special handling)
- Each agent has a unique CLI command name
- Agents accept environment variables for configuration where applicable
