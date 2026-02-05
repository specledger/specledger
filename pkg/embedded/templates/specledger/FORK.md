# SpecLedger Specification Directory

This directory contains your project's specifications using the SpecLedger framework.

## What is SpecLedger?

SpecLedger is a unified framework for:
- **Project Bootstrap**: Creating new projects with interactive CLI (`sl new`)
- **Dependency Management**: Managing external specification dependencies (`sl deps`)
- **Issue Tracking**: Tracking work across sessions using beads (bd)

## Directory Structure

```
specledger/
├── specledger.mod       # Dependency manifest (declared dependencies)
├── specledger.sum       # Lockfile (resolved dependencies with hashes)
├── FORK.md             # This file
└── specs/              # Your specification files (optional)
    ├── 001-feature-name/
    │   ├── spec.md     # Feature specification
    │   └── plan.md     # Implementation plan
    └── ...
```

## Managing Dependencies

### Add a Dependency

```bash
sl deps add git@github.com:org/spec-repo main spec.md --alias myspec
```

### List Dependencies

```bash
sl deps list
```

### Resolve (Download) Dependencies

```bash
sl deps resolve
```

Dependencies are cached locally at `~/.specledger/cache/` for offline use.

## Issue Tracking

SpecLedger uses [bd (beads)](https://github.com/steveyegge/beads) for issue tracking.

### Tracking Work

```bash
bd create "Implement feature X" --priority high
bd show sl-001           # View issue details
bd comments add sl-001 "Progress update"
bd close sl-001 "Completed"
```

### Finding Work

```bash
bd ready --limit 5       # Find issues ready to work on
bd search "database" --limit 10
```

## LLM Integration

Cached dependencies can be easily referenced by AI agents:

```markdown
## Dependencies

This component extends:
- `myspec` (git@github.com:org/spec-repo/main/spec.md)

Cached at: ~/.specledger/cache/github.com/org/spec-repo/<commit>/spec.md
```

## Configuration

Global config: `~/.config/specledger/config.yaml`

```yaml
default_project_dir: ~/demos
preferred_shell: claude-code
tui_enabled: true
```

## Documentation

- CLI Help: `sl --help`
- Documentation: https://specledger.io/docs
- GitHub: https://github.com/specledger/specledger
