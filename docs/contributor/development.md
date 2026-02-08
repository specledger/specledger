# Development Guide

Coding conventions and development workflows for SpecLedger.

## Code Conventions

### Go Style Guide

Follow standard Go conventions:
- Use `gofmt` for formatting
- Use `golangci-lint` for linting (configured in `.golangci.yml`)
- Follow effective Go guidelines: https://go.dev/doc/effective_go

### Naming Conventions

- **Commands**: `PascalCase` for exported, `camelCase` for internal
- **Variables**: `camelCase`
- **Constants**: `PascalCase` or `UPPER_SNAKE_CASE`
- **Interfaces**: `PascalCase` with `er` suffix for single-method interfaces
- **Files**: `snake_case.go`

### File Organization

```
commands/
├── bootstrap.go       # Main bootstrap command
├── deps.go            # Dependencies command
├── doctor.go          # Diagnostics command
└── ...               # Other commands
```

One command per file, matching the command name.

## Error Handling

### Use structured errors

```go
type Error struct {
    Code    string
    Message string
    Err     error
}

func (e *Error) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
    return e.Err
}
```

### Error wrapping

```go
if err != nil {
    return &Error{
        Code:    "ERR_PROJECT_EXISTS",
        Message: fmt.Sprintf("project %s already exists", name),
        Err:     err,
    }
}
```

## Logging

Use the logger from `pkg/cli/logger`:

```go
logger.Debug("Checking tool status", "tool", toolName)
logger.Info("Creating project", "name", projectName)
logger.Warn("Tool not found", "tool", toolName)
logger.Error("Failed to create project", "error", err)
```

## Testing

### Test Organization

```go
func TestNewCommand(t *testing.T) {
    tests := []struct {
        name    string
        flags   []string
        wantErr bool
    }{
        {
            name:    "valid project creation",
            flags:   []string{"--ci", "--project-name", "test", "--short-code", "tst"},
            wantErr: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Test Coverage

Target: 70%+ coverage

```bash
make test-coverage
```

## Adding New Commands

1. Create new file in `pkg/cli/commands/`
2. Implement command using Cobra
3. Register in `cmd/main.go`
4. Add tests
5. Update documentation

Example:

```go
package commands

import (
    "github.com/spf13/cobra"
)

var VarMyCommand = &cobra.Command{
    Use:   "mycommand",
    Short: "Brief description",
    Long:  `Long description with examples.`,
    RunE:  runMyCommand,
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

## Commit Messages

Follow conventional commits:

```
feat: add new feature
fix: resolve bug in dependency resolution
docs: update installation guide
refactor: improve code structure
test: add tests for dependency resolution
chore: update dependencies
```

## Pull Request Process

1. Fork the repository
2. Create a feature branch: `git checkout -b my-feature`
3. Make your changes
4. Run `make fmt && make lint && make test`
5. Commit with conventional commit message
6. Push to your fork
7. Create pull request to `specledger:main`

## Code Review Checklist

- [ ] Tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Formatting applied (`make fmt`)
- [ ] Documentation updated
- [ ] Changelog updated (if applicable)

## Performance Considerations

- Lazy-load dependencies only when needed
- Cache remote operations where possible
- Use concurrent operations for independent tasks
- Avoid unnecessary file I/O operations

## Security Considerations

- Validate all user inputs
- Use `exec.Command` carefully with user-provided arguments
- Check file permissions before operations
- Don't expose sensitive data in logs
- Follow Go security best practices
