package kafka

import (
	"time"
)

// TeamEventType represents the type of team event
type TeamEventType string

const (
	TeamCreated    TeamEventType = "TEAM_CREATED"
	MemberAdded    TeamEventType = "MEMBER_ADDED"
	MemberRemoved  TeamEventType = "MEMBER_REMOVED"
	ManagerAdded   TeamEventType = "MANAGER_ADDED"
	ManagerRemoved TeamEventType = "MANAGER_REMOVED"
)

// TeamEvent represents a team activity event
type TeamEvent struct {
	EventType    TeamEventType `json:"eventType"`
	TeamId       uint          `json:"teamId"`
	PerformedBy  string        `json:"performedBy"`
	TargetUserId string        `json:"targetUserId,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// TeamEventProducer interface for producing team events
type TeamEventProducer interface {
	ProduceTeamEvent(event TeamEvent) error
	Close() error
}
