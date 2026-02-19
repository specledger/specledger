# Research: 011-streamline-onboarding

**Date**: 2026-02-18
**Branch**: `011-streamline-onboarding`

## Prior Work

- **SL-8gh** (CLOSED): "Contributor Onboarding" — contributor documentation and setup guides. This feature extends onboarding from contributors to end-users.
- **SL-31n** (CLOSED): "Command System Enhancements" — added `/specledger.help` and enhanced all commands with Purpose sections. Foundation for discoverability.
- **SL-6t7** (CLOSED): "Enhanced Purpose Sections" — discoverability improvements across all commands.
- **SL-9c7** (OPEN): "Open Source Readiness" — includes contributor onboarding. Streamlined onboarding directly supports this.
- **SL-55w** (CLOSED): "Path Standardization" — migrated from `.specify` to `.specledger` directory. All paths now use `.specledger/`.
- **SL-6lz** (CLOSED): "Update agent context paths" — updated path references for `.specledger`.

No existing epics for constitution creation, agent launch, or TUI improvements in isolation.

## R1: How to Launch Claude Code Programmatically

**Decision**: Use `exec.Command("claude", "-p", prompt)` with `cmd.Dir` set to project directory.

**Rationale**: Claude Code CLI supports `-p` flag for passing an initial prompt. This is the documented approach for programmatic/headless invocation. The specledger codebase already uses `exec.Command` patterns for launching external processes (see `pkg/cli/auth/browser.go` for `open` command, `pkg/cli/commands/bootstrap_helpers.go` for init scripts).

**Key flags**:
- `claude -p "prompt"` — non-interactive/print mode (runs prompt, exits)
- `claude --append-system-prompt "instructions"` — adds custom instructions
- No flag exists to launch Claude in interactive mode with a pre-loaded initial message

**Critical finding**: The `-p` flag runs in headless/non-interactive mode and exits after completion. For an interactive onboarding session, we need to launch `claude` without `-p` but with an onboarding command embedded in the project's `.claude/commands/` directory. The onboarding prompt should be an embedded command file (e.g., `specledger.onboard.md`) that the user can invoke as `/specledger.onboard` when Claude starts.

**Revised approach**:
1. Launch `claude` interactively (no `-p` flag) via `exec.Command("claude")` with `cmd.Stdin`, `cmd.Stdout`, `cmd.Stderr` connected to os.Stdin/Stdout/Stderr
2. The embedded `.claude/commands/specledger.onboard.md` command file is already present in the project
3. Print a post-setup message telling the user to type `/specledger.onboard` to begin the guided workflow
4. Alternative: Use Claude Code's `SessionStart` hook to auto-run the onboarding prompt via `bd prime`-style output

**Alternatives considered**:
- `claude -p "onboarding prompt"`: Rejected — runs headless, exits after completion, not interactive
- `claude --system-prompt "..."`: Rejected — replaces default system prompt entirely, losing built-in capabilities
- `claude --append-system-prompt "..."`: Viable for injecting context but still requires interactive launch

## R2: Extending the Bubble Tea TUI

**Decision**: Add 2 new steps to the existing 5-step TUI flow: constitution principles (step 5) and agent preference (step 6).

**Rationale**: The TUI model (`pkg/cli/tui/sl_new.go`) uses a simple step counter with a switch statement. Adding new steps requires:
1. New step constants (`stepConstitution`, `stepAgentPreference`)
2. New view methods (`viewConstitution()`, `viewAgentPreference()`)
3. New cases in `Update()`, `handleEnter()`, and `View()`
4. Updated "Step X of Y" count

The existing TUI already handles text input (project name, short code) and list selection (playbook). Constitution principles can use a similar list selection with checkboxes/toggles. Agent preference uses the same list selection as playbook.

**Alternatives considered**:
- Separate TUI for constitution: Rejected — fragmented experience, user would need to run two commands
- Prompt-based (no TUI): Rejected — inconsistent with existing `sl new` experience

## R3: Making `sl init` Interactive

**Decision**: Create a new TUI model for `sl init` that reuses components from `sl new`'s TUI, presenting only the missing configuration prompts.

**Rationale**: Currently `sl init` is entirely flag-based (no TUI). To match `sl new`'s experience, `sl init` needs a TUI that:
1. Detects which configuration is already present (short code from flags, existing constitution, etc.)
2. Presents only the missing prompts
3. Skips project name and directory steps (already determined by current directory)

The TUI model can share the same step components (short code input, playbook selection, constitution, agent preference) but with a dynamic step list based on what's missing.

**Alternatives considered**:
- Reuse `sl new` TUI as-is: Rejected — `sl init` doesn't need project name/directory steps
- Simple stdin prompts: Rejected — inconsistent with Bubble Tea UI already used

## R4: Constitution Creation During Onboarding

**Decision**: For `sl new`, create a TUI step presenting default principles as a toggleable list. For `sl init`, delegate to the AI agent post-launch.

**Rationale**:
- `sl new` creates a brand-new project, so default principles (specification-first, test-first, code quality, etc.) are reasonable starting points. The TUI step presents these as a list the user can accept/modify.
- `sl init` targets existing codebases where conventions already exist. The AI agent can run `/specledger.audit` for deeper analysis than the CLI could provide, then `/specledger.constitution` to propose tailored principles.

**Constitution populated template**: Replace `[PLACEHOLDER]` tokens with concrete values. Detection of unfilled template: check for `[` followed by `ALL_CAPS_IDENTIFIER` followed by `]` pattern.

**Alternatives considered**:
- Always delegate to AI agent: Rejected for `sl new` — adds unnecessary dependency on AI agent for new projects
- Built-in heuristics for `sl init`: Rejected — AI agent provides richer analysis

## R5: Onboarding Command Design

**Decision**: Create an embedded Claude Code command `specledger.onboard.md` that orchestrates the guided workflow.

**Rationale**: The existing command system (`.claude/commands/specledger.*.md`) provides a proven pattern for structuring AI agent workflows. An onboarding command can:
1. Check constitution status
2. Guide through specify → clarify → plan → tasks → implement
3. Pause for user review at task generation
4. Be invoked manually (`/specledger.onboard`) or auto-triggered

**Workflow sequence in onboarding command**:
1. If no constitution: run `/specledger.audit` then `/specledger.constitution`
2. Welcome message explaining SpecLedger workflow
3. Ask user to describe their first feature
4. Run `/specledger.specify` with the description
5. Run `/specledger.clarify`
6. Run `/specledger.plan`
7. Run `/specledger.tasks`
8. **PAUSE**: Present tasks for user review, ask for approval
9. Only after explicit approval: run `/specledger.implement`

**Alternatives considered**:
- Inline the entire workflow in a launch prompt: Rejected — too large, hard to maintain, not reusable
- Chain of separate commands without orchestration: Rejected — user wouldn't know the sequence

## R6: Agent Availability Check

**Decision**: Use `exec.LookPath("claude")` to verify Claude Code is installed before attempting launch.

**Rationale**: Go's `exec.LookPath` is the standard way to check if a command exists in PATH. If not found, display installation instructions and complete project setup without launching.

**Alternatives considered**:
- Skip check, let exec fail: Rejected — poor UX, cryptic error message
- Check version too: Deferred — version compatibility can be addressed later

## R7: Agent Preference in Constitution

**Decision**: Add an `## Agent Preferences` section to the constitution template with a `preferred_agent` field.

**Rationale**: The constitution is the authoritative source for project preferences (per spec clarification). Adding a structured section keeps it machine-readable while fitting the existing constitution format. The section can be parsed by `sl init` to read existing preferences.

**Format**:
```markdown
## Agent Preferences

- **Preferred Agent**: Claude Code
```

**Alternatives considered**:
- YAML field in specledger.yaml: Rejected — clarification decided constitution is the source for preferences
- Separate config file: Rejected — adds file proliferation, constitution is the right home
