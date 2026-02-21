# Research: Claude CLI Launch Options & Model Configuration

**Date**: 2026-02-21
**Feature**: 597-agent-model-config
**Purpose**: Catalog all Claude Code CLI launch options, environment variables, model selection mechanisms, and configuration hierarchy to inform the agent model configuration feature design.

## Summary

Claude Code supports model selection through three mechanisms: CLI flags, environment variables, and settings files. There is no public CLI command to list available models (`claude models` does not exist — [open request #12612](https://github.com/anthropics/claude-code/issues/12612)). The Anthropic REST API provides a `/v1/models` endpoint but requires an `ANTHROPIC_API_KEY` (not available to OAuth/subscription users). Custom gateway users (`ANTHROPIC_BASE_URL`) may not have this endpoint at all.

---

## 1. Model Selection Methods

### CLI Flag

```bash
claude --model <alias|full-name>
```

| Alias | Resolves To | Notes |
|-------|-------------|-------|
| `sonnet` | Claude Sonnet 4.6 | |
| `opus` | Claude Opus 4.6 | |
| `haiku` | Claude Haiku 4.5 | |
| `opusplan` | Hybrid | Opus for planning, Sonnet for execution |
| `sonnet[1m]` | Sonnet 4.6 (1M context) | Suffix `[1m]` enables extended context |
| `opus[1m]` | Opus 4.6 (1M context) | Suffix `[1m]` enables extended context |

Full model names also accepted (e.g., `claude-opus-4-6`, `claude-sonnet-4-6`).

### Environment Variable

```bash
ANTHROPIC_MODEL=opus claude
```

### Settings File

```json
{ "model": "opus[1m]" }
```

Stored in `~/.claude/settings.json` (global) or `.claude/settings.json` (project).

---

## 2. Environment Variables — Complete Reference

### Authentication & API

| Variable | Purpose |
|----------|---------|
| `ANTHROPIC_API_KEY` | API key (X-Api-Key header) |
| `ANTHROPIC_AUTH_TOKEN` | Auth token (Authorization header) |
| `ANTHROPIC_BASE_URL` | Custom API endpoint (proxies, gateways, third-party providers) |
| `ANTHROPIC_BETAS` | Comma-separated beta features |
| `ANTHROPIC_CUSTOM_HEADERS` | Custom HTTP headers (JSON) |
| `CLAUDE_CODE_OAUTH_TOKEN` | OAuth authentication token |

### Model Override

| Variable | Purpose |
|----------|---------|
| `ANTHROPIC_MODEL` | Override default model (alias or full name) |
| `ANTHROPIC_DEFAULT_OPUS_MODEL` | Full model name for `opus` alias |
| `ANTHROPIC_DEFAULT_SONNET_MODEL` | Full model name for `sonnet` alias |
| `ANTHROPIC_DEFAULT_HAIKU_MODEL` | Full model name for `haiku` alias |
| `CLAUDE_CODE_SUBAGENT_MODEL` | Model for subagents |
| `ANTHROPIC_SMALL_FAST_MODEL` | **Deprecated** — use `ANTHROPIC_DEFAULT_HAIKU_MODEL` |

### Cloud Providers

| Variable | Purpose |
|----------|---------|
| `CLAUDE_CODE_USE_BEDROCK` | Enable AWS Bedrock |
| `CLAUDE_CODE_SKIP_BEDROCK_AUTH` | Skip Bedrock auth (for proxies) |
| `ANTHROPIC_BEDROCK_BASE_URL` | Custom Bedrock endpoint |
| `CLAUDE_CODE_USE_VERTEX` | Enable Google Vertex AI |
| `CLAUDE_CODE_SKIP_VERTEX_AUTH` | Skip Vertex auth (for proxies) |
| `ANTHROPIC_VERTEX_BASE_URL` | Custom Vertex endpoint |
| `ANTHROPIC_VERTEX_PROJECT_ID` | GCP project ID |
| `CLOUD_ML_REGION` | GCP region for Vertex AI |
| `AWS_REGION`, `AWS_PROFILE`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN` | AWS credentials |

### Performance & Behavior

| Variable | Purpose |
|----------|---------|
| `CLAUDE_CODE_EFFORT_LEVEL` | `low`, `medium`, `high` |
| `MAX_THINKING_TOKENS` | Max reasoning tokens |
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | Max response tokens (default: 32,000) |
| `DISABLE_INTERLEAVED_THINKING` | Disable interleaved reasoning |
| `DISABLE_PROMPT_CACHING` | Disable prompt caching (all models) |
| `DISABLE_PROMPT_CACHING_HAIKU` | Disable for Haiku only |
| `DISABLE_PROMPT_CACHING_SONNET` | Disable for Sonnet only |
| `DISABLE_PROMPT_CACHING_OPUS` | Disable for Opus only |

### Operation Controls

| Variable | Purpose |
|----------|---------|
| `CLAUDE_CODE_ACTION` | Permission mode: `acceptEdits`, `plan`, `bypassPermissions`, `default` |
| `CLAUDE_CODE_SHELL` | Override shell detection |
| `CLAUDE_CODE_SHELL_PREFIX` | Wrap bash commands with a prefix script |
| `CLAUDE_CODE_TMPDIR` | Override temp directory |
| `CLAUDE_CODE_SIMPLE` | `1` for minimal mode (Bash, read, edit only) |
| `CLAUDE_CODE_EXTRA_BODY` | Additional JSON for API requests |
| `CLAUDE_CODE_API_KEY_HELPER_TTL_MS` | Token refresh interval (ms) |

### Timeouts

| Variable | Purpose |
|----------|---------|
| `API_TIMEOUT_MS` | API request timeout |
| `BASH_DEFAULT_TIMEOUT_MS` | Default bash command timeout |
| `BASH_MAX_TIMEOUT_MS` | Maximum bash timeout |
| `BASH_MAX_OUTPUT_LENGTH` | Maximum bash output length |
| `CLAUDE_CODE_MAX_RETRIES` | Request retry attempts |
| `MCP_TIMEOUT` | MCP operation timeout |
| `MCP_TOOL_TIMEOUT` | MCP tool execution timeout |
| `MAX_MCP_OUTPUT_TOKENS` | Max MCP output tokens |

### Telemetry & Updates

| Variable | Purpose |
|----------|---------|
| `CLAUDE_CODE_ENABLE_TELEMETRY` | Enable OpenTelemetry |
| `DISABLE_TELEMETRY` | Disable Statsig telemetry |
| `DISABLE_ERROR_REPORTING` | Disable error reporting |
| `DISABLE_AUTOUPDATER` | Disable auto-updates |

### Proxy

| Variable | Purpose |
|----------|---------|
| `HTTP_PROXY` | HTTP proxy URL |
| `HTTPS_PROXY` | HTTPS proxy URL |
| `NO_PROXY` | Hosts to bypass proxy |

---

## 3. CLI Flags — Complete Reference

### Model & Session

| Flag | Purpose |
|------|---------|
| `--model <alias\|name>` | Set model for session |
| `--effort <level>` | Effort level: `low`, `medium`, `high` |
| `--fallback-model <model>` | Auto-fallback when primary overloaded (print mode) |
| `--continue`, `-c` | Continue most recent conversation |
| `--resume`, `-r` | Resume specific session |
| `--fork-session` | New session ID when resuming |
| `--session-id <uuid>` | Use specific session UUID |
| `--from-pr [value]` | Resume session linked to a PR |

### Permissions & Safety

| Flag | Purpose |
|------|---------|
| `--dangerously-skip-permissions` | Skip all permission prompts |
| `--allow-dangerously-skip-permissions` | Enable bypass as option without activating |
| `--permission-mode <mode>` | `plan`, `bypassPermissions`, `default`, `acceptEdits`, `dontAsk` |
| `--allowedTools <tools>` | Tools that run without prompts |
| `--disallowedTools <tools>` | Tools removed from context |
| `--tools <tools>` | Restrict available built-in tools |

### Prompts & Instructions

| Flag | Purpose |
|------|---------|
| `--system-prompt <prompt>` | Replace entire system prompt |
| `--append-system-prompt <prompt>` | Add to default system prompt |
| `--print`, `-p` | Non-interactive mode |
| `--output-format <format>` | `text`, `json`, `stream-json` |
| `--input-format <format>` | `text`, `stream-json` |
| `--json-schema <schema>` | Validate structured JSON output |

### Agents & Plugins

| Flag | Purpose |
|------|---------|
| `--agent <agent>` | Specify agent for session |
| `--agents <json>` | Define custom subagents via JSON |
| `--plugin-dir <paths>` | Load plugins from directories |
| `--disable-slash-commands` | Disable all skills |

### Configuration & MCP

| Flag | Purpose |
|------|---------|
| `--settings <file-or-json>` | Load additional settings |
| `--setting-sources <sources>` | Comma-separated: `user`, `project`, `local` |
| `--mcp-config <configs>` | Load MCP servers from JSON |
| `--strict-mcp-config` | Only use MCP from `--mcp-config` |

### Budget & Limits

| Flag | Purpose |
|------|---------|
| `--max-budget-usd <amount>` | Max dollar spend (print mode) |
| `--betas <betas>` | Beta headers for API requests |

### Other

| Flag | Purpose |
|------|---------|
| `--add-dir <dirs>` | Additional working directories |
| `--chrome` / `--no-chrome` | Chrome browser integration |
| `--debug [filter]` | Debug mode with optional category filter |
| `--ide` | Auto-connect to IDE |
| `--verbose` | Verbose logging |
| `--worktree`, `-w` | Start in isolated git worktree |
| `--tmux` | Create tmux session for worktree |
| `--version`, `-v` | Print version |

---

## 4. Configuration File Hierarchy

### Precedence (highest to lowest)

| Priority | Source | Location |
|----------|--------|----------|
| 1 | Managed settings (IT/enterprise) | `/Library/Application Support/ClaudeCode/managed-settings.json` (macOS) |
| 2 | CLI flags | Per-invocation |
| 3 | Local project settings | `.claude/settings.local.json` (gitignored) |
| 4 | Project settings | `.claude/settings.json` (git-tracked) |
| 5 | User settings | `~/.claude/settings.json` |

### Key Settings File Properties

| Setting | Type | Purpose |
|---------|------|---------|
| `model` | string | Default model alias or full name |
| `availableModels` | string[] | Restrict model choices (enterprise) |
| `apiKeyHelper` | string | Script path for dynamic auth token |
| `disableBypassPermissionsMode` | string | `"disable"` to block permission bypass |
| `hooks` | object | Lifecycle hooks (PreCompact, SessionStart, etc.) |
| `statusLine` | object | Status bar customization |
| `enabledPlugins` | object | Plugin toggles |

### Directory Structure

```
~/.claude/
├── settings.json           # Global user settings
├── settings.local.json     # Local user settings (not synced)
├── .credentials.json       # API credentials
├── CLAUDE.md               # Global instructions
├── commands/               # Global custom slash commands
├── agents/                 # User-level subagent definitions
└── projects/               # Session history per project

<project>/
├── .claude/
│   ├── settings.json       # Team-shared project settings (git)
│   ├── settings.local.json # Personal project overrides (gitignored)
│   ├── agents/             # Project subagent definitions
│   └── commands/           # Project slash commands
├── .mcp.json               # Project-scoped MCP servers
├── CLAUDE.md               # Project instructions (git)
└── CLAUDE.local.md         # Personal project instructions
```

---

## 5. Model Discovery

### Anthropic REST API

```
GET https://api.anthropic.com/v1/models
Headers:
  anthropic-version: 2023-06-01
  X-Api-Key: $ANTHROPIC_API_KEY
```

**Requires authentication** — returns `authentication_error` without a valid API key.

Response:
```json
{
  "data": [
    {
      "id": "claude-opus-4-6",
      "created_at": "2026-02-04T00:00:00Z",
      "display_name": "Claude Opus 4.6",
      "type": "model"
    }
  ],
  "first_id": "...",
  "last_id": "...",
  "has_more": true
}
```

Pagination: `?limit=N&after_id=...&before_id=...`

### Limitations

- **OAuth/subscription users**: No `ANTHROPIC_API_KEY` available, cannot query endpoint
- **Custom gateway users** (`ANTHROPIC_BASE_URL`): Endpoint may not exist or returns different models
- **No `claude models` CLI command**: Feature requested in [#12612](https://github.com/anthropics/claude-code/issues/12612), not implemented
- **In-session only**: `/model` slash command shows interactive picker, `/status` shows current model

### Recommended Approach for SpecLedger

1. **Ship with hardcoded default model list** — updated with each `sl` release
2. **Allow user-defined custom models** in config (for third-party gateways)
3. **Optionally fetch from `/v1/models`** if `ANTHROPIC_API_KEY` is available and `ANTHROPIC_BASE_URL` is not set (or points to Anthropic)
4. **Cache results** locally with TTL to avoid repeated API calls

---

## 6. Custom Gateway Support

Claude Code supports any API gateway that speaks the Anthropic Messages format:

```bash
# LiteLLM proxy
ANTHROPIC_BASE_URL=https://litellm-server:4000
ANTHROPIC_AUTH_TOKEN=sk-your-key

# LM Studio (local)
ANTHROPIC_BASE_URL=http://localhost:1234
ANTHROPIC_AUTH_TOKEN=lmstudio

# Ollama
ANTHROPIC_BASE_URL=http://localhost:11434
ANTHROPIC_AUTH_TOKEN=ollama

# DeepSeek
ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic

# Z.AI (from user's example)
ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic
ANTHROPIC_AUTH_TOKEN=[token]
```

Gateway must expose: Anthropic Messages API (`/v1/messages`), Bedrock InvokeModel, or Vertex rawPredict format.

---

## 7. Design Implications for `sl config`

### What SpecLedger should manage (environment variables injected at agent launch)

**Core set** (high priority):
- `ANTHROPIC_MODEL` — primary model selection
- `ANTHROPIC_DEFAULT_OPUS_MODEL` — opus alias override
- `ANTHROPIC_DEFAULT_SONNET_MODEL` — sonnet alias override
- `ANTHROPIC_DEFAULT_HAIKU_MODEL` — haiku alias override
- `ANTHROPIC_BASE_URL` — custom endpoint
- `ANTHROPIC_AUTH_TOKEN` — auth token (sensitive)
- `ANTHROPIC_API_KEY` — API key (sensitive)

**Secondary set** (medium priority):
- `CLAUDE_CODE_USE_BEDROCK` — cloud provider toggle
- `CLAUDE_CODE_USE_VERTEX` — cloud provider toggle
- `CLAUDE_CODE_EFFORT_LEVEL` — effort level
- `CLAUDE_CODE_SUBAGENT_MODEL` — subagent model

**CLI flags to support as config** (lower priority):
- `--dangerously-skip-permissions` / `--permission-mode`
- `--allowedTools` / `--disallowedTools`
- `--mcp-config`
- `--append-system-prompt`

### Injection strategy

SpecLedger should inject config as **environment variables** on the agent subprocess, not write to `.claude/settings.json`. This avoids conflicts with Claude's own config hierarchy and keeps concerns separated.

### Sensitive value handling

- `ANTHROPIC_API_KEY` and `ANTHROPIC_AUTH_TOKEN` must be stored with 0600 permissions
- Display masked in `sl config show` output (e.g., `sk-ant-...****`)
- Consider supporting `apiKeyHelper`-style script references in addition to raw values
