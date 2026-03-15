# Tasks Index: Multi-Coding Agent Support

Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-0120ab`
* **User Stories Source**: `specledger/001-coding-agent-support/spec.md`
* **Research Inputs**: `specledger/001-coding-agent-support/research.md`
* **Planning Details**: `specledger/001-coding-agent-support/plan.md`
* **Data Model**: `specledger/001-coding-agent-support/data-model.md`
* **Quickstart Guide**: `specledger/001-coding-agent-support/quickstart.md`

## Issue Query Hints

Use the `sl issue` CLI to query and manipulate the issue graph:

```bash
# Find all open tasks for this feature
sl issue list --label spec:001-coding-agent-support --status open

# See all issues across specs
sl issue list --all --status open

# View issue details
sl issue show SL-0120ab

# Link dependencies
sl issue link [from-id] blocks [to-id]
```

## Tasks and Phases Structure

This feature follows a 2-level graph structure:

* **Epic**: SL-0120ab → Multi-Coding Agent Support
* **Phases**: Issues of type `feature`, child of the epic
  * Setup (SL-84fe08) - Agent Infrastructure
  * P1 (SL-1f5a43) - Core Agent Launch & Config (US1+US2)
  * P2 (SL-3622d5) - Multi-Agent Setup & Shared Config (US3+US4)
  * P3 (SL-c046dd) - Per-Agent Custom Arguments (US5)
  * Polish (SL-2c8774) - Cross-Cutting Concerns
* **Tasks**: Issues of type `task`, children of each feature issue

## Convention Summary

| Type    | Description                  | Labels                                 |
| ------- | ---------------------------- | -------------------------------------- |
| epic    | Full feature epic            | `spec:001-coding-agent-support`        |
| feature | Implementation phase / story | `phase:setup`, `story:US1`, etc.       |
| task    | Implementation task          | `component:launcher`, `requirement:FR-001` |

## Phase Summary

| Phase | Feature ID | Story | Priority | Tasks |
|-------|------------|-------|----------|-------|
| Setup | SL-84fe08 | Infrastructure | 0 | 3 |
| P1 | SL-1f5a43 | US1+US2 | 1 | 5 |
| P2 | SL-3622d5 | US3+US4 | 2 | 4 |
| P3 | SL-c046dd | US5 | 3 | 2 |
| Polish | SL-2c8774 | Cross-cutting | 4 | 2 |

**Total Tasks**: 16

## Definition of Done Summary

| Issue ID | Title | DoD Items |
|----------|-------|-----------|
| SL-a6c825 | Create agent registry | Agent struct defined, 4 agents registered, case-insensitive lookup |
| SL-1fbd97 | Platform detection | IsWindows(), SupportsSymlinks(), SymlinkOrCopy() |
| SL-cd1ea4 | Config schema | agent.default, per-agent arguments/env patterns |
| SL-a5e026 | sl code command | Command registered, optional arg, defaults to config |
| SL-fc596e | Extend launcher | Accept Agent struct, env vars injected, backward compatible |
| SL-e74205 | Binary detection | CheckAgentInstalled(), install command in error |
| SL-78b826 | Per-agent config | GetAgentArguments(), project overrides global |
| SL-9abf2a | Wire command to launcher | Registry lookup, binary check, config merge |
| SL-c02b9e | TUI multi-select | Checkboxes, 4 agents, space to toggle |
| SL-dc12d7 | .agent directory | CreateAgentSharedDir(), migration, --force |
| SL-dd4d0e | Symlinks/copies | LinkAgentToShared(), platform handling |
| SL-67d7a1 | Store selected agents | Constitution field, human-readable |
| SL-419696 | Verify arguments | Multiple flags, quotes, special chars |
| SL-53773a | Verify env vars | Single/multiple env vars, spaces in values |
| SL-1ba826 | Update docs | README section, config examples, links |
| SL-dae1b5 | Validate quickstart | All scenarios verified |

## MVP Scope

**Suggested MVP**: Complete Setup + P1 (US1+US2) for basic agent launch functionality.

This delivers:
- `sl code` command to launch any coding agent
- Per-agent configuration via `sl config`
- Error messages with install instructions

## Execution Order

```
Setup (SL-84fe08) ──┬──► P1 (SL-1f5a43) ──┬──► Polish (SL-2c8774)
                    │                      │
                    ├──► P2 (SL-3622d5) ───┤
                    │                      │
                    └──► P3 (SL-c046dd) ───┘
```

P1, P2, and P3 can proceed in parallel after Setup completes.

## Status Tracking

Status is tracked in the issue store:

* **Open** → default
* **In Progress** → task being worked on
* **Closed** → complete

Use `sl issue list --status open --label spec:001-coding-agent-support` to query progress.

---

> This file is intentionally light and index-only. Implementation data lives in the issue store.
