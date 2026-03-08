# Quickstart: Agent Task Execution

**Feature**: 599-agent-task-execution
**Updated**: 2026-03-09

## Prerequisites

1. **SpecLedger CLI** installed (`sl` command available)
2. **Anthropic API key** configured:
   - `sl config set agent.api-key sk-ant-...`
3. **A spec with tasks**: `sl issue list` shows tasks in `open` status

> **Note**: No external agent binary (Goose, Claude Code) is required. The agent runs via the Anthropic Messages API directly from within `sl`.

## Quick Start

### 1. Check prerequisites

```bash
sl doctor          # Verify sl and dependencies
sl auth status     # Verify authenticated
sl config get agent.api-key  # Verify API key is set
```

### 2. Verify tasks are ready

```bash
# List all tasks for current spec
sl issue list

# Check which tasks are ready (unblocked)
sl issue list --status open
```

### 3. Run the agent

```bash
# Execute all ready tasks sequentially
sl agent run

# Execute a specific task
sl agent run --task SL-abc123

# Preview what would run (no execution)
sl agent run --dry-run
```

### 4. Monitor progress

```bash
# In another terminal
sl agent status

# View logs for a specific task
sl agent logs SL-abc123

# View last 50 lines
sl agent logs SL-abc123 --tail 50
```

### 5. Stop if needed

```bash
sl agent stop          # Graceful stop (finishes current tool call)
sl agent stop --force  # Force kill
```

## Configuration

### Agent model (via sl config)

```bash
# Set model and provider
sl config set agent.provider anthropic
sl config set agent.model claude-sonnet-4-20250514
sl config set agent.api-key sk-ant-...

# Or use profiles
sl config profile create fast-model
sl config set agent.model claude-haiku-4-5-20251001 --profile fast-model
sl config profile use fast-model
```

### Agent recipe (optional, per-spec)

Create `specledger/<spec>/agent-recipe.yaml`:

```yaml
instructions: |
  This project uses Go 1.24. Follow standard Go conventions.
  Run `make fmt` after changes. Run `make test` to verify.
max_iterations: 75
max_tokens_per_task: 150000
```

### Headless mode (CI/CD)

```bash
# Environment variables
export SPECLEDGER_AUTH_TOKEN=<supabase-token>

# Run headlessly
sl agent run --headless
```

## How It Works

The agent uses the **Anthropic Messages API** with a manual agentic loop:

1. **Task pickup**: Selects the next unblocked task from `issues.jsonl`
2. **Prompt construction**: Builds a system prompt (5-step pipeline) and user prompt (task details)
3. **Agentic loop**: Calls the Messages API with 3 tools (`bash`, `file_read`, `file_write`)
   - Model returns tool calls → `sl` executes them → results fed back → repeat
   - Loop ends when model returns final text (no more tool calls)
4. **Verification**: Checks for git commits and DoD completion
5. **Status update**: Marks task as `closed`, `needs_review`, or keeps `in_progress`

Each task's execution is logged to `.agent-runs/<run-id>/task-<id>-output.log` with every API call and tool invocation recorded.

## Typical Workflow

```
sl specify      → Create feature spec
sl clarify      → Clarify spec requirements
sl plan         → Generate implementation plan
sl tasks        → Generate tasks from plan
sl agent run    → Execute tasks with Claude API
sl agent status → Review results
# Human reviews and merges the branch
```

## Token Usage

Token usage is tracked per-task and per-run:
- View via `sl agent status` after a run
- Default budget: 100K tokens per task (configurable via `agent.max-tokens-per-task`)
- Tasks exceeding budget are marked `needs_review`
