# Contributing to SpecLedger

First off, thank you for considering contributing to SpecLedger! It's people like you that make SpecLedger such a great tool.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues as you might find that the problem has already been reported. When creating a bug report, please include:

- **Title**: A clear and descriptive title
- **Description**: A detailed description of the problem
- **Steps to Reproduce**: Step-by-step instructions to reproduce the issue
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Environment**:
  - OS and version
  - SpecLedger version (`sl --version`)
  - Go version (if building from source)
- **Screenshots/Logs**: If applicable

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Use Case**: What problem does this solve?
- **Proposed Solution**: How would you like it to work?
- **Alternatives**: What other solutions have you considered?
- **Examples**: Mockups, code examples, or documentation

### Pull Requests

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with clear, descriptive commit messages
4. Add or update tests if applicable
5. Ensure the code passes all tests (`make test`)
6. Format your code (`make fmt`)
7. Commit your changes (`git commit -m 'Add some amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

## Development Setup

### Prerequisites

| Tool | Version | Required For | Install |
|------|---------|-------------|---------|
| **Go** | 1.24+ | Building and testing | [go.dev/dl](https://go.dev/dl/) |
| **Make** | any | Build targets (`make build`, `make test`, etc.) | Xcode CLT (macOS) or `apt install make` |
| **Git** | 2.x+ | Version control | [git-scm.com](https://git-scm.com/) |
| **Docker** | 24+ | Local Supabase stack for integration/E2E tests | [docker.com](https://www.docker.com/) |
| **Supabase CLI** | 1.x+ | Local backend stack | `brew install supabase/tap/supabase` |
| **golangci-lint** | 1.x+ | Linting | `brew install golangci-lint` |

### Why Docker + Supabase CLI?

SpecLedger's backend runs on Supabase (PostgreSQL + PostgREST + GoTrue). For local development and testing, we use `supabase start` to spin up the full stack in Docker containers. This ensures:

- **Schema migrations** are validated locally before pushing
- **RLS (Row Level Security) policies** are tested with real Postgres, not mocks
- **API contracts** (pgREST endpoints) are validated against the actual PostgREST instance
- **Auth flows** work against a real GoTrue instance

Without Docker/Supabase CLI, you can still build the binary and run unit tests, but integration and E2E tests will be skipped.

### Quick Setup

```bash
# Clone the repository
git clone https://github.com/specledger/specledger.git
cd specledger

# Build
make build
# Binary at bin/sl

# Run unit tests (no Docker needed)
make test

# Start local Supabase stack (requires Docker)
supabase start

# Run integration tests (requires local Supabase stack)
make test-integration

# Run E2E tests (requires local Supabase stack)
make test-e2e

# Lint
make lint

# Format
make fmt
```

### Verifying Your Setup

```bash
# Check all tools are installed
sl doctor

# Check Supabase local stack is running
supabase status

# Run the full CI check locally
make lint && make test && make test-integration
```

## Testing

SpecLedger uses a three-tier testing strategy. See [tests/README.md](tests/README.md) for full details.

| Tier | Location | Run With | Requires |
|------|----------|----------|----------|
| **Unit** | `pkg/**/*_test.go`, `tests/issues/` | `make test` | Go only |
| **Integration** | `tests/integration/` | `make test-integration` | Docker + Supabase CLI |
| **E2E** | `tests/e2e/` | `make test-e2e` | Docker + Supabase CLI |

### Testing Guidelines

- **New CLI logic?** Add unit tests alongside the package code (`*_test.go`)
- **New pgREST interaction?** Add a go-vcr cassette in `testdata/cassettes/` for unit tests, plus an integration test against the local Supabase stack
- **New CLI command?** Add an integration test in `tests/integration/` that invokes the real binary
- **Changed a migration or RLS policy?** Update `testdata/openapi.json` and add/update an integration test
- **New quickstart scenario?** Add a corresponding E2E test function

For the testing design rationale, see:
- [Project Constitution](/.specledger/memory/constitution.md) — Principles VI, VII, VIII
- [docs/design/testing.md](docs/design/testing.md) — Layer-by-layer testing strategy

## Project Structure

```
specledger/
├── cmd/sl/              # CLI entry point
├── pkg/
│   ├── cli/             # CLI commands and utilities
│   │   ├── commands/    # Cobra command implementations
│   │   ├── comment/     # Supabase API client
│   │   ├── auth/        # OAuth/token management
│   │   ├── config/      # Configuration management
│   │   ├── tui/         # Terminal UI interactions
│   │   └── ...
│   ├── issues/          # Issue tracking (JSONL store)
│   ├── deps/            # Spec dependency resolution
│   ├── models/          # Data models
│   ├── embedded/        # Embedded templates
│   └── version/         # Version info + update checking
├── internal/            # Internal packages (spec parsing, refs)
├── tests/
│   ├── integration/     # Integration tests (real binary, real stack)
│   ├── issues/          # Issue store unit tests
│   └── e2e/             # E2E tests (quickstart scenario replay)
├── docs/design/         # Architecture and design docs
├── supabase/            # Supabase config and migrations
├── scripts/             # Installation and utility scripts
├── .specledger/         # SpecLedger project config and constitution
├── Makefile             # Build, test, lint targets
├── .goreleaser.yaml     # Cross-platform release config
└── .golangci.yml        # Lint rules
```

## Architecture

SpecLedger uses a 4-layer tooling model. See [docs/design/README.md](docs/design/README.md) for the full architecture.

| Layer | Name | Purpose |
|-------|------|---------|
| L0 | Hooks | Invisible event-driven automation |
| L1 | `sl` CLI | Deterministic data operations |
| L2 | AI Commands | Workflow orchestration (`/specledger.*`) |
| L3 | Skills | Passive context for agent decision-making |

## Commit Messages

Follow conventional commits format:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Build process or auxiliary tool changes

Example:
```
feat: add interactive mode for dependency selection

This allows users to interactively select which dependencies to resolve
when running `sl deps resolve --interactive`.
```

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

Feel free to open an issue with the `question` label, or start a discussion on GitHub Discussions.
