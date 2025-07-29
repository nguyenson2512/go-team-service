package models

type Note struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	FolderID uint   `json:"folderId"`
	OwnerID  string   `json:"ownerId"`
}
