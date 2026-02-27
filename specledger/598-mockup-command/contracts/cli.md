# CLI Contract: Mockup Command

**Branch**: `598-mockup-command` | **Date**: 2026-02-27

## Commands

### sl mockup [spec-name]

Generate a mockup from a feature specification via an interactive TUI flow.

**Usage**:
```bash
sl mockup                              # Auto-detect spec from branch
sl mockup 042-user-registration        # Explicit spec name
sl mockup --format jsx                 # Specify output format
sl mockup --force                      # Bypass frontend detection
sl mockup --dry-run                    # Write prompt to file, skip agent
sl mockup --summary                    # Compact output (agent/CI)
sl mockup --json                       # Non-interactive JSON output
```

**Arguments**:
| Argument | Required | Description |
|----------|----------|-------------|
| `spec-name` | No | Name of the spec directory (e.g., `042-user-registration`). Auto-detected from branch if omitted. |

**Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--format` | | string | `html` | Output format: `html` or `jsx` |
| `--force` | `-f` | bool | false | Bypass frontend detection check |
| `--dry-run` | | bool | false | Write prompt to file instead of launching agent |
| `--summary` | | bool | false | Compact output for agent/CI integration |
| `--json` | | bool | false | Non-interactive path, output result as JSON |

**Inputs**:
- `specledger/<spec-name>/spec.md` - Feature specification (required)
- `specledger/design_system.md` - Design system index (auto-generated if missing)

**Outputs**:
- `specledger/<spec-name>/mockup.html` - Generated mockup (HTML format, default) — created by AI agent
- `specledger/<spec-name>/mockup.jsx` - Generated mockup (JSX format, when `--format jsx`) — created by AI agent
- `specledger/<spec-name>/mockup-prompt.md` - Agent prompt (when `--dry-run`)
- `specledger/design_system.md` - Created if missing

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

**Example Output** (interactive success):
```
Resolving spec...
✓ Detected spec from branch: 042-user-registration

Detecting frontend framework...
✓ Detected: Next.js (confidence: 99%)
  Confirm framework? [Y/n]: y

Design system not found. Generate now? [Y/n]: y
✓ Scanned 47 components in 2.3s
✓ Created specledger/design_system.md

Select components to include in mockup:
  src/components/
    [x] Button (src/components/Button.tsx)
    [x] Form (src/components/Form.tsx)
    [x] Card (src/components/Card.tsx)
    [ ] Sidebar (src/components/Sidebar.tsx)
    [ ] Avatar (src/components/Avatar.tsx)
  src/components/layout/
    [ ] Header (src/components/layout/Header.tsx)
    [ ] Footer (src/components/layout/Footer.tsx)
  External (Material UI)
    [ ] TextField (@mui/material)
    [ ] Dialog (@mui/material)
  ... (47 total, grouped by directory)

Output format:
  > html
    jsx

Generating prompt...
✓ Prompt generated (estimated 2,340 tokens)

Opening editor for prompt review...
[editor opens with prompt]

What would you like to do?
  > Launch agent
    Re-edit prompt
    Write prompt to file
    Cancel

Launching Claude Code...
[agent session runs]

Agent session complete.
Changed files:
  M specledger/042-user-registration/mockup.html

Commit and push? [Y/n]: y
Select files to commit:
  [x] specledger/042-user-registration/mockup.html

Commit message: feat: generate mockup for 042-user-registration
✓ Committed: abc1234
✓ Pushed to origin/042-user-registration

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
  "components_scanned": 47,
  "components_selected": 3,
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

Refresh the design system index by rescanning the codebase.

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
- `specledger/design_system.md` - Existing design system (required)

**Outputs**:
- `specledger/design_system.md` - Updated design system

**Exit Codes**:
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Design system not found |
| 2 | Not a frontend project |
| 3 | File write error |

**Example Output** (success):
```
Updating design system index...
  Rescan components? [Y/n]: y
✓ Scanned 52 components in 1.8s
✓ Added 5 new components
✓ Removed 2 stale components
✓ Updated specledger/design_system.md
```

**Example Output** (JSON):
```json
{
  "status": "success",
  "components_total": 52,
  "components_added": 5,
  "components_removed": 2,
  "scan_duration_ms": 1800
}
```

---

## Integration: sl init

When `sl init` runs on a frontend project:

**Modified Behavior**:
1. After standard initialization, detect frontend framework
2. If frontend detected, prompt user: "Initialize design system? [Y/n]"
3. If yes (default), scan components and create `specledger/design_system.md`
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
✓ Created specledger/design_system.md

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
| Design system missing (update) | `Error: Design system not found\n\nNo design system at specledger/design_system.md\nGenerate one first with: sl mockup <spec-name>` |
| Invalid format | `Error: Invalid format 'xyz'\n\nSupported formats: html, jsx` |
| Write permission | `Error: Cannot write to specledger/\n\nCheck file permissions and try again.` |
| Agent not found | `Error: No AI agent available\n\nPrompt written to specledger/<spec-name>/mockup-prompt.md\nInstall Claude Code: npm install -g @anthropic-ai/claude-code` |
| User cancelled | _(clean exit, no error message, exit code 0)_ |
| Mockup exists | `Mockup already exists at specledger/<spec-name>/mockup.html\nOverwrite? [y/N]:` |

---

## Help Text

### sl mockup --help

```
Generate UI mockups from feature specifications using an interactive flow

Usage:
  sl mockup [spec-name] [flags]
  sl mockup [command]

Available Commands:
  update      Refresh the design system index

Examples:
  # Generate mockup (auto-detect spec from branch)
  sl mockup

  # Generate mockup for a specific spec
  sl mockup 042-user-registration

  # Generate mockup as JSX
  sl mockup 042-user-registration --format jsx

  # Write prompt to file without launching agent
  sl mockup --dry-run

  # Force generation even if not a frontend project
  sl mockup 042-user-registration --force

  # Non-interactive JSON output
  sl mockup 042-user-registration --json

Flags:
      --format string   Output format: html or jsx (default "html")
  -f, --force           Bypass frontend detection check
      --dry-run         Write prompt to file instead of launching agent
      --summary         Compact output for agent/CI integration
      --json            Non-interactive path, output result as JSON
  -h, --help            help for mockup

Use "sl mockup [command] --help" for more information about a command.
```

### sl mockup update --help

```
Refresh the design system index by rescanning the codebase

Usage:
  sl mockup update [flags]

Examples:
  # Update design system
  sl mockup update

  # Output as JSON
  sl mockup update --json

Flags:
      --json   Output result as JSON
  -h, --help   help for update
```
