# Feature Specification: Skills Registry Integration

**Feature Branch**: `610-skills-registry`
**Created**: 2026-04-05
**Status**: Draft
**Input**: GitHub issue #94 — Add `sl skills` subcommand integrating with Vercel's skills.sh registry for discovering, installing, and auditing reusable AI agent skills.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Search for Skills (Priority: P1)

As a developer setting up a new project, I want to search the skills.sh registry for relevant agent skills so that I can quickly discover community-built capabilities without leaving my CLI workflow.

**Why this priority**: Discovery is the entry point — users must find skills before they can install or audit them. Without search, no other command is useful.

**Independent Test**: Run `sl skills search "commit"` and verify a list of matching skills is displayed with name, source repo, and install count.

**Acceptance Scenarios**:

1. **Given** a search query, **When** running `sl skills search "commit"`, **Then** matching skills are displayed with name, source repository, and install count
2. **Given** a search query with `--json` flag, **When** run, **Then** output is valid JSON array of skill results
3. **Given** a search query with `--limit 5` flag, **When** run, **Then** at most 5 results are returned
4. **Given** a query with no matches, **When** run, **Then** a "no skills found" message is shown (not an error)
5. **Given** the skills.sh API is unreachable, **When** run, **Then** a user-friendly network error is shown with guidance

---

### User Story 2 - Install a Skill (Priority: P1)

As a developer who found a useful skill, I want to install it into my project so that my AI agent can use it immediately.

**Why this priority**: Installation is the core value action — the reason users engage with the registry at all.

**Independent Test**: Run `sl skills add owner/repo@skill-name` and verify the SKILL.md file is saved locally and recorded in the lock file.

**Acceptance Scenarios**:

1. **Given** a valid skill identifier `owner/repo@skill-name`, **When** running `sl skills add owner/repo@skill-name`, **Then** the SKILL.md is downloaded and saved to `.claude/skills/{skill-name}/SKILL.md`
2. **Given** a successful install, **When** complete, **Then** a `skills-lock.json` file is created or updated with the installed skill's metadata (source, version/commit, installed date)
3. **Given** a skill that is already installed, **When** running `sl skills add` for the same skill, **Then** the user is informed the skill already exists and asked to confirm overwrite
4. **Given** an invalid skill identifier (repo not found, no SKILL.md), **When** run, **Then** a clear error message explains what went wrong
5. **Given** `DISABLE_TELEMETRY` or `DO_NOT_TRACK` is not set, **When** a skill is successfully installed, **Then** an install telemetry ping is sent to skills.sh to count the install in the ecosystem
6. **Given** `DISABLE_TELEMETRY` or `DO_NOT_TRACK` is set, **When** a skill is installed, **Then** no telemetry ping is sent

---

### User Story 3 - View Skill Details and Security Audit (Priority: P2)

As a security-conscious developer, I want to view a skill's details and its Snyk security audit results before installing so that I can make an informed decision about trust.

**Why this priority**: Important for trust and safety, but many users will install directly from search results without checking audit data first.

**Independent Test**: Run `sl skills info owner/repo@skill-name` and verify skill metadata and security risk level are displayed.

**Acceptance Scenarios**:

1. **Given** a valid skill identifier, **When** running `sl skills info owner/repo@skill-name`, **Then** skill details (name, source, description) and security audit results (risk level, alert count, score, analysis date) are displayed
2. **Given** `--json` flag, **When** run, **Then** output is valid JSON with both skill metadata and audit data
3. **Given** a skill with no audit data available, **When** run, **Then** risk is shown as "unknown" with a note that no audit has been performed
4. **Given** a skill with "high" or "critical" risk, **When** displayed, **Then** a prominent warning is shown advising caution

---

### User Story 4 - List Installed Skills (Priority: P2)

As a developer managing my project's agent capabilities, I want to list all locally installed skills so that I can see what's currently available and manage them.

**Why this priority**: Supports ongoing management but not needed for initial adoption.

**Independent Test**: Run `sl skills list` after installing skills and verify all installed skills are shown with source and install date.

**Acceptance Scenarios**:

1. **Given** one or more installed skills, **When** running `sl skills list`, **Then** each skill is shown with name, source repo, and install date
2. **Given** `--json` flag, **When** run, **Then** output is valid JSON array from `skills-lock.json`
3. **Given** no installed skills, **When** run, **Then** a helpful message is shown suggesting `sl skills search`

---

### User Story 5 - Remove a Skill (Priority: P2)

As a developer who no longer needs a skill, I want to remove it cleanly so that my project stays tidy and my agent doesn't use outdated capabilities.

**Why this priority**: Supports lifecycle management but less frequent than install/search.

**Independent Test**: Run `sl skills remove skill-name` and verify the skill directory and lock file entry are removed.

**Acceptance Scenarios**:

1. **Given** an installed skill name, **When** running `sl skills remove skill-name`, **Then** the `.claude/skills/{skill-name}/` directory is deleted and the entry is removed from `skills-lock.json`
2. **Given** a skill that is not installed, **When** run, **Then** an error message states the skill is not found locally
3. **Given** `--json` flag, **When** run, **Then** confirmation is output as JSON

---

### User Story 6 - Audit Installed Skills (Priority: P3)

As a team lead reviewing project security, I want to run a security audit on all installed skills so that I can verify none have known vulnerabilities.

**Why this priority**: Valuable for security posture but not required for day-to-day skill usage.

**Independent Test**: Run `sl skills audit` and verify all installed skills are checked against the Snyk audit API with results displayed.

**Acceptance Scenarios**:

1. **Given** installed skills, **When** running `sl skills audit`, **Then** each skill's security risk level, alert count, and score are displayed
2. **Given** `--all` flag, **When** run, **Then** all installed skills are audited (same as default)
3. **Given** a specific skill name, **When** running `sl skills audit skill-name`, **Then** only that skill is audited
4. **Given** `--json` flag, **When** run, **Then** output is valid JSON with audit results per skill
5. **Given** any skill has "high" or "critical" risk, **When** audit completes, **Then** a summary warning is shown at the end

---

### Edge Cases

- What happens when skills.sh API returns rate-limited responses? -> Display a retry-after message with the wait time
- What happens when a SKILL.md file in the repo is malformed or empty? -> Show an error that the skill could not be parsed, skip installation
- What happens when `skills-lock.json` is corrupted or manually edited? -> Attempt to parse what's valid, warn about unparseable entries
- What happens when the `.claude/skills/` directory doesn't exist? -> Create it automatically on first install
- What happens when network is unavailable during `sl skills add`? -> Fail with clear network error, no partial file left behind
- What happens when two users install different versions of the same skill? -> Lock file records the commit/version, last write wins (standard git conflict resolution)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `sl skills search` MUST query the skills.sh search API and display results with name, source, and install count
- **FR-002**: `sl skills search` MUST support `--limit` flag to control number of results (default: 10)
- **FR-003**: `sl skills info` MUST display skill metadata and Snyk security audit results (risk level, alerts, score, analysis date)
- **FR-004**: `sl skills add` MUST download the SKILL.md from raw GitHub content and save to `.claude/skills/{name}/SKILL.md`
- **FR-005**: `sl skills add` MUST create or update `skills-lock.json` with skill metadata on successful install
- **FR-006**: `sl skills add` MUST send a telemetry ping to skills.sh on successful install unless `DISABLE_TELEMETRY` or `DO_NOT_TRACK` env vars are set
- **FR-007**: `sl skills remove` MUST delete the skill directory and remove its entry from `skills-lock.json`
- **FR-008**: `sl skills list` MUST read from `skills-lock.json` and display all installed skills
- **FR-009**: `sl skills audit` MUST query the Snyk audit API for installed skills and display security results
- **FR-010**: All `sl skills` subcommands MUST support `--json` output flag for agent consumption
- **FR-011**: All `sl skills` subcommands MUST handle network errors gracefully with user-friendly messages
- **FR-012**: `sl skills add` MUST warn and prompt for confirmation when overwriting an already-installed skill
- **FR-013**: Telemetry ping MUST identify the client as `v=sl-{version}` to distinguish installs via SpecLedger

### Key Entities

- **SkillResult**: A skill returned from the search API
  - id (owner/repo/skill-name), name, source repo, install count
- **SkillAudit**: Security audit data from the Snyk API
  - risk level (safe/low/medium/high/critical/unknown), alert count, score (0-100), analysis date
- **InstalledSkill**: A locally installed skill recorded in `skills-lock.json`
  - name, source (owner/repo), skill slug, installed date, content path

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can discover relevant skills in under 10 seconds using `sl skills search`
- **SC-002**: Users can install a skill in a single command without leaving the terminal
- **SC-003**: Security audit results are available for 100% of installed skills that have Snyk data
- **SC-004**: Install telemetry is accurately reported back to skills.sh (installs via `sl` appear in leaderboard counts)
- **SC-005**: All 6 subcommands (search, info, add, remove, list, audit) produce valid JSON with `--json` flag
- **SC-006**: No Node.js dependency is required — the feature works as a native Go client

### Previous work

### Epic: 601 - CLI Skills

- **sl-comment skill**: Established pattern for skill installation to `.claude/skills/` directory
- **Embedded skill templates**: Skill file management patterns in `pkg/embedded/`

### Epic: 001 - Coding Agent Support

- **Agent registry**: Pattern for agent configuration and management that skills build upon

## Dependencies & Assumptions

### Dependencies

- **skills.sh public API**: Search and telemetry endpoints (no auth required)
- **Snyk audit API via skills.sh**: Security scanning endpoint (no auth required)
- **GitHub raw content**: SKILL.md files fetched from public GitHub repos

### Assumptions

- skills.sh public API endpoints remain stable and unauthenticated
- SKILL.md files follow the standard format with YAML frontmatter
- `.claude/skills/` is the correct installation directory (consistent with Claude Code conventions)
- `skills-lock.json` lives in the project root alongside other config files
- Telemetry follows the same fire-and-forget GET pattern as the official `npx skills` CLI

## Out of Scope

- Integration with tessl.io (prohibited by their ToS — see issue #94 for rationale)
- Skill authoring or publishing (this feature is read-only from the registry)
- Automatic skill updates or version pinning
- Private/authenticated GitHub repos (only public repos supported initially)
- Caching of audit results (can be added later)
- TUI/interactive skill browser
