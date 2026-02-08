# Contributor Setup

Setup your development environment for contributing to SpecLedger.

## Prerequisites

- Go 1.24+ installed
- Git configured
- GitHub account
- Editor configured for Go (VS Code, GoLand, etc.)

## Fork and Clone

```bash
# Fork the repository at https://github.com/specledger/specledger

# Clone your fork
git clone https://github.com/YOUR_USERNAME/specledger.git
cd specledger

# Add upstream remote
git remote add upstream https://github.com/specledger/specledger.git
```

## Install Dependencies

```bash
# Install development dependencies
go mod download

# Verify installation
go build ./cmd
```

## Development Tools

### Required Tools

- **Go 1.24+**: Core language
- **git**: Version control
- **golangci-lint**: Linting (auto-installed via make)
- **gofmt**: Formatting (built into Go)

### Optional Tools

- **mise**: Tool installation manager
- **bd (Beads)**: Task tracking
- **perles**: AI agent integration

## Makefile Targets

```bash
make              # Show all available targets
make build        # Build the CLI
make install       # Install to $GOPATH/bin
make test         # Run tests
make test-coverage # Run tests with coverage
make fmt          # Format code
make lint         # Run linters
make clean        # Clean build artifacts
```

## Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test
go test -v ./pkg/cli/...
```

## Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Check formatting
gofmt -l .
```

## Development Workflow

1. Create a feature branch from `main`
2. Make your changes
3. Run tests and linting
4. Commit with clear messages
5. Push to your fork
6. Create a pull request

## Configuration

### golangci-lint

Configuration in `.golangci.yml`:
- 8 enabled linters
- 5-minute timeout
- Tests included

### CI/CD

GitHub Actions automatically runs on each pull request:
- Tests
- Linting
- Formatting check
- Coverage upload

## Next Steps

- See [Architecture](architecture.md) for project structure
- See [Development](development.md) for coding conventions
- See [Testing](testing.md) for test strategy
