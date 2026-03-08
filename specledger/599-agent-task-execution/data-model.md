# Data Model: AI Agent Task Execution Service

**Feature**: 599-agent-task-execution
**Date**: 2026-03-01
**Updated**: 2026-03-09 (Claude Go SDK migration)

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
| `agent_model` | string | Model used for execution | From resolved config (e.g., `claude-sonnet-4-20250514`) |
| `agent_provider` | string | Provider used | From resolved config (e.g., `anthropic`) |
| `headless` | bool | Whether running in headless mode | From `--headless` flag |
| `branch` | string | Git branch for commits | Current branch at run start |
| `tokens_input` | int64 | Total input tokens across all tasks | Sum of task-level input tokens |
| `tokens_output` | int64 | Total output tokens across all tasks | Sum of task-level output tokens |
| `api_calls` | int | Total Messages API calls | Sum of task-level API calls |
| `tool_calls` | int | Total tool invocations | Sum of task-level tool calls |
| `error` | string | Top-level error message if run failed | Empty on success |

**State transitions**: `running` → `completed` | `failed` | `stopped`

### 2. TaskResult

Embedded in AgentRun — records the outcome of executing a single task.

| Field | Type | Description | Constraints |
|---|---|---|---|
| `task_id` | string | Issue ID (SL-xxxxxx) | Required |
| `task_title` | string | Issue title snapshot | For display without re-querying |
| `status` | TaskResultStatus | Execution outcome | `success`, `failed`, `skipped`, `needs_review` |
| `started_at` | time.Time | Task execution start | Set before agentic loop launch |
| `ended_at` | *time.Time | Task execution end | Set after agentic loop completes |
| `exit_code` | int | Synthetic exit code | 0 = success, 1 = failed (for compat) |
| `iterations` | int | Agentic loop iterations | Number of API call rounds |
| `tokens_input` | int64 | Input tokens for this task | From `msg.Usage.InputTokens` |
| `tokens_output` | int64 | Output tokens for this task | From `msg.Usage.OutputTokens` |
| `api_calls` | int | Messages API calls for this task | Count of `Messages.New()` calls |
| `tool_calls` | int | Tool invocations for this task | Count of tool use blocks processed |
| `output_log` | string | Path to captured output log | Relative to run directory |
| `prompt_file` | string | Path to saved prompt file | Relative to run directory |
| `commits` | []string | Git commit hashes produced | Captured by comparing HEAD before/after |
| `error` | string | Error description if failed | Empty on success |
| `dod_verified` | bool | Whether DoD items were checked | True if all DoD items addressed |

**State values**:
- `success`: Agentic loop completed, DoD verified, task closed
- `failed`: API error, budget exceeded, or DoD verification failed
- `skipped`: Task blocked or circular dependency detected
- `needs_review`: Success but has manual verification DoD items, or budget/iteration limit reached

### 3. ExecutionContext

Not persisted as a separate entity — used to construct system prompt + user prompt per task.

| Field | Type | Description |
|---|---|---|
| `task` | *Issue | Full issue from JSONL store |
| `spec_context` | string | Spec directory name |
| `repo_root` | string | Repository root path |
| `branch` | string | Current git branch |
| `run_id` | string | Parent AgentRun ID |
| `lint_command` | string | Lint command (default: `make lint`) |
| `test_command` | string | Test command (default: `make test`) |
| `design_docs` | map[string]string | Relevant design artifact contents (plan.md, data-model.md, etc.) |
| `additional_instructions` | string | From agent recipe if present |

**Rendered as**: System prompt (pipeline instructions, repo context) + User prompt (task metadata). Both saved to `.agent-runs/<run-id>/task-<id>-prompt.md` for debugging.

### 4. ToolResult

In-memory only — represents the result of executing a single tool call.

| Field | Type | Description |
|---|---|---|
| `name` | string | Tool name (`bash`, `file_read`, `file_write`) |
| `output` | string | Tool output (stdout for bash, file content for read, confirmation for write) |
| `is_error` | bool | Whether the tool execution errored |
| `duration` | time.Duration | Execution time |
| `input_summary` | string | Truncated input for logging |

### 5. ExecutionResult

In-memory only — returned by the Executor after completing an agentic loop for one task.

| Field | Type | Description |
|---|---|---|
| `tokens_input` | int64 | Total input tokens |
| `tokens_output` | int64 | Total output tokens |
| `api_calls` | int | Number of API calls |
| `tool_calls` | int | Number of tool invocations |
| `iterations` | int | Number of agentic loop iterations |
| `final_text` | string | Last text response from the model |
| `error` | error | Error if loop failed |

### 6. AgentRecipe (P4 — Optional)

Custom execution configuration per spec.

| Field | Type | Description | Constraints |
|---|---|---|---|
| `instructions` | string | Additional system prompt content | Appended to pipeline prompt |
| `tools` | []string | Additional tool names to enable | Optional, future extensibility |
| `max_iterations` | int | Override max iterations per task | Optional, default from config |
| `max_tokens_per_task` | int64 | Override token budget per task | Optional, default from config |
| `pre_task_hook` | string | Shell command to run before each task | Optional |
| `post_task_hook` | string | Shell command to run after each task | Optional |

**File location**: `specledger/<spec>/agent-recipe.yaml`

---

## Relationships

```
AgentRun 1──* TaskResult       (one run contains many task results)
TaskResult *──1 Issue           (each result references one issue by ID)
AgentRun *──1 ResolvedConfig   (run uses resolved agent configuration)
AgentRun *──0..1 AgentRecipe   (run optionally uses a recipe)
ExecutionContext ──1 Issue     (context wraps one issue for the executor)
ExecutionContext ──1 AgentRun  (context references parent run)
Executor 1──* ToolHandler      (executor dispatches to registered tool handlers)
```

---

## Storage Layout

```
specledger/<spec>/
├── issues.jsonl                    # Existing — task definitions (source of truth)
├── agent-recipe.yaml               # Optional — P4 custom execution config
└── .agent-runs/                    # Gitignored — execution artifacts
    ├── agent.stop                  # Sentinel file for graceful stop
    ├── latest.json                 # Copy of most recent run metadata
    ├── run-20260301-143022-a3f5.json  # Run metadata
    └── run-20260301-143022-a3f5/      # Run artifacts directory
        ├── task-SL-abc123-prompt.md       # Saved system + user prompt
        ├── task-SL-abc123-output.log      # Structured execution log
        ├── task-SL-def456-prompt.md
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
- `started_at` must be before `ended_at`
- `commits` collected by comparing `git rev-parse HEAD` before and after execution
- `tokens_input` and `tokens_output` must be non-negative

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
    AgentModel     string        `json:"agent_model"`
    AgentProvider  string        `json:"agent_provider"`
    Headless       bool          `json:"headless"`
    Branch         string        `json:"branch"`
    TokensInput    int64         `json:"tokens_input"`
    TokensOutput   int64         `json:"tokens_output"`
    APICalls       int           `json:"api_calls"`
    ToolCalls      int           `json:"tool_calls"`
    Error          string        `json:"error,omitempty"`
}

type TaskResult struct {
    TaskID       string           `json:"task_id"`
    TaskTitle    string           `json:"task_title"`
    Status       TaskResultStatus `json:"status"`
    StartedAt    time.Time        `json:"started_at"`
    EndedAt      *time.Time       `json:"ended_at,omitempty"`
    ExitCode     int              `json:"exit_code"`
    Iterations   int              `json:"iterations"`
    TokensInput  int64            `json:"tokens_input"`
    TokensOutput int64            `json:"tokens_output"`
    APICalls     int              `json:"api_calls"`
    ToolCalls    int              `json:"tool_calls"`
    OutputLog    string           `json:"output_log"`
    PromptFile   string           `json:"prompt_file"`
    Commits      []string         `json:"commits,omitempty"`
    Error        string           `json:"error,omitempty"`
    DoDVerified  bool             `json:"dod_verified"`
}

type ToolResult struct {
    Output       string        `json:"-"`
    IsError      bool          `json:"-"`
    Duration     time.Duration `json:"-"`
    InputSummary string        `json:"-"`
}

type ExecutionResult struct {
    TokensInput  int64  `json:"tokens_input"`
    TokensOutput int64  `json:"tokens_output"`
    APICalls     int    `json:"api_calls"`
    ToolCalls    int    `json:"tool_calls"`
    Iterations   int    `json:"iterations"`
    FinalText    string `json:"final_text"`
    Error        error  `json:"-"`
}

type AgentRecipe struct {
    Instructions     string   `yaml:"instructions,omitempty"`
    Tools            []string `yaml:"tools,omitempty"`
    MaxIterations    int      `yaml:"max_iterations,omitempty"`
    MaxTokensPerTask int64    `yaml:"max_tokens_per_task,omitempty"`
    PreTaskHook      string   `yaml:"pre_task_hook,omitempty"`
    PostTaskHook     string   `yaml:"post_task_hook,omitempty"`
}
```
