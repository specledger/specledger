# Feature Specification: Multi-Coding Agent Support

**Feature Branch**: `001-coding-agent-support`
**Created**: 2026-03-07
**Status**: Draft
**Input**: User description: "multi coding agent support (claude, opencode, codex, etc), add sl code to launch the coding agent, utilize sl config to set it in global, project level, make sure the launch config can pass the arguments for each coding agent correctly e.g: --dangerously-skip-permission, etc, sl new and init can select multiple code editor and use symlink from .agent/commands and .agent/skills to the respected coding agent configuration"

## Clarifications

### Session 2026-03-15

- Q: Which coding agents should be supported in the initial implementation? → A: All 4: Claude Code, OpenCode, Copilot CLI, Codex
- Q: How should agent arguments be configured? → A: Per-agent only (e.g., `agent.claude.arguments`, `agent.opencode.arguments`)
- Q: How should the system handle Windows symlink limitations? → A: Copy files instead of symlinks on Windows (copy `.agent/commands` to `.claude/commands`, etc.)
- Q: What should the error message include when an agent binary is not found in PATH? → A: Include install command (e.g., "Error: 'claude' not found. Install: npm install -g @anthropic-ai/claude-code")
- Q: What should happen when `.agent/` directory already exists with different structure? → A: Require `--force` flag to proceed with overwrite

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Launch Coding Agent with `sl code` (Priority: P1)

As a developer, I want to launch my preferred AI coding agent from the command line using `sl code` so that I can quickly start an AI-assisted coding session with my project's configured settings.

**Why this priority**: This is the core value proposition - developers need a quick, unified way to launch any coding agent with proper configuration. Without this, the feature has no user-facing value.

**Independent Test**: Can be fully tested by running `sl code` in a project directory and verifying the agent launches with correct arguments and environment variables.

**Acceptance Scenarios**:

1. **Given** a project with agent configuration, **When** I run `sl code`, **Then** the configured default agent launches in the project directory
2. **Given** a project without agent configuration, **When** I run `sl code`, **Then** the default agent (Claude Code) launches with default settings
3. **Given** I want to launch a specific agent, **When** I run `sl code opencode`, **Then** OpenCode launches instead of the default
4. **Given** a project with `agent.claude.arguments="--dangerously-skip-permissions"`, **When** I run `sl code claude`, **Then** the agent launches with those arguments passed directly
5. **Given** global config specifies `agent.claude.env.ANTHROPIC_MODEL=opus`, **When** I run `sl code claude`, **Then** the agent launches with the correct model environment variable

---

### User Story 2 - Configure Agent Settings via `sl config` (Priority: P1)

As a developer, I want to configure coding agent settings (model, flags, environment variables) at both global and project levels using `sl config` so that my preferences are respected across all projects or per-project as needed.

**Why this priority**: Configuration is essential for the launch command to work correctly. Without proper configuration, the agent won't launch with the right settings.

**Independent Test**: Can be fully tested by running `sl config set` commands and verifying the values are stored and retrieved correctly.

**Acceptance Scenarios**:

1. **Given** I want to set a global default agent, **When** I run `sl config set --global agent.default claude`, **Then** the global config stores this preference
2. **Given** I want to set project-specific agent settings, **When** I run `sl config set agent.claude.model sonnet`, **Then** the project-level config stores this preference
3. **Given** both global and project config exist, **When** I run `sl code`, **Then** project settings override global settings
4. **Given** I want to pass custom arguments to Claude, **When** I run `sl config set agent.claude.arguments "--dangerously-skip-permissions --verbose"`, **Then** those arguments are passed directly to Claude on launch

---

### User Story 3 - Select Multiple Agents During Project Setup (Priority: P2)

As a developer setting up a new project, I want to select multiple coding agents (e.g., Claude Code, OpenCode, Copilot CLI, and Codex) during `sl new` or `sl init` so that my project is configured to work with any of my preferred agents.

**Why this priority**: Multi-agent selection enhances the setup experience but the core launch functionality can work with a single agent first.

**Independent Test**: Can be fully tested by running `sl new` or `sl init` and verifying multiple agents can be selected and configuration directories are created.

**Acceptance Scenarios**:

1. **Given** I am creating a new project, **When** I reach the agent selection step, **Then** I can select multiple agents from the list (Claude Code, OpenCode, Copilot CLI, Codex)
2. **Given** I select Claude Code and OpenCode, **When** project setup completes, **Then** both `.claude/` and `.opencode/` directories are created
3. **Given** I select multiple agents, **When** project setup completes, **Then** symlink structure is created for shared commands/skills (on macOS/Linux); files are copied on Windows
4. **Given** I have multiple agents configured, **When** I run `sl code opencode`, **Then** OpenCode launches instead of the default

---

### User Story 4 - Shared Commands and Skills via Symlinks (Priority: P2)

As a developer using multiple coding agents, I want commands and skills to be shared across agents so that I don't have to maintain duplicate configuration files.

**Why this priority**: Shared configuration reduces maintenance burden but can be implemented after basic multi-agent support is working.

**Independent Test**: Can be fully tested by creating a project with multiple agents and verifying the shared configuration structure is correct.

**Acceptance Scenarios**:

1. **Given** I select multiple agents during setup, **When** setup completes, **Then** `.agent/commands` and `.agent/skills` directories are created
2. **Given** `.agent/commands` exists on macOS/Linux, **When** setup completes for Claude, **Then** `.claude/commands` is a symlink to `../.agent/commands`
3. **Given** `.agent/skills` exists on Windows, **When** setup completes for Claude, **Then** `.claude/skills` is a copy of `.agent/skills` contents
4. **Given** I add a new command to `.agent/commands` on macOS/Linux, **Then** all linked agents can access it immediately
5. **Given** I add a new command to `.agent/commands` on Windows, **Then** I must re-run setup or manually sync to propagate changes

---

### User Story 5 - Pass Custom Arguments to Agents (Priority: P3)

As a developer, I want to pass arbitrary arguments to each coding agent via per-agent config keys so that I can customize agent behavior without needing individual config keys for each flag.

**Why this priority**: Generic argument passing provides flexibility but can be refined after core functionality is working.

**Independent Test**: Can be fully tested by setting per-agent arguments and verifying the arguments are passed to the launched agent.

**Acceptance Scenarios**:

1. **Given** I set `agent.claude.arguments="--dangerously-skip-permissions"`, **When** I run `sl code claude`, **Then** Claude launches with that flag
2. **Given** I set `agent.opencode.arguments="--model gpt-4"`, **When** I run `sl code opencode`, **Then** OpenCode receives the model flag
3. **Given** I set `agent.codex.arguments` with multiple flags, **When** I run `sl code codex`, **Then** all arguments are passed as-is to Codex
4. **Given** an agent is configured with custom env vars, **When** I run `sl code`, **Then** those env vars are injected into the agent process

---

### Edge Cases

- **Agent binary not found**: System displays error with install command (e.g., "Error: 'claude' not found. Install: npm install -g @anthropic-ai/claude-code")
- **Conflicting configuration**: Project-level settings override global settings; documented behavior
- **Launch unconfigured agent**: Agent launches with default settings if binary is available
- **Windows symlink limitations**: Files are copied instead of symlinked; changes to `.agent/` require manual sync or re-run
- **Existing `.agent/` directory**: Setup fails with message "Error: .agent/ exists. Use --force to overwrite."

### Test Cases (Edge Cases)

1. **Given** no agent configuration exists, **When** I run `sl code opencode`, **Then** OpenCode launches with default settings (no API key, no model specified)
2. **Given** `agent.claude.model_aliases.sonnet=claude-sonnet-4-20250514`, **When** I run `sl code claude`, **Then** the `ANTHROPIC_MODEL` env var is NOT set (model aliases are for user reference only, not auto-injected)
3. **Given** `agent.claude.model=claude-sonnet-4-20250514`, **When** I run `sl code claude`, **Then** `ANTHROPIC_MODEL=claude-sonnet-4-20250514` is set in the agent process

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide an `sl code [<agent>]` command to launch a coding agent (optional positional argument for agent selection; defaults to `agent.default` if omitted)
- **FR-002**: System MUST support the following agents: Claude Code, OpenCode, Copilot CLI, Codex
- **FR-003**: System MUST allow agent configuration via `sl config set` at global and project levels
- **FR-004**: System MUST support per-agent arguments via config keys (e.g., `agent.claude.arguments`, `agent.opencode.arguments`)
- **FR-005**: System MUST allow selection of multiple agents during `sl new` and `sl init`
- **FR-006**: System MUST create shared configuration structure from `.agent/commands` and `.agent/skills` to each agent's config directory using symlinks on macOS/Linux and file copies on Windows
- **FR-007**: System MUST use `agent.default` config key to determine which agent to launch when no positional argument is provided
- **FR-008**: System MUST merge configuration from global and project levels with project taking precedence
- **FR-009**: System MUST pass environment variables from config to the launched agent process
- **FR-010**: System MUST detect and report when a selected agent binary is not in PATH
- **FR-011**: System MUST display install command in error message when a selected agent is unavailable (e.g., "Error: 'claude' not found. Install: npm install -g @anthropic-ai/claude-code")
- **FR-012**: System MUST support new config keys: `agent.default`, `agent.<name>.api_key`, `agent.<name>.base_url`, `agent.<name>.model`, `agent.<name>.arguments`, `agent.<name>.env`
- **FR-013**: System MUST require `--force` flag when `.agent/` directory already exists during setup
- **FR-014**: System MUST map per-agent config values to appropriate environment variables (e.g., `agent.claude.api_key` → `ANTHROPIC_API_KEY`)
- **FR-015**: System MUST support Claude-specific model aliases via `agent.claude.model_aliases.sonnet`, `agent.claude.model_aliases.opus`, `agent.claude.model_aliases.haiku`

### Key Entities

- **Agent Definition**: Represents a coding agent with its name, CLI command, install command, and env var mappings
  - Claude Code: `claude` → `npm install -g @anthropic-ai/claude-code` (env: `ANTHROPIC_API_KEY`, `ANTHROPIC_BASE_URL`, `ANTHROPIC_MODEL`)
  - OpenCode: `opencode` → `go install github.com/opencode-ai/opencode@latest` (env: `OPENAI_API_KEY`, `OPENAI_BASE_URL`)
  - Copilot CLI: `github-copilot` → `npm install -g @github/copilot` (env: `GITHUB_TOKEN`)
  - Codex: `codex` → `npm install -g @openai/codex` (env: `OPENAI_API_KEY`, `OPENAI_BASE_URL`)
- **Agent Configuration**: User preferences including default agent, per-agent API keys, base URLs, models, custom arguments, and environment variables
- **Launch Profile**: A combination of agent selection and configuration settings used to launch an agent
- **Shared Configuration Directory**: The `.agent/` directory containing shared commands and skills, symlinked (macOS/Linux) or copied (Windows) to agent-specific directories

### Config Schema

See [plan.md](./plan.md#config-schema-refactoring-2026-03-15) for the complete config schema including per-agent keys, environment variable mappings, and deprecated keys.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can launch any supported coding agent using `sl code` or `sl code <agent>` with sub-second startup overhead
- **SC-002**: Configuration changes via `sl config` are reflected in agent launch within 1 command execution
- **SC-003**: 100% of arguments in `agent.<name>.arguments` are passed directly to the launched agent
- **SC-004**: Users can set up a project with all 4 agents in a single `sl new` or `sl init` session
- **SC-005**: Shared configuration works across all supported platforms (macOS, Linux via symlinks; Windows via file copies)
- **SC-006**: Users receive actionable error messages with install commands when an agent binary is unavailable

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
- Each agent has a unique CLI command name
- Agents accept environment variables for configuration where applicable
- Windows users understand that file copies require manual sync when `.agent/` contents change
