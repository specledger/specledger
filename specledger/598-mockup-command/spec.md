# Feature Specification: Mockup Command

**Feature Branch**: `598-mockup-command`
**Created**: 2026-02-26
**Status**: Draft
**Input**: User description: "Create new command 'sl mockup' which checks if the repository is a frontend repository, reads specledger/design_system.md for UI component indexing, generates mockups based on spec.md using the design system, creates design_system.md if missing, and initializes it during onboarding."

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Generate Mockup from Spec Using Design System (Priority: P1)

As a frontend developer working on a feature, I want to run `sl mockup <spec-name>` so that I get a visual mockup of the feature described in `specledger/<spec-name>/spec.md` that reuses my project's existing UI components rather than inventing new ones from scratch.

**Why this priority**: This is the core value proposition — generating mockups that are grounded in the project's actual design system ensures consistency and reduces rework.

**Independent Test**: Can be fully tested by running `sl mockup <spec-name>` in a frontend project that has both a `specledger/<spec-name>/spec.md` and a `specledger/design_system.md`. The output should be an HTML or JSX mockup file that references existing components from the design system.

**Acceptance Scenarios**:

1. **Given** a frontend repository with a valid `specledger/design_system.md` and a feature's `specledger/<spec-name>/spec.md`, **When** the user runs `sl mockup <spec-name>`, **Then** the system generates a mockup that maps UI elements to existing design system components and outputs it to the feature directory.
2. **Given** a frontend repository with a design system containing a Button, Form, and Card component, **When** the user runs `sl mockup <spec-name>` for a feature requiring user input, **Then** the mockup references the existing Form and Button components rather than describing generic UI elements.
3. **Given** a spec with multiple user stories at different priorities, **When** the user runs `sl mockup <spec-name>`, **Then** the mockup covers at minimum the P1 user story flows.

---

### User Story 2 - Auto-Create Design System Index When Missing (Priority: P2)

As a frontend developer in a project without a `specledger/design_system.md`, I want `sl mockup <spec-name>` to automatically scan the codebase and generate an initial design system index so that I can immediately generate mockups without manual setup.

**Why this priority**: This removes a setup barrier and makes the command usable out of the box for projects that haven't catalogued their components yet.

**Independent Test**: Can be tested by running `sl mockup <spec-name>` in a frontend repo that lacks `specledger/design_system.md`. The system should create the file by scanning the codebase for UI components, then proceed to generate the mockup.

**Acceptance Scenarios**:

1. **Given** a frontend repository without `specledger/design_system.md`, **When** the user runs `sl mockup <spec-name>`, **Then** the system scans the codebase for UI components, generates `specledger/design_system.md`, and then proceeds to generate the mockup.
2. **Given** a React project with components in `src/components/`, **When** the design system index is auto-generated, **Then** each discoverable component is listed with its file path, component name, and a brief description of its purpose.
3. **Given** a project using a component library (e.g., Material UI, Ant Design), **When** the design system is generated, **Then** both custom project components and identifiable library components used in the project are indexed.

---

### User Story 3 - Initialize Design System and Create Frontend Mockup (Priority: P2)

As a developer working on a frontend project, I want `sl mockup <spec-name>` to detect that the repository is a frontend project and initialize a design system if needed, so that I can generate mockups that leverage my project's UI components.

**Why this priority**: This ensures the mockup command works seamlessly for frontend projects and sets up the design system infrastructure automatically.

**Independent Test**: Can be tested by running `sl mockup <spec-name>` in a frontend repository. The system should detect the frontend framework, initialize the design system if missing, and generate the mockup.

**Acceptance Scenarios**:

1. **Given** a React/Vue/Svelte/Angular repository, **When** the user runs `sl mockup <spec-name>`, **Then** the system detects the frontend framework, initializes `specledger/design_system.md` if missing, and generates the mockup.
2. **Given** an ambiguous repository (e.g., a monorepo with both backend and frontend), **When** the user runs `sl mockup <spec-name>`, **Then** the system either detects the frontend portion or prompts the user to confirm the working directory.
3. **Given** a frontend repository with existing design system, **When** the user runs `sl mockup <spec-name>`, **Then** the mockup is generated using the existing design system components.

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
- What happens when the user runs `sl mockup <spec-name>` without providing a spec name? The system should abort with a message directing the user to provide a valid spec name.
- What happens when the component scan finds hundreds of components? The index should group components by directory/module and include all of them — the mockup generation step selects only relevant ones.

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

### Key Entities

- **Design System Index**: A markdown file (`specledger/design_system.md`) that catalogs all UI components in the project — each entry includes the component name, file path, a brief description, and optionally its props/inputs.
- **Mockup**: An HTML or JSX file representing the feature's UI, referencing components from the design system index. Contains screen layouts with component placements, user interaction flows, and annotations. HTML format uses semantic HTML with inline styles; JSX format outputs React-compatible component code.
- **Frontend Detection Result**: The outcome of the repository type check — identifies the frontend framework(s) in use, the component directory structure, and whether the project qualifies as a frontend repository.

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

- New `sl mockup <spec-name>` CLI command
- New `sl mockup update` CLI command
- Frontend repository detection logic
- Design system index file format and auto-generation
- Mockup generation from spec + design system (HTML or JSX output)
- `--format` flag to choose between `html` (default) and `jsx` output
- Integration with `sl init` for design system initialization
- `--force` flag to bypass frontend detection

### Out of Scope

- Interactive visual mockup editor (mockups are static HTML/JSX files)
- Design token extraction (colors, typography, spacing)
- Component screenshot generation or rendering
- Figma/Sketch/design tool integration
- Component dependency graph or usage analytics
- Cross-project design system sharing

## Dependencies & Assumptions

### Assumptions

- The mockup output format is HTML or JSX (component-based layouts with annotations), not graphical design files
- Frontend detection uses file-based heuristics (checking for package.json, framework configs, component file extensions) rather than requiring user configuration
- The design system index follows a structured markdown format that both humans and the system can read/update
- Component scanning is limited to the project's source directories and does not traverse node_modules or vendor directories
- The mockup command operates on a specified feature spec — the user must provide a spec name via `sl mockup <spec-name>`

### Dependencies

- Requires a valid spec name pointing to `specledger/<spec-name>/spec.md` file
- Relies on existing `sl init` infrastructure for onboarding integration
- Uses the project's file system structure to detect components (no external API calls needed)
