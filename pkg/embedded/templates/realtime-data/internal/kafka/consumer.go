//go:build ignore
package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Consumer wraps kafka-go reader
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
	}
}

// Read reads the next message from Kafka
func (c *Consumer) Read(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}
