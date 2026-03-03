# Implementation Plan: AI Agent Task Execution Service

**Branch**: `599-agent-task-execution` | **Date**: 2026-03-03 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specledger/599-agent-task-execution/spec.md`

## Summary

Implement `sl agent run` command that picks up tasks from `issues.jsonl`, launches a Goose AI agent for each task with a 5-step execution pipeline (think → write code → lint → git commit → test), and tracks run metadata. Builds on existing agent config (597-agent-model-config), issue store (597-issue-create-fields), and git operations.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), go-git v5 (git), gofrs/flock (file locking), gopkg.in/yaml.v3 (config), text/template (instruction rendering)
**External Dependency**: Goose CLI (Block/Square, Rust binary, invoked via `exec.Command`)
**Storage**: File-based — JSON for agent run metadata (`.agent-runs/`), JSONL for issues (existing), log files for output capture
**Testing**: `go test` / `make test`, golangci-lint / `make lint`
**Target Platform**: macOS, Linux (CLI tool)
**Project Type**: Single project (Go CLI)
**Performance Goals**: <30s overhead per task (setup, context building, status updates) beyond Goose processing time (SC-006)
**Constraints**: Sequential task execution (MVP), no concurrent agent instances, clean git state required
**Scale/Scope**: 1-20 tasks per spec, single local agent, single repo

## Constitution Check

*Note: Constitution template is not yet customized for this project. Checking against available principles.*

- [x] **Specification-First**: Spec.md complete with 4 prioritized user stories (P1-P4), edge cases, 17 FRs
- [x] **Test-First**: Test strategy: unit tests for all new packages, integration test via `--dry-run`
- [x] **Code Quality**: golangci-lint (govet, staticcheck, errcheck, gosec), gofmt
- [x] **UX Consistency**: CLI interface documented in contracts/cli-interface.md with acceptance scenarios
- [x] **Performance**: SC-006: <30s overhead per task
- [x] **Observability**: Per-task log capture, run metadata JSON, `sl agent status` and `sl agent logs` commands
- [x] **Issue Tracking**: Epic to be created via `sl issue create`

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/599-agent-task-execution/
├── plan.md              # This file
├── research.md          # Phase 0: Goose CLI research, decisions
├── data-model.md        # Phase 1: AgentRun, TaskResult, ExecutionContext models
├── quickstart.md        # Phase 1: Usage guide
├── contracts/
│   └── cli-interface.md # Phase 1: CLI command interface contracts
└── tasks.md             # Phase 2: Task breakdown (via /specledger.tasks)
```

### Source Code (repository root)

```text
pkg/cli/agent/                # NEW — Core agent execution package
├── types.go                  # AgentRun, TaskResult, RunStatus, PipelineConfig
├── selector.go               # Task selection — wraps ListReady(), priority sort
├── context.go                # ExecutionContext builder, instruction template renderer
├── goose.go                  # Goose subprocess adapter — invoke, env mapping, output capture
├── store.go                  # RunStore — persist AgentRun JSON to .agent-runs/
├── runner.go                 # Core orchestrator — task loop, status transitions
└── *_test.go                 # Unit tests for each file

pkg/cli/commands/agent.go     # NEW — sl agent run/status/logs/stop Cobra commands

pkg/cli/launcher/launcher.go  # MODIFIED — Add Goose to DefaultAgents
pkg/cli/git/git.go            # MODIFIED — Add GetHeadHash(), GetCommitsBetween()
pkg/cli/config/merge.go       # MODIFIED — Add ResolveAgentConfig()
cmd/sl/main.go                # MODIFIED — Register VarAgentCmd
.gitignore                    # MODIFIED — Add .agent-runs/ pattern
```

**Structure Decision**: Follows existing project layout. New `pkg/cli/agent/` package for agent execution logic (parallel to existing `pkg/cli/commands/`, `pkg/cli/config/`, `pkg/cli/git/`). CLI commands in `pkg/cli/commands/agent.go` (parallel to `issue.go`, `session.go`, etc.).

## Implementation Order

| Phase | File | Purpose | Dependencies |
|-------|------|---------|--------------|
| 1 | `pkg/cli/agent/types.go` | Foundation types (AgentRun, TaskResult, RunStatus) | None |
| 2 | `.gitignore` | Add `.agent-runs/` | None |
| 2 | `pkg/cli/launcher/launcher.go` | Add Goose to DefaultAgents | None |
| 3 | `pkg/cli/git/git.go` | Add GetHeadHash(), GetCommitsBetween() | None |
| 4 | `pkg/cli/config/merge.go` | Add ResolveAgentConfig() | None |
| 5 | `pkg/cli/agent/store.go` | RunStore persistence | types.go |
| 6 | `pkg/cli/agent/selector.go` | Task selection logic | types.go, pkg/issues |
| 7 | `pkg/cli/agent/context.go` | Instruction template rendering | types.go, pkg/issues |
| 8 | `pkg/cli/agent/goose.go` | Goose subprocess adapter | types.go, config/merge.go |
| 9 | `pkg/cli/agent/runner.go` | Core orchestrator | All above |
| 10 | `pkg/cli/commands/agent.go` | CLI commands | runner.go |
| 11 | `cmd/sl/main.go` | Register VarAgentCmd | commands/agent.go |

## Key Design: 5-Step Pipeline Instruction Template

The 5-step pipeline is encoded as explicit directives in the instruction markdown generated by `context.go`:

```markdown
# Task: {{.Title}}
# Task ID: {{.TaskID}} | Spec: {{.SpecContext}} | Branch: {{.Branch}}

## Task Description
{{.Description}}

## Acceptance Criteria
{{.AcceptanceCriteria}}

## Definition of Done
{{range .DoDItems}}- [ ] {{.Item}}
{{end}}

## Design Notes
{{.Design}}

# EXECUTION PIPELINE
You MUST follow these 5 steps IN ORDER. Complete each step fully before proceeding.

## Step 1: THINK (Analysis & Planning)
- Read and understand the task description, acceptance criteria, and definition of done above
- Explore the relevant codebase files
- Plan your approach: which files to create/modify, edge cases to handle
- DO NOT write any code yet

## Step 2: WRITE CODE (Implementation)
- Implement the changes as planned
- Follow project conventions: gofmt formatting, existing patterns
- Address ALL acceptance criteria and Definition of Done items

## Step 3: LINT (Code Quality)
- Run: {{.LintCommand}}
- Fix ALL lint errors
- Re-run lint to verify (up to 3 retries)
- Do NOT proceed until the linter passes cleanly

## Step 4: GIT COMMIT
- Stage your changes: git add <files>
- Commit with message format: feat({{.SpecContext}}): <description>
- Commit to the CURRENT branch ({{.Branch}}) only — do NOT create new branches

## Step 5: TEST (Verification)
- Run: {{.TestCommand}}
- If tests fail: fix the issue → re-lint → re-commit → re-test (up to 3 retries)
- Verify each Definition of Done item is addressed

## Repository Context
- Working directory: {{.RepoRoot}}
- Current branch: {{.Branch}}
- Spec directory: specledger/{{.SpecContext}}/
```

## Key Design: Goose Environment Mapping

```go
func BuildGooseEnv(resolved *config.ResolvedConfig) map[string]string {
    env := map[string]string{
        "GOOSE_MODE":                    "auto",
        "GOOSE_DISABLE_SESSION_NAMING":  "true",
    }
    if v := resolved.GetValue("agent.provider"); v != nil {
        env["GOOSE_PROVIDER"] = v.(string)
    }
    if v := resolved.GetValue("agent.model"); v != nil {
        env["GOOSE_MODEL"] = v.(string)
    }
    if v := resolved.GetValue("agent.api-key"); v != nil {
        key := v.(string)
        env["GOOSE_PROVIDER__API_KEY"] = key
        // Also set provider-specific key for compatibility
        provider, _ := resolved.GetValue("agent.provider").(string)
        switch provider {
        case "anthropic":
            env["ANTHROPIC_API_KEY"] = key
        case "openai":
            env["OPENAI_API_KEY"] = key
        }
    }
    if v := resolved.GetValue("agent.base-url"); v != nil {
        env["GOOSE_PROVIDER__HOST"] = v.(string)
    }
    // Pass through custom env vars
    if v := resolved.GetValue("agent.env"); v != nil {
        if envMap, ok := v.(map[string]string); ok {
            for k, val := range envMap {
                env[k] = val
            }
        }
    }
    return env
}
```

## Key Design: Goose Invocation

```go
args := []string{
    "run",
    "-i", instructionPath,
    "--no-session",
    "--with-builtin", "developer",
    "--max-turns", strconv.Itoa(maxTurns), // default: 50
}
if headless {
    args = append(args, "-q") // quiet mode
}
```

Output captured via `cmd.Stdout = io.MultiWriter(os.Stdout, logFile)`.

## Pipeline Failure Handling

| Scenario | Behavior |
|----------|----------|
| Lint fails after retries | Goose documents issues, commits what it has, runner marks task `failed` |
| Tests fail after retries | Goose documents failures, runner marks task `failed`, keeps issue `in_progress` |
| Goose crash/timeout | Runner catches non-zero exit, marks task `failed`, moves to next task |
| No tasks available | Exit code 2 |
| Prerequisites not met | Exit code 3 (no Goose, dirty git, not on feature branch) |

## Previous Work (from issue tracker)

### 597-agent-model-config (28 tasks, all closed)
- Agent config schema, merge logic, launcher env injection, profiles, CLI commands
- **Reused**: `AgentConfig`, `ResolvedConfig`, `AgentLauncher`, `ResolveAgentEnv()`

### 597-issue-create-fields (24 tasks, all closed)
- Issue model with DoD, dependencies, parent-child, JSONL store with locking
- **Reused**: `Store.ListReady()`, `Issue.IsReady()`, `DefinitionOfDone`, `IssueUpdate`
