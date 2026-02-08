---
name: sl-deps
description: Manage specification dependencies with sl deps commands. Use when adding, listing, removing, or resolving spec dependencies in a SpecLedger project.
---

# sl deps Dependency Management

## Overview

`sl deps` commands manage external specification dependencies stored in `specledger/specledger.yaml`. Dependencies are Git repositories containing specification documents that this project references or depends on.

## When to Use

Use `sl deps` when you need to:
- **Add a dependency** on an external specification
- **List current dependencies** to see what specs are referenced
- **Remove a dependency** that's no longer needed
- **Resolve/checkout dependencies** to work with them locally
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
- File path (within the dependency)
- Alias (short name for reference)
- Current status (resolved/unresolved)

**When to use:**
- Check what dependencies exist before adding new ones
- Verify a dependency was added successfully
- Review all external specs this project references

### sl deps add

Add a new specification dependency.

```bash
# Full syntax
sl deps add <git-url> [<branch>] [<path>] [--alias <name>]

# Examples
sl deps add git@github.com:org/specs main specs/api.md
sl deps add https://github.com/org/specs.git --alias api-specs
sl deps add git@github.com:user/repo main docs/spec.md --alias user-repo-spec
```

**Parameters:**
- `<git-url>`: Git repository URL (required)
  - SSH: `git@github.com:org/repo.git`
  - HTTPS: `https://github.com/org/repo.git`
- `<branch>`: Branch name (default: `main`)
- `<path>`: Path to spec file within repo (default: root)
- `--alias <name>`: Short reference name (optional)

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

Checkout/update dependencies locally for offline access.

```bash
sl deps resolve
```

**What it does:**
- Clones dependencies to `specledger/deps/` directory
- Resolves commit SHAs for reproducibility
- Stores resolved commit in metadata

**When to use:**
- Need offline access to dependencies
- Want to pin specific commits for reproducibility
- Preparing for work without internet access

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
dependencies:
  - url: git@github.com:org/specs.git
    branch: main
    path: spec.md
    alias: org-spec
    resolved_commit: abc123...
```

## Common Patterns

### Pattern 1: Adding API Dependencies

When your service depends on an API specification:

```bash
sl deps add git@github.com:api-team/api-specs main openapi.yaml --alias api-spec
```

This allows Claude to read the API spec when implementing your service.

### Pattern 2: Cross-Project References

When multiple projects reference shared specs:

```bash
# In project-a
sl deps add git@github.com:org/shared-specs main common.yaml --alias shared

# In project-b
sl deps add git@github.com:org/shared-specs main common.yaml --alias shared
```

Both projects now reference the same source of truth.

### Pattern 3: Hierarchical Specs

When building on top of platform specs:

```bash
# Platform team creates core spec
# Service teams add dependency on platform spec
sl deps add git@github.com:org/platform-specs main core.yaml --alias platform
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

**Dependency already exists:**
```
Error: dependency already exists
```
Solution: Use `sl deps remove` first, or check if you meant to add a different dependency.

**Invalid Git URL:**
```
Error: invalid git URL format
```
Solution: Use proper SSH or HTTPS Git URL format.

## Integration with Other Commands

The `sl deps` commands work with:
- `sl doctor` - Shows dependency status
- `sl new` - Initializes empty dependency list
- `sl init` - Initializes empty dependency list

Claude can read dependencies using the metadata when:
- Implementing features that reference external specs
- Checking compatibility with dependent specs
- Understanding project context and relationships

## Troubleshooting

**Dependencies not resolving:**
- Check Git URL is accessible
- Verify you have authentication for private repos
- Check network connectivity

**Dependencies disappeared:**
- Check `specledger/specledger.yaml` still exists
- Verify YAML format is correct
- Run `sl deps list` to confirm current state

**Need to update dependency:**
- Remove and re-add with new parameters
- Or edit `specledger/specledger.yaml` directly
- Run `sl deps resolve` to update commit SHAs
