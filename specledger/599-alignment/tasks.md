# Tasks: SDD Layer Alignment

**Epic**: SL-6cf43d - SDD Layer Alignment
**Feature**: 599-alignment
**Created**: 2026-03-02

## Overview

Consolidate AI commands from 16 to 11 by removing redundant commands, renaming for clarity, and converting audit to a skill.

**Scope**: Stream 1 - AI command consolidation (code changes)

## Issue Hierarchy

```
SL-6cf43d (Epic: SDD Layer Alignment)
├── SL-add30a (US1: Delete Redundant Commands) [P1]
│   ├── SL-6f5bf7 Delete specledger.resume.md
│   ├── SL-265810 Delete specledger.help.md
│   ├── SL-284d88 Delete specledger.adopt.md
│   ├── SL-91c1c7 Delete add-deps.md and remove-deps.md
│   └── SL-721fa7 Delete specledger.revise.md
├── SL-fe0acd (US2: Rename Analyze to Verify) [P1]
│   ├── SL-00fe17 Rename analyze.md to verify.md
│   └── SL-b802e8 Update verify.md description
├── SL-32d9fb (US3: Convert Audit to Skill) [P1]
│   ├── SL-2e8168 Create sl-audit skill
│   └── SL-3137ed Delete specledger.audit.md
├── SL-e06e81 (US4: Update Implement) [P2]
│   └── SL-804592 Add resume logic to implement
├── SL-a035df (US5: Update Onboard) [P2]
│   └── SL-0d634a Add command overview to onboard
```

## Execution Phases

### Phase 1: Delete Redundant Commands (US1) [P1]

**Feature**: SL-add30a

| Task ID | Title | Status | Parallel |
|---------|-------|--------|----------|
| SL-6f5bf7 | Delete specledger.resume.md | open | ✓ |
| SL-265810 | Delete specledger.help.md | open | ✓ |
| SL-284d88 | Delete specledger.adopt.md | open | ✓ |
| SL-91c1c7 | Delete add-deps.md and remove-deps.md | open | ✓ |
| SL-721fa7 | Delete specledger.revise.md | open | ✓ |

**Independent Test**: `ls .claude/commands/ | wc -l` shows 10 files (down from 15)

### Phase 2: Rename Analyze to Verify (US2) [P1]

**Feature**: SL-fe0acd
**Depends on**: Phase 1

| Task ID | Title | Status | Dependencies |
|---------|-------|--------|--------------|
| SL-00fe17 | Rename analyze.md to verify.md | open | Phase 1 |
| SL-b802e8 | Update verify.md description | open | SL-00fe17 |

**Independent Test**: `/specledger.verify` works, `/specledger.analyze` does not exist

### Phase 3: Convert Audit to Skill (US3) [P1]

**Feature**: SL-32d9fb
**Depends on**: Phase 2

| Task ID | Title | Status | Dependencies |
|---------|-------|--------|--------------|
| SL-2e8168 | Create sl-audit skill | open | Phase 2 |
| SL-3137ed | Delete specledger.audit.md | open | SL-2e8168 |

**Independent Test**: `skills/sl-audit/skill.md` exists, `specledger.audit.md` does not

### Phase 4-5: Update Commands (US4, US5) [P2]

**Features**: SL-e06e81, SL-a035df
**Depends on**: Phase 3
**Parallel**: Yes - US4, US5 can run in parallel

| Task ID | Title | Status | Dependencies |
|---------|-------|--------|--------------|
| SL-804592 | Add resume logic to implement | open | SL-6f5bf7 |
| SL-0d634a | Add command overview to onboard | open | SL-265810 |

**Independent Tests**:
- US4: `/specledger.implement` resumes in-progress tasks
- US5: `/specledger.onboard` shows command overview

## Dependency Graph

```
Phase 1 (US1) ─┬─ SL-6f5bf7 ──────────────────────► SL-804592 (US4)
               ├─ SL-265810 ──────────────────────► SL-0d634a (US5)
               ├─ SL-284d88
               ├─ SL-91c1c7
               └─ SL-721fa7
                        │
                        ▼
               Phase 2 (US2) ── SL-00fe17 ─► SL-b802e8
                        │
                        ▼
               Phase 3 (US3) ── SL-2e8168 ─► SL-3137ed
```

## Query Commands

```bash
# View all issues for this spec
sl issue list --label "spec:599-alignment"

# View open issues
sl issue list --status open --label "spec:599-alignment"

# View by phase
sl issue list --label "phase:us1"
sl issue list --label "phase:us2"
sl issue list --label "phase:us3"

# View ready-to-work (no blockers)
sl issue ready

# View epic tree
sl issue show SL-6cf43d
```

## Definition of Done Summary

| Issue ID | DoD Items |
|----------|-----------|
| SL-6f5bf7 | specledger.resume.md deleted, Git shows file removed |
| SL-265810 | specledger.help.md deleted, Git shows file removed |
| SL-284d88 | specledger.adopt.md deleted, Git shows file removed |
| SL-91c1c7 | add-deps.md deleted, remove-deps.md deleted, Git shows both removed |
| SL-721fa7 | specledger.revise.md deleted, Git shows file removed |
| SL-00fe17 | verify.md exists, analyze.md does not, Git shows rename |
| SL-b802e8 | Description updated with successor note, OpenSpec mentioned |
| SL-2e8168 | skill.md created, When to Load section, Key Concepts, Decision Patterns |
| SL-3137ed | audit.md deleted, Git shows removed, ls returns 11 |
| SL-804592 | In-progress check, Resume prompt, Checkpoint logic, Behavior rules |
| SL-0d634a | Command Overview section, Core commands, Utility commands, Descriptions |

## Success Criteria

- [ ] SC-001: 6 command files deleted from `.claude/commands/`
- [ ] SC-002: 1 command file renamed (analyze → verify)
- [ ] SC-003: 1 skill created (`sl-audit`), 1 command deleted (audit)
- [ ] SC-004: 2 commands updated (implement, onboard)
- [ ] SC-005: Final command count is 9
- [ ] SC-006: All removed functionality (except revise) is absorbed by remaining commands

## MVP Scope

**MVP = US1 + US2 + US3** (all P1 stories)

This delivers:
- 6 redundant commands deleted
- analyze renamed to verify
- audit converted to skill
- Final count: 11 commands

**Post-MVP = US4 + US5 + US6** (P2 stories)

These add enhanced functionality but are not required for the consolidation to be complete.

## References

- [spec.md](spec.md) - Feature specification
- [plan.md](plan.md) - Implementation plan
- [research.md](research.md) - Consolidation analysis
