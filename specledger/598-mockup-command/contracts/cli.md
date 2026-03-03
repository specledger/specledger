# CLI Contract: Mockup Command

**Branch**: `598-mockup-command` | **Date**: 2026-02-27

## Commands

### sl mockup [prompt...]

Generate a mockup from a feature specification. Spec is auto-detected from branch.

**Usage**:
```bash
sl mockup                              # Interactive flow with confirmations
sl mockup focus on login form          # Skip confirmations, launch agent directly
sl mockup -y                           # Auto-confirm all prompts
sl mockup --format jsx                 # Specify output format
sl mockup --force                      # Bypass frontend detection
sl mockup --dry-run                    # Write prompt to file, skip agent
sl mockup --json                       # Non-interactive JSON output
```

**Arguments**:
| Argument | Required | Description |
|----------|----------|-------------|
| `prompt...` | No | Additional instructions for the AI agent. When provided, confirmations are skipped. |

**Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--format` | | string | `html` | Output format: `html` or `jsx` |
| `--force` | `-f` | bool | false | Bypass frontend detection check |
| `--dry-run` | | bool | false | Write prompt to file instead of launching agent |
| `--summary` | | bool | false | Compact output for agent/CI integration |
| `--json` | | bool | false | Non-interactive path, output result as JSON |
| `--yes` | `-y` | bool | false | Auto-confirm all prompts and launch agent directly |
| `--prompt` | `-p` | string | | Additional instructions for the AI agent |

**Inputs**:
- `specledger/<spec-name>/spec.md` - Feature specification (required)
- `.specledger/memory/design-system.md` - Design system index (auto-generated if missing)

**Outputs**:
- `specledger/<spec-name>/mockup.html` - Generated mockup (HTML format, default) — created by AI agent
- `specledger/<spec-name>/mockup.jsx` - Generated mockup (JSX format, when `--format jsx`) — created by AI agent
- `specledger/<spec-name>/mockup-prompt.md` - Agent prompt (when `--dry-run`)
- `.specledger/memory/design-system.md` - Created if missing

**Exit Codes**:
| Code | Meaning |
|------|---------|
| 0 | Success (or user cancelled) |
| 1 | Spec not found |
| 2 | Not a frontend project (without `--force`) |
| 3 | Spec has no user scenarios |
| 4 | File write error |
| 5 | Invalid format value |
| 6 | Agent launch failed |

**Example Output** (interactive - `sl mockup`):
```
Resolving spec...
✓ Detected spec from branch: 042-user-registration

Detecting frontend framework...
✓ Detected: Next.js (confidence: 99%)

Design system not found.
Generate design system now? [Y/n]: y
✓ Extracted design tokens
✓ Created .specledger/memory/design-system.md

Generating prompt...
✓ Prompt generated (estimated 2,340 tokens)

Mockup prompt is ready.
  > Review/edit the prompt in vim
    Proceed with the generated prompt

[editor opens with prompt]

Prompt ready - what next?
  > Launch AI agent with this prompt
    Re-edit the prompt
    Write prompt to a file
    Cancel

Launching Claude Code...
[agent session runs - agent asks user about commit]

✓ Mockup saved to specledger/042-user-registration/mockup.html
```

**Example Output** (with prompt - `sl mockup focus on login form`):
```
Resolving spec...
✓ Detected spec from branch: 042-user-registration

Detecting frontend framework...
✓ Detected: Next.js (confidence: 99%)

✓ Loaded design system

Generating prompt...
✓ Prompt generated (estimated 2,450 tokens)

Launching Claude Code...
[agent session runs - agent asks user about commit]

✓ Mockup saved to specledger/042-user-registration/mockup.html
```

**Example Output** (dry-run):
```
Resolving spec...
✓ Detected spec from branch: 042-user-registration

Detecting frontend framework...
✓ Detected: Next.js (confidence: 99%)

[... interactive steps ...]

✓ Prompt written to specledger/042-user-registration/mockup-prompt.md
  Run your agent manually with this prompt, or re-run without --dry-run.
```

**Example Output** (JSON — non-interactive):
```json
{
  "status": "success",
  "framework": "nextjs",
  "spec_name": "042-user-registration",
  "mockup_path": "specledger/042-user-registration/mockup.html",
  "prompt_path": "",
  "format": "html",
  "design_system_created": true,
  "agent_launched": true,
  "committed": true
}
```

**Example Output** (error - not frontend):
```
Error: Not a frontend project

No frontend framework detected in this repository.
Use --force to bypass this check, or run from a frontend project directory.
```

---

### sl mockup update

Refresh the design system by re-extracting global CSS and design tokens.

**Usage**:
```bash
sl mockup update
sl mockup update --json
```

**Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | | bool | false | Output result as JSON |

**Inputs**:
- `.specledger/memory/design-system.md` - Existing design system (required)

**Outputs**:
- `.specledger/memory/design-system.md` - Updated design system

**Exit Codes**:
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Design system not found |
| 2 | Not a frontend project |
| 3 | File write error |

**Example Output** (success):
```
Re-extract design tokens? [Y/n]: y
Re-extracting design tokens...
✓ Extracted design tokens
✓ Updated .specledger/memory/design-system.md
```

**Example Output** (JSON):
```json
{
  "status": "success",
  "scan_duration_ms": 1800
}
```

---

## Integration: sl init

When `sl init` runs on a frontend project:

**Modified Behavior**:
1. After standard initialization, detect frontend framework
2. If frontend detected, prompt user: "Initialize design system? [Y/n]"
3. If yes (default), scan components and create `.specledger/memory/design-system.md`
4. If no, skip design system initialization

**Non-Interactive Mode** (`sl init --ci`):
- Auto-create design system if frontend detected
- No prompt shown

**Example Output** (interactive):
```
✓ Created specledger/specledger.yaml
✓ Created .specledger/ directory

Detected frontend framework: React
Initialize design system? [Y/n]: y
✓ Scanned 23 components in 1.2s
✓ Created .specledger/memory/design-system.md

SpecLedger initialized successfully!
```

---

## Error Messages

| Scenario | Message |
|----------|---------|
| Spec not found | `Error: Spec 'xyz' not found\n\nNo spec.md file at specledger/xyz/spec.md\nCreate a spec first with: sl specify xyz` |
| No spec on branch | `Error: Cannot detect spec\n\nNot on a feature branch and no spec-name provided.\nProvide a spec name: sl mockup <spec-name>` |
| No user scenarios | `Error: Spec has no user scenarios\n\nThe spec.md file has no user scenarios to generate mockups from.\nAdd user scenarios with: sl clarify xyz` |
| Not frontend | `Error: Not a frontend project\n\nNo frontend framework detected in this repository.\nUse --force to bypass this check, or run from a frontend project directory.` |
| Design system missing (update) | `Error: Design system not found\n\nNo design system at .specledger/memory/design-system.md\nGenerate one first with: sl mockup <spec-name>` |
| Invalid format | `Error: Invalid format 'xyz'\n\nSupported formats: html, jsx` |
| Write permission | `Error: Cannot write to specledger/\n\nCheck file permissions and try again.` |
| Agent not found | `Error: No AI agent available\n\nPrompt written to specledger/<spec-name>/mockup-prompt.md\nInstall Claude Code: npm install -g @anthropic-ai/claude-code` |
| User cancelled | _(clean exit, no error message, exit code 0)_ |
| Mockup exists | `Mockup already exists at specledger/<spec-name>/mockup.html\nOverwrite? [y/N]:` |

---

## Help Text

### sl mockup --help

```
Generate UI mockups from feature specifications.

Auto-detects spec from current branch. Pass instructions for the AI agent.

Usage:
  sl mockup [prompt...] [flags]
  sl mockup [command]

Available Commands:
  update      Refresh the design system by re-extracting global CSS and design tokens

Examples:
  sl mockup                                    # Interactive flow
  sl mockup help me gen mockup ui for spec     # With custom instructions
  sl mockup focus on the login form            # With custom instructions
  sl mockup -y                                 # Auto-confirm all prompts
  sl mockup --format jsx                       # Generate JSX mockup

Flags:
      --format string   Output format: html or jsx (default "html")
  -f, --force           Bypass frontend detection check
      --dry-run         Write prompt to file instead of launching agent
      --summary         Compact output for agent/CI integration
      --json            Non-interactive path, output result as JSON
  -y, --yes             Auto-confirm all prompts and launch agent directly
  -p, --prompt string   Additional instructions for the AI agent
  -h, --help            help for mockup

Use "sl mockup [command] --help" for more information about a command.
```

### sl mockup update --help

```
Refresh the design system by re-extracting global CSS and design tokens.

Re-extracts CSS variables, theme colors, and styling patterns.

Usage:
  sl mockup update [flags]

Examples:
  sl mockup update          # Interactive update
  sl mockup update --json   # Output as JSON

Flags:
      --json   Output result as JSON
  -h, --help   help for update
```
