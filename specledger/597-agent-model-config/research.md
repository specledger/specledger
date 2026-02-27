# Research: 597-agent-model-config

**Date**: 2026-02-21
**Purpose**: Consolidate research findings, resolve unknowns, and document technical decisions for the agent model configuration feature.

## Prior Work

### Issue Tracker (beads)

No existing open or closed issues relate to agent model configuration, profiles, or config management. The current issue tracker (37 total, 34 closed) is entirely under spec `136-revise-comments`. The `597-agent-model-config` branch is new work with no tracked issues yet.

Related closed work that informs this feature:
- **sl-tmq**: Extended `AgentLauncher` with `LaunchWithPrompt` method (P1, foundational)
- **sl-c1h**: US5: Launch Coding Agent with Prompt (P2)
- **sl-dc1**: Implement agent launch and post-agent state detection (P2)

### Existing Research Spikes

- **[001-claude-cli-launch-options.md](research/001-claude-cli-launch-options.md)**: Comprehensive catalog of Claude Code's 50+ environment variables, CLI flags, model selection methods, config file hierarchy, model discovery limitations, and gateway support. Key finding: SpecLedger should inject config as env vars on the subprocess (not write to `.claude/settings.json`), focus on a core set of ~8 high-priority env vars, and provide an `agent.env` catch-all for the remaining 40+.

- **[002-config-precedence-patterns.md](research/002-config-precedence-patterns.md)**: Precedence patterns from git config, Claude Code settings, npm, and Terraform. All use layered file-based config with no interactive prompting on conflicts. Claude Code's `.local.json` pattern directly informs the `specledger.local.yaml` personal override design.

---

## Research Findings

### R1: Current Config System Analysis

**Decision**: Extend existing `Config` struct with nested `AgentConfig`
**Rationale**: The current config system (`pkg/cli/config/config.go`) is a simple 8-field struct with `Load()`/`Save()` operating on `~/.specledger/config.yaml` (consolidated under `~/.specledger/` alongside credentials). Adding a nested `AgentConfig` struct follows YAML nesting conventions and keeps the existing API intact.
**Alternatives considered**: Separate config file for agent settings (rejected — adds file proliferation), flat keys in existing struct (rejected — 20+ new fields would bloat the flat struct).

**Current Config struct** (8 fields):
```go
type Config struct {
    DefaultProjectDir  string `yaml:"default_project_dir"`
    PreferredShell     string `yaml:"preferred_shell"`
    TUIEnabled         bool   `yaml:"tui_enabled"`
    AutoInstallDeps    bool   `yaml:"auto_install_deps"`
    FallbackToPlainCLI bool   `yaml:"fallback_to_plain_cli"`
    LogLevel           string `yaml:"log_level"`
    Theme              string `yaml:"theme"`
    Language           string `yaml:"language"`
}
```

**Proposed extension**: Add `Agent AgentConfig` and `Profiles map[string]AgentConfig` fields.

### R2: Current Launcher Analysis

**Decision**: Add `BuildEnv()` method to `AgentLauncher` for env var injection
**Rationale**: The launcher currently uses `exec.Command()` without explicit `cmd.Env` — meaning the subprocess inherits the parent's full environment. Adding `cmd.Env = append(os.Environ(), resolvedEnvVars...)` preserves existing behavior while adding config-driven overrides.
**Alternatives considered**: Writing to `.claude/settings.json` (rejected — spec explicitly forbids this), using CLI flags only (rejected — not all settings have flag equivalents).

**Current launch pattern** (no env injection):
```go
cmd := exec.Command(l.Command)
cmd.Dir = l.Dir
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
return cmd.Run()
```

**Required change**: Insert `cmd.Env = l.BuildEnv()` before `cmd.Run()`.

### R3: Current Metadata / Project Config Analysis

**Decision**: Add `AgentConfig` section to `ProjectMetadata` for team-local config; introduce `specledger.local.yaml` for personal overrides
**Rationale**: `ProjectMetadata` already manages project-level YAML (`specledger/specledger.yaml`). Adding agent config as a nested section keeps one file for team-shared settings. A parallel `specledger.local.yaml` (gitignored) handles personal overrides without new directories.
**Alternatives considered**: Separate `specledger/agent-config.yaml` (rejected — file proliferation), using `.claude/settings.local.json` directly (rejected — spec mandates SpecLedger manages its own config).

**No local override mechanism exists today** — this is entirely new.

### R4: Config Key Schema Design

**Decision**: Define a `ConfigKeyDef` registry in code that maps config keys to types, env vars, CLI flags, defaults, and validation
**Rationale**: A centralized registry enables: (1) validation at `sl config set` time, (2) type-appropriate TUI rendering, (3) automated env var injection, (4) `sl config show` with scope indicators. This pattern is used by Viper, Cobra, and other Go config libraries.
**Alternatives considered**: Viper library (rejected — over-engineered for this use case, adds large dependency), untyped string map (rejected — no validation, poor TUI experience).

**Core config keys** (from research/001):
| Config Key | Type | Env Var | Sensitive | Default |
|---|---|---|---|---|
| `agent.base-url` | string | `ANTHROPIC_BASE_URL` | no | — |
| `agent.auth-token` | string | `ANTHROPIC_AUTH_TOKEN` | yes | — |
| `agent.api-key` | string | `ANTHROPIC_API_KEY` | yes | — |
| `agent.model` | string | `ANTHROPIC_MODEL` | no | — |
| `agent.model.sonnet` | string | `ANTHROPIC_DEFAULT_SONNET_MODEL` | no | — |
| `agent.model.opus` | string | `ANTHROPIC_DEFAULT_OPUS_MODEL` | no | — |
| `agent.model.haiku` | string | `ANTHROPIC_DEFAULT_HAIKU_MODEL` | no | — |
| `agent.subagent-model` | string | `CLAUDE_CODE_SUBAGENT_MODEL` | no | — |
| `agent.provider` | enum | — | no | `anthropic` |
| `agent.permission-mode` | enum | — | no | `default` |
| `agent.skip-permissions` | bool | — | no | `false` |
| `agent.effort` | enum | — | no | — |
| `agent.allowed-tools` | string-list | — | no | — |
| `agent.env` | string-map | *(each entry)* | no | `{}` |

### R5: Config Merge Strategy

**Decision**: Implement a simple layer-based merge using Go struct embedding and YAML unmarshalling
**Rationale**: Load each config file independently, then merge with a `MergeConfigs(defaults, global, teamLocal, personalLocal, profile)` function that applies non-zero values from each layer. This is simple, testable, and matches how git config works.
**Alternatives considered**: YAML merge keys (`<<:` anchors — rejected, not supported well in Go YAML v3), Viper's automatic merge (rejected — we'd take on the full Viper dependency for one feature).

**Merge algorithm**:
1. Start with built-in defaults
2. Overlay global config (`~/.specledger/config.yaml`)
3. Overlay active profile values (if a profile is active)
4. Overlay team-local config (`specledger/specledger.yaml`)
5. Overlay personal-local config (`specledger/specledger.local.yaml`)
6. For each layer, only non-zero/non-empty values override

### R6: TUI Feasibility

**Decision**: TUI config editor is feasible with current dependencies. Use `huh` forms for editing, `bubbles/list` for navigation, `lipgloss` for layout.
**Rationale**: `huh` v0.8.0 is already in go.mod (indirect). It provides `Input`, `Select`, `Confirm`, `MultiSelect` field types that map directly to config key types (string, enum, boolean, string-list). The `bubbles/table` component handles the key-value display. `lipgloss.JoinHorizontal` handles the two-pane layout (nav + editor).
**Spike needed**: Promote `huh` from indirect to direct dependency. Build a minimal prototype with list navigation + form editing to validate the two-pane pattern.

**Component mapping**:
| Config Type | huh Component | bubbles Fallback |
|---|---|---|
| string | `huh.Input` | `textinput.Model` |
| bool | `huh.Confirm` | manual toggle |
| enum | `huh.Select[string]` | manual radio |
| string-list | `huh.MultiSelect[string]` | manual list |
| string-map | custom `huh.Group` | manual key-value list |

### R7: Profile Storage Format

**Decision**: Store profiles as a YAML map within the config file (both global and project-level)
**Rationale**: Profiles are config bundles — storing them alongside config avoids new files. A profile is simply a named `AgentConfig` struct. The active profile name is a top-level config key.

**YAML structure**:
```yaml
# In config.yaml or specledger.yaml
agent:
  base-url: "https://api.anthropic.com"
  model: "sonnet"

profiles:
  work:
    agent:
      base-url: "https://litellm.corp.example.com"
      model.sonnet: "gpt-4"
      auth-token: "sk-corp-..."
  local:
    agent:
      base-url: "http://localhost:11434"
      model: "llama3"

active-profile: "work"
```

### R8: Interactive TUI Descoped

**Decision**: Remove interactive TUI (config editor) from this spec. Defer to a separate TUI spec.
**Rationale**: A research spike (`research/003-tui-framework-spike.md`) found that the existing SpecLedger TUI is step-based form wizards only (95% hand-rolled, 1 of 20 bubbles components used). A real config editor needs an interactive tree, pane layout, and inline editing — capabilities that don't exist in the codebase today. Building a reusable TUI shell is significant standalone work that also benefits the revise flow.

| Option | Effort | Ships Core Value | Framework Risk |
|---|---|---|---|
| **A: Custom Bubble Tea shell** | 6-10 days | No (delays config) | Low (same stack) |
| **B: Adopt tview** | 3-5 days | No + framework split | High (two stacks) |
| **C: Defer TUI, CLI-only** | 0 extra days | **Yes (fastest)** | None |

**Choice**: Option C — ship CLI-only config now. The core value (replacing shell aliases with persistent config) is fully achievable through `sl config` subcommands. A future TUI spec should build a reusable shell component that serves both config editing and revise comment viewing.

**TUI mockup preserved**: See `quickstart.md` (removed) and `research/003-tui-framework-spike.md` Section 6 for the original TUI design concept.

### R9: Sensitive Value Handling

**Decision**: Minimal approach — file permissions (0600) and display masking. No encryption, no secret manager integration in this feature.
**Rationale**: Sensitive values (auth tokens, API keys) live in personal-local config (gitignored) or global config (user home). File permissions protect at rest. Display masking (`sk-ant-...****`) prevents shoulder-surfing. Secret manager integration (1Password, SOPS, etc.) is explicitly out of scope per spec assumptions.
**Future path**: Config schema includes an `apiKeyHelper`-style field (reference to a script that returns the token) — this is the same pattern Claude Code uses and enables integration with any secret manager.
