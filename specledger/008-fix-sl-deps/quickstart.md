# Quickstart Guide: SpecLedger Dependencies

**Feature**: 008-fix-sl-deps
**Version**: 1.1.0
**Last Updated**: 2026-02-09

## Overview

SpecLedger dependencies allow you to reference external specifications from other Git repositories. Dependencies are cached locally for offline access and can be referenced across projects using a simple `alias:artifact` syntax.

---

## Prerequisites

- SpecLedger CLI v1.1.0 or later
- Git installed and configured
- Access to dependency repositories (SSH keys or tokens)

---

## 5-Minute Getting Started

### 1. Initialize a Project with Dependencies

```bash
# Create a new project (automatically configures artifact_path)
sl new my-service --type service

# Or initialize an existing project
sl init --artifact-path specledger/
```

### 2. Add Your First Dependency

```bash
# Add a SpecLedger repository (auto-detects artifact_path)
sl deps add git@github.com:org/platform-specs main specs/ --alias platform

# Add a non-SpecLedger repository (specify artifact_path manually)
sl deps add https://github.com/external/api-docs --artifact-path docs/openapi/ --alias api
```

### 3. Resolve Dependencies

```bash
# Download and cache all dependencies
sl deps resolve
```

Output:
```
Resolving Dependencies
Resolving 2 dependencies...

1. git@github.com:org/platform-specs
   Alias:  platform
   Cache:  /Users/user/.specledger/cache/platform-specs
   Status: ✓ abc123def4

2. https://github.com/external/api-docs
   Cache:  /Users/user/.specledger/cache/api-docs
   Status: ✓ def456abc1

Resolved 2/2 dependencies
```

### 4. List Dependencies

```bash
sl deps list
```

Output:
```
Dependencies (2 total)

1. git@github.com:org/platform-specs
   Branch:  main
   Path:    platform
   Alias:   platform
   Artifact Path: specs/
   Status:  ✓ abc123def4

2. https://github.com/external/api-docs
   Alias:   api
   Artifact Path: docs/openapi/
   Status:  ✓ def456abc1
```

### 5. Reference Artifacts

Once dependencies are resolved, reference them using the `alias:artifact` format:

```
# Reference platform's core spec
See platform:core.md for platform requirements.

# Reference API documentation
The API contract is defined in api:openapi.yaml.
```

---

## Common Workflows

### Adding SpecLedger Repository Dependencies

For repositories using SpecLedger, the `artifact_path` is auto-discovered:

```bash
sl deps add git@github.com:org/shared-specs --alias shared
```

The system will:
1. Clone the repository temporarily
2. Read its `specledger.yaml`
3. Extract the `artifact_path` value
4. Store it in your dependency configuration

### Adding Non-SpecLedger Repository Dependencies

For repositories not using SpecLedger, specify the `artifact_path` manually:

```bash
sl deps add https://github.com/external/repo --artifact-path path/to/specs --alias external
```

### Updating Dependencies

Check for and apply updates:

```bash
# Check all dependencies for updates
sl deps update

# Update a specific dependency
sl deps update platform

# Update all without prompting
sl deps update --yes
```

### Removing Dependencies

```bash
# Remove by alias
sl deps remove platform

# Remove by URL
sl deps remove git@github.com:org/specs
```

---

## Understanding artifact_path

### What is artifact_path?

The `artifact_path` is the directory within a repository where specification documents are stored. It's configured in `specledger.yaml`:

```yaml
# In platform-specs/specledger.yaml
artifact_path: specs/
```

### How It Works

When you reference an artifact:

```
Reference: platform:api.md
```

The system resolves it as:

```
<project.artifact_path> + <dependency.path> + "/" + "api.md"
```

Example:
```
project.artifact_path: specledger/
dependency.path: platform
dependency.artifact_path: specs/ (for finding files in cache)

Result: specledger/platform/api.md
```

### Project vs Dependency artifact_path

| Context | artifact_path Meaning |
|---------|---------------------|
| **Project** | Where YOUR project stores artifacts (e.g., `specledger/`) |
| **Dependency** | where the DEPENDENCY stores artifacts (e.g., `specs/`) |

---

## Directory Structure

### After Adding Dependencies

```
my-service/
├── specledger/
│   ├── specledger.yaml          # Project metadata with dependencies
│   └── .clerk/                   # Your project's artifacts
└── .claude/
    └── commands/                 # Claude Code integration

~/.specledger/cache/              # Global cache (all projects)
├── platform-specs/               # Cloned dependencies
│   └── specs/
│       ├── core.md
│       └── api.md
└── api-docs/
    └── docs/
        └── openapi/
            └── openapi.yaml
```

---

## Reference Resolution Examples

### Example 1: Simple Reference

```yaml
# specledger.yaml
artifact_path: specledger/
dependencies:
  - url: git@github.com:org/api-specs
     artifact_path: docs/
     path: api
     alias: api
```

```
Reference: api:openapi.yaml
Resolves to: specledger/api/openapi.yaml
```

### Example 2: Nested Path

```yaml
# specledger.yaml
artifact_path: docs/specifications/
dependencies:
  - url: git@github.com:org/platform
     artifact_path: specs/
     path: platform-core
     alias: platform
```

```
Reference: platform:auth/requirements.md
Resolves to: docs/specifications/platform-core/auth/requirements.md
```

### Example 3: Multiple Dependencies

```yaml
# specledger.yaml
artifact_path: specledger/
dependencies:
  - url: git@github.com:org/auth-specs
     alias: auth
     path: auth
     artifact_path: specs/
  - url: git@github.com:org/data-specs
     alias: data
     path: data
     artifact_path: documentation/
```

```
References:
- auth:oauth.md → specledger/auth/oauth.md
- data:schema.md → specledger/data/schema.md
```

---

## Configuration Examples

### Minimal Configuration

```yaml
# specledger.yaml
version: "1.0.0"

project:
  name: my-service
  short_code: ms
  created: 2026-02-09T10:00:00Z
  modified: 2026-02-09T10:00:00Z
  version: "0.1.0"

artifact_path: specledger/

playbook:
  name: specledger
  version: "1.0.0"

dependencies:
  - url: git@github.com:org/specs
    alias: specs
```

### Full Configuration

```yaml
# specledger.yaml
version: "1.0.0"

project:
  name: my-service
  short_code: ms
  created: 2026-02-09T10:00:00Z
  modified: 2026-02-09T10:00:00Z
  version: "0.1.0"

artifact_path: specledger/

playbook:
  name: specledger
  version: "1.0.0"
  applied_at: 2026-02-09T10:00:00Z
  structure:
    - specledger/
    - .claude/

dependencies:
  - url: git@github.com:org/platform-specs
    branch: main
    artifact_path: specs/
    path: platform
    alias: platform
    resolved_commit: abc123def456789abc123def456789abc123def4
    framework: both
    import_path: @platform/spec

  - url: https://github.com/external/api-docs
    branch: v2.0
    artifact_path: docs/openapi/
    path: api-docs
    alias: api-docs
    resolved_commit: def456abc123789def456abc123789def456abc12
    framework: none
```

---

## Troubleshooting

### Issue: "artifact_path not specified"

**Error**:
```
Error: artifact_path must be specified for non-SpecLedger repositories
```

**Solution**: Add the `--artifact-path` flag:

```bash
sl deps add <url> --artifact-path <path>
```

### Issue: "Failed to detect artifact_path"

**Error**:
```
Warning: Could not detect artifact_path from specledger.yaml
```

**Solution**: The dependency may not be a SpecLedger repository. Specify manually:

```bash
sl deps add <url> --artifact-path <path>
```

### Issue: "Dependency already exists"

**Error**:
```
Error: dependency already exists: git@github.com:org/specs
```

**Solution**: Remove the existing dependency first:

```bash
sl deps remove git@github.com:org/specs
sl deps add git@github.com:org/specs <new-parameters>
```

### Issue: "Failed to clone repository"

**Error**:
```
Warning: Failed to clone git@github.com:org/private-specs: authentication failed
```

**Solution**:
1. Set up SSH keys for private repositories
2. Or use HTTPS with a personal access token
3. Verify you have access to the repository

### Issue: "Reference not found"

**Error**:
```
Error: artifact not found: specledger/platform/api.md
```

**Solution**:
1. Run `sl deps resolve` to ensure dependencies are cached
2. Verify the artifact exists in the dependency repository
3. Check the `artifact_path` configuration

---

## Best Practices

### 1. Use Meaningful Aliases

```bash
# Good
sl deps add git@github.com:org/platform-auth-specs --alias auth

# Less clear
sl deps add git@github.com:org/platform-auth-specs --alias specs
```

### 2. Pin Specific Branches for Production

```bash
sl deps add git@github.com:org/specs v1.0 --alias prod-specs
```

### 3. Commit resolved_commit Values

After running `sl deps resolve`, commit the updated `specledger.yaml`:

```bash
git add specledger/specledger.yaml
git commit -m "Resolve dependency commits"
```

This ensures team members use the same dependency versions.

### 4. Use Reference Format in Documentation

Instead of hardcoding paths, use the reference format:

```markdown
<!-- Good -->
See [auth:oauth.md](auth:oauth.md) for OAuth requirements.

<!-- Less flexible -->
See [OAuth requirements](../deps/auth/specs/oauth.md).
```

### 5. Run Resolve Before Offline Work

```bash
sl deps resolve
```

This ensures all dependencies are cached locally.

---

## Advanced Usage

### Custom Cache Location

```bash
export SPECLEDGER_CACHE_DIR=/custom/cache/path
sl deps resolve
```

### Force Re-download

```bash
sl deps resolve --no-cache
```

### Batch Add Dependencies

Create a script:

```bash
#!/bin/bash
deps=(
  "git@github.com:org/spec1 main specs/ spec1"
  "git@github.com:org/spec2 develop docs/ spec2"
)

for dep in "${deps[@]}"; do
  sl deps add $dep
done

sl deps resolve
```

---

## Migration from Previous Versions

### If Your Project Existed Before v1.1.0

1. **Update specledger.yaml**:

```yaml
# Add this line
artifact_path: specledger/
```

2. **Re-resolve dependencies**:

```bash
sl deps resolve --no-cache
```

The system will auto-discover `artifact_path` for SpecLedger dependencies.

---

## Next Steps

- Read the [CLI API contracts](contracts/cli-api.md) for detailed command reference
- Review the [data model](data-model.md) for entity definitions
- Check the [research document](research.md) for implementation details

---

## Support

- Issues: https://github.com/specledger/specledger/issues
- Documentation: https://docs.specledger.dev
