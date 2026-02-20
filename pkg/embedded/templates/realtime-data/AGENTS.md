# Real-Time Data Pipeline Template

## Overview

This template provides a Kafka-based real-time data streaming pipeline for event processing, data ingestion, and stream analytics. It includes producer and consumer patterns for building scalable data pipelines.

## Technology Stack

- **Language**: Go 1.24+
- **Messaging**: Apache Kafka
- **Kafka Client**: segmentio/kafka-go
- **Infrastructure**: Docker Compose (Kafka + Zookeeper)

## Directory Structure

```
.
├── cmd/
│   ├── producer/           # Data producer service
│   │   └── main.go
│   └── consumer/           # Data consumer service
│       └── main.go
├── internal/
│   ├── kafka/
│   │   ├── producer.go     # Kafka producer implementation
│   │   ├── consumer.go     # Kafka consumer implementation
│   │   └── config.go       # Connection configuration
│   ├── models/             # Message schemas
│   ├── processors/         # Message processing logic
│   └── handlers/           # Event handlers
├── tests/                  # Test files
├── docker-compose.yml      # Kafka infrastructure
└── go.mod
```

## Development Commands

### Start Infrastructure
```bash
docker-compose up -d          # Start Kafka and Zookeeper
```

### Run Producer
```bash
go run ./cmd/producer         # Start producing messages
```

### Run Consumer
```bash
go run ./cmd/consumer         # Start consuming messages
```

### Testing
```bash
go test ./...                 # Run all tests
```

## Kafka Development Guidelines

### Producer Best Practices
- Use async producers for high throughput
- Implement idempotent producers to prevent duplicates
- Batch messages when possible
- Handle retries with exponential backoff

### Consumer Best Practices
- Use consumer groups for scalability
- Commit offsets after successful processing
- Handle rebalances gracefully
- Implement dead letter queues for failed messages

### Message Design
- Use Avro/Protobuf for schema evolution
- Include message timestamps
- Add correlation IDs for tracing

## Code Guidelines

- Process messages idempotently
- Log with structured logging (slog)
- Monitor consumer lag
- Test with embedded Kafka or testcontainers

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
