---
description: Manage specification dependencies using the SpecLedger CLI deps commands.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Overview

The SpecLedger CLI provides comprehensive dependency management for specifications. Dependencies are external specifications that your project references or builds upon.

## When to Use

Use this command when:
- **Adding dependencies** - Your spec references other specifications
- **Listing dependencies** - Understanding what specs your project depends on
- **Resolving dependencies** - Fetching and validating external specs
- **Updating dependencies** - Keeping specs at latest compatible versions
- **Removing dependencies** - Cleaning up unused spec references
- **Checking conflicts** - Finding dependency issues

## Commands Reference

### sl deps add

Add a specification dependency to your project.

**Usage:**
```bash
sl deps add <repo-url> [branch] [spec-path] [--alias <name>]
```

**Examples:**
```bash
# Basic usage
sl deps add git@github.com:org/api-spec.git

# With specific branch and spec path
sl deps add git@github.com:org/api-spec.git main specs/api.md

# With alias for easy reference
sl deps add git@github.com:org/api-spec.git --alias api
```

### sl deps list

List all declared dependencies.

**Usage:**
```bash
sl deps list [--include-transitive]
```

**Output:**
```
Dependencies (3):

1. git@github.com:org/api-spec.git
   Version: main
   Spec: specs/api.md
   Alias: api

2. git@github.com:org/auth-spec.git
   Version: v2.0
   Spec: specs/auth.md

3. git@github.com:org/db-spec.git
   Version: main
   Spec: spec.md
```

### sl deps resolve

Resolve all dependencies and generate spec.sum lockfile.

**Usage:**
```bash
sl deps resolve [--no-cache] [--deep]
```

This:
- Fetches external specifications
- Validates versions and commits
- Generates cryptographic hashes
- Creates lockfile for reproducible builds

### sl deps update

Update dependencies to latest compatible versions.

**Usage:**
```bash
sl deps update [repo-url] [--force]
```

### sl deps remove

Remove a dependency.

**Usage:**
```bash
sl deps remove <repo-url> <spec-path>
```

## Referencing Dependencies

Once a dependency is added, reference it in your specifications:

**In spec.md:**
```markdown
## API Integration

This component depends on: `api` (git@github.com:org/api-spec.git/main/specs/api.md)

Key requirements from api:
- Authentication: Must use OAuth2 flow
- Rate limits: 100 req/min per user
```

**In templates:**
```markdown
## Database Schema

Extends base schema defined in: `db-spec` (git@github.com:org/db-spec.git)

Additional fields:
- user_preferences: JSONB
```

## Conflict Detection

Check for dependency issues:

```bash
# Check for conflicts
sl conflict check

# Detect potential issues
sl conflict detect
```

## Best Practices

1. **Use meaningful aliases** - Short, memorable names for frequently referenced specs
2. **Pin versions** - Specify branches or tags for reproducibility
3. **Document references** - Always note which parts of external specs you use
4. **Regular updates** - Keep dependencies updated with `sl deps update`
5. **Resolve before committing** - Run `sl deps resolve` to update lockfile

## Error Handling

**Common errors:**

- **"Not a SpecLedger project"** - Run `sl new` or create `specs/spec.mod` first
- **"Invalid repository URL"** - Use `git@` format or HTTPS URL
- **"Dependency not found"** - Check the URL, branch, and spec path
- **"Signature verification failed"** - External spec may be corrupted

## Getting Help

```bash
sl deps --help
sl deps <command> --help
```
