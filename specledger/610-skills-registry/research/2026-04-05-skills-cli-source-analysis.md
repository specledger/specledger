# Research: Official skills CLI (Node.js) Source Analysis

**Date**: 2026-04-05
**Context**: Validate our spec's user stories against the actual features and behavior of the official `npx skills` CLI before committing to implementation.
**Time-box**: 30 minutes

## Question

How does the official skills CLI handle search result sorting, security/audit surfacing, and what features does it ship that our spec may be missing or misrepresenting?

## Findings

### Finding 1: Search API is Minimal — No Sort Options

**Source**: `src/find.ts:32-59`

The search API has only two parameters:

```
GET https://skills.sh/api/search?q={query}&limit=10
```

- `q` — search query (min 2 chars before triggering)
- `limit` — hardcoded to 10, not configurable from CLI

**Sorting is client-side only** — results from the API are sorted by install count descending (`sort((a, b) => (b.installs || 0) - (a.installs || 0))`). There are no server-side sort options. No sort flag exists on the CLI.

**Confidence**: High (read actual implementation)

**Impact on our spec**: Our FR-002 specifies `--limit` flag. The upstream API supports `limit` as a param, so we can pass it through. However, we should NOT promise sort options — just document that results are sorted by popularity (installs). This is a backend decision, not something we control.

### Finding 2: Search Command is `find`, Not `search`

**Source**: `src/find.ts`, `src/cli.ts`

The official CLI uses `npx skills find [query]`, not `search`. It has two modes:

1. **Interactive mode** (no query): fzf-style UI with debounced search, arrow key navigation, auto-install on selection
2. **Non-interactive mode** (with query): prints top 6 results with source and install count, then exits

Non-interactive output format (line 295-304):
```
owner/repo@skill-name  12.3K installs
└ https://skills.sh/slug
```

**Impact on our spec**: Our spec calls this `sl skill search`. That's fine — we're not cloning their CLI, we're building our own. But we should consider whether we want interactive mode (P3 at most) or just non-interactive for v1. The interactive TUI is substantial code (~180 lines of raw readline/ANSI).

### Finding 3: Security Audit is Shown During `add`, NOT in `find`/`search`

**Source**: `src/add.ts:108-154, 1186-1435`

This is a critical finding. The audit data is:
- **Fetched in parallel** during `skills add` (fire-and-forget, 3s timeout)
- **Displayed BEFORE the install confirmation prompt** as a "Security Risk Assessments" note
- **NOT shown in search results** — `find` shows no security info at all

The audit table has **3 columns from 3 partners**:

| Column | Partner | What it measures |
|--------|---------|-----------------|
| Gen | ATH (App Threat Intel) | General risk rating |
| Socket | Socket Supply Chain | Alert count for supply chain issues |
| Snyk | Snyk Vulnerabilities | Vulnerability risk rating |

Risk levels: Critical (red bold), High (red), Med (yellow), Low/Safe (green), or `--` if no data.

**API endpoint**:
```
GET https://add-skill.vercel.sh/audit?source={owner/repo}&skills={slug1,slug2}
```

Response structure per skill per partner:
```json
{
  "slug": {
    "ath": { "risk": "safe", "alerts": 0, "score": 100, "analyzedAt": "..." },
    "socket": { "risk": "low", "alerts": 0, "score": 95, "analyzedAt": "..." },
    "snyk": { "risk": "safe", "alerts": 0, "score": 100, "analyzedAt": "..." }
  }
}
```

**Impact on our spec**: 
1. Our issue #94 only mentions Snyk — the audit API actually returns **3 partners** (ath, socket, snyk). We should surface all three.
2. Our spec has a separate `sl skill info` command for audit. The official CLI doesn't have an `info` command — it bakes audit into `add`. We may want to keep `info` as a value-add but should also show audit during `add`.
3. Our spec has a standalone `sl skill audit` command. The official CLI has no standalone audit command either — it only audits during install. Our `audit` command for batch-checking installed skills is a genuine value-add.

### Finding 4: Two Lock File Systems — Global and Local

**Source**: `src/skill-lock.ts`, `src/local-lock.ts`

The official CLI has **two separate lock files**:

| Lock File | Location | Purpose | Checked into git? |
|-----------|----------|---------|-------------------|
| `.skill-lock.json` | `~/.agents/` (global) | Tracks all installs, update hashes, agent prefs | No |
| `skills-lock.json` | Project root (local) | Minimal entries for reproducible installs | Yes |

**Global lock** (`SkillLockEntry`): source, sourceType, sourceUrl, ref, skillPath, skillFolderHash (GitHub tree SHA), installedAt, updatedAt, pluginName

**Local lock** (`LocalSkillLockEntry`): source, ref, sourceType, computedHash (SHA-256 of file contents). Intentionally minimal and timestamp-free to minimize merge conflicts. Skills sorted alphabetically for deterministic output.

**Impact on our spec**: Our spec only mentions `skills-lock.json` (the local one). For v1, the local lock is probably sufficient. But we should be aware that the global lock powers `check`/`update` commands. If we want those later, we'll need a global lock too.

### Finding 5: `check` and `update` Commands Exist

**Source**: `src/cli.ts:370-605`

The official CLI has two commands we didn't spec:

- **`skills check`**: Reads global lock, calls GitHub Trees API to compare folder SHAs, reports which skills have updates available
- **`skills update`**: Finds outdated skills, re-runs `npx skills add` for each to update them

These use GitHub token (GITHUB_TOKEN, GH_TOKEN, or `gh auth token`) for API access to avoid rate limiting.

**Impact on our spec**: We explicitly put these out of scope ("Automatic skill updates or version pinning"). This is correct for v1 — they're complex and require the global lock infrastructure.

### Finding 6: `remove` Command is More Featured Than Expected

**Source**: `src/remove.ts`, README

The official `remove` supports:
- Interactive selection (no args — pick from installed)
- Multiple skills at once
- `--agent` filter (remove from specific agents only)
- `--skill '*'` wildcard
- `--all` shorthand for everything
- Updates both global and local lock files

**Impact on our spec**: Our spec's remove is simpler (remove by name, one at a time). This is fine for v1.

### Finding 7: Telemetry Implementation Details

**Source**: `src/telemetry.ts`

```
GET https://add-skill.vercel.sh/t?v={version}&ci={1|undefined}&event={type}&...params
```

Key behaviors:
- Fire-and-forget (no error handling, never blocks)
- **Skipped entirely for private repos** (checked via GitHub API)
- CI detection via 8 env vars (CI, GITHUB_ACTIONS, GITLAB_CI, etc.)
- Events: find, install, remove, check, update, experimental_sync
- Install event includes: source, skills (comma-separated), agents (comma-separated)

**Impact on our spec**: Our FR-006 and FR-013 cover this well. We should also skip telemetry for private repos (not currently in our spec). The version identifier should be `v=specledger-{version}` (not `sl-{version}` as our spec says — "sl" is ambiguous).

### Finding 8: Source Format Flexibility

**Source**: README, `src/source-parser.ts`

The official CLI accepts many source formats:
- `owner/repo` (GitHub shorthand)
- Full GitHub/GitLab URLs
- Direct tree paths (`github.com/.../tree/main/skills/name`)
- Git SSH URLs
- Local paths (`./my-skills`)

Our spec only covers `owner/repo@skill-name`. This is fine for v1, but worth noting.

### Finding 9: No Standalone `info` Command Exists Upstream

The official CLI has: `add`, `find`, `list`, `remove`, `check`, `update`, `init`. There is no `info` command. Our proposed `sl skill info` is a net-new feature that combines skill metadata with audit data in a single view. This is a genuine value-add over the official CLI.

## Decisions

- **Decision 1**: Keep `sl skill search` (our naming) rather than `find` — we're building our own CLI, not cloning theirs. But drop any plans for interactive/TUI mode in v1.
- **Decision 2**: Surface all 3 audit partners (ATH, Socket, Snyk) not just Snyk. The API returns all three and they each measure different things.
- **Decision 3**: Show audit data during `sl skill add` flow (like upstream), AND keep `sl skill info` as a standalone pre-install check (our value-add), AND keep `sl skill audit` for batch checking installed skills.
- **Decision 4**: Use only the local lock file (`skills-lock.json`) for v1. Skip global lock — it's only needed for `check`/`update` which are out of scope.
- **Decision 5**: Skip telemetry for private repos (match upstream behavior). Use `v=specledger-{version}` for the version identifier.

## Recommendations

1. **Update spec FR-003 and US3**: Replace "Snyk security audit" references with "security risk assessments from ATH, Socket, and Snyk" to match actual API response
2. **Add audit display to `sl skill add` flow**: Currently our spec only shows audit in `info` and `audit` commands — the official CLI shows it during install which is the most impactful moment
3. **Update spec FR-013**: Change `v=sl-{version}` to `v=specledger-{version}` for clarity
4. **Add private repo telemetry skip**: Add a requirement that telemetry is skipped for private repos
5. **Clarify search result count**: Document that we'll pass `--limit` to the API but default to 10 (matching upstream), and that results are always sorted by popularity
6. **Do NOT add sort options**: The backend API doesn't support server-side sorting, and client-side sort of 10 results by a single field (installs) is not worth a CLI flag
7. **Consider `--list` flag on `add`**: The official CLI supports `npx skills add repo --list` to preview available skills before installing — useful for repos with multiple skills

## References

- `/Users/vids/specledger/skills/src/find.ts` — Search/find implementation
- `/Users/vids/specledger/skills/src/add.ts:73-154` — Audit display and security table rendering
- `/Users/vids/specledger/skills/src/skill-lock.ts` — Global lock file structure
- `/Users/vids/specledger/skills/src/local-lock.ts` — Local (project) lock file structure
- `/Users/vids/specledger/skills/src/telemetry.ts` — Telemetry implementation
- `/Users/vids/specledger/skills/src/list.ts` — List command with JSON output
- `/Users/vids/specledger/skills/src/cli.ts:370-605` — Check/update commands
- `https://skills.sh/api/search` — Search API endpoint
- `https://add-skill.vercel.sh/audit` — Security audit API endpoint
- `https://add-skill.vercel.sh/t` — Telemetry endpoint
