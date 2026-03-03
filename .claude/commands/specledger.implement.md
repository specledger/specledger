---
description: Execute the implementation plan by processing and executing all tasks defined in tasks.md
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Purpose

Execute the implementation plan by processing all tasks in tasks.md. This command orchestrates the actual coding work following the defined task order and dependencies.

**When to use**: After `/specledger.tasks` generates the task list.

## Outline

1. **Sync issues before starting** (required):

   ```bash
   git pull --rebase origin $(git branch --show-current)
   ```

   This ensures you have the latest changes from other team members.

2. Run `.specledger/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks` from repo root and parse FEATURE_DIR and AVAILABLE_DOCS list. All paths must be absolute. For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

3. **Check checklists status** (if FEATURE_DIR/checklists/ exists):
   - Scan all checklist files in the checklists/ directory
   - For each checklist, count:
     - Total items: All lines matching `- [ ]` or `- [X]` or `- [x]`
     - Completed items: Lines matching `- [X]` or `- [x]`
     - Incomplete items: Lines matching `- [ ]`
   - Create a status table:
     ```text
     | Checklist | Total | Completed | Incomplete | Status |
     |-----------|-------|-----------|------------|--------|
     | ux.md     | 12    | 12        | 0          | ✓ PASS |
     | test.md   | 8     | 5         | 3          | ✗ FAIL |
     | security.md | 6   | 6         | 0          | ✓ PASS |
     ```
   - Calculate overall status:
     * **PASS**: All checklists have 0 incomplete items
     * **FAIL**: One or more checklists have incomplete items

   - **If any checklist is incomplete**:
     * Display the table with incomplete item counts
     * **STOP** and ask: "Some checklists are incomplete. Do you want to proceed with implementation anyway? (yes/no)"
     * Wait for user response before continuing
     * If user says "no" or "wait" or "stop", halt execution
     * If user says "yes" or "proceed" or "continue", proceed to step 4

   - **If all checklists are complete**:
     * Display the table showing all checklists passed
     * Automatically proceed to step 4

4. Load and analyze the implementation context:
   - **REQUIRED**: Read tasks.md for the complete task list and execution plan
   - **REQUIRED**: Read plan.md for tech stack, architecture, and file structure
   - **IF EXISTS**: Read data-model.md for entities and relationships
   - **IF EXISTS**: Read contracts/ for API specifications and test requirements
   - **IF EXISTS**: Read research.md for technical decisions and constraints
   - **IF EXISTS**: Read quickstart.md for integration scenarios

5. **Project Setup Verification**:
   - **REQUIRED**: Create/verify ignore files based on actual project setup:

   **Detection & Creation Logic**:
   - Check if the following command succeeds to determine if the repository is a git repo (create/verify .gitignore if so):

     ```sh
     git rev-parse --git-dir 2>/dev/null
     ```
   - Check if Dockerfile* exists or Docker in plan.md → create/verify .dockerignore
   - Check if .eslintrc* exists → create/verify .eslintignore
   - Check if eslint.config.* exists → ensure the config's `ignores` entries cover required patterns
   - Check if .prettierrc* exists → create/verify .prettierignore
   - Check if .npmrc or package.json exists → create/verify .npmignore (if publishing)
   - Check if terraform files (*.tf) exist → create/verify .terraformignore
   - Check if .helmignore needed (helm charts present) → create/verify .helmignore

   **If ignore file already exists**: Verify it contains essential patterns, append missing critical patterns only
   **If ignore file missing**: Create with full pattern set for detected technology

   **Common Patterns by Technology** (from plan.md tech stack):
   - **Node.js/JavaScript**: `node_modules/`, `dist/`, `build/`, `*.log`, `.env*`
   - **Python**: `__pycache__/`, `*.pyc`, `.venv/`, `venv/`, `dist/`, `*.egg-info/`
   - **Java**: `target/`, `*.class`, `*.jar`, `.gradle/`, `build/`
   - **C#/.NET**: `bin/`, `obj/`, `*.user`, `*.suo`, `packages/`
   - **Go**: `*.exe`, `*.test`, `vendor/`, `*.out`
   - **Universal**: `.DS_Store`, `Thumbs.db`, `*.tmp`, `*.swp`, `.vscode/`, `.idea/`

   **Tool-Specific Patterns**:
   - **Docker**: `node_modules/`, `.git/`, `Dockerfile*`, `.dockerignore`, `*.log*`, `.env*`, `coverage/`
   - **ESLint**: `node_modules/`, `dist/`, `build/`, `coverage/`, `*.min.js`
   - **Prettier**: `node_modules/`, `dist/`, `build/`, `coverage/`, `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`
   - **Terraform**: `.terraform/`, `*.tfstate*`, `*.tfvars`, `.terraform.lock.hcl`

6. Read tasks.md structure and extract:
   - **Task phases**: Setup, Tests, Core, Integration, Polish
   - **Task dependencies**: Sequential vs parallel execution rules
   - **Task details**: ID, description, file paths, design + acceptance criteria
   - **Task comments**: Important notes and modifications to original plan
   - **Execution flow**: Order and dependency requirements

7. Execute implementation following the task plan:
   - **Phase-by-phase execution**: Complete each phase before moving to the next
   - **Respect dependencies**: Run sequential tasks in order, parallel tasks [P] can run together
   - **Follow TDD approach**: Execute test tasks before their corresponding implementation tasks
   - **File-based coordination**: Tasks affecting the same files must run sequentially
   - **Validation checkpoints**: Verify each phase completion before proceeding
   - **Read issue fields before implementation**:
     - Use `sl issue show <id>` to retrieve the issue's details
     - Read the `design` field for technical approach and file references
     - Read the `acceptance_criteria` field for requirements and success criteria
     - Use these fields to guide implementation decisions

8. Implementation execution rules:
   - **Setup first**: Initialize project structure, dependencies, configuration
   - **Tests before code**: If you need to write tests for contracts, entities, and integration scenarios
   - **Core development**: Implement models, services, CLI commands, endpoints
   - **Integration work**: Database connections, middleware, logging, external services
   - **Polish and validation**: Unit tests, performance optimization, documentation
   - **Verify against acceptance criteria**: Before marking task complete, verify implementation satisfies all acceptance_criteria
   - **Check off DoD items progressively**: As subtasks complete, use `sl issue update <id> --check-dod "Item text"` to mark relevant DoD items as verified
   - **Only close after all DoD items checked**: Ensure all Definition of Done items are marked complete before closing an issue

9. Progress tracking and error handling:
   - Find ready tasks using: `sl issue ready`
   - If no ready tasks, display blocking issues and offer options
   - Update issue status with: `sl issue update <id> --status in_progress`
   - Report progress after each completed task
   - Halt execution if any sequential task fails
   - For ready tasks, continue with successful tasks, report failed ones
   - Provide clear error messages with context for debugging
   - Suggest next steps if implementation cannot proceed
   - **IMPORTANT** For completed tasks, make sure to close the issue: `sl issue close <id> --reason "Completed"`

9a. **Definition of Done Verification** (before closing issues):

   Before closing any issue, verify its Definition of Done items:

   1. **Read the issue's DoD**: Use `sl issue show <id>` to get the definition_of_done field
   2. **Attempt automated verification** for each DoD item:

   **Automated Verification Patterns**:

   | Pattern | Verification Method | Example |
   |---------|---------------------|---------|
   | `file exists: <path>` | Check if file exists | `test -f src/auth/login.go` |
   | `directory exists: <path>` | Check if directory exists | `test -d src/models` |
   | `command succeeds: <cmd>` | Run command, check exit code | `go build ./...` |
   | `tests pass` | Run test suite | `go test ./...` |
   | `syntax valid: <file>` | Run linter | `golangci-lint run <file>` |
   | `no errors in <file>` | Parse for error patterns | `! grep -q "error" file.log` |

   3. **Interactive fallback**: For items that cannot be verified automatically:
      - Prompt user: "Is '<item>' complete? (y/n)"
      - Wait for response
      - Record the response

   4. **Handle verification failures**:
      - Display failed items with specific reasons
      - Example: "DoD item failed: 'file exists: src/missing.go' - File not found"
      - Require explicit `--force` confirmation to proceed
      - Log verification results for audit trail

   5. **Verification result display**:
      ```
      DoD Verification Results for SL-xxxxx:
      ✓ file exists: src/auth/login.go
      ✓ tests pass
      ✗ syntax valid: src/auth/login.go - 2 linting errors
      ? User confirmed: Is 'UI is intuitive' complete? → yes

      2/3 automated checks passed. 1 interactive confirmation.
      Proceed with closing? (--force required)
      ```

10. Completion validation:
   - Verify all required tasks are completed
   - Check that implemented features match the original specification
   - Validate that tests pass and coverage meets requirements
   - Confirm the implementation follows the technical plan
   - Report final status with summary of completed work

## Issue Tracking Commands

Use the built-in `sl issue` commands for issue management:

| Action | Command |
|--------|---------|
| Create issue | `sl issue create --title "..." --type task` |
| **Find ready tasks** | `sl issue ready` |
| List all open issues | `sl issue list --status open` |
| Show issue details | `sl issue show <id>` |
| Update issue | `sl issue update <id> --status in_progress` |
| Close issue | `sl issue close <id> --reason "Completed"` |
| Link dependencies | `sl issue link <from> blocks <to>` |
| List across all specs | `sl issue list --all` |

## Definition of Done

Before closing an issue, verify the Definition of Done criteria using the verification process in step 9a:

### Verification Process

1. **Check the issue's definition_of_done field**: `sl issue show <id>`
2. **Run automated verification** for patterns that can be checked programmatically
3. **Prompt for interactive confirmation** for subjective items
4. **Display verification results** with pass/fail status
5. **Require `--force`** if any automated checks fail

### Automated Verification Patterns

| DoD Item Pattern | Command |
|-----------------|---------|
| File/directory exists | `test -f <path>` or `test -d <path>` |
| Tests pass | `go test ./...` or equivalent |
| Build succeeds | `go build ./...` or equivalent |
| Linting passes | `golangci-lint run` or equivalent |

### Interactive Confirmation

For items that cannot be automated (e.g., "UI is intuitive", "Documentation is clear"):
- Prompt the user directly
- Record the response
- Include in verification results

### Force Close

If verification fails but the issue should still be closed:
- Use `sl issue close <id> --reason "Completed" --force`
- Document the reason for bypassing DoD

Note: This command assumes a complete task breakdown exists in tasks.md. If tasks are incomplete or missing, suggest running `/specledger.tasks` first to regenerate the task list.
