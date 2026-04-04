# sl Dependency Management

**When to Load**: Triggered when tasks involve cross-repo specification dependencies, `sl deps` commands, artifact caching, or multi-repo dependency resolution.

## Overview

`sl deps` manages external specification dependencies between repositories. Dependencies are declared in `specledger.yaml`, cached locally at `~/.specledger/cache/`, and optionally symlinked into your project's artifacts directory for direct access.

## Subcommands

| Command | Purpose | Output Mode |
|---------|---------|-------------|
| `sl deps add <url> [branch] -a <alias> [--link]` | Declare a new dependency | Progress + confirmation |
| `sl deps list` | Show all declared dependencies | Table (repo, version, resolved status) |
| `sl deps remove <url>` | Remove dependency declaration (cache kept) | Confirmation |
| `sl deps resolve` | Download and cache all dependencies | Progress per dependency |
| `sl deps update [url]` | Pull latest versions (all if no URL) | Progress + confirmation |
| `sl deps link` | Symlink cached deps into project artifacts | Confirmation per link |
| `sl deps unlink [alias]` | Remove symlinks (all if no alias) | Confirmation |

## Decision Criteria

### sl deps vs sl issue link

These are fundamentally different concepts:

**Use `sl deps`** for **repository-level artifact dependencies**:
- Your project references specifications from another repository
- You need external spec artifacts resolved locally for planning or implementation
- Multi-repo projects sharing specification artifacts

**Use `sl issue link`** for **work-item relationships** within a project:
- Task A blocks task B
- Issues that are related but not blocking

### Resolve Only vs Resolve + Link

**Resolve only (`sl deps resolve`):**
- CI environments where symlinks aren't needed
- Checking dependency availability before a workflow
- You'll reference cached artifacts by their cache path

**Resolve + link (`sl deps resolve --link`):**
- You want dependencies accessible at `<artifact_path>/deps/<alias>/`
- Claude Code or other tools need to read dependency artifacts directly
- Active development referencing external specs

### Cache Behavior

- `resolve` downloads to `~/.specledger/cache/` (reuses cache by default)
- `resolve --no-cache` / `-n` forces fresh download
- `remove` removes the declaration but **keeps the cache**
- `update` fetches latest versions and refreshes cache

## Workflow Patterns

### Pattern 1: Add and Link a New Dependency

```bash
# SpecLedger repos: artifact_path auto-detected
sl deps add git@github.com:org/api-spec --alias api

# Non-SpecLedger repos: specify artifact path manually
sl deps add https://github.com/org/api-docs --alias docs --artifact-path docs/openapi/

# Resolve and symlink in one step
sl deps resolve --link
```

### Pattern 2: Update Dependencies

```bash
sl deps update              # Update all to latest
sl deps update git@github.com:org/spec  # Update one
sl deps link                # Re-link after update
```

### Pattern 3: Clean Up

```bash
sl deps unlink api          # Remove symlink for one
sl deps unlink              # Remove all symlinks
sl deps remove git@github.com:org/api-spec  # Remove declaration (cache remains)
```

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| "failed to find project root" | No `specledger.yaml` in parent dirs | Run from within a SpecLedger project, or `sl init` first |
| "invalid repository URL" | URL doesn't match git URL patterns | Use `git@host:org/repo` or `https://host/org/repo` format |
| "Could not auto-detect artifact_path" | Remote repo lacks `specledger.yaml` | Add `--artifact-path` flag to specify manually |
| Symlink errors | Missing cache or permissions | Run `sl deps resolve` first, check directory permissions |

## Token Efficiency (D21)

- **list**: Compact table (~200 tokens for 10 dependencies)
- **add/remove/resolve/update**: Progress indicators + minimal confirmation
- **link/unlink**: One line per symlink created/removed

For full flag details, run `sl deps <subcommand> --help`.
