//go:build ignore
package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Producer wraps kafka-go writer
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// Send sends a message to Kafka
func (p *Producer) Send(ctx context.Context, key, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}
