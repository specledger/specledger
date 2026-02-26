# Feature Specification: Generate .claudeignore

**Feature Branch**: `322-generate-claudeignore`
**Created**: 2026-02-24
**Status**: In Progress

> **Note:** See [preliminary-research.md](./preliminary-research.md) for background research on `.claudeignore` vs `permissions.deny`.
> See [implementation-plan.md](./implementation-plan.md) for implementation details.

## Overview

Generate .claudeignore or other agent ignore files while running `sl init` or `sl new` commands. This feature automatically creates ignore files that prevent Claude and other AI agents from processing unnecessary files and directories, improving context efficiency and reducing token usage.

## User Scenarios & Testing

### User Story 1 - Initialize Project with Agent Ignore Files (Priority: P1)

As a developer, I want to automatically generate .claudeignore files when initializing a new project so that I don't have to manually configure which files Claude should ignore.

**Acceptance Scenarios**:

1. **Given** a user runs `sl init` in a new project directory, **When** the command completes, **Then** a .claudeignore file is created from a static template embedded in the sl binary, and the agent is instructed to explore the codebase and enhance it with project-specific patterns.

2. **Given** a user runs `sl new <project-name>`, **When** the project is scaffolded, **Then** the generated project includes a .claudeignore file from the static template, and the agent is instructed to enhance it based on the selected project type.

3. **Given** a .claudeignore file already exists, **When** running `sl init`, **Then** the existing file is preserved and the agent suggests improvements based on the current template (user can accept/reject).

---

### User Story 2 - Customize Ignore Patterns (Priority: P2)

As a developer, I want to be able to customize the .claudeignore file to match my project's specific needs so that I can control exactly what Claude sees.

**Acceptance Scenarios**:

1. **Given** a .claudeignore file exists, **When** I edit it with custom patterns, **Then** the changes are respected on subsequent Claude interactions.

---

## Requirements

### Functional Requirements

- **FR-001**: System MUST generate a .claudeignore file when `sl init` is executed in a project without one
- **FR-002**: System MUST generate a .claudeignore file when `sl new <project-name>` is executed
- **FR-003**: System MUST include sensible default ignore patterns (node_modules, .git, build artifacts, etc.)
- **FR-004**: System MUST NOT overwrite existing .claudeignore files
- **FR-005**: System MUST support custom ignore patterns using gitignore-style syntax
- **FR-006**: System MUST document the ignore format and available patterns (both `.claudeignore` and `permissions.deny`)
- **FR-007**: System MUST support language/framework-specific ignore patterns (Python, Node.js, Go, etc.)
- **FR-008**: System MUST allow users to extend default patterns without losing them
- **FR-009**: System SHOULD prefer `permissions.deny` in `.claude/settings.json` as the primary mechanism (modern approach)

### Non-Functional Requirements

- **NFR-001**: .claudeignore generation MUST complete in < 100ms
- **NFR-002**: Default patterns MUST be maintainable and version-controlled
- **NFR-003**: System MUST be compatible with existing .gitignore patterns

## Success Criteria

- **SC-001**: 100% of new projects created with `sl new` include a .claudeignore file
- **SC-002**: 95% of `sl init` executions successfully generate .claudeignore when missing
- **SC-003**: Existing .claudeignore files are preserved in 100% of cases
- **SC-004**: Default patterns cover at least 90% of common project types
- **SC-005**: Documentation is complete and includes examples for 5+ languages/frameworks
