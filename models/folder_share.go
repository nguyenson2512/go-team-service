package models

type FolderShare struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	FolderID uint   `json:"folderId"`
	UserID   string   `json:"userId"`
	Access   string `json:"access"` // "read" or "write"
}
