---
name: deps-manager
description: Manage specification dependencies using the SpecLedger CLI deps commands. Use when working with external specification dependencies, dependency resolution, or linking specs across repositories.
---

# Deps Manager

## Overview

The SpecLedger CLI provides a comprehensive set of commands to manage external specification dependencies. Dependencies are specified in `specs/spec.mod` and can be resolved, validated, and locked for reproducible builds.

## When to Use

Use this skill when:
- **Adding new dependencies** - Referencing specs from other repositories
- **Listing dependencies** - Understanding what specs a project depends on
- **Resolving dependencies** - Fetching and validating external specs
- **Updating dependencies** - Keeping specs at the latest compatible version
- **Removing dependencies** - Cleaning up unused spec references
- **Transitive dependencies** - Understanding dependency trees

## CLI Commands Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `sl deps add <repo-url>` | Add a specification dependency |
| `sl deps list` | List all declared dependencies |
| `sl deps resolve` | Resolve all dependencies and generate spec.sum |
| `sl deps update` | Update dependencies to latest compatible versions |
| `sl deps remove <repo-url> <spec-path>` | Remove a dependency |

### Adding Dependencies

**Basic usage:**
```bash
sl deps add github.com/org/project-spec
```

**With branch and spec path:**
```bash
sl deps add github.com/org/project-spec main specs/api.md
```

**With alias:**
```bash
sl deps add github.com/org/project-spec --alias myapi
```

**Usage in code/specs:**
```markdown
## API Specification

This spec depends on: `myapi` (github.com/org/project-spec/main/specs/api.md)
```

**Variables in spec templates:**
```markdown
## API Response Schema

```yaml
{{ $myapi := (lookup "github.com/org/project-spec" "main" "specs/api.yml" "ResponseSchema") }}
{{ $myapi }}
```

## Listing Dependencies

```bash
sl deps list
```

Output:
```
Dependencies (2):

1. github.com/org/project-spec
   Version: main
   Spec: specs/api.md
   Alias: myapi

2. github.com/org/referenced-spec
   Version: v1.0
   Spec: specs/reference.md
```

### Resolving Dependencies

```bash
sl deps resolve
```

This command:
1. Reads `specs/spec.mod`
2. Fetches external specifications from Git
3. Validates versions and schemas
4. Generates `specs/spec.sum` with cryptographic hashes
5. Provides lockfile with commit hashes for reproducibility

**Skip cache:**
```bash
sl deps resolve --no-cache
```

**Verbose output:**
```bash
sl deps resolve --no-cache
```

### Updating Dependencies

```bash
# Update all dependencies to latest compatible versions
sl deps update

# Force update a specific dependency
sl deps update --force github.com/org/project-spec
```

### Removing Dependencies

```bash
sl deps remove github.com/org/project-spec specs/api.md
```

## Dependency Workflow

### Typical Project Setup

```bash
# 1. Initialize a new project
sl new

# 2. Add your first dependency
sl deps add github.com/org/your-org/spec-base main specs/base.md --alias base

# 3. Resolve to fetch and validate
sl deps resolve

# 4. Add more dependencies as needed
sl deps add github.com/org/your-org/ui-specs main specs/ui.md

# 5. Resolve again
sl deps resolve

# 6. Verify your lockfile
cat specs/spec.sum
```

### Working with Dependencies in Specs

When you add a dependency with an alias, you can reference it in your specs:

**In markdown specs:**
```markdown
# User Authentication Spec

## Dependencies
- `base` (github.com/org/spec-base/main/specs/auth-base.md)

## Content
{{ $base := (lookup "github.com/org/spec-base" "main" "specs/auth-base.md") }}
{{ $base }}

## Extending Base
We extend the base auth spec with custom flows...
```

**Variables from dependencies:**
```yaml
# project-spec.md
api_endpoint: {{ (lookup "github.com/org/api-spec" "main" "specs/endpoints.yml" "api_endpoint") }}

rate_limit: {{ (lookup "github.com/org/api-spec" "main" "specs/endpoints.yml" "rate_limit") }}
```

## Dependency Format

**spec.mod file structure:**
```yaml
manifest_version: 1
dependencies:
  - repository_url: github.com/org/project-spec
    version: main
    spec_path: specs/project.md
    alias: myproject
    added_at: 2024-01-15T10:00:00Z
```

**spec.sum file structure:**
```json
{
  "lockfile_version": "1",
  "dependencies": [
    {
      "repository_url": "github.com/org/project-spec",
      "commit_hash": "abc123def456",
      "content_hash": "xyz789",
      "spec_path": "specs/project.md",
      "branch": "main",
      "size": 1024,
      "fetched_at": "2024-01-15T10:00:00Z"
    }
  ],
  "total_size": 1024,
  "generated_at": "2024-01-15T10:00:00Z"
}
```

## Advanced Patterns

### Multiple Aliases for Same Repo

```bash
sl deps add github.com/org/shared main specs/common.md --alias common
sl deps add github.com/org/shared main specs/ui.md --alias ui
```

### Conditional Dependencies

```markdown
# Conditional spec loading

{{ if (hasDependency "github.com/org/core-spec") }}
{{ $core := (lookup "github.com/org/core-spec" "main" "specs/core.md") }}
{{ $core }}
{{ end }}
```

### Dependency Version Locking

```bash
# Add with specific version
sl deps add github.com/org/project-spec v1.2.3

# This locks to the commit at v1.2.3
# To update, run: sl deps update
```

## Troubleshooting

### "Not a SpecLedger project"

Make sure you're in a project directory with `.specledger/`:
```bash
ls .specledger/
sl deps list
```

### "Dependency not found"

```bash
# Check if added correctly
sl deps list

# Check the spec.mod file directly
cat specs/spec.mod
```

### "Failed to resolve dependencies"

```bash
# Try with --no-cache to re-fetch
sl deps resolve --no-cache

# Check internet connection and repo URLs
# Verify the repo exists at the specified path
```

### "Invalid repository URL"

Use valid Git URLs:
- ✅ `github.com/org/project`
- ✅ `https://github.com/org/project.git`
- ✅ `git@github.com:org/project.git`

## Integration with Other Skills

### With bd-issue-tracking

When discovering issues during dependency work:

```bash
# Create issue for discovered problem
bd create "Found: Dependency URL is incorrect"

# Link issues to track progress
bd dep add current-task-id dependency-issue-id --type discovered-from
```

### With TUI (gum)

For interactive project setup with dependencies:

```bash
# Use gum-enhanced prompts
sl new

# This uses gum input/confirm/choose for:
# - Project name
# - Short code
# - Playbook selection
# - Agent shell selection
```

## Best Practices

1. **Use aliases** for common dependencies to make specs more readable
2. **Lock versions** when you need reproducibility
3. **Resolve regularly** to keep `spec.sum` up to date
4. **Document dependencies** in your README
5. **Use transitive flags** when needed to understand the full dependency tree
6. **Resolve with --no-cache** when pushing changes to ensure fresh specs

## Files Reference

| File | Description |
|------|-------------|
| `specs/spec.mod` | Dependency manifest (add/remove/update deps here) |
| `specs/spec.sum` | Lockfile with resolved commits and hashes |
| `.specledger/` | Project database directory |

## Related Skills

- **[bd-issue-tracking](.claude/skills/bd-issue-tracking/SKILL.md)** - Use for tracking work across sessions with complex dependencies
- **[sl-bootstrap](.claude/skills/sl-bootstrap/SKILL.md)** - Interactive project setup with TUI

## Quick Commands Cheat Sheet

```bash
# Add dependency
sl deps add <repo-url> [branch] [spec-path] [--alias <name>]

# List dependencies
sl deps list [--include-transitive]

# Resolve all
sl deps resolve [--no-cache]

# Update all
sl deps update [--force] [repo-url]

# Remove dependency
sl deps remove <repo-url> <spec-path>
```

---

**For more details on bd-issue-tracking**, see: [bd-issue-tracking](.claude/skills/bd-issue-tracking/SKILL.md)
