package models

type NoteShare struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	NoteID uint   `json:"noteId"`
	UserID string   `json:"userId"`
	Access string `json:"access"` // "read" or "write"
}
