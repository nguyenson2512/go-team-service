package entities

import "time"

// Folder represents a folder entity
type Folder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Notes     []Note    `gorm:"foreignKey:FolderID" json:"notes,omitempty"`
}
