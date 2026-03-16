# Testing Design — Layer-by-Layer Strategy

> **Governing principles:** [Project Constitution](../../.specledger/memory/constitution.md) — Principles VI (Contract-First Testing), VII (Supabase Local Stack), VIII (Quickstart-Driven Validation).
>
> **Implementation details:** [tests/README.md](../../tests/README.md) — Current tooling, conventions, and test inventory.

This document bridges the 4-layer architecture (defined in [README.md](README.md)) with the 3-tier testing strategy (defined in the constitution). It specifies what to test at each layer and which tier handles it.

---

## Layer ↔ Tier Mapping

| Layer | What to test | Tier 1 (Unit) | Tier 2 (Integration) | Tier 3 (E2E) |
|-------|-------------|---------------|----------------------|---------------|
| **L0 Hooks** | Silent resilience, cache-then-retry, no UX blocking | Mock filesystem + clock | Hook fires on real git commit | Full session capture flow |
| **L1 CLI** | Flag parsing, output format, error messages, offline behavior | `httptest` + go-vcr cassettes | Real `supabase start` stack | Binary invocation via `exec.Command` |
| **L2 Commands** | Workflow sequencing, L2→L1 calls, prerequisite validation | Template rendering, prompt assembly | Command runs against local stack | Quickstart scenario replay |
| **L3 Skills** | Skill loading triggers, content accuracy, cross-layer sync | Skill file parsing, section validation | Skill-guided CLI invocations | Agent session with skill context |

---

## L0: Hooks

**What hooks do:** Invisible, event-driven automation (e.g., session capture on `git commit`).

**Testing approach:**
- **Unit:** Test the hook handler logic in isolation. Mock filesystem for state files (`~/.specledger/session-state.json`), mock clock for timing, mock HTTP for API calls. Verify:
  - Hook detects correct trigger events (e.g., `git commit` in PostToolUse)
  - Cache-then-retry logic works when API is unavailable
  - Hook never blocks or delays the user's operation
  - Error logging writes to `~/.specledger/capture-errors.log`
- **Integration:** Install hook config, run a real `git commit`, verify session state was captured.
- **E2E:** Full workflow: commit → hook fires → session captured → verify via `sl session` CLI.

**Key contract:** Hooks must be silent. A test that causes user-visible output from a hook is a failing test.

---

## L1: CLI (`sl` binary)

**What the CLI does:** Deterministic data operations — CRUD on issues, comments, deps, config.

**Testing approach:**

### pgREST Contract Testing (API-backed commands)

Commands that call Supabase/pgREST (e.g., `sl comment list`, `sl comment reply`, `sl comment resolve`):

1. **go-vcr cassettes (Tier 1):** Record real pgREST request/response pairs. Commit cassettes to `testdata/cassettes/`. Replay in unit tests for fast, deterministic validation of:
   - Request construction (URL, headers, query params, RLS-aware auth)
   - Response parsing (JSON → Go structs)
   - Error handling (401 retry, RLS violations, network errors)

2. **OpenAPI contract snapshot (Tier 2):** Fetch the auto-generated spec from PostgREST root endpoint after `supabase start`. Diff against committed `testdata/openapi.json`. Fail CI if they diverge — this catches migration-induced API surface changes.

3. **Live stack validation (Tier 2):** Run commands against real local Supabase. Verify:
   - Migrations apply cleanly
   - RLS policies enforce expected access (test with different auth contexts)
   - pgREST endpoints return expected shapes and status codes
   - Auth flows work end-to-end (token refresh, credential reload)

### Local-First Commands

Commands that operate on local files (e.g., `sl issue list`, `sl deps add`):

- **Unit (Tier 1):** Test with temp directories and fixture JSONL files
- **Integration (Tier 2):** Binary invocation in isolated temp project directories
- **E2E (Tier 3):** Part of quickstart scenario replay

### Output Format Compliance

All CLI output must conform to the Two-Level Output Design ([cli.md](cli.md)):

- **JSON mode:** Validate against documented schema (wrapped objects, not flat arrays)
- **Human mode:** Verify truncation rules, footer hints, compact format
- **Error mode:** Verify "what failed → why → fix command" structure

---

## L2: AI Commands

**What commands do:** Multi-step workflow orchestration. Commands call L1 tools and consume L3 skill context.

**Testing approach:**
- **Unit (Tier 1):** Test template rendering and prompt assembly. Verify:
  - Command frontmatter is valid
  - Prerequisite checks pass/fail correctly
  - L2→L1 tool calls are correctly formed
- **Integration (Tier 2):** Run commands against a local Supabase stack. Verify the full workflow pipeline:
  - `specify` creates valid spec.md
  - `clarify` loads comments from pgREST and produces clarification questions
  - `plan` generates plan.md with quickstart.md
  - `tasks` generates dependency-ordered task list
- **E2E (Tier 3):** Quickstart scenario replay (see below)

**Key contract:** The workflow pipeline order (`specify → clarify → plan → tasks → implement`) is immutable. Tests must verify that commands enforce prerequisite ordering.

---

## L3: Skills

**What skills do:** Passive context injection — teach agents how to use CLI tools.

**Testing approach:**
- **Unit (Tier 1):** Validate skill file structure:
  - All 7 required sections present (Overview, Subcommands, Decision Criteria, JSON Parsing, Workflow Patterns, Error Handling, Token Efficiency)
  - Subcommand tables match actual CLI flags and arguments
  - Decision criteria are internally consistent
- **Sync validation:** Skills must stay synchronized with CLI changes. Test that:
  - Every `sl <command>` flag documented in a skill exists in the actual Cobra command
  - Workflow patterns in skills produce valid CLI invocations
  - Error handling tables cover all known error codes

**Key contract:** Skills are complementary, not redundant. A skill must not duplicate information available via `sl <command> --help`.

---

## Quickstart-Driven E2E Testing

> **Principle VIII:** quickstart.md scenarios map 1:1 to E2E test functions.

### How it works

Each spec's `quickstart.md` (written during `/specledger.plan`) contains step-by-step CLI usage scenarios. These are the specification for E2E tests.

```
quickstart.md scenario          →  Go test function
─────────────────────────────────────────────────────
"Login and list comments"       →  Test_009_LoginAndListComments
"Reply to a review comment"     →  Test_009_ReplyToReviewComment
"Resolve with reason"           →  Test_009_ResolveWithReason
```

### Task generation requirement

When `/specledger.tasks` generates the task list for a feature, it MUST include:

1. **A task for reviewing quickstart.md** — Verify each scenario is testable as a CLI invocation
2. **A task per quickstart scenario** — Create the corresponding E2E test function
3. **A task for contract snapshots** — Update `testdata/openapi.json` and go-vcr cassettes if the feature touches pgREST

### Test structure

```go
// tests/e2e/<spec>_test.go
//go:build e2e

func TestMain(m *testing.M) {
    // Start supabase local stack
    // Build sl binary
    // Seed test data
    code := m.Run()
    // Tear down
    os.Exit(code)
}

func Test_<SpecCode>_<ScenarioName>(t *testing.T) {
    // Maps directly to a quickstart.md scenario
    // Invokes real sl binary via exec.Command
    // Asserts on stdout, stderr, exit code, and DB state
}
```

---

## Supabase Local Stack in Tests

> **Principle VII:** Every feature branch maintains a working local Supabase environment.

### Lifecycle in tests

```
TestMain setup:
  1. supabase start (or verify already running)
  2. Apply migrations (supabase db reset if needed)
  3. Seed test data via pgREST or direct SQL
  4. Build sl binary with test config pointing to localhost:54321

Test execution:
  - Each test gets isolated data (unique project/spec per test)
  - Tests clean up their own data or use transaction rollback

TestMain teardown:
  - supabase stop (only if TestMain started it)
```

### CI requirements

- Docker must be available (GitHub Actions: `services` or DinD)
- `supabase` CLI installed
- Migrations in `supabase/migrations/` applied before test run
- Contract snapshot diffed after migrations

### Data branching

For feature branches that modify schema:
1. Create a Supabase branch: `supabase branches create <feature-name>`
2. Apply feature migrations to the branch
3. Run integration/E2E tests against the branch
4. Merge branch when feature merges to main

---

## Contract Snapshot Workflow

```
Developer changes a migration
        ↓
supabase start (applies new migration)
        ↓
Fetch OpenAPI spec: curl localhost:54321/rest/v1/ > testdata/openapi.json
        ↓
Re-record affected go-vcr cassettes (run tests in record mode)
        ↓
Commit updated snapshots alongside migration
        ↓
CI validates: live spec == committed spec (no drift)
```

---

## References

- [Project Constitution](../../.specledger/memory/constitution.md) — Principles VI, VII, VIII
- [tests/README.md](../../tests/README.md) — Current implementation, tooling, running tests
- [cli.md](cli.md) — L1 output format and error design (tested in Tier 1 + 2)
- [commands.md](commands.md) — L2 workflow pipeline (tested in Tier 2 + 3)
- [hooks.md](hooks.md) — L0 silent resilience (tested in Tier 1 + 2)
- [skills.md](skills.md) — L3 skill sync requirements (tested in Tier 1)
