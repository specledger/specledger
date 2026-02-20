//go:build ignore
package main

import (
	"context"
	"log"

	"example.com/project/internal/kafka"
)

func main() {
	brokers := []string{"localhost:9092"}
	topic := "events"
	groupID := "consumer-group"

	consumer := kafka.NewConsumer(brokers, topic, groupID)
	defer consumer.Close()

	ctx := context.Background()

	log.Println("Starting consumer...")
	for {
		msg, err := consumer.Read(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		log.Printf("Received message: key=%s value=%s", string(msg.Key), string(msg.Value))
	}
}
