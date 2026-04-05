# API Contracts: skills.sh Registry

**Date**: 2026-04-05
**Source**: Analyzed from official CLI at `skills/src/`

**Registry model**: skills.sh has no publish/submission process. Skills are automatically indexed via install telemetry — when any user runs `npx skills add <source>`, the install event registers the skill on the leaderboard. The registry indexes skills from any source (GitHub, GitLab, well-known HTTPS endpoints, any git host), not only GitHub. GitHub dominates current coverage due to the fast download API and ecosystem momentum.

## Search API

```
GET https://skills.sh/api/search?q={query}&limit={limit}
```

**Request**:
| Param | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| q | string | yes | — | Search query (min 2 chars) |
| limit | int | no | 10 | Max results to return |

**Response** (200):
```json
{
  "skills": [
    {
      "id": "owner/repo/skill-name",
      "name": "skill-name",
      "installs": 12345,
      "source": "owner/repo"
    }
  ]
}
```

**Error handling**: Non-200 responses return empty results (no error payload documented).

## Audit API

```
GET https://add-skill.vercel.sh/audit?source={owner/repo}&skills={slug1,slug2}
```

**Request**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| source | string | yes | Repository identifier (owner/repo) |
| skills | string | yes | Comma-separated skill slugs |

**Response** (200):
```json
{
  "skill-slug": {
    "ath": {
      "risk": "safe",
      "alerts": 0,
      "score": 100,
      "analyzedAt": "2026-03-01T00:00:00Z"
    },
    "socket": {
      "risk": "low",
      "alerts": 0,
      "score": 95,
      "analyzedAt": "2026-03-01T00:00:00Z"
    },
    "snyk": {
      "risk": "safe",
      "alerts": 0,
      "score": 100,
      "analyzedAt": "2026-03-01T00:00:00Z"
    }
  }
}
```

**Risk levels**: `safe`, `low`, `medium`, `high`, `critical`, `unknown`

**Timeout**: 3 seconds (non-blocking in add flow).

## Telemetry API

```
GET https://add-skill.vercel.sh/t?v={version}&event={type}&source={source}&skills={skills}&agents={agents}
```

**Request** (all query params):
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| v | string | no | Client version (e.g., `specledger-1.2.0`) |
| ci | string | no | `"1"` if CI environment detected |
| event | string | yes | Event type: `install`, `remove`, `find` |
| source | string | yes | Repository identifier |
| skills | string | yes | Comma-separated skill names |
| agents | string | no | Comma-separated agent names |

**Response**: Ignored (fire-and-forget). No error handling.

**Gating** (skip telemetry if any):
- `DISABLE_TELEMETRY` or `DO_NOT_TRACK` env var set
- Source repo is private
- CI environment detected (CI, GITHUB_ACTIONS, GITLAB_CI, CIRCLECI, TRAVIS, BUILDKITE, JENKINS_URL, TEAMCITY_VERSION)

## GitHub Raw Content

```
GET https://raw.githubusercontent.com/{owner}/{repo}/{ref}/skills/{skill-name}/SKILL.md
```

**Response** (200): Raw markdown content of SKILL.md file.

**Error**: 404 if skill or repo doesn't exist.

## GitHub Trees API (Skill Discovery)

```
GET https://api.github.com/repos/{owner}/{repo}/git/trees/{ref}?recursive=1
```

**Auth** (optional): `Authorization: Bearer {GITHUB_TOKEN}` for rate limit headroom.

**Response** (200):
```json
{
  "sha": "abc123",
  "tree": [
    {
      "path": "skills/skill-name/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "def456",
      "size": 1234
    }
  ]
}
```

**Skill discovery logic**: Filter tree entries matching `*/SKILL.md` patterns. Extract skill name from parent directory.

## Local Lock File Contract

**File**: `./skills-lock.json` (project root)

```json
{
  "version": 1,
  "skills": {
    "skill-name": {
      "source": "owner/repo",
      "ref": "main",
      "sourceType": "github",
      "computedHash": "sha256-hex-string"
    }
  }
}
```

**Write rules**:
- Skills sorted alphabetically by name (deterministic output, clean diffs)
- 2-space indent + trailing newline
- Timestamps intentionally omitted (minimize merge conflicts)
