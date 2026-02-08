---
description: Resolve and checkout all dependencies locally
---

Checkout all dependencies to `specledger/deps/` for offline access and commit SHA resolution.

## What It Does

1. **Clone dependencies** to `specledger/deps/<alias>/`
2. **Resolve commit SHAs** for reproducibility
3. **Update metadata** with resolved commits
4. **Handle duplicates** and missing branches

## Quick Usage

```bash
sl deps resolve
```

## Execution Flow

1. **Read dependencies** from `specledger/specledger.yaml`

2. **For each dependency**:
   - Check if already resolved (commit SHA exists)
   - Clone or update repository
   - Checkout specified branch
   - Resolve current commit SHA
   - Store commit SHA in metadata

3. **Report results**:
   - Number of dependencies resolved
   - Any warnings or errors encountered

## Output Format

```
Resolving dependencies...

  ✓ api-specs (git@github.com:org/specs.git)
    Branch: main
    Commit: abc123def456...
    Path: specledger/deps/api-specs

  ✓ shared-specs (https://github.com/org/shared.git)
    Branch: develop
    Commit: def456abc123...
    Path: specledger/deps/shared-specs

Resolved 2 dependencies
```

## Directory Structure

After resolution:
```
specledger/
├── deps/
│   ├── api-specs/          # Cloned repo
│   │   └── openapi.yaml
│   └── shared-specs/       # Cloned repo
│       └── common/
│           └── spec.md
└── specledger.yaml         # Updated with commit SHAs
```

## Error Handling

**Invalid branch:**
```
⚠ Warning: branch 'feature' not found for api-specs
  Using default branch 'main' instead
```
Resolution: Uses default branch, updates metadata.

**Authentication required:**
```
Error: failed to clone git@github.com:org/private-specs.git
  Please ensure you have access to this repository
```
Resolution: Set up SSH keys or use HTTPS with token.

**Network error:**
```
⚠ Warning: failed to clone api-specs: network timeout
  Skipping for now, you can retry later
```
Resolution: Dependencies without commits remain unresolved.

## When to Run Resolution

Run `sl deps resolve` when:
- Setting up a new development machine
- After adding new dependencies
- Preparing for offline work
- Want to pin specific commits for reproducibility

## Reproducibility

Resolved commit SHAs ensure:
- Same version of dependency is used across builds
- Team members work with identical specs
- Historical builds can be reproduced
- Changes in dependencies are tracked

## Integration

- Use with `sl deps list` to see resolution status
- Run after `sl deps add` to checkout new dependency
- Commit `specledger.yaml` to share resolved SHAs with team
