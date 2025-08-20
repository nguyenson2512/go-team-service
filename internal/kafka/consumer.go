package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

// TeamEventConsumer represents a Kafka consumer for team events
type TeamEventConsumer struct {
	reader *kafka.Reader
}

// NewTeamEventConsumer creates a new Kafka consumer for team events
func NewTeamEventConsumer(brokers []string, topic string, groupID string) *TeamEventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &TeamEventConsumer{
		reader: reader,
	}
}

// Consume starts consuming team events from Kafka
func (c *TeamEventConsumer) Consume(ctx context.Context) {
	log.Println("Starting Kafka consumer for team events...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer...")
			return
		default:
			// Read message from Kafka
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message from Kafka: %v", err)
				continue
			}

			// Parse the event
			var event TeamEvent
			err = json.Unmarshal(msg.Value, &event)
			if err != nil {
				log.Printf("Error unmarshaling event: %v", err)
				continue
			}

			// Log the event (in a real implementation, this would send to ElasticSearch)
			log.Printf("Received team event: %+v", event)

			// In a real implementation, you would send this to ElasticSearch here
			// For example:
			// err = sendToElasticSearch(event)
			// if err != nil {
			//     log.Printf("Error sending event to ElasticSearch: %v", err)
			// }
		}
	}
}

// Close closes the Kafka consumer
func (c *TeamEventConsumer) Close() error {
	return c.reader.Close()
}

// sendToElasticSearch would send the event to ElasticSearch in a real implementation
// func sendToElasticSearch(event TeamEvent) error {
//     // Implementation would go here
//     return nil
// }
