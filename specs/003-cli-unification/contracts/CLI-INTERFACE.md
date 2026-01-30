# CLI Interface Contracts

**Feature**: CLI Unification (003-cli-unification)
**Date**: 2026-01-30

## Overview

This document defines the interface contracts for the unified CLI tool. These contracts define the expected behavior, flags, arguments, and output formats for all CLI commands.

## Command Structure

### Global Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--help` | `-h` | bool | false | Show help and exit |
| `--version` | `-v` | bool | false | Show version and exit |
| `--ci` | | bool | false | Force non-interactive mode (skip TUI) |
| `--simple` | | bool | false | Force plain CLI mode (no TUI) |
| `--dry-run` | | bool | false | Preview changes without executing |

### Version Output Format

```
sl version 1.0.0
Built: 2026-01-30T10:30:45Z
Commit: 55a256b
Platform: darwin/arm64
```

### Help Output Format

```
SpecLedger - Specification dependency management

USAGE:
    sl [COMMAND] [FLAGS] [ARGUMENTS]

COMMANDS:
    new, bootstrap       Start interactive TUI for project bootstrap
    deps [SUBCOMMAND]    Manage specification dependencies
    refs [SUBCOMMAND]    Validate and manage references
    graph [SUBCOMMAND]   Visualize dependency graphs
    vendor [SUBCOMMAND]  Manage vendor specifications
    conflict [SUBCOMMAND] Resolve dependency conflicts
    update [SUBCOMMAND]  Update dependencies

FLAGS:
    -h, --help       Show help and exit
    -v, --version    Show version and exit
    --ci             Force non-interactive mode
    --simple         Force plain CLI mode
    --dry-run        Preview changes without executing

EXAMPLES:
    sl new                    Start interactive project bootstrap
    sl new --project-name myproj --short-code myp
    sl deps list
    sl deps add git@github.com:org/spec.git
```

---

## Command: bootstrap / new

### Interactive Mode (TUI)

**When**: No `--ci` flag, interactive terminal detected

**Invocations**:
- `sl new`
- `sl bootstrap`

**Arguments**: None

**Output**: Interactive TUI prompts for:
1. Project name
2. Short code (2-4 letters)
3. Playbook selection (Default, Data Science, Platform Engineering)
4. Agent shell (Claude Code, Gemini CLI, Codex)

**Exit Codes**:
- 0: Success
- 1: Error

---

### Non-Interactive Mode (Flags)

**When**: `--ci` flag provided or non-interactive environment detected

**Invocations**:
- `sl new --project-name <name> --short-code <code> [--playbook <type>] [--shell <shell>]`
- `sl bootstrap --project-name <name> --short-code <code>`

**Arguments**:
| Argument | Required | Type | Description |
|----------|----------|------|-------------|
| `--project-name` | Yes | string | Project name (alphanumeric, hyphens, underscores) |
| `--short-code` | Yes | string | Short code (2-4 lowercase letters) |
| `--playbook` | No | string | Playbook type (default, data-science, platform-engineering) |
| `--shell` | No | string | Agent shell (claude-code, gemini-cli, codex) |

**Output**: Plain text output to stdout

**Exit Codes**:
- 0: Success
- 1: Error

**Error Scenarios**:
1. Missing required arguments → Error message with usage
2. Invalid project name format → Error with validation message
3. Invalid short code format → Error with validation message
4. Directory already exists → Error with location
5. Permission denied → Error with alternative location suggestion

**Example**:
```
$ sl new --project-name myproject --short-code myp --playbook default

✓ Project created: ~/demos/myproject
✓ Beads prefix: myp
✓ Playbook: Default (General SWE)
✓ Agent Shell: Claude Code

Next steps:
  cd ~/demos/myproject
  claude
```

---

## Command: deps

### Subcommand: deps list

**When**: Called from within a SpecLedger project

**Invocations**:
- `sl deps list`
- `sl deps list --include-transitive`
- `sl deps list --format table`

**Arguments**: None

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--include-transitive` | bool | Include transitive dependencies |
| `--format` | string | Output format (table, json) |

**Output**:

**Table Format** (default):
```
REPOSITORY                    BRANCH        SPEC PATH            ALIAS    STATUS
────────────────────────────────────────────────────────────────────────────────
git@github.com:org/a.git      main          specs/a.md           a        active
git@github.com:org/b.git      develop       specs/b.md           b        active
git@github.com:org/c.git      feature       specs/c.md           c        locked
```

**JSON Format** (`--format json`):
```json
{
  "dependencies": [
    {
      "repository": "git@github.com:org/a.git",
      "branch": "main",
      "spec_path": "specs/a.md",
      "alias": "a",
      "status": "active",
      "added_at": "2026-01-30T10:30:00Z"
    }
  ],
  "transitive_count": 2
}
```

**Exit Codes**:
- 0: Success
- 1: Not a SpecLedger project
- 2: No dependencies

---

### Subcommand: deps add

**When**: Called from within a SpecLedger project

**Invocations**:
- `sl deps add <repo-url> [branch] [spec-path] [--alias]`

**Arguments**:
| Argument | Required | Type | Description |
|----------|----------|------|-------------|
| `repo-url` | Yes | string | Git repository URL |
| `branch` | No | string | Git branch (default: main) |
| `spec-path` | No | string | Path to spec file (default: specs/sdd-control-plane.md) |

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--alias` | string | Local alias for referencing |
| `--dry-run` | bool | Preview without adding |

**Output**:
```
✓ Dependency added:
  Repository: git@github.com:org/spec.git
  Branch: main
  Spec path: specs/sdd-control-plane.md
  Alias: spec
  Status: active
```

**Exit Codes**:
- 0: Success
- 1: Not a SpecLedger project
- 2: Invalid repository URL
- 3: Already exists
- 4: Signature verification failed
- 5: Network error

---

### Subcommand: deps resolve

**When**: Called from within a SpecLedger project

**Invocations**:
- `sl deps resolve [--no-cache] [--deep]`

**Arguments**: None

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--no-cache` | bool | Skip cache, re-resolve all |
| `--deep` | bool | Include transitive dependencies |

**Output**:
```
Resolving dependencies...
  a → b (transitive)
  c → d (transitive)
  ✓ All dependencies verified

Lockfile updated: .specledger/lockfile.yaml
Verification status: all-clean
```

**Exit Codes**:
- 0: Success
- 1: Not a SpecLedger project
- 2: Dependency verification failed

---

### Subcommand: deps update

**When**: Called from within a SpecLedger project

**Invocations**:
- `sl deps update [repo-url] [--force]`

**Arguments**:
| Argument | Required | Type | Description |
|----------|----------|------|-------------|
| `repo-url` | No | string | Specific repository to update |

**Flags**:
| Flag | Type | Description |
|------|------|-------------|
| `--force` | bool | Force update even if unchanged |

**Output**:
```
Updating dependencies...
  git@github.com:org/a.git: main → develop
  ✓ Updated: 1 dependency

Lockfile updated: .specledger/lockfile.yaml
```

**Exit Codes**:
- 0: Success
- 1: Not a SpecLedger project
- 2: Update failed

---

### Subcommand: deps remove

**When**: Called from within a SpecLedger project

**Invocations**:
- `sl deps remove <repo-url> <spec-path>`

**Arguments**:
| Argument | Required | Type | Description |
|----------|----------|------|-------------|
| `repo-url` | Yes | string | Repository URL |
| `spec-path` | Yes | string | Spec file path |

**Flags**: None

**Output**:
```
✓ Dependency removed: git@github.com:org/a.git (specs/a.md)
Lockfile updated: .specledger/lockfile.yaml
```

**Exit Codes**:
- 0: Success
- 1: Not a SpecLedger project
- 2: Dependency not found
- 3: Cannot remove (has dependencies)

---

## Command: refs

### Subcommand: refs validate

**Invocations**:
- `sl refs validate`
- `sl refs validate --all`

**Output**:
```
Validating references...
  ✓ src/main.go references sdd-control-plane spec
  ✓ src/utils.go references database spec
  ✓ No invalid references found
```

**Exit Codes**:
- 0: Success
- 1: Invalid references found

---

## Command: graph

### Subcommand: graph show

**Invocations**:
- `sl graph show`
- `sl graph show --format dot`

**Output**:

**Text Format** (default):
```
Project dependencies:
  a → b → c
    └── d

Nodes:
  a (active)
  b (active)
  c (active)
  d (active)
```

**DOT Format** (`--format dot`):
```
digraph dependencies {
  a -> b;
  a -> d;
  b -> c;
  d [style=dashed];
}
```

**Exit Codes**:
- 0: Success
- 1: No dependencies to show

---

## Command: vendor

### Subcommand: vendor add

**Invocations**:
- `sl vendor add <spec-url>`

**Output**:
```
✓ Spec vendored to vendor/specs/spec.md
  Checksum: abc123...
```

---

### Subcommand: vendor list

**Invocations**:
- `sl vendor list`

**Output**:
```
Vendored specs:
  vendor/specs/sdd-control-plane.md (verified)
  vendor/specs/database-spec.md (verified)
```

---

## Command: conflict

### Subcommand: conflict list

**Invocations**:
- `sl conflict list`

**Output**:
```
Potential conflicts:
  1. a → b (version mismatch)
  2. c → d (duplicate dependency)
```

---

### Subcommand: conflict resolve

**Invocations**:
- `sl conflict resolve <conflict-id>`

**Output**:
```
✓ Resolved: a → b (used b's spec)
```

---

## Command: update

### Subcommand: update self

**Invocations**:
- `sl update self`

**Output**:
```
Checking for updates...
  Latest version: 1.0.1
  Current version: 1.0.0
  ✓ Available: 1.0.1

Download: https://github.com/org/cli/releases/latest
```

---

## Error Messages

### General Error Format

```
ERROR: [error description]

Usage: sl [COMMAND] [FLAGS] [ARGUMENTS]
```

### Specific Errors

**Invalid Command**:
```
ERROR: Unknown command "invalid"

Available commands:
  new, bootstrap       Start interactive project bootstrap
  deps [SUBCOMMAND]    Manage specification dependencies
  ...
```

**Missing Dependency**:
```
ERROR: gum not found

The TUI requires gum. Install from:
  macOS:  brew install gum
  Linux:  go install github.com/charmbracelet/gum@latest

You can also continue without TUI using --ci flag.
```

**Permission Denied**:
```
ERROR: Permission denied: ~/demos/myproject

Cannot write to this directory. Please choose:
  1. A directory where you have write permissions
  2. Current directory: ./myproject
  3. Provide full path with --project-dir flag
```

**Invalid Input**:
```
ERROR: Invalid project name: my-project

Project name must be alphanumeric, hyphens, or underscores only.
```

**Project Exists**:
```
ERROR: ~/demos/myproject already exists

Bootstrap command requires a new directory. Options:
  1. Choose a different project name
  2. Use existing project: cd ~/demos/myproject
  3. Remove existing directory: rm -rf ~/demos/myproject
```

**Not a SpecLedger Project**:
```
ERROR: Not a SpecLedger project

This command requires running from within a SpecLedger project.

Expected to find .specledger directory.
```

---

## Exit Code Standards

| Exit Code | Scenario |
|-----------|----------|
| 0 | Success |
| 1 | Any error |
| 2 | Command not found or invalid arguments |
| 130 | User cancellation (Ctrl+C) |

---

## Logging Format

### Debug Logging

```
2026/01/30 10:30:45 [DEBUG] Starting bootstrap with project name "myproject"
2026/01/30 10:30:45 [DEBUG] Checking for gum dependency... not found
2026/01/30 10:30:45 [DEBUG] Asking user about dependency installation
2026/01/30 10:30:50 [INFO] Installing gum...
2026/01/30 10:30:52 [DEBUG] gum installed successfully
2026/01/30 10:30:52 [DEBUG] Project created at ~/demos/myproject
```

### Error Logging

```
2026/01/30 10:30:45 [ERROR] Failed to create project directory: permission denied
2026/01/30 10:30:45 [ERROR] Aborting bootstrap due to error
```

---

## Summary

This document defines the complete interface contracts for the unified CLI tool:

1. **Global flags** for version, help, and mode selection
2. **bootstrap/new command** with both TUI and non-interactive modes
3. **deps subcommands** for all dependency management operations
4. **Other command groups** (refs, graph, vendor, conflict, update)
5. **Error handling** with standard formats and exit codes
6. **Logging** for debugging

All commands follow consistent patterns for arguments, flags, output formats, and error handling.
