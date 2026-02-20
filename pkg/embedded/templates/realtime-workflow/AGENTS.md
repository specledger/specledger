# Real-Time Workflow Template

## Overview

This template provides a Temporal-based real-time workflow processing system for event-driven applications, long-running business processes, and distributed transactions.

## Technology Stack

- **Language**: Go 1.24+
- **Orchestration**: Temporal (workflow engine)
- **Infrastructure**: Docker Compose (Temporal + PostgreSQL)
- **Pattern**: Event-driven workflows

## Directory Structure

```
.
├── cmd/
│   ├── worker/main.go      # Temporal worker process
│   └── starter/            # Optional workflow trigger
├── internal/
│   ├── workflows/          # Workflow definitions
│   │   └── workflow.go
│   ├── activities/         # Activity implementations
│   │   └── activities.go
│   ├── models/             # Data models
│   └── config/             # Configuration
├── tests/                  # Test files
├── docker-compose.yml      # Temporal infrastructure
└── go.mod
```

## Development Commands

### Start Infrastructure
```bash
docker-compose up -d          # Start Temporal server
```

### Run Worker
```bash
go run ./cmd/worker           # Start workflow worker
```

### Testing
```bash
go test ./...                 # Run all tests
go test ./internal/workflows/ # Test workflows
```

## Workflow Development Guidelines

### Workflow Patterns
- **Saga Pattern**: Multi-step transactions with compensations
- **Human-in-the-Loop**: Workflows that wait for external signals
- **Scheduled Workflows**: Recurring or delayed execution
- **Long-Running**: Processes that span hours/days/weeks

### Signal and Query
```go
// Receive signals
workflow.GetSignalChannel(ctx, "approval").Receive(ctx, &approval)

// Handle queries
workflow.SetQueryHandler(ctx, "status", func() (string, error) {
    return status, nil
})
```

### Best Practices
- Workflows must be deterministic (no random, time, or I/O)
- Use activities for all external interactions
- Design for failure: handle partial completions
- Use child workflows for modularity

## Code Guidelines

- Keep workflows readable and linear
- Activities should be atomic operations
- Use workflow timers for delays
- Implement compensations for rollback scenarios

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
