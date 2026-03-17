# SpecLedger Test Strategy

> **Governing principles:** See the [Project Constitution](../.specledger/memory/constitution.md) — specifically Principles VI (Contract-First Testing), VII (Supabase Local Stack), and VIII (Quickstart-Driven Validation).
>
> **Layer-by-layer design:** See [docs/design/testing.md](../docs/design/testing.md) — maps the 4-layer architecture (L0-L3) to the 3 test tiers.

This document describes the current testing implementation, tooling, and conventions. It is the source of truth for _how_ we test; the constitution is the source of truth for _why_; the design doc specifies _what_ each layer tests.

---

## Testing Tiers

### Tier 1: Unit Tests

**Location:** `pkg/**/`, `internal/**/`, `tests/issues/`
**Run with:** `make test` or `go test ./pkg/... ./internal/... ./cmd/...`
**Build tag:** none (always runs)

Fast, isolated tests for business logic. No network, no Docker, no filesystem side effects beyond `t.TempDir()`.

**Patterns in use:**
- Table-driven tests for validation logic (`tests/issues/issue_test.go`)
- `httptest.NewServer()` for mocking Supabase/pgREST responses (`pkg/cli/comment/client_test.go`)
- Mock interfaces for auth and credentials (`mockAuthProvider`)
- Temp directories for JSONL store tests (`tests/issues/store_test.go`)

**Planned additions:**
- [ ] **go-vcr cassettes** — Record real pgREST request/response pairs into `testdata/cassettes/`. Replay them for deterministic HTTP tests without a running Supabase instance. Library: [dnaeon/go-vcr](https://github.com/dnaeon/go-vcr)
- [ ] **OpenAPI contract snapshots** — Fetch the auto-generated OpenAPI spec from PostgREST root endpoint (`GET /rest/v1/`), commit as `testdata/openapi.json`. CI step diffs against live spec to detect schema drift after migrations.

### Tier 2: Integration Tests

**Location:** `tests/integration/`
**Run with:** `make test-integration`
**Build tag:** `//go:build integration`

Tests that require external tools or a running Supabase local stack. These validate real behavior, not mocks.

**Patterns in use:**
- Real `sl` binary built at test time via `go build` (`test_helper.go`)
- Binary invocation via `exec.Command()` — tests the actual CLI, not cobra internals
- Isolated temp directories per test
- Output and exit code assertions

**Current coverage:**
| Area | File | What it tests |
|------|------|---------------|
| Bootstrap | `bootstrap_test.go` | `sl new`, `sl init`, project scaffolding, playbook application |
| Dependencies | `deps_test.go` | `sl deps add/list/remove`, duplicate detection, project root discovery |
| Doctor | `doctor_test.go` | `sl doctor`, JSON output, tool detection, exit codes |

**Planned additions:**
- [ ] **Supabase local stack tests** — Use `supabase start` in `TestMain` to spin up the full local stack (Postgres, PostgREST, GoTrue). Point integration tests at `localhost:54321`. Validate:
  - Schema migrations apply cleanly
  - RLS policies enforce expected access patterns
  - pgREST endpoints return expected shapes
  - Auth flows work end-to-end against GoTrue
- [ ] **Contract validation** — After `supabase start`, fetch the live OpenAPI spec and diff against committed `testdata/openapi.json`. Fail if they diverge.
- [ ] **Supabase data branching in CI** — Use Supabase branching to create isolated DB environments per feature branch. Ensures migration changes are tested in isolation.

### Tier 3: E2E Tests

**Location:** `tests/e2e/` (to be created)
**Run with:** `make test-e2e`
**Build tag:** `//go:build e2e`

Full CLI-to-backend validation. These invoke the real `sl` binary against a real local Supabase stack. No mocks, no fakes. Orchestrated entirely via Go's `testing` package.

**Design:**
- `TestMain` starts `supabase start`, builds `sl` binary, seeds test data
- Each test function maps to a quickstart.md scenario (see [Quickstart-Driven Validation](#quickstart-driven-validation))
- Tests run CLI commands via `exec.Command()` and assert on:
  - stdout/stderr output
  - Exit codes
  - Database state (queried via pgREST or direct SQL)
  - File system artifacts (JSONL files, config files)
- `TestMain` tears down the Supabase stack after all tests complete

**Planned additions:**
- [ ] Create `tests/e2e/` directory structure
- [ ] `TestMain` with Supabase lifecycle management
- [ ] Golden file pattern for CLI output validation (consider [go-cmdtest](https://github.com/google/go-cmdtest) or [testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript))
- [ ] Test generation from quickstart.md scenarios

---

## Quickstart-Driven Validation

> **Principle VIII:** The `quickstart.md` generated during planning defines user scenarios that map 1:1 to E2E test cases.

Each spec's `quickstart.md` (at `specledger/<spec>/quickstart.md`) contains the CLI commands and expected behavior a user would follow. These are not just documentation — they are the specification for E2E tests.

**The mapping:**

```
Spec user stories / FRs
        ↓
  quickstart.md scenarios (written during /specledger.plan)
        ↓
  Go e2e test functions (written during /specledger.implement)
```

**Convention:**
- Each quickstart scenario gets a corresponding `Test_<Spec>_<Scenario>` function in `tests/e2e/`
- Test names reference the spec short code (e.g., `Test_009_LoginAndComment`)
- If a quickstart scenario has no test → it's a gap to be filed as an issue
- Task lists generated by `/specledger.tasks` must include a task for translating quickstart scenarios to e2e tests

---

## Supabase Local Stack

> **Principle VII:** Every feature branch maintains a working local Supabase environment.

**Setup:** `supabase start` spins up the full stack via Docker Compose:
- PostgreSQL (with migrations applied)
- PostgREST (pgREST API)
- GoTrue (auth)
- Storage, Realtime, etc.

**Requirements:**
- Docker must be running
- `supabase` CLI installed (checked by `sl doctor`)
- Migrations in `supabase/migrations/` are applied on start

**Rules:**
- If you change a migration, RLS policy, or edge function → update and validate the local stack
- CI runs `supabase start` → integration/e2e tests → `supabase stop`
- Supabase data branching isolates feature work from main

---

## Contract Snapshots

> **Principle VI:** Contracts are snapshotted on disk and validated in tests.

**Planned file locations:**

```
testdata/
├── openapi.json              # PostgREST OpenAPI spec snapshot
├── cassettes/                # go-vcr recorded HTTP interactions
│   ├── comment_create.yaml
│   ├── comment_list.yaml
│   └── ...
└── fixtures/                 # Seed data for e2e tests
    ├── projects.json
    ├── specs.json
    └── ...
```

**Workflow:**
1. Run `supabase start` locally
2. Fetch OpenAPI spec: `curl http://localhost:54321/rest/v1/ > testdata/openapi.json`
3. Record cassettes: run tests in record mode
4. Commit snapshots — they are part of the codebase
5. CI diffs live spec against snapshot — fails on drift

---

## Running Tests

```bash
# Unit tests (fast, no dependencies)
make test

# Integration tests (requires external tools, may require Docker)
make test-integration

# E2E tests (requires Docker + supabase CLI)
make test-e2e          # (planned)

# All tests with coverage
make test-coverage
```

---

## Adding New Tests

1. **Unit test for new logic?** → Add `_test.go` alongside the package code
2. **Testing pgREST interaction?** → Add a go-vcr cassette in `testdata/cassettes/` and a unit test
3. **Testing CLI command end-to-end?** → Add to `tests/integration/` with `//go:build integration`
4. **Testing a quickstart scenario?** → Add to `tests/e2e/` with `//go:build e2e`, name it after the spec
5. **Changed a migration or RLS policy?** → Update `testdata/openapi.json`, add/update integration test validating the change

---

## References

- [Project Constitution](../.specledger/memory/constitution.md) — Governing principles (VI, VII, VIII)
- [go-vcr](https://github.com/dnaeon/go-vcr) — HTTP record/replay for Go
- [PostgREST OpenAPI](https://docs.postgrest.org/en/v12/references/api/openapi.html) — Auto-generated API spec
- [testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript) — CLI golden-file testing
- [go-cmdtest](https://github.com/google/go-cmdtest) — CLI output testing with golden files
- [Supabase Local Development](https://supabase.com/docs/guides/local-development) — Local stack setup
- [Supabase Branching](https://supabase.com/docs/guides/deployment/branching) — Data branching for feature isolation
