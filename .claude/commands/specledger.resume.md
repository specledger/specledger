---
description: Execute the implementation plan by processing and executing all tasks defined in tasks.md
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Outline

1. Run `.specledger/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks` from repo root and parse FEATURE_DIR and AVAILABLE_DOCS list. All paths must be absolute. For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

2. Read tasks.md structure and use `sl issue` to extract:
   - **Task phases**: Setup, Tests, Core, Integration, Polish
   - **Task dependencies**: Sequential vs parallel execution rules
   - **Task details**: ID, description, file paths, design + acceptance criteria
   - **Task comments**: Important notes and modifications to original plan
   - **Execution flow**: Order and dependency requirements

3. Execute implementation following the task plan:
   - **Phase-by-phase execution**: Complete each phase before moving to the next
   - **Respect dependencies**: Run sequential tasks in order, parallel tasks [P] can run together
   - **File-based coordination**: Tasks affecting the same files must run sequentially
   - **Validation checkpoints**: Verify each phase completion before proceeding

4. Implementation execution rules:
   - **Setup first**: Initialize project structure, dependencies, configuration
   - **Core development**: Implement models, services, CLI commands, endpoints
   - **Integration work**: Database connections, middleware, logging, external services
   - **Polish and validation**: Unit tests, performance optimization, documentation

5. Progress tracking and error handling:
   - Find open tasks using: `sl issue list --status open --label "spec:<slug>"`
   - Update issue status: `sl issue update <id> --status in_progress`
   - Report progress after each completed task
   - Halt execution if any sequential task fails
   - For ready tasks, continue with successful tasks, report failed ones
   - Provide clear error messages with context for debugging
   - Suggest next steps if implementation cannot proceed
   - **IMPORTANT** For completed tasks: `sl issue close <id> --reason "Completed"`

6. Completion validation:
   - Verify all required tasks are completed
   - Check that implemented features match the original specification
   - Validate that tests pass and coverage meets requirements
   - Confirm the implementation follows the technical plan
   - Report final status with summary of completed work

## Issue Tracking Commands

| Action | Command |
|--------|---------|
| List open issues | `sl issue list --status open` |
| Show issue details | `sl issue show <id>` |
| Update issue | `sl issue update <id> --status in_progress` |
| Close issue | `sl issue close <id> --reason "Completed"` |

Note: This command assumes a complete task breakdown exists in tasks.md. If tasks are incomplete or missing, suggest running `/specledger.tasks` first to regenerate the task list.
