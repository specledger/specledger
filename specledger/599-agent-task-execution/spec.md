# Feature Specification: AI Agent Task Execution Service

**Feature Branch**: `599-agent-task-execution`
**Created**: 2026-03-01
**Status**: Draft
**Input**: User description: "Create a service to execute spec tasks for SpecLedger users via AI agents (Goose), authenticated with Supabase tokens, inspired by Stripe's minions architecture"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Local Agent Task Pickup (Priority: P1)

A developer has completed their specification and all tasks are confirmed in `sl issues`. They want to dispatch those tasks to an AI agent running locally on their machine. The developer runs a single command (e.g., `sl agent run`) which reads the task list, selects the next available unblocked task, constructs an execution context (task description, acceptance criteria, design notes, repository context), and launches a Goose agent session to execute it. The agent works autonomously — writing code, running tests, and committing results — then marks the task as complete and picks up the next one.

**Why this priority**: This is the core value proposition — turning confirmed specs into working code without manual developer intervention. Without this, nothing else matters.

**Independent Test**: Can be fully tested by creating a spec with 2-3 tasks, running `sl agent run`, and verifying the agent picks up tasks in dependency order, executes them, and updates their status.

**Acceptance Scenarios**:

1. **Given** a spec with 3 tasks (T1 blocks T2, T3 is independent) all in `open` status, **When** the user runs `sl agent run`, **Then** the system selects T1 and T3 (unblocked tasks), launches an agent for the first available, and sets its status to `in_progress`.
2. **Given** a running agent completes a task successfully (tests pass, code compiles), **When** the agent finishes, **Then** the task status is updated to `closed`, completion timestamp is recorded, and the system picks up the next unblocked task.
3. **Given** a running agent encounters a failure it cannot resolve, **When** the agent exhausts its retry budget, **Then** the task remains `in_progress` with failure notes appended, and the system moves on to the next available task rather than blocking.
4. **Given** no Goose installation is detected on the machine, **When** the user runs `sl agent run`, **Then** the system displays a clear error with installation instructions.

---

### User Story 2 - Task Execution Monitoring and Status (Priority: P2)

A developer has dispatched tasks to the agent and wants to monitor progress. They can see which tasks are running, which completed, which failed, and review agent output logs. The developer runs `sl agent status` to get a dashboard view of the current execution state.

**Why this priority**: Visibility into agent work builds trust and enables developers to intervene when needed. Critical for adoption but not required for basic execution.

**Independent Test**: Can be tested by launching an agent run, then using `sl agent status` in another terminal to verify real-time progress visibility.

**Acceptance Scenarios**:

1. **Given** an agent is currently executing task SL-abc123, **When** the user runs `sl agent status`, **Then** they see the task ID, title, elapsed time, and current agent activity summary.
2. **Given** 2 tasks have completed and 1 failed, **When** the user runs `sl agent status`, **Then** they see a summary showing completed count, failed count, remaining count, and can drill into failure details.
3. **Given** no agent is currently running, **When** the user runs `sl agent status`, **Then** they see the results of the last run with a timestamp.

---

### User Story 3 - Cloud Agent Execution (Priority: P3)

A developer or team lead wants to run agent task execution in a cloud environment (CI/CD pipeline, cloud VM, or container) rather than on their local machine. They configure the agent service with environment variables for authentication and Goose provider settings, then trigger execution remotely. The system uses the same task pickup and execution logic but runs headlessly without any interactive prompts.

**Why this priority**: Cloud execution enables scaling to multiple parallel agents and unattended overnight runs. Important for teams but not required for individual MVP use.

**Independent Test**: Can be tested by configuring environment variables (Supabase token, Goose provider credentials) and running `sl agent run --headless` in a containerized environment, verifying tasks are picked up and executed without user interaction.

**Acceptance Scenarios**:

1. **Given** valid environment variables for Supabase auth and Goose provider, **When** `sl agent run --headless` is executed in a container, **Then** the agent authenticates, fetches tasks, and executes them without any interactive prompts.
2. **Given** the agent is running headlessly and the auth token expires mid-execution, **When** a refresh token is available, **Then** the system automatically refreshes the token and continues execution.
3. **Given** multiple agent instances are launched in parallel (one per task), **When** they access the same task list, **Then** each agent claims a unique task and no task is executed by more than one agent.

---

### User Story 4 - Execution Configuration and Recipes (Priority: P4)

A developer wants to customize how the agent executes tasks — specifying which model provider to use, setting turn limits, providing additional context files, or defining project-specific execution recipes. They use the existing `sl config` system (from 597-agent-model-config) to set agent execution preferences and optionally create reusable recipe files for their project.

**Why this priority**: Customization enables teams to optimize agent behavior for their specific codebase and workflows. Nice-to-have after core execution works.

**Independent Test**: Can be tested by setting custom agent config values and verifying the agent uses them during task execution.

**Acceptance Scenarios**:

1. **Given** the user has configured `agent.model` to a specific model via `sl config set`, **When** `sl agent run` is executed, **Then** the Goose session uses the configured model.
2. **Given** a recipe file exists at `specledger/<spec>/agent-recipe.yaml`, **When** `sl agent run` is executed for that spec, **Then** the recipe instructions are included in the agent's execution context.

---

### Edge Cases

- What happens when a task has circular dependencies (T1 blocks T2, T2 blocks T1)? System detects the cycle and reports it as an error, skipping both tasks.
- What happens when the agent modifies files that conflict with another concurrent agent's changes? In MVP (single local agent), this is avoided by sequential execution. For cloud parallel mode, each agent works on an isolated git branch/worktree.
- What happens when the Goose process crashes mid-task? The task remains in `in_progress` status. On the next `sl agent run`, the system detects stale in-progress tasks and offers to retry or skip them.
- What happens when the user's LLM provider rate-limits the agent? The system respects Goose's built-in retry and backoff behavior. If the provider remains unavailable, the task is paused with an appropriate error message.
- What happens when a task's definition of done includes manual verification steps? The agent marks the task as `needs_review` instead of `closed`, flagging it for human verification.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST authenticate users via existing Supabase credentials stored by `sl auth login` before allowing task execution.
- **FR-002**: System MUST read the task list from `specledger/<spec>/issues.jsonl` and identify all tasks eligible for execution (status `open`, not blocked by incomplete tasks).
- **FR-003**: System MUST resolve task dependencies using `blocked_by` and `blocks` fields, executing tasks only when all their blockers are completed.
- **FR-004**: System MUST construct an execution context for each task containing: task title, description, acceptance criteria, definition of done, design notes, and relevant repository file paths.
- **FR-005**: System MUST launch a Goose agent session with the constructed execution context, passing it as an instruction to `goose run`.
- **FR-006**: System MUST update task status in the JSONL store as execution progresses: `open` → `in_progress` → `closed` (or back to `open` on failure).
- **FR-007**: System MUST capture and store agent execution logs for each task, accessible via `sl agent status` or `sl agent logs <task-id>`.
- **FR-008**: System MUST verify Goose is installed and accessible before attempting task execution, providing clear installation guidance if missing.
- **FR-009**: System MUST support headless execution mode (no interactive prompts) for cloud/CI environments via `--headless` flag.
- **FR-010**: System MUST use the agent configuration from the existing config hierarchy (597-agent-model-config) to determine model provider, model name, and other Goose settings.
- **FR-011**: System MUST support executing a single specific task via `sl agent run --task <task-id>` in addition to automatic task pickup.
- **FR-012**: System MUST prevent concurrent execution of the same task by using a locking mechanism (file lock or status-based claim).
- **FR-013**: System MUST support stopping a running agent gracefully via `sl agent stop`, allowing the current Goose session to complete its turn before terminating.

### Key Entities

- **Agent Run**: A single invocation of `sl agent run` that may execute one or more tasks. Attributes: run ID, spec context, start time, end time, status (running/completed/failed), task execution results.
- **Task Execution**: The execution of a single task by a Goose agent. Attributes: task ID, agent session ID, start time, end time, status, output log path, git branch/commit (if changes were made).
- **Execution Context**: The compiled information package sent to Goose for a task. Includes task metadata, codebase context, and execution instructions.
- **Agent Recipe**: An optional YAML configuration file that customizes agent behavior for a specific spec or project. Includes Goose recipe format fields (instructions, extensions, settings).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can go from confirmed spec with tasks to working code with a single command, completing the full cycle in under 5 minutes per task (excluding agent execution time).
- **SC-002**: 80% of well-defined tasks (clear acceptance criteria, design notes, bounded scope) are completed successfully by the agent on the first attempt.
- **SC-003**: Failed tasks provide sufficient diagnostic information (logs, error context) for a developer to understand and manually resolve the issue within 10 minutes.
- **SC-004**: Task execution respects dependency ordering — no task begins before all its blockers are resolved, verified across 100% of test runs.
- **SC-005**: Headless cloud execution works without any user interaction once configured, supporting unattended runs of 10+ tasks.
- **SC-006**: Agent execution time overhead (setup, context building, status updates) adds less than 30 seconds per task beyond the actual Goose processing time.

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

- Goose is installed locally by the user (MVP does not bundle or auto-install Goose).
- The user's LLM provider API key is configured either in Goose's own config or via SpecLedger's agent config (`sl config set agent.api-key`).
- Tasks in `issues.jsonl` are well-formed with sufficient description, acceptance criteria, and design notes for an AI agent to execute meaningfully.
- The repository is in a clean git state (no uncommitted changes) before agent execution begins.
- For MVP, tasks are executed sequentially (one at a time) on the local machine. Parallel execution is a P3 enhancement for cloud environments.
- Goose's `goose run` CLI command is the primary integration point — SpecLedger does not embed or fork the Goose runtime.
- Agent execution creates commits on a feature branch, not directly on main.
