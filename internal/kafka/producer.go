package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type teamEventProducer struct {
	producer *kafka.Writer
	topic    string
}

// NewTeamEventProducer creates a new Kafka producer for team events
func NewTeamEventProducer(brokers []string, topic string) TeamEventProducer {
	// Create Kafka writer
	writer := &kafka.Writer{
		Addr:  kafka.TCP(brokers...),
		Topic: topic,
	}

	return &teamEventProducer{
		producer: writer,
		topic:    topic,
	}
}

// ProduceTeamEvent sends a team event to Kafka
func (p *teamEventProducer) ProduceTeamEvent(event TeamEvent) error {
	// Marshal the event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("team-%d", event.TeamId)),
		Value: eventData,
	}

	// Send message to Kafka
	err = p.producer.WriteMessages(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	log.Printf("Successfully sent team event to Kafka: %s", event.EventType)
	return nil
}

// Close closes the Kafka producer
func (p *teamEventProducer) Close() error {
	return p.producer.Close()
}
