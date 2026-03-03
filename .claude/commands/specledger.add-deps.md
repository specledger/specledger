---
description: Add a new specification dependency to the project
---

Add an external specification as a dependency.

## User Input

```text
$ARGUMENTS
```

**First argument is the Git URL** (required).
**Second argument** is optional branch name.
**--alias flag is required**.
**--artifact-path flag is optional** (for non-SpecLedger repos).

## Quick Usage

```bash
sl deps add git@github.com:org/specs --alias specs
sl deps add https://github.com/org/api-docs --alias api --artifact-path docs/openapi/
sl deps add git@github.com:user/repo develop --alias platform
```

## Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `git-url` | Yes | - | Git repository URL (SSH or HTTPS) |
| `branch` | No | `main` | Branch to checkout |
| `--alias` | **Yes** | - | Short name for the dependency (used as reference path) |
| `--artifact-path` | No | auto-detected | Path to artifacts within dependency repo |

## Execution Flow

1. **Parse arguments**:
   - Extract Git URL (first argument)
   - Parse optional branch
   - Validate --alias is provided
   - Validate --artifact-path if provided

2. **Detect framework**:
   - Check if repository uses SpecKit, OpenSpec, or both
   - Display detected framework type

3. **Auto-detect artifact_path** (if not manually specified):
   - For SpecLedger repos: Clone temporarily and read specledger.yaml
   - Extract artifact_path value
   - Store in dependency configuration

4. **Auto-download dependency**:
   - Clone repository to `~/.specledger/cache/<alias>/`
   - Resolve current commit SHA
   - Store commit in dependency configuration

5. **Add to metadata**:
   - Update `specledger/specledger.yaml`
   - Add new dependency with all resolved information
   - Save updated metadata

6. **Report success**:
   - Show what was added
   - Display resolved commit SHA

## Examples

```bash
# Add SpecLedger repository (auto-detects artifact_path)
sl deps add git@github.com:org/platform-specs --alias platform

# Add with specific branch
sl deps add git@github.com:org/specs develop --alias specs

# Add non-SpecLedger repository (specify artifact_path manually)
sl deps add https://github.com/org/api-docs --alias api --artifact-path docs/openapi/

# Add with both branch and artifact-path
sl deps add git@github.com:user/repo.git v1.0 --alias user-repo --artifact-path specifications/
```

## Error Handling

**Missing --alias flag:**
```
Error: required flag(s) "alias" not set
```
Solution: Add the `--alias` flag with a short name for the dependency.

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

**Not in a SpecLedger project:**
```
Error: failed to find project root: not in a SpecLedger project
```
Solution: Navigate to your project directory (contains `specledger/specledger.yaml`).

## Auto-Download Behavior

When you add a dependency, it is **automatically downloaded** to the cache:
- Cache location: `~/.specledger/cache/<alias>/`
- Current commit SHA is resolved and stored
- No separate `sl deps resolve` command needed (resolve is for manual refresh only)

## Artifact Path Discovery

**For SpecLedger repositories:**
- System clones the repository temporarily
- Reads `specledger.yaml` from the dependency
- Extracts `artifact_path` value
- Stores in your dependency configuration

**For non-SpecLedger repositories:**
- Use `--artifact-path` flag to specify manually
- Example: `--artifact-path docs/openapi/`
- This tells SpecLedger where artifacts are located within the dependency

## When to Add Dependencies

Add a dependency when:
- Your service implements an API defined elsewhere
- Building on top of shared specification documents
- Need to reference standards or protocols
- Creating hierarchical spec relationships
