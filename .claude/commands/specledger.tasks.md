---
description: Generate an actionable, dependency-ordered tasks.md for the feature based on available design artifacts.
handoffs:
  - label: Analyze For Consistency
    agent: specledger.analyze
    prompt: Run a project analysis for consistency
    send: true
  - label: Implement Project
    agent: specledger.implement
    prompt: Start the implementation in phases
    send: true
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Purpose

Generate actionable, dependency-ordered tasks from the implementation plan. Tasks are created using the built-in issue tracking system for tracking through implementation.

**When to use**: After `/specledger.plan` completes successfully.

## Outline

1. **Setup**: Run `.specledger/scripts/bash/check-prerequisites.sh --json` from repo root and parse FEATURE_DIR and AVAILABLE_DOCS list. All paths must be absolute. For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

2. **Load design documents**: Read from FEATURE_DIR:
   - **Required**: plan.md (tech stack, libraries, structure), spec.md (user stories with priorities)
   - **Optional**: data-model.md (entities), contracts/ (API endpoints), research.md (decisions), quickstart.md (test scenarios)
   - Note: Not all projects have all documents. Generate tasks based on what's available.

3. *** Epic Creation**: From spec.md, create an epic using `sl issue` for the feature using the feature name from plan.md.
   - Use the CLI `sl issue create` to create one top-level issue of type `epic`
   - Title: pulled from plan.md or feature folder name
   - Description: summarize from plan.md + top of spec.md
   - Labels:
     - `spec:<feature-slug>` from feature folder name (e.g. `spec:006-authz-authn-rbac`)
     - `component:<primary-components>` from plan.md (e.g. `component:webapp`)
   - The system auto-generates the ID in format `SL-xxxxxx`

4. **Execute task generation workflow** (follow the template structure):
   - Load plan.md and extract tech stack, libraries, project structure
   - **Load spec.md and extract user stories with their priorities (P1, P2, P3, etc.)**
   - If data-model.md exists: Extract entities → map to user stories
   - If contracts/ exists: Each file → map endpoints to user stories
   - If research.md exists: Extract decisions → generate setup tasks
   - **Generate tasks ORGANIZED BY USER STORY**:
     - Setup tasks (shared infrastructure needed by all stories)
     - **Foundational tasks (prerequisites that must complete before ANY user story can start)**
     - For each user story (in priority order P1, P2, P3...):
       - Group all tasks needed to complete JUST that story
       - Include models, services, endpoints, UI components specific to that story
       - Mark which tasks depends on others
       - If tests requested: Include tests specific to that story
     - Polish/Integration tasks (cross-cutting concerns)
   - **Tests are OPTIONAL**: Only generate test tasks if explicitly requested in the feature spec or user asks for TDD approach
   - Apply task rules:
     - Different files (no dependencies) = allow parallel
     - Same file = sequential requires dependency
     - If tests requested: Tests before implementation (TDD order)
   - Number tasks sequentially (T001, T002...)
   - Generate dependency graph showing user story completion order
   - Create parallel execution examples per user story
   - Validate task completeness (each user story has all needed tasks, independently testable)

5. **Generate tasks.md**: Use `.specledger/templates/tasks-template.md` as structure, fill with:
   - Correct feature name from plan.md
   - Correct Epic ID and Labels
   - Correct User stories source, research input reference, planning details, data model and contracts paths (if available)
   This file does **not** list every task.
   - tasks.md acts as an index for querying issues
   - For each phase and task described below also create corresponding issues:
   - Phase 1: Setup tasks (project initialization) create a `feature`-type issue:
      - Use `sl issue create --type feature`
      - Labels:
         - `phase:<phase-name>` (e.g. `phase:setup`, `phase:us1`, `phase:foundational`)
         - All carry the `spec:<feature-slug>` label
      - For each task, create with `--type task`
      - Tasks must include all content in the `--description` field (there is no separate design/acceptance/dod flag):
         - `--title` (short summary, under 80 characters)
         - `--description` Multi-line text containing:
           - WHY: Problem statement (what to implement, where, inputs/outputs)
           - DESIGN: HOW to build, which files, references (can change during work)
           - ACCEPTANCE: WHAT success looks like (stable criteria)
           - DEFINITION OF DONE: Checklist items [ ] derived from acceptance criteria
         - `--priority` (from story priority, 0=critical, 1=high, 2=normal, 3=low)
         - Labels:
            - `story:US1`, `story:US2`, etc. (mapped from spec.md)
            - `component:<area>` (e.g. `component:auth`, `component:db`)
            - `fr:<requirement-id>` if available
            - `spec:<feature-slug>`
   - Phase 2: Foundational tasks (blocking prerequisites for all user stories)
   - Phase 3+: One phase per user story (in priority order from spec.md)
     - Each phase includes: story goal, independent test criteria, tests (if requested), implementation tasks
     - Clear [Story] labels (US1, US2, US3...) for each task
     - Issue dependencies to identify parallelizable tasks within each story
     - Checkpoint markers after each story phase
   - Structure:
     - Top-level Epic ID + description
     - List of features (setup, foundational, stories)
     - Commands to filter each group:
       ```bash
       sl issue list --label "spec:<slug>"
       sl issue list --status open
       ```
     - MVP and incremental delivery summary
     - Links back to spec.md and plan.md
   - Final Phase: Polish & cross-cutting concerns
   - Review priorities for execution order
   - Review ensuring clear file paths for each task
   - Review task dependencies
   - Implementation strategy section (MVP first, incremental delivery)
   - Suggested MVP scope (typically just User Story 1)

6. **Report**: Provide a summary:
   - Total tasks created
   - Breakdown per user story
   - Story testability (is each story independently verifiable?)
   - Parallel opportunities identified
   - Independent test criteria for each story
   - Suggested MVP (usually US1 or first P1 story)

Context for task generation: $ARGUMENTS

The tasks.md should be immediately executable - each task must be specific enough that an LLM can complete it without additional context. Particularly focus on breaking down by user story to enable independent implementation and testing against agreed interfaces and with clear instructions not to duplicate work (Provide assumptions based on task completion by other agents for parallel work).

## Task Generation Rules

**IMPORTANT**: Tests are optional. Only generate test tasks if the user explicitly requested testing or TDD approach in the feature specification.

**CRITICAL**: Tasks MUST be organized by user story to enable independent implementation and testing.

1. **From User Stories (spec.md)** - PRIMARY ORGANIZATION:
   - Each user story (P1, P2, P3...) gets its own phase (feature issue)
   - Map all related components to their story:
     - Models needed for that story
     - Services needed for that story
     - Endpoints/UI needed for that story
     - If tests requested: Tests specific to that story
   - Mark story dependencies (most stories should be independent)

2. **From Contracts**:
   - Map each contract/endpoint → to the user story it serves
   - If tests requested: Each contract → contract test task [P] before implementation in that story's phase

3. **From Data Model**:
   - Map each entity → to the user story(ies) that need it
   - If entity serves multiple stories: Put in earliest story or Setup phase
   - Relationships → service layer tasks in appropriate story phase

4. **From Setup/Infrastructure**:
   - Shared infrastructure → Setup phase (Phase 1)
   - Foundational/blocking tasks → Foundational phase (Phase 2)
     - Examples: Database schema setup, authentication framework, core libraries, base configurations
     - These MUST complete before any user story can be implemented
   - Story-specific setup → within that story's phase

5. **Ordering**:
   - Phase 1: Setup (project initialization)
   - Phase 2: Foundational (blocking prerequisites - must complete before user stories)
   - Phase 3+: User Stories in priority order (P1, P2, P3...)
     - Within each story: Tests (if requested) → Models → Services → Endpoints → Integration
   - Final Phase: Polish & Cross-Cutting Concerns
   - **DO NOT generate a linear checklist** — build a **task graph** using dependencies
   - Each user story phase should be a complete, independently testable increment

## Issue Content Structure

Each generated issue MUST have the following structured content in the `--description` field:

| Field | Purpose | Example |
|-------|---------|---------|
| `--title` | Concise summary (under 80 chars) | "Add user authentication to login page" |
| `--description` | Multi-line content with sections | See format below |

**Description Format** (all in one `--description` string):
```
WHY: [Problem statement - what to implement, where, inputs/outputs]

DESIGN: [HOW to build - file paths, module references, approach decisions]

ACCEPTANCE: [WHAT success looks like - measurable, testable outcomes]

DEFINITION OF DONE:
[ ] Item 1
[ ] Item 2
[ ] Item 3
```

**Field Guidelines**:
- `--title`: Action-oriented, specific, concise
- `--description`: Include WHY, DESIGN, ACCEPTANCE, and DEFINITION OF DONE sections
- All structured content goes in description since `sl issue create` only supports `--title` and `--description` flags

## Definition of Done Population

When creating issues, derive `definition_of_done` items from the acceptance criteria in spec.md:

1. **Extract acceptance criteria**: Parse each "Then" clause from acceptance scenarios in spec.md
2. **Convert to checklist items**: Transform each criterion into a verifiable statement
3. **Include in issue creation**: Add items to the `definition_of_done` field

**Example conversion**:
- Spec acceptance: "Then the user can log in with valid credentials"
- DoD item: "User can authenticate with valid username/password"
- Spec acceptance: "Then invalid credentials show an error message"
- DoD item: "Invalid credentials display appropriate error message"

**DoD Summary in tasks.md**: After creating all issues, include a DoD Summary section in tasks.md:

```markdown
## Definition of Done Summary

| Issue ID | DoD Items |
|----------|-----------|
| SL-xxxxx | - Item 1<br>- Item 2<br>- Item 3 |
| SL-yyyyy | - Item 1<br>- Item 2 |
```

This provides quick visibility into verification requirements for each task.

## Label Conventions

| Label                | Purpose                             |
| -------------------- | ----------------------------------- |
| `spec:<slug>`        | All tasks in this feature spec      |
| `phase:<name>`       | Setup, US1, polish, etc.            |
| `story:US1`          | User story traceability             |
| `requirement:FR-001` | Functional requirement traceability |
| `component:<area>`   | Mapping to plan-defined modules     |
| `test:<type>`        | Test-related tasks                  |

## Example CLI Calls

### Epic

```bash
sl issue create --title "Login Feature" --description "Implement user authentication with OAuth2 support" --type epic --labels "spec:006-login-auth,component:webapp" --priority 1
```

### Feature (Phase)

```bash
sl issue create --title "Setup Phase" --description "Initialize project structure and dependencies" --type feature --labels "spec:006-login-auth,phase:setup,component:infra" --priority 1
```

### Task

```bash
sl issue create --title "Add React LoginForm" --description "WHY: Users need to authenticate via the login form.

DESIGN: Files: src/components/LoginForm.tsx, src/hooks/useAuth.ts. Use React Hook Form for validation.

ACCEPTANCE: User can enter credentials, form validates input, submits to auth API.

DEFINITION OF DONE:
[ ] LoginForm component created
[ ] Form validation working
[ ] Auth API integration complete" --type task --labels "spec:006-login-auth,story:US1,component:webapp" --priority 2
```

### Add Dependencies

```bash
sl issue link SL-xxxxx blocks SL-yyyyy
```

**IMPORTANT**: The `sl issue create` command only supports these flags:
- `--title` (required)
- `--description` (optional, but include all WHY/DESIGN/ACCEPTANCE/DoD content here)
- `--type` (epic, feature, task, bug)
- `--labels` (comma-separated)
- `--priority` (0-5)
- `--spec` (override spec context)
- `--force` (skip duplicate detection)

There are NO separate `--design`, `--acceptance`, or `--definition_of_done` flags.

## Error Handling

When `sl issue create` or `sl issue link` commands fail, handle errors gracefully:

### Automatic Error Recovery

1. **Sanitize special characters**: If description contains quotes, newlines, or special characters:
   - Escape single quotes: `'` → `'\''`
   - Escape double quotes within strings
   - Replace literal newlines with `\n` or remove if problematic
   - Sanitize any shell metacharacters

2. **Retry with corrected parameters**:
   - First attempt: Use original parameters
   - If fails: Sanitize and retry once
   - If still fails: Proceed to manual error handling

3. **Report clear errors**:
   - Display the specific error message from the CLI
   - Identify which parameter caused the issue
   - Suggest remediation steps

### Common Error Scenarios

| Error | Cause | Resolution |
|-------|-------|------------|
| "label format invalid" | Special characters in label | Sanitize to alphanumeric + dashes |
| "description too long" | Field exceeds limit | Truncate with ellipsis, log warning |
| "duplicate issue" | Same title/labels exist | Skip with warning, use `--force` if intentional |
| "file system error" | Permissions, disk space | Display path, suggest remediation |

### Example Error Handling Flow

```bash
# Attempt 1: Original command
sl issue create --title "Feature" --description "With 'quotes'" --type task

# If fails with quote error, sanitize and retry:
sl issue create --title "Feature" --description "With '\''quotes'\''" --type task

# If still failing, report and continue:
echo "Warning: Could not create issue 'Feature'. Error: [specific error]"
echo "Suggestion: [remediation step]"
```
