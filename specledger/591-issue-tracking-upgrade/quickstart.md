# Quickstart: Built-In Issue Tracker

**Feature**: 591-issue-tracking-upgrade | **Date**: 2026-02-18

## Installation

The issue tracker is built into the `sl` CLI. No additional installation required.

```bash
# Verify installation
sl issue --help
```

## Quick Start

### Creating Issues

```bash
# Create a simple task
sl issue create --title "Add input validation" --type task

# Create with all options
sl issue create \
  --title "Implement user authentication" \
  --description "OAuth2 integration with JWT tokens" \
  --type feature \
  --priority 1 \
  --labels "component:auth,phase:setup"

# Create a bug
sl issue create \
  --title "Fix login timeout" \
  --type bug \
  --priority 0 \
  --description "Users are logged out after 5 minutes instead of 1 hour"
```

**Output**:
```
Created issue SL-a3f5d8
  Title: Add input validation
  Type: task
  Priority: 1
  Spec: 010-my-feature

View: sl issue show SL-a3f5d8
```

### Listing Issues

```bash
# List all issues in current spec
sl issue list

# List with status filter
sl issue list --status open
sl issue list --status in_progress
sl issue list --status closed

# List with type filter
sl issue list --type bug
sl issue list --type task

# List across all specs
sl issue list --all

# List from specific spec
sl issue list --spec 010-my-feature

# Combine filters
sl issue list --all --status open --type bug
```

**Output** (default table format):
```
ID          TITLE                        STATUS       TYPE      PRIORITY  SPEC
SL-a3f5d8   Add input validation         open         task      1         010-my-feature
SL-b4e6f9   Fix login timeout            in_progress  bug       0         010-my-feature
SL-c7e1a2   Implement OAuth2             open         feature   2         010-my-feature
```

**Output** (JSON format with `--json`):
```json
[
  {
    "id": "SL-a3f5d8",
    "title": "Add input validation",
    "status": "open",
    "issue_type": "task",
    "priority": 1,
    "spec_context": "010-my-feature"
  }
]
```

### Viewing Issue Details

```bash
# Show full issue details
sl issue show SL-a3f5d8

# JSON output for scripting
sl issue show SL-a3f5d8 --json
```

**Output**:
```
Issue: SL-a3f5d8
  Title: Add input validation
  Type: task
  Status: open
  Priority: 1 (high)
  Spec: 010-my-feature

Description:
  Implement input validation for user registration form

Labels:
  - component:api

Created: 2026-02-18 10:00:00
Updated: 2026-02-18 10:00:00

Definition of Done:
  [ ] Unit tests written
  [ ] Code review approved
  [x] Implementation complete
```

### Updating Issues

```bash
# Update status
sl issue update SL-a3f5d8 --status in_progress

# Update priority
sl issue update SL-a3f5d8 --priority 0

# Add notes
sl issue update SL-a3f5d8 --notes "Started implementation, found edge case with email validation"

# Add labels
sl issue update SL-a3f5d8 --add-label "needs-review"

# Remove labels
sl issue update SL-a3f5d8 --remove-label "needs-review"

# Update assignee
sl issue update SL-a3f5d8 --assignee alice
```

### Closing Issues

```bash
# Close an issue
sl issue close SL-a3f5d8

# Close with reason
sl issue close SL-a3f5d8 --reason "Completed in PR #42"

# Force close (skip definition of done check)
sl issue close SL-a3f5d8 --force
```

**Output**:
```
Closing issue SL-a3f5d8...

Checking definition of done:
  ✓ Implementation complete
  ✗ Unit tests written
  ✗ Code review approved

Definition of done not met. Use --force to close anyway.
```

### Managing Dependencies

```bash
# Issue A blocks Issue B (A must complete before B can start)
sl issue link SL-a3f5d8 blocks SL-b4e6f9

# Show dependency tree
sl issue show SL-b4e6f9 --tree

# List all blocked issues
sl issue list --blocked

# Remove dependency
sl issue unlink SL-a3f5d8 blocks SL-b4e6f9
```

**Dependency Tree Output**:
```
SL-b4e6f9: Fix login timeout [in_progress]
└── blocked_by: SL-a3f5d8: Add input validation [open]
```

### Migration from Beads

```bash
# Migrate existing Beads data (also cleans up Beads/Perles dependencies)
sl issue migrate

# Check what would be migrated (dry run)
sl issue migrate --dry-run

# Keep Beads data after migration (don't remove .beads folder)
sl issue migrate --keep-beads
```

**Output**:
```
Migrating Beads issues...

Reading .beads/issues.jsonl...
  Found 45 issues to migrate

Mapping issues to specs:
  010-auth: 12 issues
  011-payment: 8 issues
  012-notifications: 15 issues
  migrated (unmapped): 10 issues

Writing new issue files...
  ✓ specledger/010-auth/issues.jsonl (12 issues)
  ✓ specledger/011-payment/issues.jsonl (8 issues)
  ✓ specledger/012-notifications/issues.jsonl (15 issues)
  ⚠ specledger/migrated/issues.jsonl (10 issues - could not determine spec)

Cleaning up Beads dependencies...
  ✓ Removed .beads/ directory
  ✓ Removed 'beads' from mise.toml
  ✓ Removed 'perles' from mise.toml

Migration complete!
  45 issues migrated
  10 issues need manual review (see specledger/migrated/)

Next steps:
  1. Review issues in specledger/migrated/ and move to appropriate specs
  2. Run 'sl issue list --all' to verify migration
```

### Repair Corrupted Files

```bash
# Repair a corrupted issues.jsonl
sl issue repair --spec 010-my-feature

# Repair all spec files
sl issue repair --all
```

**Output**:
```
Repairing specledger/010-my-feature/issues.jsonl...

Valid lines: 15
Invalid lines: 2
  Line 23: Invalid JSON - skipped
  Line 45: Missing required field 'id' - skipped

Recovered 15 valid issues
2 issues could not be recovered

Backup saved to: specledger/010-my-feature/issues.jsonl.bak
```

## Workflow Examples

### Feature Development Workflow

```bash
# 1. Start working on a feature branch
git checkout -b 020-user-profiles

# 2. Create epic for the feature
sl issue create \
  --title "User Profile Feature" \
  --type epic \
  --description "Complete user profile management"

# 3. Create tasks for the epic
sl issue create --title "Design profile schema" --type task --priority 0
sl issue create --title "Implement profile API" --type task --priority 1
sl issue create --title "Add profile UI" --type task --priority 2

# 4. Link dependencies
sl issue link SL-abc123 blocks SL-def456  # schema blocks API
sl issue link SL-def456 blocks SL-ghi789  # API blocks UI

# 5. Work through tasks
sl issue update SL-abc123 --status in_progress
# ... implement ...
sl issue close SL-abc123 --reason "Schema designed and reviewed"

# 6. Continue with next task
sl issue update SL-def456 --status in_progress
```

### Bug Fix Workflow

```bash
# 1. Create bug report
sl issue create \
  --title "Profile update fails with special characters" \
  --type bug \
  --priority 0 \
  --description "Updating profile with emoji in bio causes 500 error"

# 2. Add definition of done
sl issue update SL-xyz789 --definition-of-done '[
  {"item": "Root cause identified", "checked": false},
  {"item": "Fix implemented", "checked": false},
  {"item": "Test case added", "checked": false},
  {"item": "Code review passed", "checked": false}
]'

# 3. Start fixing
sl issue update SL-xyz789 --status in_progress

# 4. Update DoD as you progress
sl issue update SL-xyz789 --check-dod "Root cause identified"
sl issue update SL-xyz789 --check-dod "Fix implemented"

# 5. Close when all criteria met
sl issue close SL-xyz789
# If DoD not complete, will prompt or require --force
```

### Cross-Spec Planning

```bash
# List all open issues across all specs
sl issue list --all --status open

# Find issues in a specific component
sl issue list --all --label "component:auth"

# Check for duplicate issues across specs
sl issue list --all --check-duplicates
```

## Command Reference

| Command | Description |
|---------|-------------|
| `sl issue create` | Create a new issue |
| `sl issue list` | List issues |
| `sl issue show <id>` | Show issue details |
| `sl issue update <id>` | Update issue fields |
| `sl issue close <id>` | Close an issue |
| `sl issue link <id1> <type> <id2>` | Create dependency |
| `sl issue unlink <id1> <type> <id2>` | Remove dependency |
| `sl issue migrate` | Migrate from Beads |
| `sl issue repair` | Repair corrupted files |

## Flags Reference

### Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |
| `--quiet` | Minimal output (IDs only) |

### create Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--title` | Yes | - | Issue title |
| `--description` | No | "" | Issue description |
| `--type` | No | task | Issue type (epic/feature/task/bug) |
| `--priority` | No | 2 | Priority (0-5, 0=highest) |
| `--labels` | No | [] | Comma-separated labels |
| `--spec` | No | auto | Spec context (auto-detected from branch) |
| `--force` | No | false | Skip duplicate detection |

### list Flags

| Flag | Description |
|------|-------------|
| `--status` | Filter by status (open/in_progress/closed) |
| `--type` | Filter by type (epic/feature/task/bug) |
| `--priority` | Filter by priority (0-5) |
| `--label` | Filter by label (can be repeated) |
| `--all` | List across all specs |
| `--spec <name>` | List from specific spec |
| `--blocked` | Show only blocked issues |
| `--check-duplicates` | Check for similar issues |

### update Flags

| Flag | Description |
|------|-------------|
| `--title` | New title |
| `--description` | New description |
| `--status` | New status |
| `--priority` | New priority |
| `--assignee` | Assignee |
| `--notes` | Update notes |
| `--design` | Update design notes |
| `--acceptance-criteria` | Update acceptance criteria |
| `--add-label` | Add label |
| `--remove-label` | Remove label |
| `--check-dod <item>` | Check a definition of done item |
| `--definition-of-done` | Set full definition of done (JSON) |

### close Flags

| Flag | Description |
|------|-------------|
| `--reason` | Close reason/comment |
| `--force` | Skip definition of done check |

### migrate Flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Show what would be migrated without making changes |
| `--keep-beads` | Keep .beads folder and mise.toml entries after migration |

## Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `Not on a feature branch` | Running outside `###-branch` context | Use `--spec` flag or checkout a feature branch |
| `Issue not found: SL-xxxxxx` | Invalid issue ID | Check ID with `sl issue list` |
| `Spec directory not found` | No spec directory exists | Run spec initialization first |
| `Definition of done not met` | Trying to close issue with incomplete DoD | Complete DoD items or use `--force` |
| `Duplicate detected` | Similar issue already exists | Review duplicates or use `--force` |
| `Cannot create cycle` | Dependency would create circular reference | Remove conflicting dependency |
| `File locked` | Another process has issues.jsonl open | Wait or close other process |
