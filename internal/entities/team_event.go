package entities

import "time"

// TeamEventRecord represents a persisted team event consumed from Kafka
// It is separate from kafka.TeamEvent to decouple storage concerns
// and avoid import cycles.
type TeamEventRecord struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	EventType    string    `json:"eventType" gorm:"column:eventType;index"`
	TeamId       uint      `json:"teamId" gorm:"column:teamId;index"`
	PerformedBy  string    `json:"performedBy" gorm:"column:performedBy;index"`
	TargetUserId string    `json:"targetUserId" gorm:"column:targetUserId"`
	Timestamp    time.Time `json:"timestamp" gorm:"column:timestamp;index"`
	CreatedAt    time.Time `json:"createdAt" gorm:"column:createdAt;autoCreateTime"`
}

func (TeamEventRecord) TableName() string {
	return "TeamEvents"
} 