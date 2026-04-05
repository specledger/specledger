# Feature Specification: Skills Registry Integration

**Feature Branch**: `610-skills-registry`
**Created**: 2026-04-05
**Status**: Draft
**Input**: GitHub issue #94 — Add `sl skill` subcommand integrating with Vercel's skills.sh registry for discovering, installing, and auditing reusable AI agent skills.

## Clarifications

### Session 2026-04-05

- Q: How should `sl skill add` determine which agent directories to install to? → A: Read configured agents from `specledger.yaml` (set during `sl init`). Install SKILL.md to each agent's mapped path. Defer symlink-vs-copy preference to a future `sl config` enhancement.
- Q: Should the lock file format match the official Vercel `skills-lock.json` schema? → A: Yes, commit to compatibility now. Local lock file MUST use the same schema (version, skills map with source/ref/sourceType/computedHash).
- Q: Should search results support pagination? → A: No. The upstream API returns max 10 results. Just support `--limit` and add footer hints per CLI design principles. Pagination is YAGNI.
- Q: Should we surface all 3 audit partners (ATH, Socket, Snyk) or just Snyk? → A: Surface all 3 partners, matching the official CLI's Security Risk Assessments table.
- Q: Should audit data be shown inline during `sl skill add`? → A: Yes. Fetch audit in parallel during add, display before install confirmation. Non-blocking (3s timeout, skip on failure).
- Q: What agent paths should v1 target for installation? → A: Use configured agents from `specledger.yaml`. If only Claude Code is configured, install to `.claude/skills/` only. No symlink complexity in v1.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Search for Skills (Priority: P1)

As a developer setting up a new project, I want to search the skills.sh registry for relevant agent skills so that I can quickly discover community-built capabilities without leaving my CLI workflow.

**Why this priority**: Discovery is the entry point — users must find skills before they can install or audit them. Without search, no other command is useful.

**Independent Test**: Run `sl skill search "commit"` and verify a list of matching skills is displayed with name, source repo, and install count.

**Acceptance Scenarios**:

1. **Given** a search query, **When** running `sl skill search "commit"`, **Then** matching skills are displayed in compact format: name, source repo, install count (truncated to 80 chars per CLI design principles)
2. **Given** a search query with `--json` flag, **When** run, **Then** output is valid JSON array of skill results (complete, untruncated)
3. **Given** a search query with `--limit 5` flag, **When** run, **Then** at most 5 results are returned (default: 10)
4. **Given** a query with no matches, **When** run, **Then** a "no skills found" message is shown (not an error)
5. **Given** the skills.sh API is unreachable, **When** run, **Then** a user-friendly error is shown to stderr with suggested fix per CLI design principles
6. **Given** search results are displayed, **When** output completes, **Then** a footer hint suggests the next step: `Use 'sl skill info <slug>' for details or 'sl skill add <slug>' to install`

---

### User Story 2 - Install a Skill (Priority: P1)

As a developer who found a useful skill, I want to install it into my project so that my AI agent can use it immediately.

**Why this priority**: Installation is the core value action — the reason users engage with the registry at all.

**Independent Test**: Run `sl skill add owner/repo@skill-name` and verify the SKILL.md file is saved to configured agent paths and recorded in `skills-lock.json`.

**Acceptance Scenarios**:

1. **Given** a valid skill identifier (`owner/repo@skill-name` shorthand or full git URL), **When** running `sl skill add owner/repo@skill-name`, **Then** the SKILL.md is downloaded (via GitHub API for shorthand, or `git clone --depth 1` for full URLs) and saved to each configured agent's skill path (e.g., `.claude/skills/{skill-name}/SKILL.md`, `.agents/skills/{skill-name}/SKILL.md`) based on agents in `specledger.yaml`
2. **Given** a successful install, **When** complete, **Then** a `skills-lock.json` file is created or updated using the official Vercel schema (version, skills map with source, ref, sourceType, computedHash fields)
3. **Given** a skill that is already installed, **When** running `sl skill add` for the same skill, **Then** the user is informed the skill already exists and asked to confirm overwrite
4. **Given** an invalid skill identifier (repo not found, no SKILL.md, unreachable git host), **When** run, **Then** a clear error to stderr explains what went wrong with suggested fix (e.g., "Repository not found. Check the source format: owner/repo, owner/repo@skill, or a full git URL")
5. **Given** `DISABLE_TELEMETRY` or `DO_NOT_TRACK` is not set and the repo is public, **When** a skill is successfully installed, **Then** an install telemetry ping is sent to skills.sh to count the install in the ecosystem
6. **Given** `DISABLE_TELEMETRY` or `DO_NOT_TRACK` is set, or the repo is private, **When** a skill is installed, **Then** no telemetry ping is sent
7. **Given** the audit API is reachable, **When** running `sl skill add`, **Then** security risk assessments (ATH, Socket, Snyk) are displayed before the install confirmation prompt
8. **Given** the audit API is unreachable or times out (3s), **When** running `sl skill add`, **Then** installation proceeds without audit data (non-blocking)

---

### User Story 3 - View Skill Details and Security Audit (Priority: P2)

As a security-conscious developer, I want to view a skill's details and its security audit results before installing so that I can make an informed decision about trust.

**Why this priority**: Important for trust and safety, but many users will install directly from search results without checking audit data first.

**Independent Test**: Run `sl skill info owner/repo@skill-name` and verify skill metadata and security risk levels from all 3 audit partners are displayed.

**Acceptance Scenarios**:

1. **Given** a valid skill identifier, **When** running `sl skill info owner/repo@skill-name`, **Then** skill details (name, source, description) and security audit results from ATH (general threat), Socket (supply chain alerts), and Snyk (vulnerabilities) are displayed with risk level, alert count, score, and analysis date
2. **Given** `--json` flag, **When** run, **Then** output is valid JSON with both skill metadata and audit data for all available partners
3. **Given** a skill with no audit data available for a partner, **When** run, **Then** that partner's column shows `--` (unknown)
4. **Given** a skill with "high" or "critical" risk from any partner, **When** displayed, **Then** a prominent warning is shown advising caution

---

### User Story 4 - List Installed Skills (Priority: P2)

As a developer managing my project's agent capabilities, I want to list all locally installed skills so that I can see what's currently available and manage them.

**Why this priority**: Supports ongoing management but not needed for initial adoption.

**Independent Test**: Run `sl skill list` after installing skills and verify all installed skills are shown with source and install date.

**Acceptance Scenarios**:

1. **Given** one or more installed skills, **When** running `sl skill list`, **Then** each skill is shown in compact format with name and source repo
2. **Given** `--json` flag, **When** run, **Then** output is valid JSON array from `skills-lock.json`
3. **Given** no installed skills, **When** run, **Then** a helpful message with footer hint suggests `sl skill search`

---

### User Story 5 - Remove a Skill (Priority: P2)

As a developer who no longer needs a skill, I want to remove it cleanly so that my project stays tidy and my agent doesn't use outdated capabilities.

**Why this priority**: Supports lifecycle management but less frequent than install/search.

**Independent Test**: Run `sl skill remove skill-name` and verify the skill directory and lock file entry are removed.

**Acceptance Scenarios**:

1. **Given** an installed skill name, **When** running `sl skill remove skill-name`, **Then** the skill directory is deleted from all configured agent paths and the entry is removed from `skills-lock.json`
2. **Given** a skill that is not installed, **When** run, **Then** an error message to stderr states the skill is not found locally with suggested fix
3. **Given** `--json` flag, **When** run, **Then** confirmation is output as JSON

---

### User Story 6 - Audit Installed Skills (Priority: P3)

As a team lead reviewing project security, I want to run a security audit on all installed skills so that I can verify none have known vulnerabilities.

**Why this priority**: Valuable for security posture but not required for day-to-day skill usage.

**Independent Test**: Run `sl skill audit` and verify all installed skills are checked against the audit API with ATH, Socket, and Snyk results displayed.

**Acceptance Scenarios**:

1. **Given** installed skills, **When** running `sl skill audit`, **Then** each skill's security risk assessments are displayed in a table with ATH, Socket, and Snyk columns showing risk level, alert count, and score
2. **Given** `--all` flag, **When** run, **Then** all installed skills are audited (same as default)
3. **Given** a specific skill name, **When** running `sl skill audit skill-name`, **Then** only that skill is audited
4. **Given** `--json` flag, **When** run, **Then** output is valid JSON with audit results per skill per partner
5. **Given** any skill has "high" or "critical" risk from any partner, **When** audit completes, **Then** a summary warning is shown at the end

---

### User Story 7 - Agent Discovers sl skill Commands (Priority: P2)

As an AI agent working on a project with SpecLedger, I need an embedded skill template that teaches me when and how to use `sl skill` commands so that I can help users find, install, and audit skills without the user needing to know the exact CLI syntax.

**Why this priority**: Without this skill, agents won't know `sl skill` exists. The skill template is what bridges CLI availability to agent discoverability — matching the existing pattern for `sl-comment`, `sl-deps`, and `sl-audit`.

**Independent Test**: Load the `sl-skill` skill into an agent and verify it triggers when a user asks about finding or installing agent skills, and that it provides correct command syntax.

**Acceptance Scenarios**:

1. **Given** a project bootstrapped with `sl init`, **When** the embedded skill templates are applied, **Then** an `sl-skill` skill is present in the agent's skills directory
2. **Given** the `sl-skill` skill is loaded, **When** a user asks "find a skill for commit messages" or "install a skill", **Then** the agent knows to use `sl skill search` and `sl skill add` with correct syntax
3. **Given** the `sl-skill` skill content, **When** reviewed, **Then** it documents all 6 subcommands (search, info, add, remove, list, audit) with usage examples and trigger conditions

---

### Edge Cases

- What happens when skills.sh API returns rate-limited responses? -> Display a retry-after message with the wait time
- What happens when a SKILL.md file in the repo is malformed or empty? -> Fail the install with a clear error (match upstream behavior — require valid YAML frontmatter with name and description fields)
- What happens when `skills-lock.json` is corrupted or invalid JSON? -> Fail fast with an error suggesting the user fix or delete the file
- What happens when the agent skill directory doesn't exist? -> Create it automatically on first install
- What happens when network is unavailable during `sl skill add`? -> Fail with clear network error to stderr, no partial file left behind

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `sl skill search` MUST query the skills.sh search API and display results in compact format with name, source, install count, and footer hint for next steps
- **FR-002**: `sl skill search` MUST support `--limit` flag to control number of results (default: 10)
- **FR-003**: `sl skill info` MUST display skill metadata and security audit results from all available partners (ATH, Socket, Snyk) with risk level, alerts, score, and analysis date
- **FR-004**: `sl skill add` MUST download the SKILL.md from the source repository (via GitHub API for `owner/repo` shorthand, or `git clone --depth 1` for full git URLs) and save to each configured agent's skill path as defined in `specledger.yaml`
- **FR-005**: `sl skill add` MUST create or update `skills-lock.json` using the official Vercel local lock schema (version, skills map with source, ref, sourceType, computedHash)
- **FR-006**: `sl skill add` MUST send a telemetry ping to skills.sh on successful install unless `DISABLE_TELEMETRY` or `DO_NOT_TRACK` env vars are set, or the source repo is private
- **FR-007**: `sl skill remove` MUST delete the skill directory from all configured agent paths and remove its entry from `skills-lock.json`
- **FR-008**: `sl skill list` MUST read from `skills-lock.json` and display all installed skills
- **FR-009**: `sl skill audit` MUST query the audit API for installed skills and display security results from all available partners (ATH, Socket, Snyk)
- **FR-010**: All `sl skill` subcommands MUST support `--json` output flag for agent consumption (complete, untruncated, pipeable)
- **FR-011**: All `sl skill` subcommands MUST handle network errors with actionable error messages to stderr per CLI design principles (what failed, why, suggested fix)
- **FR-012**: `sl skill add` MUST warn and prompt for confirmation when overwriting an already-installed skill
- **FR-013**: Telemetry ping MUST identify the client as `v=specledger-{version}` to distinguish installs via SpecLedger
- **FR-014**: `sl skill add` MUST fetch and display security risk assessments (ATH, Socket, Snyk) before the install confirmation prompt (non-blocking, 3s timeout)
- **FR-015**: `sl skill add` MUST determine target installation directories from the agents configured in `specledger.yaml`
- **FR-016**: An embedded `sl-skill` skill template MUST be shipped in `pkg/embedded/` and installed during `sl init`, following the same pattern as `sl-comment`, `sl-deps`, and `sl-audit` skills

### Key Entities

- **SkillResult**: A skill returned from the search API
  - id (owner/repo/skill-name), name, source repo, install count
- **SkillAudit**: Security audit data from the audit API, per partner
  - partner (ath/socket/snyk), risk level (safe/low/medium/high/critical/unknown), alert count, score (0-100), analysis date
- **InstalledSkill**: A locally installed skill recorded in `skills-lock.json`
  - name, source (owner/repo), ref, sourceType, computedHash (SHA-256 of skill folder contents)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can discover relevant skills using `sl skill search` with results displayed in a single command invocation
- **SC-002**: Users can install a skill in a single command without leaving the terminal
- **SC-003**: Security audit results from all available partners are displayed for 100% of skills that have audit data
- **SC-004**: Install telemetry is accurately reported back to skills.sh (installs via `sl` appear in leaderboard counts)
- **SC-005**: All 6 subcommands (search, info, add, remove, list, audit) produce valid JSON with `--json` flag
- **SC-008**: The `sl-skill` embedded skill template is present after `sl init` and correctly guides agents to use `sl skill` commands
- **SC-006**: No Node.js dependency is required — the feature works as a native Go client
- **SC-007**: `skills-lock.json` produced by `sl skill add` is parseable by the official `npx skills` CLI (format compatibility)

### Previous work

### Epic: 601 - CLI Skills

- **sl-comment skill**: Established pattern for skill installation to `.claude/skills/` directory
- **Embedded skill templates**: Skill file management patterns in `pkg/embedded/`

### Epic: 001 - Coding Agent Support

- **Agent registry**: Pattern for agent configuration and management that skills build upon

## Dependencies & Assumptions

### Dependencies

- **skills.sh public API**: Search and telemetry endpoints (no auth required)
- **Audit API via skills.sh**: Security scanning endpoint returning ATH, Socket, and Snyk partner data (no auth required)
- **GitHub raw content**: SKILL.md files fetched from public GitHub repos
- **specledger.yaml**: Agent configuration for determining installation target paths

### Assumptions

- skills.sh public API endpoints remain stable and unauthenticated
- SKILL.md files follow the standard format with YAML frontmatter (name and description required)
- Agent skill paths are determined by the agents configured in `specledger.yaml` during `sl init`
- `skills-lock.json` lives in the project root and follows the official Vercel local lock schema
- Telemetry follows the same fire-and-forget GET pattern as the official `npx skills` CLI
- Telemetry is skipped for private repos (matching upstream behavior)

## Out of Scope

- Integration with tessl.io (prohibited by their ToS — see issue #94 for rationale)
- Skill authoring or publishing (this feature is read-only from the registry)
- Automatic skill updates or version pinning (requires global lock infrastructure)
- Private/authenticated repos (only public repos supported initially)
- Well-known RFC 8615 skill sources (can be added later)
- Local filesystem skill sources (development-only, can be added later)
- Caching of audit results (can be added later)
- TUI/interactive skill browser
- Symlink-vs-copy installation preference (deferred to future `sl config` enhancement)
- Pagination of search results (upstream API returns max 10, YAGNI)
