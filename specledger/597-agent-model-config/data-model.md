# Data Model: Advanced Agent Model Configuration

**Feature**: 597-agent-model-config
**Date**: 2026-02-21

## Entities

### ConfigKeyDef — Config Key Registry Entry

Defines a single configurable setting. Registered at compile time in the schema registry.

| Field | Type | Description |
|-------|------|-------------|
| `Key` | `string` | Dot-namespaced key (e.g., `agent.base-url`) |
| `Type` | `ConfigKeyType` | One of: `string`, `bool`, `enum`, `string-list`, `string-map` |
| `EnvVar` | `string` | Environment variable to inject (e.g., `ANTHROPIC_BASE_URL`). Empty if key maps to a CLI flag instead. |
| `CLIFlag` | `string` | CLI flag to pass to agent (e.g., `--permission-mode`). Empty if key maps to an env var instead. |
| `Default` | `interface{}` | Default value. `nil` means no default (key is optional). |
| `Sensitive` | `bool` | If true, value is masked in display and stored with 0600 permissions. |
| `Description` | `string` | Human-readable description for TUI and help output. |
| `EnumValues` | `[]string` | Valid values for enum type. Empty for non-enum types. |
| `Category` | `string` | Display category for TUI grouping (e.g., "Provider", "Models", "Launch Flags"). |

**Enum: ConfigKeyType**
```
string | bool | enum | string-list | string-map
```

### AgentConfig — Agent Configuration Values

Holds the values for all agent-related configuration keys. Used in both global config and project metadata.

Fields tagged `sensitive:"true"` trigger CLI guardrails: masked display, 0600 file permissions, and a warning when stored in a git-tracked scope without `--personal`.

| Field | YAML Key | Type | Sensitive | Maps To |
|-------|----------|------|-----------|---------|
| `BaseURL` | `base-url` | `string` | | `ANTHROPIC_BASE_URL` |
| `AuthToken` | `auth-token` | `string` | **yes** | `ANTHROPIC_AUTH_TOKEN` |
| `APIKey` | `api-key` | `string` | **yes** | `ANTHROPIC_API_KEY` |
| `Model` | `model` | `string` | | `ANTHROPIC_MODEL` |
| `ModelSonnet` | `model.sonnet` | `string` | | `ANTHROPIC_DEFAULT_SONNET_MODEL` |
| `ModelOpus` | `model.opus` | `string` | | `ANTHROPIC_DEFAULT_OPUS_MODEL` |
| `ModelHaiku` | `model.haiku` | `string` | | `ANTHROPIC_DEFAULT_HAIKU_MODEL` |
| `SubagentModel` | `subagent-model` | `string` | | `CLAUDE_CODE_SUBAGENT_MODEL` |
| `Provider` | `provider` | `enum` | | Sets `CLAUDE_CODE_USE_BEDROCK` or `CLAUDE_CODE_USE_VERTEX` |
| `PermissionMode` | `permission-mode` | `enum` | | `--permission-mode` flag |
| `SkipPermissions` | `skip-permissions` | `bool` | | `--dangerously-skip-permissions` flag |
| `Effort` | `effort` | `enum` | | `--effort` flag |
| `AllowedTools` | `allowed-tools` | `string-list` | | `--allowedTools` flag |
| `Env` | `env` | `map[string]string` | | Each key-value injected as env var |

**Go struct tag convention**:
```go
type AgentConfig struct {
    AuthToken string `yaml:"auth-token" sensitive:"true"`
    APIKey    string `yaml:"api-key"    sensitive:"true"`
    // ... non-sensitive fields omit the sensitive tag
}
```

The config schema loader reads `sensitive:"true"` struct tags at init time to populate `ConfigKeyDef.Sensitive`, keeping the single source of truth on the Go struct. This drives:
1. **Display masking** — `sl config show` renders `****` for sensitive values
2. **File permissions** — files containing sensitive values written with 0600
3. **Scope warning** — CLI warns when a sensitive field targets a git-tracked scope (team-local) without `--personal`, recommending `--personal` to store in the gitignored `specledger.local.yaml`

**Enum values:**
- `Provider`: `anthropic` (default), `bedrock`, `vertex`
- `PermissionMode`: `default`, `plan`, `bypassPermissions`, `acceptEdits`, `dontAsk`
- `Effort`: `low`, `medium`, `high`

### Profile — Named Agent Configuration Bundle

| Field | YAML Key | Type | Description |
|-------|----------|------|-------------|
| `Name` | *(map key)* | `string` | Profile identifier (e.g., "work", "personal", "local") |
| `Agent` | `agent` | `AgentConfig` | Agent configuration values bundled in this profile |

### ConfigFile — Root Config File Schema

Used for both global config (`~/.config/specledger/config.yaml`) and project metadata extensions.

| Field | YAML Key | Type | Description |
|-------|----------|------|-------------|
| *(existing fields)* | — | — | All existing `Config` / `ProjectMetadata` fields preserved |
| `Agent` | `agent` | `AgentConfig` | Agent configuration values at this scope |
| `Profiles` | `profiles` | `map[string]AgentConfig` | Named profiles |
| `ActiveProfile` | `active-profile` | `string` | Currently active profile name (empty = none) |

### ResolvedConfig — Merged Configuration

The result of merging all config layers. Used at agent launch time.

| Field | Type | Description |
|-------|------|-------------|
| `Values` | `map[string]ResolvedValue` | All resolved config key-value pairs |
| `ActiveProfile` | `string` | Name of the active profile (empty if none) |

### ResolvedValue — Single Resolved Key

| Field | Type | Description |
|-------|------|-------------|
| `Key` | `string` | Config key name |
| `Value` | `interface{}` | Effective value after merge |
| `Source` | `ConfigScope` | Where this value came from |
| `Sensitive` | `bool` | Whether to mask in display |

**Enum: ConfigScope**
```
default | global | profile | team-local | personal-local
```

## Relationships

```
ConfigKeyDef (registry, compile-time)
    └── defines type/validation for → AgentConfig fields

Config (global, ~/.config/specledger/config.yaml)
    ├── Agent: AgentConfig
    ├── Profiles: map[string]AgentConfig
    └── ActiveProfile: string

ProjectMetadata (team-local, specledger/specledger.yaml)
    ├── Agent: AgentConfig
    ├── Profiles: map[string]AgentConfig
    └── ActiveProfile: string

PersonalLocal (specledger/specledger.local.yaml — gitignored)
    ├── Agent: AgentConfig
    └── ActiveProfile: string

ResolvedConfig = merge(defaults, global, profile, team-local, personal-local)
    └── used by → AgentLauncher.BuildEnv() → cmd.Env
```

## State Transitions

### Config Value Lifecycle

```
[unset/default] → sl config set → [set at scope]
[set at scope]  → sl config unset → [unset, falls back to next layer]
[any state]     → sl config show → [display with scope indicator]
```

### Profile Lifecycle

```
[none] → sl config profile create → [exists, inactive]
[exists, inactive] → sl config profile use → [active]
[active] → sl config profile use <other> → [inactive]
[exists] → sl config profile delete → [removed]
```

## Config File Examples

### Global Config (`~/.config/specledger/config.yaml`)

```yaml
default_project_dir: "~/demos"
preferred_shell: "zsh"
tui_enabled: true
auto_install_deps: false
fallback_to_plain_cli: true
log_level: "debug"
theme: "default"
language: "en"

# New agent config section
agent:
  model: "sonnet"
  api-key: "sk-ant-api03-..."

profiles:
  personal:
    agent:
      base-url: ""
      model: "opus"
  work:
    agent:
      base-url: "https://litellm.corp.example.com"
      model.sonnet: "gpt-4-turbo"
      auth-token: "sk-corp-..."
  local:
    agent:
      base-url: "http://localhost:11434"
      model: "llama3"
      skip-permissions: true

active-profile: ""
```

### Team Project Config (`specledger/specledger.yaml`)

```yaml
version: "1.0.0"
project:
  name: "my-project"
  short_code: "MP"
  # ... existing fields ...

# New agent config section (git-tracked, team-shared)
agent:
  base-url: "https://litellm.corp.example.com"
  provider: "anthropic"
  permission-mode: "default"
  env:
    CLAUDE_CODE_EFFORT_LEVEL: "high"
    MAX_THINKING_TOKENS: "16000"
```

### Personal Project Config (`specledger/specledger.local.yaml` — gitignored)

```yaml
agent:
  auth-token: "sk-my-personal-token-..."
  model: "opus"
```
