# CLI Contract: Mockup Command

**Branch**: `598-mockup-command` | **Date**: 2026-02-27

## Commands

### sl mockup \<spec-name\>

Generate a mockup from a feature specification.

**Usage**:
```bash
sl mockup <spec-name>
sl mockup <spec-name> --force
sl mockup <spec-name> --json
```

**Arguments**:
| Argument | Required | Description |
|----------|----------|-------------|
| `spec-name` | Yes | Name of the spec directory (e.g., `042-user-registration`) |

**Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--force` | `-f` | bool | false | Bypass frontend detection check |
| `--json` | | bool | false | Output result as JSON |

**Inputs**:
- `specledger/<spec-name>/spec.md` - Feature specification (required)
- `specledger/design_system.md` - Design system index (auto-generated if missing)

**Outputs**:
- `specledger/<spec-name>/mockup.md` - Generated mockup
- `specledger/design_system.md` - Created if missing

**Exit Codes**:
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Spec not found |
| 2 | Not a frontend project (without `--force`) |
| 3 | Spec has no user scenarios |
| 4 | File write error |

**Example Output** (success):
```
Detecting frontend framework...
✓ Detected: Next.js (confidence: 99%)

Design system not found. Generating...
✓ Scanned 47 components in 2.3s
✓ Created specledger/design_system.md

Generating mockup for 042-user-registration...
✓ Generated 3 screens from 2 user stories
✓ Mockup saved to specledger/042-user-registration/mockup.md
```

**Example Output** (JSON):
```json
{
  "status": "success",
  "framework": "nextjs",
  "spec_name": "042-user-registration",
  "mockup_path": "specledger/042-user-registration/mockup.md",
  "design_system_created": true,
  "components_scanned": 47,
  "screens_generated": 3
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
| No user scenarios | `Error: Spec has no user scenarios\n\nThe spec.md file has no user scenarios to generate mockups from.\nAdd user scenarios with: sl clarify xyz` |
| Not frontend | `Error: Not a frontend project\n\nNo frontend framework detected in this repository.\nUse --force to bypass this check, or run from a frontend project directory.` |
| Design system missing (update) | `Error: Design system not found\n\nNo design system at specledger/design_system.md\nGenerate one first with: sl mockup <spec-name>` |
| Write permission | `Error: Cannot write to specledger/\n\nCheck file permissions and try again.` |

---

## Help Text

### sl mockup --help

```
Generate UI mockups from feature specifications

Usage:
  sl mockup <spec-name> [flags]
  sl mockup [command]

Available Commands:
  update      Refresh the design system index

Examples:
  # Generate mockup for a spec
  sl mockup 042-user-registration

  # Force generation even if not a frontend project
  sl mockup 042-user-registration --force

  # Output as JSON
  sl mockup 042-user-registration --json

Flags:
  -f, --force   Bypass frontend detection check
      --json    Output result as JSON
  -h, --help    help for mockup

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
