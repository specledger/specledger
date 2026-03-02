# CLI Interface Contracts: Agent Task Execution

**Feature**: 599-agent-task-execution
**Date**: 2026-03-01

This feature adds CLI commands only (no REST/RPC API). Contracts are defined as CLI interface specifications.

---

## Command: `sl agent run`

**Description**: Execute tasks for the current spec using a Goose AI agent.

### Usage

```
sl agent run [flags]
```

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--spec` | string | (auto-detect from branch) | Spec context to execute tasks for |
| `--task` | string | (none) | Execute a specific task by ID |
| `--headless` | bool | false | Run without interactive prompts (for CI/CD) |
| `--max-turns` | int | 50 | Maximum Goose turns per task |
| `--dry-run` | bool | false | Show tasks that would be executed without running them |

### Behavior

1. **Auto-detect spec**: Parse current git branch name → spec context
2. **Verify prerequisites**: Goose installed, clean git state, valid spec
3. **Discover tasks**: Load `issues.jsonl`, filter ready (open, unblocked) tasks
4. **Sequential execution**: For each task in priority order:
   a. Update status → `in_progress`
   b. Generate instruction file
   c. Launch `goose run -i <file>` with configured environment
   d. Capture output and exit code
   e. Verify DoD items
   f. Update status → `closed` / `needs_review` / keep `in_progress` (failure)
5. **Save run metadata**: Write `AgentRun` JSON to `.agent-runs/`

### Exit Codes

| Code | Meaning |
|---|---|
| 0 | All tasks completed successfully |
| 1 | One or more tasks failed (partial success) |
| 2 | No tasks available to execute |
| 3 | Prerequisites not met (no Goose, dirty git, etc.) |

### Output (stdout)

```
Agent Run: run-20260301-143022-a3f5
Spec: 599-agent-task-execution
Branch: 599-agent-task-execution
Model: claude-sonnet-4-20250514 (anthropic)

Tasks discovered: 5 (3 ready, 2 blocked)

[1/3] Executing: SL-abc123 - Implement agent run command
  Status: in_progress
  ... (Goose output) ...
  Status: closed ✓ (2m 34s)

[2/3] Executing: SL-def456 - Add task selection logic
  Status: in_progress
  ... (Goose output) ...
  Status: closed ✓ (1m 12s)

[3/3] Executing: SL-ghi789 - Write unit tests
  Status: in_progress
  ... (Goose output) ...
  Status: failed ✗ (exit code 1)

Summary:
  Completed: 2/3
  Failed: 1/3
  Duration: 4m 18s
```

---

## Command: `sl agent status`

**Description**: Show status of the current or most recent agent run.

### Usage

```
sl agent status [flags]
```

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--spec` | string | (auto-detect) | Spec context |
| `--run` | string | (latest) | Specific run ID |
| `--json` | bool | false | Output as JSON |

### Output (stdout)

```
Last Run: run-20260301-143022-a3f5
Status: completed
Started: 2026-03-01 14:30:22
Ended: 2026-03-01 14:34:40
Duration: 4m 18s

Tasks:
  ✓ SL-abc123  Implement agent run command     closed        2m 34s
  ✓ SL-def456  Add task selection logic         closed        1m 12s
  ✗ SL-ghi789  Write unit tests                 in_progress   0m 32s  (exit 1)

Summary: 2 completed, 1 failed, 0 skipped
```

### JSON Output

```json
{
  "id": "run-20260301-143022-a3f5",
  "spec_context": "599-agent-task-execution",
  "status": "completed",
  "started_at": "2026-03-01T14:30:22Z",
  "ended_at": "2026-03-01T14:34:40Z",
  "tasks_total": 3,
  "tasks_completed": 2,
  "tasks_failed": 1,
  "task_results": [...]
}
```

---

## Command: `sl agent logs`

**Description**: View captured output from a task execution.

### Usage

```
sl agent logs <task-id> [flags]
```

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--spec` | string | (auto-detect) | Spec context |
| `--run` | string | (latest) | Specific run ID |
| `--tail` | int | 0 | Show only last N lines |
| `--follow` | bool | false | Follow output in real-time (if running) |

### Output

Raw captured Goose output for the specified task.

---

## Command: `sl agent stop`

**Description**: Gracefully stop a running agent.

### Usage

```
sl agent stop [flags]
```

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--spec` | string | (auto-detect) | Spec context |
| `--force` | bool | false | Force kill after timeout |

### Behavior

1. Read PID from `.agent-runs/agent.pid`
2. Send SIGTERM to Goose process
3. Wait up to 30 seconds for graceful shutdown
4. If `--force` or timeout: send SIGKILL
5. Update run status → `stopped`
6. Clean up PID file

---

## Internal Interface: Goose Invocation

Not a user-facing contract but documents the subprocess invocation pattern.

### Command Construction

```go
// For each task:
args := []string{
    "run",
    "-i", instructionFilePath,
    "--no-session",
    "--with-builtin", "developer",
    "--max-turns", strconv.Itoa(maxTurns),
}

// Headless mode adds:
if headless {
    args = append(args, "-q")
}

cmd := exec.Command("goose", args...)
cmd.Dir = repoRoot
cmd.Env = buildGooseEnv(resolvedConfig)
cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
```

### Environment Variables Injected

```
GOOSE_PROVIDER=<from agent.provider>
GOOSE_MODEL=<from agent.model>
GOOSE_PROVIDER__API_KEY=<from agent.api-key>
GOOSE_PROVIDER__HOST=<from agent.base-url>
GOOSE_MODE=auto                          # Always auto for task execution
GOOSE_DISABLE_SESSION_NAMING=true        # No interactive naming
GOOSE_MAX_TURNS=<from config or default>
<agent.env.*>                            # Custom env vars passed through
```
