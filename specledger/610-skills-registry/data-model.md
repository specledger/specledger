# Data Model: Skills Registry Integration

**Branch**: `610-skills-registry` | **Date**: 2026-04-05

## Entities

### SkillSource

Parsed representation of a user-provided skill identifier.

| Field | Type | Description |
|-------|------|-------------|
| Owner | string | GitHub org/user (e.g., "vercel-labs") |
| Repo | string | Repository name (e.g., "agent-skills") |
| SkillFilter | string (optional) | Specific skill name from `@skill` syntax |
| Ref | string | Git ref, defaults to "main" |

**Validation**:
- Owner and Repo are required
- Must contain exactly one `/` separator
- `@` splits skill filter from repo path
- No path traversal (`..`) in any segment

**Identity**: `{Owner}/{Repo}` (with optional `@{SkillFilter}`)

### SkillMetadata

Parsed from SKILL.md YAML frontmatter.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Skill identifier (lowercase, hyphens) |
| Description | string | What the skill does |
| Slug | string | Full path: `{owner}/{repo}/{name}` |
| Source | string | `{owner}/{repo}` |

**Validation**:
- Name and Description are required in frontmatter
- Skills with `metadata.internal: true` are hidden unless `INSTALL_INTERNAL_SKILLS=1`

### SkillSearchResult

Returned from the skills.sh search API.

| Field | Type | Description |
|-------|------|-------------|
| ID | string | Full slug: `{owner}/{repo}/{skill-name}` |
| Name | string | Skill display name |
| Source | string | `{owner}/{repo}` |
| Installs | int | Total install count |

**Sort**: Client-side by Installs descending (API does not support server-side sort).

### PartnerAudit

Security audit data for a single partner.

| Field | Type | Description |
|-------|------|-------------|
| Risk | string | One of: safe, low, medium, high, critical, unknown |
| Alerts | int | Number of security alerts |
| Score | int | Security score (0-100) |
| AnalyzedAt | time.Time | When the analysis was last run |

### SkillAuditResult

Audit results for a single skill across all partners.

| Field | Type | Description |
|-------|------|-------------|
| Slug | string | Skill identifier |
| ATH | *PartnerAudit | App Threat Intel (general risk) |
| Socket | *PartnerAudit | Socket Supply Chain alerts |
| Snyk | *PartnerAudit | Snyk vulnerability scan |

**Note**: Any partner may be nil if no data available.

### LocalSkillLockEntry

A single entry in `skills-lock.json`. Schema matches official Vercel format.

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Source | string | `source` | `{owner}/{repo}` |
| Ref | string (optional) | `ref` | Git branch/tag used |
| SourceType | string | `sourceType` | Always "github" for v1 |
| ComputedHash | string | `computedHash` | SHA-256 hex of skill folder contents |

### LocalSkillLockFile

The `skills-lock.json` file. Matches official Vercel schema v1.

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Version | int | `version` | Always 1 |
| Skills | map[string]LocalSkillLockEntry | `skills` | Keyed by skill name, sorted alphabetically on write |

**File location**: Project root `./skills-lock.json`

**Hash computation** (SHA-256):
1. Recursively collect all files in skill directory
2. Skip `.git` and `node_modules` directories
3. Sort files by relative path (deterministic)
4. For each file: hash relative path + file contents
5. Return hex digest

## Relationships

```
SkillSource ──parses──> owner/repo@skill
     │
     ├── search ──> SkillSearchResult[]
     │
     ├── info ──> SkillMetadata + SkillAuditResult
     │
     └── add ──> downloads SKILL.md
                  │
                  ├── writes to agent paths (from registry)
                  ├── updates LocalSkillLockFile
                  ├── fetches SkillAuditResult (non-blocking)
                  └── sends telemetry (fire-and-forget)

LocalSkillLockFile ──read by──> list, remove, audit
```

## State Transitions

Skills have a simple lifecycle:

```
[Not Installed] ──add──> [Installed] ──remove──> [Not Installed]
                              │
                              └── audit ──> [Installed + Audited]
```

No complex state machine. Skills are either installed (present in lock file + agent directories) or not.
