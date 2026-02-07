# Feature Specification: Embedded Templates

**Feature Branch**: `005-embedded-templates`
**Created**: 2026-02-07
**Status**: Draft
**Input**: User description: "make it not platform agnostic, it can support multiple tamplate from remote but for now the opensource just embed the current tempalte"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Project with Embedded Templates (Priority: P1)

A developer creates a new SpecLedger project and wants the templates (Claude Code commands, skills, bash scripts, and file templates) to be automatically set up from the embedded Spec Kit playbook templates included with SpecLedger.

**Why this priority**: This is the core MVP functionality - without embedded templates, users would have to manually copy or configure templates for every new project, significantly reducing the value proposition of SpecLedger as a bootstrap tool.

**Independent Test**: Can be fully tested by running `sl new --framework speckit` and verifying that:
1. The `.claude/` folder contains SpecLedger commands and skills
2. The `specledger/scripts/` folder contains helper scripts
3. The `specledger/templates/` folder contains file templates
4. The `.beads/` folder is initialized with configuration
5. All template files match the embedded versions

**Acceptance Scenarios**:

1. **Given** a user runs `sl new myproject --framework speckit`, **When** project creation completes, **Then** the project contains all embedded template files in the correct locations
2. **Given** a user runs `sl init --framework speckit` in an existing repo, **When** initialization completes, **Then** the repo contains all embedded template files
3. **Given** embedded templates exist, **When** a project is created, **Then** all template files are copied from the embedded templates folder to the project

---

### User Story 2 - List Available Template Playbooks (Priority: P2)

A developer wants to see what template/playbook options are available before creating a project. SpecLedger should show which playbooks are included (currently: Spec Kit).

**Why this priority**: Important for user discovery and setting expectations for what's available. Users need to know what frameworks/playbooks are supported before creating a project.

**Independent Test**: Can be tested by running a command like `sl template list` and verifying it displays available embedded templates (currently just "speckit" from the Spec Kit playbook).

**Acceptance Scenarios**:

1. **Given** a user runs `sl template list`, **When** the command executes, **Then** it shows "speckit" as an available embedded template
2. **Given** a user runs `sl new --framework`, **When** they press tab for autocomplete, **Then** it shows available template options

---

### User Story 3 - Future: Remote Template Support (Priority: P3)

A developer wants to use a custom template playbook from a remote repository (e.g., a company-specific template hosted on GitHub). SpecLedger should support fetching and applying templates from remote URLs.

**Why this priority**: This is a future enhancement. The current MVP focuses on embedded templates, but the architecture should support remote templates without major refactoring.

**Independent Test**: Can be tested by running `sl new --template-url https://github.com/org/custom-template` and verifying templates are fetched and applied from the remote URL.

**Acceptance Scenarios**:

1. **Given** a user specifies a remote template URL, **When** project creation runs, **Then** SpecLedger clones the template repo and applies its templates
2. **Given** a remote template is specified, **When** the URL is invalid or inaccessible, **Then** SpecLedger shows a clear error message

---

### Edge Cases

- What happens when embedded templates folder is missing or corrupted?
- How does system handle conflicts when template files already exist in target project?
- What happens if user specifies both `--framework` and `--template-url` flags?
- How does system handle remote template URLs that require authentication?
- What happens when remote template repository has invalid structure?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: SpecLedger MUST ship with embedded Spec Kit playbook templates in the codebase
- **FR-002**: When creating a new project with `sl new --framework speckit`, system MUST copy all embedded template files to the new project
- **FR-003**: When initializing with `sl init --framework speckit`, system MUST copy all embedded template files to the existing repository
- **FR-004**: System MUST preserve directory structure when copying templates (e.g., `.claude/`, `specledger/scripts/`, `specledger/templates/`)
- **FR-005**: System MUST support listing available embedded templates via command
- **FR-006**: System architecture MUST support future remote template fetching without major refactoring
- **FR-007**: System MUST validate that embedded templates folder exists before project creation
- **FR-008**: System MUST show clear error messages if template files cannot be copied
- **FR-009**: System MUST handle existing files gracefully (skip, overwrite, or merge based on strategy)
- **FR-010**: System MUST support multiple template playbooks in embedded folder (prepare for future additions like OpenSpec templates)

### Key Entities

- **Embedded Template**: A collection of files (commands, skills, scripts, file templates) that define a SpecLedger playbook for a specific SDD framework (e.g., Spec Kit, OpenSpec)
- **Template Manifest**: Metadata file describing the template (name, version, framework type, required tools)
- **Template Source**: Location of templates - currently "embedded" in codebase, future supports "remote" URLs

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create a new SpecLedger project with all templates applied in under 10 seconds
- **SC-002**: 100% of embedded template files are successfully copied to new projects
- **SC-003**: Users can discover available templates without documentation lookup (via `sl template list`)
- **SC-004**: Template architecture supports adding new embedded playbooks (e.g., OpenSpec) without code changes to core logic
- **SC-005**: Remote template fetching can be implemented as a feature addition without refactoring embedded template logic

## Assumptions

1. Current embedded templates are based on the Spec Kit playbook from the `001-sdd-control-plane` branch
2. Embedded templates are stored in `templates/` folder in the SpecLedger repository
3. Template copying happens during project creation/initialization phase
4. Remote template support is a future feature - current scope is embedded templates only
5. Template files are static (no template variable substitution required in this phase)
6. Users have read permissions to the embedded templates folder during project creation

## Scope Exclusions

- Remote template fetching (deferred to future feature)
- Template variable substitution/interpolation (templates are copied as-is)
- Template version management or updates
- Custom template creation tools (users manually edit templates after project creation)
- Template validation or schema enforcement
- Hybrid template scenarios (mixing multiple templates)
