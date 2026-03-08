# Feature Specification: Push-Triggered Scheduler Strategy

**Feature Branch**: `127-specledger-scheduler-push-strategy`
**Created**: 2026-03-08
**Status**: Draft
**Input**: User description: "Research how to utilize git hook push mechanic to trigger task execution. When user approves specs and pushes via git push, the system should go to the issues or plans to run sl implement."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Push-Triggered Implementation Execution (Priority: P1)

A developer finishes specifying and planning a feature using SpecLedger. They approve the spec and tasks, then run `git push`. A git hook detects that approved SpecLedger artifacts are being pushed and automatically triggers `sl implement` for the approved feature, running the implementation workflow without manual intervention.

**Why this priority**: This is the core value proposition - automating the transition from planning to implementation via a familiar git workflow. Without this, the feature has no purpose.

**Independent Test**: Can be fully tested by creating a spec with approved status, pushing the branch, and verifying that `sl implement` is triggered automatically for that feature.

**Acceptance Scenarios**:

1. **Given** a feature branch with approved spec and tasks, **When** the user runs `git push`, **Then** the push hook detects the approved artifacts and triggers `sl implement` for that feature.
2. **Given** a feature branch with a draft (unapproved) spec, **When** the user runs `git push`, **Then** no automatic implementation is triggered and the push completes normally.
3. **Given** a feature branch with approved artifacts but `sl implement` is already running or completed, **When** the user runs `git push`, **Then** the hook skips triggering to avoid duplicate execution.

---

### User Story 2 - Hook Installation and Management (Priority: P2)

A developer sets up SpecLedger on their project and wants to enable push-triggered implementation. They run a command to install the git push hook. They can also uninstall it or check its status. The hook integrates cleanly with any existing git hooks the project may have.

**Why this priority**: Users need a reliable way to opt into and manage this behavior. Without installation tooling, the feature is not accessible.

**Independent Test**: Can be tested by running the install command and verifying the hook file exists in `.git/hooks/`, then uninstalling and verifying removal.

**Acceptance Scenarios**:

1. **Given** a SpecLedger project without the push hook, **When** the user runs the hook install command, **Then** a pre-push or post-push git hook is installed that triggers implementation on approved specs.
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
- **FR-005**: The push hook MUST trigger `sl implement` for the detected approved feature after a successful push.
- **FR-006**: The push hook MUST NOT block or delay the git push operation itself. Implementation execution should happen asynchronously after the push completes.
- **FR-007**: The push hook MUST NOT trigger implementation if the feature's implementation is already in progress or completed.
- **FR-008**: The hook installation MUST preserve any existing git hooks in the project.
- **FR-009**: The push hook MUST only trigger on feature branches (matching the `NNN-feature-name` pattern).
- **FR-010**: System MUST log hook execution details including timestamp, feature detected, action taken, and outcome.
- **FR-011**: The push hook MUST fail gracefully (log error, do not block push) if any error occurs during detection or triggering.

### Approval Detection

- **FR-012**: System MUST define what constitutes an "approved" spec - a spec.md with `**Status**: Approved` and a corresponding tasks.md present in the feature directory.
- **FR-013**: System MUST check approval status by reading the spec file from the working tree at push time.

### Key Entities

- **Push Hook**: The git hook script installed in `.git/hooks/` that intercepts push events and checks for approved SpecLedger artifacts.
- **Approval Status**: A marker in the spec indicating that the feature is ready for implementation (Status field set to "Approved").
- **Hook Execution Log**: A record of each hook invocation with detection results and actions taken.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can go from approved spec to running implementation with zero manual commands beyond `git push`.
- **SC-002**: Hook installation completes in under 5 seconds and requires a single command.
- **SC-003**: The push hook adds less than 2 seconds of overhead to the git push operation.
- **SC-004**: 100% of pushes on non-feature branches or with unapproved specs complete without triggering implementation.
- **SC-005**: Hook execution logs are available for the last 50 push events.

### Previous work

- **Session Capture Hook (Feature 010)**: Established the Claude Code hook infrastructure in `pkg/cli/hooks/claude.go` for installing PostToolUse hooks. This feature follows a similar pattern but targets git hooks instead of Claude Code hooks.
- **SDD Workflow Streamline (Feature 598)**: Streamlined the bash CLI and spec verification workflow, which this feature builds upon for the implementation trigger.

## Dependencies & Assumptions

### Assumptions

- Git hooks are the appropriate mechanism for this trigger (as opposed to CI/CD pipelines or file watchers). Git hooks run locally on the developer's machine.
- The `post-push` event does not exist natively in git; the implementation will likely use `pre-push` (which runs before the push) or a wrapper around `git push`. This is a key technical consideration for the planning phase.
- The `sl` binary will be available in PATH during hook execution.
- Only one feature is associated with each feature branch (the branch name determines the feature).
- Asynchronous execution after push is preferred over blocking the push to wait for implementation to complete.
- Standard git hook chaining practices (e.g., checking for existing hooks, appending rather than replacing) will be followed.

### Constraints

- Must work with the existing SpecLedger directory structure (`specledger/NNN-feature-name/`).
- Must not require changes to the remote repository or server-side hooks.
- Must work across macOS and Linux environments.
