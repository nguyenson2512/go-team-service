package repository

import (
	"team-service/internal/entities"

	"gorm.io/gorm"
)

type ShareRepository interface {
	// Folder sharing
	CreateFolderShare(share *entities.FolderShare) error
	UpdateFolderShare(share *entities.FolderShare) error
	DeleteFolderShare(folderID uint, userID string) error
	GetFolderShare(folderID uint, userID string) (*entities.FolderShare, error)
	GetFolderShares(folderID uint) ([]entities.FolderShare, error)

	// Note sharing
	CreateNoteShare(share *entities.NoteShare) error
	UpdateNoteShare(share *entities.NoteShare) error
	DeleteNoteShare(noteID uint, userID string) error
	GetNoteShare(noteID uint, userID string) (*entities.NoteShare, error)
	GetNoteShares(noteID uint) ([]entities.NoteShare, error)

	// Bulk operations
	DeleteNoteSharesByNoteID(noteID uint) error
	DeleteFolderSharesByFolderID(folderID uint) error
}

type shareRepository struct {
	db *gorm.DB
}

func NewShareRepository(db *gorm.DB) ShareRepository {
	return &shareRepository{db: db}
}

// Folder sharing methods
func (r *shareRepository) CreateFolderShare(share *entities.FolderShare) error {
	return r.db.Create(share).Error
}

func (r *shareRepository) UpdateFolderShare(share *entities.FolderShare) error {
	return r.db.Save(share).Error
}

func (r *shareRepository) DeleteFolderShare(folderID uint, userID string) error {
	return r.db.Where("folder_id = ? AND user_id = ?", folderID, userID).Delete(&entities.FolderShare{}).Error
}

func (r *shareRepository) GetFolderShare(folderID uint, userID string) (*entities.FolderShare, error) {
	var share entities.FolderShare
	err := r.db.Where("folder_id = ? AND user_id = ?", folderID, userID).First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *shareRepository) GetFolderShares(folderID uint) ([]entities.FolderShare, error) {
	var shares []entities.FolderShare
	err := r.db.Where("folder_id = ?", folderID).Find(&shares).Error
	return shares, err
}

// Note sharing methods
func (r *shareRepository) CreateNoteShare(share *entities.NoteShare) error {
	return r.db.Create(share).Error
}

func (r *shareRepository) UpdateNoteShare(share *entities.NoteShare) error {
	return r.db.Save(share).Error
}

func (r *shareRepository) DeleteNoteShare(noteID uint, userID string) error {
	return r.db.Where("note_id = ? AND user_id = ?", noteID, userID).Delete(&entities.NoteShare{}).Error
}

func (r *shareRepository) GetNoteShare(noteID uint, userID string) (*entities.NoteShare, error) {
	var share entities.NoteShare
	err := r.db.Where("note_id = ? AND user_id = ?", noteID, userID).First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *shareRepository) GetNoteShares(noteID uint) ([]entities.NoteShare, error) {
	var shares []entities.NoteShare
	err := r.db.Where("note_id = ?", noteID).Find(&shares).Error
	return shares, err
}

// Bulk operations
func (r *shareRepository) DeleteNoteSharesByNoteID(noteID uint) error {
	return r.db.Where("note_id = ?", noteID).Delete(&entities.NoteShare{}).Error
}

func (r *shareRepository) DeleteFolderSharesByFolderID(folderID uint) error {
	return r.db.Where("folder_id = ?", folderID).Delete(&entities.FolderShare{}).Error
}
