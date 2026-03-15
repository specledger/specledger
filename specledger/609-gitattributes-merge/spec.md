# Feature Specification: Gitattributes Merge for Auto-Generated File Markers

**Feature Branch**: `609-gitattributes-merge`
**Created**: 2026-03-12
**Status**: Draft
**Input**: User description: "PRs are cluttered with auto-generated specledger artifacts. GitHub supports collapsing these when marked with linguist-generated in .gitattributes. The current .gitattributes template is empty and uses simple copy (skip-if-exists / overwrite-if-force), which either loses user content or never adds entries. NOTE: spec.md, plan.md are reviewable artifacts!"

## Clarifications

### Session 2026-03-12

- Q: Which files should be marked linguist-generated? → A: Only `issues.jsonl` and `tasks.md` — both are machine-generated indexes. Checklists, research, templates, commands, and skills are reviewable and should remain visible in PRs.
- Q: When .gitattributes has a malformed sentinel (begin without end), what should the system do? → A: Auto-fix by treating everything from the begin marker to EOF as the managed section and replacing it. This is safe since the file is tracked in git — no data loss risk.
- Q: Should FR-009 (mergeable manifest support) stay in the spec? → A: Keep as-is. It describes a needed capability (merge vs copy distinction) without prescribing implementation.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - First-time init populates .gitattributes (Priority: P1)

A developer runs `sl init` in a new project that has no `.gitattributes` file. After initialization, the project contains a `.gitattributes` with `linguist-generated` markers for machine-generated indexes (`issues.jsonl` and `tasks.md`). When they create a PR, GitHub automatically collapses these files in the diff view, letting reviewers focus on meaningful changes.

**Why this priority**: This is the core use case — most new projects don't have a `.gitattributes` file yet, and this delivers the primary value of decluttered PR views.

**Independent Test**: Run `sl init` in an empty git repo and verify `.gitattributes` contains the expected `linguist-generated` patterns within sentinel markers.

**Acceptance Scenarios**:

1. **Given** a project with no `.gitattributes`, **When** `sl init` completes, **Then** a `.gitattributes` file exists with specledger-managed patterns wrapped in sentinel comments (`# >>> specledger-generated` / `# <<< specledger-generated`)
2. **Given** a project with no `.gitattributes`, **When** the developer creates a PR with specledger artifacts, **Then** GitHub collapses `issues.jsonl` and `tasks.md` in the PR diff view
3. **Given** a project with no `.gitattributes`, **When** the developer creates a PR with spec.md, plan.md, checklists, research, commands, or skills changes, **Then** GitHub shows these files expanded because they are reviewable artifacts

---

### User Story 2 - Re-init merges into existing .gitattributes (Priority: P1)

A developer runs `sl init` in a project that already has a `.gitattributes` file (e.g., from GitHub's template with entries like `*.pbxproj binary`). After re-initialization, the existing user-managed content is preserved and the specledger-managed section is added or updated.

**Why this priority**: Many projects already have `.gitattributes` for binary files, LFS, or other tooling. Overwriting would lose critical configuration; skipping would never add the needed entries. Merging is essential.

**Independent Test**: Create a `.gitattributes` with custom content, run `sl init`, verify custom content is preserved and specledger section is appended.

**Acceptance Scenarios**:

1. **Given** a project with an existing `.gitattributes` containing user-managed entries, **When** `sl init` runs, **Then** user-managed entries are fully preserved and the specledger sentinel block is appended
2. **Given** a project with an existing `.gitattributes` already containing a specledger sentinel block, **When** `sl init` runs again, **Then** only the content between sentinels is updated while everything outside is preserved
3. **Given** a project with an existing `.gitattributes`, **When** `sl init --force` runs, **Then** the same merge behavior occurs (not an overwrite) — force does not destroy user content in mergeable files

---

### User Story 3 - Idempotent re-runs (Priority: P2)

A developer runs `sl init` multiple times (e.g., after upgrading specledger). Each run produces the same `.gitattributes` content — no duplicate sentinel blocks, no content drift.

**Why this priority**: Upgrade safety. Developers should be able to re-init freely without worrying about file corruption.

**Independent Test**: Run `sl init` three times in succession and diff the `.gitattributes` after each run — all three should be identical.

**Acceptance Scenarios**:

1. **Given** a project where `sl init` has already been run, **When** `sl init` is run again with no changes, **Then** the `.gitattributes` file content is byte-identical to before
2. **Given** a project with a specledger sentinel block, **When** the specledger version ships updated patterns, **Then** `sl init` updates only the content between sentinels

---

### Edge Cases

- What happens when `.gitattributes` has a sentinel begin marker but no end marker? The system auto-fixes by replacing everything from the begin marker to EOF with the correct sentinel block. This is safe because `.gitattributes` is git-tracked.
- What happens when `.gitattributes` has only whitespace or comments? Sentinel block is appended normally.
- What happens when the user manually edits content inside the sentinel block? Their edits are overwritten on next `sl init` — this is expected and documented by the sentinel comments.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `.gitattributes` template MUST include `linguist-generated=true` patterns for: `specledger/*/issues.jsonl` and `specledger/*/tasks.md` (machine-generated indexes only)
- **FR-002**: The `.gitattributes` template MUST NOT mark `spec.md` or `plan.md` as linguist-generated, since these are reviewable artifacts
- **FR-003**: The `.gitattributes` template MUST NOT mark checklists, research, templates, commands, or skills as linguist-generated, since these represent reviewable changes
- **FR-004**: The system MUST use sentinel comment markers (`# >>> specledger-generated` / `# <<< specledger-generated`) to delimit the managed section
- **FR-005**: When no `.gitattributes` exists, the system MUST create one containing only the sentinel-wrapped managed section
- **FR-006**: When a `.gitattributes` exists without sentinels, the system MUST append the sentinel block without modifying existing content
- **FR-007**: When a `.gitattributes` exists with sentinels, the system MUST replace only the content between sentinels
- **FR-008**: The merge operation MUST be idempotent — repeated runs produce identical output
- **FR-009**: The playbook manifest MUST support declaring files as "mergeable" to distinguish them from normal copy behavior
- **FR-010**: The `--force` flag MUST NOT cause mergeable files to be overwritten — merge behavior applies regardless of force setting
- **FR-011**: When a sentinel begin marker exists without a matching end marker, the system MUST auto-fix by treating everything from the begin marker to EOF as the managed section and replacing it with the correct sentinel block (safe because the file is git-tracked)

### Key Entities

- **Sentinel Block**: A delimited section in a text file, bounded by begin/end comment markers, containing content managed by specledger
- **Mergeable File**: A structure file declared in the playbook manifest that uses merge (not copy) semantics during init

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After `sl init`, 100% of auto-generated artifact patterns are marked with `linguist-generated=true` in `.gitattributes`
- **SC-002**: After `sl init` in a project with existing `.gitattributes`, 100% of pre-existing user content is preserved
- **SC-003**: Running `sl init` N times produces a byte-identical `.gitattributes` for N >= 2
- **SC-004**: PRs containing specledger artifacts have auto-generated files collapsed by default in GitHub's diff view

### Previous work

- **GitHub Issue #74**: [BUG] PRs are loitered with autogenerated files — the originating bug report
- **PR #70**: Referenced in issue as example of cluttered PR view (review comment r2921999907)
- **135-fix-missing-chmod-x**: Previous work on template file handling (executable permissions)

## Assumptions

- GitHub's `linguist-generated` attribute uses gitattributes glob patterns, which support `*` and `**` wildcards
- The `#` comment syntax is appropriate for `.gitattributes` sentinel markers
- `spec.md` and `plan.md` are intentionally excluded from `linguist-generated` marking because they contain human-reviewable design decisions
- `tasks.md` IS marked as linguist-generated because it is a machine-generated index for LLM agents (task data lives in `issues.jsonl`, managed via `sl issue` commands)
- Checklists, research, templates, commands, and skills are intentionally excluded — they represent reviewable changes that should be visible in PRs
