package models

type Folder struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"ownerId"`
	Notes   []Note `gorm:"foreignKey:FolderID" json:"notes"`
}
