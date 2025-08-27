package kafka

import (
	"context"
	"encoding/json"
	"log"
	"team-service/internal/entities"
	"team-service/internal/repository"
	"team-service/pkg/cache"

	"github.com/segmentio/kafka-go"
)

// TeamEventConsumer represents a Kafka consumer for team events
type TeamEventConsumer struct {
	reader    *kafka.Reader
	eventRepo repository.TeamEventRepository
	teamCache cache.TeamCache
}

// NewTeamEventConsumer creates a new Kafka consumer for team events
func NewTeamEventConsumer(brokers []string, topic string, groupID string, eventRepo repository.TeamEventRepository, teamCache cache.TeamCache) *TeamEventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &TeamEventConsumer{
		reader:    reader,
		eventRepo: eventRepo,
		teamCache: teamCache,
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

			log.Printf("Received team event: %+v", event)

			record := &entities.TeamEventRecord{
				EventType:    string(event.EventType),
				TeamId:       event.TeamId,
				PerformedBy:  event.PerformedBy,
				TargetUserId: event.TargetUserId,
				Timestamp:    event.Timestamp,
			}
			if err := c.eventRepo.Create(record); err != nil {
				log.Printf("Failed to persist team event: %v", err)
			}

			// Update cache for member changes
			switch event.EventType {
			case MemberAdded:
				if c.teamCache != nil && event.TargetUserId != "" {
					if err := c.teamCache.AddTeamMember(ctx, event.TeamId, event.TargetUserId); err != nil {
						log.Printf("Failed to add member to cache: %v", err)
					}
				}
			case MemberRemoved:
				if c.teamCache != nil && event.TargetUserId != "" {
					if err := c.teamCache.RemoveTeamMember(ctx, event.TeamId, event.TargetUserId); err != nil {
						log.Printf("Failed to remove member from cache: %v", err)
					}
				}
			}
		}
	}
}

// Close closes the Kafka consumer
func (c *TeamEventConsumer) Close() error {
	return c.reader.Close()
}
