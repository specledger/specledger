# Feature Specification: Mockup Command

**Feature Branch**: `598-mockup-command`
**Created**: 2026-02-26
**Status**: Draft
**Input**: User description: "Create new command 'sl mockup' which checks if the repository is a frontend repository, reads specledger/design_system.md for UI component indexing, generates mockups based on spec.md using the design system, creates design_system.md if missing, and initializes it during onboarding."

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Generate Mockup from Spec Using Design System (Priority: P1)

As a frontend developer working on a feature, I want to run `sl mockup` so that I get a visual mockup of the feature described in `spec.md` that reuses my project's existing UI components rather than inventing new ones from scratch.

**Why this priority**: This is the core value proposition — generating mockups that are grounded in the project's actual design system ensures consistency and reduces rework.

**Independent Test**: Can be fully tested by running `sl mockup` in a frontend project that has both a `spec.md` and a `specledger/design_system.md`. The output should be a mockup artifact that references existing components from the design system.

**Acceptance Scenarios**:

1. **Given** a frontend repository with a valid `specledger/design_system.md` and a current feature's `spec.md`, **When** the user runs `sl mockup`, **Then** the system generates a mockup that maps UI elements to existing design system components and outputs it to the feature directory.
2. **Given** a frontend repository with a design system containing a Button, Form, and Card component, **When** the user runs `sl mockup` for a feature requiring user input, **Then** the mockup references the existing Form and Button components rather than describing generic UI elements.
3. **Given** a spec with multiple user stories at different priorities, **When** the user runs `sl mockup`, **Then** the mockup covers at minimum the P1 user story flows.

---

### User Story 2 - Auto-Create Design System Index When Missing (Priority: P2)

As a frontend developer in a project without a `specledger/design_system.md`, I want `sl mockup` to automatically scan the codebase and generate an initial design system index so that I can immediately generate mockups without manual setup.

**Why this priority**: This removes a setup barrier and makes the command usable out of the box for projects that haven't catalogued their components yet.

**Independent Test**: Can be tested by running `sl mockup` in a frontend repo that lacks `specledger/design_system.md`. The system should create the file by scanning the codebase for UI components, then proceed to generate the mockup.

**Acceptance Scenarios**:

1. **Given** a frontend repository without `specledger/design_system.md`, **When** the user runs `sl mockup`, **Then** the system scans the codebase for UI components, generates `specledger/design_system.md`, and then proceeds to generate the mockup.
2. **Given** a React project with components in `src/components/`, **When** the design system index is auto-generated, **Then** each discoverable component is listed with its file path, component name, and a brief description of its purpose.
3. **Given** a project using a component library (e.g., Material UI, Ant Design), **When** the design system is generated, **Then** both custom project components and identifiable library components used in the project are indexed.

---

### User Story 3 - Abort Gracefully for Non-Frontend Repositories (Priority: P2)

As a developer working on a backend or CLI project, I want `sl mockup` to detect that the repository is not a frontend project and inform me clearly, so I don't waste time on an unsupported workflow.

**Why this priority**: Prevents confusion and wasted effort. Equal priority with US2 as both are critical for a good first-run experience.

**Independent Test**: Can be tested by running `sl mockup` in a Go or Python backend repository. The command should exit with a clear message explaining it only works for frontend projects.

**Acceptance Scenarios**:

1. **Given** a Go backend repository, **When** the user runs `sl mockup`, **Then** the system displays a message like "This repository does not appear to be a frontend project. The mockup command requires a frontend codebase." and exits without generating any files.
2. **Given** an ambiguous repository (e.g., a monorepo with both backend and frontend), **When** the user runs `sl mockup`, **Then** the system either detects the frontend portion or prompts the user to confirm the working directory.

---

### User Story 4 - Initialize Design System During Onboarding (Priority: P3)

As a new user onboarding a frontend project to SpecLedger, I want the `sl init` / onboarding process to automatically create `specledger/design_system.md` so that the design system index is ready when I first use `sl mockup`.

**Why this priority**: Improves the onboarding experience but is not blocking since US2 handles the case where the file is missing at mockup time.

**Independent Test**: Can be tested by running `sl init` on a frontend project and verifying that `specledger/design_system.md` is created alongside other initialization artifacts.

**Acceptance Scenarios**:

1. **Given** a frontend repository being initialized with `sl init`, **When** onboarding completes, **Then** `specledger/design_system.md` exists in the specledger directory with an auto-generated index of UI components.
2. **Given** a non-frontend repository being initialized with `sl init`, **When** onboarding completes, **Then** `specledger/design_system.md` is NOT created (since mockup is not applicable).

---

### Edge Cases

- What happens when the design system file exists but is empty or malformed? The system should treat it as missing and re-generate it, warning the user.
- What happens when the spec.md has no user scenarios or functional requirements to mockup? The system should abort with a message directing the user to complete the spec first.
- What happens when the frontend project uses an uncommon framework not in the detection heuristics? The system should provide a `--force` flag to bypass detection and allow the user to proceed.
- What happens when the user runs `sl mockup` outside of a feature branch (no spec.md context)? The system should abort with a message directing the user to create a feature first.
- What happens when the component scan finds hundreds of components? The index should group components by directory/module and include all of them — the mockup generation step selects only relevant ones.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST detect whether the current repository is a frontend project by checking for frontend indicators (package.json with frontend dependencies, framework config files, or source directories with frontend component files).
- **FR-002**: System MUST abort with a clear, actionable error message when run in a non-frontend repository.
- **FR-003**: System MUST read and parse `specledger/design_system.md` to build an index of available UI components with their names, file paths, and descriptions.
- **FR-004**: System MUST generate a mockup based on the current feature's `spec.md`, mapping UI needs to existing design system components wherever possible.
- **FR-005**: System MUST auto-generate `specledger/design_system.md` by scanning the codebase for UI components when the file does not exist.
- **FR-006**: System MUST integrate with the `sl init` onboarding flow to create `specledger/design_system.md` for frontend projects during initialization.
- **FR-007**: System MUST output the generated mockup to the current feature's specledger directory (e.g., `specledger/<feature>/mockup.md`).
- **FR-008**: System MUST support at minimum React (.tsx/.jsx), Vue (.vue), Svelte (.svelte), and Angular (.component.ts) component detection.
- **FR-009**: System MUST handle the case where `spec.md` contains no user scenarios by displaying a helpful error message directing the user to run the specify workflow first.
- **FR-010**: System MUST allow users to manually edit `specledger/design_system.md` and respect manual additions/modifications on subsequent runs.
- **FR-011**: System MUST provide a `--force` flag to bypass frontend detection for edge cases (e.g., uncommon frameworks).

### Key Entities

- **Design System Index**: A markdown file (`specledger/design_system.md`) that catalogs all UI components in the project — each entry includes the component name, file path, a brief description, and optionally its props/inputs.
- **Mockup**: A markdown-based visual representation of the feature's UI, referencing components from the design system index. Contains screen layouts, component placements, user interaction flows, and annotations.
- **Frontend Detection Result**: The outcome of the repository type check — identifies the frontend framework(s) in use, the component directory structure, and whether the project qualifies as a frontend repository.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Users can generate a mockup from an existing spec in under 30 seconds (excluding initial design system scan).
- **SC-002**: Generated mockups reference existing design system components rather than proposing entirely new UI elements for common patterns.
- **SC-003**: Frontend detection correctly identifies frontend projects across React, Vue, Svelte, and Angular projects.
- **SC-004**: The auto-generated design system index captures discoverable UI components in the scanned codebase.
- **SC-005**: Users who run `sl mockup` on a non-frontend repo receive a clear abort message within 2 seconds.
- **SC-006**: After onboarding a frontend project, `specledger/design_system.md` is present and populated without additional user action.

### Previous work

- **597-issue-create-fields**: Issue create fields enhancement — most recent CLI command addition, establishes patterns for new commands.
- **011-streamline-onboarding**: Streamlined onboarding — directly related since this feature extends the onboarding flow to initialize design_system.md.

## Scope & Boundaries

### In Scope

- New `sl mockup` CLI command
- Frontend repository detection logic
- Design system index file format and auto-generation
- Mockup generation from spec + design system
- Integration with `sl init` for design system initialization
- `--force` flag to bypass frontend detection

### Out of Scope

- Interactive visual mockup editor (mockups are static markdown)
- Design token extraction (colors, typography, spacing)
- Component screenshot generation or rendering
- Figma/Sketch/design tool integration
- Component dependency graph or usage analytics
- Cross-project design system sharing

## Dependencies & Assumptions

### Assumptions

- The mockup output format is markdown-based (ASCII mockups, component references, and flow descriptions), not graphical
- Frontend detection uses file-based heuristics (checking for package.json, framework configs, component file extensions) rather than requiring user configuration
- The design system index follows a structured markdown format that both humans and the system can read/update
- Component scanning is limited to the project's source directories and does not traverse node_modules or vendor directories
- The mockup command operates on the current feature branch's spec.md — the user must be on a feature branch

### Dependencies

- Requires an active feature branch with a `spec.md` file (from the specify workflow)
- Relies on existing `sl init` infrastructure for onboarding integration
- Uses the project's file system structure to detect components (no external API calls needed)
