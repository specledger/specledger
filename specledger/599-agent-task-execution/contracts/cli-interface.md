# CLI Interface Contracts: Agent Task Execution

**Feature**: 599-agent-task-execution
**Date**: 2026-03-01
**Updated**: 2026-03-09 (Claude Go SDK migration)

This feature adds CLI commands only (no REST/RPC API). Contracts are defined as CLI interface specifications.

---

## Command: `sl agent run`

**Description**: Execute tasks for the current spec using the Claude API via the Anthropic Go SDK.

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
| `--max-iterations` | int | 50 | Maximum agentic loop iterations per task |
| `--dry-run` | bool | false | Show tasks that would be executed without running them |

### Behavior

1. **Auto-detect spec**: Parse current git branch name → spec context
2. **Verify prerequisites**: API key configured, clean git state, valid spec
3. **Discover tasks**: Load `issues.jsonl`, filter ready (open, unblocked) tasks
4. **Sequential execution**: For each task in priority order:
   a. Update status → `in_progress`
   b. Build system prompt (pipeline instructions) + user prompt (task metadata)
   c. Run agentic loop via Anthropic Messages API with `bash`, `file_read`, `file_write` tools
   d. Log every API call and tool invocation
   e. Verify DoD items and capture git commits
   f. Update status → `closed` / `needs_review` / keep `in_progress` (failure)
5. **Save run metadata**: Write `AgentRun` JSON to `.agent-runs/`

### Exit Codes

| Code | Meaning |
|---|---|
| 0 | All tasks completed successfully |
| 1 | One or more tasks failed (partial success) |
| 2 | No tasks available to execute |
| 3 | Prerequisites not met (no API key, dirty git, etc.) |

### Output (stdout)

```
Agent Run: run-20260301-143022-a3f5
Spec: 599-agent-task-execution
Branch: 599-agent-task-execution
Model: claude-sonnet-4-20250514 (anthropic)

Tasks discovered: 5 (3 ready, 2 blocked)

[1/3] Executing: SL-abc123 - Implement agent run command
  Status: in_progress
  Iteration 1: THINK — exploring codebase (3 tool calls)
  Iteration 2: WRITE — creating types.go (2 tool calls)
  Iteration 3: LINT — running linter (1 tool call)
  Iteration 4: COMMIT — staging and committing (2 tool calls)
  Iteration 5: TEST — running tests (1 tool call)
  Status: closed ✓ (2m 34s, 45K tokens, 9 tool calls)

[2/3] Executing: SL-def456 - Add task selection logic
  Status: in_progress
  ...
  Status: closed ✓ (1m 12s, 28K tokens, 6 tool calls)

[3/3] Executing: SL-ghi789 - Write unit tests
  Status: in_progress
  ...
  Status: failed ✗ (token budget exceeded at 100K tokens)

Summary:
  Completed: 2/3
  Failed: 1/3
  Duration: 4m 18s
  Total tokens: 173K (input: 125K, output: 48K)
  Total API calls: 15
  Total tool calls: 24
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
Model: claude-sonnet-4-20250514

Tasks:
  ✓ SL-abc123  Implement agent run command     closed        2m 34s  45K tokens
  ✓ SL-def456  Add task selection logic         closed        1m 12s  28K tokens
  ✗ SL-ghi789  Write unit tests                 in_progress   0m 32s  100K tokens (budget exceeded)

Summary: 2 completed, 1 failed, 0 skipped
Tokens: 173K total (125K input, 48K output)
API calls: 15 | Tool calls: 24
```

### JSON Output

```json
{
  "id": "run-20260301-143022-a3f5",
  "spec_context": "599-agent-task-execution",
  "status": "completed",
  "started_at": "2026-03-01T14:30:22Z",
  "ended_at": "2026-03-01T14:34:40Z",
  "agent_model": "claude-sonnet-4-20250514",
  "agent_provider": "anthropic",
  "tasks_total": 3,
  "tasks_completed": 2,
  "tasks_failed": 1,
  "tokens_input": 125000,
  "tokens_output": 48000,
  "api_calls": 15,
  "tool_calls": 24,
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

### Output

Structured execution log showing API calls, tool invocations, and model text output:

```
[14:30:22] API CALL #1 (2,450 in / 830 out tokens)
[14:30:25] TOOL: bash {"command": "ls pkg/cli/agent/"}
[14:30:25] RESULT: (success, 234 bytes, 0.3s)
[14:30:25] TEXT: I'll start by examining the existing code structure...
[14:30:25] API CALL #2 (3,280 in / 1,200 out tokens)
[14:30:30] TOOL: file_write {"path": "pkg/cli/agent/types.go"}
[14:30:30] RESULT: (success, wrote 1,482 bytes, 0.1s)
...
[14:32:15] LOOP COMPLETE: 8 iterations, 12 tool calls, 45,230 total tokens
```

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
| `--force` | bool | false | Force kill the process |

### Behavior

1. Write stop sentinel file to `.agent-runs/agent.stop`
2. Runner detects sentinel between agentic loop iterations
3. Current tool execution completes, then loop exits
4. If `--force`: send SIGKILL to the process via PID file
5. Update run status → `stopped`
6. Clean up sentinel and PID files

---

## Internal Interface: Anthropic SDK Invocation

Not a user-facing contract but documents the API call pattern.

### Client Construction

```go
client := anthropic.NewClient(
    anthropic.WithAPIKey(resolvedConfig.GetValue("agent.api-key").(string)),
    // optional: option.WithBaseURL(...)
)
```

### Messages API Call

```go
msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
    Model:     anthropic.Model(resolvedConfig.GetValue("agent.model").(string)),
    MaxTokens: 4096,
    System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
    Messages:  messages,
    Tools:     tools,  // bash, file_read, file_write
})
```

### Tool Definitions

```go
tools := []anthropic.ToolUnionParam{
    {OfTool: &anthropic.ToolParam{
        Name:        "bash",
        Description: anthropic.String("Execute a shell command"),
        InputSchema: GenerateSchema[BashInput](),
    }},
    {OfTool: &anthropic.ToolParam{
        Name:        "file_read",
        Description: anthropic.String("Read file contents"),
        InputSchema: GenerateSchema[FileReadInput](),
    }},
    {OfTool: &anthropic.ToolParam{
        Name:        "file_write",
        Description: anthropic.String("Write content to a file"),
        InputSchema: GenerateSchema[FileWriteInput](),
    }},
}
```
