# Real-Time Workflow Template

Temporal-based real-time workflow processing.

## Structure

```
cmd/
├── worker/main.go        # Temporal worker process
└── starter/main.go       # Workflow starter CLI

internal/
├── workflows/            # Workflow definitions
├── activities/           # Activity implementations
├── models/               # Data models
└── config/               # Configuration

tests/                    # Test files
```

## Technologies

- **Framework**: Temporal Go SDK
- **Development**: Docker Compose (Temporal + PostgreSQL + Elasticsearch)

## Getting Started

1. Start Temporal: `docker-compose up -d`
2. Run worker: `go run cmd/worker/main.go`
3. Trigger workflow: `go run cmd/starter/main.go`
