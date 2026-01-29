# Quick Start: Spec Dependency Linking

**Get started in 5 minutes** with external specification dependency management.

## Prerequisites

- Go 1.21+ installed
- Git installed
- Access to a Git repository (GitHub, GitLab, etc.)

## Installation

### 1. Use the Existing `sl` Bootstrap Script

```bash
# The sl script already exists in ./sl
# Make sure it's executable
chmod +x ./sl

# Or install it globally
cp ./sl /usr/local/bin/sl
```

### 2. Verify Installation

```bash
./sl --version
# or if installed globally
sl --version
# Should output: sl version 1.0.0
```

## Quick Setup

### Bootstrap a Project

```bash
# Use the existing sl bootstrap script
cd ~/demos/your-project
# or create a new project
./sl
# This creates the full SpecLedger infrastructure including spec.md
```

### Add Your First External Dependency

```bash
# Add authentication service spec from GitHub
sl deps add https://github.com/example/auth-service v1.2.0 specs/auth.md

# Add with alias
sl deps add https://github.com/example/user-service main specs/user.md --alias users
```

### Resolve Dependencies

```bash
sl deps resolve
```

This will:
- Fetch the external specifications
- Generate `spec.sum` with cryptographic hashes
- Cache the specifications locally

### Validate References

```bash
# Validate external references in your spec.md
sl refs validate
```

## Basic Workflow

### 1. Declare Dependencies

Edit `spec.mod` to declare your dependencies:

```text
# spec.mod
require https://github.com/example/auth-service v1.2.0 specs/auth.md
require https://github.com/example/user-service main specs/user.md id my-users
```

### 2. Reference External Specs

Use markdown links in your `spec.md`:

```markdown
# My Service API

This service depends on the authentication system from [Auth Service](https://github.com/example/auth-service#auth-service#authentication).

For user management, see [User Management](https://github.com/example/user-service#my-users#user-api).
```

### 3. Resolve and Validate

```bash
# Resolve all dependencies
sl deps resolve

# Validate all references
sl refs validate
```

### 4. Update Dependencies

```bash
# Update to latest versions
sl deps update

# Update specific dependency
sl deps update https://github.com/example/auth-service
```

## Commands Reference

### Dependency Management

```bash
# Add a dependency
sl deps add <repo-url> [branch] [path] [--alias <name>]

# List dependencies
sl deps list [--include-transitive]

# Remove a dependency
sl deps remove <repo-url> <spec-path>

# Update dependencies
sl deps update [--force] [specific-repo]
```

### Resolution and Validation

```bash
# Resolve dependencies (generate spec.sum)
sl deps resolve [--no-cache] [--deep]

# Validate external references
sl refs validate [--strict]

# Check for conflicts
sl conflicts check
```

### Graph Visualization

```bash
# Show dependency graph
sl graph show [--format=svg|json|text]

# Export graph to file
sl graph export --format=svg --output=deps.svg

# Show transitive dependencies
sl graph transitive
```

### Vendoring

```bash
# Vendor dependencies for offline use
sl vendor --output=specs/vendor

# Update vendored dependencies
sl vendor update

# Clean vendor directory
sl vendor clean
```

## Configuration

### Authentication

For private repositories, configure authentication:

```bash
# Set GitHub token
sl config set github-token YOUR_GITHUB_TOKEN

# Set GitLab token
sl config set gitlab-token YOUR_GITLAB_TOKEN

# Use SSH key
sl config set ssh-key ~/.ssh/id_rsa
```

### Cache Settings

```bash
# Configure cache size
sl config set cache-size 100

# Set cache timeout
sl config set cache-ttl 1h

# Clear cache
sl cache clear
```

## Troubleshooting

### Common Issues

**1. Permission Denied Error**

```bash
# Check authentication
sl auth check

# Re-authenticate
sl auth login
```

**2. Dependency Resolution Failed**

```bash
# Check network connectivity
curl -I https://github.com

# Clear cache and retry
sl cache clear
sl deps resolve --no-cache
```

**3. Reference Validation Failed**

```bash
# Check specific reference
sl refs validate --verbose

# List all references
sl refs list
```

### Debug Mode

```bash
# Enable debug logging
sl --debug deps resolve

# Show detailed information
sl deps resolve --verbose
```

## Examples

### Example 1: Microservice Dependencies

```bash
# Add common dependencies
sl deps add https://github.com/company/common-apis v2.1.0 specs/common.md
sl deps add https://github.com/company/database-schemas v1.3.0 specs/database.md

# Reference in service spec
# my-service/spec.md
"""
# User Service

Implements user management using [Common APIs](https://github.com/company/common-apis#common-apis#user-api).

Database schema defined in [Database Schema](https://github.com/company/database-schemas#db-schema#users).
"""

# Resolve and validate
sl deps resolve
sl refs validate
```

### Example 2: Cross-Team Sharing

```bash
# Team A publishes their spec
cd team-a-project
sl deps add https://github.com/team-b/shared-types v1.0.0 specs/types.md

# Team B can use Team A's spec
cd team-b-project
sl deps add https://github.com/team-a/project-a main specs/a-specs.md

# Both teams can reference each other's specs
# team-b/spec.md
"""
# Integration Service

Uses [Team A Types](https://github.com/team-a/project-a#a-project#user-type)
and [Team B Types](https://github.com/team-b/shared-types#shared-types#common-type).
"""
```

## Next Steps

1. **Read the full documentation** at [docs.sl.com](https://docs.sl.com)
2. **Explore advanced features** like conflict resolution and vendoring
3. **Set up CI/CD integration** for automatic dependency updates
4. **Join the community** on [GitHub Discussions](https://github.com/sl/sl/discussions)

## Support

- **Documentation**: [docs.sl.o](https://docs.sl.o)
- **Issues**: [GitHub Issues](https://github.com/sl/sl/issues)
- **Discussions**: [GitHub Discussions](https://github.com/sl/sl/discussions)
- **Email**: support@sl.o