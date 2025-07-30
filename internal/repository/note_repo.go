package repository

import (
	"team-service/internal/entities"

	"gorm.io/gorm"
)

type NoteRepository interface {
	Create(note *entities.Note) error
	GetByID(id uint) (*entities.Note, error)
	GetByIDWithAccess(id uint, userID string) (*entities.Note, error)
	Update(note *entities.Note) error
	Delete(id uint) error
	GetByFolderID(folderID uint) ([]entities.Note, error)
	GetByOwnerID(ownerID string) ([]entities.Note, error)
}

type noteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) NoteRepository {
	return &noteRepository{db: db}
}

func (r *noteRepository) Create(note *entities.Note) error {
	return r.db.Create(note).Error
}

func (r *noteRepository) GetByID(id uint) (*entities.Note, error) {
	var note entities.Note
	err := r.db.First(&note, id).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (r *noteRepository) GetByIDWithAccess(id uint, userID string) (*entities.Note, error) {
	var note entities.Note
	err := r.db.Where(`
		id = ? AND (
			owner_id = ? OR 
			id IN (SELECT note_id FROM note_shares WHERE user_id = ?)
		)
	`, id, userID, userID).First(&note).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (r *noteRepository) Update(note *entities.Note) error {
	return r.db.Save(note).Error
}

func (r *noteRepository) Delete(id uint) error {
	return r.db.Delete(&entities.Note{}, id).Error
}

func (r *noteRepository) GetByFolderID(folderID uint) ([]entities.Note, error) {
	var notes []entities.Note
	err := r.db.Where("folder_id = ?", folderID).Find(&notes).Error
	return notes, err
}

func (r *noteRepository) GetByOwnerID(ownerID string) ([]entities.Note, error) {
	var notes []entities.Note
	err := r.db.Where("owner_id = ?", ownerID).Find(&notes).Error
	return notes, err
}
