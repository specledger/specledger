# Tasks Index: AI Agent Task Execution Service

Beads Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through Beads CLI.

## Feature Tracking

* **Beads Epic ID**: `sl-74p`
* **User Stories Source**: `specledger/599-agent-task-execution/spec.md`
* **Research Inputs**: `specledger/599-agent-task-execution/research.md`
* **Planning Details**: `specledger/599-agent-task-execution/plan.md`
* **Data Model**: `specledger/599-agent-task-execution/data-model.md`
* **Contract Definitions**: `specledger/599-agent-task-execution/contracts/`

## Beads Query Hints

Use the `bd` CLI to query and manipulate the issue graph:

```bash
# Find all open tasks for this feature
bd list --label spec:599-agent-task-execution --status open --limit 10

# Find ready tasks to implement
bd ready --label spec:599-agent-task-execution --limit 5

# See full dependency tree
bd dep tree --reverse sl-74p

# View issues by story
bd list --label story:US1 --label spec:599-agent-task-execution
bd list --label story:US2 --label spec:599-agent-task-execution

# View issues by phase
bd list --type feature --label spec:599-agent-task-execution

# View issues by component
bd list --label component:agent --label spec:599-agent-task-execution
bd list --label component:cli --label spec:599-agent-task-execution

# Define dependencies
bd dep add <dependent-id> <blocker-id> --type blocks

# Show dependencies for an issue
bd dep tree sl-w47
```

## Tasks and Phases Structure

This feature follows Beads' 2-level graph structure:

* **Epic**: sl-74p — AI Agent Task Execution Service
* **Phases**: Beads issues of type `feature`, children of the epic
  * **sl-r9u** — Setup Phase (project initialization)
  * **sl-8hk** — Foundational Phase (blocking prerequisites)
  * **sl-p64** — User Story 1: Local Agent Task Pickup (P1, MVP)
  * **sl-1lz** — User Story 2: Task Execution Monitoring and Status (P2)
  * **sl-8u1** — User Story 3: Cloud Agent Execution (P3)
  * **sl-9ou** — User Story 4: Execution Configuration and Recipes (P4)
  * **sl-sl3** — Polish and Cross-Cutting Concerns
* **Tasks**: Issues of type `task`, children of each feature (phase)

## Phase Dependencies

```
Setup (sl-r9u) ──blocks──▶ Foundational (sl-8hk) ──blocks──▶ US1 (sl-p64) ──blocks──▶ US2 (sl-1lz)
                                                        │                        ├──▶ US3 (sl-8u1)
                                                        │                        ├──▶ US4 (sl-9ou)
                                                        │                        └──▶ Polish (sl-sl3)
                                                        ├──blocks──▶ US2 (sl-1lz)
                                                        ├──blocks──▶ US3 (sl-8u1)
                                                        └──blocks──▶ US4 (sl-9ou)
```

## Convention Summary

| Type    | Description                  | Labels                                         |
| ------- | ---------------------------- | ---------------------------------------------- |
| epic    | Full feature epic            | `spec:599-agent-task-execution`                |
| feature | Implementation phase / story | `phase:<name>`, `story:<US#>`                  |
| task    | Implementation task          | `component:<x>`, `requirement:<fr-id>`         |

## Phase Breakdown

### Phase 1: Setup (sl-r9u) — 4 tasks

| Task ID | Title | Component | Priority |
|---------|-------|-----------|----------|
| sl-ext  | Create pkg/cli/agent/ package structure | agent | P1 |
| sl-0vr  | Register sl agent command group in Cobra | cli | P1 |
| sl-395  | Add Goose to DefaultAgents in launcher | launcher | P1 |
| sl-v75  | Add .agent-runs/ to project gitignore | infra | P1 |

All 4 tasks can run **in parallel** (no internal dependencies).

### Phase 2: Foundational (sl-8hk) — 7 tasks

| Task ID | Title | Component | Blocked By |
|---------|-------|-----------|------------|
| sl-zv3  | Implement AgentRun and TaskResult types | agent | — |
| sl-f26  | Implement AgentRun JSON persistence store | agent | sl-zv3 |
| sl-xxt  | Implement task selector with priority ordering | agent | — |
| sl-9mm  | Implement execution context builder | agent | — |
| sl-cq9  | Implement Goose subprocess adapter | agent | — |
| sl-hky  | Implement Goose environment variable mapping | agent | sl-cq9 |
| sl-3z8  | Implement spec context auto-detection | agent | — |

Parallel groups: {sl-zv3, sl-xxt, sl-9mm, sl-cq9, sl-3z8} then {sl-f26, sl-hky}.

### Phase 3: User Story 1 — Local Agent Task Pickup (sl-p64) — 6 tasks (MVP)

| Task ID | Title | Component | Blocked By |
|---------|-------|-----------|------------|
| sl-w47  | Implement core runner orchestrator | agent | — |
| sl-8pa  | Implement sl agent run Cobra command | cli | sl-w47 |
| sl-r6m  | Implement task status updates during execution | agent | sl-w47 |
| sl-0ky  | Implement agent output capture and log storage | agent | sl-w47 |
| sl-blo  | Implement git state verification and commit tracking | agent | sl-w47 |
| sl-se6  | Implement cycle detection and stale task handling | agent | sl-w47 |

sl-w47 first, then {sl-8pa, sl-r6m, sl-0ky, sl-blo, sl-se6} in parallel.

### Phase 4: User Story 2 — Monitoring and Status (sl-1lz) — 2 tasks

| Task ID | Title | Component | Blocked By |
|---------|-------|-----------|------------|
| sl-7h1  | Implement sl agent status command | cli | — |
| sl-shr  | Implement sl agent logs command | cli | — |

Both can run **in parallel**.

### Phase 5: User Story 3 — Cloud Agent Execution (sl-8u1) — 4 tasks

| Task ID | Title | Component | Blocked By |
|---------|-------|-----------|------------|
| sl-iq2  | Implement headless execution mode | agent | — |
| sl-wsy  | Implement task claim locking | agent | — |
| sl-yw3  | Implement sl agent stop command | cli | sl-wsy |
| sl-9um  | Implement auth token auto-refresh | agent | — |

Parallel: {sl-iq2, sl-wsy, sl-9um}, then sl-yw3.

### Phase 6: User Story 4 — Configuration and Recipes (sl-9ou) — 2 tasks

| Task ID | Title | Component | Blocked By |
|---------|-------|-----------|------------|
| sl-etv  | Implement agent recipe YAML loading | agent | — |
| sl-7vq  | Implement config-driven model/provider selection | agent | — |

Both can run **in parallel**.

### Phase 7: Polish (sl-sl3) — 3 tasks

| Task ID | Title | Component | Blocked By |
|---------|-------|-----------|------------|
| sl-6ym  | Polish error messages and user guidance | cli | — |
| sl-3pp  | Implement run completion summary output | cli | — |
| sl-qtz  | Validate quickstart.md workflow end-to-end | docs | — |

All can run **in parallel**.

## Implementation Strategy

### MVP Scope (User Story 1 only)

The minimum viable product is **Phases 1-3** (Setup + Foundational + US1):
- 17 tasks total
- Delivers: `sl agent run` with task pickup, execution, status updates
- Independently testable: create spec with tasks, run `sl agent run`, verify execution

### Incremental Delivery

1. **MVP**: Setup → Foundational → US1 (sl agent run works end-to-end)
2. **Visibility**: US2 (sl agent status + sl agent logs)
3. **Cloud**: US3 (headless mode + locking + stop)
4. **Customization**: US4 (recipes + config-driven model selection)
5. **Quality**: Polish phase

### Story Testability

| Story | Independent Test | Requires |
|-------|-----------------|----------|
| US1   | Create spec with 2-3 tasks, run `sl agent run`, verify execution | Goose installed, API key configured |
| US2   | After US1 run, use `sl agent status` and `sl agent logs` | Completed US1 run |
| US3   | Run `sl agent run --headless` in container with env vars | Docker, env vars |
| US4   | Set custom config, create recipe, verify agent uses them | US1 working |

## Agent Execution Flow

MCP agents and AI workflows should:

1. **Assume `bd init` already done** by `specledger.tasks`
2. **Use `bd create`** to directly generate Beads issues
3. **Set metadata and dependencies** in the graph, not markdown
4. **Use this markdown only as a navigational anchor**

> Agents MUST NOT output tasks into this file. They MUST use Beads CLI to record all task and phase structure.

## Example Queries for Agents

```bash
# Get all tasks in tree structure for the feature
bd dep tree --reverse sl-74p

# Get all tasks by label
bd list --label spec:599-agent-task-execution --label story:US1

# Add a new task
bd create "Fix edge case in runner" -t task --parent sl-p64 --label spec:599-agent-task-execution --label component:agent

# Update notes on a task
bd update sl-w47 --notes "Consider using errgroup for parallel task future"

# Add a comment based on findings
bd comments add sl-w47 "Goose exit code behavior verified in PR #4621"

# Mark task as completed
bd close sl-ext --reason "Package structure created, all files compile"

# Check overall progress
bd count --label spec:599-agent-task-execution --status open
bd count --label spec:599-agent-task-execution --status closed
```

## Status Tracking

Status is tracked only in Beads:

* **Open** — default
* **In Progress** — task being worked on
* **Blocked** — dependency unresolved
* **Closed** — complete

Use `bd ready`, `bd list --status`, `bd count` with appropriate filters to query progress.

---

> This file is intentionally light and index-only. Implementation data lives in Beads. Update this file only to point humans and agents to canonical query paths and feature references.
