# Research: Configuration Precedence & UX Patterns

**Date**: 2026-02-21
**Feature**: 597-agent-model-config
**Purpose**: Document configuration precedence patterns from established tools to inform `sl config` UX design during planning phase.

## Precedence Patterns in Established Tools

### Git Config
- System (`/etc/gitconfig`) < Global (`~/.gitconfig`) < Local (`.git/config`) < Worktree (`.git/config.worktree`)
- `git config --local` / `--global` / `--system` to target scope
- No interactive prompting on conflicts — layering handles it naturally

### Claude Code Settings
- Managed settings (IT/enterprise) > CLI flags > Local project settings (`.claude/settings.local.json`, gitignored) > Project settings (`.claude/settings.json`, git-tracked) > User settings (`~/.claude/settings.json`)
- `.local.json` pattern for personal overrides that don't go into git
- See research/001 Section 4 for full hierarchy

### npm/Node.js
- `.npmrc` at project, user (`~/.npmrc`), and global levels
- Project-level wins over user-level

### Terraform
- Environment variables > CLI flags > `.tfvars` files > variable defaults
- Workspace-scoped state and variables

## Recommended SpecLedger Precedence

| Priority | Layer | Location | Git Status | Use Case |
|----------|-------|----------|------------|----------|
| 1 (highest) | Personal project override | `specledger/specledger.local.yaml` | gitignored | Developer's personal auth tokens, model preferences |
| 2 | Team project config | `specledger/specledger.yaml` | tracked | Team-shared gateway URL, required models |
| 3 | Global user config | `~/.specledger/config.yaml` | N/A | Personal defaults across all projects |
| 4 | Active profile values | stored in applicable scope | varies | Bundled provider configurations |
| 5 (lowest) | Built-in defaults | compiled into binary | N/A | Sensible out-of-box behavior |

## UX Design Notes (for quickstart/plan phase)

### CLI Subcommands (to be designed in quickstart)
- `sl config set <key> <value>` — set at default scope (local if in project, global otherwise)
- `sl config set --global <key> <value>` — set at global scope
- `sl config get <key>` — show effective value with scope indicator
- `sl config show` — show all settings with scope indicators
- `sl config unset <key>` — remove at default scope
- `sl config profile create|use|list|delete <name>` — profile management
- `sl config` (no subcommand, interactive terminal) — launch TUI

### Sensitive Value Editing in TUI
- Password-style input fields (asterisk masking during entry)
- Reveal/hide toggle — Bubble Tea `textinput` supports `EchoMode`
- Needs feasibility spike: current TUI only has bootstrap screens, no pane-based layouts

### Map/Dictionary Config in TUI
- `agent.env` as key-value editor (add/remove/edit entries)
- Needs spike on Bubble Tea list/table components for key-value editing

### Profile Workflow
- Profiles bundle multiple config keys into a named set
- Active profile acts as a default layer below explicit overrides
- Profile values can be set at either global or project scope
- UX for creating/managing profiles to be designed in quickstart
