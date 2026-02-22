# AI Chatbot Template (Google ADK)

Go-based AI chatbot using Google Agent Development Kit.

## Structure

```
cmd/
├── chatbot/main.go       # CLI + Web UI launcher (ADK full launcher)
└── server/main.go        # Programmatic runner example

internal/
├── agents/               # Agent definitions
│   └── chatbot.go
├── tools/                # Custom function tools
│   └── tools.go
└── middleware/            # Request/response middleware

configs/                  # Configuration files
tests/                    # Test files
```

## Technologies

- **Framework**: Google Agent Development Kit (google.golang.org/adk)
- **LLM**: Google Gemini (gemini-2.5-flash)
- **Language**: Go 1.24+

## Getting Started

1. Set API key: `export GOOGLE_API_KEY=your-key`
2. Install deps: `go mod tidy`
3. Run CLI: `go run ./cmd/chatbot`
4. Run Web UI: `go run ./cmd/chatbot web api webui`
