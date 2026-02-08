# SpecLedger Project Guidelines

This project uses **SpecLedger** - a unified system that combines:

1. **Issue Tracking** (via `bd` - beads) - Track complex work with dependency graphs
2. **Specification Dependency Management** (via `sl` CLI) - Manage external spec dependencies

---

## Issue Tracking with `bd`

**IMPORTANT**: Use `bd` for ALL issue tracking. Do NOT use markdown TODOs or task lists.

### Why `bd`?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Auto-syncs to `.beads/issues.jsonl` for version control
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

### Quick Start

```bash
# Check for ready work
bd ready [--limit 10]

# Create new issues
bd create "Issue title" -t bug|feature|task -p 0-4
bd create "Found bug" -p 1 --deps discovered-from:bd-123

# Claim and update
bd update bd-42 --status in_progress
bd update bd-42 --priority 1

# Document discoveries
bd comments add bd-42 "Discovered: API requires auth header"

# Complete work
bd close bd-42 --reason "Completed"
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue with `--deps discovered-from:<id>`
5. **Complete**: `bd close <id> --reason "Done"`
6. **Commit together**: Always commit `.beads/issues.jsonl` with your code changes

### Auto-Sync

`bd` automatically syncs with git:
- Exports to `.beads/issues.jsonl` after changes (5s debounce)
- Imports from JSONL when newer (e.g., after `git pull`)

**See**: [skills/specledger-issue-tracking](.claude/skills/specledger-issue-tracking/README.md) for details.

---

## Specification Dependencies with `sl deps`

**IMPORTANT**: Use `sl deps` to manage external specification dependencies.

### Why Dependency Management?

- Track which external specs your project references
- Cache dependencies locally at `~/.specledger/cache/` for offline use
- Enable LLMs to easily reference cached specifications
- Reproducible builds with locked commit hashes

### Quick Start

```bash
# Add a dependency
sl deps add git@github.com:org/api-spec main spec.md --alias api

# List all dependencies
sl deps list

# Download and cache dependencies
sl deps resolve

# Update to latest versions
sl deps update

# Remove a dependency
sl deps remove git@github.com:org/api-spec
```

### File Structure

- `specledger/specledger.yaml` - Dependency manifest (add/remove/update deps here)
- `specledger/specledger.sum` - Lockfile with resolved commits and hashes
- `~/.specledger/cache/` - Local cache of downloaded dependencies

**See**: [skills/specledger-deps](.claude/skills/specledger-deps/README.md) for details.

---

## Integration: Using Both Together

When working on features that depend on external specifications:

1. **Create an issue** for the feature: `bd create "Implement API integration" -p 1`
2. **Add dependencies** needed: `sl deps add git@github.com:org/api-spec`
3. **Download dependencies**: `sl deps resolve`
4. **Reference cached specs** in your implementation
5. **Close the issue**: `bd close <id> --reason "Completed"`

---

## Important Rules

- ✅ Use `bd` for ALL task tracking
- ✅ Use `sl deps` for specification dependencies
- ✅ Check `bd ready` before asking "what should I work on?"
- ✅ Run `sl deps resolve` after adding dependencies
- ✅ Commit `.beads/issues.jsonl` with code changes
- ✅ Commit `specledger/specledger.sum` lockfile
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems

---

## Commit & Pull Request Guidelines

Follow conventional commit prefixes (`feat:`, `fix:`, `chore:`, `docs:`). Keep messages imperative and under 72 characters. Reference related issues in the body.

Example:
```
feat: add API integration with external auth spec

Implements OAuth2 flow as specified in external auth dependency.
Closes bd-42.

Co-Authored-By: Claude (GLM-4.7-Flash) <noreply@anthropic.com>
```
