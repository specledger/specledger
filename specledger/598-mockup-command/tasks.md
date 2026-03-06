# Tasks Index: Mockup Command

**Status**: Complete | **Epic**: SL-675f7d (closed)

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
├── Phase 1: Shared Infrastructure (SL-a1b2c3) ─ Extract prompt/editor from revise
│   ├── T000: Extract shared editor utilities into pkg/cli/prompt/editor.go
│   │         (DetectEditor, EditPrompt moved from revise/editor.go)
│   ├── T000b: Extract shared prompt utilities into pkg/cli/prompt/prompt.go
│   │          (RenderTemplate, EstimateTokens, PrintTokenWarnings)
│   └── T000c: Refactor revise to delegate to pkg/cli/prompt/
│              (revise/editor.go → thin wrapper, revise/prompt.go → delegates)
│
├── Phase 2: Setup (SL-acaa90) ─ Package + command skeleton
│   ├── T001: Create shared types in pkg/cli/mockup/types.go
│   │         (FrameworkType, DesignSystem, MockupPromptContext, StyleInfo,
│   │          SpecContent, MockupFormat — NO component types)
│   └── T002: Create mockup command skeleton + register in main.go
│             (VarMockupCmd with --format, --force, --dry-run, --summary, --json flags)
│
├── Phase 3: Domain Logic (SL-e6485b) ─ Detector, style scanner, design system, spec parser
│   ├── T003: Implement frontend framework detector
│   ├── T004: Implement CSS token extraction + component lib detection (stylescan.go)
│   │         (Extract colors, fonts, CSS variables, Tailwind v4, component libs)
│   ├── T004b: Implement app structure scanner (appscan.go)
│   │          (Layout files, component dirs, global styles — framework-aware)
│   ├── T005: Implement design system file I/O with YAML frontmatter
│   └── T006: Implement spec parser (specparser.go)
│             (Parse spec.md into SpecContent: title, user stories, requirements)
│
├── Phase 4: Prompt & Template (SL-d4e5f6) ─ Prompt builder + template + tests
│   ├── T007: Create mockup prompt template (prompt.tmpl)
│   │         (Agent instructions with spec, design tokens, framework, format)
│   ├── T008: Implement MockupPromptContext builder (mockupprompt.go)
│   │         (Assemble context from gathered data, render template)
│   └── T009: Add tests for prompt generation
│
├── Phase 5: Interactive Flow (SL-f7a8b9) ─ 6-step wiring in mockup.go
│   ├── T010: Wire spec resolution (arg / branch detection / picker)
│   │         (Reuse issues.NewContextDetector, fallback to huh.Select)
│   ├── T011: Wire auto framework detection (NO confirmation)
│   │         (lipgloss display only, --force bypass for non-frontend)
│   ├── T012: Wire design system check/generate with prompts
│   │         (Extract CSS tokens, no component scanning)
│   ├── T013: Wire prompt generation → editor review → action menu
│   │         (Uses pkg/cli/prompt/editor.go, Launch/Re-edit/Write/Cancel)
│   ├── T014: Wire agent launch with fallback
│   │         (launcher.LaunchWithPrompt, writePromptToFile fallback)
│   └── T015: Wire post-agent commit/push flow
│             (Reuse stagingAndCommitFlow pattern from revise.go)
│
├── Phase 6: Update Command (SL-a00981) ─ sl mockup update handler
│   └── T016: Implement sl mockup update handler
│             (Re-extract CSS tokens, update .specledger/memory/design-system.md)
│
├── Phase 7: Init Integration (SL-b0f5d0) ─ bootstrap.go
│   └── T017: Add frontend detection to bootstrap.go
│
└── Phase 8: Polish (SL-c71d1f) ─ JSON, errors, tests
    ├── T018: Add --json non-interactive path for both commands
    ├── T019: Comprehensive error handling and edge cases
    │         (Agent not found, mockup exists, user cancel, malformed design system)
    └── T020: Unit tests for all domain modules
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
                       └──────────────────────────────────────┼──→ T013 (editor/action menu)
                                                              │
T001 (types) ──┬──→ T003 (detector) ──→ T011 (auto detection)
               │                                │
               ├──→ T004 (style scanner) ───────┤
               │         │                      │
               ├──→ T004b (app scanner) ────────┤
               │         │                      │
               │         └──→ T012 (design system flow)
               │                       │
               ├──→ T005 (designsystem)┼──→ T016 (update handler)
               │                       └──→ T017 (bootstrap)
               │
               ├──→ T006 (specparser) ──→ T008 (prompt builder)
               │                               │
               └──→ T002 (cmd skeleton)        │
                       │                       │
                       └───────────────────────┴──→ T010 (spec resolution)
                                                        │
T007 (prompt.tmpl) ──→ T008 (prompt builder) ──→ T009 (tests)
                              │
                              └──→ T013 (editor/action menu)
                                        │
                                        └──→ T014 (agent launch)
                                                  │
                                                  └──→ T015 (commit/push)

T010-T015 (interactive flow) ──┬──→ T018 (JSON output)
                               └──→ T019 (error handling)
T016 (update)                ──┬──→ T018 (JSON output)
                               └──→ T019 (error handling)
T018 + T019                    ──→ T020 (all tests)
```

### Parallel Execution Opportunities

**Within Phase 1 (Shared Infra)**: T000 and T000b can run in parallel. T000c depends on both.

**Phase 2 + Phase 3**: After T001 completes, T003, T004, T004b, T005, T006 can run in parallel. T002 (cmd skeleton) can run in parallel with domain logic.

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
| T001     | types.go with FrameworkType, DesignSystem, StyleInfo, MockupPromptContext, SpecContent; compiles |
| T002     | mockup.go with VarMockupCmd, all flags (--format, --force, --dry-run, --summary, --json), registered in main.go |
| T003     | 3-tier detection, all frameworks, IsFrontend=false for non-frontend |
| T004     | CSS token extraction: colors, fonts, CSS variables, Tailwind v4, component libs; excluded dirs skipped |
| T004b    | App structure scanner: layouts, components (max 50), global styles (max 30); framework-aware (Next.js App/Pages Router, SvelteKit, Nuxt, etc.) |
| T005     | Load/Write design system with YAML frontmatter; edge case handling |
| T006     | ParseSpec extracts title, user stories, requirements; error on empty; SpecContent populated |
| T007     | prompt.tmpl with agent instructions template; embedded in binary |
| T008     | BuildMockupPrompt assembles context (no style data — agent reads design-system.md) and renders template |
| T009     | Tests verify prompt output for various contexts |
| T010     | Spec resolved from arg, branch, or picker; correct error for missing spec |
| T011     | Framework auto-detected and displayed with lipgloss (no confirmation), --force bypasses |
| T012     | Design system loaded or auto-generated with CSS token extraction |
| T013     | Editor opens, action menu works (all 4 options), uses shared prompt package |
| T014     | Agent launched if available, prompt written to file if not, install instructions shown |
| T015     | Post-agent commit/push flow works with file multi-select and push |
| T016     | Update handler re-extracts CSS tokens, updates .specledger/memory/design-system.md |
| T017     | Bootstrap integration, interactive prompt, CI auto-create, skip non-frontend |
| T018     | JSON output for both commands, non-JSON suppressed, jq-parseable |
| T019     | All edge cases handled per contract error messages |
| T020     | All test files pass, table-driven tests, coverage on domain modules |

## Implementation Strategy

### MVP (Suggested: Phase 1 + 2 + 3 + 4 + 5)

The MVP delivers **User Story 1** — the full mockup flow from spec to agent-generated mockup. This covers:
- Shared `pkg/cli/prompt/` package (extracted from revise)
- `pkg/cli/mockup/` package with all types, domain logic, and prompt builder
- `sl mockup [spec-name]` command with streamlined 6-step flow
- Spec resolution, auto framework detection, design system (CSS tokens), prompt, agent launch, commit/push
- `--dry-run` flag for non-agent workflow

**MVP scope**: T000-T015 (15 tasks), priority 1.

### Incremental Delivery

1. **MVP**: Phase 1-5 (Shared Infra + Setup + Domain + Prompt + Interactive Flow) — Full mockup generation
2. **Increment 1**: Phase 6 (Update Command) — `sl mockup update`
3. **Increment 2**: Phase 7 (Init Integration) — `sl init` design system setup
4. **Final**: Phase 8 (Polish) — JSON output, error handling, comprehensive tests

### Story Testability

| Story | Independently Testable? | Test Criteria |
|-------|------------------------|---------------|
| US1   | Yes | Run `sl mockup` on feature branch → flow completes, agent generates mockup |
| US2   | Yes | Run `sl mockup` without .specledger/memory/design-system.md → prompted to generate, file created, flow continues |
| US3   | Yes | Run `sl mockup` on React/Go project → correct detection displayed, auto-proceed or error |
| US4   | Yes | Run `sl mockup update` → design system refreshed with CSS tokens |
| US5   | Yes | Run `sl init` on frontend project → .specledger/memory/design-system.md created |

---

> This file is an index only. Implementation data lives in `sl issue`. Update this file only to point humans and agents to canonical query paths and feature references.
