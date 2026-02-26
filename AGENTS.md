# Repository Guidelines

## Issue Tracking with `sl issue`

**IMPORTANT**: This project uses the built-in **`sl issue`** commands for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why `sl issue`?

- No external dependencies: Built directly into the `sl` CLI
- Git-friendly: Issues stored as JSONL files, one per spec
- Agent-optimized: JSON output, ready work detection, dependency links
- Prevents duplicate tracking systems and confusion

### Quick Start

**Check for open issues:**

```bash
sl issue list --status open
sl issue list --all  # across all specs
```

**Create new issues:**

```bash
sl issue create --title "Issue title" --type bug|feature|task --priority 0-4
sl issue create --title "Fix login" --type bug --priority 1
```

**Update issues:**

```bash
sl issue update SL-abc123 --status in_progress
sl issue update SL-abc123 --priority 1
```

**Complete work:**

```bash
sl issue close SL-abc123 --reason "Completed: Fixed login validation"
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)

### Issue IDs

Issues use deterministic IDs in format `SL-xxxxxx` (6 hex characters derived from SHA-256 hash of spec context + title + timestamp).

### Spec Storage

Issues are stored per-spec in `specledger/<spec>/issues.jsonl` to avoid merge conflicts when working on different features.

### Workflow for AI Agents

1. **Check open issues**: `sl issue list --status open`
2. **Claim your task**: `sl issue update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Complete**: `sl issue close <id> --reason "Done"`
5. **Commit together**: Always commit the `specledger/<spec>/issues.jsonl` file together with the code changes

### Dependency Management

Link issues with dependencies:

```bash
sl issue link SL-abc123 blocks SL-def456  # abc123 blocks def456
```

### Important Rules

- Use `sl issue` for ALL task tracking
- Issues are stored per-spec in `specledger/<spec>/issues.jsonl`
- Check `sl issue list --status open` before asking "what should I work on?"
- Do NOT create markdown TODO lists
- Do NOT use external issue trackers
- Do NOT duplicate tracking systems

## Commit & Pull Request Guidelines

Follow the existing conventional prefixes (`feat:`, `fix:`, `chore:`, `docs:`) and keep messages imperative and under 72 characters. Reference related issues in the body and mention migrations, proto changes, or new binaries explicitly. PRs should include a concise summary, testing evidence (`make test-unit`, `make test-integration`, etc.), and screenshots or CLI transcripts when behavior changes. Request reviews from domain owners and ensure generated artifacts and docs stay in sync with code changes.

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
