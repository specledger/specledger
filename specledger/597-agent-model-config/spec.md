# Feature Specification: Advanced Agent Model Configuration

**Feature Branch**: `597-agent-model-config`
**Created**: 2026-02-21
**Status**: Draft
**Input**: User description: "Add advanced model selection to the agent launcher - users want to control how the agent is launched with configurable model providers, model names, base URLs, and auth tokens. These should be configurable in specledger.yaml or user home directory with local/global settings hierarchy. Expand configuration management UI to support local/global settings with dropdowns for booleans and enums."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Agent Model Overrides via CLI (Priority: P1)

A developer wants to use an alternative AI provider (e.g., a self-hosted endpoint or third-party API-compatible service) instead of the default Claude Code setup. They run a configuration command to set the model names, base URL, and authentication token. These settings are saved so that every time the agent launches, the correct environment is used automatically — no need for manual shell aliases or environment variable exports.

**Why this priority**: This is the core value proposition. Without the ability to persist agent launch configuration, users must rely on manual shell aliases, which are error-prone, not portable across machines, and not visible to other team members.

**Independent Test**: Can be fully tested by running `sl config set agent.base-url https://api.z.ai/api/anthropic` followed by `sl config show` to verify persistence, and then launching an agent to confirm the environment variables are applied.

**Acceptance Scenarios**:

1. **Given** a user with no existing agent configuration, **When** they run a command to set the agent base URL and model names, **Then** the values are persisted to configuration and displayed when viewing settings.
2. **Given** a user with saved agent model configuration, **When** they launch the agent via `sl new` or `sl init`, **Then** the agent process receives the configured environment variables (model names, base URL, auth token).
3. **Given** a user with a configured auth token, **When** the configuration is saved, **Then** the token value is stored securely with restricted file permissions and masked when displayed in output.

---

### User Story 2 - Local vs Global Configuration Hierarchy (Priority: P2)

A developer works on multiple projects. They have a global default configuration (e.g., their personal API endpoint and credentials) but need to override specific settings for one project (e.g., a different model or base URL required by the team). They set global defaults in their home directory and project-level overrides in the project's `specledger/specledger.yaml`. The system merges these, with local settings taking precedence.

**Why this priority**: Multi-project support is essential for professional developers who work across different teams and environments. Without local/global layering, users must reconfigure for every project switch.

**Independent Test**: Can be fully tested by setting a global config value, then overriding it at the project level, and verifying `sl config show` displays the merged result with the local override winning.

**Acceptance Scenarios**:

1. **Given** a global config sets `agent.base-url` to `https://default.example.com`, **When** no project-level override exists, **Then** `sl config show` displays the global base URL with a "global" scope indicator.
2. **Given** a global config sets `agent.base-url` and a project config overrides it, **When** the user runs `sl config show`, **Then** the project-level value is displayed as the effective value with a "local" scope indicator.
3. **Given** a user sets a value with `--global` flag, **When** they view the configuration, **Then** the value is stored in the user's home directory configuration, not in the project.

---

### User Story 3 - Interactive Configuration Management (Priority: P3)

A developer wants to review and modify all available SpecLedger settings through an interactive terminal interface. They run `sl config` (without subcommands) and see a TUI screen showing all configurable options organized by category. Boolean options show as toggles, enum-style options show as selectable lists, and text fields allow inline editing. They can see which values come from global vs local scope and where overrides exist.

**Why this priority**: A visual configuration interface lowers the barrier to discovery and reduces errors. Users don't need to memorize config key names or valid values — the UI guides them.

**Independent Test**: Can be fully tested by launching the interactive TUI, navigating through all categories, toggling a boolean setting, selecting an enum value, and editing a text field, then verifying the changes persist.

**Acceptance Scenarios**:

1. **Given** a user runs `sl config` interactively, **When** the TUI loads, **Then** all configurable settings are shown organized by category (General, Agent, etc.) with their current values and scope (local/global).
2. **Given** a boolean setting is displayed, **When** the user presses space or enter on it, **Then** the value toggles between enabled/disabled and the change is saved.
3. **Given** an enum setting (e.g., agent choice) is displayed, **When** the user navigates to it, **Then** a list of valid options is shown and the user can select one.
4. **Given** a user modifies a setting in the TUI, **When** they choose to save at local scope, **Then** the value is written to the project configuration; when global, to the home directory configuration.

---

### User Story 4 - Custom Agent Profiles (Priority: P3)

A developer frequently switches between different AI provider configurations (e.g., "work" profile using a corporate endpoint, "personal" profile using the default, "experimental" using a cutting-edge model). They define named profiles that bundle together a set of agent configuration values and can quickly switch between them.

**Why this priority**: Profiles are a convenience feature that builds on the core config system. They reduce friction for users with multiple environments but are not required for the feature to be useful.

**Independent Test**: Can be fully tested by creating two named profiles with different settings, switching between them, and verifying the active profile's values are applied when launching the agent.

**Acceptance Scenarios**:

1. **Given** a user creates a profile named "work" with specific model and URL settings, **When** they activate the "work" profile, **Then** subsequent agent launches use the profile's values.
2. **Given** multiple profiles exist, **When** the user lists profiles, **Then** all profiles are shown with the currently active one highlighted.
3. **Given** a user switches from profile "work" to "personal", **When** they launch the agent, **Then** the new profile's environment variables are used.

---

### Edge Cases

- What happens when a configured agent command is not found in PATH? The system should warn the user and fall back to default behavior (as it does today).
- What happens when both a profile and explicit overrides are set? Explicit overrides should take precedence over the active profile.
- What happens when a user removes a local override? The global value should take effect immediately.
- What happens when configuration keys are invalid or unrecognized? The system should reject them with a clear error message listing valid keys.
- What happens when a user migrates from the old constitution-based agent preference? The system should read the existing preference and offer to migrate it to the new configuration format.
- What happens when the auth token in configuration has expired? The system should pass it through as-is — token validity is the responsibility of the downstream provider, not SpecLedger.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to configure agent launch environment variables (model names, base URL, auth token) through a persistent configuration mechanism.
- **FR-002**: System MUST support a two-tier configuration hierarchy: global (user home directory) and local (project-level), where local settings override global settings.
- **FR-003**: System MUST provide a `sl config` command with subcommands to get, set, list, and unset configuration values.
- **FR-004**: System MUST support a `--global` flag to target the global configuration and default to local (project-level) when in a project context.
- **FR-005**: System MUST mask sensitive values (auth tokens) when displaying configuration to the terminal.
- **FR-006**: System MUST store sensitive configuration values (auth tokens) in files with restricted permissions (owner read/write only).
- **FR-007**: System MUST inject configured agent environment variables into the agent subprocess when launching.
- **FR-008**: System MUST provide an interactive TUI for browsing and editing configuration when `sl config` is run without subcommands in an interactive terminal.
- **FR-009**: The TUI MUST render boolean settings as toggles, enum settings as selectable lists, and text settings as editable fields.
- **FR-010**: System MUST display the effective value and its source scope (global, local, or default) for each configuration key.
- **FR-011**: System MUST support named agent profiles that bundle multiple agent configuration values together.
- **FR-012**: System MUST validate configuration keys and values, rejecting unknown keys with a helpful error message.
- **FR-013**: System MUST migrate existing agent preference from CONSTITUTION.md to the new configuration format when detected, with user confirmation.

### Key Entities

- **Configuration Key**: A named setting with a defined type (string, boolean, enum), scope (global/local), and optional default value. Keys are namespaced by category (e.g., `agent.base-url`, `general.tui-enabled`).
- **Configuration Scope**: The level at which a value is set — "default" (built-in), "global" (user home), or "local" (project). Resolution follows local > global > default precedence.
- **Agent Profile**: A named collection of agent configuration values (model names, base URL, auth token) that can be activated as a unit. One profile may be active at a time.
- **Agent Environment**: The set of environment variables passed to the agent subprocess at launch time, derived from the resolved configuration (merged defaults, global, local, and profile values).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can configure and launch an agent with custom model settings in under 2 minutes, without needing manual shell aliases or environment variable exports.
- **SC-002**: Configuration changes made via `sl config set` are reflected immediately in subsequent `sl config show` output and agent launches.
- **SC-003**: 100% of configurable settings are accessible through both the CLI subcommands and the interactive TUI.
- **SC-004**: Users with multiple projects can maintain different agent configurations per project without interference between projects.
- **SC-005**: Sensitive values (auth tokens) are never displayed in full in terminal output.
- **SC-006**: Existing users with agent preferences in CONSTITUTION.md are offered a migration path to the new system without losing their settings.

### Previous work

- **008-cli-auth**: Established the authentication and credentials system (`~/.specledger/credentials.json`) including secure storage patterns with restricted file permissions.
- **Agent Launcher (pkg/cli/launcher)**: Current agent launching system with `AgentOption` struct and subprocess execution — the foundation this feature extends.
- **Global Config (pkg/cli/config)**: Existing global config system at `~/.config/specledger/config.yaml` with `DefaultConfig()`, `Load()`, `Save()` pattern — will be extended with agent settings.
- **Project Metadata (pkg/cli/metadata)**: Existing `specledger/specledger.yaml` schema — will be extended with local configuration overrides.
- **Bootstrap TUI (pkg/cli/tui)**: Existing Bubble Tea TUI for `sl new` and `sl init` with radio buttons, checkboxes, and text inputs — patterns to reuse for config management TUI.

## Dependencies & Assumptions

### Assumptions

- The agent launch mechanism will continue to use subprocess execution with environment variable injection (the standard pattern for CLI tools).
- The existing `~/.config/specledger/config.yaml` location is the appropriate home for global configuration (follows XDG conventions already established).
- Project-level configuration will be added to the existing `specledger/specledger.yaml` file rather than creating a separate config file, to keep the project structure simple.
- Agent profiles are stored alongside configuration (in the same files), not in separate profile files.
- The set of configurable agent environment variables includes at minimum: model names (sonnet, opus, haiku), base URL, and auth token — matching the common `ANTHROPIC_*` environment variable pattern.
- Boolean and enum types for settings are determined by a schema defined in code, not by user input.
