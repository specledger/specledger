# Quickstart: Shared Project Root Resolution

**Date**: 2026-03-23
**Feature**: 610-shared-project-root

## Scenario 1: sl doctor --template from subdirectory (Issue #81 fix)

```bash
# Navigate to a subdirectory within a SpecLedger project
cd /path/to/my-project/pkg/cli/commands

# Run doctor --template — should succeed, not error
sl doctor --template
# Expected: Templates updated successfully (or "up to date")
```

## Scenario 2: sl doctor from project root (regression check)

```bash
# Navigate to the project root
cd /path/to/my-project

# Run doctor — should work exactly as before
sl doctor
# Expected: Full doctor output with CLI version and template status
```

## Scenario 3: sl doctor from outside any project (error handling)

```bash
# Navigate to a directory with no SpecLedger project above it
cd /tmp

# Run doctor --template — should fail with a clear message
sl doctor --template
# Expected error: "not in a SpecLedger project (no specledger.yaml found).
#   Run 'sl init' to create one, or navigate to a project directory."
```

## Scenario 4: Other commands from subdirectory (consistency)

```bash
# Navigate to a subdirectory
cd /path/to/my-project/pkg/cli/commands

# Run various project-aware commands
sl session list    # Should work
sl comment list    # Should work
sl deps list       # Already works (has its own findProjectRoot)
```

## Scenario 5: sl doctor --json from subdirectory

```bash
cd /path/to/my-project/some/deep/subdir

sl doctor --json
# Expected: Valid JSON output with template_version and template_update_available fields
```
