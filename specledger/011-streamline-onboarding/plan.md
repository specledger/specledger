# Implementation Plan: Streamlined Onboarding Experience

**Branch**: `011-streamline-onboarding` | **Date**: 2026-02-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specledger/011-streamline-onboarding/spec.md`

## Summary

Enhance `sl new` and `sl init` commands to provide a unified, guided onboarding experience. `sl new` gains two new TUI steps: constitution creation (with default principles) and agent preference selection. `sl init` becomes interactive, presenting missing configuration prompts via a new TUI, then launching the selected AI coding agent. A new embedded command (`/specledger.onboard`) orchestrates the guided workflow: audit → constitution → specify → clarify → plan → tasks (review pause) → implement.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), Bubble Tea + Bubbles + Lipgloss (TUI), go-git v5, YAML v3
**Storage**: File-based (`.specledger/memory/constitution.md`, `specledger/specledger.yaml`)
**Testing**: Go standard `testing` package, integration tests via built binary + `exec.Command`
**Target Platform**: macOS, Linux (cross-platform CLI)
**Project Type**: Single project (Go CLI)
**Performance Goals**: TUI renders instantly; agent launch <2s; constitution template detection <100ms
**Constraints**: Must work offline (except AI agent features); no new external dependencies
**Scale/Scope**: Single-user CLI tool; changes touch ~7 files, ~1 new package, ~1 new embedded command

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Note: The project constitution (`.specledger/memory/constitution.md`) is currently an unfilled template. The checks below use the template's example principles as guidance.

- [x] **Specification-First**: Spec.md complete with 4 prioritized user stories, 18 functional requirements, 7 success criteria
- [x] **Test-First**: Test strategy defined — unit tests for TUI steps and launcher, integration tests for bootstrap flow (see quickstart.md)
- [x] **Code Quality**: Go standard tooling — `go vet`, `gofmt`, existing CI pipeline
- [x] **UX Consistency**: User flows documented with 19 acceptance scenarios across 4 user stories
- [x] **Performance**: Agent launch <2s, constitution detection <100ms, no latency-sensitive operations
- [x] **Observability**: Non-fatal errors logged with warnings (existing pattern); post-setup message confirms status
- [x] **Issue Tracking**: Beads epic to be created during `/specledger.tasks` phase

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/011-streamline-onboarding/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 research output
├── data-model.md        # Phase 1 data model
├── quickstart.md        # Phase 1 development quickstart
└── tasks.md             # Phase 2 output (/specledger.tasks - NOT created by /specledger.plan)
```

### Source Code (repository root)

```text
pkg/
├── cli/
│   ├── commands/
│   │   ├── bootstrap.go              # MODIFY: wire agent launch, init TUI
│   │   └── bootstrap_helpers.go      # MODIFY: add launchAgent(), constitutionCheck()
│   ├── tui/
│   │   ├── sl_new.go                 # MODIFY: add constitution + agent pref steps
│   │   └── sl_init.go                # NEW: interactive TUI for sl init
│   └── launcher/
│       └── launcher.go               # NEW: agent availability check + launch
├── embedded/
│   └── templates/
│       └── specledger/
│           ├── .claude/commands/
│           │   └── specledger.onboard.md  # NEW: guided onboarding command
│           └── .specledger/memory/
│               └── constitution.md        # MODIFY: add Agent Preferences section

tests/
└── integration/
    └── bootstrap_test.go             # MODIFY: add tests for new TUI steps + agent launch
```

**Structure Decision**: Single project structure (Go CLI). All changes extend existing packages (`cli/commands`, `cli/tui`) with one new package (`cli/launcher`) and one new embedded command template. No new top-level directories.

## Design Decisions

### D1: Agent Launch Strategy

The Go CLI launches the AI agent as an interactive subprocess with stdio passthrough. The user's terminal is handed over to the agent process.

```go
cmd := exec.Command("claude")
cmd.Dir = projectDir
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
cmd.Run() // Blocking — CLI waits for agent to exit
```

Post-setup, the CLI prints a message instructing the user to type `/specledger.onboard` to begin the guided workflow. The embedded `specledger.onboard.md` command handles the full workflow orchestration.

### D2: Constitution Detection

```go
func isConstitutionPopulated(path string) bool {
    content, err := os.ReadFile(path)
    if err != nil {
        return false // Missing file = not populated
    }
    // Check for placeholder pattern: [ALL_CAPS_IDENTIFIER]
    re := regexp.MustCompile(`\[[A-Z_]{3,}\]`)
    return !re.Match(content)
}
```

### D3: sl init TUI Step Selection

The `sl init` TUI dynamically determines which steps to show:

```
always show: short_code (if not provided via flag)
always show: playbook (if not provided via flag)
always show: agent_preference (unless constitution has it)
never show: constitution (delegated to AI agent)
never show: project_name, directory (derived from cwd)
```

### D4: Onboarding Command Branching

The `specledger.onboard.md` command checks constitution status and branches:

```
sl new path:  constitution exists → confirm → specify workflow
sl init path: constitution missing → /specledger.audit → /specledger.constitution → specify workflow
```

The command uses the AskUserQuestion tool to pause at task review and get explicit approval before `/specledger.implement`.

### D5: Default Constitution Principles

For `sl new`, the TUI presents 5 pre-selected default principles. Users can deselect any or keep all. The principles are written to `.specledger/memory/constitution.md` replacing the template placeholders, plus an `## Agent Preferences` section with the selected agent.

### D6: CI Mode Behavior

In CI mode (`--ci` flag or no TTY):
- Constitution: Uses default principles (all 5 selected) without prompting
- Agent preference: Defaults to "None" (no agent launched)
- All other behavior unchanged from current CI mode

## Complexity Tracking

No complexity violations identified. All changes follow existing patterns (TUI steps, embedded templates, exec.Command for external processes).

## Post-Phase 1 Constitution Re-check

- [x] **Specification-First**: Confirmed — spec complete, plan references spec throughout
- [x] **Test-First**: Confirmed — test strategy in quickstart.md, integration test modifications planned
- [x] **Code Quality**: Confirmed — standard Go tooling, no new linting requirements
- [x] **UX Consistency**: Confirmed — TUI extensions follow existing step pattern, consistent UI
- [x] **Performance**: Confirmed — no new latency-sensitive operations
- [x] **Observability**: Confirmed — existing warning/error logging patterns reused
- [x] **Issue Tracking**: Confirmed — Beads epic creation deferred to `/specledger.tasks`
