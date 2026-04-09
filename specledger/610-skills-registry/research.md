# Phase 0 Research: Skills Registry Integration

**Date**: 2026-04-05
**Branch**: `610-skills-registry`
**Spec**: [spec.md](spec.md)

## Prior Work

| Spec | Related Feature | Relevance |
|------|----------------|-----------|
| 001-coding-agent-support | Agent registry, `internal/agent/registry.go` | Agent→ConfigDir mapping we'll use for skill installation paths |
| 597-agent-model-config | Per-agent config, `pkg/cli/config/agent_settings.go` | Config merge patterns for agent-aware settings |
| 601-cli-skills | `sl comment` CLI, `sl-comment` skill | Cobra subcommand pattern, HTTP client pattern, skill installation to `.claude/skills/` |
| 596-doctor-version-update | Template management, playbook copy/symlink | Playbook system creates `.agents/skills/` with symlinks — same pattern we'll follow |
| 135-fix-missing-chmod-x | `applyEmbeddedSkills` | File permission handling for skill installation |

## Decision 1: HTTP Client Architecture

**Decision**: Create a standalone `pkg/cli/skills/client.go` with a plain `*http.Client` — no auth required.

**Rationale**: All three skills.sh APIs (search, audit, telemetry) are public and unauthenticated. The existing `comment/client.go` pattern includes Supabase auth retry logic that isn't needed here. A simpler client without auth reduces complexity.

**Alternatives considered**:
- Reuse `comment/client.go` pattern → rejected, auth overhead for public APIs
- Use `net/http` directly in each command → rejected, duplicates timeout/error handling

## Decision 2: Agent Path Resolution

**Decision**: Use `internal/agent/registry.go` to resolve agent names to skill installation paths. Read configured agents from `specledger.yaml` (the single source of truth for agent preferences, per #147).

**Rationale**: The agent registry defines `ConfigDir` for each agent (claude→`.claude`, opencode→`.opencode`, etc.). `specledger.yaml` stores which agents the project uses (set during `sl init`). The constitution is NOT the source for agent preferences (#147).

**Path resolution logic**:
1. Read configured agents from `specledger.yaml` via metadata package
2. For each agent, look up `ConfigDir` from `internal/agent/registry.go`
3. Skill install path = `{projectRoot}/{ConfigDir}/skills/{skill-name}/SKILL.md`
4. If multi-agent: also create canonical copy at `.agents/skills/{skill-name}/` (matching playbook behavior)

**Alternatives considered**:
- Hardcode `.claude/skills/` only → rejected, breaks multi-agent projects
- Read from constitution → rejected, constitution is not the source of truth for agent preferences (#147)

## Decision 3: Lock File Format

**Decision**: Use exact Vercel `skills-lock.json` local schema (version 1).

**Schema** (from `skills/src/local-lock.ts`):
```json
{
  "version": 1,
  "skills": {
    "skill-name": {
      "source": "owner/repo",
      "ref": "main",
      "sourceType": "github",
      "computedHash": "<sha256-hex>"
    }
  }
}
```

**Rationale**: Interoperability with `npx skills` CLI. Users who install skills via either tool get a consistent lock file. The `computedHash` is SHA-256 of all files in the skill directory (sorted by relative path, each file contributes path + contents to the hash).

**Alternatives considered**:
- Custom lock format → rejected, breaks interoperability (SC-007)
- Use global lock format (v3 with timestamps) → rejected, YAGNI — global lock is only needed for `check`/`update` which are out of scope

## Decision 4: Source Identifier Format

**Decision**: Support two fetch paths:
1. `owner/repo[@skill-name]` shorthand → GitHub fast path (Trees API + raw content)
2. Full git URLs (HTTPS, SSH, GitLab, any git host) → `git clone --depth 1` fallback

**Rationale**: The skills.sh registry indexes skills from **any source** via install telemetry — not GitHub-only. GitHub dominates current coverage but GitLab, Bitbucket, and well-known HTTPS endpoints are all valid sources. The official CLI supports 5 source types. GitHub shorthand covers most public skills. Git clone covers everything else (GitLab, Bitbucket, self-hosted, any git host). Well-known (RFC 8615) and local paths are deferred — niche and development-only respectively.

**Parsing logic**:
1. Check for `@` separator → extract skill filter
2. If input matches `owner/repo` format → delegate to `cligit.ParseRepoFlag()`, set type=github
3. If input is a full URL (HTTPS/SSH) → delegate to `cligit.ParseRepoURL()`, set type=git
4. GitHub type: GitHub Trees API for discovery, raw.githubusercontent.com for SKILL.md content
5. Git type: `git clone --depth 1` to temp dir, scan for SKILL.md files, clean up

**Alternatives considered**:
- GitHub only → rejected, too limiting — skills can be hosted anywhere
- Full parity with Node.js CLI (well-known, local) → rejected, YAGNI for v1
- Git clone only (no GitHub fast path) → rejected, slower for the most common case

## Decision 5: Skill Discovery in Repos

**Decision**: Use GitHub Trees API to discover skills in a repository without cloning.

**Rationale**: The Node.js CLI uses a "blob-based fast path" for allowlisted owners that fetches the repo tree via API instead of cloning. We generalize this approach for all public repos — no git clone needed, faster, no temp directory management.

**API call**: `GET https://api.github.com/repos/{owner}/{repo}/git/trees/{ref}?recursive=1`
- Filter for paths matching `skills/*/SKILL.md` or `*/SKILL.md`
- Fetch each SKILL.md via raw GitHub content URL
- Support optional GitHub token (GITHUB_TOKEN, GH_TOKEN) for rate limit headroom

**Alternatives considered**:
- Git clone with `--depth 1` → rejected, slower, requires temp dirs, git dependency
- Raw content URL guessing → rejected, can't discover which skills exist in a repo

## Decision 6: Telemetry Implementation

**Decision**: Fire-and-forget GET request matching upstream protocol.

**Endpoint**: `GET https://add-skill.vercel.sh/t?v=specledger-{version}&event=install&source={owner/repo}&skills={name}&agents=claude-code`

**Gating logic**:
1. Skip if `DISABLE_TELEMETRY` or `DO_NOT_TRACK` env var set
2. Skip if source repo is private (check via GitHub API, skip telemetry on error)
3. Skip in CI (detect via CI, GITHUB_ACTIONS, GITLAB_CI, etc.)
4. Fire-and-forget: goroutine with 3s timeout, no error handling

**Alternatives considered**:
- POST request → rejected, upstream uses GET with query params
- Synchronous → rejected, must not block CLI

## Decision 7: Audit Data Display

**Decision**: Show all 3 partners (ATH, Socket, Snyk) in a table matching the Node.js CLI format.

**API**: `GET https://add-skill.vercel.sh/audit?source={owner/repo}&skills={slug1,slug2}`

**Display contexts**:
1. During `sl skill add` — fetched in parallel, shown before confirmation (non-blocking, 3s timeout)
2. In `sl skill info` — fetched synchronously, shown with skill details
3. In `sl skill audit` — batch query for all installed skills

**Table format** (human output):
```
                  Gen          Socket       Snyk
skill-name        Safe         0 alerts     Low Risk
other-skill       High Risk    2 alerts     Med Risk
```

## Decision 8: Command Pattern Classification

**Decision**: `sl skill` is a **Data CRUD** pattern per CLI design principles.

**Rationale**: It performs deterministic operations on skill entities (search, install, remove, list, audit). No AI reasoning involved. Returns structured data. Follows the same pattern as `sl deps` and `sl comment`.

**Constraints from Data CRUD pattern**:
- `--spec` flag for override, ContextDetector for auto-detect (needed for lock file location)
- No AI reasoning. Returns structured data.
- MUST fail with clear error + suggested fix
- `--json` flag for complete, pipeable output

## Decision 9: Testing Strategy

**Decision**: Table-driven unit tests for client/parser + integration tests for full CLI flow.

**Tiers**:
1. **Unit tests** (`pkg/cli/skills/*_test.go`): Mock HTTP responses for API client, test source parsing, lock file read/write, hash computation
2. **Integration tests** (`tests/integration/skills_test.go`): Build `sl` binary, run `sl skill search/add/remove/list` against mock server or live API
3. **No E2E/Supabase tests needed**: `sl skill` doesn't interact with Supabase at all

**Alternatives considered**:
- Live API tests only → rejected, flaky in CI, rate limits
- VCR/cassette recording → considered for future, overkill for v1 with simple GET endpoints
