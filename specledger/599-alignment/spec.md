# Feature Specification: SDD Layer Alignment

**Feature Branch**: `599-alignment`
**Created**: 2026-03-02
**Status**: Draft
**Input**: User description: "alignment (update skills, cli, commands), create new task but refer to 598-sdd-workflow-streamline spec and plan"

## Overview

Standardize and align AI skills, AI commands, and CLI patterns across the SDD workflow to reduce overlapping responsibilities and ensure consistent layer interactions.

**Scope Note**: This is **Stream 1** of a 3-stream alignment effort:
| Stream | Focus | Feature |
|--------|-------|---------|
| **1** | AI skills/commands/CLI alignment (this spec) | 599-alignment |
| 2 | Bash script → Go CLI migration | 598-sdd-workflow-streamline |
| 3 | New CLI + skills (comment, checkpoint, research) | Future |

**Layer Architecture** (from 598):
| Layer | Name | Runtime | Purpose |
|-------|------|---------|---------|
| 0 | Hooks | Invisible, event-driven | Auto-capture sessions on commit |
| 1 | `sl` CLI | Go binary, no AI needed | Data operations, CRUD, standalone tooling |
| 2 | AI Commands | Agent shell prompts | AI workflow orchestration (specify→implement) |
| 3 | Skills | Passive context injection | Domain knowledge, progressively loaded |

**Core Problem**: Overlapping responsibilities between layers cause confusion:
- AI commands (L2) contain business logic that belongs in CLI (L1)
- Skills (L3) duplicate information already in CLI help text
- Inconsistent patterns across commands increase cognitive load
- No clear boundaries between what's CLI vs. AI command responsibility

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Define Layer Responsibilities (Priority: P1)

As a SpecLedger maintainer, I need clear documentation of what belongs in each layer (CLI vs. AI Commands vs. Skills) so that future development follows consistent patterns.

**Why this priority**: Without clear boundaries, layers become entangled and maintenance burden increases.

**Independent Test**: Review layer responsibility doc; verify it provides decision criteria for "should this go in CLI or AI command?"

**Acceptance Scenarios**:

1. **Given** a new feature idea, **When** applying layer decision criteria, **Then** it's clear whether it belongs in L1 (CLI), L2 (AI Command), or L3 (Skill)
2. **Given** existing overlapping functionality, **When** reviewing layer doc, **Then** it's clear which layer owns it
3. **Given** the CLI constitution (from 598), **When** layer doc is created, **Then** it references and extends CLI patterns

---

### User Story 2 - Audit Existing Overlaps (Priority: P1)

As a SpecLedger maintainer, I need an audit of where AI commands and skills duplicate CLI functionality so that I can identify consolidation opportunities.

**Why this priority**: Cannot reduce overlap without first identifying where it exists.

**Independent Test**: Review audit output; verify it lists specific files/functions with overlap classification.

**Acceptance Scenarios**:

1. **Given** all AI command files, **When** audited, **Then** each file is classified: pure-orchestration / has-business-logic / duplicates-CLI
2. **Given** all skill files, **When** audited, **Then** each skill is classified: unique-knowledge / duplicates-help / outdated
3. **Given** the audit report, **When** reviewed, **Then** it provides actionable recommendations (keep/merge/remove) for each item

---

### User Story 3 - Standardize AI Command Patterns (Priority: P2)

As an AI agent developer, I need AI commands to follow a consistent structure so that I can easily understand and modify any command.

**Why this priority**: Consistency reduces cognitive load; P2 because commands work today but are hard to maintain.

**Independent Test**: Compare 3 AI command files; verify they follow the same structural template.

**Acceptance Scenarios**:

1. **Given** an AI command file, **When** reading it, **Then** it follows the standard sections: Purpose, When to Use, Outline, Behavior Rules
2. **Given** multiple AI commands, **When** comparing them, **Then** they use consistent terminology for CLI invocations
3. **Given** an AI command that calls CLI, **When** reviewing the call pattern, **Then** it uses `--json` for structured data and documents expected output

---

### User Story 4 - Standardize Skill Patterns (Priority: P2)

As a skill author, I need a template for what skills should contain so that skills are focused on domain knowledge, not CLI documentation.

**Why this priority**: Skills currently mix domain knowledge with CLI syntax; separating them improves maintainability.

**Independent Test**: Review existing skills against template; verify they focus on "when to use" not "how to use".

**Acceptance Scenarios**:

1. **Given** the skill template, **When** authoring a new skill, **Then** it contains: When to Load, Key Concepts, Decision Patterns, CLI Reference (link only)
2. **Given** an existing skill, **When** reviewing against template, **Then** CLI syntax examples are minimal or delegated to `--help`
3. **Given** skills reference CLI commands, **When** CLI changes, **Then** skill updates are minimal (no embedded syntax to update)

---

### User Story 5 - Document Cross-Layer Interactions (Priority: P3)

As a developer, I need examples of how layers should interact so that I understand the launcher pattern and data flow.

**Why this priority**: Helps new contributors understand architecture; P3 because code works today.

**Independent Test**: Review interaction doc; verify it shows L2→L1 call patterns with examples.

**Acceptance Scenarios**:

1. **Given** the interaction doc, **When** reading it, **Then** it shows the launcher pattern (AI command → CLI → parse output)
2. **Given** the interaction doc, **When** reading it, **Then** it shows the skill loading trigger pattern
3. **Given** the interaction doc, **When** reading it, **Then** it documents L1→L0 (CLI configures hooks) convenience patterns

---

### Edge Cases

- What if an AI command has business logic that can't move to CLI? → Document as exception with justification
- What if a skill needs CLI syntax examples for clarity? → Keep minimal, link to `--help` for full syntax
- What if layers have legitimate shared concerns? → Document in both places with DRY principle noted

---

## Clarifications

### Session 2026-03-02

- Q: What is the scope of this feature? → A: Stream 1 only - AI skills/commands/CLI alignment. Stream 2 (bash migration) in 598, Stream 3 (new CLI/skills) in future features.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Layer responsibility document MUST be created with clear decision criteria for L1/L2/L3 placement
- **FR-002**: Audit report MUST classify all existing AI commands and skills by overlap status
- **FR-003**: AI command template MUST be defined with standard sections (Purpose, When to Use, Outline, Behavior Rules)
- **FR-004**: Skill template MUST be defined focusing on domain knowledge, not CLI syntax duplication
- **FR-005**: Cross-layer interaction patterns MUST be documented with examples
- **FR-006**: All AI commands MUST use `--json` for CLI data retrieval (not parsing human output)
- **FR-007**: Skills MUST link to CLI `--help` for syntax rather than duplicating it
- **FR-008**: Layer responsibility document MUST reference and extend CLI constitution from 598

### Key Entities

- **Layer Responsibility Doc**: Document defining what belongs in L0 (Hooks), L1 (CLI), L2 (AI Commands), L3 (Skills) with decision criteria
- **Overlap Audit**: Report classifying existing commands/skills by overlap status and recommending actions
- **AI Command Template**: Standard structure for `.claude/commands/*.md` files
- **Skill Template**: Standard structure for skill markdown files focusing on domain knowledge

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Layer responsibility document exists with decision criteria for L1/L2/L3 placement
- **SC-002**: Audit report covers 100% of AI commands (`.claude/commands/*.md`) with classification
- **SC-003**: Audit report covers 100% of skills with classification
- **SC-004**: AI command template defined and documented
- **SC-005**: Skill template defined and documented
- **SC-006**: Cross-layer interaction examples documented for L2→L1 and L3 loading patterns

### Previous work

### Epic: 598 - SDD Workflow Streamline

- **CLI Constitution**: Patterns for `sl` CLI commands (Data CRUD, token-efficient output, etc.)
- **Layer Architecture**: Four-layer model (Hooks, CLI, AI Commands, Skills)
- **Launcher Pattern**: AI commands invoke CLI, parse JSON output

## Dependencies & Assumptions

### Dependencies

- **598-sdd-workflow-streamline**: Layer architecture and CLI constitution defined in 598 provide foundation for this alignment work.

### Assumptions

- CLI constitution from 598 is stable and can be extended
- Existing AI commands and skills can be audited without modifying them
- Templates are documentation-only (no code generation required)

## Out of Scope

- **Stream 2**: Bash script → Go CLI migration (covered by 598)
- **Stream 3**: New CLI commands and skills (comment, checkpoint, research) - future features
- Modifying existing AI commands or skills (audit only)
- Changes to the core SDD workflow (specify→clarify→plan→tasks→implement is immutable)
- Code implementation - this is a documentation/audit feature
