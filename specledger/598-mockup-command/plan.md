# Implementation Plan: Mockup Command

**Branch**: `598-mockup-command` | **Date**: 2026-02-27 | **Spec**: [specledger/598-mockup-command/spec.md](spec.md)
**Input**: Feature specification from `specledger/598-mockup-command/spec.md`

## Summary

Add `sl mockup [spec-name]` and `sl mockup update` commands with a streamlined interactive TUI flow. The command auto-detects the spec from the branch, auto-detects the framework, checks/generates the design system (CSS tokens only), builds an AI agent prompt from templates and context, and lets the user review before launching the agent. The AI agent generates the mockup (HTML or JSX), and the user can optionally commit/push the result. Shared editor/prompt utilities are extracted from `revise` into `pkg/cli/prompt/` for reuse.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), go-git v5 (git), gopkg.in/yaml.v3 (YAML parsing)
**Storage**: File-based — Markdown with YAML frontmatter (`design-system.md`), HTML (`mockup.html`) or JSX (`mockup.jsx`)
**Testing**: `go test` (unit tests), table-driven tests following existing patterns
**Target Platform**: macOS, Linux (CLI binary via GoReleaser)
**Project Type**: Single CLI binary
**Performance Goals**: Mockup generation <30s (SC-001/SC-005), design system update <10s (SC-007)
**Constraints**: Offline-capable (no external API calls for detection/scanning), skip `node_modules`/`vendor`
**Scale/Scope**: Scan up to ~1000 component files, 11 frameworks (React/Next.js/Vue/Nuxt/Svelte/SvelteKit/Angular/Astro/SolidJS/Qwik/Remix)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution is in template state (not yet ratified). Proceeding with best-practice gates:

- [x] **Specification-First**: Spec.md complete with 5 prioritized user stories (P1-P3)
- [x] **Test-First**: Test strategy defined — unit tests for detector, scanner, generator; integration tests for end-to-end CLI flow
- [x] **Code Quality**: golangci-lint configured in CI (`.golangci.yml`), `gofmt` enforced
- [x] **UX Consistency**: User flows documented in acceptance scenarios (US1-US5)
- [x] **Performance**: Metrics defined — <30s mockup, <10s update (SC-001, SC-005, SC-007)
- [x] **Observability**: CLI output with progress indicators (checkmarks, timing), `--json` flag for machine output
- [ ] **Issue Tracking**: No issue tracker epic required (constitution not ratified)

**Complexity Violations**: None identified

### Post-Design Re-check

- [x] Data model covers all entities from spec (DetectionResult, Component, DesignSystem, Mockup, ScanResult)
- [x] CLI contract covers both commands (`sl mockup <spec-name>`, `sl mockup update`) with flags, exit codes, error messages
- [x] Integration point with `sl init` documented in contracts/cli.md
- [x] No new external Go dependencies required — uses stdlib + existing yaml.v3

## Project Structure

### Documentation (this feature)

```text
specledger/598-mockup-command/
├── plan.md              # This file
├── research.md          # Phase 0 output (complete)
├── data-model.md        # Phase 1 output (complete)
├── contracts/
│   └── cli.md           # Phase 1 output (complete)
└── tasks.md             # Phase 2 output (created by /specledger.tasks)
```

### Source Code (repository root)

```text
pkg/cli/
├── prompt/                    # Shared editor/prompt utilities (NEW — extracted from revise)
│   ├── editor.go              # DetectEditor(), EditPrompt() — moved from revise/editor.go
│   ├── editor_test.go         # Editor tests
│   ├── prompt.go              # RenderTemplate(), EstimateTokens(), PrintTokenWarnings()
│   └── prompt_test.go         # Prompt rendering tests
├── commands/
│   └── mockup.go              # Cobra command definitions (VarMockupCmd, mockupUpdateCmd)
├── mockup/                    # Domain logic package (NEW)
│   ├── detector.go            # Frontend framework detection (Tier 1-3 heuristics, 11 frameworks)
│   ├── detector_test.go       # Detection unit tests
│   ├── appscan.go             # App structure scanner (layouts, components, global styles)
│   ├── appscan_test.go        # App structure scanner tests
│   ├── stylescan.go           # CSS token extraction (colors, fonts, variables, component libs)
│   ├── stylescan_test.go      # Style scanning tests
│   ├── designsystem.go        # Design system read/write (YAML frontmatter + markdown body)
│   ├── designsystem_test.go   # Design system I/O tests
│   ├── specparser.go          # Parse spec.md content (title, user stories, requirements)
│   ├── specparser_test.go     # Spec parser tests
│   ├── mockupprompt.go        # MockupPromptContext builder, template renderer
│   ├── mockupprompt_test.go   # Prompt builder tests
│   ├── prompt.tmpl            # Embedded Go template for AI agent instructions
│   └── types.go               # Shared types (FrameworkType, DesignSystem, MockupPromptContext, etc.)
└── revise/                    # MODIFIED — refactor to use pkg/cli/prompt/
    ├── editor.go              # REPLACED — thin wrapper around prompt.EditPrompt()
    └── prompt.go              # MODIFIED — delegates to prompt.RenderTemplate()

pkg/cli/commands/
└── bootstrap.go               # MODIFIED — add frontend detection + design system init
```

**Structure Decision**: Follows the established pattern of domain logic in `pkg/cli/<feature>/` (like `revise/`, `session/`, `playbooks/`) with command definitions in `pkg/cli/commands/mockup.go`. Shared editor/prompt utilities are extracted from `revise` into `pkg/cli/prompt/` so both `revise` and `mockup` can reuse them. The `revise` package is refactored to delegate to the shared package (thin wrappers for backward compatibility).

## Previous Work

| Spec | Relevance | Reuse |
|------|-----------|-------|
| **597-issue-create-fields** | Most recent CLI command | Cobra patterns, flag handling, `--json` output |
| **011-streamline-onboarding** | Extends onboarding flow | `sl init` integration point in `bootstrap.go` |
| **591-issue-tracking-upgrade** | File-based storage | `pkg/issues/context.go` for spec detection |
| **136-revise-comments** | Domain logic separation | `pkg/cli/revise/` package pattern |
| **596-doctor-version-update** | File scanning | File traversal patterns |

## Architecture

### Command Flow

```
sl mockup [spec-name]
  │
  ├─ 1. Resolve spec
  │     ├─ If arg given: use directly
  │     ├─ If on feature branch: auto-detect via issues.NewContextDetector
  │     └─ If neither: interactive spec picker (huh.Select)
  ├─ 2. Auto-detect framework (detector.go)
  │     ├─ Tier 1: Config files (next.config.js, angular.json, etc.)
  │     ├─ Tier 2: package.json dependencies
  │     ├─ Tier 3: File extension scan
  │     └─ Display result with lipgloss (no confirmation needed)
  ├─ 3. Design system check/generate (designsystem.go)
  │     ├─ If exists: load and display
  │     ├─ If missing: prompt to generate → extract CSS tokens → write file
  │     └─ If malformed: warn and re-generate
  ├─ 4. Generate mockup prompt (mockup/prompt.go + prompt.tmpl)
  │     ├─ Parse spec content (specparser.go)
  │     ├─ Build MockupPromptContext from gathered data
  │     └─ Render template → prompt string
  ├─ 5. Edit & confirm prompt (pkg/cli/prompt/editor.go)
  │     ├─ Open $EDITOR with generated prompt
  │     └─ Action menu: Launch / Re-edit / Write to file / Cancel
  ├─ 6. Launch AI agent (pkg/cli/launcher/)
  │     ├─ If agent available: launcher.LaunchWithPrompt(prompt)
  │     └─ If no agent: writePromptToFile() + install instructions
  └─ 7. Post-agent commit/push (stagingAndCommitFlow pattern)
        ├─ Detect changed files → display summary
        ├─ huh.Confirm → file multi-select → commit message input
        └─ Stage → commit → push

sl mockup update
  │
  ├─ 1. Validate design-system.md exists
  ├─ 2. Load existing design system
  ├─ 3. Confirm rescan (huh.Confirm in interactive mode)
  ├─ 4. Re-extract CSS tokens (stylescan.go)
  └─ 5. Write updated design-system.md
```

### Key Design Decisions

1. **AI agent generates mockup content, not Go code** — The Go command orchestrates the interactive flow (framework detection, component selection, prompt building) and launches the AI agent. The agent produces the actual HTML/JSX mockup file. This leverages the agent's ability to understand UI requirements and produce high-quality layouts, while Go handles the structured context gathering that agents struggle with.

2. **Extract shared `pkg/cli/prompt/` package from `revise`** — Editor detection (`DetectEditor`), prompt editing (`EditPrompt`), template rendering, and token estimation are shared concerns. Extracting them avoids code duplication and ensures consistent UX between `revise` and `mockup`. The `revise` package delegates to `prompt` via thin wrappers.

3. **Reuse `stagingAndCommitFlow` for post-agent git operations** — The commit/push flow (changed file display, multi-select, commit message, push) is identical to `revise`. Rather than duplicating, `mockup.go` calls the same flow function.

4. **Spec-name is optional — auto-detect from branch** — Uses `issues.NewContextDetector` to extract spec name from the current branch (e.g., `598-mockup-command` → spec name). Falls back to interactive picker if not on a feature branch. Matches how other `sl` commands resolve context.

5. **Domain package at `pkg/cli/mockup/`** — Follows `revise/`, `session/` pattern. Keeps detection, scanning, and prompt logic separate from Cobra command wiring.

6. **Tiered framework detection (11 frameworks)** — Config files first (97-99% confidence), package.json fallback, file extension last resort. Returns `DetectionResult` with confidence score; `IsFrontend` only if confidence >= 70. Supports React, Next.js, Vue, Nuxt, Svelte, SvelteKit, Angular, Astro, SolidJS, Qwik, and Remix.

7. **YAML frontmatter + Markdown for design-system.md** — Machine-parseable metadata (version, framework, last_scanned, style, app_structure) in frontmatter, human-readable summary in markdown body (overview table, ASCII directory tree, color palette). The markdown body is for humans only — the AI agent reads the YAML frontmatter.

8. **Framework-aware app structure scanning** — `ScanAppStructure()` detects layouts, component dirs (1-level, max 50), and global stylesheets (max 30) using framework-specific heuristics. Next.js distinguishes App Router vs Pages Router. Skips `node_modules/`, `vendor/`, `.git/`, `dist/`, `build/`.

9. **Agent reads design-system.md directly** — The prompt template instructs the agent to read `.specledger/memory/design-system.md` rather than embedding style data in the prompt. This keeps prompts concise and makes the design system the single source of truth. `BuildMockupPromptContext` no longer takes `ds` or `style` parameters.

10. **Component library detection** — `ScanStyles()` detects 20+ component libraries (shadcn/ui, MUI, Chakra UI, Ant Design, Mantine, etc.) from package.json dependencies and stores them in `StyleInfo.ComponentLibs`.

11. **HTML/JSX mockup output** — Generated by the AI agent. HTML (default) uses semantic HTML with inline styles for immediate browser preview. JSX outputs React-compatible component code. Output is saved to `specledger/<spec-name>/mockup.html` or `mockup.jsx`.

12. **No new external dependencies** — Uses stdlib plus existing `gopkg.in/yaml.v3`, `huh`, `lipgloss`, and `launcher`. Avoids adding bloat.

### Codebase Integration Points

| Integration | File | Change |
|-------------|------|--------|
| Command registration | `cmd/sl/main.go` | Add `rootCmd.AddCommand(commands.VarMockupCmd)` |
| Command definition | `pkg/cli/commands/mockup.go` | New file — `VarMockupCmd` + `mockupUpdateCmd` |
| Init flow | `pkg/cli/commands/bootstrap.go` | Add frontend detection + design system init after base setup |
| Spec context | `pkg/issues/context.go` | Reuse `NewContextDetector` for auto-detecting spec name |
| UI output | `pkg/cli/ui/` | Reuse `ui.Checkmark()`, `ui.Bold()` for consistent output |
| Agent launch | `pkg/cli/launcher/launcher.go` | Reuse `NewAgentLauncher`, `LaunchWithPrompt`, `DefaultAgents` |
| Shared prompt | `pkg/cli/prompt/` | **New package** — extracted editor/prompt utilities |
| Revise refactor | `pkg/cli/revise/editor.go` | Refactor to delegate to `pkg/cli/prompt/editor.go` |
| Revise refactor | `pkg/cli/revise/prompt.go` | Refactor to delegate to `pkg/cli/prompt/prompt.go` |
| Commit flow | `pkg/cli/commands/revise.go` | Reuse `stagingAndCommitFlow` pattern (may extract to shared package) |

## Complexity Tracking

> No violations identified. Two package additions follow established patterns. Shared extraction reduces long-term duplication.

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Package count | +2 (`pkg/cli/mockup/`, `pkg/cli/prompt/`) | `mockup` for domain logic, `prompt` for shared editor/template utilities |
| New files | ~17 (8 source + 7 test + 1 command + 1 template) | Includes appscan.go + appscan_test.go for app structure scanning |
| External deps | 0 new | All needed functionality in stdlib + yaml.v3 + existing huh/lipgloss/launcher |
| Modified files | 3 (`main.go`, `bootstrap.go`, `revise.go`) + 2 refactored (`revise/editor.go`, `revise/prompt.go`) | Shared extraction requires refactoring revise |
