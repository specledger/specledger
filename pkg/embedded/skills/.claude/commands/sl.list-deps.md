---
description: List all specification dependencies in the current project
---

List all dependencies stored in `specledger/specledger.yaml`.

Shows:
- Dependency URL (Git repository)
- Branch name
- File path within repository
- Alias (short reference name)
- Resolved commit SHA (if available)

Run this command when you need to:
- Check what external specs this project references
- Verify a dependency was added successfully
- Review the dependency graph before making changes

```bash
sl deps list
```

**Output format:**
```
Dependencies (3):
  api-specs (git@github.com:org/api-specs.git)
    Branch: main
    Path: openapi.yaml
    Commit: abc123...

  shared-specs (https://github.com/org/shared.git)
    Branch: develop
    Path: common/spec.md
    Commit: def456...

  platform-core (git@github.com:org/platform.git)
    Branch: main
    Path: core.yaml
    Not resolved
```

**Common patterns:**

1. **Check before adding**: Run `sl deps list` first to avoid duplicates
2. **Verify after add**: Confirm the dependency was added correctly
3. **Review impact**: Understand what removing a dependency would affect

**When no dependencies exist:**
```
No dependencies found.
Add dependencies with: sl deps add <git-url> [<branch>] [<path>]
```
