package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func ConsumeFailureEvents() {
	// Kafka config
	topic := "url_failures"
	group := "notification-group"
	broker := "hcaas_kafka:9092"

	ctx := context.Background()

	// Create Kafka reader (consumer)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       topic,
		GroupID:     group,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
		MaxAttempts: 10,
	})

	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("[ERROR] Failed to close reader: %v", err)
		}
	}()

	log.Printf("[INFO] Kafka consumer started | broker=%s | topic=%s | group=%s", broker, topic, group)

	// Loop to read messages
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("[ERROR] Failed to read message: %v", err)
			continue
		}

		log.Printf("[MESSAGE] Received at %s | partition=%d offset=%d key=%s value=%s",
			time.Now().Format(time.RFC3339),
			msg.Partition,
			msg.Offset,
			string(msg.Key),
			string(msg.Value),
		)
	}
}
