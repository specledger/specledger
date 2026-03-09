# Tasks Index: SDD Workflow Streamline

This index provides query hints for navigating the issue graph for this feature implementation.
All tasks are managed through `sl issue` CLI. This file does **not** contain tasks directly.

## Feature Tracking

* **Epic ID**: `SL-e5ae24`
* **User Stories Source**: `specledger/598-sdd-workflow-streamline/spec.md`
* **Research Inputs**: `specledger/598-sdd-workflow-streamline/research.md`
* **Planning Details**: `specledger/598-sdd-workflow-streamline/plan.md`
* **Data Model**: `specledger/598-sdd-workflow-streamline/data-model.md`
* **Contract Definitions**: `specledger/598-sdd-workflow-streamline/contracts/`

## Quick Start

```bash
# View the epic and all children
sl issue show SL-e5ae24

# List all issues for this spec
sl issue list --label spec:598-sdd-workflow-streamline --status all

# Find ready tasks (no blocking dependencies)
sl issue list --label spec:598-sdd-workflow-streamline --status open --limit 10
```

## Phase Overview

| Phase | Issue ID | Priority | Blocked By | Description |
|-------|----------|----------|------------|-------------|
| **Setup** | SL-391a76 | P1 | - | Project structure and package scaffolding |
| **Foundational** | SL-a2173e | P1 | Setup | Shared infrastructure (client extraction, utilities) |
| **US1** | SL-c8facb | P1 | Foundational | Inventory and classify workflow components |
| **US6** | SL-fb2567 | P1 | Foundational | Spike AI command |
| **US7** | SL-f31153 | P1 | Foundational | Checkpoint AI command |
| **US11** | SL-8f8d2d | P1 | Foundational | sl comment CLI (list/show/reply/resolve) |
| **US16** | SL-c49ce7 | P1 | Foundational | sl spec + sl context CLI (bash replacement) |
| **US2** | SL-9e40e2 | P2 | - | Consolidate dependency commands |
| **US3** | SL-bac864 | P2 | - | Verify command + sl-audit skill |
| **US4** | SL-e4d9aa | P2 | US11 | Skills with progressive loading |
| **US8** | SL-93dfa7 | P2 | - | sl doctor --template |
| **US9** | SL-3ae06f | P2 | - | Update CLI README |
| **US12** | SL-95bda1 | P2 | - | Context detection fallback chain |
| **US13** | SL-8d0c29 | P2 | US16 | Phase out bash scripts |
| **US14** | SL-61ba32 | P2 | US11 | Clarify absorbs revise |
| **US5** | SL-632335 | P3 | - | Document consolidated workflow |
| **US10** | SL-0dea96 | P3 | - | CHANGELOG for templates |
| **US15** | SL-462774 | P3 | - | Session lifecycle hooks |
| **Polish** | SL-ffac2e | P3 | - | Cross-cutting concerns |

## Dependency Graph

```
Setup (SL-391a76)
    └── Foundational (SL-a2173e)
            ├── US1 (SL-c8facb) - Inventory
            ├── US6 (SL-fb2567) - Spike
            ├── US7 (SL-f31153) - Checkpoint
            ├── US11 (SL-8f8d2d) - sl comment CLI ──┬── US4 (SL-e4d9aa) - Skills
            │                                       └── US14 (SL-61ba32) - Clarify
            └── US16 (SL-c49ce7) - sl spec/context ──── US13 (SL-8d0c29) - Phase out bash

US2 (SL-9e40e2), US3 (SL-bac864), US8 (SL-93dfa7), US9 (SL-3ae06f), US12 (SL-95bda1)
    → Independent P2 phases (no blocking dependencies)

US5 (SL-632335), US10 (SL-0dea96), US15 (SL-462774), Polish (SL-ffac2e)
    → P3 phases (lowest priority)
```

## Query Hints by Phase

### Setup Phase
```bash
sl issue list --label phase:setup --status open
sl issue show SL-391a76
```

### Foundational Phase
```bash
sl issue list --label phase:foundational --status open
sl issue show SL-a2173e
```

### User Story 11 - sl comment CLI (Critical Path)
```bash
sl issue list --label story:US11 --status open
sl issue show SL-8f8d2d
```

### User Story 16 - Bash Script Replacement (Critical Path)
```bash
sl issue list --label story:US16 --status open
sl issue show SL-c49ce7
```

### All P1 Tasks
```bash
sl issue list --label spec:598-sdd-workflow-streamline --status open --priority 1
```

### All P2 Tasks
```bash
sl issue list --label spec:598-sdd-workflow-streamline --status open --priority 2
```

## Task Summary

| Type | Count | Description |
|------|-------|-------------|
| Epic | 1 | Top-level feature container |
| Feature (Phases) | 19 | User story implementations + setup/foundational/polish |
| Task | 38 | Individual implementation tasks |
| **Total** | **58** | All issues |

### Tasks by User Story

| Story | Priority | Tasks | Key Deliverables |
|-------|----------|-------|------------------|
| Setup | P1 | 3 | Package scaffolding (comment, spec, context) |
| Foundational | P1 | 5 | Client extraction, utilities, branch naming |
| US1 | P1 | 2 | Component inventory, CLI constitution |
| US6 | P1 | 1 | Spike AI command template |
| US7 | P1 | 1 | Checkpoint AI command template |
| US11 | P1 | 5 | sl comment list/show/reply/resolve + wire CLI |
| US16 | P1 | 5 | sl spec info/create/setup-plan, sl context update + wire CLI |
| US2 | P2 | 1 | sl deps graph |
| US3 | P2 | 2 | verify rename, sl-audit skill |
| US4 | P2 | 2 | sl-comment skill, update existing skills |
| US8 | P2 | 2 | --template flag, stale detection |
| US9 | P2 | 1 | README update |
| US12 | P2 | 2 | yaml alias, git heuristic |
| US13 | P2 | 1 | Update AI templates to use sl CLI |
| US14 | P2 | 2 | Update clarify, remove revise template |
| US5 | P3 | 1 | Workflow documentation |
| US10 | P3 | 1 | CHANGELOG for templates |
| US15 | P3 | 1 | Session hooks documentation |

## Parallel Execution Opportunities

The following phases can be worked in parallel after Foundational completes:

**Parallel Group A (P1 - Critical Path)**:
- US1 (Inventory) - Documentation only
- US6 (Spike) - AI command template
- US7 (Checkpoint) - AI command template
- US11 (sl comment) - CLI implementation
- US16 (sl spec/context) - CLI implementation

**Parallel Group B (P2 - Independent)**:
- US2, US3, US8, US9, US12 - No cross-dependencies

**Parallel Group C (P2 - After US11)**:
- US4 (Skills) - Depends on US11
- US14 (Clarify) - Depends on US11

**Parallel Group D (P2 - After US16)**:
- US13 (Phase out bash) - Depends on US16

## MVP Scope

**Recommended MVP** (P1 stories only):
1. Setup → Foundational (blocking)
2. US11 (sl comment CLI) - Enables comment management
3. US16 (sl spec/context CLI) - Enables cross-platform support

This delivers the two core work streams defined in plan.md:
- Comment CRUD extraction
- Bash script replacement

## Definition of Done Summary

All tasks include Definition of Done items. View with:
```bash
sl issue show <issue-id>
```

Example:
```bash
sl issue show SL-f51c06  # sl comment list task
```

## Agent Execution Notes

1. **Assume packages exist** after Setup phase completes
2. **US11 and US16 are independent** - can be worked in parallel
3. **US4 and US14 wait for US11** - need sl comment CLI first
4. **US13 waits for US16** - need sl spec/context CLI first
5. **P3 phases are polish** - can be deferred

## Labels Reference

| Label | Purpose |
|-------|---------|
| `spec:598-sdd-workflow-streamline` | All issues in this feature |
| `phase:setup` | Setup phase tasks |
| `phase:foundational` | Foundational phase tasks |
| `phase:us11` | User Story 11 phase |
| `story:US11` | Tasks for US11 |
| `component:cli` | CLI implementation tasks |
| `component:commands` | AI command template tasks |
| `component:skills` | Skill file tasks |
| `component:docs` | Documentation tasks |
| `requirement:FR-XXX` | Functional requirement traceability |
