# Testing Strategy

Test approach and coverage requirements for SpecLedger.

## Test Organization

```
tests/
├── unit/           # Unit tests for individual packages
├── integration/    # Integration tests for command workflows
└── e2e/           # End-to-end tests (if needed)
```

## Unit Tests

### What to Test

- **Command execution**: Flags, arguments, and outputs
- **Configuration parsing**: YAML validation and defaults
- **Dependency resolution**: URL parsing, framework detection
- **Tool checking**: Tool detection and validation
- **Metadata operations**: CRUD operations on specledger.yaml

### Example

```go
func TestDepsAdd(t *testing.T) {
    // Setup test project
    tmpDir := t.TempDir()
    // ... create test project

    // Add dependency
    err := depsAdd(context.Background(), tmpDir, "git@github.com:org/spec")

    // Assert
    assert.NoError(t, err)
    assert.FileExists(t, filepath.Join(tmpDir, "specledger", "dependencies.yaml"))
}
```

## Integration Tests

### What to Test

- **Complete workflows**: `sl new`, `sl init`, `sl deps add`
- **Multi-step operations**: Dependency resolution, caching
- **Framework integration**: Spec Kit, OpenSpec setup
- **Tool installation**: Prerequisite checking

### Example

```go
func TestNewProjectWorkflow(t *testing.T) {
    // Run: sl new --ci --project-name test --short-code tst
    cmd := exec.Command(slBin, "new", "--ci", "--project-name", "test", "--short-code", "tst")
    output, err := cmd.CombinedOutput()

    assert.NoError(t, err)
    assert.Contains(t, string(output), "Project created successfully")
}
```

## Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test -v ./pkg/cli/commands

# Run specific test
go test -v ./pkg/cli/commands -run TestDepsAdd
```

## Coverage Goals

- **Target**: 70%+ coverage
- **Critical paths**: 90%+ coverage
- **Main packages**: 80%+ coverage

## Test Helpers

### Temporary Directories

Use `t.TempDir()` for test isolation:

```go
func TestSomething(t *testing.T) {
    tmpDir := t.TempDir()
    // tmpDir is automatically cleaned up
}
```

### Mock Filesystem

Use testing/mocks or create test fixtures:

```go
func TestWithMockYAML(t *testing.T) {
    yamlContent := `
project:
  name: test
  short_code: tst
`
    // Create test file with yamlContent
    // Run test
}
```

## Continuous Integration

Tests run automatically on:
- Every push to `main` branch
- Every pull request
- Every tag (for releases)

### CI Workflow

```yaml
jobs:
  test:
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Run tests
        run: make test
      - name: Run coverage
        run: make test-coverage
```

## Writing Good Tests

1. **Test behavior, not implementation**
2. **Use table-driven tests** for multiple scenarios
3. **Use descriptive test names**
4. **Assert on errors, not on success**
5. **Clean up resources** (defer cleanup, t.TempDir())

## Test Maintenance

- Update tests when functionality changes
- Remove tests for deprecated features
- Keep tests fast (<100ms per test ideally)
- Use test fixtures for complex setup

## Next Steps

- See [Development](development.md) for coding conventions
- See [Architecture](architecture.md) for project structure
