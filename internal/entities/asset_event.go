package entities

import "time"

// AssetEventRecord represents a persisted asset event for audit logging
type AssetEventRecord struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	EventType string    `json:"eventType" gorm:"column:eventType;index"`
	AssetType string    `json:"assetType" gorm:"column:assetType;index"`
	AssetId   string    `json:"assetId" gorm:"column:assetId;index"`
	OwnerId   string    `json:"ownerId" gorm:"column:ownerId;index"`
	ActionBy  string    `json:"actionBy" gorm:"column:actionBy;index"`
	Timestamp time.Time `json:"timestamp" gorm:"column:timestamp;index"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:createdAt;autoCreateTime"`
}

func (AssetEventRecord) TableName() string {
	return "AssetEvents"
} 