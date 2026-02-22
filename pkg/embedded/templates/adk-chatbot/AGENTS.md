# AI Chatbot Template (Google ADK)

## Overview

This template provides a Go-based AI chatbot using Google's Agent Development Kit (ADK). It uses Gemini models with custom tools, session management, and supports both CLI and web interfaces out of the box via the ADK launcher.

## Technology Stack

- **Language**: Go 1.24+
- **Framework**: Google Agent Development Kit (google.golang.org/adk v0.5.0)
- **LLM**: Google Gemini (via google.golang.org/genai)
- **Session**: In-memory (ADK session service)
- **Deployment**: CLI, Web UI, or Google Cloud Run

## Directory Structure

```
.
├── cmd/
│   ├── chatbot/main.go        # Full launcher (CLI + web + A2A)
│   └── server/main.go         # Programmatic runner example
├── internal/
│   ├── agents/
│   │   └── chatbot.go         # Agent definitions
│   ├── tools/
│   │   └── tools.go           # Custom function tools
│   └── middleware/             # Request/response middleware
├── configs/
│   └── example.env.sample     # Environment configuration
├── tests/                     # Test files
└── go.mod
```

## Development Commands

### Run CLI (Interactive Console)

```bash
export GOOGLE_API_KEY=your-key
go run ./cmd/chatbot
```

### Run Web UI

```bash
go run ./cmd/chatbot web api webui
# Opens at http://localhost:8080
```

### Run Programmatic Example

```bash
go run ./cmd/server
```

### Testing

```bash
go test ./...
```

## Agent Development Guidelines

### Creating Agents

1. Define agent in `internal/agents/`
2. Use `llmagent.New()` with a `Config` struct
3. Set `Name`, `Model`, `Description`, `Instruction`, and `Tools`

### Adding Custom Tools

```go
type MyArgs struct {
    Input string `json:"input"` // description for LLM
}
type MyResult struct {
    Output string `json:"output"` // description for LLM
}

func myHandler(ctx tool.Context, input MyArgs) (MyResult, error) {
    return MyResult{Output: "result"}, nil
}

myTool, err := functiontool.New(functiontool.Config{
    Name:        "my_tool",
    Description: "What this tool does.",
}, myHandler)
```

### Multi-Agent Workflows

ADK supports composing agents hierarchically:

- **Sequential**: `sequentialagent.New()` runs sub-agents in order
- **Parallel**: `parallelagent.New()` runs sub-agents concurrently
- **Loop**: `loopagent.New()` repeats until exit condition

### Using Google Search

```go
import "google.golang.org/adk/tool/geminitool"

tools := []tool.Tool{geminitool.GoogleSearch{}}
```

### MCP Tool Integration

```go
import "google.golang.org/adk/tool/mcptoolset"

mcpTools, err := mcptoolset.New(ctx, mcptoolset.Config{...})
```

## Code Guidelines

- Keep agents focused on a single responsibility
- Define tool input/output as typed structs with JSON tags
- Use structured logging (slog) for observability
- Handle errors explicitly in tool handlers
- Test tools independently from agent logic

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
