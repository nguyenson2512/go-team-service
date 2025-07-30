package usecases

import (
	"errors"
	"team-service/internal/entities"
	"team-service/internal/repository"

	"gorm.io/gorm"
)

type NoteService interface {
	CreateNote(title, body string, folderID uint, userID string) (*entities.Note, error)
	GetNote(id uint, userID string) (*entities.Note, error)
	UpdateNote(id uint, title, body, userID string) (*entities.Note, error)
	DeleteNote(id uint, userID string) error
}

type noteService struct {
	noteRepo   repository.NoteRepository
	folderRepo repository.FolderRepository
	shareRepo  repository.ShareRepository
	db         *gorm.DB
}

func NewNoteService(noteRepo repository.NoteRepository, folderRepo repository.FolderRepository, shareRepo repository.ShareRepository, db *gorm.DB) NoteService {
	return &noteService{
		noteRepo:   noteRepo,
		folderRepo: folderRepo,
		shareRepo:  shareRepo,
		db:         db,
	}
}

func (s *noteService) CreateNote(title, body string, folderID uint, userID string) (*entities.Note, error) {
	// Check if user owns the folder
	folder, err := s.folderRepo.GetByID(folderID)
	if err != nil {
		return nil, err
	}

	if folder.OwnerID != userID {
		return nil, errors.New("folder not found or access denied")
	}

	note := &entities.Note{
		Title:    title,
		Body:     body,
		FolderID: folderID,
		OwnerID:  userID,
	}

	err = s.noteRepo.Create(note)
	if err != nil {
		return nil, err
	}

	return note, nil
}

func (s *noteService) GetNote(id uint, userID string) (*entities.Note, error) {
	note, err := s.noteRepo.GetByIDWithAccess(id, userID)
	if err != nil {
		return nil, err
	}
	return note, nil
}

func (s *noteService) UpdateNote(id uint, title, body, userID string) (*entities.Note, error) {
	note, err := s.noteRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check ownership or write access
	if note.OwnerID != userID {
		share, err := s.shareRepo.GetNoteShare(note.ID, userID)
		if err != nil {
			return nil, errors.New("access denied")
		}
		if share.Access != "write" {
			return nil, errors.New("write permission required")
		}
	}

	note.Title = title
	note.Body = body

	err = s.noteRepo.Update(note)
	if err != nil {
		return nil, err
	}

	return note, nil
}

func (s *noteService) DeleteNote(id uint, userID string) error {
	note, err := s.noteRepo.GetByID(id)
	if err != nil {
		return err
	}

	if note.OwnerID != userID {
		return errors.New("only owner can delete the note")
	}

	// Use transaction to delete note and all related shares
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete note shares
		if err := tx.Where("note_id = ?", note.ID).Delete(&entities.NoteShare{}).Error; err != nil {
			return err
		}

		// Delete the note
		if err := tx.Delete(&entities.Note{}, note.ID).Error; err != nil {
			return err
		}

		return nil
	})
}
