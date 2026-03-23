# Contributing to SpecLedger

Thank you for considering contributing to SpecLedger!

**Required reading before contributing:** [docs/design/README.md](docs/design/README.md) — the 4-layer architecture and design contracts all contributions must respect.

---

## Two Contribution Paths

SpecLedger uses itself for contributions. Choose the path that matches the scope of your change.

### Path 1: Vibe-coded (small fixes)

Use this path for changes that are:
- Isolated to ~3 files or fewer
- Low architectural risk (bug fix, typo, small enhancement)
- No new commands, no schema changes, no cross-layer impact

**Workflow:**
1. Create a Claude Code plan (not a full spec)
2. Implement the fix
3. Validate against [`.specledger/memory/constitution.md`](.specledger/memory/constitution.md) and [`docs/design/`](docs/design/) guidelines
4. `make lint && make test && make fmt`
5. Open a PR

> The PR will be reviewed against the constitution and design docs. If the reviewer determines the change has broader impact than expected, they may request a full SDD flow.

### Path 2: Full SDD (features / large changes)

Use this path for changes that are:
- Multi-file, cross-package, or cross-layer
- New commands, new API surface, schema migrations
- Anything that affects the 4-layer model contracts

**Workflow:**
1. `/specledger.specify` — write the spec
2. `/specledger.clarify` — answer questions, resolve reviewer comments from core team
3. `/specledger.plan` — design the implementation
4. **Alignment with core team** on spec + plan (via app.specledger.io review comments)
5. `/specledger.tasks` — generate task breakdown
6. `/specledger.verify` — cross-validate artifacts
7. `/specledger.implement` — execute
8. Open a PR

---

## Pre-push Checklist

Run these before every PR:

```bash
make lint   # golangci-lint v2
make test   # unit tests
make fmt    # formatting
```

---

## Reporting Bugs

Before creating a bug report, check existing issues. Include:

- **Title**: Clear and descriptive
- **Description**: Detailed description of the problem
- **Steps to Reproduce**: Step-by-step instructions
- **Expected Behavior**: What you expected
- **Actual Behavior**: What happened
- **Environment**: OS, `sl --version`, Go version (if building from source)

---

## Development Setup

### Prerequisites

| Tool | Version | Required For | Install |
|------|---------|-------------|---------|
| **Go** | 1.24+ | Building and testing | [go.dev/dl](https://go.dev/dl/) |
| **Make** | any | Build targets (`make build`, `make test`, etc.) | Xcode CLT (macOS) or `apt install make` |
| **Git** | 2.x+ | Version control | [git-scm.com](https://git-scm.com/) |
| **Docker** | 24+ | Local Supabase stack for integration/E2E tests | [docker.com](https://www.docker.com/) |
| **Supabase CLI** | 1.x+ | Local backend stack | `brew install supabase/tap/supabase` |
| **golangci-lint** | v2+ | Linting | `mise install golangci-lint` |

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

---

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

---

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

---

## Architecture

SpecLedger uses a 4-layer tooling model. See [docs/design/README.md](docs/design/README.md) for the full architecture.

| Layer | Name | Purpose |
|-------|------|---------|
| L0 | Hooks | Invisible event-driven automation |
| L1 | `sl` CLI | Deterministic data operations |
| L2 | AI Commands | Workflow orchestration (`/specledger.*`) |
| L3 | Skills | Passive context for agent decision-making |

---

## Commit Messages

Follow [conventional commits](https://www.conventionalcommits.org/) format:

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

---

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

Open an issue with the `question` label, or start a discussion on GitHub Discussions.
