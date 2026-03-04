# Feature Specification: SDD Layer Alignment

**Feature Branch**: `599-alignment`
**Created**: 2026-03-02
**Status**: Draft
**Input**: User description: "alignment (update skills, cli, commands), create new task but refer to 598-sdd-workflow-streamline spec and plan"

## Overview

Consolidate AI commands from 15 to 9 by removing redundant commands, renaming for clarity, and converting audit to a skill. This is **Stream 1** of a 3-stream alignment effort.

| Stream | Focus | Feature | Status |
|--------|-------|---------|--------|
| **1** | AI command consolidation | 599-alignment | **This spec** |
| 2 | Bash script → Go CLI migration | TBD | Planned |
| 3 | New CLI + skills (comment, checkpoint, research) | TBD | Future |

**Stream 1 Scope** (this spec - code changes):
- Remove 6 redundant AI commands
- Rename 1 command (analyze → verify)
- Convert 1 command to skill (audit → sl-audit)
- Update 2 commands to absorb removed functionality

**Consolidation Summary** (from 598):

| Action | Count | Commands |
|--------|-------|----------|
| **KEEP** | 8 | specify, tasks, checklist, implement, clarify, plan, onboard, constitution |
| **REMOVE** | 6 | resume, help, adopt, add-deps, remove-deps, revise (pre-deleted) |
| **RENAME** | 1 | analyze → verify |
| **→ SKILL** | 1 | audit → sl-audit skill |

**Final count**: 15 → 9 commands (revise.md was pre-deleted in commit 773a293)

**Layer Architecture** (from 598):
| Layer | Name | Runtime | Purpose |
|-------|------|---------|---------|
| 0 | Hooks | Invisible, event-driven | Auto-capture sessions on commit |
| 1 | `sl` CLI | Go binary, no AI needed | Data operations, CRUD, standalone tooling |
| 2 | AI Commands | Agent shell prompts | AI workflow orchestration (specify→implement) |
| 3 | Skills | Passive context injection | Domain knowledge, progressively loaded |

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remove Redundant Commands (Priority: P1)

As a SpecLedger user, I want redundant AI commands removed so that the command set is smaller and easier to navigate.

**Why this priority**: Reduces maintenance burden and user confusion immediately.

**Independent Test**: Run `ls .claude/commands/` and verify 6 fewer files.

**Acceptance Scenarios**:

1. **Given** `.claude/commands/`, **When** consolidation complete, **Then** `resume.md` is deleted (merged into implement)
2. **Given** `.claude/commands/`, **When** consolidation complete, **Then** `help.md` is deleted (merged into onboard)
3. **Given** `.claude/commands/`, **When** consolidation complete, **Then** `adopt.md` is deleted (context detection replaces)
4. **Given** `.claude/commands/`, **When** consolidation complete, **Then** `add-deps.md` and `remove-deps.md` are deleted (agent calls `sl deps` directly)
5. **Given** `.claude/commands/`, **When** consolidation complete, **Then** `revise.md` is deleted (absorbed by clarify)

---

### User Story 2 - Rename Analyze to Verify (Priority: P1)

As a SpecLedger user, I want the `analyze` command renamed to `verify` to align with OpenSpec terminology.

**Why this priority**: Terminology alignment improves discoverability.

**Independent Test**: Run `/specledger.verify` and verify it works; verify `/specledger.analyze` does not exist.

**Acceptance Scenarios**:

1. **Given** `specledger.analyze.md`, **When** renamed, **Then** file is `specledger.verify.md`
2. **Given** the verify command, **When** description reads, **Then** it notes successor to analyze
3. **Given** users familiar with SpecKit, **When** looking for analyze, **Then** documentation notes verify is successor

---

### User Story 3 - Convert Audit to Skill (Priority: P1)

As a SpecLedger user, I want codebase audit as a skill (not AI command) since it provides passive context, not workflow orchestration.

**Why this priority**: Proper layer placement - audit is reconnaissance context, not multi-step workflow.

**Independent Test**: Load `sl-audit` skill; verify it provides audit patterns without being an AI command.

**Acceptance Scenarios**:

1. **Given** `specledger.audit.md`, **When** converted, **Then** skill exists at `skills/sl-audit/skill.md`
2. **Given** the sl-audit skill, **When** loaded, **Then** it provides codebase reconnaissance patterns
3. **Given** `.claude/commands/`, **When** conversion complete, **Then** `specledger.audit.md` is deleted

---

### User Story 4 - Update Implement to Absorb Resume (Priority: P2)

As a SpecLedger user, I want `implement` to handle resumption so I don't need a separate `resume` command.

**Why this priority**: Consolidates related functionality; implement already handles task execution.

**Independent Test**: Start implementation, exit, run `/specledger.implement` again; verify it resumes correctly.

**Acceptance Scenarios**:

1. **Given** `specledger.implement.md`, **When** updated, **Then** it checks for in-progress tasks and resumes
2. **Given** an in-progress task, **When** implement runs, **Then** it continues from last checkpoint
3. **Given** the implement command, **When** reading, **Then** it has explicit resume behavior documented

---

### User Story 5 - Update Onboard to Absorb Help (Priority: P2)

As a SpecLedger user, I want `onboard` to include help information so I don't need a separate `help` command.

**Why this priority**: Onboarding is the natural place for workflow overview and command discovery.

**Independent Test**: Run `/specledger.onboard`; verify it includes command overview previously in help.

**Acceptance Scenarios**:

1. **Given** `specledger.onboard.md`, **When** updated, **Then** it includes command overview section
2. **Given** the onboard command, **When** running, **Then** it shows available commands with descriptions
3. **Given** the onboard command, **When** complete, **Then** user knows core workflow and available commands

---

### Edge Cases

- What if user has muscle memory for old commands? → Document migration path in AGENTS.md
- What about external docs referencing removed commands? → Add deprecation notices, update docs
- What if removed commands had unique functionality? → Verify all functionality is absorbed before deletion

---

## Clarifications

### Session 2026-03-02

- Q: What is the scope of this feature? → A: Stream 1 only - AI command consolidation (code changes)
- Q: Should output be documentation or code? → A: Code changes - remove/rename/convert/update AI commands

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 6 AI commands MUST be deleted: resume, help, adopt, add-deps, remove-deps, revise
- **FR-002**: `specledger.analyze.md` MUST be renamed to `specledger.verify.md`
- **FR-003**: `specledger.audit.md` MUST be converted to `skills/sl-audit/skill.md`
- **FR-004**: `specledger.implement.md` MUST be updated to absorb resume functionality
- **FR-005**: `specledger.onboard.md` MUST be updated to absorb help functionality
- **FR-006**: Final command count MUST be 9 (down from 15)

### Key Entities

- **AI Command**: Markdown file in `.claude/commands/` that orchestrates AI workflows
- **Skill**: Markdown file in `skills/` that provides passive domain knowledge
- **Consolidation**: Process of removing, renaming, or converting commands to reduce redundancy

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 6 command files deleted from `.claude/commands/` ✅ DONE
- **SC-002**: 1 command file renamed (analyze → verify) ✅ DONE
- **SC-003**: 1 skill created (`sl-audit`), 1 command deleted (audit) ✅ DONE
- **SC-004**: 2 commands updated (implement, onboard) - Partial (needs resume logic, command overview)
- **SC-005**: Final command count is 9 ✅ DONE (base count before 601 additions)
- **SC-006**: All removed functionality (except revise) is absorbed by remaining commands

### Implementation Notes

**Completed in this alignment:**
- Deleted: resume.md, help.md, adopt.md, add-deps.md, remove-deps.md, analyze.md, audit.md (7 files)
- Renamed: analyze.md → verify.md (verify.md already existed, deleted analyze.md)
- Created: sl-audit skill deployed to `.claude/skills/sl-audit/`
- Updated: clarify.md updated to use `sl comment list` instead of `sl revise --summary`

**Post-alignment command count**: 9 base commands + 2 from 601 (spike, checkpoint) = 11 total

### Previous work

### Epic: 598 - SDD Workflow Streamline

- **Consolidation Decisions**: Remove 7, rename 1, convert 1 (audit → skill)
- **Layer Architecture**: Four-layer model (Hooks, CLI, AI Commands, Skills)
- **Launcher Pattern**: AI commands invoke CLI, parse JSON output

## Dependencies & Assumptions

### Dependencies

- **598-sdd-workflow-streamline**: Consolidation decisions from 598 spec define what to remove/rename/convert

### Assumptions

- All functionality in removed commands can be absorbed by remaining commands
- Users will be notified of command removal via documentation
- `sl deps` CLI exists for dependency management (replacing add-deps/remove-deps)

## Out of Scope

- **Stream 2** (TBD): Bash script → Go CLI migration - `sl spec info/create/setup-plan`, `sl context update`
- **Stream 3** (TBD): New CLI commands and skills - `sl comment`, `sl checkpoint`, `sl research`
  - **Note**: `clarify.md` update to use `sl comment` is Stream 3 (depends on `sl comment` CLI)
- Changes to CLI binary (`sl`)
- Changes to the core SDD workflow (specify→clarify→plan→tasks→implement is immutable)
