package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type assetEventProducer struct {
	producer *kafka.Writer
	topic    string
}

// NewAssetEventProducer creates a new Kafka producer for asset events
func NewAssetEventProducer(brokers []string, topic string) AssetEventProducer {
	// Create Kafka writer
	writer := &kafka.Writer{
		Addr:  kafka.TCP(brokers...),
		Topic: topic,
	}

	return &assetEventProducer{
		producer: writer,
		topic:    topic,
	}
}

// ProduceAssetEvent sends an asset event to Kafka
func (p *assetEventProducer) ProduceAssetEvent(event AssetEvent) error {
	// Marshal the event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal asset event: %w", err)
	}

	// Create Kafka message
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("%s-%d", event.AssetType, event.AssetId)),
		Value: eventData,
	}

	// Send message to Kafka
	err = p.producer.WriteMessages(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to write asset event message to Kafka: %w", err)
	}

	log.Printf("Successfully sent asset event to Kafka: %s", event.EventType)
	return nil
}

// Close closes the asset event producer
func (p *assetEventProducer) Close() error {
	return p.producer.Close()
} 