# Feature Specification: Bash Script to Go CLI Migration

**Feature Branch**: `600-bash-cli-migration`
**Created**: 2026-03-02
**Status**: Draft
**Input**: User description: "create new specs for stream 2"

## Overview

Replace bash scripts with Go CLI commands for cross-platform support, eliminating `jq`/`bash`/`sed`/`grep` dependencies. This is **Stream 2** of the 3-stream SDD alignment effort.

| Stream | Focus | Feature | Status |
|--------|-------|---------|--------|
| 1 | AI command consolidation | 599-alignment | Complete |
| **2** | **Bash script → Go CLI migration** | **600-bash-cli-migration** | **This spec** |
| 3 | New CLI + skills | TBD | Future |

**Stream 2 Scope** (this spec - Go code changes):
- Replace 6 bash scripts with 4 Go CLI commands
- Eliminate external dependencies (jq, grep, sed, git CLI)
- Provide JSON output for agent consumption
- Ensure cross-platform compatibility (macOS, Linux, Windows)

**Script → Command Mapping**:

| Script | Go Command | Complexity |
|--------|------------|------------|
| `common.sh` | Absorbed into `internal/ref/` | N/A |
| `check-prerequisites.sh` | `sl spec info` | Medium |
| `create-new-feature.sh` | `sl spec create` | High |
| `setup-plan.sh` | `sl spec setup-plan` | Low |
| `update-agent-context.sh` | `sl context update` | High |
| `adopt-feature-branch.sh` | Superseded by ContextDetector | N/A |

**Dependencies Eliminated**:

| Dependency | Used By | Go Replacement |
|------------|---------|----------------|
| `jq` | common.sh, adopt-feature-branch.sh | `encoding/json` |
| `grep` | check-prerequisites.sh, update-agent-context.sh | `regexp`, `strings` |
| `sed` | update-agent-context.sh | `strings.Replace`, `regexp` |
| `git` CLI | create-new-feature.sh | `go-git/v5` |

## User Scenarios & Testing *(mandatory)*

### User Story 1 - sl spec info Command (Priority: P1)

As an AI agent executing SDD workflows, I need `sl spec info` to get feature paths and prerequisite validation so that I can verify feature state before proceeding.

**Why this priority**: Replaces `check-prerequisites.sh` which is called by multiple AI commands.

**Independent Test**: Run `sl spec info --json` and verify JSON output with FEATURE_DIR and AVAILABLE_DOCS fields.

**Acceptance Scenarios**:

1. **Given** a feature branch, **When** running `sl spec info --json`, **Then** output contains FEATURE_DIR, BRANCH, FEATURE_SPEC paths
2. **Given** `--require-plan` flag, **When** plan.md missing, **Then** command exits with error and message
3. **Given** `--require-tasks` flag, **When** tasks.md missing, **Then** command exits with error and message
4. **Given** `--include-tasks` flag, **When** tasks.md exists, **Then** output includes AVAILABLE_DOCS with tasks.md
5. **Given** `--paths-only` flag, **When** run, **Then** output is minimal JSON with just paths (no doc discovery)

---

### User Story 2 - sl spec create Command (Priority: P1)

As a developer creating a new feature, I need `sl spec create` to generate a feature branch and spec directory so that I can start a new feature without bash scripts.

**Why this priority**: Replaces `create-new-feature.sh` which is called by `/specledger.specify`.

**Independent Test**: Run `sl spec create --number 999 --short-name "test" --json` and verify branch and spec directory created.

**Acceptance Scenarios**:

1. **Given** a short-name and number, **When** running `sl spec create`, **Then** branch `NNN-short-name` is created
2. **Given** a description with stop-words, **When** generating short-name, **Then** common words (the, a, an, etc.) are filtered
3. **Given** a description with acronyms, **When** generating short-name, **Then** acronyms are preserved (OAuth2, API, JWT)
4. **Given** a long description, **When** generating branch name, **Then** name is truncated to 244 bytes (GitHub limit)
5. **Given** `--json` flag, **When** run, **Then** output contains BRANCH_NAME, FEATURE_DIR, SPEC_FILE paths
6. **Given** existing local feature with same number, **When** run, **Then** error with collision message

---

### User Story 3 - sl spec setup-plan Command (Priority: P2)

As a developer starting implementation planning, I need `sl spec setup-plan` to copy plan templates so that I can begin planning without bash scripts.

**Why this priority**: Replaces `setup-plan.sh` which is called by `/specledger.plan`.

**Independent Test**: Run `sl spec setup-plan --json` and verify plan.md created with template content.

**Acceptance Scenarios**:

1. **Given** a feature directory, **When** running `sl spec setup-plan`, **Then** plan.md is created from template
2. **Given** plan.md already exists, **When** run, **Then** error with "already exists" message
3. **Given** `--json` flag, **When** run, **Then** output contains PLAN_FILE path

---

### User Story 4 - sl context update Command (Priority: P1)

As a developer updating agent context files, I need `sl context update` to parse plan.md and update agent files so that AI assistants have current feature context.

**Why this priority**: Replaces `update-agent-context.sh` which is called after plan changes.

**Independent Test**: Run `sl context update claude` and verify CLAUDE.md updated with plan metadata.

**Acceptance Scenarios**:

1. **Given** plan.md with Technical Context, **When** running `sl context update claude`, **Then** CLAUDE.md Active Technologies section is updated
2. **Given** CLAUDE.md with manual additions, **When** updating, **Then** content between `<!-- MANUAL ADDITIONS -->` markers is preserved
3. **Given** duplicate technology entries, **When** updating, **Then** entries are deduplicated (not appended)
4. **Given** `--agent` flag with different agent, **When** run, **Then** correct agent file is updated (gemini, copilot, cursor, etc.)
5. **Given** `--json` flag, **When** run, **Then** output contains updated file paths

---

### User Story 5 - Cross-Platform Support (Priority: P2)

As a Windows developer, I need all `sl` commands to work without bash so that I can use SpecLedger on any platform.

**Why this priority**: Enables broader adoption but P2 since core functionality works first.

**Independent Test**: Run all 4 new commands on Windows and verify no bash/jq/sed/grep errors.

**Acceptance Scenarios**:

1. **Given** Windows platform, **When** running `sl spec info`, **Then** no "bash: command not found" errors
2. **Given** Windows platform, **When** running `sl spec create`, **Then** branch created with correct path separators
3. **Given** Windows platform, **When** running `sl context update`, **Then** file paths use correct separators
4. **Given** any platform, **When** running commands, **Then** output is identical (JSON format)

---

### User Story 6 - Context Detection Fallback Chain (Priority: P2)

As a developer working on a branch that doesn't follow SpecLedger's naming convention, I need `sl` commands to automatically detect which spec the branch belongs to without manual configuration.

**Why this priority**: Enables workflows with JIRA, Linear, or GitHub UI-created branches.

**Independent Test**: Create a branch named `feature/PROJ-123`, edit files in `specledger/600-bash-cli-migration/`, run `sl spec info` and verify auto-detection works.

**Acceptance Scenarios**:

1. **Given** a branch named `598-sdd-workflow-streamline`, **When** `ContextDetector` runs, **Then** it resolves via regex match (step 1)
2. **Given** a branch named `feature/fix-login` with an alias in `specledger.yaml`, **When** `ContextDetector` runs, **Then** it resolves via yaml alias lookup (step 2)
3. **Given** a branch named `johns-auth-work` with commits touching `specledger/042-auth-improvements/`, **When** `ContextDetector` runs, **Then** it resolves via git heuristic (step 3)
4. **Given** a branch with no regex match, no alias, and no git heuristic result, **When** `ContextDetector` runs in interactive mode, **Then** it lists available specs and prompts user to pick one, saving alias to yaml (step 4)
5. **Given** a saved alias in `specledger.yaml`, **When** any future `sl` command runs on that branch, **Then** it auto-resolves without re-prompting
6. **Given** non-interactive mode (CI, `--spec` flag), **When** detection fails, **Then** the `--spec` flag overrides all detection steps

---

### Edge Cases

- What if git repo is in detached HEAD? → `sl spec info` returns error with suggestion to checkout branch or set SPECIFY_FEATURE
- What if feature number collides with remote branch? → `sl spec create` checks remote (gap identified in 598 research)
- What if plan.md has malformed metadata? → `sl context update` logs warning, skips malformed fields
- What if CLAUDE.md doesn't have markers? → `sl context update` adds markers and appends section
- What if branch name exceeds 244 bytes? → Truncate with warning, preserve feature number
- What if git heuristic finds multiple specledger dirs? → Fall through to step 4 (interactive prompt)
- What if git heuristic finds zero specledger dirs? → Fall through to step 4 (interactive prompt)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `sl spec info` MUST output JSON with FEATURE_DIR, BRANCH, FEATURE_SPEC paths
- **FR-002**: `sl spec info` MUST support `--require-plan`, `--require-tasks`, `--include-tasks`, `--paths-only` flags
- **FR-003**: `sl spec create` MUST generate branch name from number + short-name with stop-word filtering
- **FR-004**: `sl spec create` MUST enforce 244-byte branch name limit (GitHub constraint)
- **FR-005**: `sl spec create` MUST create git branch and spec directory with template files
- **FR-006**: `sl spec setup-plan` MUST copy plan template to feature directory
- **FR-007**: `sl context update` MUST parse plan.md Technical Context section
- **FR-008**: `sl context update` MUST preserve manual additions between markers
- **FR-009**: `sl context update` MUST deduplicate entries (not append)
- **FR-010**: All commands MUST work on macOS, Linux, Windows without bash/jq/sed/grep

### Context Detection Requirements (from 598 US12)

- **FR-011**: `ContextDetector` MUST implement 4-step fallback chain: regex → yaml alias → git heuristic → interactive prompt
- **FR-012**: Branch aliases MUST be stored in `specledger.yaml` under `branch_aliases` key and version-controlled
- **FR-013**: Non-interactive mode (`--spec` flag) MUST override all detection steps
- **FR-014**: SPECIFY_FEATURE env var MUST be checked first and bypass all detection steps

### Key Entities

- **Feature Context**: Resolved from current branch, contains feature directory, spec file, plan file paths
- **Branch Name**: Generated from feature number + short-name with stop-word filtering and length limits
- **Agent Context**: Agent-specific file (CLAUDE.md, GEMINI.md, etc.) updated with plan metadata
- **Manual Additions**: User-edited content preserved between markers during context updates

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 4 new CLI commands available (`sl spec info`, `sl spec create`, `sl spec setup-plan`, `sl context update`)
- **SC-002**: 0 bash script dependencies in new command code paths
- **SC-003**: All commands produce identical JSON output on macOS, Linux, Windows
- **SC-004**: `sl spec create` handles stop-words, acronyms, and length limits correctly
- **SC-005**: `sl context update` preserves manual additions and deduplicates entries
- **SC-006**: AI commands updated to use new CLI instead of bash scripts

### Previous work

### Epic: 598 - SDD Workflow Streamline

- **Bash Script Analysis**: 6 scripts to replace with 4 Go commands
- **Dependencies**: jq, grep, sed, git CLI to eliminate
- **Cross-cutting**: Token-efficient output pattern (D21)

### Epic: 599 - SDD Layer Alignment

- **AI Command Updates**: Commands will reference new CLI instead of bash scripts
- **Layer Architecture**: CLI is L1, should handle data operations

## Dependencies & Assumptions

### Dependencies

- **go-git/v5**: Already a dependency for git operations
- **Cobra**: CLI framework already in use
- **599-alignment**: AI commands updated to use new CLI

### Assumptions

- `pkg/cli/spec/` package exists for context detection
- Template files exist in `pkg/embedded/templates/`
- Agent file mappings are known (17+ agents)

## Out of Scope

- **Stream 1**: AI command consolidation (599-alignment)
- **Stream 3**: New CLI commands (comment, checkpoint, research)
- Changes to AI command content (only which CLI they call)
- TUI features (separate future work)
