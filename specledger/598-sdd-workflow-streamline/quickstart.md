# Quickstart: SDD Workflow Streamline

**Branch**: `598-sdd-workflow-streamline`

## Work Stream 1: `sl comment` (Comment CRUD)

### Setup
```bash
# Ensure authenticated
sl auth status

# Build latest sl binary
make build
```

### Test: List Comments
```bash
# List open comments for current spec
bin/sl comment list --status open

# JSON output for agent consumption
bin/sl comment list --status open --json

# Filter by spec
bin/sl comment list --spec 136-revise-comments --status all
```

### Test: Show Comment Details
```bash
# Show a specific comment with thread context
bin/sl comment show df77879f-416a-4d98-a130-5015d3546433
```

### Test: Reply to Comment
```bash
# Post a reply
bin/sl comment reply df77879f-416a-4d98-a130-5015d3546433 --content "Addressed in this PR"
```

### Test: Resolve Comment
```bash
# Resolve with reason (posts reply first, then marks resolved)
bin/sl comment resolve df77879f-416a-4d98-a130-5015d3546433 --reason "Fixed in commit abc123"

# Resolve without reason
bin/sl comment resolve df77879f-416a-4d98-a130-5015d3546433
```

### Verify: Refactored `sl revise` Still Works
```bash
# Interactive mode (launcher pattern)
bin/sl revise

# Summary mode (should still work, now delegates to comment package)
bin/sl revise --summary

# Auto mode with fixture
bin/sl revise --auto fixture.json --dry-run
```

---

## Work Stream 2: `sl spec` + `sl context` (Bash Replacement)

### Test: `sl spec info` (replaces check-prerequisites.sh)
```bash
# Paths only (JSON)
bin/sl spec info --json --paths-only

# With validation
bin/sl spec info --json --require-tasks --include-tasks

# Compare output with bash script
diff <(bin/sl spec info --json --paths-only) <(.specledger/scripts/bash/check-prerequisites.sh --json --paths-only)
```

### Test: `sl spec create` (replaces create-new-feature.sh)
```bash
# Create a test feature
bin/sl spec create --json --number 999 --short-name "test-feature" "Test feature description"

# Verify: branch created, spec dir exists, spec.md from template
git branch | grep 999-test-feature
ls specledger/999-test-feature/spec.md

# Cleanup
git branch -d 999-test-feature
rm -rf specledger/999-test-feature
```

### Test: `sl spec setup-plan` (replaces setup-plan.sh)
```bash
# Setup plan for current feature
bin/sl spec setup-plan --json

# Compare output with bash script
diff <(bin/sl spec setup-plan --json) <(.specledger/scripts/bash/setup-plan.sh --json)
```

### Test: `sl context update` (replaces update-agent-context.sh)
```bash
# Update Claude context
bin/sl context update claude

# Verify CLAUDE.md was updated
git diff CLAUDE.md

# Update all existing agent files
bin/sl context update
```

---

## Validation Matrix

| Test | Command | Expected |
|------|---------|----------|
| Comment list (human) | `sl comment list` | Formatted comment list grouped by file |
| Comment list (JSON) | `sl comment list --json` | Valid JSON with threads |
| Comment show | `sl comment show <id>` | Full comment with thread |
| Comment reply | `sl comment reply <id> --content "test"` | Reply created, visible in list |
| Comment resolve | `sl comment resolve <id> --reason "done"` | Comment + replies marked resolved |
| Revise still works | `sl revise --summary` | Same output as before refactor |
| Spec info JSON | `sl spec info --json --paths-only` | Matches check-prerequisites.sh output |
| Spec create | `sl spec create --json --number 999 --short-name "t" "test"` | Branch + dir + spec.md |
| Setup plan | `sl spec setup-plan --json` | Matches setup-plan.sh output |
| Context update | `sl context update claude` | CLAUDE.md updated with plan data |
