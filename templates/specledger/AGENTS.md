# SpecLedger Directory Guidelines

## Purpose

The `specledger/` directory contains your project's specification documents and related artifacts. This is where you define your system's architecture, features, and technical decisions.

## Directory Structure

```
specledger/
├── spec.md              # Main feature specification
├── plan.md              # Implementation plan (optional)
├── tasks.md             # Task breakdown (optional)
├── contracts/           # Interface contracts and agreements
└── memory/              # Agent memory and context
```

## Using SpecLedger Commands

This project uses the unified SpecLedger CLI (`sl`) for specification management:

### Dependency Management

```bash
# List specification dependencies
sl deps list

# Add a new dependency
sl deps add git@github.com:org/spec.git main specs/sdd.md

# Resolve dependencies and generate lockfile
sl deps resolve

# Update dependencies
sl deps update

# Remove a dependency
sl deps remove git@github.com:org/spec.git specs/sdd.md
```

### Reference Validation

```bash
# Validate external spec references
sl refs validate

# List all references
sl refs list
```

### Conflict Detection

```bash
# Check for dependency conflicts
sl conflict check

# Detect potential conflicts
sl conflict detect
```

## Writing Specifications

1. **Start with spec.md**: Define your feature requirements, user stories, and acceptance criteria
2. **Add contracts**: Define API contracts and interfaces in `contracts/`
3. **Break down tasks**: Use `sl tasks` or manually create `tasks.md` for implementation tracking
4. **Link dependencies**: Use `sl deps add` to reference external specifications

## Best Practices

- Keep specifications focused and atomic
- Use concrete acceptance criteria
- Link related specs as dependencies
- Update specs as requirements evolve
- Commit spec changes alongside code changes

## Getting Help

```bash
sl --help
sl deps --help
sl <command> --help
```
