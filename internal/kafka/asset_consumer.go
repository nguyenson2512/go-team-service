package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"team-service/internal/entities"
	"team-service/internal/repository"
	"team-service/pkg/cache"

	"github.com/segmentio/kafka-go"
)

// AssetEventConsumer represents a Kafka consumer for asset events
type AssetEventConsumer struct {
	reader             *kafka.Reader
	assetCache         cache.AssetCache
	accessControlCache cache.AccessControlCache
	eventRepo          repository.AssetEventRepository
}

// NewAssetEventConsumer creates a new Kafka consumer for asset events
func NewAssetEventConsumer(brokers []string, topic string, groupID string, assetCache cache.AssetCache, accessControlCache cache.AccessControlCache, eventRepo repository.AssetEventRepository) *AssetEventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &AssetEventConsumer{
		reader:             reader,
		assetCache:         assetCache,
		accessControlCache: accessControlCache,
		eventRepo:          eventRepo,
	}
}

// Consume starts consuming asset events from Kafka
func (c *AssetEventConsumer) Consume(ctx context.Context) {
	log.Println("Starting Kafka consumer for asset events...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka asset consumer...")
			return
		default:
			// Read message from Kafka
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading asset message from Kafka: %v", err)
				continue
			}

			// Parse the event
			var event AssetEvent
			err = json.Unmarshal(msg.Value, &event)
			if err != nil {
				log.Printf("Error unmarshaling asset event: %v", err)
				continue
			}

			log.Printf("Received asset event: %+v", event)

			// Store audit log
			if c.eventRepo != nil {
				record := &entities.AssetEventRecord{
					EventType: string(event.EventType),
					AssetType: event.AssetType,
					AssetId:   event.AssetId,
					OwnerId:   event.OwnerId,
					ActionBy:  event.ActionBy,
					Timestamp: event.Timestamp,
				}
				if err := c.eventRepo.Create(record); err != nil {
					log.Printf("Failed to store asset event audit log: %v", err)
				}
			}

			// Handle cache invalidation based on event type
			if c.assetCache != nil {
				switch event.EventType {
				case FolderDeleted, NoteDeleted:
					// Delete from cache
					if event.AssetType == "folder" {
						if assetID, err := strconv.ParseUint(event.AssetId, 10, 32); err == nil {
							if err := c.assetCache.DeleteFolder(ctx, uint(assetID)); err != nil {
								log.Printf("Failed to delete folder from cache: %v", err)
							}
						}
					} else if event.AssetType == "note" {
						if assetID, err := strconv.ParseUint(event.AssetId, 10, 32); err == nil {
							if err := c.assetCache.DeleteNote(ctx, uint(assetID)); err != nil {
								log.Printf("Failed to delete note from cache: %v", err)
							}
						}
					}
				case FolderCreated, FolderUpdated, NoteCreated, NoteUpdated:
					// These events are handled by write-through caching in the services
					// The cache is already updated when the DB operation succeeds
					log.Printf("Asset event %s handled by write-through caching", event.EventType)
				case FolderShared, NoteShared:
					// Update ACL in Redis
					if c.accessControlCache != nil && event.TargetUserId != "" && event.AccessType != "" {
						if err := c.accessControlCache.SetAssetAccess(ctx, event.AssetId, event.TargetUserId, event.AccessType); err != nil {
							log.Printf("Failed to update ACL in Redis for asset %s: %v", event.AssetId, err)
						} else {
							log.Printf("Updated ACL in Redis for asset %s, user %s, access %s", event.AssetId, event.TargetUserId, event.AccessType)
						}
					}
				case FolderUnshared, NoteUnshared:
					// Remove access from ACL in Redis
					if c.accessControlCache != nil && event.TargetUserId != "" {
						if err := c.accessControlCache.RemoveAssetAccess(ctx, event.AssetId, event.TargetUserId); err != nil {
							log.Printf("Failed to remove ACL in Redis for asset %s: %v", event.AssetId, err)
						} else {
							log.Printf("Removed ACL in Redis for asset %s, user %s", event.AssetId, event.TargetUserId)
						}
					}
				}
			}
		}
	}
}

// Close closes the asset event consumer
func (c *AssetEventConsumer) Close() error {
	return c.reader.Close()
}
