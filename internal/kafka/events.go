package kafka

import (
	"time"
)

// TeamEventType represents the type of team event
type TeamEventType string

// AssetEventType represents the type of asset event
type AssetEventType string

const (
	TeamCreated    TeamEventType = "TEAM_CREATED"
	MemberAdded    TeamEventType = "MEMBER_ADDED"
	MemberRemoved  TeamEventType = "MEMBER_REMOVED"
	ManagerAdded   TeamEventType = "MANAGER_ADDED"
	ManagerRemoved TeamEventType = "MANAGER_REMOVED"
)

const (
	// Asset creation/update/deletion events
	FolderCreated AssetEventType = "FOLDER_CREATED"
	FolderUpdated AssetEventType = "FOLDER_UPDATED"
	FolderDeleted AssetEventType = "FOLDER_DELETED"
	NoteCreated   AssetEventType = "NOTE_CREATED"
	NoteUpdated   AssetEventType = "NOTE_UPDATED"
	NoteDeleted   AssetEventType = "NOTE_DELETED"

	// Asset sharing events
	FolderShared   AssetEventType = "FOLDER_SHARED"
	FolderUnshared AssetEventType = "FOLDER_UNSHARED"
	NoteShared     AssetEventType = "NOTE_SHARED"
	NoteUnshared   AssetEventType = "NOTE_UNSHARED"
)

// TeamEvent represents a team activity event
type TeamEvent struct {
	EventType    TeamEventType `json:"eventType"`
	TeamId       uint          `json:"teamId"`
	PerformedBy  string        `json:"performedBy"`
	TargetUserId string        `json:"targetUserId,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// AssetEvent represents an asset change event
type AssetEvent struct {
	EventType    AssetEventType `json:"eventType"`
	AssetType    string         `json:"assetType"` // "folder" or "note"
	AssetId      string         `json:"assetId"`   // Changed to string to match UUID requirement
	OwnerId      string         `json:"ownerId"`
	ActionBy     string         `json:"actionBy"`
	TargetUserId string         `json:"targetUserId,omitempty"`
	AccessType   string         `json:"accessType,omitempty"`
	Timestamp    time.Time      `json:"timestamp"`
}

// TeamEventProducer interface for producing team events
type TeamEventProducer interface {
	ProduceTeamEvent(event TeamEvent) error
	Close() error
}

// AssetEventProducer interface for producing asset events
type AssetEventProducer interface {
	ProduceAssetEvent(event AssetEvent) error
	Close() error
}
