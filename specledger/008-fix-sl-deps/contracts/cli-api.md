# CLI API Contracts: SpecLedger Dependencies

**Feature**: 008-fix-sl-deps
**Date**: 2026-02-09
**Phase**: Phase 1 - Design & Contracts

## Overview

This document defines the CLI API contracts for the SpecLedger dependencies commands. These contracts specify the command interface, flags, arguments, input/output formats, and error handling.

---

## Command: sl deps add

### Signature
```bash
sl deps add <repo-url> [<branch>] [<path>] [flags]
```

### Arguments
| Position | Name | Type | Required | Default | Description |
|----------|------|------|----------|---------|-------------|
| 0 | repo-url | string | Yes | - | Git repository URL (SSH or HTTPS) |
| 1 | branch | string | No | `main` | Git branch name |
| 2 | path | string | No | `<alias>` | Reference path within project's artifact_path |

### Flags
| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `-a, --alias <name>` | string | No | `<generated>` | Short name for the dependency |
| `--artifact-path <path>` | string | Conditional | - | Artifact path in dependency repo (required for non-SpecLedger repos) |

### Input
None (arguments and flags only)

### Output Format
```
Detecting Framework
Checking git@github.com:org/specs...
  Framework:  Spec Kit

Dependency added
  Repository:  git@github.com:org/specs
  Alias:       specs
  Branch:      main
  Path:        spec.md
  Framework:   Spec Kit
  Import Path: @specs/spec

Next: sl deps resolve
```

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Invalid URL format |
| 2 | Dependency already exists |
| 3 | Alias already exists |
| 4 | Failed to save metadata |

### Error Conditions
| Condition | Error Message | Resolution |
|-----------|---------------|------------|
| Invalid repo URL | `invalid repository URL: <url>` | Use valid SSH or HTTPS Git URL |
| Dependency exists | `dependency already exists: <url>` | Remove first or use different URL |
| Alias exists | `alias already exists: <alias>` | Use different alias |
| Not in project | `failed to find project root: not in a SpecLedger project` | Navigate to project directory |

### Examples
```bash
# Add SpecLedger repo (auto-detect artifact_path)
sl deps add git@github.com:org/platform-specs main specs/ --alias platform

# Add non-SpecLedger repo (manual artifact_path)
sl deps add https://github.com/external/api.git --artifact-path docs/openapi/

# Add with all parameters
sl deps add git@github.com:user/repo.git develop docs/spec.md --alias user-spec
```

---

## Command: sl deps list

### Signature
```bash
sl deps list
```

### Arguments
None

### Flags
None

### Input
None

### Output Format
```
Dependencies (3 total)

1. git@github.com:org/platform-specs
   Branch:  develop
   Path:    platform
   Alias:   platform
   Framework: Both
   Import:  @platform/spec
   Status:  ✓ abc123de

2. https://github.com/external/api.git
   Alias:   api-docs
   Status:  not resolved (run sl deps resolve)

3. git@github.com:org/shared
   Path:    shared
   Status:  not resolved (run sl deps resolve)
```

**When no dependencies exist**:
```
Dependencies

No dependencies declared.

Add dependencies with:
  sl deps add git@github.com:org/spec
```

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success |

### Examples
```bash
sl deps list
```

---

## Command: sl deps remove

### Signature
```bash
sl deps remove <url-or-alias>
```

### Arguments
| Position | Name | Type | Required | Description |
|----------|------|------|----------|-------------|
| 0 | url-or-alias | string | Yes | Dependency URL or alias |

### Flags
None

### Input
None

### Output Format
```
Dependency removed
  git@github.com:org/specs
```

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Not in project |
| 2 | Dependency not found |
| 3 | Failed to save metadata |

### Error Conditions
| Condition | Error Message | Resolution |
|-----------|---------------|------------|
| Not found | `dependency not found: <url-or-alias>` | Check with `sl deps list` first |

### Examples
```bash
# Remove by URL
sl deps remove git@github.com:org/specs

# Remove by alias
sl deps remove platform
```

---

## Command: sl deps resolve

### Signature
```bash
sl deps resolve [flags]
```

### Arguments
None

### Flags
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-n, --no-cache` | bool | `false` | Ignore cached dependencies, force re-download |

### Input
None

### Output Format
```
Resolving Dependencies
Resolving 3 dependencies...

1. git@github.com:org/platform-specs
   Alias:  platform
   Branch: develop
   Cache:  /Users/user/.specledger/cache/platform-specs
   Status: ✓ abc123def4

2. https://github.com/external/api.git
   Cache:  /Users/user/.specledger/cache/api-docs
   Status: cloning...
   Status: ✓ def456abc1

3. git@github.com:org/shared
   Cache:  /Users/user/.specledger/cache/shared
   Status: cloning...
   Warning: Failed to clone git@github.com:org/shared: network timeout

Resolved 2/3 dependencies
Warning: Some dependencies failed to resolve
```

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success (all or some resolved) |
| 1 | Not in project |
| 2 | No dependencies to resolve |

### Behavior

1. For each dependency:
   - Check if already cached and `--no-cache` not set
   - Clone to `~/.specledger/cache/<dir-name>/`
   - Checkout specified branch
   - Resolve current commit SHA
   - Store commit in metadata

2. Auto-discovery for SpecLedger repos:
   - Read `specledger.yaml` from cloned dependency
   - Extract `artifact_path` value
   - Update dependency's artifact_path field

3. Cache directory naming:
   - Use alias if available
   - Otherwise generate from URL (replace `:`, `/`, `.git` with `-`)

### Error Conditions
| Condition | Behavior |
|-----------|----------|
| Network timeout | Warning, continue with next dependency |
| Authentication failed | Error, stop resolving |
| Invalid branch | Warning, use default branch |
| Disk full | Error, stop resolving |

### Examples
```bash
# Resolve all dependencies
sl deps resolve

# Force re-download (ignore cache)
sl deps resolve --no-cache
```

---

## Command: sl deps update

### Signature
```bash
sl deps update [url-or-alias] [flags]
```

### Arguments
| Position | Name | Type | Required | Description |
|----------|------|------|----------|-------------|
| 0 | url-or-alias | string | No | Specific dependency to update (default: all) |

### Flags
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-y, --yes` | bool | `false` | Auto-confirm all updates |

### Input
None

### Output Format
```
Checking for Updates
Checking 3 dependencies for updates...

1. platform-specs (git@github.com:org/platform-specs)
   Current: abc123def4 (main)
   Latest:  def456abc1 (main)
   Update?  [y/N] y
   Updated to def456abc1

2. api-docs (https://github.com/external/api.git)
   Current: def456abc1
   Latest:  def456abc1
   No update available

3. shared (git@github.com:org/shared)
   Current: (not resolved)
   Latest:  1234567890 (main)
   Update?  [y/N] y
   Updated to 1234567890

Updated 2 dependencies
```

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Not in project |
| 2 | No dependencies |

### Behavior

1. Fetch latest from remote for each dependency
2. Compare with current resolved commit
3. Prompt for confirmation (unless `--yes` flag)
4. Update and re-resolve commit SHA

### Examples
```bash
# Check and update all
sl deps update

# Update specific dependency
sl deps update platform

# Update all without prompting
sl deps update --yes
```

---

## Command: sl init (Modified)

### Changes
Add `artifact_path` detection and configuration.

### Signature
```bash
sl init [flags]
```

### Flags (Existing + New)
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--artifact-path <path>` | string | auto-detect | Path to artifacts directory |

### Behavior Changes

1. **Detect artifact_path**: Scan for common directories
   - `specledger/`
   - `specs/`
   - `docs/specs/`
   - `documentation/`

2. **Prompt user**: Confirm detected path or enter custom path

3. **Write to specledger.yaml**: Store confirmed artifact_path

### Example Interaction
```
Detecting project structure...
Found specledger/ directory
Use 'specledger/' as artifact path? [Y/n]: y

Initialized SpecLedger project
  Artifact path: specledger/
```

---

## Command: sl new (Modified)

### Changes
Always set `artifact_path: specledger/` for new projects.

### Behavior

1. Create project structure
2. Set `artifact_path: specledger/` in metadata
3. No user interaction required

---

## Global Error Format

All commands use consistent error formatting:

```
Error: <error message>

Run 'sl deps <command> --help' for usage information.
```

For warnings:
```
⚠ Warning: <warning message>
```

---

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `SPECLEDGER_CACHE_DIR` | Override default cache location (`~/.specledger/cache`) |
| `GIT_SSH_COMMAND` | Custom SSH command for private repos |
| `NO_COLOR` | Disable colored output |

---

## Configuration Files

### specledger.yaml Structure
```yaml
version: "1.0.0"
project:
  name: string
  short_code: string
  created: timestamp
  modified: timestamp
  version: string
artifact_path: string              # NEW
playbook:
  name: string
  version: string
  applied_at: timestamp (optional)
  structure: []string (optional)
task_tracker:                      # optional
  choice: "beads" | "none"
  enabled_at: timestamp (optional)
dependencies: []Dependency
```

### Dependency Structure
```yaml
url: string
branch: string (optional)
path: string (optional)
alias: string (optional)
artifact_path: string (optional)   # NEW
resolved_commit: string (optional)
framework: "speckit" | "openspec" | "both" | "none" (optional)
import_path: string (optional)
```

---

## Backward Compatibility

### Loading Old Files

Files without `artifact_path`:
- Project: Default to `"specledger/"`
- Dependencies: Auto-discover or leave empty

### Saving

- Only write `artifact_path` if non-empty
- Use `omitempty` YAML tag

---

## Testing Contract

### Unit Tests Must Cover

1. Argument parsing (position, flags)
2. Default value application
3. Error code returns
4. Output format validation

### Integration Tests Must Cover

1. Full command execution
2. File system effects (cache, metadata)
3. Git operations (clone, checkout)
4. Network error handling

---

**Next**: Generate quickstart documentation.
