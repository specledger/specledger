# Batch Data Processing Template

Temporal-based batch data processing pipeline.

## Structure

```
workflows/
├── workflow.go           # Temporal workflow definitions
└── activity.go           # Temporal activity definitions

cmd/
├── worker/main.go        # Temporal worker process
└── starter/main.go       # Workflow starter CLI

internal/
├── extractors/           # Data extraction logic
├── transformers/         # Data transformation logic
└── loaders/              # Data loading logic

config/                   # Configuration files
tests/                    # Test files
```

## Technologies

- **Orchestration**: Temporal Go SDK
- **Development**: Temporal CLI for debugging
- **Persistence**: PostgreSQL (Temporal backend)

## Getting Started

1. Start Temporal: `docker-compose up -d`
2. Run worker: `go run cmd/worker/main.go`
3. Start workflow: `go run cmd/starter/main.go`
