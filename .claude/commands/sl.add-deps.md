---
description: Add a new specification dependency to the project
---

Add an external specification as a dependency.

## User Input

```text
$ARGUMENTS
```

**First argument is the Git URL** (required).
**Remaining arguments** are parsed as: `[branch] [path] [--alias name]`

## Quick Usage

```bash
sl deps add git@github.com:org/specs main specs/api.md
sl deps add https://github.com/org/specs.git --alias api
sl deps add git@github.com:user/repo develop docs/spec.md --alias user-spec
```

## Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `git-url` | Yes | - | Git repository URL (SSH or HTTPS) |
| `branch` | No | `main` | Branch to checkout |
| `path` | No | root | Path to spec file within repo |
| `--alias` | No | auto-generated | Short name for reference |

## Execution Flow

1. **Parse arguments**:
   - Extract Git URL (first argument)
   - Parse optional branch, path, alias
   - Validate URL format

2. **Check for duplicates**:
   - Run `sl deps list` to check existing dependencies
   - Warn if URL or alias already exists
   - Ask user to confirm if duplicate detected

3. **Add to metadata**:
   - Update `specledger/specledger.yaml`
   - Add new dependency to dependencies array
   - Save updated metadata

4. **Report success**:
   - Show what was added
   - Next steps: `sl deps resolve` to checkout locally

## Examples

```bash
# Add with all parameters
sl deps add git@github.com:api-team/specs main openapi.yaml --alias api-spec

# Add with URL only (uses defaults)
sl deps add https://github.com/org/specs.git

# Add with branch and alias
sl deps add git@github.com:user/repo.git develop --alias user-repo
```

## Error Handling

**Invalid Git URL:**
```
Error: invalid git URL format
Valid formats:
  SSH: git@github.com:org/repo.git
  HTTPS: https://github.com/org/repo.git
```
Solution: Fix the URL format and retry.

**Dependency already exists:**
```
Error: dependency already exists with URL: git@github.com:org/specs.git
```
Solution: Use `sl deps remove` first if replacing, or check if you meant a different dependency.

**Not in a SpecLedger project:**
```
Error: failed to find project root: not in a SpecLedger project
```
Solution: Navigate to your project directory (contains `specledger/specledger.yaml`).

## When to Add Dependencies

Add a dependency when:
- Your service implements an API defined elsewhere
- Building on top of shared specification documents
- Need to reference standards or protocols
- Creating hierarchical spec relationships

## Integration

After adding, consider:
- Run `sl deps resolve` to checkout locally
- Run `sl deps graph` to see dependency relationships
- Use `sl deps list` to verify addition
