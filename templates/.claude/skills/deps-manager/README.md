# Deps Manager Skill

This skill documents the SpecLedger CLI `deps` commands for managing external specification dependencies.

## Commands

| Command | Description |
|---------|-------------|
| `sl deps add <repo-url>` | Add a specification dependency |
| `sl deps list` | List all declared dependencies |
| `sl deps resolve` | Resolve all dependencies and generate spec.sum |
| `sl deps update` | Update dependencies to latest compatible versions |
| `sl deps remove <repo-url> <spec-path>` | Remove a dependency |

## Quick Start

```bash
# Add a dependency
sl deps add github.com/org/project-spec main specs/project.md --alias myproject

# List all dependencies
sl deps list

# Resolve and generate lockfile
sl deps resolve

# Update to latest versions
sl deps update
```

## See Also

- **[bd-issue-tracking](../bd-issue-tracking/README.md)** - Track work across sessions with dependency graphs
- **[gum-tui](../gum-tui/README.md)** - Interactive terminal UI with beautiful prompts
