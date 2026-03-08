# Research: AI Agent Task Execution Service

**Feature**: 599-agent-task-execution
**Date**: 2026-03-01
**Updated**: 2026-03-09 (Claude Go SDK migration)

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

**Impact on 599**: Reuse config resolution from `ResolvedConfig` to get API key, model, provider, base URL. The existing `AgentLauncher` subprocess pattern is being **replaced** by direct Anthropic API calls via the Claude Go SDK. Config schema keys (`agent.api-key`, `agent.model`, `agent.provider`, `agent.base-url`) are reused directly.

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

### 1. Claude Go SDK — Replacing Goose CLI

**Decision**: Use `github.com/anthropics/anthropic-sdk-go` for direct API-driven agent execution, replacing Goose CLI subprocess

**Rationale**:
- **No external binary dependency** — eliminates Goose installation requirement (FR-008 simplified)
- **Full control over tool execution** — inspect every tool call, handle errors granularly per-step
- **In-process observability** — emit metrics at every tool call, not just start/end (FR-007 enhanced)
- **Native Go integration** — no subprocess overhead, no env var mapping, no output parsing
- **Pipeline orchestration** — can split 5-step pipeline into separate prompts or run as single agentic loop
- **Cost control** — track token usage per task, enforce budgets

**SDK Details** (v1.26.0, MIT license, Go 1.22+):
```
go get -u github.com/anthropics/anthropic-sdk-go@v1.26.0
```

**Alternatives considered**:
- **Goose CLI subprocess** (previous plan) — rejected: external binary dependency, black-box execution, limited observability, env var mapping complexity
- **Community Agent SDK wrappers** (schlunsen/claude-agent-sdk-go, lancekrogers/claude-code-go) — rejected: wrap Claude Code CLI as subprocess, not official, adds another binary dependency
- **Claude Agent SDK (Python)** — rejected: Python, not Go; no official Go port
- **OpenAI Go SDK** — rejected: not using OpenAI models; Anthropic SDK is the native choice

### 2. Agentic Loop Pattern

**Decision**: Manual agentic loop (not BetaToolRunner) for maximum control

The core execution pattern is a message loop that continues until the model stops requesting tool calls:

```go
func (e *Executor) RunAgenticLoop(ctx context.Context, systemPrompt string, userPrompt string, tools []anthropic.ToolUnionParam) (*anthropic.Message, error) {
    messages := []anthropic.MessageParam{
        anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
    }

    var lastMsg *anthropic.Message
    for iteration := 0; iteration < e.maxIterations; iteration++ {
        msg, err := e.client.Messages.New(ctx, anthropic.MessageNewParams{
            Model:     e.model,
            MaxTokens: e.maxTokens,
            System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
            Messages:  messages,
            Tools:     tools,
        })
        if err != nil {
            return nil, fmt.Errorf("API call failed at iteration %d: %w", iteration, err)
        }
        lastMsg = msg

        // Append assistant response to history
        messages = append(messages, msg.ToParam())

        // Check for tool use blocks
        var toolResults []anthropic.ContentBlockParamUnion
        for _, block := range msg.Content {
            if tu, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
                result := e.executeTool(ctx, tu.Name, tu.JSON.Input.Raw())
                toolResults = append(toolResults, anthropic.NewToolResultBlock(block.ID, result.Output, result.IsError))
                e.emitToolMetric(tu.Name, result) // observability hook
            }
        }

        // No tool calls → model is done
        if len(toolResults) == 0 {
            break
        }

        // Feed tool results back as user message
        messages = append(messages, anthropic.NewUserMessage(toolResults...))
    }
    return lastMsg, nil
}
```

**Key SDK patterns used**:
- `msg.ToParam()` — converts assistant response back to `MessageParam` for conversation history
- `anthropic.NewToolResultBlock(id, content, isError)` — creates a tool result
- Tool results are sent as a user message containing `ContentBlockParamUnion` items
- Loop breaks when model returns text without tool calls (stop_reason = `end_turn`)

**Why manual loop over BetaToolRunner**:
- `BetaToolRunner.RunToCompletion()` hides tool execution — we need per-tool observability
- Manual loop lets us inject pipeline step transitions (THINK → WRITE → LINT → COMMIT → TEST)
- Can abort mid-loop if a specific tool fails (e.g., lint fails 3 times)
- Can log every tool call and result to the output log file
- Can enforce per-step token budgets

**Alternative considered**: `BetaToolRunner` — rejected for MVP due to limited observability hooks; may reconsider if SDK adds event callbacks

### 3. Tool Definitions

**Decision**: Implement 3 core tools: `bash`, `file_read`, `file_write`

These tools replicate the essential capabilities that Goose's `developer` extension provides:

#### Tool: `bash`
```go
type BashInput struct {
    Command string `json:"command" jsonschema_description:"Shell command to execute"`
    Timeout int    `json:"timeout,omitempty" jsonschema_description:"Timeout in seconds (default 120)"`
}
```
- Executes shell commands via `exec.Command("bash", "-c", command)`
- Captures stdout + stderr
- Enforces timeout (default 120s, configurable)
- Returns combined output (truncated if >100KB)
- **Security**: Runs in the repo working directory only, inherits minimal env

#### Tool: `file_read`
```go
type FileReadInput struct {
    Path   string `json:"path" jsonschema_description:"Absolute or relative file path to read"`
    Offset int    `json:"offset,omitempty" jsonschema_description:"Line offset to start reading from"`
    Limit  int    `json:"limit,omitempty" jsonschema_description:"Maximum number of lines to read"`
}
```
- Reads file contents with optional line range
- Returns content with line numbers
- Error if file doesn't exist or is binary

#### Tool: `file_write`
```go
type FileWriteInput struct {
    Path    string `json:"path" jsonschema_description:"File path to write"`
    Content string `json:"content" jsonschema_description:"Full file content to write"`
}
```
- Creates or overwrites files
- Creates parent directories as needed
- Returns confirmation with byte count

**Why only 3 tools (not more)**:
- `bash` covers: git operations, linting, testing, building, any CLI tool
- `file_read` + `file_write` cover: all file operations the model needs
- Fewer tools = smaller prompt overhead, fewer failure modes
- Additional tools (e.g., `file_edit` for partial edits, `glob` for search) can be added later if needed

**Schema generation** using `invopop/jsonschema`:
```go
import "github.com/invopop/jsonschema"

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
    reflector := jsonschema.Reflector{
        AllowAdditionalProperties: false,
        DoNotReference:            true,
    }
    var v T
    schema := reflector.Reflect(v)
    return anthropic.ToolInputSchemaParam{
        Properties: schema.Properties,
    }
}
```

**Alternatives considered**:
- Goose's full toolkit (shell, text_editor, analyze, screen_capture, image_processor) — rejected: over-scoped for code tasks
- MCP server integration — deferred to P4: adds protocol complexity

### 4. Streaming vs Non-Streaming

**Decision**: Non-streaming for MVP, streaming for real-time output display

**MVP**: Use `client.Messages.New()` (synchronous). Simpler, easier to log, sufficient for headless/CI execution.

**Future (P2 observability)**:
```go
stream := client.Messages.NewStreaming(ctx, params)
for stream.Next() {
    event := stream.Current()
    // Write to log file in real-time
    // Emit progress events to observability platform
}
```

Event types: `ContentBlockStartEvent`, `ContentBlockDeltaEvent` (with `TextDelta`, `InputJSONDelta`), `ContentBlockStopEvent`, `MessageStartEvent`, `MessageStopEvent`.

### 5. Client Configuration

**Decision**: Map existing SpecLedger agent config directly to SDK client constructor

```go
func NewAnthropicClient(resolved *config.ResolvedConfig) (*anthropic.Client, string, error) {
    apiKey := resolved.GetValue("agent.api-key")
    if apiKey == nil {
        return nil, "", fmt.Errorf("agent.api-key not configured; run: sl config set agent.api-key <key>")
    }

    opts := []option.RequestOption{
        anthropic.WithAPIKey(apiKey.(string)),
    }

    if baseURL := resolved.GetValue("agent.base-url"); baseURL != nil {
        opts = append(opts, option.WithBaseURL(baseURL.(string)))
    }

    model := "claude-sonnet-4-20250514" // default
    if m := resolved.GetValue("agent.model"); m != nil {
        model = m.(string)
    }

    client := anthropic.NewClient(opts...)
    return client, model, nil
}
```

**No Goose env var mapping needed** — config values are used directly in Go code.

| SpecLedger Config Key | SDK Usage | Notes |
|---|---|---|
| `agent.api-key` | `anthropic.WithAPIKey()` | Required |
| `agent.base-url` | `option.WithBaseURL()` | Optional, for proxies |
| `agent.model` | `MessageNewParams.Model` | Default: claude-sonnet-4-20250514 |
| `agent.provider` | Determines SDK client type | MVP: anthropic only |

### 6. Task Selection Algorithm

**Decision**: Use existing `Store.ListReady()` with priority-based ordering (unchanged from previous plan)

**Algorithm**:
1. Call `store.ListReady(nil)` to get all unblocked open tasks
2. Sort by Priority (ascending), then by CreatedAt (ascending) for tie-breaking
3. If `--task <id>` flag provided, select that specific task (validate it's unblocked)
4. Pick the first task from sorted list

### 7. Execution Context as System + User Prompt

**Decision**: Replace instruction file generation with structured system prompt + user prompt pair

Previously, Goose received a single markdown instruction file. With the SDK, we construct:

**System prompt** (constant per run):
```
You are a software engineer executing tasks for the SpecLedger project.
You have access to bash, file_read, and file_write tools.

EXECUTION PIPELINE — follow these 5 steps IN ORDER:
1. THINK: Analyze the task, explore the codebase, plan your approach
2. WRITE CODE: Implement changes following project conventions
3. LINT: Run the lint command, fix all errors (up to 3 retries)
4. GIT COMMIT: Stage and commit with format: feat({spec}): <description>
5. TEST: Run tests, fix failures (up to 3 retries), verify DoD items

Repository: {repo_root}
Branch: {branch}
Spec: specledger/{spec_context}/
Lint command: {lint_command}
Test command: {test_command}
```

**User prompt** (per task):
```
Execute the following task:

# {title}
Task ID: {task_id}

## Description
{description}

## Acceptance Criteria
{acceptance_criteria}

## Definition of Done
{dod_checklist}

## Design Notes
{design}
```

**Advantages over instruction file**:
- System prompt is reused across tasks in the same run (fewer tokens)
- User prompt is the natural place for task-specific content
- No temporary file I/O

**Instruction file still generated** for logging/debugging: saved to `.agent-runs/<run-id>/task-<id>-prompt.md`

### 8. Task Status Management

**Decision**: Unchanged from previous plan — use existing Issue `Update()` with status transitions

**Flow**:
1. Before execution: `Update(id, {Status: "in_progress"})` — claim the task
2. On success (model completes + DoD verified): `Update(id, {Status: "closed", ClosedAt: now})`
3. On success with manual verification items: `Update(id, {Status: "needs_review"})`
4. On failure: Keep status as `in_progress`, append failure notes to `Notes` field
5. On stale detection (next run finds in_progress tasks): Offer retry/skip

**Completion criteria change**: Instead of "Goose exit code 0", completion is determined by the agentic loop finishing without error AND git commits being produced AND DoD items addressed.

### 9. Agent Run Tracking

**Decision**: Same storage layout, updated metadata to track SDK-specific fields

**New fields in AgentRun**:
- `tokens_input` (int64) — total input tokens across all API calls for this run
- `tokens_output` (int64) — total output tokens
- `api_calls` (int) — total Messages API calls
- `tool_calls` (int) — total tool invocations

**New fields in TaskResult**:
- `tokens_input` / `tokens_output` — per-task token usage
- `api_calls` / `tool_calls` — per-task counts
- `iterations` (int) — number of agentic loop iterations

These replace `exit_code` (no longer a subprocess) — though we keep a synthetic `exit_code` for backwards compatibility (0 = success, 1 = failed).

### 10. Logging and Output Capture

**Decision**: Log every API call and tool execution to per-task log files

**Log format**: Structured text with timestamps
```
[2026-03-09 14:30:22] API CALL #1 (tokens: 2450 in, 830 out)
[2026-03-09 14:30:25] TOOL CALL: bash {"command": "ls pkg/cli/agent/"}
[2026-03-09 14:30:25] TOOL RESULT: (success, 234 bytes)
[2026-03-09 14:30:25] API CALL #2 (tokens: 3280 in, 1200 out)
[2026-03-09 14:30:30] TOOL CALL: file_write {"path": "pkg/cli/agent/types.go", ...}
[2026-03-09 14:30:30] TOOL RESULT: (success, wrote 1482 bytes)
...
[2026-03-09 14:32:15] LOOP COMPLETE: 8 iterations, 12 tool calls, 45230 total tokens
```

Also capture the full model text output (thinking, explanations) interleaved with tool calls.

### 11. Git Workflow

**Decision**: Unchanged — agent commits to current feature branch. Git operations performed by the model via the `bash` tool (e.g., `git add`, `git commit`).

**Pre-execution checks**:
1. Verify clean git state (no uncommitted changes)
2. Verify current branch matches spec context
3. After each task, capture new commit hashes by comparing HEAD before/after

### 12. Graceful Stop Mechanism

**Decision**: Context cancellation instead of PID-based signals

Since the agent now runs in-process (no subprocess), graceful stop uses Go's `context.Context`:

```go
ctx, cancel := context.WithCancel(context.Background())

// sl agent stop writes a stop sentinel file
// The runner checks for this file between loop iterations
if stopRequested(runDir) {
    cancel()
}
```

**Implementation**:
- Stop sentinel: `.agent-runs/agent.stop` (presence = stop requested)
- Runner checks between API calls (between loop iterations)
- Context cancellation causes current API call to abort
- Current tool execution completes before stopping (if bash command is running)
- PID file still used for process-level kill (`sl agent stop --force`)

### 13. Token Budget and Cost Control

**Decision**: Configurable per-task token budget with soft limit

**Config keys** (new):
- `agent.max-tokens-per-task` — default 100,000 (soft limit, warns at 80%)
- `agent.max-iterations` — default 50 (hard limit on agentic loop iterations)

**Enforcement**:
- Track cumulative `input_tokens + output_tokens` per task
- At 80% of budget: inject a system message urging the model to wrap up
- At 100%: stop the loop, mark task as `needs_review` with budget note

### 14. Provider Support

**Decision**: Anthropic-only for MVP via official SDK

The existing `agent.provider` config supports `anthropic`, `bedrock`, `vertex`. For MVP:
- `anthropic` → `anthropic.NewClient()` with API key
- `bedrock` / `vertex` → error with message "Coming soon; use agent.provider=anthropic"

Future: The SDK supports Bedrock and Vertex via `anthropic.WithAWSBedrock()` and `anthropic.WithGoogleVertex()`.

### 15. Notification System (FR-016)

**Decision**: Unchanged — defer bot notification to post-MVP; log `needs_review` status prominently

---

## Migration Summary: Goose → Claude Go SDK

| Aspect | Goose (old) | Claude Go SDK (new) |
|--------|-------------|---------------------|
| Integration | `exec.Command("goose", ...)` subprocess | `anthropic.NewClient()` in-process |
| Dependency | Goose CLI binary (Rust) | Go library (`go get`) |
| Tools | Goose `developer` extension (black box) | Custom `bash`, `file_read`, `file_write` |
| Configuration | Env vars: `GOOSE_PROVIDER`, `GOOSE_MODEL`, etc. | Direct: `anthropic.WithAPIKey()`, model param |
| Observability | Parse stdout/logs after | Per-tool-call metrics in real-time |
| Pipeline control | Single instruction file | System prompt + user prompt, per-step hooks |
| Output capture | `io.MultiWriter` on subprocess stdout | Structured log of API calls + tool results |
| Stop mechanism | SIGTERM to Goose PID | `context.Cancel()` on API client |
| Token tracking | Not available | Built-in via `msg.Usage` |
| Cost | Goose abstracts away | Direct token counting and budgeting |
