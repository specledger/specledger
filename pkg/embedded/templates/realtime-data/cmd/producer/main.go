//go:build ignore
package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"example.com/project/internal/kafka"
)

func main() {
	brokers := []string{"localhost:9092"}
	topic := "events"

	producer := kafka.NewProducer(brokers, topic)
	defer producer.Close()

	ctx := context.Background()

	// Example: Send messages in a loop
	for i := 0; ; i++ {
		event := map[string]interface{}{
			"id":        i,
			"timestamp": time.Now().Unix(),
			"data":      "sample event",
		}

		value, err := json.Marshal(event)
		if err != nil {
			log.Printf("Error marshaling event: %v", err)
			continue
		}

		if err := producer.Send(ctx, nil, value); err != nil {
			log.Printf("Error sending message: %v", err)
		} else {
			log.Printf("Sent message %d", i)
		}

		time.Sleep(time.Second)
	}
}
