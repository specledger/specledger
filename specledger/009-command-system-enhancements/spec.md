# Feature Specification: Command System Enhancements

**Feature Branch**: `009-command-system-enhancements`
**Created**: 2026-02-10
**Status**: Draft
**Input**: User description: "Document changes in diff - new commands, updated commands, utility scripts, and bash script fixes"

## Summary

This feature documents a significant update to the SpecLedger command system including:
- **3 New Commands**: `/specledger.audit`, `/specledger.help`, `/specledger.revise`
- **2 Updated Commands**: `/specledger.adopt` (audit mode), `/specledger.implement` (Supabase sync)
- **8 Commands Enhanced**: Added "Purpose" and "When to use" sections for discoverability
- **2 New Utility Scripts**: `pull-issues.js`, `review-comments.js`
- **Bash Script Fixes**: Updated paths from `.specify` to `.specledger`
- **AGENTS.md Cleanup**: Simplified and focused on bd issue tracking

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Access Command Quick Reference (Priority: P1)

A developer new to SpecLedger wants to quickly discover available commands and understand the workflow. They run `/specledger.help` to see a categorized list of all commands with descriptions and workflow guidance.

**Why this priority**: Discoverability is essential for adoption - users must know what tools are available before they can use them.

**Independent Test**: Run `/specledger.help` and verify it displays all commands organized by workflow stage with clear descriptions.

**Acceptance Scenarios**:

1. **Given** user runs `/specledger.help`, **When** command executes, **Then** displays commands grouped by: Core Workflow, Analysis & Validation, Setup & Configuration, and Collaboration
2. **Given** user sees help output, **When** reviewing content, **Then** each command has a description and workflow examples are provided

---

### User Story 2 - Perform Codebase Audit (Priority: P1)

A developer encountering an unfamiliar codebase wants to understand its structure before creating specifications. They run `/specledger.audit` to perform a two-phase analysis: quick reconnaissance followed by deep module analysis.

**Why this priority**: Understanding an existing codebase is a prerequisite for creating accurate specifications from existing code.

**Independent Test**: Run `/specledger.audit` on a Go project and verify it identifies tech stack, architecture pattern, modules, and generates JSON cache.

**Acceptance Scenarios**:

1. **Given** user runs `/specledger.audit` in a Go project, **When** quick phase completes, **Then** detects Go language, identifies `go.mod`, finds entry points like `main.go`
2. **Given** quick phase completed, **When** deep phase runs, **Then** discovers logical modules, extracts key functions, data models, and builds dependency graph
3. **Given** audit completes, **When** checking output, **Then** `scripts/audit-cache.json` contains project metadata, modules, and dependencies

---

### User Story 3 - Address Review Comments (Priority: P2)

A developer has pushed changes and received review comments from team members via Supabase. They run `/specledger.revise` to fetch unresolved comments and address them interactively.

**Why this priority**: Streamlined review workflow improves team collaboration and ensures feedback is systematically addressed.

**Independent Test**: Run `/specledger.revise` after review comments exist and verify it fetches, displays, and allows addressing each comment.

**Acceptance Scenarios**:

1. **Given** user is logged in via `sl login`, **When** runs `/specledger.revise`, **Then** fetches all unresolved review comments for the repository
2. **Given** comments are fetched, **When** processing each comment, **Then** displays context, selected text, and presents options to address feedback
3. **Given** user selects an option, **When** edit is applied, **Then** file is updated and comment can be marked as resolved

---

### User Story 4 - Sync Issues Before Implementation (Priority: P2)

A developer starting implementation wants to ensure they have the latest issue state from team members. The `/specledger.implement` command now syncs issues from Supabase before beginning work.

**Why this priority**: Prevents working on issues already claimed by others and ensures issue state is current.

**Independent Test**: Run `/specledger.implement` and verify it syncs issues from Supabase before checking prerequisites.

**Acceptance Scenarios**:

1. **Given** user runs `/specledger.implement`, **When** sync runs, **Then** executes `node scripts/pull-issues.js` with repo owner/name from git remote
2. **Given** sync completes, **When** checking `.beads/issues.jsonl`, **Then** contains latest issues, dependencies, and comments from Supabase
3. **Given** user is not logged in, **When** sync fails, **Then** stops and prompts user to run `sl login`

---

### User Story 5 - Create Spec from Audit Data (Priority: P3)

A developer has run `/specledger.audit` and wants to create a specification for a discovered module. They use `/specledger.adopt --module-id [ID] --from-audit` to generate a spec from cached audit data.

**Why this priority**: Streamlines spec creation for existing code by leveraging pre-analyzed module data.

**Independent Test**: Run `/specledger.adopt --module-id user-auth --from-audit` after audit and verify spec is generated from cached data.

**Acceptance Scenarios**:

1. **Given** `scripts/audit-cache.json` exists, **When** user runs adopt with `--from-audit`, **Then** reads module data from cache instead of analyzing branch
2. **Given** module found in cache, **When** generating spec, **Then** User Scenarios inferred from key functions, Requirements from API contracts, Entities from data models

---

### Edge Cases

- What happens when user runs `/specledger.revise` but is not logged in? → Error: "Credentials file not found. Run 'sl login' to authenticate first."
- What happens when `/specledger.audit` runs in empty directory? → Error: "No source files found in the specified scope."
- What happens when audit cache is stale? → Error: "Audit cache stale - run /specledger.audit --force"
- What happens when module-id doesn't exist in cache? → Error: "Module [ID] not found in audit cache"

## Requirements *(mandatory)*

### Functional Requirements

**New Commands:**
- **FR-001**: System MUST provide `/specledger.help` command that displays all available commands organized by workflow stage
- **FR-002**: System MUST provide `/specledger.audit` command with two-phase analysis: quick reconnaissance (~15 min) and deep module analysis (~30+ min)
- **FR-003**: System MUST provide `/specledger.revise` command that fetches review comments from Supabase and allows interactive resolution

**Updated Commands:**
- **FR-004**: System MUST sync issues from Supabase as first step of `/specledger.implement` to ensure latest issue state
- **FR-005**: System MUST support `--from-audit` mode in `/specledger.adopt` to create specs from cached audit data

**Discoverability:**
- **FR-006**: All commands MUST include "Purpose" and "When to use" sections in their documentation

**Utility Scripts:**
- **FR-007**: System MUST provide `scripts/pull-issues.js` utility to sync beads issues from Supabase to `.beads/issues.jsonl`
- **FR-008**: System MUST provide `scripts/review-comments.js` utility with subcommands: `by-path`, `by-project`, `by-change`, `resolve`

**Path Corrections:**
- **FR-009**: Bash scripts MUST use `.specledger` directory instead of `.specify` for path detection
- **FR-010**: Bash scripts MUST use `specledger/` directory instead of `specs/` for feature directories

### Key Entities

- **Command**: A SpecLedger slash command with description, purpose, and execution flow
- **Audit Cache**: JSON file at `scripts/audit-cache.json` containing project metadata, modules, and dependencies
- **Module**: A logical unit of code discovered during audit, containing key functions, data models, and API contracts
- **Review Comment**: Feedback from team members stored in Supabase with file path, selected text, and resolution status
- **Beads Issue**: Task/bug/feature tracked in `.beads/issues.jsonl` and synced with Supabase

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can discover all available commands in under 30 seconds using `/specledger.help`
- **SC-002**: Codebase audit completes quick phase in under 15 minutes for projects under 50,000 LOC
- **SC-003**: Review comments are fetched and displayed within 5 seconds of running `/specledger.revise`
- **SC-004**: Issue sync completes within 10 seconds before implementation starts
- **SC-005**: All commands include purpose descriptions, improving new user onboarding experience
- **SC-006**: Path-related errors are eliminated after bash script fixes

### Previous work

- **008-cli-auth**: Browser-based OAuth authentication for CLI - provides `sl login` used by revise and implement commands
- **006-opensource-readiness**: Core SpecLedger infrastructure and command framework

## Change Summary

### Files Added (5)
| File | Description |
|------|-------------|
| `.claude/commands/specledger.audit.md` | New codebase audit command |
| `.claude/commands/specledger.help.md` | New help/reference command |
| `.claude/commands/specledger.revise.md` | New review comments command |
| `scripts/pull-issues.js` | Sync beads issues from Supabase |
| `scripts/review-comments.js` | Query/manage review comments |

### Files Modified (13)
| File | Changes |
|------|---------|
| `.claude/commands/specledger.adopt.md` | Simplified, added `--from-audit` mode |
| `.claude/commands/specledger.implement.md` | Added Supabase sync as step 1 |
| `.claude/commands/specledger.*.md` (8 files) | Added Purpose sections |
| `.specledger/scripts/bash/*.sh` (5 files) | Fixed paths: `.specify` → `.specledger`, `specs/` → `specledger/` |
| `AGENTS.md` | Simplified, removed dependency section, focused on bd |

### Statistics
- **Lines added**: ~1,519
- **Lines removed**: ~694
- **Net change**: +825 lines
