# Tasks Index: Mockup Command

Issue graph index for the tasks and phases of the mockup command feature implementation.
This index does **not contain tasks directly** — those are fully managed through `sl issue`.

## Feature Tracking

* **Epic ID**: `SL-675f7d`
* **User Stories Source**: `specledger/598-mockup-command/spec.md`
* **Research Inputs**: `specledger/598-mockup-command/research.md`
* **Planning Details**: `specledger/598-mockup-command/plan.md`
* **Data Model**: `specledger/598-mockup-command/data-model.md`
* **Contract Definitions**: `specledger/598-mockup-command/contracts/`

## Issue Query Hints

```bash
# Find all open tasks for this feature
sl issue list --label spec:598-mockup-command --status open

# Find ready tasks (no blockers)
sl issue ready --label spec:598-mockup-command

# Show dependency tree for epic
sl issue show SL-675f7d

# View issues by phase
sl issue list --label "phase:setup" --label "spec:598-mockup-command"
sl issue list --label "phase:foundational" --label "spec:598-mockup-command"
sl issue list --label "phase:us1" --label "spec:598-mockup-command"
sl issue list --label "phase:us2" --label "spec:598-mockup-command"
sl issue list --label "phase:us3" --label "spec:598-mockup-command"
sl issue list --label "phase:us4" --label "spec:598-mockup-command"
sl issue list --label "phase:us5" --label "spec:598-mockup-command"
sl issue list --label "phase:polish" --label "spec:598-mockup-command"

# View issues by user story
sl issue list --label "story:US1" --label "spec:598-mockup-command"
sl issue list --label "story:US4" --label "spec:598-mockup-command"

# View issues by component
sl issue list --label "component:mockup" --label "spec:598-mockup-command"
sl issue list --label "component:cli" --label "spec:598-mockup-command"
```

## Tasks and Phases Structure

```
Epic: SL-675f7d (Mockup Command)
├── Phase 1: Setup (SL-acaa90) ─ Package + command skeleton
│   ├── SL-462fc9: Create shared types in pkg/cli/mockup/types.go
│   └── SL-da1182: Create mockup command skeleton + register in main.go
│
├── Phase 2: Foundational (SL-e6485b) ─ Detector, scanner, design system I/O
│   ├── SL-001c77: Implement frontend framework detector
│   ├── SL-9c05bf: Implement component scanner with framework-specific handlers
│   └── SL-dd3052: Implement design system file I/O with YAML frontmatter
│
├── Phase 3: US1 - Generate Mockup (SL-1e3e83) ─ P1 MVP
│   ├── SL-0b1e92: Implement spec parser
│   ├── SL-bdc5a2: Implement mockup generator with ASCII screens
│   └── SL-497332: Wire sl mockup command handler
│
├── Phase 4: US2 - Auto-Create Design System (SL-520c1b) ─ P2
│   ├── SL-e55a8e: Add auto-generation when design_system.md missing
│   └── SL-632f43: Add third-party library detection
│
├── Phase 5: US3 - Frontend Detection Flow (SL-2b6615) ─ P2
│   ├── SL-87cce5: Integrate detection into command with --force
│   └── SL-e776d8: Handle ambiguous repos and monorepos
│
├── Phase 6: US4 - Update Design System (SL-a00981) ─ P2
│   ├── SL-354a9b: Implement sl mockup update handler
│   └── SL-1033e3: Add merge logic with manual preservation
│
├── Phase 7: US5 - Init Integration (SL-b0f5d0) ─ P3
│   └── SL-078da6: Add frontend detection to bootstrap.go
│
└── Phase 8: Polish (SL-c71d1f) ─ P3
    ├── SL-bf1ecc: Add --json output to both commands
    ├── SL-e17fc9: Comprehensive error handling and edge cases
    └── SL-d47cb2: Unit tests for all domain modules
```

## Convention Summary

| Type    | Description                  | Labels                                          |
| ------- | ---------------------------- | ----------------------------------------------- |
| epic    | Full feature epic            | `spec:598-mockup-command`                       |
| feature | Implementation phase / story | `phase:<name>`, `story:<US#>`                   |
| task    | Implementation task          | `component:<x>`, `requirement:<fr-id>`          |

## Dependency Graph

```
T001 (types) ──┬──→ T003 (detector) ──┬──→ T008 (handler) ──→ T011 (detection) → T012 (ambiguous)
               │                       │         ↑                    ↑
               ├──→ T004 (scanner) ────┤    T002 (cmd) ─────────────┘
               │         │             │         │
               │         ├──→ T009 (auto-gen)    ├──→ T013 (update) → T014 (merge)
               │         │                       │
               │         └──→ T010 (3rd-party)   │
               │                                  │
               └──→ T005 (designsystem) ──┬──→ T007 (generator) → T008
                         │                │         ↑
                         │                │    T006 (parser) ──┘
                         │                │
                         ├──→ T009 (auto-gen)
                         ├──→ T013 (update handler)
                         ├──→ T014 (merge logic)
                         └──→ T015 (bootstrap)

T008 (handler) ──┬──→ T016 (JSON output)
                 └──→ T017 (error handling)
T013 (update)  ──┬──→ T016 (JSON output)
                 └──→ T017 (error handling)
```

### Parallel Execution Opportunities

**Within Phase 2 (Foundational)**: After T001 completes, T003, T004, T005 can run in parallel.

**Within Phase 3 (US1)**: T006 (parser) can run in parallel with foundational tasks.

**US2 + US4 parallelism**: After foundational completes, US2 and US4 tasks can run in parallel since they touch different parts of the flow.

**US3 + US5 parallelism**: After foundational + US1 complete, US3 (detection integration) and US5 (bootstrap integration) can run in parallel.

## Definition of Done Summary

| Issue ID   | DoD Items |
|------------|-----------|
| SL-462fc9  | types.go with all entities, FrameworkType enum, yaml+json tags, compiles |
| SL-da1182  | mockup.go with VarMockupCmd, flags, registered in main.go, help text correct |
| SL-001c77  | 3-tier detection, all frameworks, IsFrontend=false for non-frontend |
| SL-9c05bf  | Per-framework scanning, props extraction, excluded dirs skipped |
| SL-dd3052  | Load/Write/Init design system, manual markers preserved, edge case handling |
| SL-0b1e92  | ParseSpec extracts user stories + priorities, error on empty |
| SL-bdc5a2  | ASCII wireframes, component mapping, WriteMockup to markdown, P1 covered |
| SL-497332  | Full flow orchestration, correct exit codes, error messages match contracts |
| SL-e55a8e  | Auto-create on missing, scan stats output, zero-component handling |
| SL-632f43  | Import patterns for 5 UI libs, external components, deduplication |
| SL-87cce5  | Detection output, --force bypass, exit code 2 on non-frontend |
| SL-e776d8  | Ambiguous detection, backend indicators, monorepo patterns, warning |
| SL-354a9b  | Update handler, existence validation, rescan, stats output |
| SL-1033e3  | Merge with manual preservation, add/remove/unchanged stats |
| SL-078da6  | Bootstrap integration, interactive prompt, CI auto-create, skip non-frontend |
| SL-bf1ecc  | JSON output for both commands, non-JSON suppressed, jq-parseable |
| SL-e17fc9  | All edge cases handled, error messages match contracts |
| SL-d47cb2  | 4 test files, table-driven tests, all pass |

## Implementation Strategy

### MVP (Suggested: Phase 1 + 2 + 3)

The MVP delivers **User Story 1** — generating mockups from specs using an existing design system. This covers:
- `pkg/cli/mockup/` package with all types
- `sl mockup <spec-name>` command registered and functional
- Frontend detection, component scanning, design system I/O
- Spec parsing and mockup generation
- End-to-end flow producing `mockup.md`

**MVP scope**: 8 tasks (T001-T008), priority 1.

### Incremental Delivery

1. **MVP**: Phase 1-3 (Setup + Foundational + US1) — Core mockup generation
2. **Increment 1**: Phase 4-5 (US2 + US3) — Auto-create + detection flow
3. **Increment 2**: Phase 6 (US4) — Update command
4. **Increment 3**: Phase 7 (US5) — Init integration
5. **Final**: Phase 8 (Polish) — JSON output, error handling, tests

### Story Testability

| Story | Independently Testable? | Test Criteria |
|-------|------------------------|---------------|
| US1   | Yes | Run `sl mockup <spec>` with existing design_system.md → mockup.md generated |
| US2   | Yes | Run `sl mockup <spec>` without design_system.md → file auto-created, mockup generated |
| US3   | Yes | Run `sl mockup <spec>` on React/Go project → correct detection/error behavior |
| US4   | Yes | Run `sl mockup update` → design system refreshed with stats |
| US5   | Yes | Run `sl init` on frontend project → design_system.md created |

---

> This file is an index only. Implementation data lives in `sl issue`. Update this file only to point humans and agents to canonical query paths and feature references.
