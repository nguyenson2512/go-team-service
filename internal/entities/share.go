package entities

// FolderShare represents folder sharing permissions
type FolderShare struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	FolderID uint   `json:"folderId"`
	UserID   string `json:"userId"`
	Access   string `json:"access"` // "read" or "write"
}

// NoteShare represents note sharing permissions
type NoteShare struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	NoteID uint   `json:"noteId"`
	UserID string `json:"userId"`
	Access string `json:"access"` // "read" or "write"
}
