# Implementation Plan: Skills Registry Integration

**Branch**: `610-skills-registry` | **Date**: 2026-04-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specledger/610-skills-registry/spec.md`

## Summary

Add `sl skill` subcommand (singular, per CLI naming convention) with 6 subcommands (search, info, add, remove, list, audit) that integrate with Vercel's skills.sh registry. Implemented as a native Go HTTP client against 4 public APIs (skills.sh search, audit, telemetry, GitHub raw content/trees). Skills are installed to agent-specific directories resolved from the agent registry, tracked in a Vercel-compatible `skills-lock.json`.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), net/http (API client), crypto/sha256 (lock file hashing), gopkg.in/yaml.v3 (SKILL.md frontmatter), dnaeon/go-vcr v4 (test only — VCR cassettes)
**Storage**: `skills-lock.json` (Vercel-compatible local lock file, project root)
**Testing**: Two-tier: (1) `dnaeon/go-vcr` v4 cassettes for API client unit tests, (2) `httptest.Server` for full CLI E2E integration tests. Endpoint base URLs configurable via ENV vars (`SKILLS_API_URL`, `SKILLS_AUDIT_URL`, `GITHUB_API_URL`) for testability.
**Target Platform**: darwin/linux (CLI binary via GoReleaser)
**Project Type**: Single project (Go module)
**Performance Goals**: All commands complete in <2s excluding network RTT
**Constraints**: No Node.js dependency. Public APIs only (no auth). Fire-and-forget telemetry.
**Scale/Scope**: 6 subcommands, ~1200 LOC estimated across client + commands + lock file

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Specification-First**: Spec.md complete with 6 prioritized user stories, 15 FRs, 7 SCs
- [x] **Test-First**: Test strategy defined — unit tests for client/parser, integration tests for CLI flow (see research.md Decision 9)
- [x] **Code Quality**: `golangci-lint v2` (existing), `make fmt` (existing)
- [x] **UX Consistency**: All 6 commands follow CLI design principles (compact output, footer hints, `--json`, stderr errors)
- [x] **Performance**: <2s per command (excluding network). Audit fetched in parallel during add (3s timeout)
- [x] **Observability**: Errors to stderr with actionable guidance per Principle 2. `--verbose` flag planned for debug output
- [x] **Issue Tracking**: Epic to be created with `sl issue create --type epic` during task generation phase

**Complexity Violations**: None identified. All decisions follow YAGNI/KISS.

## Project Structure

### Documentation (this feature)

```text
specledger/610-skills-registry/
├── plan.md              # This file
├── research.md          # Phase 0: decisions and rationale
├── data-model.md        # Phase 1: entities and relationships
├── quickstart.md        # Phase 1: E2E test scenarios
├── contracts/
│   └── skills-sh-api.md # Phase 1: API contracts for all endpoints
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (repository root)

```text
pkg/cli/skills/
├── client.go            # HTTP client for skills.sh APIs (search, audit, telemetry)
├── client_test.go       # Unit tests with httptest mock server
├── source.go            # Source identifier parsing (owner/repo@skill)
├── source_test.go       # Table-driven parser tests
├── lock.go              # skills-lock.json read/write (Vercel-compatible)
├── lock_test.go         # Lock file serialization/hash tests
├── hash.go              # SHA-256 folder hash computation
├── hash_test.go         # Hash determinism tests
├── discover.go          # GitHub Trees API skill discovery
├── discover_test.go     # Discovery with mock GitHub API
├── install.go           # SKILL.md download + write to agent paths
├── install_test.go      # Installation path resolution tests
├── telemetry.go         # Fire-and-forget telemetry ping
└── telemetry_test.go    # Telemetry gating tests (env vars, CI, private repo)

pkg/cli/commands/
└── skill.go             # Cobra command definitions (search, info, add, remove, list, audit)

pkg/embedded/templates/specledger/skills/
└── sl-skill/
    └── SKILL.md         # Agent skill template for sl skill commands

tests/integration/
└── skills_test.go       # Full CLI integration tests (httptest-based E2E)

tests/testdata/cassettes/skills/
├── search.yaml          # VCR cassette: skills.sh search API
├── audit.yaml           # VCR cassette: audit API (ATH, Socket, Snyk)
├── github_trees.yaml    # VCR cassette: GitHub Trees API (skill discovery)
└── github_raw.yaml      # VCR cassette: raw.githubusercontent.com (SKILL.md fetch)
```

**Structure Decision**: Follows existing pattern — `pkg/cli/skills/` for business logic (matching `pkg/cli/comment/`, `pkg/cli/spec/`), `pkg/cli/commands/skills.go` for Cobra wiring (matching `comment.go`, `deps.go`). No new top-level directories.

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│  pkg/cli/commands/skills.go                             │
│  Cobra commands: search, info, add, remove, list, audit │
│  Presentation: human (compact) vs JSON (complete)       │
├─────────────────────────────────────────────────────────┤
│  pkg/cli/skills/                                        │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐  │
│  │ client   │ │ source   │ │ lock     │ │ install   │  │
│  │ .go      │ │ .go      │ │ .go      │ │ .go       │  │
│  │          │ │          │ │          │ │           │  │
│  │ Search() │ │ Parse()  │ │ Read()   │ │ Install() │  │
│  │ Audit()  │ │          │ │ Write()  │ │ Remove()  │  │
│  │ Track()  │ │          │ │ Add()    │ │           │  │
│  │ Info()   │ │          │ │ Remove() │ │           │  │
│  └────┬─────┘ └──────────┘ └────┬─────┘ └─────┬─────┘  │
│       │                         │              │        │
│  ┌────┴─────┐ ┌──────────┐ ┌───┴──────┐ ┌─────┴─────┐  │
│  │ discover │ │ hash     │ │telemetry │ │ agent     │  │
│  │ .go      │ │ .go      │ │ .go      │ │ registry  │  │
│  │          │ │          │ │          │ │ (existing)│  │
│  │ Trees()  │ │Compute() │ │ Track()  │ │ All()    │  │
│  └────┬─────┘ └──────────┘ └────┬─────┘ └───────────┘  │
├───────┼──────────────────────────┼──────────────────────┤
│       ▼                          ▼                      │
│  External APIs                                          │
│  • skills.sh/api/search                                 │
│  • add-skill.vercel.sh/audit                            │
│  • add-skill.vercel.sh/t (telemetry)                    │
│  • raw.githubusercontent.com (SKILL.md content)         │
│  • api.github.com/repos/.../git/trees (discovery)       │
└─────────────────────────────────────────────────────────┘
```

### Data Flow: `sl skill add owner/repo@skill-name`

```
1. Parse source → SkillSource{Owner, Repo, SkillFilter}
2. Discover skills via GitHub Trees API (or direct fetch if filter specified)
3. Fetch SKILL.md content from raw.githubusercontent.com
4. Parse YAML frontmatter → SkillMetadata{Name, Description}
5. Fetch audit data in parallel (3s timeout, non-blocking)
6. Display audit table (if available)
7. Prompt for confirmation (unless --yes)
8. Resolve agent paths from registry + constitution
9. Write SKILL.md to each agent's skills directory
10. Compute SHA-256 hash of installed folder
11. Update skills-lock.json with new entry
12. Fire telemetry ping (goroutine, fire-and-forget)
```

## Implementation Phases

### Phase 1: Core Client & Source Parser (P1 foundation)

**Files**: `source.go`, `source_test.go`, `client.go`, `client_test.go`

1. Implement `ParseSource(input string) (*SkillSource, error)`:
   - `owner/repo[@skill]` shorthand → delegate to `cligit.ParseRepoFlag()` (reuse from `pkg/cli/git/git.go:77-84`), set type=github
   - Full git URLs (HTTPS/SSH/GitLab) → delegate to `cligit.ParseRepoURL()` (from `git.go:54-65`), set type=git
   - Returns `SkillSource{Owner, Repo, SkillFilter, Ref, Type, URL}` where Type is `github` or `git`
2. Implement `Client` struct with configurable base URLs (`SKILLS_API_URL`, `SKILLS_AUDIT_URL`, `GITHUB_API_URL` env vars for testability)
3. Client methods: `Search(query, limit)`, `FetchAudit(source, skills)`, `FetchSkillContent(owner, repo, ref, skillPath)`
4. Unit tests with `dnaeon/go-vcr` v4 cassettes — record real API responses, replay in CI
5. Cassettes stored at `tests/testdata/cassettes/skills/`

### Phase 2: Lock File & Hashing (P1 foundation)

**Files**: `lock.go`, `lock_test.go`, `hash.go`, `hash_test.go`

1. Implement `ReadLocalLock(path)` / `WriteLocalLock(path, lock)` matching Vercel schema
2. Implement `ComputeFolderHash(dir)` — SHA-256 of sorted file paths + contents
3. Implement `AddSkill(lock, name, entry)` / `RemoveSkill(lock, name)`
4. Tests for serialization, deterministic hashing, empty dir, nested files

### Phase 3: Skill Discovery & Installation (P1 core)

**Files**: `discover.go`, `discover_test.go`, `install.go`, `install_test.go`

1. Implement `DiscoverSkills(source)` with two paths and fallback:
   - **GitHub fast path** (type=github): GitHub Trees API to enumerate skills, raw.githubusercontent.com to fetch SKILL.md content
   - **Git clone fallback** (type=git, OR github fast path fails with 404): `git clone --depth 1` to temp dir, scan for SKILL.md files, clean up
   - If `owner/repo` shorthand fails GitHub API (404), automatically retry via `git clone https://github.com/{owner}/{repo}` before erroring — handles repos without Trees API access
2. Implement `InstallSkill(metadata, content, agentPaths)` — write SKILL.md to resolved paths
3. Implement `RemoveSkill(name, agentPaths, lockPath)` — delete dirs + update lock
4. Agent path resolution: read constitution → look up ConfigDir from registry → build paths

### Phase 4: Telemetry (P1)

**Files**: `telemetry.go`, `telemetry_test.go`

1. Implement `Track(event, params)` — fire-and-forget GET with 3s timeout
2. Gating: check env vars, CI detection, private repo check
3. Tests for gating logic (no network calls in tests)

### Phase 5: Cobra Commands (P1 + P2 + P3)

**Files**: `pkg/cli/commands/skill.go`

1. Register `VarSkillCmd` with 6 subcommands in `cmd/sl/main.go`
2. Implement each run function following Two-Level Output Design:
   - `runSkillSearch` — compact table + footer hint (human) / JSON array (json)
   - `runSkillAdd` — audit display + confirmation + install + telemetry
   - `runSkillInfo` — metadata + audit table (human) / full JSON (json)
   - `runSkillList` — compact list + footer (human) / lock file JSON (json)
   - `runSkillRemove` — delete + update lock + confirmation
   - `runSkillAudit` — batch audit table + warning summary
3. All commands: `--json` flag, `--help` with 2+ examples, errors to stderr with suggested fix

### Phase 6: Agent Skill Template

**Files**: `pkg/embedded/templates/specledger/skills/sl-skill/SKILL.md`

1. Create `sl-skill` embedded skill template following existing pattern (sl-comment, sl-deps, sl-audit) (ensure to load the skill-creator skill while working on this)
2. Skill teaches AI agents when/how to use `sl skill` commands
3. Include trigger conditions, JSON output examples, security audit interpretation

### Phase 7: Integration Tests & Polish

**Files**: `tests/integration/skills_test.go`, `tests/testdata/cassettes/skills/`

1. **VCR cassettes**: Record real skills.sh, audit, and GitHub API responses for deterministic replay
2. **E2E tests**: Build `sl` binary, run quickstart scenarios with `SKILLS_API_URL` pointing to `httptest.Server`
3. Verify human output format (compact, footer hints)
4. Verify JSON output is valid and complete
5. Verify error messages follow CLI design principles
6. Verify lock file interoperability (write → read back → fields match)
7. Pattern compliance check: Data CRUD pattern constraints

## Key Design Decisions (from research spikes)

### Command Naming: `sl skill` (singular)
All 16 existing top-level commands use singular naming (`sl comment`, `sl issue`, `sl spec`). `sl deps` is the sole exception. We follow the convention with `sl skill`.

### Source Parsing: Reuse `ParseRepoFlag()` (DRY)
`pkg/cli/git/git.go:77-84` provides `ParseRepoFlag(flag string) (owner, repo, error)` with 16+ test cases. We compose it in `source.go` with skills-specific `@skill-name` suffix parsing. No duplication, minimal coupling (pure function).

### Testing: Two-tier VCR + httptest (adopted from tfc-cli)
- **Tier 1**: `dnaeon/go-vcr` v4 cassettes for API client unit tests (deterministic, no network)
- **Tier 2**: `httptest.Server` for full CLI E2E (dynamic responses, full binary invocation)
- **ENV vars**: `SKILLS_API_URL`, `SKILLS_AUDIT_URL`, `GITHUB_API_URL` for endpoint injection
- Not a conflict with Constitution Principle VII (Supabase stack) — `sl skill` has zero Supabase interaction

### Agent Skill Template: `sl-skill`
New embedded skill at `pkg/embedded/templates/specledger/skills/sl-skill/SKILL.md` ensures AI agents can discover and use `sl skill` commands when users ask about finding/installing skills.

## Complexity Tracking

No violations. All decisions follow YAGNI/KISS:
- No symlink complexity (deferred to #164)
- No global lock file (only needed for check/update, out of scope)
- No interactive TUI (out of scope)
- No local path sources (v1 = remote only)
- No well-known RFC 8615 sources (niche, can add later)
- No auth required (all public APIs)
- Reuse existing `ParseRepoFlag()` instead of duplicating URL parsing
