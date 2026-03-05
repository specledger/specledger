# Feature Specification: AI Agent Task Execution Service

**Feature Branch**: `599-agent-task-execution`
**Created**: 2026-03-01
**Status**: Draft
**Input**: User description: "Create a service to execute spec tasks for SpecLedger users via AI agents (Goose), authenticated with Supabase tokens, inspired by Stripe's minions architecture"

## Clarifications

### Session 2026-03-01

- Q: How should the agent organize its git commits? → A: Hierarchical branching — each main task gets its own branch off the spec's feature branch. Sub-tasks branch off the main task's branch and auto-merge back when complete. Only the main task branch requires human review before merging into the spec branch.
- Q: How should the system verify a task is successfully completed? → A: Exit code (0) + Definition of Done checklist verification for individual tasks. Human review is required only for main tasks (not sub-tasks). Sub-tasks auto-merge into the main task branch upon completion. Some main tasks with manual verification steps trigger a confirmation message via bot to a human reviewer. Test infrastructure is not yet available — DoD abstracts test requirements and will integrate when available.
- Q: What scope should `sl agent run` target by default? → A: Auto-detect spec context from the current git branch name, consistent with other `sl` commands. Allow `--spec` flag to override.

### Session 2026-03-05

- Q: What is the value of local agent task pickup vs `/specledger.implement`? → A: Local execution is a lower-priority convenience for individual developers who want to test agent execution on their own machine before committing to cloud-scheduled runs. The primary execution path is cloud-triggered scheduling.
- Q: Should monitoring use CLI commands or observability platforms? → A: Emit execution metrics to configurable logging/observability platforms (Splunk, Sentry) rather than relying on CLI status commands.
- Q: How should credentials be configured for agent execution? → A: Configuration must support (1) git credentials for task execution, and (2) specledger access token scoping — whether tasks are owner-scoped or use a shared master credential for the minion pool.
- Q: Should tasks have timeout limits? → A: Yes, configurable per-task and global default timeout limits.
- Q: How should circular dependencies be resolved? → A: Rather than skipping, use a mock/stub approach — T1 runs with a stub of T2's output, T2 runs with a stub of T1's output, then conflicts are resolved and merged into one commit.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cloud-Triggered Agent Execution with Sub-Minions (Priority: P1)

When a spec plan is approved on the SpecLedger UI, the system automatically schedules task execution. The main task is broken down into small, manageable sub-tasks, each assigned to a sub-minion agent for parallel execution. Each sub-task is checked out into its own branch off the main task's branch. Sub-minions execute their sub-tasks autonomously — writing code, running tests, and committing results. When a sub-task completes, it is automatically merged into the main task's branch without human review. When all sub-tasks are complete, the main task is marked as `needs_review` for human review before merging into the spec's feature branch.

**Why this priority**: This is the core value proposition — automated, cloud-triggered execution that scales via parallel sub-minions. Turning approved spec plans into working code without manual intervention is the primary workflow.

**Independent Test**: Can be tested by approving a spec plan on the UI, verifying the system schedules execution, breaks the main task into sub-tasks, dispatches sub-minions, and auto-merges completed sub-task branches into the main task branch.

**Acceptance Scenarios**:

1. **Given** a spec plan is approved on the SpecLedger UI, **When** the system receives the approval event, **Then** it automatically schedules agent execution for all tasks in the spec, breaking each main task into sub-tasks and dispatching sub-minion agents.
2. **Given** a main task with 3 sub-tasks (S1, S2, S3), **When** sub-minions are dispatched, **Then** each sub-task gets its own branch off the main task's branch (e.g., `599-agent-task-execution/T1/S1`), and sub-minions execute in parallel.
3. **Given** sub-task S1 completes successfully, **When** the sub-minion finishes, **Then** S1's branch is automatically merged into the main task's branch without human review, and the sub-task status is set to `closed`.
4. **Given** all sub-tasks for a main task are complete, **When** the last sub-task merges, **Then** the main task is marked as `needs_review` and a notification is sent to the designated reviewer for human review.
5. **Given** a sub-task exceeds its configured timeout limit, **When** the timeout is reached, **Then** the sub-task is paused with a timeout error, and the system continues with remaining sub-tasks.

---

### User Story 2 - Execution Observability and Metrics (Priority: P2)

A developer or team lead wants to monitor agent execution progress and health. The system emits execution metrics and logs to configurable observability platforms (Splunk, Sentry, or other logging services) rather than requiring CLI polling. Metrics include task status transitions, execution durations, success/failure rates, and error details.

**Why this priority**: Visibility into agent work builds trust and enables teams to intervene when needed. Observability platform integration provides persistent, searchable, and alertable monitoring compared to ephemeral CLI output.

**Independent Test**: Can be tested by configuring an observability platform endpoint, running agent execution, and verifying metrics are emitted with correct task status, timing, and error data.

**Acceptance Scenarios**:

1. **Given** an observability platform is configured (e.g., Splunk endpoint via `sl config set agent.observability.endpoint`), **When** a task transitions status (open → in_progress → closed), **Then** a structured metric event is emitted to the configured platform with task ID, status, timestamp, and duration.
2. **Given** a task fails during execution, **When** the failure is recorded, **Then** an error event is emitted to the observability platform with error details, stack trace context, and the task's execution log path.
3. **Given** an agent run completes (all tasks processed), **When** the run finishes, **Then** a summary metric is emitted with total tasks, completed count, failed count, and total execution time.

---

### User Story 3 - Local Agent Task Pickup (Priority: P3)

A developer wants to test agent execution locally on their own machine before committing to cloud-scheduled runs. They run `sl agent run` which reads the task list, selects the next available unblocked task, constructs an execution context, and launches a Goose agent session to execute it. This serves as a development and testing convenience for individual developers.

**Why this priority**: Local execution is a convenience for individual developers to validate agent behavior on their codebase before relying on cloud-triggered scheduling. It is not the primary execution path.

**Independent Test**: Can be fully tested by creating a spec with 2-3 tasks, running `sl agent run`, and verifying the agent picks up tasks in dependency order, executes them, and updates their status.

**Acceptance Scenarios**:

1. **Given** a spec with 3 tasks (T1 blocks T2, T3 is independent) all in `open` status, **When** the user runs `sl agent run`, **Then** the system selects T1 and T3 (unblocked tasks), launches an agent for the first available, and sets its status to `in_progress`.
2. **Given** a running agent completes a task successfully (tests pass, code compiles), **When** the agent finishes, **Then** the task status is updated to `closed`, completion timestamp is recorded, and the system picks up the next unblocked task.
3. **Given** a running agent encounters a failure it cannot resolve, **When** the agent exhausts its retry budget, **Then** the task remains `in_progress` with failure notes appended, and the system moves on to the next available task rather than blocking.
4. **Given** no Goose installation is detected on the machine, **When** the user runs `sl agent run`, **Then** the system displays a clear error with installation instructions.

---

### User Story 4 - Execution Configuration, Credentials, and Recipes (Priority: P4)

A developer or team lead wants to customize how the agent executes tasks — specifying which model provider to use, setting turn limits, configuring task timeout limits, providing git credentials for execution, configuring specledger access tokens (owner-scoped or shared master credential), and defining project-specific execution recipes. They use the existing `sl config` system (from 597-agent-model-config) to set agent execution preferences and optionally create reusable recipe files for their project.

**Why this priority**: Customization enables teams to optimize agent behavior for their specific codebase and workflows. Nice-to-have after core execution works.

**Independent Test**: Can be tested by setting custom agent config values and verifying the agent uses them during task execution.

**Acceptance Scenarios**:

1. **Given** the user has configured `agent.model` to a specific model via `sl config set`, **When** `sl agent run` is executed, **Then** the Goose session uses the configured model.
2. **Given** a recipe file exists at `specledger/<spec>/agent-recipe.yaml`, **When** `sl agent run` is executed for that spec, **Then** the recipe instructions are included in the agent's execution context.
3. **Given** the user has configured `agent.git-credentials` with a token, **When** the agent executes tasks, **Then** it uses the configured git credentials for all git operations (clone, push, merge).
4. **Given** the user has configured `agent.access-token` with a specledger token, **When** the agent picks up tasks, **Then** it only picks up tasks belonging to the token's owner (or all tasks if using a master credential).
5. **Given** a task has a configured timeout of 30 minutes via `agent.task-timeout`, **When** the task execution exceeds 30 minutes, **Then** the system terminates the agent session and marks the task with a timeout error.

---

### Edge Cases

- What happens when a task has circular dependencies (T1 blocks T2, T2 blocks T1)? The system resolves circular dependencies using a mock/stub approach: T1 runs with a stub of T2's expected output, T2 runs with a stub of T1's expected output. Both sessions run in parallel on separate branches, then conflicts are resolved and merged into a single commit.
- What happens when the agent modifies files that conflict with another concurrent agent's changes? Each sub-minion works on an isolated branch off the main task's branch. Merge conflicts are detected during auto-merge and flagged for resolution.
- What happens when the Goose process crashes mid-task? The task remains in `in_progress` status. On the next run, the system detects stale in-progress tasks (exceeding the configured timeout or heartbeat threshold) and offers to retry or skip them.
- What happens when the user's LLM provider rate-limits the agent? The system respects Goose's built-in retry and backoff behavior. If the provider remains unavailable, the task is paused with an appropriate error message.
- What happens when a task's definition of done includes manual verification steps? For main tasks, the agent marks the task as `needs_review` instead of `closed` and sends a confirmation message via bot to the designated human reviewer. Sub-tasks do not require human review and auto-merge upon completion.
- What happens when all sub-tasks in a main task are completed? The sub-task branches are all merged into the main task branch. The main task is marked as `needs_review` for human review. No auto-merge to the spec branch occurs.
- What happens when all main tasks in a run are completed? The system produces a summary and emits a completion metric to the configured observability platform. The spec branch remains ready for human-reviewed merge to main.
- What happens when a task exceeds its configured timeout limit? The system terminates the agent session gracefully, marks the task with a timeout error, and moves on to the next available task.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST authenticate users via existing Supabase credentials stored by `sl auth login` before allowing task execution.
- **FR-002**: System MUST read the task list from `specledger/<spec>/issues.jsonl` and identify all tasks eligible for execution (status `open`, not blocked by incomplete tasks).
- **FR-003**: System MUST resolve task dependencies using `blocked_by` and `blocks` fields, executing tasks only when all their blockers are completed.
- **FR-004**: System MUST construct an execution context for each task containing: task title, description, acceptance criteria, definition of done, design notes, and relevant repository file paths.
- **FR-005**: System MUST launch a Goose agent session with the constructed execution context, passing it as an instruction to `goose run`.
- **FR-006**: System MUST update task status in the JSONL store as execution progresses: `open` → `in_progress` → `closed` (or `needs_review` for main tasks requiring manual verification, or back to `open` on failure). Completion requires Goose exit code 0 AND all Definition of Done checklist items addressed.
- **FR-007**: System MUST emit structured execution metrics and logs to a configurable observability platform (Splunk, Sentry, or other logging services) for each task status transition, failure, and run completion.
- **FR-008**: System MUST verify Goose is installed and accessible before attempting task execution, providing clear installation guidance if missing.
- **FR-009**: System MUST support headless execution mode (no interactive prompts) for cloud/CI environments via `--headless` flag.
- **FR-010**: System MUST use the agent configuration from the existing config hierarchy (597-agent-model-config) to determine model provider, model name, and other Goose settings.
- **FR-011**: System MUST support executing a single specific task via `sl agent run --task <task-id>` in addition to automatic task pickup.
- **FR-012**: System MUST prevent concurrent execution of the same task by using a locking mechanism (file lock or status-based claim).
- **FR-013**: System MUST support stopping a running agent gracefully via `sl agent stop`, allowing the current Goose session to complete its turn before terminating.
- **FR-014**: System MUST auto-detect the target spec context from the current git branch name (e.g., `599-agent-task-execution` → `specledger/599-agent-task-execution/`), with a `--spec` flag to override.
- **FR-015**: System MUST implement hierarchical branching: main tasks branch off the spec's feature branch, sub-tasks branch off the main task's branch. Sub-task branches auto-merge into the main task branch upon completion.
- **FR-016**: System MUST send a notification via bot to a designated reviewer when a main task is marked `needs_review` due to manual verification steps in its Definition of Done. Sub-tasks do not require human review.
- **FR-017**: System MUST NOT auto-merge main task branches to the spec branch. Main task merge requires human review.
- **FR-018**: System MUST automatically schedule task execution when a spec plan is approved on the SpecLedger UI, breaking main tasks into sub-tasks and dispatching sub-minion agents.
- **FR-019**: System MUST support configurable task timeout limits (per-task via task metadata and global default via `agent.task-timeout` config). Tasks exceeding the timeout are terminated gracefully and marked with a timeout error.
- **FR-020**: System MUST resolve circular dependencies using a mock/stub approach — each task in the cycle runs with a stub of the other's expected output on separate branches, then results are merged and conflicts resolved.
- **FR-021**: System MUST support configuring git credentials (`agent.git-credentials`) for agent git operations in cloud environments.
- **FR-022**: System MUST support configuring specledger access token scoping (`agent.access-token`) — either owner-scoped (agent picks up tasks belonging to the token's owner) or master credential (agent picks up all tasks).

### Key Entities

- **Agent Run**: A single invocation of agent execution (cloud-triggered or local) that may execute one or more main tasks. Attributes: run ID, spec context, start time, end time, status (running/completed/failed), task execution results.
- **Main Task**: A top-level task from the spec's task list. Attributes: task ID, branch name, sub-task list, status, reviewer assignment. Requires human review before merge.
- **Sub-Task**: A smaller, manageable unit broken down from a main task. Attributes: sub-task ID, parent task ID, branch name, status, timeout. Auto-merges into main task branch upon completion without human review.
- **Task Execution**: The execution of a single task (main or sub) by a Goose agent (sub-minion). Attributes: task ID, agent session ID, start time, end time, status, output log path, git branch, timeout limit.
- **Execution Context**: The compiled information package sent to Goose for a task. Includes task metadata, codebase context, and execution instructions.
- **Agent Recipe**: An optional YAML configuration file that customizes agent behavior for a specific spec or project. Includes Goose recipe format fields (instructions, extensions, settings).
- **Sub-Minion**: An individual Goose agent instance dispatched to execute a specific sub-task. Multiple sub-minions can run in parallel for different sub-tasks of the same main task.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Approved spec plans automatically trigger agent execution in the cloud, completing the scheduling and dispatch cycle within 2 minutes of approval.
- **SC-002**: 80% of well-defined sub-tasks (clear acceptance criteria, design notes, bounded scope) are completed successfully by sub-minion agents on the first attempt.
- **SC-003**: Failed tasks provide sufficient diagnostic information (logs, error context, observability metrics) for a developer to understand and manually resolve the issue within 10 minutes.
- **SC-004**: Task execution respects dependency ordering — no task begins before all its blockers are resolved, verified across 100% of test runs.
- **SC-005**: Headless cloud execution works without any user interaction once configured, supporting unattended runs of 10+ tasks with parallel sub-minion dispatch.
- **SC-006**: Agent execution time overhead (setup, context building, status updates) adds less than 30 seconds per task beyond the actual Goose processing time.
- **SC-007**: Sub-task branches auto-merge into main task branches with zero human intervention, achieving 95%+ clean merge rate for well-scoped sub-tasks.
- **SC-008**: Execution metrics are emitted to the configured observability platform within 5 seconds of each task status transition.

### Previous work

### Epic: SL-aaf16e - Advanced Agent Model Configuration (597-agent-model-config)

- **Agent Config Schema**: Established the configuration structure for agent model, provider, base URL, auth tokens, and permission modes across global/team/personal scopes.
- **Config Merge Logic**: Implemented multi-layer config resolution (default → global → profile → team-local → personal-local) used to determine agent execution settings.
- **Agent Launcher BuildEnv**: Created the mechanism to inject resolved agent configuration as environment variables into agent subprocesses.

### Related: 597-issue-create-fields

- **Issue JSONL Store**: File-based issue storage with JSONL format, file locking, and full CRUD operations — the foundation for task state management during agent execution.
- **Issue Dependencies**: `blocked_by` and `blocks` fields with cycle detection — critical for dependency-ordered task execution.
- **Definition of Done**: Checklist-based completion criteria per task — used by agents to verify task completion.

## Assumptions

- Goose is installed and available in the execution environment (locally by the user, or pre-installed in cloud containers).
- The user's LLM provider API key is configured either in Goose's own config or via SpecLedger's agent config (`sl config set agent.api-key`).
- Tasks in `issues.jsonl` are well-formed with sufficient description, acceptance criteria, and design notes for an AI agent to execute meaningfully.
- The repository is in a clean git state (no uncommitted changes) before agent execution begins.
- Cloud-triggered execution is the primary workflow. Local execution is a convenience for development and testing.
- Goose's `goose run` CLI command is the primary integration point — SpecLedger does not embed or fork the Goose runtime.
- Hierarchical branching: main tasks branch from the spec's feature branch, sub-tasks branch from the main task's branch. Sub-tasks auto-merge; main tasks require human review.
- Final merge of the spec branch to main always requires human review — the agent never auto-merges to main.
- Test infrastructure is not yet available. Definition of Done checklist abstracts test requirements; test runner integration will be added when infrastructure exists.
- An observability platform (Splunk, Sentry, or equivalent) is available and configured for execution metric emission in production use.
