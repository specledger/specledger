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

A developer works on multiple projects with different team requirements. They have global default configuration (their personal preferences and credentials) but one project requires a different provider endpoint mandated by the team. Additionally, each team member may need personal overrides (such as their own auth token) that should not be committed to git. The system supports three configuration layers — global user defaults, team-shared project configuration (git-tracked), and personal project overrides (gitignored) — and merges them with a clear, predictable precedence where more specific layers override broader ones.

**Why this priority**: Multi-project support is essential for professional developers who work across different teams and environments. The personal project override layer (gitignored) solves the critical need for team members to maintain individual settings (e.g., personal auth tokens) alongside shared team configuration, without conflicts. This pattern is well-established in tools like git config and Claude Code's own `.local.json` settings.

**Independent Test**: Can be fully tested by setting a global config value, overriding it at the team project level, then adding a personal project override, and verifying the system displays the merged result with the correct precedence and scope indicators.

**Acceptance Scenarios**:

1. **Given** a global config sets a base URL, **When** no project-level override exists, **Then** viewing configuration displays the global value with a "global" scope indicator.
2. **Given** a global config sets a base URL and a team project config overrides it, **When** the user views configuration, **Then** the project-level value is displayed as the effective value with a "local" scope indicator.
3. **Given** a team project config sets a value and a personal project override (gitignored) exists for the same key, **When** the user views configuration, **Then** the personal override is displayed as the effective value.
4. **Given** a user sets a value targeting the global scope, **When** they view the configuration, **Then** the value is stored in the user home directory configuration, not in the project.
5. **Given** a user removes a personal project override, **When** they view configuration, **Then** the team project value takes effect immediately.

---

### User Story 3 - Custom Agent Profiles (Priority: P2)

A developer frequently switches between different AI provider configurations (e.g., "work" profile using a corporate endpoint, "personal" profile using the default, "experimental" using a cutting-edge model). They define named profiles that bundle together a set of agent configuration values and can quickly switch between them.

**Why this priority**: The original user request is a shell alias that bundles multiple environment variables together — that is essentially a profile. Profiles are the natural abstraction for users who need to switch between provider configurations (e.g., corporate gateway vs. personal API vs. local endpoint) without re-running multiple config commands.

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
- What happens when a team member has personal project overrides that differ from the team-shared configuration? The personal override (gitignored) takes precedence for that user without affecting other team members' configurations.
- What happens when the auth token in configuration has expired? The system should pass it through as-is — token validity is the responsibility of the downstream provider, not SpecLedger.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to configure agent launch environment variables (model names, base URL, auth token) through a persistent configuration mechanism.
- **FR-002**: System MUST support a three-tier configuration hierarchy: global (user home directory), team-local (project-level, git-tracked), and personal-local (project-level, gitignored), where more specific layers override broader ones (personal-local > team-local > global > default).
- **FR-003**: System MUST provide a `sl config` command with subcommands to get, set, list, and unset configuration values.
- **FR-004**: System MUST support `--global` and `--personal` flags on `set` and `unset` to target the global and personal-local (gitignored) scopes respectively. Default (no flag) targets team-local (project-level, git-tracked) when in a project context.
- **FR-005**: System MUST mask sensitive values (auth tokens) when displaying configuration to the terminal.
- **FR-006**: System MUST store sensitive configuration values (auth tokens) in files with restricted permissions (owner read/write only).
- **FR-007**: System MUST inject configured agent environment variables into the agent subprocess when launching. The subprocess MUST inherit the current process environment, with configured values taking precedence over existing environment variables.
- **FR-007a**: System MUST support an `agent.env` configuration key that accepts a map of arbitrary key-value pairs, each injected as an environment variable into the agent subprocess. This enables configuration of agent-specific environment variables beyond the predefined configuration keys.
- **FR-008**: System MUST display the effective value and its source scope (global, local, or default) for each configuration key.
- **FR-011**: System MUST support named agent profiles that bundle multiple agent configuration values together.
- **FR-012**: System MUST validate configuration keys and values, rejecting unknown keys with a helpful error message.

### Key Entities

- **Configuration Key**: A named setting with a defined type (string, boolean, enum, or string-map), scope (global/local), and optional default value. Keys are namespaced by category (e.g., `agent.base-url`, `agent.env`).
- **Configuration Scope**: The level at which a value is set — "default" (built-in), "global" (user home), "team-local" (project, git-tracked), or "personal-local" (project, gitignored). Resolution follows personal-local > team-local > global > profile > default precedence.
- **Agent Profile**: A named collection of agent configuration values (model names, base URL, auth token) that can be activated as a unit. One profile may be active at a time.
- **Agent Environment**: The set of environment variables passed to the agent subprocess at launch time, derived from the resolved configuration (merged defaults, global, local, and profile values).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can configure and launch an agent with custom model settings in under 2 minutes, without needing manual shell aliases or environment variable exports.
- **SC-002**: Configuration changes made via `sl config set` are reflected immediately in subsequent `sl config show` output and agent launches.
- **SC-003**: 100% of configurable settings are accessible through the CLI subcommands (`set`, `get`, `show`, `unset`).
- **SC-004**: Users with multiple projects can maintain different agent configurations per project without interference between projects.
- **SC-005**: Sensitive values (auth tokens) are never displayed in full in terminal output.

### Previous work

- **008-cli-auth**: Established the authentication and credentials system (`~/.specledger/credentials.json`) including secure storage patterns with restricted file permissions.
- **Agent Launcher (pkg/cli/launcher)**: Current agent launching system with `AgentOption` struct and subprocess execution — the foundation this feature extends.
- **Global Config (pkg/cli/config)**: Existing global config system at `~/.config/specledger/config.yaml` with `DefaultConfig()`, `Load()`, `Save()` pattern — will be extended with agent settings.
- **Project Metadata (pkg/cli/metadata)**: Existing `specledger/specledger.yaml` schema — will be extended with local configuration overrides.
- **Bootstrap TUI (pkg/cli/tui)**: Existing Bubble Tea TUI for `sl new` and `sl init` with radio buttons, checkboxes, and text inputs. An interactive config editor TUI was descoped from this feature — see `research/003-tui-framework-spike.md` for findings and a future TUI spec recommendation.

## Dependencies & Assumptions

### Assumptions

- The agent launch mechanism will continue to use subprocess execution with environment variable injection (the standard pattern for CLI tools).
- The existing `~/.config/specledger/config.yaml` location is the appropriate home for global configuration (follows XDG conventions already established).
- Project-level configuration will be added to the existing `specledger/specledger.yaml` file rather than creating a separate config file, to keep the project structure simple.
- Agent profiles are stored alongside configuration (in the same files), not in separate profile files.
- The set of configurable agent environment variables includes at minimum: model names (sonnet, opus, haiku), base URL, and auth token — matching the common `ANTHROPIC_*` environment variable pattern.
- Boolean and enum types for settings are determined by a schema defined in code, not by user input.
- Personal project-level configuration is stored in a gitignored file alongside the team-shared project configuration, following the established `.local` convention used by Claude Code and other tools.
- The agent subprocess inherits the current process environment, with configured values overriding any matching environment variables.
- Secrets management integration (e.g., 1Password, SOPS, Bitwarden, AWS Secrets Manager) is out of scope for this feature. The design should not preclude future integration (e.g., secret interpolation syntax in YAML values).
- Sensitive config fields are identified by Go struct tags (`sensitive:"true"`) on the `AgentConfig` struct. The CLI uses these tags to drive display masking, file permissions (0600), and scope warnings when sensitive values target git-tracked config without `--personal`. This is best-effort guardrailing — teams should additionally adopt pre-commit hooks (e.g., Yelp's `detect-secrets`) for defense-in-depth.
- Interactive TUI config editor was descoped from this feature after a research spike (see `research/003-tui-framework-spike.md`). The existing SpecLedger TUI is step-based form wizards only; a full config editor requires building a reusable TUI shell (tree nav, panes, inline editing) estimated at 6-10 days — recommended as a separate spec that also benefits the revise flow.
