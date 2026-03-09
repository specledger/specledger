# Feature Specification: Mockup Command

**Feature Branch**: `598-mockup-command`
**Created**: 2026-02-26
**Status**: Draft
**Input**: User description: "Create new command 'sl mockup' which checks if the repository is a frontend repository, reads .specledger/memory/design-system.md for UI component indexing, generates mockups based on spec.md using the design system, creates design-system.md if missing, and initializes it during onboarding."

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Generate Mockup via Interactive Flow (Priority: P1)

As a frontend developer working on a feature, I want to run `sl mockup [prompt...]` so that an interactive flow guides me through design system setup and prompt review before launching an AI agent to generate a mockup grounded in my project's design tokens.

**Why this priority**: This is the core value proposition — an agent-driven flow that generates mockups grounded in the project's design tokens ensures consistency and produces higher-quality output than a non-interactive generator.

**Independent Test**: Can be fully tested by running `sl mockup` on a feature branch in a frontend project with both a spec and design system. The interactive flow should guide through each step and launch the AI agent to produce the mockup file.

**Acceptance Scenarios**:

1. **Given** a frontend repository on a feature branch `598-mockup-command`, **When** the user runs `sl mockup` (no argument), **Then** the system auto-detects the spec from the branch name and proceeds with the interactive flow (confirmations enabled).
2. **Given** a frontend repository on a feature branch, **When** the user runs `sl mockup focus on login form` (with arguments), **Then** the system skips confirmations and launches the agent directly with the user's instructions.
3. **Given** the interactive flow is running, **When** framework detection completes, **Then** the system displays the detected framework with lipgloss styling and proceeds automatically.
4. **Given** a design system exists, **When** the flow proceeds, **Then** the system includes the design tokens in the AI agent prompt so mockups use consistent styling.
5. **Given** the prompt is generated (interactive mode), **When** the user proceeds, **Then** the system opens the user's `$EDITOR` for prompt review and shows an action menu: Launch / Re-edit / Write to file / Cancel.
6. **Given** the user selects "Launch", **When** an AI agent is available, **Then** the system launches the agent with the prompt via `launcher.LaunchWithPrompt()`.
7. **Given** the agent session is running, **When** the mockup is generated, **Then** the AI agent asks the user if they want to commit and push (the agent handles git operations, not CLI).
8. **Given** a `--dry-run` flag, **When** the prompt is generated, **Then** the system writes the prompt to a file instead of launching the agent and exits.
9. **Given** a `-y` or `--yes` flag, **When** the command runs, **Then** the system skips all confirmations and launches the agent directly.
10. **Given** a frontend repository with design tokens (colors, spacing, typography), **When** the user completes the interactive flow, **Then** the generated prompt instructs the agent to use the project's design tokens and search the codebase for existing components rather than inventing new styles.

---

### User Story 2 - Auto-Create Design System When Missing (Priority: P2)

As a frontend developer in a project without a `.specledger/memory/design-system.md`, I want the interactive flow to detect the missing file, extract global CSS styles (design tokens, variables, theme configuration), and generate an initial design system document so that the AI agent can generate mockups consistent with my project's visual identity.

**Why this priority**: This removes a setup barrier and ensures mockups match the project's design language. Component discovery is left to the AI agent which can search and identify them directly.

**Independent Test**: Can be tested by running `sl mockup` in a frontend repo that lacks `.specledger/memory/design-system.md`. The interactive flow should prompt the user to generate the design system, extract global CSS, and then proceed.

**Acceptance Scenarios**:

1. **Given** a frontend repository without `.specledger/memory/design-system.md`, **When** the interactive flow reaches the design system step, **Then** the system prompts the user to generate the design system, extracts global CSS/tokens, and creates `.specledger/memory/design-system.md`.
2. **Given** a React project with Tailwind CSS configuration, **When** the design system is auto-generated, **Then** the user sees a lipgloss-styled summary of discovered design tokens (colors, spacing, typography) before proceeding.
3. **Given** a project using CSS variables in global stylesheets, **When** the design system is generated, **Then** all CSS custom properties (--color-primary, --spacing-lg, etc.) are extracted and documented.

---

### User Story 3 - Automatic Framework Detection (Priority: P2)

As a developer working on a frontend project, I want the interactive flow to auto-detect my frontend framework and display the result with lipgloss styling, so that the mockup process can proceed quickly without unnecessary confirmation steps.

**Why this priority**: This ensures the mockup command works seamlessly for frontend projects with minimal friction.

**Independent Test**: Can be tested by running `sl mockup` in a frontend repository. The system should detect the frontend framework, display it with confidence score, and proceed automatically.

**Acceptance Scenarios**:

1. **Given** a React/Next.js/Vue/Nuxt/Svelte/SvelteKit/Angular/Astro/SolidJS/Qwik/Remix repository, **When** the interactive flow reaches framework detection, **Then** the system displays the detected framework with confidence score using lipgloss and proceeds automatically.
2. **Given** a non-frontend repository without `--force`, **When** the user runs `sl mockup`, **Then** the system displays a styled error and exits with code 2.
3. **Given** a non-frontend repository with `--force`, **When** the user runs `sl mockup`, **Then** the system proceeds with framework=unknown.

---

### User Story 4 - Update Design System (Priority: P2)

As a developer who has updated my project's global CSS or design tokens, I want to run `sl mockup update` to refresh the design system so that future mockups reflect current styling.

**Why this priority**: Allows developers to keep the design system in sync with their codebase as styles evolve.

**Independent Test**: Can be tested by modifying global CSS/tokens in a frontend project and running `sl mockup update`. The system should re-extract and update `.specledger/memory/design-system.md`.

**Acceptance Scenarios**:

1. **Given** a frontend repository with an existing `.specledger/memory/design-system.md`, **When** the user modifies global CSS or Tailwind config and runs `sl mockup update`, **Then** the system re-extracts design tokens and updates the design system.
2. **Given** a design system that has been manually edited, **When** the user runs `sl mockup update`, **Then** the system respects manual additions/modifications and merges new discoveries with existing entries.
3. **Given** a project where CSS variables have been removed, **When** the user runs `sl mockup update`, **Then** the system identifies and removes stale entries from the design system.

---

### User Story 5 - Initialize Design System During Onboarding (Priority: P3)

As a new user onboarding a frontend project to SpecLedger, I want the `sl init` / onboarding process to automatically create `.specledger/memory/design-system.md` so that the design system index is ready when I first use `sl mockup`.

**Why this priority**: Improves the onboarding experience but is not blocking since US3 handles the case where the file is missing at mockup time.

**Independent Test**: Can be tested by running `sl init` on a frontend project and verifying that `.specledger/memory/design-system.md` is created alongside other initialization artifacts.

**Acceptance Scenarios**:

1. **Given** a frontend repository being initialized with `sl init`, **When** onboarding completes, **Then** `.specledger/memory/design-system.md` exists in the specledger directory with an auto-generated index of UI components.
2. **Given** a non-frontend repository being initialized with `sl init`, **When** onboarding completes, **Then** `.specledger/memory/design-system.md` is NOT created (since mockup is not applicable).

---

### Edge Cases

- What happens when the design system file exists but is empty or malformed? The system should treat it as missing and re-generate it, warning the user.
- What happens when the spec.md has no user scenarios or functional requirements to mockup? The system should abort with a message directing the user to complete the spec first.
- What happens when the frontend project uses an uncommon framework not in the detection heuristics? The system should provide a `--force` flag to bypass detection and allow the user to proceed.
- What happens when the user runs `sl mockup` without providing a spec name and is not on a feature branch? The system should present an interactive picker of available specs via `huh.Select`.
- What happens when the project has complex CSS with many design tokens? The system should extract and organize tokens by category (colors, spacing, typography, etc.) for clarity.
- What happens when no AI agent is available (not installed)? The system should write the prompt to a file and display install instructions (same as `revise` fallback).
- What happens when a mockup file already exists at the output path? The system should prompt for overwrite confirmation before proceeding.
- What happens when the user cancels at any interactive step? The system should exit cleanly with code 0 and no partial output.
- What happens when `--json` is used? The system should run a non-interactive path (skip huh prompts, use flag values directly) and output structured JSON.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST detect whether the current repository is a frontend project by checking for frontend indicators (package.json with frontend dependencies, framework config files, or source directories with frontend component files).
- **FR-002**: System MUST initialize the design system by extracting global CSS (design tokens, CSS variables, theme config) if the repository is a frontend project and `.specledger/memory/design-system.md` is missing.
- **FR-003**: System MUST read and parse `.specledger/memory/design-system.md` to provide the AI agent with the project's design tokens and styling conventions.
- **FR-004**: System MUST generate a mockup based on the feature's primary source files — `spec.md` (user stories, acceptance criteria), `requirements.md` (functional requirements), and `data-model.md` (data structures) — mapping UI needs to existing design system components wherever possible.
- **FR-005**: System MUST auto-generate `.specledger/memory/design-system.md` by extracting global CSS styles, design tokens, and CSS variables when the file does not exist.
- **FR-006**: System MUST integrate with the `sl init` onboarding flow to create `.specledger/memory/design-system.md` for frontend projects during initialization.
- **FR-007**: System MUST output the generated mockup to the feature directory as HTML or JSX (e.g., `specledger/<spec-name>/mockup.html` or `specledger/<spec-name>/mockup.jsx`), defaulting to HTML.
- **FR-013**: System MUST support a `--format` flag accepting `html` or `jsx` to control the mockup output format.
- **FR-008**: System MUST support extracting design tokens from common styling approaches: CSS variables, Tailwind config (v3 and v4), styled-components themes, CSS-in-JS theme objects, SCSS/Less variables, UnoCSS, Panda CSS, and StyleX.
- **FR-023**: System MUST detect component libraries in use (e.g., shadcn/ui, MUI, Chakra UI, Ant Design, Mantine, Headless UI, NextUI, daisyUI, PrimeReact, PrimeVue, Vuetify, Element Plus) from package.json dependencies and include them in the design system.
- **FR-024**: System MUST scan app structure — layout files, component directories, and global stylesheets — using framework-aware heuristics (e.g., Next.js App Router vs Pages Router, SvelteKit `+layout.svelte`, Nuxt `layouts/`).
- **FR-025**: The AI agent prompt MUST instruct the agent to read `.specledger/memory/design-system.md` directly rather than embedding style data in the prompt, keeping the prompt concise and the design system as the single source of truth.
- **FR-009**: System MUST handle the case where `spec.md` contains no user scenarios by displaying a helpful error message directing the user to run the specify workflow first.
- **FR-010**: System MUST allow users to manually edit `.specledger/memory/design-system.md` and respect manual additions/modifications on subsequent runs.
- **FR-011**: System MUST provide a `--force` flag to bypass frontend detection for edge cases (e.g., uncommon frameworks).
- **FR-012**: System MUST support `sl mockup update` command to refresh the design system by re-extracting global CSS and design tokens.
- **FR-014**: System MUST support a `--dry-run` flag that writes the generated prompt to a file instead of launching the AI agent. _(US1)_
- **FR-015**: System MUST support a `--summary` flag for compact output suitable for agent integration and CI environments. _(US1)_
- **FR-016**: System MUST allow spec-name to be optional, auto-detecting from the current branch via `issues.NewContextDetector`. If not on a feature branch and no argument given, present an interactive spec picker. _(US1)_
- **FR-017**: System MUST launch the configured AI agent with the generated prompt using `launcher.LaunchWithPrompt()`. If no agent is available, fall back to writing the prompt to a file with install instructions. _(US1)_
- **FR-018**: The AI agent MUST ask the user if they want to commit and push after generating the mockup. Git operations are handled within the agent session. CLI provides a fallback commit flow if user exits agent without committing. _(US1)_
- **FR-019**: System MUST display interactive confirmation at key steps — design system generation and prompt review — using `huh` forms and `lipgloss` displays. Confirmations are skipped when user provides arguments or uses `-y` flag. _(US1, US2, US3)_
- **FR-020**: System MUST extract shared editor and prompt utilities from the `revise` package into a reusable `pkg/cli/prompt/` package for use by both `revise` and `mockup` commands. _(US1)_
- **FR-021**: System MUST support `-y` / `--yes` flag to skip all confirmations and launch agent directly. _(US1)_
- **FR-022**: System MUST accept positional arguments as additional instructions for the AI agent. When arguments are provided, confirmations are skipped. _(US1)_

### Key Entities

- **Design System**: A markdown file (`.specledger/memory/design-system.md`) with YAML frontmatter containing machine-readable data (framework, style tokens, app structure, component libs) and a human-readable markdown body (overview table, ASCII directory tree, color palette). The AI agent reads this file directly at runtime.
- **App Structure**: Part of the design system — describes the project's router type, layout files, component directories (1-level deep, max 50), and global stylesheets (max 30). Framework-aware: Next.js distinguishes App Router vs Pages Router; SvelteKit looks for `+layout.svelte`; Nuxt checks `layouts/`; etc.
- **Mockup**: An HTML or JSX file representing the feature's UI. **Generated by the AI agent, not by Go code.** The agent reads the design system and codebase to match existing visual style.
- **Frontend Detection Result**: The outcome of the repository type check — identifies the frontend framework in use (11 frameworks supported), confidence score, component directories, and indicators.
- **Mockup Prompt Context**: The template rendering context passed to the agent prompt — includes spec name, spec path/title, framework, format, output path, and user prompt. Style/design-system data is NOT embedded — the prompt template instructs the agent to read `.specledger/memory/design-system.md` directly.
- **Spec Content**: Parsed content from `spec.md` — title, user stories, requirements, and full content text used to build the agent prompt.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Users can generate a mockup from an existing spec in under 30 seconds (excluding initial design system scan).
- **SC-002**: Generated mockups use the project's design tokens (colors, spacing, typography) consistently.
- **SC-003**: Frontend detection correctly identifies frontend projects across React, Next.js, Vue, Nuxt, Svelte, SvelteKit, Angular, Astro, SolidJS, Qwik, and Remix projects.
- **SC-004**: The auto-generated design system captures the project's design tokens (CSS variables, Tailwind config, theme objects).
- **SC-005**: Users running `sl mockup <spec-name>` on a frontend repo receive a mockup within 30 seconds.
- **SC-006**: After onboarding a frontend project, `.specledger/memory/design-system.md` is present and populated without additional user action.
- **SC-007**: Running `sl mockup update` refreshes the design system (re-extracts design tokens) in under 10 seconds.

### Previous work

- **597-issue-create-fields**: Issue create fields enhancement — most recent CLI command addition, establishes patterns for new commands.
- **011-streamline-onboarding**: Streamlined onboarding — directly related since this feature extends the onboarding flow to initialize design-system.md.

## Scope & Boundaries

### In Scope

- New `sl mockup [spec-name]` CLI command with interactive TUI flow
- New `sl mockup update` CLI command
- Automatic frontend repository detection
- Design system file format (global CSS/design tokens) and auto-generation
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
- Component indexing or cataloging (the AI agent discovers components via codebase search)
- Component screenshot generation or rendering
- Figma/Sketch/design tool integration
- Component dependency graph or usage analytics
- Cross-project design system sharing

## Dependencies & Assumptions

### Assumptions

- The AI agent generates the mockup content (HTML or JSX) — Go code orchestrates the interactive flow, gathers context, builds the prompt, and launches the agent
- The mockup output format is HTML or JSX (component-based layouts with annotations), not graphical design files
- Frontend detection uses file-based heuristics (checking for package.json, framework configs, component file extensions) rather than requiring user configuration
- The design system follows a structured markdown format documenting design tokens that both humans and the system can read/update
- The AI agent is responsible for discovering and using existing components via codebase search — the design system only provides styling context
- The mockup command follows the same interactive pattern as `sl revise` — editor integration, agent launch, commit/push flow
- Spec-name is optional; the system auto-detects from the current branch name using `issues.NewContextDetector`

### Dependencies

- Requires a valid spec name pointing to `specledger/<spec-name>/spec.md` file (provided or auto-detected)
- Relies on existing `sl init` infrastructure for onboarding integration
- Uses the project's file system structure to detect components (no external API calls needed)
- Depends on `pkg/cli/launcher/` for AI agent launch (existing package)
- Depends on `pkg/cli/revise/editor.go` and `prompt.go` for shared editor/prompt utilities (to be extracted into `pkg/cli/prompt/`)
- Depends on `pkg/issues/context.go` for spec auto-detection from branch name
