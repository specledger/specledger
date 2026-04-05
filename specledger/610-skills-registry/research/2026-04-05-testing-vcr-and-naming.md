# Research: Integration Testing Strategy, Command Naming, and Code Reuse

**Date**: 2026-04-05
**Context**: Pre-task-generation review of plan.md — 3 open questions from reviewer that impact task decomposition.
**Time-box**: 30 minutes

## Question 1: Integration Testing — VCR Cassettes vs httptest

### Findings

**Existing SpecLedger pattern** (Constitution Principle VI/VII): Uses Supabase local stack for integration testing. This is for database-backed commands (`sl comment`, `sl session`). Not applicable to `sl skill` which calls external public APIs with no Supabase dependency.

**Reference: tfc-cli VCR setup** (`so0k/tfc-cli`):

Uses a **two-tier strategy**:

| Tier | Tool | Where | Purpose |
|------|------|-------|---------|
| Unit | `dnaeon/go-vcr` v4 | `internal/tfc/client_test.go` | Replay recorded API responses against the HTTP client |
| E2E | `httptest.Server` | `tests/e2e/helpers_test.go` | Full CLI binary invocation against mock server |

**Key VCR patterns from tfc-cli**:
- Cassettes stored in `tests/testdata/cassettes/*.yaml` (human-readable YAML)
- Custom matcher: matches by HTTP method + URL path only (ignores headers/query params → prevents brittle tests)
- Recording opt-in: `RECORD_CASSETTES=1 go test ...` (skipped by default)
- Client receives injected `*http.Client` from VCR recorder: `rec.GetDefaultClient()`

**Endpoint override for testability**:
- tfc-cli uses ENV var: `TFC_ADDRESS=http://localhost:PORT/api/v2`
- The Node.js skills CLI already supports this: `SKILLS_API_URL` env var (find.ts line 15)
- We need **3 ENV vars** for our Go client:
  - `SKILLS_API_URL` (default: `https://skills.sh`) — search API
  - `SKILLS_AUDIT_URL` (default: `https://add-skill.vercel.sh`) — audit + telemetry
  - `GITHUB_API_URL` (default: `https://api.github.com`) — trees API + raw content

### Decision

**Adopt the tfc-cli two-tier pattern** for `sl skill`:

1. **Unit tests** (`pkg/cli/skills/client_test.go`): Use `dnaeon/go-vcr` v4 with cassettes for search, audit, and GitHub APIs. Record real responses once, replay deterministically in CI.

2. **Integration tests** (`tests/integration/skills_test.go`): Use `httptest.Server` for full CLI E2E. Build `sl` binary, run commands with `SKILLS_API_URL` pointing to mock server.

3. **Client constructor** accepts base URLs as params (not hardcoded), enabling both VCR and httptest injection.

**Rationale**: VCR cassettes give us real API response shapes without network dependency. httptest gives us full CLI integration coverage. The two tiers complement each other — VCR validates client behavior, httptest validates command wiring + output formatting.

**This is NOT a conflict with Constitution Principle VII** (Supabase local stack). That principle applies to commands that interact with Supabase. `sl skill` has zero Supabase interaction — its external dependencies are skills.sh, GitHub, and Vercel's audit API.

---

## Question 2: Command Naming — `sl skill` vs `sl skills`

### Findings

**All 17 top-level commands in `sl` use SINGULAR naming**:

| Command | Use |
|---------|-----|
| `sl comment` | SINGULAR |
| `sl issue` | SINGULAR |
| `sl spec` | SINGULAR |
| `sl config` | SINGULAR |
| `sl session` | SINGULAR |
| `sl auth` | SINGULAR |
| `sl playbook` | SINGULAR |
| `sl doctor` | SINGULAR |
| `sl context` | SINGULAR |
| `sl code` | SINGULAR |
| `sl graph` | SINGULAR |
| `sl mockup` | SINGULAR |
| `sl revise` | SINGULAR |
| `sl init` | SINGULAR |
| `sl new` | SINGULAR |
| `sl version` | SINGULAR |
| `sl deps` | **PLURAL** (only exception) |

### Decision

**Use `sl skill`** (singular) for consistency with the overwhelming convention.

`sl deps` is the sole outlier and shouldn't set the pattern. The spec and plan should be updated from `sl skills` to `sl skill` throughout.

**Impact**: All references in spec.md, plan.md, quickstart.md, and research.md need updating. The Node.js CLI uses `skills` (plural) but that's their convention, not ours.

---

## Question 3: Agent Skill Template for `sl skill` Command

### Findings

**Existing embedded skills** at `pkg/embedded/templates/specledger/skills/`:
- `sl-audit/` — teaches agents about `sl audit`
- `sl-comment/` — teaches agents about `sl comment`
- `sl-deps/` — teaches agents about `sl deps`
- `sl-issue-tracking/` — teaches agents about `sl issue`

**Pattern**: Each skill has a `SKILL.md` with YAML frontmatter (name, description) and instructions for when/how the agent should use the corresponding `sl` command.

### Decision

**Yes, we need an `sl-skill` embedded skill template.** Add it to `pkg/embedded/templates/specledger/skills/sl-skill/SKILL.md` as part of the implementation tasks.

**Content should cover**:
- When to trigger: user asks about finding, installing, or managing agent skills
- How to use: `sl skill search`, `sl skill add`, `sl skill list`, etc.
- JSON output examples for agent consumption
- Security audit interpretation guidance

---

## Question 4: Git URL Parsing — DRY vs WET

### Findings

**Existing code**: `pkg/cli/git/git.go` provides `ParseRepoURL(rawURL string) (owner, repo string, err error)` — lines 54-84.
- Handles SSH, HTTPS, ssh://, custom hosts, ports, dotted repo names
- 16+ test cases in `git_test.go:209-344`
- Already reused by `pkg/cli/session/capture.go:157-161`
- Returns exactly what we need: `(owner, repo, error)`

**Also exists**: `ParseRepoFlag(flag string) (owner, repo string, err error)` — lines 77-84.
- Parses simple `owner/repo` format (no protocol prefix)
- This is the exact format `sl skill add owner/repo@skill` uses

### Decision

**DRY — reuse `ParseRepoFlag()` from `pkg/cli/git/git.go`** for the `owner/repo` portion. Add skill-specific `@skill-name` parsing on top.

**Rationale**:
- `ParseRepoFlag` is a pure function with no side effects — zero coupling risk
- It's already tested with edge cases we'd need to cover anyway
- The "WET" alternative would duplicate 30 lines of regex matching with no benefit
- The `@skill-name` suffix parsing is skills-specific and belongs in `pkg/cli/skills/source.go`

**Implementation**:
```go
// pkg/cli/skills/source.go
func ParseSource(input string) (*SkillSource, error) {
    // 1. Split on @ to extract skill filter
    // 2. Delegate owner/repo parsing to cligit.ParseRepoFlag()
    // 3. Return SkillSource with Owner, Repo, SkillFilter, Ref
}
```

This follows the principle of composing existing utilities rather than duplicating them, while keeping skills-specific logic (the `@` syntax) in the skills package.

## References

- `so0k/tfc-cli/internal/tfc/client_test.go` — VCR cassette test setup
- `so0k/tfc-cli/internal/tfc/record_test.go` — Cassette recording logic
- `so0k/tfc-cli/tests/e2e/helpers_test.go` — E2E test helpers with httptest
- `specledger/specledger/.claude/worktrees/sl-skills/pkg/cli/git/git.go:54-84` — ParseRepoURL/ParseRepoFlag
- `specledger/specledger/.claude/worktrees/sl-skills/pkg/cli/git/git_test.go:209-344` — URL parsing tests
- `specledger/specledger/.claude/worktrees/sl-skills/cmd/sl/main.go` — Command registration
- `specledger/specledger/.claude/worktrees/sl-skills/pkg/embedded/templates/specledger/skills/` — Embedded skill templates
