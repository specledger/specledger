# Data Model: AI Agent Task Execution Service

**Feature**: 599-agent-task-execution
**Date**: 2026-03-01

## Entities

### 1. AgentRun

Represents a single invocation of `sl agent run` that may execute one or more tasks.

| Field | Type | Description | Constraints |
|---|---|---|---|
| `id` | string | Unique run identifier | Format: `run-<timestamp>-<4-hex>` (e.g., `run-20260301-143022-a3f5`) |
| `spec_context` | string | Spec being executed | Required. Format: `###-feature-name` |
| `status` | RunStatus | Current run status | `running`, `completed`, `failed`, `stopped` |
| `started_at` | time.Time | Run start timestamp | Set on creation |
| `ended_at` | *time.Time | Run end timestamp | Set on completion/failure/stop |
| `tasks_total` | int | Total tasks targeted | Set on creation after task discovery |
| `tasks_completed` | int | Successfully completed count | Incremented on task close |
| `tasks_failed` | int | Failed task count | Incremented on task failure |
| `tasks_skipped` | int | Skipped task count | Blocked tasks that couldn't run |
| `task_results` | []TaskResult | Per-task execution results | Appended after each task |
| `agent_command` | string | Agent CLI used | `goose` |
| `agent_model` | string | Model used for execution | From resolved config |
| `agent_provider` | string | Provider used | From resolved config |
| `headless` | bool | Whether running in headless mode | From `--headless` flag |
| `branch` | string | Git branch for commits | Current branch at run start |
| `error` | string | Top-level error message if run failed | Empty on success |

**State transitions**: `running` → `completed` | `failed` | `stopped`

### 2. TaskResult

Embedded in AgentRun — records the outcome of executing a single task.

| Field | Type | Description | Constraints |
|---|---|---|---|
| `task_id` | string | Issue ID (SL-xxxxxx) | Required |
| `task_title` | string | Issue title snapshot | For display without re-querying |
| `status` | TaskResultStatus | Execution outcome | `success`, `failed`, `skipped`, `needs_review` |
| `started_at` | time.Time | Task execution start | Set before Goose launch |
| `ended_at` | *time.Time | Task execution end | Set after Goose exits |
| `exit_code` | int | Goose process exit code | 0 = success |
| `output_log` | string | Path to captured output log | Relative to run directory |
| `instruction_file` | string | Path to generated instruction file | Relative to run directory |
| `commits` | []string | Git commit hashes produced | Captured by comparing HEAD before/after |
| `error` | string | Error description if failed | Empty on success |
| `dod_verified` | bool | Whether DoD items were checked | True if all DoD items addressed |

**State values**:
- `success`: Exit code 0, DoD verified, task closed
- `failed`: Non-zero exit code or DoD verification failed
- `skipped`: Task blocked or circular dependency detected
- `needs_review`: Success but has manual verification DoD items

### 3. ExecutionContext

Not persisted as a separate entity — generated as a temporary instruction file per task.

| Field | Type | Description |
|---|---|---|
| `task` | *Issue | Full issue from JSONL store |
| `spec_context` | string | Spec directory name |
| `repo_root` | string | Repository root path |
| `branch` | string | Current git branch |
| `run_id` | string | Parent AgentRun ID |
| `design_docs` | map[string]string | Relevant design artifact contents (plan.md, data-model.md, etc.) |
| `additional_instructions` | string | From agent recipe if present |

**Rendered as**: Markdown instruction file at `.agent-runs/<run-id>/task-<id>-instructions.md`

### 4. AgentRecipe (P4 — Optional)

Extends Goose recipe format with SpecLedger-specific fields.

| Field | Type | Description | Constraints |
|---|---|---|---|
| `instructions` | string | Additional Goose instructions | Appended to task context |
| `extensions` | []Extension | Goose MCP extensions to enable | Optional |
| `settings` | map[string]string | Goose settings overrides | Optional |
| `max_turns` | int | Override max turns per task | Optional, default from config |
| `pre_task_hook` | string | Shell command to run before each task | Optional |
| `post_task_hook` | string | Shell command to run after each task | Optional |

**File location**: `specledger/<spec>/agent-recipe.yaml`

---

## Relationships

```
AgentRun 1──* TaskResult       (one run contains many task results)
TaskResult *──1 Issue           (each result references one issue by ID)
AgentRun *──1 AgentConfig       (run uses resolved agent configuration)
AgentRun *──0..1 AgentRecipe    (run optionally uses a recipe)
ExecutionContext ──1 Issue      (context wraps one issue for Goose)
ExecutionContext ──1 AgentRun   (context references parent run)
```

---

## Storage Layout

```
specledger/<spec>/
├── issues.jsonl                    # Existing — task definitions (source of truth)
├── agent-recipe.yaml               # Optional — P4 custom execution config
└── .agent-runs/                    # Gitignored — execution artifacts
    ├── agent.pid                   # PID of running agent (lock mechanism)
    ├── latest.json                 # Copy of most recent run metadata
    ├── run-20260301-143022-a3f5.json  # Run metadata
    └── run-20260301-143022-a3f5/      # Run artifacts directory
        ├── task-SL-abc123-instructions.md  # Generated instruction file
        ├── task-SL-abc123-output.log       # Captured Goose output
        ├── task-SL-def456-instructions.md
        └── task-SL-def456-output.log
```

**Gitignore**: Add `.agent-runs/` to project `.gitignore` — execution artifacts are ephemeral and machine-specific.

---

## Validation Rules

### AgentRun
- `id` must be unique (timestamp + random hex ensures this)
- `spec_context` must match an existing `specledger/<spec>/` directory
- `branch` must match the current git branch
- `status` transitions: only `running` → {`completed`, `failed`, `stopped`}

### TaskResult
- `task_id` must exist in `issues.jsonl`
- `exit_code` must be captured (default -1 if process crashed)
- `started_at` must be before `ended_at`
- `commits` collected by comparing `git rev-parse HEAD` before and after execution

### ExecutionContext
- `task.Status` must be `open` before execution
- `task.BlockedBy` must all be `closed`
- Repository must be in clean git state (no uncommitted changes)

---

## Go Type Definitions

```go
// pkg/cli/agent/types.go

type RunStatus string

const (
    RunStatusRunning   RunStatus = "running"
    RunStatusCompleted RunStatus = "completed"
    RunStatusFailed    RunStatus = "failed"
    RunStatusStopped   RunStatus = "stopped"
)

type TaskResultStatus string

const (
    TaskResultSuccess     TaskResultStatus = "success"
    TaskResultFailed      TaskResultStatus = "failed"
    TaskResultSkipped     TaskResultStatus = "skipped"
    TaskResultNeedsReview TaskResultStatus = "needs_review"
)

type AgentRun struct {
    ID             string        `json:"id"`
    SpecContext    string        `json:"spec_context"`
    Status         RunStatus     `json:"status"`
    StartedAt      time.Time     `json:"started_at"`
    EndedAt        *time.Time    `json:"ended_at,omitempty"`
    TasksTotal     int           `json:"tasks_total"`
    TasksCompleted int           `json:"tasks_completed"`
    TasksFailed    int           `json:"tasks_failed"`
    TasksSkipped   int           `json:"tasks_skipped"`
    TaskResults    []TaskResult  `json:"task_results"`
    AgentCommand   string        `json:"agent_command"`
    AgentModel     string        `json:"agent_model"`
    AgentProvider  string        `json:"agent_provider"`
    Headless       bool          `json:"headless"`
    Branch         string        `json:"branch"`
    Error          string        `json:"error,omitempty"`
}

type TaskResult struct {
    TaskID          string           `json:"task_id"`
    TaskTitle       string           `json:"task_title"`
    Status          TaskResultStatus `json:"status"`
    StartedAt       time.Time        `json:"started_at"`
    EndedAt         *time.Time       `json:"ended_at,omitempty"`
    ExitCode        int              `json:"exit_code"`
    OutputLog       string           `json:"output_log"`
    InstructionFile string           `json:"instruction_file"`
    Commits         []string         `json:"commits,omitempty"`
    Error           string           `json:"error,omitempty"`
    DoDVerified     bool             `json:"dod_verified"`
}

type AgentRecipe struct {
    Instructions string            `yaml:"instructions,omitempty"`
    Extensions   []RecipeExtension `yaml:"extensions,omitempty"`
    Settings     map[string]string `yaml:"settings,omitempty"`
    MaxTurns     int               `yaml:"max_turns,omitempty"`
    PreTaskHook  string            `yaml:"pre_task_hook,omitempty"`
    PostTaskHook string            `yaml:"post_task_hook,omitempty"`
}

type RecipeExtension struct {
    Type string `yaml:"type"`  // builtin, stdio
    Name string `yaml:"name"`
    Cmd  string `yaml:"cmd,omitempty"`
    Args []string `yaml:"args,omitempty"`
}
```
