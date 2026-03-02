# Research: AI Agent Task Execution Service

**Feature**: 599-agent-task-execution
**Date**: 2026-03-01

## Prior Work

### Epic: SL-aaf16e — Advanced Agent Model Configuration (597-agent-model-config)

**Status**: Complete (27/27 issues closed, PR #40 merged)

Provides the entire agent configuration infrastructure:
- **AgentConfig struct** with 15+ fields: BaseURL, AuthToken, APIKey, Model, Provider, PermissionMode, etc.
- **Config schema registry** (`pkg/cli/config/schema.go`): Centralized key definitions with types, env vars, defaults, sensitivity flags
- **Multi-layer config merge** (`pkg/cli/config/merge.go`): personal-local > team-local > profile > global > default with scope tracking
- **AgentLauncher** (`pkg/cli/launcher/launcher.go`): `BuildEnv()` injects resolved config as environment variables into agent subprocesses, `IsAvailable()` checks PATH, `Launch()`/`LaunchWithPrompt()` for subprocess execution
- **CLI commands**: `sl config set/get/show/unset` with scope flags, profile management
- **Profiles**: Named agent profiles (e.g., "work", "experimental") switchable via `sl config profile`

**Impact on 599**: Reuse AgentLauncher pattern for Goose invocation. Extend `DefaultAgents` to include Goose. Use `ResolveAgentEnv()` to inject provider/model config into Goose environment variables.

### Related: 597-issue-create-fields

**Status**: Complete (18/18 issues closed)

Provides the task metadata model:
- **Issue model** (`pkg/issues/issue.go`): ID, Title, Description, Status (open/in_progress/closed), Priority, BlockedBy, Blocks, AcceptanceCriteria, DefinitionOfDone, Design, Notes, ParentID
- **Store** (`pkg/issues/store.go`): JSONL-based CRUD with file locking (`gofrs/flock`), atomic writes
- **Dependency resolution** (`pkg/issues/dependencies.go`): `AddDependency()`, `RemoveDependency()`, `DetectCycles()`, DFS cycle detection, `GetBlockedIssuesWithBlockers()`
- **Ready-to-work queries**: `ListReady(filter)` returns issues with status=open and no open blockers — exactly what agent task pickup needs
- **DoD workflow**: `ChecklistItem` with `Checked` field, CLI `--check-dod` for marking items complete

**Impact on 599**: `ListReady()` is the task pickup query. Issue `Update()` manages status transitions. DoD verification uses existing checklist logic.

### Related: 598-spec-close (Draft)

Lifecycle tracking for completed specs. Complements agent execution by providing post-completion tracking. Low priority dependency.

---

## Research Topics

### 1. Goose CLI Integration

**Decision**: Use `goose run` with instruction files for task execution

**Rationale**:
- `goose run` is designed for automated, non-interactive execution — it auto-exits when the task is complete
- Instruction files (`-i`) allow rich, multi-line context without shell escaping issues
- `--no-session` prevents session accumulation during automated runs
- `--output-format json` enables structured result parsing
- Exit code 0 = success, non-zero = failure (fixed in goose PR #4621)

**Key CLI flags for our use case**:
```
goose run \
  -i <instruction-file>       # Task execution context
  --no-session                 # No persistent session needed
  --with-builtin developer     # Enable developer tools (file editing, shell)
  --provider <provider>        # From agent config
  --model <model>              # From agent config
  --output-format json         # For structured result capture
  --max-turns <N>              # Prevent runaway execution
  -q                           # Quiet mode for log cleanliness
```

**Environment variables for headless mode**:
```
GOOSE_MODE=auto                # Auto-approve tool operations
GOOSE_DISABLE_SESSION_NAMING=true
GOOSE_MAX_TURNS=50             # Configurable safety limit
```

**Alternatives considered**:
- `goose session` (interactive) — rejected: requires human interaction, not suitable for automation
- Embedding Goose as a library — rejected: Goose is a Rust binary, no Go SDK; CLI invocation is the documented integration point
- Recipe files — considered for future: recipes provide parameterized workflows but add complexity for MVP

### 2. Goose Detection and Installation

**Decision**: Use `exec.LookPath("goose")` + `goose --version` for detection

**Rationale**: Consistent with existing `AgentLauncher.IsAvailable()` pattern. Version check ensures the installed goose is the Block Inc variant, not the unrelated `goose` Homebrew package.

**Installation guidance**:
```
brew install block-goose-cli
```

**Alternative considered**: Auto-install via Homebrew — rejected: follows spec assumption that "Goose is installed locally by the user"

### 3. Task Selection Algorithm

**Decision**: Use existing `Store.ListReady()` with priority-based ordering

**Rationale**: `ListReady()` already returns issues with status=open and no open blockers. Adding priority sort (lower number = higher priority) gives deterministic, dependency-respecting task order.

**Algorithm**:
1. Call `store.ListReady(nil)` to get all unblocked open tasks
2. Sort by Priority (ascending), then by CreatedAt (ascending) for tie-breaking
3. If `--task <id>` flag provided, select that specific task (validate it's unblocked)
4. Pick the first task from sorted list

**Alternatives considered**:
- Topological sort of full dependency graph — rejected for MVP: `ListReady()` already handles this implicitly by filtering blocked tasks. Topological sort would be needed for parallel execution (P3)
- Random selection — rejected: non-deterministic, harder to debug

### 4. Execution Context Construction

**Decision**: Generate a temporary instruction file per task

**Rationale**: Instruction files allow rich, structured context without shell escaping issues. The file is passed to `goose run -i <file>`.

**Context template structure**:
```markdown
# Task: {title}
## Task ID: {id}
## Spec: {spec_context}

## Description
{description}

## Acceptance Criteria
{acceptance_criteria}

## Definition of Done
{definition_of_done_checklist}

## Design Notes
{design}

## Additional Notes
{notes}

## Repository Context
- Working directory: {repo_root}
- Branch: {current_branch}
- Spec directory: specledger/{spec_context}/

## Instructions
1. Read and understand the task requirements above
2. Implement the changes described in the description and design notes
3. Ensure all acceptance criteria are met
4. Verify each Definition of Done item
5. Commit your changes to the current branch with a descriptive message
6. Do NOT create new branches — commit directly to {current_branch}
```

**File location**: `specledger/<spec>/.agent-runs/<run-id>/task-<id>-instructions.md` (temp, gitignored)

### 5. Task Status Management

**Decision**: Use existing Issue `Update()` with status field transitions

**Rationale**: The JSONL store already supports status updates with file locking. Status transitions: open → in_progress → closed/needs_review.

**Flow**:
1. Before execution: `Update(id, {Status: "in_progress"})` — claim the task
2. On success (exit 0 + DoD verified): `Update(id, {Status: "closed", ClosedAt: now})`
3. On success with manual verification items: `Update(id, {Status: "needs_review"})`
4. On failure: Keep status as `in_progress`, append failure notes to `Notes` field
5. On stale detection (next run finds in_progress tasks): Offer retry/skip

**Locking**: File locking via `gofrs/flock` prevents concurrent modifications. For MVP (sequential execution), this is sufficient. Cloud parallel mode (P3) would need distributed locking.

### 6. Agent Run Tracking

**Decision**: New `AgentRun` model stored as JSON files in `specledger/<spec>/.agent-runs/`

**Rationale**: Agent runs are ephemeral operational data, not spec artifacts. Separate storage from issues.jsonl keeps concerns clean. JSON (not JSONL) since runs are individually accessed, not streamed.

**Storage layout**:
```
specledger/<spec>/.agent-runs/
├── latest.json              # Symlink/copy of most recent run
├── <run-id>.json            # Run metadata
└── <run-id>/
    ├── task-<id>-instructions.md  # Generated instruction file
    └── task-<id>-output.log       # Captured agent output
```

**Alternatives considered**:
- Store in issues.jsonl — rejected: conflates task definition with execution history
- SQLite — rejected: over-engineering for MVP, introduces new dependency
- Single runs.jsonl — considered: simpler but harder to manage individual runs

### 7. Goose Environment Variable Mapping

**Decision**: Map SpecLedger agent config to Goose environment variables

**Mapping table**:
| SpecLedger Config Key | Goose Env Var |
|---|---|
| `agent.provider` | `GOOSE_PROVIDER` |
| `agent.model` | `GOOSE_MODEL` |
| `agent.api-key` | `GOOSE_PROVIDER__API_KEY` + provider-specific (e.g., `ANTHROPIC_API_KEY`) |
| `agent.base-url` | `GOOSE_PROVIDER__HOST` |
| `agent.env.*` | Passed through directly |
| (hardcoded) | `GOOSE_MODE=auto` (headless) |
| (hardcoded) | `GOOSE_DISABLE_SESSION_NAMING=true` |

**Implementation**: Extend `ResolvedConfig.GetEnvVars()` or create a Goose-specific adapter that reads the resolved config and produces Goose-compatible env vars. The existing `AgentLauncher.SetEnv()` + `BuildEnv()` pattern handles injection.

### 8. Logging and Output Capture

**Decision**: Capture Goose stdout/stderr to per-task log files

**Rationale**: Essential for `sl agent status` and debugging. Use `cmd.Stdout = io.MultiWriter(os.Stdout, logFile)` for real-time display + capture.

**Log format**: Raw text output from Goose. Structured metadata (timing, exit code) stored in the run JSON.

**Alternatives considered**:
- `--output-format json` — considered for future: enables structured parsing but adds complexity for MVP
- Structured logging — rejected for MVP: Goose output is already formatted

### 9. Git Workflow

**Decision**: All agent commits go to the current feature branch (spec's branch)

**Rationale**: Per spec requirement FR-015, no per-task branches. The agent commits directly to the current branch.

**Pre-execution checks**:
1. Verify clean git state (no uncommitted changes)
2. Verify current branch matches spec context
3. After each task, verify commits were made to the correct branch

**Alternatives considered**:
- Per-task branches with merge — rejected: spec explicitly states "Single branch per run"
- Git worktrees for isolation — deferred to P3 (cloud parallel execution)

### 10. Graceful Stop Mechanism (FR-013)

**Decision**: Signal-based stop with PID tracking

**Rationale**: Store the Goose process PID in a lock file. `sl agent stop` sends SIGTERM, allowing Goose to finish its current turn. If unresponsive after timeout, send SIGKILL.

**Implementation**:
- PID file: `specledger/<spec>/.agent-runs/agent.pid`
- `sl agent stop`: Read PID, send SIGTERM, wait up to 30s, then SIGKILL
- `sl agent run` cleans up PID file on normal exit

### 11. Notification System (FR-016)

**Decision**: Defer bot notification to post-MVP; log `needs_review` status prominently

**Rationale**: The spec mentions "sends a confirmation message via bot" but the notification infrastructure (Telegram/Slack bot) is not yet defined. For MVP, tasks marked `needs_review` will be prominently displayed in `sl agent status` output.

**Future**: Integrate with whatever bot/notification system is established.
