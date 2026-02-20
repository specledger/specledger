# Real-Time Data Pipeline Template

Kafka-based real-time data streaming pipeline.

## Structure

```
cmd/
├── producer/main.go      # Kafka producer service
├── consumer/main.go      # Kafka consumer service
└── processor/main.go     # Stream processor service

internal/
├── kafka/
│   ├── producer.go       # Producer implementation
│   ├── consumer.go       # Consumer implementation
│   └── config.go         # Kafka configuration
├── handlers/             # Message handlers
├── models/               # Data models
└── processors/           # Stream processing logic

configs/                  # Configuration files
deployments/              # Docker/Kubernetes manifests
tests/                    # Test files
```

## Technologies

- **Streaming**: segmentio/kafka-go (pure Go, best performance)
- **Processing**: Goka (stream processing framework)
- **Development**: Docker Compose (Kafka + Zookeeper + Schema Registry)

## Getting Started

1. Start Kafka: `docker-compose -f deployments/docker-compose.yml up -d`
2. Run producer: `go run cmd/producer/main.go`
3. Run consumer: `go run cmd/consumer/main.go`
