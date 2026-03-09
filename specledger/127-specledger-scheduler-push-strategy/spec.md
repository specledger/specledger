# Feature Specification: Push-Triggered Scheduler Strategy

**Feature Branch**: `127-specledger-scheduler-push-strategy`
**Created**: 2026-03-08
**Status**: Draft
**Input**: User description: "Research how to utilize git hook push mechanic to trigger task execution. When user approves specs and pushes via git push, the system should go to the issues or plans to run sl implement."

## Clarifications

### Session 2026-03-09

- Q: Should the hook trigger `sl implement` as a single Go binary process with goroutines or spawn separate commands per task? → A: Single process + goroutines. The hook invokes `sl implement` once, which reads plan.md internally and uses goroutines/task queue to execute tasks.
- Q: Should tasks from plan.md execute sequentially or in parallel? → A: Sequential in dependency order. Tasks execute one by one as defined in plan.md to avoid race conditions.
- Q: Should `/specledger.approve` require all artifacts before allowing approval? → A: Yes, require spec.md + plan.md + tasks.md to all exist and be non-empty before approval.
- Q: Where should hook execution logs be stored? → A: `.specledger/logs/push-hook.log`, consistent with SpecLedger directory conventions.

## User Scenarios & Testing *(mandatory)*

### User Story 0 - Spec Approval Command (Priority: P1)

A developer finishes specifying and planning a feature using SpecLedger. They run `/specledger.approve` (or `sl approve`) to mark the spec as approved. The command validates that all required artifacts (spec.md, plan.md, tasks.md) exist and are non-empty before setting `**Status**: Approved` in spec.md. This approval is the gate that enables push-triggered implementation.

**Why this priority**: Without an approval mechanism, there is no trigger signal for the push hook. This is a prerequisite for the core automation.

**Independent Test**: Can be tested by running the approve command with complete vs incomplete artifacts and verifying status change.

**Acceptance Scenarios**:

1. **Given** a feature with spec.md, plan.md, and tasks.md all present and non-empty, **When** the user runs `sl approve`, **Then** spec.md's Status field is updated from "Draft" to "Approved".
2. **Given** a feature missing plan.md or tasks.md, **When** the user runs `sl approve`, **Then** the command fails with a clear message listing which artifacts are missing.
3. **Given** a feature already approved, **When** the user runs `sl approve`, **Then** the command reports that the spec is already approved.

---

### User Story 1 - Push-Triggered Implementation Execution (Priority: P1)

A developer finishes specifying and planning a feature using SpecLedger. They approve the spec via `sl approve`, then run `git push`. A `pre-push` git hook detects that approved SpecLedger artifacts exist for the current feature branch and spawns `sl implement` as a background process. The `sl implement` process reads plan.md, breaks down the tasks, and executes them sequentially using internal goroutines/task queue. The push completes without waiting for implementation to finish.

**Why this priority**: This is the core value proposition - automating the transition from planning to implementation via a familiar git workflow. Without this, the feature has no purpose.

**Independent Test**: Can be fully tested by creating a spec with approved status, pushing the branch, and verifying that `sl implement` is triggered automatically for that feature.

**Acceptance Scenarios**:

1. **Given** a feature branch with approved spec and tasks, **When** the user runs `git push`, **Then** the push hook detects the approved artifacts and triggers `sl implement` for that feature.
2. **Given** a feature branch with a draft (unapproved) spec, **When** the user runs `git push`, **Then** no automatic implementation is triggered and the push completes normally.
3. **Given** a feature branch with approved artifacts but `sl implement` is already running (`.specledger/exec.lock` exists with a live PID), **When** the user runs `git push`, **Then** the hook skips triggering to avoid duplicate execution and logs a message.
4. **Given** a feature branch with a stale `.specledger/exec.lock` (PID no longer running), **When** the user runs `git push`, **Then** the hook removes the stale lock, logs a warning, and proceeds to trigger `sl implement` normally.
5. **Given** a push-triggered implementation completes successfully, **Then** the generated code is committed to a `<feature-branch>/implement` sub-branch and the developer's working tree remains unchanged.

---

### User Story 2 - Hook Installation and Management (Priority: P2)

A developer sets up SpecLedger on their project and wants to enable push-triggered implementation. They run a command to install the git push hook. They can also uninstall it or check its status. The hook integrates cleanly with any existing git hooks the project may have.

**Why this priority**: Users need a reliable way to opt into and manage this behavior. Without installation tooling, the feature is not accessible.

**Independent Test**: Can be tested by running the install command and verifying the hook file exists in `.git/hooks/`, then uninstalling and verifying removal.

**Acceptance Scenarios**:

1. **Given** a SpecLedger project without the push hook, **When** the user runs the hook install command, **Then** a `pre-push` git hook is installed that triggers implementation on approved specs.
2. **Given** a project with existing git hooks, **When** the user installs the SpecLedger push hook, **Then** the existing hooks are preserved and the SpecLedger hook is added alongside them.
3. **Given** a project with the push hook installed, **When** the user runs the uninstall command, **Then** the SpecLedger hook is removed without affecting other hooks.
4. **Given** a project, **When** the user checks hook status, **Then** they see whether the push hook is currently installed and active.

---

### User Story 3 - Execution Feedback and Logging (Priority: P3)

After a push triggers implementation, the developer wants to know what happened. They can see a log of what the hook detected, whether implementation was triggered, and where to find the output. If something went wrong (e.g., missing dependencies, malformed spec), they receive clear feedback.

**Why this priority**: Observability is important for trust and debugging, but the feature works without it at a basic level.

**Independent Test**: Can be tested by triggering a push with an approved spec and checking that a log entry is created with hook execution details.

**Acceptance Scenarios**:

1. **Given** a push that triggers implementation, **When** the hook executes, **Then** a log entry is recorded with timestamp, feature name, and execution status.
2. **Given** a push where the hook encounters an error (e.g., spec is malformed), **When** the hook runs, **Then** the error is logged and the push is not blocked.
3. **Given** a developer wants to review past hook executions, **When** they check the log, **Then** they see a history of push-triggered implementations with outcomes.

---

### Edge Cases

- What happens when multiple approved features exist on the same branch? The hook should process only the feature associated with the current branch.
- What happens when the push is rejected by the remote? The hook should handle this gracefully without leaving partial state.
- What happens when the user pushes from a non-feature branch (e.g., main)? No implementation should be triggered.
- What happens when `sl` binary is not available in PATH during hook execution? The hook should fail gracefully with a clear error message.
- What happens during a force push? The hook should behave the same as a normal push.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a command to install a git push hook that monitors for approved SpecLedger artifacts.
- **FR-002**: System MUST provide a command to uninstall the git push hook cleanly.
- **FR-003**: System MUST provide a command to check the current installation status of the push hook.
- **FR-004**: The push hook MUST detect when pushed commits contain approved spec/task artifacts for the current feature branch.
- **FR-005**: The push hook MUST trigger `sl implement` as a single background process for the detected approved feature. The process reads plan.md and executes tasks sequentially using internal goroutines/task queue.
- **FR-006**: The push hook MUST use the `pre-push` git hook and spawn `sl implement` as a background process (detached from the push). The push MUST NOT be blocked or delayed by implementation execution.
- **FR-007**: The push hook MUST NOT trigger implementation if a `.specledger/exec.lock` file exists AND the process ID recorded in it is still running. The lock file MUST contain the PID and feature name. `sl implement` MUST create the lock on start and remove it on completion (success or failure).
- **FR-008**: The hook installation MUST preserve any existing git hooks in the project.
- **FR-009**: The push hook MUST only trigger on feature branches (matching the `NNN-feature-name` pattern).
- **FR-010**: System MUST log hook execution details to `.specledger/logs/push-hook.log` including timestamp, feature detected, action taken, and outcome.
- **FR-011**: The push hook MUST fail gracefully (log error, do not block push) if any error occurs during detection or triggering.
- **FR-015**: If a stale `.specledger/exec.lock` is detected (recorded PID no longer running), the hook MUST remove it and proceed normally, logging a warning about the stale lock.
- **FR-016**: `sl implement` MUST commit generated code to a sub-branch named `<feature-branch>/implement`. The developer's working tree MUST NOT be modified. A summary of changes MUST be written to `.specledger/logs/<feature>-result.md`.

### Approval Command

- **FR-012**: System MUST provide an `sl approve` command (and `/specledger.approve` Claude command) that sets `**Status**: Approved` in spec.md.
- **FR-013**: The approve command MUST validate that spec.md, plan.md, and tasks.md all exist and are non-empty before allowing approval. If any artifact is missing or empty, the command MUST fail with a clear error listing missing artifacts.
- **FR-014**: System MUST check approval status by reading the spec file's Status field from the working tree at push time. An "approved" spec is one with `**Status**: Approved`.

### Key Entities

- **Push Hook**: The git hook script installed in `.git/hooks/` that intercepts push events and checks for approved SpecLedger artifacts.
- **Approval Status**: A marker in the spec indicating that the feature is ready for implementation (Status field set to "Approved").
- **Hook Execution Log**: A structured log file at `.specledger/logs/push-hook.log` recording each hook invocation with detection results and actions taken.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can go from approved spec to running implementation with zero manual commands beyond `git push`.
- **SC-002**: Hook installation completes in under 5 seconds and requires a single command.
- **SC-003**: The push hook adds less than 2 seconds of overhead to the git push operation.
- **SC-004**: 100% of pushes on non-feature branches or with unapproved specs complete without triggering implementation.
- **SC-005**: Hook execution logs are available for the last 50 push events.
- **SC-006**: After push-triggered implementation completes, the developer can review all generated code via `git diff <feature-branch>..<feature-branch>/implement` without any working tree modifications.

### Previous work

- **Session Capture Hook (Feature 010)**: Established the Claude Code hook infrastructure in `pkg/cli/hooks/claude.go` for installing PostToolUse hooks. This feature follows a similar pattern but targets git hooks instead of Claude Code hooks.
- **SDD Workflow Streamline (Feature 598)**: Streamlined the bash CLI and spec verification workflow, which this feature builds upon for the implementation trigger.

## Dependencies & Assumptions

### Assumptions

- Git hooks are the appropriate mechanism for this trigger (as opposed to CI/CD pipelines or file watchers). Git hooks run locally on the developer's machine.
- The `pre-push` git hook will be used. It runs before the push completes but spawns `sl implement` as a detached background process so the push is not blocked. This avoids needing a wrapper or non-existent `post-push` hook.
- The `sl` binary will be available in PATH during hook execution.
- Only one feature is associated with each feature branch (the branch name determines the feature).
- Asynchronous execution after push is preferred over blocking the push to wait for implementation to complete.
- Standard git hook chaining practices (e.g., checking for existing hooks, appending rather than replacing) will be followed.

### Constraints

- Must work with the existing SpecLedger directory structure (`specledger/NNN-feature-name/`).
- Must not require changes to the remote repository or server-side hooks.
- Must work across macOS and Linux environments.
