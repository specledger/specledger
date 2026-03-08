# Implementation Plan: AI Agent Task Execution Service

**Branch**: `599-agent-task-execution` | **Date**: 2026-03-03 | **Updated**: 2026-03-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specledger/599-agent-task-execution/spec.md`

## Summary

Implement `sl agent run` command that picks up tasks from `issues.jsonl`, launches an agentic loop via the Anthropic Go SDK (`anthropic-sdk-go`) with custom tools (`bash`, `file_read`, `file_write`) for each task following a 5-step execution pipeline (think → write code → lint → git commit → test), and tracks run metadata with per-tool-call observability. Builds on existing agent config (597-agent-model-config), issue store (597-issue-create-fields), and git operations.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), go-git v5 (git), gofrs/flock (file locking), gopkg.in/yaml.v3 (config), anthropic-sdk-go v1.26+ (Claude API), invopop/jsonschema (tool schema generation)
**External Dependency**: None — Claude API called directly via Go SDK (replaces Goose CLI binary)
**Storage**: File-based — JSON for agent run metadata (`.agent-runs/`), JSONL for issues (existing), log files for output capture
**Testing**: `go test` / `make test`, golangci-lint / `make lint`
**Target Platform**: macOS, Linux (CLI tool)
**Project Type**: Single project (Go CLI)
**Performance Goals**: <30s overhead per task (setup, context building, status updates) beyond LLM processing time (SC-006)
**Constraints**: Sequential task execution (MVP), no concurrent agent instances, clean git state required
**Scale/Scope**: 1-20 tasks per spec, single local agent, single repo

## Constitution Check

*Note: Constitution template is not yet customized for this project. Checking against available principles.*

- [x] **Specification-First**: Spec.md complete with 4 prioritized user stories (P1-P4), edge cases, 22 FRs
- [x] **Test-First**: Test strategy: unit tests for all new packages, integration test via `--dry-run`
- [x] **Code Quality**: golangci-lint (govet, staticcheck, errcheck, gosec), gofmt
- [x] **UX Consistency**: CLI interface documented in contracts/cli-interface.md with acceptance scenarios
- [x] **Performance**: SC-006: <30s overhead per task
- [x] **Observability**: Per-tool-call logging, token tracking, run metadata JSON, `sl agent status` and `sl agent logs` commands
- [x] **Issue Tracking**: Epic to be created via `sl issue create`

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/599-agent-task-execution/
├── plan.md              # This file
├── research.md          # Phase 0: Claude Go SDK research, decisions
├── data-model.md        # Phase 1: AgentRun, TaskResult, ExecutionContext models
├── quickstart.md        # Phase 1: Usage guide
├── contracts/
│   └── cli-interface.md # Phase 1: CLI command interface contracts
└── tasks.md             # Phase 2: Task breakdown (via /specledger.tasks)
```

### Source Code (repository root)

```text
pkg/cli/agent/                # NEW — Core agent execution package
├── types.go                  # AgentRun, TaskResult, RunStatus, ToolResult
├── tools.go                  # Tool definitions (bash, file_read, file_write) + schema generation
├── executor.go               # Agentic loop — Messages API calls, tool dispatch, iteration control
├── selector.go               # Task selection — wraps ListReady(), priority sort
├── context.go                # System prompt + user prompt construction from task metadata
├── store.go                  # RunStore — persist AgentRun JSON to .agent-runs/
├── runner.go                 # Core orchestrator — task loop, status transitions, logging
├── logger.go                 # Structured log writer for API calls and tool executions
└── *_test.go                 # Unit tests for each file

pkg/cli/commands/agent.go     # NEW — sl agent run/status/logs/stop Cobra commands

pkg/cli/config/merge.go       # MODIFIED — Add ResolveAgentConfig()
cmd/sl/main.go                # MODIFIED — Register VarAgentCmd
.gitignore                    # MODIFIED — Add .agent-runs/ pattern
go.mod                        # MODIFIED — Add anthropic-sdk-go, invopop/jsonschema
```

**Structure Decision**: Follows existing project layout. New `pkg/cli/agent/` package for agent execution logic (parallel to existing `pkg/cli/commands/`, `pkg/cli/config/`, `pkg/cli/git/`). CLI commands in `pkg/cli/commands/agent.go` (parallel to `issue.go`, `session.go`, etc.). Goose-specific files (`goose.go`) removed — replaced by `executor.go` (SDK client) and `tools.go` (custom tools).

## Implementation Order

| Phase | File | Purpose | Dependencies |
|-------|------|---------|--------------|
| 1 | `go.mod` | Add anthropic-sdk-go, invopop/jsonschema | None |
| 2 | `pkg/cli/agent/types.go` | Foundation types (AgentRun, TaskResult, RunStatus, ToolResult) | None |
| 3 | `.gitignore` | Add `.agent-runs/` | None |
| 4 | `pkg/cli/agent/tools.go` | Tool definitions + bash/file_read/file_write handlers | types.go |
| 5 | `pkg/cli/agent/logger.go` | Structured log writer | types.go |
| 6 | `pkg/cli/agent/executor.go` | Agentic loop (Messages API + tool dispatch) | types.go, tools.go, logger.go |
| 7 | `pkg/cli/config/merge.go` | Add ResolveAgentConfig() | None |
| 8 | `pkg/cli/agent/store.go` | RunStore persistence | types.go |
| 9 | `pkg/cli/agent/selector.go` | Task selection logic | types.go, pkg/issues |
| 10 | `pkg/cli/agent/context.go` | System + user prompt construction | types.go, pkg/issues |
| 11 | `pkg/cli/agent/runner.go` | Core orchestrator | All above |
| 12 | `pkg/cli/commands/agent.go` | CLI commands | runner.go |
| 13 | `cmd/sl/main.go` | Register VarAgentCmd | commands/agent.go |

## Key Design: Agentic Loop (executor.go)

The executor implements the manual agentic loop pattern from the Anthropic Go SDK:

```go
// pkg/cli/agent/executor.go

type Executor struct {
    client        *anthropic.Client
    model         anthropic.Model
    maxTokens     int64
    maxIterations int
    tokenBudget   int64
    tools         []anthropic.ToolUnionParam
    toolHandlers  map[string]ToolHandler
    logger        *Logger
}

type ToolHandler func(ctx context.Context, input json.RawMessage) ToolResult

type ToolResult struct {
    Output  string
    IsError bool
}

func (e *Executor) Run(ctx context.Context, systemPrompt string, userPrompt string) (*ExecutionResult, error) {
    messages := []anthropic.MessageParam{
        anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
    }

    result := &ExecutionResult{}

    for iteration := 0; iteration < e.maxIterations; iteration++ {
        msg, err := e.client.Messages.New(ctx, anthropic.MessageNewParams{
            Model:     e.model,
            MaxTokens: e.maxTokens,
            System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
            Messages:  messages,
            Tools:     e.tools,
        })
        if err != nil {
            return result, fmt.Errorf("API call failed at iteration %d: %w", iteration, err)
        }

        // Track token usage
        result.TokensInput += msg.Usage.InputTokens
        result.TokensOutput += msg.Usage.OutputTokens
        result.APICalls++

        e.logger.LogAPICall(iteration, msg.Usage)

        // Append assistant response to history
        messages = append(messages, msg.ToParam())

        // Process tool use blocks
        var toolResults []anthropic.ContentBlockParamUnion
        for _, block := range msg.Content {
            switch v := block.AsAny().(type) {
            case anthropic.TextBlock:
                e.logger.LogText(v.Text)
            case anthropic.ToolUseBlock:
                result.ToolCalls++
                handler, ok := e.toolHandlers[v.Name]
                if !ok {
                    toolResults = append(toolResults,
                        anthropic.NewToolResultBlock(block.ID, "unknown tool: "+v.Name, true))
                    continue
                }
                tr := handler(ctx, json.RawMessage(v.JSON.Input.Raw()))
                e.logger.LogToolCall(v.Name, v.JSON.Input.Raw(), tr)
                toolResults = append(toolResults,
                    anthropic.NewToolResultBlock(block.ID, tr.Output, tr.IsError))
            }
        }

        // No tool calls → model is done
        if len(toolResults) == 0 {
            break
        }

        // Token budget check
        totalTokens := result.TokensInput + result.TokensOutput
        if e.tokenBudget > 0 && totalTokens > e.tokenBudget {
            e.logger.LogBudgetExceeded(totalTokens, e.tokenBudget)
            break
        }

        // Feed tool results back
        messages = append(messages, anthropic.NewUserMessage(toolResults...))
    }

    result.Iterations = len(messages) / 2 // approximate
    return result, nil
}
```

## Key Design: Tool Definitions (tools.go)

```go
// pkg/cli/agent/tools.go

func BuildTools(workDir string) ([]anthropic.ToolUnionParam, map[string]ToolHandler) {
    tools := []anthropic.ToolUnionParam{
        {OfTool: &anthropic.ToolParam{
            Name:        "bash",
            Description: anthropic.String("Execute a shell command in the repository directory"),
            InputSchema: GenerateSchema[BashInput](),
        }},
        {OfTool: &anthropic.ToolParam{
            Name:        "file_read",
            Description: anthropic.String("Read the contents of a file"),
            InputSchema: GenerateSchema[FileReadInput](),
        }},
        {OfTool: &anthropic.ToolParam{
            Name:        "file_write",
            Description: anthropic.String("Write content to a file, creating it if necessary"),
            InputSchema: GenerateSchema[FileWriteInput](),
        }},
    }

    handlers := map[string]ToolHandler{
        "bash":       NewBashHandler(workDir),
        "file_read":  NewFileReadHandler(workDir),
        "file_write": NewFileWriteHandler(workDir),
    }

    return tools, handlers
}

type BashInput struct {
    Command string `json:"command" jsonschema:"required" jsonschema_description:"Shell command to execute"`
    Timeout int    `json:"timeout,omitempty" jsonschema_description:"Timeout in seconds (default 120)"`
}

type FileReadInput struct {
    Path   string `json:"path" jsonschema:"required" jsonschema_description:"File path to read (absolute or relative to repo root)"`
    Offset int    `json:"offset,omitempty" jsonschema_description:"Line number to start reading from"`
    Limit  int    `json:"limit,omitempty" jsonschema_description:"Maximum number of lines to read"`
}

type FileWriteInput struct {
    Path    string `json:"path" jsonschema:"required" jsonschema_description:"File path to write"`
    Content string `json:"content" jsonschema:"required" jsonschema_description:"Complete file content to write"`
}
```

## Key Design: 5-Step Pipeline System Prompt

The 5-step pipeline is encoded in the system prompt, not as separate API calls:

```go
// pkg/cli/agent/context.go

func BuildSystemPrompt(cfg *ContextConfig) string {
    return fmt.Sprintf(`You are a software engineer executing tasks for a SpecLedger project.
You have access to bash, file_read, and file_write tools.

EXECUTION PIPELINE — follow these 5 steps IN ORDER. Complete each step fully before proceeding.

## Step 1: THINK (Analysis & Planning)
- Read and understand the task description, acceptance criteria, and definition of done
- Use file_read and bash to explore the relevant codebase files
- Plan your approach: which files to create/modify, edge cases to handle
- DO NOT write any code yet

## Step 2: WRITE CODE (Implementation)
- Implement the changes as planned
- Follow project conventions: gofmt formatting, existing patterns
- Address ALL acceptance criteria and Definition of Done items

## Step 3: LINT (Code Quality)
- Run: %s
- Fix ALL lint errors
- Re-run lint to verify (up to 3 retries)
- Do NOT proceed until the linter passes cleanly

## Step 4: GIT COMMIT
- Stage your changes: git add <files>
- Commit with message format: feat(%s): <description>
- Commit to the CURRENT branch (%s) only — do NOT create new branches

## Step 5: TEST (Verification)
- Run: %s
- If tests fail: fix the issue → re-lint → re-commit → re-test (up to 3 retries)
- Verify each Definition of Done item is addressed

## Repository Context
- Working directory: %s
- Current branch: %s
- Spec directory: specledger/%s/
`, cfg.LintCommand, cfg.SpecContext, cfg.Branch, cfg.TestCommand, cfg.RepoRoot, cfg.Branch, cfg.SpecContext)
}

func BuildUserPrompt(task *issues.Issue) string {
    // Render task metadata as user prompt
    // Title, description, acceptance criteria, DoD checklist, design notes
}
```

## Key Design: Client Configuration

```go
// pkg/cli/agent/executor.go

func NewExecutor(resolved *config.ResolvedConfig, logger *Logger) (*Executor, error) {
    apiKey := resolved.GetValue("agent.api-key")
    if apiKey == nil {
        return nil, fmt.Errorf("agent.api-key not configured; run: sl config set agent.api-key <key>")
    }

    opts := []option.RequestOption{
        anthropic.WithAPIKey(apiKey.(string)),
    }
    if baseURL := resolved.GetValue("agent.base-url"); baseURL != nil {
        opts = append(opts, option.WithBaseURL(baseURL.(string)))
    }

    model := anthropic.ModelClaudeSonnet4_20250514 // default
    if m := resolved.GetValue("agent.model"); m != nil {
        model = anthropic.Model(m.(string))
    }

    maxIterations := 50 // default
    tokenBudget := int64(100000) // default

    client := anthropic.NewClient(opts...)
    tools, handlers := BuildTools(repoRoot)

    return &Executor{
        client:        client,
        model:         model,
        maxTokens:     4096,
        maxIterations: maxIterations,
        tokenBudget:   tokenBudget,
        tools:         tools,
        toolHandlers:  handlers,
        logger:        logger,
    }, nil
}
```

## Pipeline Failure Handling

| Scenario | Behavior |
|----------|----------|
| Lint fails after retries | Model documents issues, commits what it has, runner marks task `failed` |
| Tests fail after retries | Model documents failures, runner marks task `failed`, keeps issue `in_progress` |
| API error (rate limit, network) | Runner retries with exponential backoff (3 attempts), then marks task `failed` |
| Token budget exceeded | Loop stops, task marked `needs_review` with budget note |
| Max iterations reached | Loop stops, task marked `needs_review` with iteration note |
| Context cancellation (stop) | Current tool completes, loop exits, task stays `in_progress` |
| No tasks available | Exit code 2 |
| Prerequisites not met (no API key, dirty git) | Exit code 3 |

## Previous Work (from issue tracker)

### 597-agent-model-config (28 tasks, all closed)
- Agent config schema, merge logic, launcher env injection, profiles, CLI commands
- **Reused**: `AgentConfig`, `ResolvedConfig`, config keys (`agent.api-key`, `agent.model`, `agent.provider`, `agent.base-url`)

### 597-issue-create-fields (24 tasks, all closed)
- Issue model with DoD, dependencies, parent-child, JSONL store with locking
- **Reused**: `Store.ListReady()`, `Issue.IsReady()`, `DefinitionOfDone`, `IssueUpdate`
