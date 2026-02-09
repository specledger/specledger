---
name: sl-deps
description: Manage specification dependencies with sl deps commands. Use when adding, listing, removing, or resolving spec dependencies in a SpecLedger project.
---

# sl deps Dependency Management

## Overview

`sl deps` commands manage external specification dependencies stored in `specledger/specledger.yaml`. Dependencies are Git repositories containing specification documents that this project references or depends on.

## Key Concepts

### artifact_path

The `artifact_path` is a fundamental concept in SpecLedger dependencies. It specifies where specification artifacts are stored within a repository.

**Two types of artifact_path:**

1. **Project artifact_path**: Where YOUR project stores its artifacts (configured in your project's specledger.yaml)
   - Example: `artifact_path: specledger/`
   - Default for new projects: `specledger/`

2. **Dependency artifact_path**: Where a DEPENDENCY stores its artifacts (auto-discovered or manually specified)
   - Example: `artifact_path: specs/`
   - Auto-discovered for SpecLedger repos (read from their specledger.yaml)
   - Manually specified via `--artifact-path` flag for non-SpecLedger repos

### Reference Resolution

When you reference an artifact from a dependency, the path is resolved as:

```
<project.artifact_path> + "deps/" + <dependency.alias> + "/" + <artifact_name>
```

**Example:**
```
Project artifact_path: specledger/
Dependency alias: platform
Artifact name: api.md

Resolved path: specledger/deps/platform/api.md
```

### Auto-Download

Dependencies are **automatically downloaded** when you add them:
- Cached to `~/.specledger/cache/<alias>/`
- Current commit SHA is resolved and stored
- No separate `sl deps resolve` command needed (resolve is for manual refresh only)

## When to Use

Use `sl deps` when you need to:
- **Add a dependency** on an external specification
- **List current dependencies** to see what specs are referenced
- **Remove a dependency** that's no longer needed
- **Resolve/checkout dependencies** to refresh them manually
- **View dependency graph** to understand relationships

## Project Context

`sl deps` commands must be run from within a SpecLedger project directory (one containing `specledger/specledger.yaml`).

### Finding Project Root

The `sl deps` commands automatically find the project root by searching for `specledger/specledger.yaml` in parent directories. This means you can run commands from any subdirectory within your project.

Example:
```bash
cd src/components  # Deep in project tree
sl deps list        # Still works - finds project root automatically
```

## Available Commands

### sl deps list

List all dependencies in the current project.

```bash
sl deps list
```

**Output includes:**
- Dependency URL
- Branch name
- Alias (short name for reference)
- Artifact Path (where artifacts are located in the dependency)
- Framework type (SpecKit, OpenSpec, Both, None)
- Current status (resolved with commit SHA)

**When to use:**
- Check what dependencies exist before adding new ones
- Verify a dependency was added successfully
- Review all external specs this project references

### sl deps add

Add a new specification dependency.

```bash
# Full syntax (alias is required)
sl deps add <git-url> [<branch>] --alias <name> [--artifact-path <path>]

# Examples
sl deps add git@github.com:org/platform-specs --alias platform
sl deps add https://github.com:org/api-docs --alias api --artifact-path docs/openapi/
sl deps add git@github.com:user/repo develop --alias user-repo
```

**Parameters:**
- `<git-url>`: Git repository URL (required)
  - SSH: `git@github.com:org/repo.git`
  - HTTPS: `https://github.com/org/repo.git`
- `<branch>`: Branch name (default: `main`)
- `--alias <name>`: Short reference name (**required**)
- `--artifact-path <path>`: Path to artifacts within dependency repo (optional, auto-detected for SpecLedger repos)

**Behavior:**
1. Validates the Git URL
2. Detects framework type (SpecKit, OpenSpec, Both, None)
3. Auto-detects artifact_path for SpecLedger repos
4. Automatically downloads/clones the dependency to cache
5. Resolves and stores current commit SHA
6. Adds dependency to specledger.yaml

**When to use:**
- This project needs to reference another specification
- Building on top of existing spec documents
- Creating a dependency graph between related specs

### sl deps remove

Remove a dependency by URL or alias.

```bash
sl deps remove <url-or-alias>
sl deps remove git@github.com:org/specs.git
sl deps remove api-specs  # Using alias
```

**When to use:**
- Dependency is no longer relevant
- Replacing with a different reference
- Cleaning up unused dependencies

### sl deps resolve

Manually refresh all dependencies (like `go mod download`).

```bash
sl deps resolve
```

**What it does:**
- Refreshes dependencies in `~/.specledger/cache/`
- Updates to latest commits on configured branches
- Resolves commit SHAs for reproducibility
- Stores resolved commits in metadata

**Note:** This is typically only needed after cloning a project or for manual refresh. The `sl deps add` command automatically downloads dependencies.

**When to use:**
- Need offline access to dependencies
- Want to refresh to latest commits
- Preparing for work without internet access

### sl deps update

Check for and apply updates to dependencies.

```bash
sl deps update              # Check all dependencies
sl deps update <alias>      # Update specific dependency
```

**When to use:**
- Want to see if dependencies have new commits
- Updating to latest versions of dependencies
- Preparing for a dependency update cycle

### sl deps graph

Display dependency relationships as a graph.

```bash
sl deps graph
```

**Output:**
- Visual representation of dependency tree
- Shows which specs depend on others
- Helps understand impact of changes

## Dependency Storage

Dependencies are stored in `specledger/specledger.yaml`:

```yaml
artifact_path: specledger/  # Where this project stores its artifacts
dependencies:
  - url: git@github.com:org/platform-specs
    branch: main
    alias: platform
    artifact_path: specs/    # Where the dependency stores its artifacts (auto-detected)
    resolved_commit: abc123...
    framework: both
    import_path: @platform/spec
```

## Common Patterns

### Pattern 1: Adding SpecLedger Repository Dependencies

When adding a SpecLedger repository, artifact_path is auto-detected:

```bash
sl deps add git@github.com:org/platform-specs --alias platform
```

The system will:
1. Detect SpecLedger framework
2. Clone the repo temporarily
3. Read artifact_path from its specledger.yaml
4. Store the detected artifact_path

### Pattern 2: Adding Non-SpecLedger Repository Dependencies

For non-SpecLedger repos, manually specify artifact_path:

```bash
sl deps add https://github.com/org/api-docs --alias api --artifact-path docs/openapi/
```

### Pattern 3: Cross-Project References

When multiple projects reference shared specs:

```bash
# In project-a
sl deps add git@github.com:org/shared-specs --alias shared

# In project-b
sl deps add git@github.com:org/shared-specs --alias shared
```

Both projects now reference the same source of truth.

### Pattern 4: Hierarchical Specs

When building on top of platform specs:

```bash
# Platform team creates core spec
# Service teams add dependency on platform spec
sl deps add git@github.com:org/platform-specs --alias platform
```

## Session Start Checklist

When starting a session in a SpecLedger project:

```
Deps Check:
- [ ] Run sl deps list to see current dependencies
- [ ] Check if dependencies are resolved (committed SHAs present)
- [ ] Ask user if they want to add any new dependencies
- [ ] Report dependency context: "This project has X dependencies: [summary]"
```

## Error Handling

**Not in a SpecLedger project:**
```
Error: failed to find project root: not in a SpecLedger project (no specledger/specledger.yaml found)
```
Solution: Navigate to your SpecLedger project directory first.

**Missing --alias flag:**
```
Error: required flag(s) "alias" not set
```
Solution: Add the `--alias` flag when adding dependencies.

**Invalid artifact-path:**
```
Error: invalid artifact-path: must be a relative path
```
Solution: Use a relative path without `../` or leading `/`.

**Dependency already exists:**
```
Error: dependency already exists: git@github.com:org/specs
```
Solution: Use `sl deps remove` first if replacing, or check if you meant a different dependency.

**Failed to clone repository:**
```
Warning: Failed to clone git@github.com:org/private-specs: authentication failed
```
Solution:
- Set up SSH keys for private repositories
- Or use HTTPS with a personal access token
- Verify you have access to the repository

## Best Practices

1. **Use meaningful aliases**: `sl deps add git@github.com:org/platform-auth --alias auth` (not `--alias specs`)
2. **Pin branches for production**: `sl deps add git@github.com:org/specs v1.0 --alias prod-specs`
3. **Commit resolved commits**: After `sl deps resolve`, commit the updated specledger.yaml
4. **Run resolve before offline work**: `sl deps resolve` ensures all dependencies are cached locally
5. **Use reference format**: Instead of hardcoding paths, use `alias:artifact` format in documentation

## Integration with Claude Code

The `sl deps` commands integrate with Claude Code to help AI assistants understand your project's dependency structure. When you add dependencies:

1. Claude can read specifications from dependencies
2. Cross-references between specs are automatically resolved
3. AI assistants can trace API contracts and requirements across projects

## Cache Location

Dependencies are cached globally at:
```
~/.specledger/cache/<alias>/
```

This cache is shared across all SpecLedger projects on your machine, making dependency management efficient.
