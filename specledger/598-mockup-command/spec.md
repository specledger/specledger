# Feature Specification: Mockup Command

**Feature Branch**: `598-mockup-command`
**Created**: 2026-02-26
**Status**: Draft
**Input**: User description: "Create new command 'sl mockup' which checks if the repository is a frontend repository, reads specledger/design_system.md for UI component indexing, generates mockups based on spec.md using the design system, creates design_system.md if missing, and initializes it during onboarding."

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Generate Mockup via Interactive Flow (Priority: P1)

As a frontend developer working on a feature, I want to run `sl mockup [spec-name]` so that an interactive flow guides me through framework detection, design system setup, component selection, and prompt review before launching an AI agent to generate a mockup grounded in my project's actual UI components.

**Why this priority**: This is the core value proposition — an interactive, agent-driven flow that generates mockups grounded in the project's actual design system ensures consistency, gives the developer control at each step, and produces higher-quality output than a non-interactive generator.

**Independent Test**: Can be fully tested by running `sl mockup` on a feature branch (or `sl mockup <spec-name>` explicitly) in a frontend project with both a spec and design system. The interactive flow should guide through each step and launch the AI agent to produce the mockup file.

**Acceptance Scenarios**:

1. **Given** a frontend repository on a feature branch `598-mockup-command`, **When** the user runs `sl mockup` (no argument), **Then** the system auto-detects the spec from the branch name and proceeds with the interactive flow.
2. **Given** the interactive flow is running, **When** framework detection completes, **Then** the system displays the detected framework with lipgloss styling and asks for confirmation (or override).
3. **Given** a design system exists, **When** the flow reaches the component step, **Then** the user is presented a `huh.MultiSelect` of available components to include in the mockup prompt.
4. **Given** the user has confirmed components and format, **When** the prompt is generated, **Then** the system opens the user's `$EDITOR` for prompt review and shows an action menu: Launch / Re-edit / Write to file / Cancel.
5. **Given** the user selects "Launch", **When** an AI agent is available, **Then** the system launches the agent with the prompt via `launcher.LaunchWithPrompt()`.
6. **Given** the agent session completes, **When** there are uncommitted changes, **Then** the system offers to commit and push using the `stagingAndCommitFlow` pattern (file multi-select, commit message input, push).
7. **Given** a `--dry-run` flag, **When** the prompt is generated, **Then** the system writes the prompt to a file instead of launching the agent and exits.
8. **Given** a frontend repository with a design system containing a Button, Form, and Card component, **When** the user completes the interactive flow for a feature requiring user input, **Then** the generated prompt instructs the agent to reference the existing Form and Button components rather than inventing new ones.

---

### User Story 2 - Auto-Create Design System Index When Missing (Priority: P2)

As a frontend developer in a project without a `specledger/design_system.md`, I want the interactive flow to detect the missing file, scan the codebase, and generate an initial design system index so that I can immediately proceed with mockup generation.

**Why this priority**: This removes a setup barrier and makes the command usable out of the box for projects that haven't catalogued their components yet.

**Independent Test**: Can be tested by running `sl mockup` in a frontend repo that lacks `specledger/design_system.md`. The interactive flow should prompt the user to generate the design system, scan components, and then proceed to the component selection step.

**Acceptance Scenarios**:

1. **Given** a frontend repository without `specledger/design_system.md`, **When** the interactive flow reaches the design system step, **Then** the system prompts the user to generate the design system, scans the codebase, and creates `specledger/design_system.md`.
2. **Given** a React project with components in `src/components/`, **When** the design system index is auto-generated, **Then** the user sees a lipgloss-styled summary of discovered components before proceeding.
3. **Given** a project using a component library (e.g., Material UI, Ant Design), **When** the design system is generated, **Then** both custom project components and identifiable library components used in the project are indexed.

---

### User Story 3 - Interactive Framework Detection and Confirmation (Priority: P2)

As a developer working on a frontend project, I want the interactive flow to detect my frontend framework, display the result with lipgloss styling, and let me confirm or override it, so that the mockup context is always correct.

**Why this priority**: This ensures the mockup command works seamlessly for frontend projects and gives the user control over framework selection.

**Independent Test**: Can be tested by running `sl mockup` in a frontend repository. The system should detect the frontend framework, display it with confidence score, and allow the user to confirm or override.

**Acceptance Scenarios**:

1. **Given** a React/Vue/Svelte/Angular repository, **When** the interactive flow reaches framework detection, **Then** the system displays the detected framework with confidence score using lipgloss and prompts for confirmation.
2. **Given** an ambiguous repository (e.g., a monorepo with both backend and frontend), **When** framework detection runs, **Then** the system presents multiple detected frameworks and lets the user pick one via `huh.Select`.
3. **Given** a non-frontend repository without `--force`, **When** the user runs `sl mockup`, **Then** the system displays a styled error and exits with code 2.

---

### User Story 4 - Update Design System Index (Priority: P2)

As a developer who has added new UI components to my project, I want to run `sl mockup update` to refresh the design system index so that new components are available for future mockups.

**Why this priority**: Allows developers to keep the design system index in sync with their codebase as components evolve.

**Independent Test**: Can be tested by adding new components to a frontend project and running `sl mockup update`. The system should scan the codebase and update `specledger/design_system.md` with the new components.

**Acceptance Scenarios**:

1. **Given** a frontend repository with an existing `specledger/design_system.md`, **When** the user adds new UI components and runs `sl mockup update`, **Then** the system scans the codebase and updates the design system index with the new components.
2. **Given** a design system that has been manually edited, **When** the user runs `sl mockup update`, **Then** the system respects manual additions/modifications and merges new discoveries with existing entries.
3. **Given** a project where components have been removed, **When** the user runs `sl mockup update`, **Then** the system identifies and removes stale component entries from the design system index.

---

### User Story 5 - Initialize Design System During Onboarding (Priority: P3)

As a new user onboarding a frontend project to SpecLedger, I want the `sl init` / onboarding process to automatically create `specledger/design_system.md` so that the design system index is ready when I first use `sl mockup`.

**Why this priority**: Improves the onboarding experience but is not blocking since US3 handles the case where the file is missing at mockup time.

**Independent Test**: Can be tested by running `sl init` on a frontend project and verifying that `specledger/design_system.md` is created alongside other initialization artifacts.

**Acceptance Scenarios**:

1. **Given** a frontend repository being initialized with `sl init`, **When** onboarding completes, **Then** `specledger/design_system.md` exists in the specledger directory with an auto-generated index of UI components.
2. **Given** a non-frontend repository being initialized with `sl init`, **When** onboarding completes, **Then** `specledger/design_system.md` is NOT created (since mockup is not applicable).

---

### Edge Cases

- What happens when the design system file exists but is empty or malformed? The system should treat it as missing and re-generate it, warning the user.
- What happens when the spec.md has no user scenarios or functional requirements to mockup? The system should abort with a message directing the user to complete the spec first.
- What happens when the frontend project uses an uncommon framework not in the detection heuristics? The system should provide a `--force` flag to bypass detection and allow the user to proceed.
- What happens when the user runs `sl mockup` without providing a spec name and is not on a feature branch? The system should present an interactive picker of available specs via `huh.Select`.
- What happens when the component scan finds hundreds of components? The index should group components by directory/module and include all of them — the component multi-select step lets the user filter.
- What happens when no AI agent is available (not installed)? The system should write the prompt to a file and display install instructions (same as `revise` fallback).
- What happens when a mockup file already exists at the output path? The system should prompt for overwrite confirmation before proceeding.
- What happens when the user cancels at any interactive step? The system should exit cleanly with code 0 and no partial output.
- What happens when `--json` is used? The system should run a non-interactive path (skip huh prompts, use flag values directly) and output structured JSON.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST detect whether the current repository is a frontend project by checking for frontend indicators (package.json with frontend dependencies, framework config files, or source directories with frontend component files).
- **FR-002**: System MUST initialize the design system if the repository is a frontend project and `specledger/design_system.md` is missing.
- **FR-003**: System MUST read and parse `specledger/design_system.md` to build an index of available UI components with their names, file paths, and descriptions.
- **FR-004**: System MUST generate a mockup based on the feature's `specledger/<spec-name>/spec.md`, mapping UI needs to existing design system components wherever possible.
- **FR-005**: System MUST auto-generate `specledger/design_system.md` by scanning the codebase for UI components when the file does not exist.
- **FR-006**: System MUST integrate with the `sl init` onboarding flow to create `specledger/design_system.md` for frontend projects during initialization.
- **FR-007**: System MUST output the generated mockup to the feature directory as HTML or JSX (e.g., `specledger/<spec-name>/mockup.html` or `specledger/<spec-name>/mockup.jsx`), defaulting to HTML.
- **FR-013**: System MUST support a `--format` flag accepting `html` or `jsx` to control the mockup output format.
- **FR-008**: System MUST support at minimum React (.tsx/.jsx), Vue (.vue), Svelte (.svelte), and Angular (.component.ts) component detection.
- **FR-009**: System MUST handle the case where `spec.md` contains no user scenarios by displaying a helpful error message directing the user to run the specify workflow first.
- **FR-010**: System MUST allow users to manually edit `specledger/design_system.md` and respect manual additions/modifications on subsequent runs.
- **FR-011**: System MUST provide a `--force` flag to bypass frontend detection for edge cases (e.g., uncommon frameworks).
- **FR-012**: System MUST support `sl mockup update` command to refresh the design system index by rescanning the codebase.
- **FR-014**: System MUST support a `--dry-run` flag that writes the generated prompt to a file instead of launching the AI agent. _(US1)_
- **FR-015**: System MUST support a `--summary` flag for compact output suitable for agent integration and CI environments. _(US1)_
- **FR-016**: System MUST allow spec-name to be optional, auto-detecting from the current branch via `issues.NewContextDetector`. If not on a feature branch and no argument given, present an interactive spec picker. _(US1)_
- **FR-017**: System MUST launch the configured AI agent with the generated prompt using `launcher.LaunchWithPrompt()`. If no agent is available, fall back to writing the prompt to a file with install instructions. _(US1)_
- **FR-018**: System MUST offer to commit and push changes after the agent session completes, reusing the `stagingAndCommitFlow` pattern (file multi-select, commit message, push). _(US1)_
- **FR-019**: System MUST display interactive confirmation at each major step — framework detection, design system generation, component selection, format selection, and prompt review — using `huh` forms and `lipgloss` displays. _(US1, US2, US3)_
- **FR-020**: System MUST extract shared editor and prompt utilities from the `revise` package into a reusable `pkg/cli/prompt/` package for use by both `revise` and `mockup` commands. _(US1)_

### Key Entities

- **Design System Index**: A markdown file (`specledger/design_system.md`) that catalogs all UI components in the project — each entry includes the component name, file path, a brief description, and optionally its props/inputs.
- **Mockup**: An HTML or JSX file representing the feature's UI, referencing components from the design system index. Contains screen layouts with component placements, user interaction flows, and annotations. HTML format uses semantic HTML with inline styles; JSX format outputs React-compatible component code. **Generated by the AI agent, not by Go code.**
- **Frontend Detection Result**: The outcome of the repository type check — identifies the frontend framework(s) in use, the component directory structure, and whether the project qualifies as a frontend repository.
- **Mockup Prompt Context**: The template rendering context passed to the agent prompt — includes spec name, parsed spec content, framework, format, output path, selected components, external libraries, and design system presence.
- **Spec Content**: Parsed content from `spec.md` — title, user stories, requirements, and full content text used to build the agent prompt.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Users can generate a mockup from an existing spec in under 30 seconds (excluding initial design system scan).
- **SC-002**: Generated mockups reference existing design system components rather than proposing entirely new UI elements for common patterns.
- **SC-003**: Frontend detection correctly identifies frontend projects across React, Vue, Svelte, and Angular projects.
- **SC-004**: The auto-generated design system index captures discoverable UI components in the scanned codebase.
- **SC-005**: Users running `sl mockup <spec-name>` on a frontend repo receive a mockup within 30 seconds.
- **SC-006**: After onboarding a frontend project, `specledger/design_system.md` is present and populated without additional user action.
- **SC-007**: Running `sl mockup update` refreshes the design system index in under 10 seconds.

### Previous work

- **597-issue-create-fields**: Issue create fields enhancement — most recent CLI command addition, establishes patterns for new commands.
- **011-streamline-onboarding**: Streamlined onboarding — directly related since this feature extends the onboarding flow to initialize design_system.md.

## Scope & Boundaries

### In Scope

- New `sl mockup [spec-name]` CLI command with interactive TUI flow
- New `sl mockup update` CLI command
- Frontend repository detection logic with interactive confirmation
- Design system index file format and auto-generation
- AI agent-driven mockup generation from spec + design system (HTML or JSX output)
- `--format` flag to choose between `html` (default) and `jsx` output
- `--dry-run` flag to write prompt to file without launching agent
- `--summary` flag for compact output (agent/CI integration)
- Agent launch via `launcher.LaunchWithPrompt()` with fallback to file
- Editor integration for prompt review (reused from `revise`)
- Post-agent commit/push flow (reused `stagingAndCommitFlow` pattern)
- Shared `pkg/cli/prompt/` package extracted from `revise`
- Integration with `sl init` for design system initialization
- `--force` flag to bypass frontend detection
- Auto-detection of spec-name from branch via `issues.NewContextDetector`

### Out of Scope

- Go code that directly generates mockup content (the AI agent generates it)
- Design token extraction (colors, typography, spacing)
- Component screenshot generation or rendering
- Figma/Sketch/design tool integration
- Component dependency graph or usage analytics
- Cross-project design system sharing

## Dependencies & Assumptions

### Assumptions

- The AI agent generates the mockup content (HTML or JSX) — Go code orchestrates the interactive flow, gathers context, builds the prompt, and launches the agent
- The mockup output format is HTML or JSX (component-based layouts with annotations), not graphical design files
- Frontend detection uses file-based heuristics (checking for package.json, framework configs, component file extensions) rather than requiring user configuration
- The design system index follows a structured markdown format that both humans and the system can read/update
- Component scanning is limited to the project's source directories and does not traverse node_modules or vendor directories
- The mockup command follows the same interactive pattern as `sl revise` — editor integration, agent launch, commit/push flow
- Spec-name is optional; the system auto-detects from the current branch name using `issues.NewContextDetector`

### Dependencies

- Requires a valid spec name pointing to `specledger/<spec-name>/spec.md` file (provided or auto-detected)
- Relies on existing `sl init` infrastructure for onboarding integration
- Uses the project's file system structure to detect components (no external API calls needed)
- Depends on `pkg/cli/launcher/` for AI agent launch (existing package)
- Depends on `pkg/cli/revise/editor.go` and `prompt.go` for shared editor/prompt utilities (to be extracted into `pkg/cli/prompt/`)
- Depends on `pkg/issues/context.go` for spec auto-detection from branch name
