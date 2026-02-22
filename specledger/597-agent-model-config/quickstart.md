# Quickstart: Agent Model Configuration

**Feature**: 597-agent-model-config

This document shows the user experience for configuring agent launch settings in SpecLedger.

---

## 1. Replace a Shell Alias with Persistent Config

**Before** (brittle shell alias):
```bash
alias zc='ANTHROPIC_DEFAULT_SONNET_MODEL=glm-4.7-flash \
  ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5 \
  ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic \
  ANTHROPIC_AUTH_TOKEN=[token] \
  claude code --dangerously-skip-permissions'
```

**After** (persistent SpecLedger config):
```bash
# Set agent configuration (saved to project config)
sl config set agent.base-url https://api.z.ai/api/anthropic
sl config set agent.model.sonnet glm-4.7-flash
sl config set agent.model.opus glm-5
sl config set --personal agent.auth-token [token]
sl config set agent.skip-permissions true

# Verify
sl config show
```

> **Sensitive values**: Fields tagged as sensitive in the config schema (e.g.,
> `auth-token`, `api-key`) should be stored with `--personal` to keep them in
> `specledger/specledger.local.yaml` (gitignored). The CLI warns when a
> sensitive field is stored in a git-tracked scope without `--personal`.
> Teams should also consider entropy-based or keyword pre-commit hooks
> (e.g., Yelp's `detect-secrets`) as an additional safeguard.

**Output of `sl config show`**:
```
Agent Configuration
  agent.base-url          https://api.z.ai/api/anthropic    [local]
  agent.model.sonnet      glm-4.7-flash                     [local]
  agent.model.opus        glm-5                              [local]
  agent.auth-token        ****[token-last4]                  [local]
  agent.skip-permissions  true                                [local]

General
  tui_enabled             true                                [global]
  log_level               debug                               [default]
  ...
```

Now every `sl new` or `sl init` launch automatically applies these settings.

---

## 2. Create and Use Profiles

```bash
# Create a "work" profile for the corporate gateway
sl config profile create work
sl config set agent.base-url https://litellm.corp.example.com --profile work
sl config set agent.model.sonnet gpt-4-turbo --profile work
sl config set --personal agent.auth-token sk-corp-... --profile work

# Create a "local" profile for Ollama
sl config profile create local
sl config set agent.base-url http://localhost:11434 --profile local
sl config set agent.model llama3 --profile local
sl config set agent.skip-permissions true --profile local

# List profiles
sl config profile list
```

**Output**:
```
Profiles:
  work     https://litellm.corp.example.com  (3 settings)
  local    http://localhost:11434              (3 settings)

No active profile. Use: sl config profile use <name>
```

```bash
# Activate a profile
sl config profile use work

# Verify — profile values shown with [profile] scope
sl config show
```

**Output**:
```
Active Profile: work

Agent Configuration
  agent.base-url          https://litellm.corp.example.com   [profile:work]
  agent.model.sonnet      gpt-4-turbo                        [profile:work]
  agent.auth-token        ****corp                            [profile:work]
  ...
```

```bash
# Switch profiles
sl config profile use local

# Deactivate all profiles
sl config profile use --none
```

---

## 3. Local vs Global Configuration

```bash
# Set a global default (applies to all projects)
sl config set --global agent.api-key sk-ant-api03-...
sl config set --global agent.model sonnet

# In a specific project, override the base URL (team-shared)
cd ~/projects/corp-project
sl config set agent.base-url https://litellm.corp.example.com

# Add a personal override (gitignored, not shared with team)
sl config set --personal agent.auth-token sk-my-personal-...
```

**Output of `sl config show` in the project**:
```
Agent Configuration
  agent.base-url          https://litellm.corp.example.com   [local]
  agent.auth-token        ****onal                            [personal]
  agent.api-key           ****03-                             [global]
  agent.model             sonnet                              [global]
```

**Precedence** (highest to lowest):
1. Personal project override (`specledger/specledger.local.yaml` — gitignored)
2. Team project config (`specledger/specledger.yaml` — git tracked)
3. Global user config (`~/.config/specledger/config.yaml`)
4. Active profile values
5. Built-in defaults

---

## 4. Custom Environment Variables

For agent environment variables not covered by the built-in config keys:

```bash
# Set arbitrary env vars via the agent.env map
sl config set agent.env.CLAUDE_CODE_EFFORT_LEVEL high
sl config set agent.env.MAX_THINKING_TOKENS 16000
sl config set agent.env.DISABLE_PROMPT_CACHING true
```

**Output of `sl config show`**:
```
Agent Configuration
  ...
  agent.env
    CLAUDE_CODE_EFFORT_LEVEL    high                          [local]
    MAX_THINKING_TOKENS          16000                         [local]
    DISABLE_PROMPT_CACHING       true                          [local]
```

All `agent.env` entries are injected as environment variables on the agent subprocess.

---

## 5. Migration from Shell Alias

When SpecLedger detects an agent preference in `CONSTITUTION.md`:

```
ℹ  Found agent preference "Claude Code" in CONSTITUTION.md
   This can be migrated to the new sl config system.

   Migrate now? [Y/n]
```

On confirmation, the preference is written to config and subsequent agent launches use the config system.

---

## 6. Config Key Reference

| Key | Type | Description |
|-----|------|-------------|
| `agent.base-url` | string | Custom API endpoint URL |
| `agent.auth-token` | string | Auth token (sensitive, masked) |
| `agent.api-key` | string | API key (sensitive, masked) |
| `agent.model` | string | Default model (alias or full name) |
| `agent.model.sonnet` | string | Model for "sonnet" alias |
| `agent.model.opus` | string | Model for "opus" alias |
| `agent.model.haiku` | string | Model for "haiku" alias |
| `agent.subagent-model` | string | Model for subagents |
| `agent.provider` | enum | Provider: `anthropic`, `bedrock`, `vertex` |
| `agent.permission-mode` | enum | `default`, `plan`, `bypassPermissions`, `acceptEdits`, `dontAsk` |
| `agent.skip-permissions` | bool | Skip permission prompts |
| `agent.effort` | enum | Effort level: `low`, `medium`, `high` |
| `agent.allowed-tools` | list | Tools allowed without prompts |
| `agent.env` | map | Arbitrary env vars injected into agent |
