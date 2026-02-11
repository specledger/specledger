---
description: Execute the implementation plan by processing and executing all tasks defined in tasks.md (Beads)
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Purpose

Execute the implementation plan by processing all tasks in tasks.md (Beads). This command orchestrates the actual coding work following the defined task order and dependencies.

**When to use**: After `/specledger.tasks` generates the task list.

## Outline

1. **Sync issues from Supabase before starting** (required):

   ### Step 1.1: Check Authentication
   ```bash
   sl auth status
   ```
   If not logged in → **STOP** and prompt user to run `sl auth login` first.

   ### Step 1.2: Get Credentials

   **Option A: Use CLI commands (preferred):**
   ```bash
   ACCESS_TOKEN=$(sl auth token)
   ```

   **Option B: Fallback - extract from credentials file:**
   ```bash
   ACCESS_TOKEN=$(python3 -c "import json; print(json.load(open('$HOME/.specledger/credentials.json'))['access_token'])")
   ```

   **Beads Supabase config:**
   ```bash
   BEADS_SUPABASE_URL="https://lmjpnzplurfnojfqtqly.supabase.co"
   ```

   ### Step 1.3: Detect Repository Info
   ```bash
   remoteUrl=$(git remote get-url origin)
   # Parse owner/repo from URL (e.g., github.com/owner/repo.git)
   if [[ "$remoteUrl" =~ github\.com[:/]([^/]+)/([^/.]+) ]]; then
       REPO_OWNER="${BASH_REMATCH[1]}"
       REPO_NAME="${BASH_REMATCH[2]%.git}"
   fi
   ```

   ### Step 1.4: Fetch Project ID
   ```bash
   curl -s "${BEADS_SUPABASE_URL}/rest/v1/projects?select=id&repo_owner=eq.${REPO_OWNER}&repo_name=eq.${REPO_NAME}" \
     -H "apikey: ${ACCESS_TOKEN}" \
     -H "Authorization: Bearer ${ACCESS_TOKEN}"
   ```

   If project not found → **STOP** and show error: "Project not found: {REPO_OWNER}/{REPO_NAME}"

   ### Step 1.5: Fetch Issues, Dependencies, and Comments

   **Fetch Issues:**
   ```bash
   curl -s "${BEADS_SUPABASE_URL}/rest/v1/bd_issues?select=*&project_id=eq.${PROJECT_ID}&order=created_at.asc" \
     -H "apikey: ${ACCESS_TOKEN}" \
     -H "Authorization: Bearer ${ACCESS_TOKEN}"
   ```

   **Fetch Dependencies:**
   ```bash
   curl -s "${BEADS_SUPABASE_URL}/rest/v1/bd_dependencies?select=*&project_id=eq.${PROJECT_ID}" \
     -H "apikey: ${ACCESS_TOKEN}" \
     -H "Authorization: Bearer ${ACCESS_TOKEN}"
   ```

   **Fetch Comments:**
   ```bash
   curl -s "${BEADS_SUPABASE_URL}/rest/v1/bd_comments?select=*&project_id=eq.${PROJECT_ID}" \
     -H "apikey: ${ACCESS_TOKEN}" \
     -H "Authorization: Bearer ${ACCESS_TOKEN}"
   ```

   ### Step 1.6: Build and Write JSONL

   For each issue, build a JSON object:
   ```json
   {
     "id": "issue.id",
     "title": "issue.title",
     "status": "issue.status",
     "priority": "issue.priority",
     "issue_type": "issue.issue_type",
     "created_at": "issue.created_at",
     "updated_at": "issue.updated_at",
     "description": "issue.description (if exists)",
     "design": "issue.design (if exists)",
     "acceptance_criteria": "issue.acceptance_criteria (if exists)",
     "closed_at": "issue.closed_at (if exists)",
     "labels": ["issue.labels (if exists)"],
     "dependencies": [{"issue_id": "...", "depends_on_id": "...", "type": "..."}],
     "comments": [{"id": "...", "author": "...", "text": "...", "created_at": "..."}]
   }
   ```

   Write to `.beads/issues.jsonl` (one JSON object per line).

   ### Step 1.7: Display Sync Summary

   ```text
   🔄 Syncing beads issues from Supabase...

   ✓ Found project: {REPO_OWNER}/{REPO_NAME} ({PROJECT_ID})
   ✓ Fetched {n} issues
   ✓ Fetched {m} dependencies
   ✓ Fetched {k} comments
   ✓ Wrote {n} issues to .beads/issues.jsonl

   📊 Summary:
      - Issues: {n}
      - With dependencies: {count}
      - With comments: {count}

   ✅ Sync complete! Beads daemon will auto-import changes.
   ```

   If sync fails (network error, API error) → **STOP** and show error with details.

   This ensures you see latest issue status from other team members and prevents working on issues already claimed by others

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

6. Read tasks.md structure and use Beads to extract:
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

8. Implementation execution rules:
   - **Setup first**: Initialize project structure, dependencies, configuration
   - **Tests before code**: If you need to write tests for contracts, entities, and integration scenarios
   - **Core development**: Implement models, services, CLI commands, endpoints
   - **Integration work**: Database connections, middleware, logging, external services
   - **Polish and validation**: Unit tests, performance optimization, documentation

9. Progress tracking and error handling:
   - Find ready tasks using `bd ready --label "spec:......"`
   - Update Beads Issues progress with Comments and Progress
   - Report progress after each completed task
   - Halt execution if any sequential task fails
   - For ready tasks, continue with successful tasks, report failed ones
   - Provide clear error messages with context for debugging
   - Suggest next steps if implementation cannot proceed
   - **IMPORTANT** For completed tasks, make sure to add relevant comments and update the Beads task status to "Closed"

10. Completion validation:
   - Verify all required tasks are completed
   - Check that implemented features match the original specification
   - Validate that tests pass and coverage meets requirements
   - Confirm the implementation follows the technical plan
   - Report final status with summary of completed work

Note: This command assumes a complete task breakdown exists in tasks.md. If tasks are incomplete or missing, suggest running `/specledger.tasks` first to regenerate the task list.

---

## Supabase Sync Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| JWT expired / PGRST303 | Access token expired | Run `sl auth login` to refresh token |
| 401 Unauthorized | Session expired | Run `sl auth login` again |
| 403 Forbidden | No permission | Check access rights for the project |
| Project not found | Repo not registered | Ensure project is registered in SpecLedger |
| Credentials file not found | Not logged in | Run `sl auth login` first |
| Network error | Connection issue | Check internet connection and retry |

**Note:** If CLI commands fail in TUI mode, use the fallback method to extract credentials directly from the JSON file.
