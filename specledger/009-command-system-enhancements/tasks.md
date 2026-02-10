# Tasks Index: Command System Enhancements

Beads Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through Beads CLI.

**Status**: CLOSED (all tasks completed - documents existing changes)

## Feature Tracking

* **Beads Epic ID**: `SL-13u`
* **User Stories Source**: `specledger/009-command-system-enhancements/spec.md`
* **Research Inputs**: N/A (documents existing changes)
* **Planning Details**: `specledger/009-command-system-enhancements/plan.md`
* **Data Model**: N/A (no data entities)
* **Contract Definitions**: N/A (no API contracts)

## Beads Query Hints

Use the `bd` CLI to query and manipulate the issue graph:

```bash
# Find all tasks for this feature
bd list --label spec:009-command-system-enhancements --limit 30

# Find tasks by story
bd list --label spec:009-command-system-enhancements --label story:US1

# See dependencies for epic
bd dep tree SL-13u

# View issues by component
bd list --label component:commands --label spec:009-command-system-enhancements

# Show all phases
bd list --type feature --label spec:009-command-system-enhancements
```

## Tasks and Phases Structure

```
Epic: SL-13u (Command System Enhancements) CLOSED
├── Feature: SL-if7 (Setup: Path Standardization) CLOSED
│   ├── Task: SL-ftu (Update adopt-feature-branch.sh paths) CLOSED
│   ├── Task: SL-awv (Update create-new-feature.sh paths) CLOSED
│   ├── Task: SL-dnd (Update common.sh paths) CLOSED
│   ├── Task: SL-7pi (Update setup-plan.sh paths) CLOSED
│   └── Task: SL-eak (Update update-agent-context.sh paths) CLOSED
├── Feature: SL-96v (US1: Help Command) CLOSED
│   └── Task: SL-ttc (Create specledger.help.md) CLOSED
├── Feature: SL-ky5 (US2: Audit Command) CLOSED
│   └── Task: SL-c1n (Create specledger.audit.md) CLOSED
├── Feature: SL-2x9 (US3: Revise Command) CLOSED
│   └── Task: SL-zig (Create specledger.revise.md) CLOSED
├── Feature: SL-mye (US4: Implement Sync) CLOSED
│   └── Task: SL-0tv (Add Supabase sync) CLOSED
├── Feature: SL-7o3 (US5: Adopt from Audit) CLOSED
│   └── Task: SL-tku (Add --from-audit mode) CLOSED
├── Feature: SL-6t7 (Enhanced: Purpose Sections) CLOSED
│   └── Task: SL-7kz (Add Purpose sections to 8 commands) CLOSED
├── Feature: SL-4v4 (Utility Scripts) CLOSED
│   ├── Task: SL-b5g (Create pull-issues.js) CLOSED
│   └── Task: SL-abo (Create review-comments.js) CLOSED
└── Task: SL-pl5 (Simplify AGENTS.md) CLOSED
```

## Phase Summary

| Phase | Feature ID | Description | Tasks | Status |
|-------|------------|-------------|-------|--------|
| Setup | SL-if7 | Path Standardization | 5 | CLOSED |
| US1 (P1) | SL-96v | Help Command | 1 | CLOSED |
| US2 (P1) | SL-ky5 | Audit Command | 1 | CLOSED |
| US3 (P2) | SL-2x9 | Revise Command | 1 | CLOSED |
| US4 (P2) | SL-mye | Implement Sync | 1 | CLOSED |
| US5 (P3) | SL-7o3 | Adopt from Audit | 1 | CLOSED |
| Enhanced | SL-6t7 | Purpose Sections | 1 | CLOSED |
| Scripts | SL-4v4 | Utility Scripts | 2 | CLOSED |

## User Story Mapping

| Story | Priority | Feature ID | Requirements | Status |
|-------|----------|------------|--------------|--------|
| US1: Help Command | P1 | SL-96v | FR-001 | CLOSED |
| US2: Audit Command | P1 | SL-ky5 | FR-002 | CLOSED |
| US3: Revise Command | P2 | SL-2x9 | FR-003 | CLOSED |
| US4: Implement Sync | P2 | SL-mye | FR-004 | CLOSED |
| US5: Adopt from Audit | P3 | SL-7o3 | FR-005 | CLOSED |

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

- CLOSED **MVP (US1+US2)**: Help and Audit commands
- CLOSED **P2 Stories (US3+US4)**: Revise command and Implement sync
- CLOSED **P3 Stories (US5)**: Adopt from audit mode
- CLOSED **Supporting Work**: Purpose sections, utility scripts, path fixes

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
