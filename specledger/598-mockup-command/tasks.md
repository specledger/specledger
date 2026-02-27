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
sl issue list --label "phase:shared-infra" --label "spec:598-mockup-command"
sl issue list --label "phase:setup" --label "spec:598-mockup-command"
sl issue list --label "phase:domain" --label "spec:598-mockup-command"
sl issue list --label "phase:prompt" --label "spec:598-mockup-command"
sl issue list --label "phase:interactive" --label "spec:598-mockup-command"
sl issue list --label "phase:update" --label "spec:598-mockup-command"
sl issue list --label "phase:init" --label "spec:598-mockup-command"
sl issue list --label "phase:polish" --label "spec:598-mockup-command"

# View issues by user story
sl issue list --label "story:US1" --label "spec:598-mockup-command"
sl issue list --label "story:US4" --label "spec:598-mockup-command"

# View issues by component
sl issue list --label "component:mockup" --label "spec:598-mockup-command"
sl issue list --label "component:prompt" --label "spec:598-mockup-command"
sl issue list --label "component:cli" --label "spec:598-mockup-command"
```

## Tasks and Phases Structure

```
Epic: SL-675f7d (Mockup Command)
├── Phase 1: Shared Infrastructure (SL-TBD) ─ Extract prompt/editor from revise
│   ├── T000: Extract shared editor utilities into pkg/cli/prompt/editor.go
│   │         (DetectEditor, EditPrompt moved from revise/editor.go)
│   ├── T000b: Extract shared prompt utilities into pkg/cli/prompt/prompt.go
│   │          (RenderTemplate, EstimateTokens, PrintTokenWarnings)
│   └── T000c: Refactor revise to delegate to pkg/cli/prompt/
│              (revise/editor.go → thin wrapper, revise/prompt.go → delegates)
│
├── Phase 2: Setup (SL-acaa90) ─ Package + command skeleton
│   ├── T001: Create shared types in pkg/cli/mockup/types.go
│   │         (FrameworkType, Component, DesignSystem, MockupPromptContext,
│   │          PromptComponent, SpecContent, MockupFormat)
│   └── T002: Create mockup command skeleton + register in main.go
│             (VarMockupCmd with --format, --force, --dry-run, --summary, --json flags)
│
├── Phase 3: Domain Logic (SL-e6485b) ─ Detector, scanner, design system, spec parser
│   ├── T003: Implement frontend framework detector
│   ├── T004: Implement component scanner with framework-specific handlers
│   ├── T005: Implement design system file I/O with YAML frontmatter
│   └── T006: Implement spec parser (specparser.go)
│             (Parse spec.md into SpecContent: title, user stories, requirements)
│
├── Phase 4: Prompt & Template (SL-TBD) ─ Prompt builder + template + golden tests
│   ├── T007: Create mockup prompt template (prompt.tmpl)
│   │         (Agent instructions with spec content, components, framework, format)
│   ├── T008: Implement MockupPromptContext builder (mockup/prompt.go)
│   │         (Assemble context from gathered data, render template)
│   └── T009: Add golden file tests for prompt generation
│             (Verify prompt output matches expected for various contexts)
│
├── Phase 5: Interactive Flow (SL-TBD) ─ Full 10-step wiring in mockup.go
│   ├── T010: Wire spec resolution (arg / branch detection / picker)
│   │         (Reuse issues.NewContextDetector, fallback to huh.Select)
│   ├── T011: Wire framework detection with interactive confirmation
│   │         (lipgloss display, huh.Confirm, --force bypass)
│   ├── T012: Wire design system check/generate with prompts
│   │         (Auto-generate flow, component count display)
│   ├── T013: Wire component multi-select and format selection
│   │         (huh.MultiSelect for components, huh.Select for format)
│   ├── T014: Wire prompt generation → editor review → action menu
│   │         (Uses pkg/cli/prompt/editor.go, Launch/Re-edit/Write/Cancel)
│   ├── T015: Wire agent launch with fallback
│   │         (launcher.LaunchWithPrompt, writePromptToFile fallback)
│   └── T016: Wire post-agent commit/push flow
│             (Reuse stagingAndCommitFlow pattern from revise.go)
│
├── Phase 6: Update Command (SL-a00981) ─ sl mockup update handler
│   ├── T017: Implement sl mockup update handler with interactive confirm
│   └── T018: Add merge logic with manual preservation
│
├── Phase 7: Init Integration (SL-b0f5d0) ─ bootstrap.go
│   └── T019: Add frontend detection to bootstrap.go
│
└── Phase 8: Polish (SL-c71d1f) ─ JSON, errors, tests
    ├── T020: Add --json non-interactive path for both commands
    ├── T021: Comprehensive error handling and edge cases
    │         (Agent not found, mockup exists, user cancel, malformed design system)
    └── T022: Unit tests for all domain modules
```

## Convention Summary

| Type    | Description                  | Labels                                          |
| ------- | ---------------------------- | ----------------------------------------------- |
| epic    | Full feature epic            | `spec:598-mockup-command`                       |
| feature | Implementation phase / story | `phase:<name>`, `story:<US#>`                   |
| task    | Implementation task          | `component:<x>`, `requirement:<fr-id>`          |

## Dependency Graph

```
T000 (shared editor) ──┬──→ T000b (shared prompt) ──→ T000c (revise refactor)
                       │                                      │
                       └──────────────────────────────────────┼──→ T014 (editor/action menu)
                                                              │
T001 (types) ──┬──→ T003 (detector) ──┬──→ T011 (detection confirm)
               │                       │
               ├──→ T004 (scanner) ────┤
               │         │             │
               │         └──→ T012 (design system flow)
               │                       │
               ├──→ T005 (designsystem)┤
               │         │             │
               │         ├──→ T017 (update handler) → T018 (merge)
               │         └──→ T019 (bootstrap)
               │
               ├──→ T006 (specparser) ──→ T008 (prompt builder)
               │                               │
               └──→ T002 (cmd skeleton)        │
                       │                       │
                       └───────────────────────┴──→ T010 (spec resolution)
                                                        │
T007 (prompt.tmpl) ──→ T008 (prompt builder) ──→ T009 (golden tests)
                              │
                              └──→ T013 (component/format select)
                                        │
                                        └──→ T014 (editor/action menu)
                                                  │
                                                  └──→ T015 (agent launch)
                                                            │
                                                            └──→ T016 (commit/push)

T010-T016 (interactive flow) ──┬──→ T020 (JSON output)
                               └──→ T021 (error handling)
T017-T018 (update)           ──┬──→ T020 (JSON output)
                               └──→ T021 (error handling)
T020 + T021                    ──→ T022 (all tests)
```

### Parallel Execution Opportunities

**Within Phase 1 (Shared Infra)**: T000 and T000b can run in parallel. T000c depends on both.

**Phase 2 + Phase 3**: After T001 completes, T003, T004, T005, T006 can run in parallel. T002 (cmd skeleton) can run in parallel with domain logic.

**Phase 4**: T007 (template) can start as soon as types are defined. T008 depends on T006 + T007.

**Phase 5**: T010-T016 must be sequential (they wire the interactive flow steps in order), but Phase 5 can start as soon as Phases 1-4 complete.

**Phase 6 + Phase 7**: Can run in parallel with Phase 5 (they touch different parts of the flow).

**Phase 8**: Depends on all prior phases.

## Definition of Done Summary

| Task     | DoD Items |
|----------|-----------|
| T000     | editor.go in pkg/cli/prompt/ with DetectEditor, EditPrompt; compiles; tests pass |
| T000b    | prompt.go in pkg/cli/prompt/ with RenderTemplate, EstimateTokens; compiles; tests pass |
| T000c    | revise/editor.go and revise/prompt.go delegate to prompt package; all existing revise tests still pass |
| T001     | types.go with all entities including MockupPromptContext, PromptComponent, SpecContent; compiles |
| T002     | mockup.go with VarMockupCmd, all flags (--format, --force, --dry-run, --summary, --json), registered in main.go, help text matches contract |
| T003     | 3-tier detection, all frameworks, IsFrontend=false for non-frontend |
| T004     | Per-framework scanning, props extraction, excluded dirs skipped |
| T005     | Load/Write/Init design system, manual markers preserved, edge case handling |
| T006     | ParseSpec extracts title, user stories, requirements; error on empty; SpecContent populated |
| T007     | prompt.tmpl with agent instructions template; embedded in binary |
| T008     | BuildMockupPrompt assembles context and renders template; prompt string output |
| T009     | Golden file tests verify prompt output for React HTML, React JSX, Vue HTML, empty components |
| T010     | Spec resolved from arg, branch, or picker; correct error for missing spec |
| T011     | Framework displayed with lipgloss, confirmed with huh.Confirm, --force bypasses |
| T012     | Design system loaded or auto-generated with interactive prompts |
| T013     | Component multi-select and format select work; selections stored in context |
| T014     | Editor opens, action menu works (all 4 options), uses shared prompt package |
| T015     | Agent launched if available, prompt written to file if not, install instructions shown |
| T016     | Post-agent commit/push flow works with file multi-select and push |
| T017     | Update handler validates existence, rescans with confirm, displays stats |
| T018     | Merge preserves manual entries, add/remove/unchanged stats correct |
| T019     | Bootstrap integration, interactive prompt, CI auto-create, skip non-frontend |
| T020     | JSON output for both commands, non-JSON suppressed, jq-parseable |
| T021     | All edge cases handled per contract error messages |
| T022     | All test files pass, table-driven tests, coverage on domain modules |

## Implementation Strategy

### MVP (Suggested: Phase 1 + 2 + 3 + 4 + 5)

The MVP delivers **User Story 1** — the full interactive mockup flow from spec to agent-generated mockup. This covers:
- Shared `pkg/cli/prompt/` package (extracted from revise)
- `pkg/cli/mockup/` package with all types, domain logic, and prompt builder
- `sl mockup [spec-name]` command with complete 10-step interactive flow
- Spec resolution, framework detection, design system, component selection, prompt, agent launch, commit/push
- `--dry-run` flag for non-agent workflow

**MVP scope**: T000-T016 (16 tasks), priority 1.

### Incremental Delivery

1. **MVP**: Phase 1-5 (Shared Infra + Setup + Domain + Prompt + Interactive Flow) — Full interactive mockup generation
2. **Increment 1**: Phase 6 (Update Command) — `sl mockup update`
3. **Increment 2**: Phase 7 (Init Integration) — `sl init` design system setup
4. **Final**: Phase 8 (Polish) — JSON output, error handling, comprehensive tests

### Story Testability

| Story | Independently Testable? | Test Criteria |
|-------|------------------------|---------------|
| US1   | Yes | Run `sl mockup` on feature branch → interactive flow completes, agent generates mockup |
| US2   | Yes | Run `sl mockup` without design_system.md → prompted to generate, file created, flow continues |
| US3   | Yes | Run `sl mockup` on React/Go project → correct detection displayed, confirm/error behavior |
| US4   | Yes | Run `sl mockup update` → design system refreshed with interactive confirm + stats |
| US5   | Yes | Run `sl init` on frontend project → design_system.md created |

---

> This file is an index only. Implementation data lives in `sl issue`. Update this file only to point humans and agents to canonical query paths and feature references.
