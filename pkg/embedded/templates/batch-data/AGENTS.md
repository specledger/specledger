# Batch Data Processing Template

## Overview

This template provides a Temporal-based batch data processing pipeline with ETL workflow capabilities. It's designed for scheduled data processing, batch jobs, and long-running data transformations.

## Technology Stack

- **Language**: Go 1.24+
- **Orchestration**: Temporal (workflow engine)
- **Infrastructure**: Docker Compose (Temporal + PostgreSQL)
- **Pattern**: Extract-Transform-Load (ETL)

## Directory Structure

```
.
├── workflows/
│   ├── workflow.go         # Temporal workflow definitions
│   └── activity.go         # Temporal activity definitions
├── cmd/
│   ├── worker/main.go      # Temporal worker process
│   └── starter/main.go     # Workflow starter CLI
├── internal/
│   ├── extractors/         # Data extraction logic
│   ├── transformers/       # Data transformation logic
│   └── loaders/            # Data loading logic
├── config/                 # Configuration files
├── tests/                  # Test files
├── docker-compose.yml      # Temporal + PostgreSQL setup
└── go.mod
```

## Development Commands

### Start Infrastructure
```bash
docker-compose up -d          # Start Temporal and PostgreSQL
```

### Run Worker
```bash
go run ./cmd/worker           # Start Temporal worker
```

### Trigger Workflow
```bash
go run ./cmd/starter          # Start a workflow execution
```

### Testing
```bash
go test ./...                 # Run all tests
go test ./workflows/...       # Test workflows specifically
```

## Workflow Development Guidelines

### Creating New Workflows
1. Define the workflow in `workflows/workflow.go`
2. Define activities in `workflows/activity.go`
3. Register in `cmd/worker/main.go`

### Activity Best Practices
- Activities should be idempotent
- Use retries with backoff for external calls
- Keep activities focused and single-purpose

### ETL Pattern
- **Extractors**: Fetch data from sources (APIs, databases, files)
- **Transformers**: Process and clean data
- **Loaders**: Write to destination (database, file, API)

## Code Guidelines

- Workflows must be deterministic
- Activities handle all I/O operations
- Use context for cancellation
- Log with structured logging (slog)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
