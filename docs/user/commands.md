# SpecLedger Commands Reference

Complete reference for all SpecLedger CLI commands.

## Project Commands

### `sl new`
Create a new project.

```bash
sl new                    # Interactive TUI mode
sl new --ci              # Non-interactive mode
sl new --ci --project-name <name> --short-code <code>
```

### `sl init`
Initialize SpecLedger in an existing repository.

```bash
sl init
sl init --framework <name>    # speckit, openspec
```

## Diagnostic Commands

### `sl doctor`
Check installation status of required tools.

```bash
sl doctor                # Human-readable output
sl doctor --json         # JSON for CI/CD
```

## Dependency Commands

### `sl deps list`
List all dependencies.

```bash
sl deps list
```

### `sl deps add`
Add a dependency.

```bash
sl deps add <url>
sl deps add <url> --alias <name>
sl deps add git@github.com:org/spec --alias api
```

### `sl deps remove`
Remove a dependency.

```bash
sl deps remove <url>
```

### `sl deps resolve`
Download and cache dependencies.

```bash
sl deps resolve
```

### `sl deps update`
Update dependencies to latest versions.

```bash
sl deps update
```

## Framework Commands

### `sl playbook`
Run framework-specific workflows.

```bash
sl playbook              # List available workflows
sl playbook <name>      # Run specific workflow
```

### `sl graph`
Show dependency graph.

```bash
sl graph
```

### `sl refs`
Manage reference resolution.

```bash
sl refs list
sl refs resolve
```

## Migration Commands

### `sl migrate`
Convert legacy configuration files.

```bash
sl migrate
sl migrate --dry-run
```

## Vendor Commands

### `sl vendor list`
List vendored dependencies.

```bash
sl vendor list
```

### `sl vendor add`
Add a vendored dependency.

```bash
sl vendor add <url>
```

### `sl vendor remove`
Remove a vendored dependency.

```bash
sl vendor remove <url>
```

## Getting Help

```bash
sl --help
sl <command> --help
```
