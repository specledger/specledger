# Integration Guide: Spec Dependency Linking with `sl` CLI

**Updated**: 2026-01-30
**Branch**: `002-spec-dependency-linking`

## Overview

The Spec Dependency Linking feature integrates seamlessly with the existing `sl` (SpecLedger) bootstrap script. This document outlines how the new dependency management functionality will be integrated into the existing CLI toolchain.

## Current `sl` Script

The existing `sl` script (`./sl`) provides:

1. **Project Bootstrapping**: Creates new SpecLedger projects with complete infrastructure
2. **Tool Installation**: Sets up `bd`, `perles`, and other tools via `mise.toml`
3. **Issue Tracking**: Initializes Beads issue tracker
4. **Configuration**: Sets up Claude Code skills and commands

## Integration Strategy

### 1. CLI Architecture

The new dependency management commands will be added to the existing `sl` tool:

```bash
# Existing commands
sl --help
sl bootstrap  # (same as running ./sl)

# New dependency management commands
sl deps add <repo-url> [branch] [path]
sl deps list
sl deps resolve
sl deps update
sl refs validate
sl graph show
sl vendor
```

### 2. File Structure Integration

```
# After integration
sl
├── bootstrap     # (existing) - Create new project
├── deps          # (new) - Dependency management
│   ├── add.go
│   ├── list.go
│   ├── resolve.go
│   └── update.go
├── refs          # (new) - Reference management
│   ├── validate.go
│   └── list.go
├── graph         # (new) - Graph visualization
├── vendor        # (new) - Vendoring
├── config        # (existing) - Tool configuration
└── version.go    # (existing) - Version info
```

### 3. Mise Configuration Update

The `mise.toml` file will be updated to include the new `sl` dependency management tool:

```toml
[tools]
  # Existing tools
  "github.com/steveyegge/beads" = "0.28.0"
  "github.com/specledger/perles" = "latest"

  # New dependency management tool
  "github.com/specledger/specledger" = "1.0.0"
```

### 4. Bootstrap Script Enhancement

The existing `sl` bootstrap script will be enhanced to:

1. **Initialize `spec.mod`**: Create initial dependency manifest file
2. **Add Common Dependencies**: Pre-populate with team-specific dependencies
3. **Setup Configuration**: Initialize authentication and cache settings
4. **Integration Documentation**: Add dependency management to next steps

## Implementation Steps

### Phase 1: Core CLI Extension

1. **Extend `main.go`**: Add subcommand parsing using cobra
2. **Create Command Structure**: Implement `sl deps`, `sl refs`, `sl graph`, `sl vendor` commands
3. **Update Help System**: Integrate new commands into existing help system

### Phase 2: File Integration

1. **Bootstrap Enhancement**: Modify `sl` to create `spec.mod` during bootstrap
2. **Config Management**: Store dependency settings in existing config structure
3. **Cache Integration**: Use existing cache directory structure

### Phase 3: Tool Integration

1. **Mise.toml Update**: Automatically add new tool to mise configuration
2. **Version Management**: Keep version aligned with existing tools
3. **Documentation Update**: Update AGENTS.md with new commands

## Backward Compatibility

### Existing Commands Unchanged
```bash
# These continue to work exactly as before
./sl                    # Bootstrap new project
bd ready               # Find unblocked issues
bd create              # Create new issue
perles                 # Launch TUI
```

### New Commands Added
```bash
# New functionality
sl deps add ...         # Add external dependency
sl deps resolve         # Resolve dependencies
sl refs validate        # Validate references
```

## Migration Path

### For Existing Projects

1. **Manual Enable**: Run `sl init-deps` in existing projects
2. **Automatic Detection**: Commands detect if `spec.mod` exists
3. **Graceful Fallback**: Commands work in non-dependency projects

### For New Projects

1. **Bootstrap Creates `spec.mod`**: New projects get dependency support automatically
2. **Pre-populated Dependencies**: Common team dependencies added by default
3. **Setup Complete**: Ready for external dependency usage

## Configuration Files Integration

### Existing Structure
```
.spec-config/
├── auth.json      # Authentication tokens (new entries added)
├── cache.db       # Cache metadata (extended for dependencies)
└── config.yaml    # Tool configuration (updated)
```

### New Configuration Options
```yaml
# .spec-config/config.yaml
dependencies:
  cache-size: 100        # MB
  cache-ttl: 1h          # Time to live
  default-branch: main   # Default branch for dependencies
auth:
  github-token: ""       # Added by sl config set
  gitlab-token: ""       # Added by sl config set
```

## Error Handling Integration

### Existing Error Patterns
```bash
ERROR: Must run from specledger repository root
ERROR: gum not found. Install from https://github.com/charmbracelet/gum
ERROR: mise not found. Install from https://mise.jdx.dev
```

### New Error Messages
```bash
ERROR: Invalid repository URL
ERROR: Dependency resolution failed
ERROR: Reference validation failed
ERROR: Authentication required for private repository
```

## Performance Considerations

### Cache Sharing
- **Existing Cache**: `bd` and other tools' cache respected
- **Dedicated Cache**: Separate cache for dependency content
- **Cache Coordination**: Shared configuration for cache management

### Resource Management
- **Memory Limits**: 512MB limit for dependency resolution
- **Parallel Operations**: Leverage existing concurrency patterns
- **Cleanup**: Existing cleanup mechanisms extended

## Testing Integration

### Existing Test Structure
```
tests/
├── unit/
├── integration/
└── fixtures/
```

### New Test Categories
```bash
# New test suites
tests/unit/deps/
tests/unit/refs/
tests/integration/deps-resolve/
tests/integration/refs-validate/
```

## Documentation Updates

### AGENTS.md Enhancement
- Add new command references
- Update workflow examples
- Add troubleshooting section

### CLAUDE.md Updates
- Add dependency management best practices
- Update tool integration guidance
- Add security considerations

## Rollout Plan

### Step 1: Core CLI (Week 1)
- Implement `sl deps`, `sl refs` commands
- Update bootstrap script

### Step 2: Advanced Features (Week 2)
- Implement graph visualization
- Add vendoring support

### Step 3: Integration (Week 3)
- Update mise configuration
- Enhance documentation
- Add backward compatibility

### Step 4: Testing & Polish (Week 4)
- Comprehensive testing
- Performance optimization
- Community feedback

## Troubleshooting

### Common Issues

**Command not found**
```bash
# Solution: Update mise
mise install
```

**Cache conflicts**
```bash
# Solution: Clear specific cache
sl cache clear deps
```

**Authentication issues**
```bash
# Solution: Re-authenticate
sl auth login
```

## Support

- **Existing Channels**: GitHub Issues, Discord, Email
- **Documentation**: Updated AGENTS.md, CLAUDE.md
- **Training**: Updated examples and tutorials