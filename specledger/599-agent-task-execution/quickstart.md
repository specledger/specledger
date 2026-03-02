# Quickstart: Agent Task Execution

**Feature**: 599-agent-task-execution

## Prerequisites

1. **SpecLedger CLI** installed (`sl` command available)
2. **Goose** installed: `brew install block-goose-cli`
3. **LLM provider configured** (one of):
   - `sl config set agent.provider anthropic && sl config set agent.api-key <key>`
   - Or Goose's own config at `~/.config/goose/config.yaml`
4. **A spec with tasks**: `sl issue list` shows tasks in `open` status

## Quick Start

### 1. Check prerequisites

```bash
sl doctor          # Verify sl and dependencies
goose --version    # Verify Goose is installed
sl auth status     # Verify authenticated
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

# Follow real-time output
sl agent logs SL-abc123 --follow
```

### 5. Stop if needed

```bash
sl agent stop          # Graceful stop
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
max_turns: 75
extensions:
  - type: builtin
    name: developer
```

### Headless mode (CI/CD)

```bash
# Environment variables
export SPECLEDGER_AUTH_TOKEN=<supabase-token>

# Run headlessly
sl agent run --headless
```

## Typical Workflow

```
sl specify      → Create feature spec
sl clarify      → Clarify spec requirements
sl plan         → Generate implementation plan
sl tasks        → Generate tasks from plan
sl agent run    → Execute tasks with AI agent
sl agent status → Review results
# Human reviews and merges the branch
```
