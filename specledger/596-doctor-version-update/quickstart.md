# Quickstart: Doctor Version and Template Update

**Feature**: 596-doctor-version-update
**Date**: 2026-02-20

## Overview

The `sl doctor` command now includes:
1. **CLI Version Check**: Compares your installed version against the latest GitHub release
2. **Template Update Offers**: Prompts to update project templates when a newer CLI is detected

## Usage Scenarios

### Scenario 1: Check CLI Version

Run `sl doctor` to see your version status:

```bash
$ sl doctor

╭──────────────────────────────────────────────────────────╮
│                  SpecLedger Doctor                        │
│                  Environment Check                        │
╰──────────────────────────────────────────────────────────╯

Core Tools
──────────

  ✓ mise (version 2024.1.3)

SpecLedger CLI
──────────────

  ✓ Version: 1.0.0 (latest: 1.2.0)
  ⚠ Update available! Download from:
    https://github.com/specledger/specledger/releases/latest

╭──────────────────────────────────────────────────────────╮
│           All core tools installed                        │
╰──────────────────────────────────────────────────────────╯
```

### Scenario 2: Update Available - CLI Only

When CLI is outdated but you're not in a project:

```bash
$ cd ~
$ sl doctor

SpecLedger CLI
──────────────

  ⚠ Version: 1.0.0 (latest: 1.2.0)

  Update available!
    brew upgrade specledger          # Homebrew
    go install .../cmd/sl@latest     # Go install
    https://github.com/.../releases  # Binary download
```

### Scenario 3: Template Update Offer

When in a project with outdated templates:

```bash
$ cd my-project
$ sl doctor

╭──────────────────────────────────────────────────────────╮
│                  SpecLedger Doctor                        │
╰──────────────────────────────────────────────────────────╯

Core Tools
──────────

  ✓ mise (version 2024.1.3)

SpecLedger CLI
──────────────

  ✓ Version: 1.2.0 (latest)

Project Templates
─────────────────

  ⚠ Your templates (v1.0.0) are older than CLI (v1.2.0)

  ╭────────────────────────────────────────────────────────╮
  │ Template Update Available                               │
  │                                                         │
  │ Update templates to match CLI v1.2.0?                   │
  │                                                         │
  │   > Yes, update templates                               │
  │     No, skip for now                                    │
  │                                                         │
  ╰────────────────────────────────────────────────────────╯
```

### Scenario 4: Template Update with Customized Files

When some files have been customized:

```bash
$ sl doctor

...

  ╭────────────────────────────────────────────────────────╮
  │ Template Update Available                               │
  │                                                         │
  │ Warning: 2 files have been customized and will be       │
  │ skipped:                                                │
  │   • commands/custom-workflow.md                         │
  │   • skills/my-team-conventions.md                       │
  │                                                         │
  │ Update remaining templates?                              │
  │   > Yes, update templates                               │
  │     No, skip for now                                    │
  │                                                         │
  ╰────────────────────────────────────────────────────────╯
```

After accepting:

```
  Updating templates...
    ✓ .claude/commands/specledger.specify.md
    ✓ .claude/commands/specledger.plan.md
    ✓ .claude/commands/specledger.tasks.md
    ⏭ .claude/commands/custom-workflow.md (customized, skipped)
    ⏭ .claude/skills/my-team-conventions.md (customized, skipped)

  Updated 3 templates, skipped 2 customized files
```

### Scenario 5: Templates Already Current

When templates are up-to-date:

```bash
$ sl doctor

Project Templates
─────────────────

  ✓ Templates: v1.2.0 (current)
```

### Scenario 6: CI/CD Mode (JSON Output)

For automation pipelines:

```bash
$ sl doctor --json
```

```json
{
  "status": "pass",
  "tools": [
    {"name": "mise", "installed": true, "version": "2024.1.3", "category": "core"}
  ],
  "cli_version": "1.0.0",
  "cli_latest_version": "1.2.0",
  "cli_update_available": true,
  "cli_update_instructions": "Download from https://github.com/specledger/specledger/releases/latest",
  "template_version": "0.9.0",
  "template_update_available": true,
  "template_customized_files": ["commands/custom-workflow.md"]
}
```

**Note**: In JSON mode, template updates are NOT performed (no interactive prompt).

### Scenario 7: Offline / Network Error

When version check fails:

```bash
$ sl doctor

SpecLedger CLI
──────────────

  ✓ Version: 1.0.0
  ⚠ Could not check for updates (network unreachable)
```

The doctor command continues with other checks.

## Validation Commands

### Test Version Check

```bash
# Run doctor and check version output
sl doctor | grep -A2 "SpecLedger CLI"
```

### Test JSON Output

```bash
# Validate JSON output format
sl doctor --json | jq '.cli_version, .cli_update_available'
```

### Test Template Update Flow

```bash
# 1. Initialize a project with current CLI
sl init

# 2. Simulate older templates by editing specledger.yaml
# Change template_version to an older version

# 3. Run doctor - should offer template update
sl doctor
```

### Test Customized File Detection

```bash
# 1. Modify a template file
echo "# Custom addition" >> .claude/commands/specledger.specify.md

# 2. Run doctor - should detect customization
sl doctor
```

## Troubleshooting

### Version check always fails

- Check network connectivity
- Verify GitHub API is accessible
- Check if rate limited (unauthenticated: 60 requests/hour)

### Template update corrupts custom files

- Customized files are **never** overwritten
- Check the "skipped" summary after update
- Use git to review changes: `git diff .claude/`

### Template version not updating

- Verify `template_version` in `specledger/specledger.yaml`
- Must accept the update prompt for version to change
