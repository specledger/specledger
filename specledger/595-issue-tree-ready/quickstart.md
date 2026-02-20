# Quickstart: Issue Tree View and Ready Command

**Feature**: 595-issue-tree-ready
**Date**: 2026-02-20

## Prerequisites

- SpecLedger CLI installed
- Initialized project with `.specledger/` directory
- At least one spec with issues

## Scenario 1: View Issue Dependencies as Tree

### Setup

```bash
# Create some issues with dependencies
sl issue create --title "Setup database" --type task --priority 1
# Output: Created issue SL-abc123

sl issue create --title "Create schema" --type task --priority 2
# Output: Created issue SL-def456

sl issue create --title "Add indexes" --type task --priority 2
# Output: Created issue SL-ghi789

# Create dependency: Setup blocks Create schema
sl issue link SL-abc123 blocks SL-def456

# Create dependency: Create schema blocks Add indexes
sl issue link SL-def456 blocks SL-ghi789
```

### Execution

```bash
sl issue list --tree
```

### Expected Output

```
595-issue-tree-ready (3 issues)
└── SL-abc123 [open] Setup database
    └── SL-def456 [open] Create schema
        └── SL-ghi789 [open] Add indexes
```

### Validation

- [ ] Tree shows hierarchical structure
- [ ] Indentation and connecting lines visible
- [ ] Issue IDs, status, and titles displayed
- [ ] Parent-child relationships match `blocks` links

---

## Scenario 2: List Ready-to-Work Issues

### Setup

```bash
# Using issues from Scenario 1
# SL-abc123 blocks SL-def456 which blocks SL-ghi789

# Close the blocking issue
sl issue close SL-abc123 --reason "Database setup complete"
```

### Execution

```bash
sl issue ready
```

### Expected Output

```
ID         TITLE                    STATUS  PRIORITY
SL-def456  Create schema            open    2
```

Note: SL-ghi789 is NOT ready because SL-def456 is still open.

### Validation

- [ ] Only unblocked issues shown
- [ ] SL-def456 appears (its blocker SL-abc123 is closed)
- [ ] SL-ghi789 does NOT appear (still blocked by SL-def456)

---

## Scenario 3: All Issues Blocked

### Setup

```bash
# Create mutually blocking issues
sl issue create --title "Task A" --type task
# SL-xxx111

sl issue create --title "Task B" --type task
# SL-xxx222

sl issue link SL-xxx111 blocks SL-xxx222
sl issue link SL-xxx222 blocks SL-xxx111
# This creates a cycle - neither is ready
```

### Execution

```bash
sl issue ready
```

### Expected Output

```
No ready issues found.

Blocked issues:
  SL-xxx111 "Task A" is blocked by:
    - SL-xxx222 "Task B" (open)
  SL-xxx222 "Task B" is blocked by:
    - SL-xxx111 "Task A" (open)
```

### Validation

- [ ] Clear message when no ready issues
- [ ] Shows which issues are blocked
- [ ] Shows blocking issues with their status

---

## Scenario 4: Cross-Spec Ready Issues

### Setup

```bash
# Create issues in current spec
sl issue create --title "Feature task" --type task
# SL-aaa111

# Switch to another spec (simulated)
# In another spec directory:
sl issue create --title "Other task" --type task --spec 591-issue-tracking-upgrade
# SL-bbb222
```

### Execution

```bash
sl issue ready --all
```

### Expected Output

```
ID         TITLE           STATUS  PRIORITY  SPEC
SL-aaa111  Feature task    open    2         595-issue-tree-ready
SL-bbb222  Other task      open    2         591-issue-tracking-upgrade
```

### Validation

- [ ] Issues from all specs shown
- [ ] Spec context column included
- [ ] Only ready issues appear

---

## Scenario 5: JSON Output for Scripting

### Execution

```bash
sl issue ready --json
```

### Expected Output

```json
[
  {
    "id": "SL-def456",
    "title": "Create schema",
    "status": "open",
    "priority": 2,
    "spec_context": "595-issue-tree-ready",
    "created_at": "2026-02-20T10:00:00Z",
    "updated_at": "2026-02-20T10:30:00Z"
  }
]
```

### Validation

- [ ] Valid JSON array
- [ ] All issue fields present
- [ ] Parseable by scripts

---

## Scenario 6: Cycle Detection in Tree

### Setup

```bash
# Create a cycle
sl issue create --title "Task X" --type task
# SL-cyc111

sl issue create --title "Task Y" --type task
# SL-cyc222

sl issue link SL-cyc111 blocks SL-cyc222
sl issue link SL-cyc222 blocks SL-cyc111
```

### Execution

```bash
sl issue list --tree
```

### Expected Output

```
⚠ Warning: Cyclic dependencies detected
  Cycle: SL-cyc111 → SL-cyc222 → SL-cyc111

595-issue-tree-ready (2 issues with cycle)
├── SL-cyc111 [open] Task X ⚠
│   └── SL-cyc222 [open] Task Y ⚠
│       └── SL-cyc111 [open] Task X (cycle)
└── SL-cyc222 [open] Task Y ⚠
    └── SL-cyc111 [open] Task X ⚠
        └── SL-cyc222 [open] Task Y (cycle)
```

### Validation

- [ ] Warning displayed at top
- [ ] Cycle path shown
- [ ] Cyclic nodes marked with ⚠
- [ ] Tree still renders (doesn't crash)

---

## Scenario 7: Implement Workflow Integration

### Execution

```bash
# Run implement workflow
/specledger.implement
```

### Expected Behavior

1. Workflow queries `sl issue ready` instead of `sl issue list --status open`
2. Only unblocked tasks presented as options
3. If all tasks blocked, shows blocking issues

### Sample Workflow Output

```
Checking for ready issues...

Ready tasks available:
  [1] SL-def456: Create schema (priority 2)
  [2] SL-ghi789: Add indexes (priority 2)

Which task would you like to work on?
```

If no ready tasks:
```
Checking for ready issues...

No ready issues found. All open issues are blocked.

Blocked issues:
  SL-xxx111 "Task A" blocked by SL-xxx222 (open)

Would you like to:
  [1] View the dependency tree
  [2] Work on a blocking issue
  [3] Exit
```

### Validation

- [ ] Workflow uses ready command
- [ ] Only unblocked tasks shown
- [ ] Helpful message when all blocked

---

## Test Script

Run all scenarios automatically:

```bash
#!/bin/bash
# test-595-quickstart.sh

set -e

echo "=== Scenario 1: Tree View ==="
# ... setup and validate

echo "=== Scenario 2: Ready Issues ==="
# ... setup and validate

echo "=== Scenario 3: All Blocked ==="
# ... setup and validate

echo "=== Scenario 4: Cross-Spec ==="
# ... setup and validate

echo "=== Scenario 5: JSON Output ==="
# ... setup and validate

echo "=== Scenario 6: Cycle Detection ==="
# ... setup and validate

echo "=== Scenario 7: Implement Integration ==="
# ... manual validation required

echo "All scenarios passed!"
```

---

## Troubleshooting

### "No issues found"

- Ensure you're in a spec directory or use `--all` flag
- Check `sl issue list` to verify issues exist

### Tree not rendering correctly

- Check for broken dependency references (deleted issues)
- Run `sl issue repair` to fix corrupted data

### Ready list seems wrong

- Verify blocker status with `sl issue show <id>`
- Ensure blocker's status is "closed" not just "in_progress"

### Cycle warning appears

- Review dependency chain with `sl issue show <id> --tree`
- Use `sl issue unlink` to break unwanted cycles
