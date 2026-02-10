# Tasks Index: Command System Enhancements

Beads Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through Beads CLI.

**Status**: ✅ All tasks completed (existing changes documented)

## Feature Tracking

* **Beads Epic ID**: `SL-31n`
* **User Stories Source**: `specledger/009-command-system-enhancements/spec.md`
* **Research Inputs**: N/A (documents existing changes)
* **Planning Details**: `specledger/009-command-system-enhancements/plan.md`
* **Data Model**: N/A (no data entities)
* **Contract Definitions**: N/A (no API contracts)

## Beads Query Hints

Use the `bd` CLI to query and manipulate the issue graph:

```bash
# Find all tasks for this feature
bd list --label spec:009-command-system-enhancements --limit 20

# Find tasks by story
bd list --label spec:009-command-system-enhancements --label story:US1

# See dependencies for epic
bd dep tree SL-31n

# View issues by component
bd list --label component:commands --label spec:009-command-system-enhancements

# Show all phases
bd list --type feature --label spec:009-command-system-enhancements
```

## Tasks and Phases Structure

```
Epic: SL-31n (Command System Enhancements) ✅ CLOSED
├── Feature: SL-nid (Setup: Path Standardization) ✅ CLOSED
│   ├── Task: SL-bhr (Update adopt-feature-branch.sh) ✅ CLOSED
│   ├── Task: SL-qqk (Update create-new-feature.sh) ✅ CLOSED
│   ├── Task: SL-p3l (Update common.sh) ✅ CLOSED
│   ├── Task: SL-9j2 (Update setup-plan.sh) ✅ CLOSED
│   └── Task: SL-k0r (Update update-agent-context.sh) ✅ CLOSED
├── Feature: SL-irz (US1: Help Command) ✅ CLOSED
│   └── Task: SL-qlo (Create specledger.help.md) ✅ CLOSED
├── Feature: SL-s7x (US2: Audit Command) ✅ CLOSED
│   └── Task: SL-cu2 (Create specledger.audit.md) ✅ CLOSED
├── Feature: SL-cq8 (US3: Revise Command) ✅ CLOSED
│   └── Task: SL-vxp (Create specledger.revise.md) ✅ CLOSED
├── Feature: SL-5lx (US4: Implement Sync) ✅ CLOSED
│   └── Task: SL-2h2 (Add Supabase sync) ✅ CLOSED
├── Feature: SL-jrt (US5: Adopt from Audit) ✅ CLOSED
│   └── Task: SL-b73 (Add --from-audit mode) ✅ CLOSED
├── Feature: SL-gg1 (Enhanced Commands) ✅ CLOSED
│   └── Task: SL-2hw (Add Purpose sections) ✅ CLOSED
├── Feature: SL-3kq (Utility Scripts) ✅ CLOSED
│   ├── Task: SL-f88 (Create pull-issues.js) ✅ CLOSED
│   └── Task: SL-agk (Create review-comments.js) ✅ CLOSED
└── Task: SL-6tp (Simplify AGENTS.md) ✅ CLOSED
```

## Phase Summary

| Phase | Feature ID | Description | Tasks | Status |
|-------|------------|-------------|-------|--------|
| Setup | SL-nid | Path Standardization | 5 | ✅ Done |
| US1 (P1) | SL-irz | Help Command | 1 | ✅ Done |
| US2 (P1) | SL-s7x | Audit Command | 1 | ✅ Done |
| US3 (P2) | SL-cq8 | Revise Command | 1 | ✅ Done |
| US4 (P2) | SL-5lx | Implement Sync | 1 | ✅ Done |
| US5 (P3) | SL-jrt | Adopt from Audit | 1 | ✅ Done |
| Enhanced | SL-gg1 | Purpose Sections | 1 | ✅ Done |
| Scripts | SL-3kq | Utility Scripts | 2 | ✅ Done |

## User Story Mapping

| Story | Priority | Feature ID | Requirements | Status |
|-------|----------|------------|--------------|--------|
| US1: Help Command | P1 | SL-irz | FR-001 | ✅ Done |
| US2: Audit Command | P1 | SL-s7x | FR-002 | ✅ Done |
| US3: Revise Command | P2 | SL-cq8 | FR-003 | ✅ Done |
| US4: Implement Sync | P2 | SL-5lx | FR-004 | ✅ Done |
| US5: Adopt from Audit | P3 | SL-jrt | FR-005 | ✅ Done |

## Implementation Statistics

| Metric | Count |
|--------|-------|
| Total Epic | 1 |
| Total Features (Phases) | 8 |
| Total Tasks | 14 |
| Tasks Completed | 14 |
| Completion Rate | 100% |

## MVP Scope

Since this documents **existing changes**, all work is already complete:

- ✅ **MVP (US1+US2)**: Help and Audit commands
- ✅ **P2 Stories (US3+US4)**: Revise command and Implement sync
- ✅ **P3 Stories (US5)**: Adopt from audit mode
- ✅ **Supporting Work**: Purpose sections, utility scripts, path fixes

## Verification Checklist

To verify the changes work correctly:

- [ ] Run `/specledger.help` - displays categorized commands
- [ ] Run `/specledger.audit` - detects tech stack and modules
- [ ] Run `/specledger.revise` - fetches review comments (requires login)
- [ ] Run `/specledger.implement` - syncs issues before starting
- [ ] Run `/specledger.adopt --from-audit` - uses cached audit data

---

> This file is intentionally an index-only document. Implementation data lives in Beads.
> All tasks are closed as the feature documents existing changes already in the diff.
