package usecases

import (
	"context"
	"errors"
	"fmt"
	"team-service/internal/entities"
	"team-service/internal/kafka"
	"team-service/internal/repository"
	"team-service/pkg/cache"
	"time"

	"gorm.io/gorm"
)

type NoteService interface {
	CreateNote(title, body string, folderID uint, userID string) (*entities.Note, error)
	GetNote(id uint, userID string) (*entities.Note, error)
	UpdateNote(id uint, title, body, userID string) (*entities.Note, error)
	DeleteNote(id uint, userID string) error
}

type noteService struct {
	noteRepo      repository.NoteRepository
	folderRepo    repository.FolderRepository
	shareRepo     repository.ShareRepository
	cache         cache.AssetCache
	assetProducer kafka.AssetEventProducer
	db            *gorm.DB
}

func NewNoteService(noteRepo repository.NoteRepository, folderRepo repository.FolderRepository, shareRepo repository.ShareRepository, cache cache.AssetCache, assetProducer kafka.AssetEventProducer, db *gorm.DB) NoteService {
	return &noteService{
		noteRepo:      noteRepo,
		folderRepo:    folderRepo,
		shareRepo:     shareRepo,
		cache:         cache,
		assetProducer: assetProducer,
		db:            db,
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

	// Write-through: Update cache after successful DB write
	if s.cache != nil {
		ctx := context.Background()
		if err := s.cache.SetNote(ctx, note); err != nil {
			// Log cache error but don't fail the operation
			_ = err
		}
	}

	// Send asset event for note creation
	if s.assetProducer != nil {
		event := kafka.AssetEvent{
			EventType: kafka.NoteCreated,
			AssetType: "note",
			AssetId:   fmt.Sprintf("%d", note.ID),
			OwnerId:   note.OwnerID,
			ActionBy:  userID,
			Timestamp: time.Now(),
		}
		s.assetProducer.ProduceAssetEvent(event)
	}

	return note, nil
}

func (s *noteService) GetNote(id uint, userID string) (*entities.Note, error) {
	ctx := context.Background()
	
	// Try cache first
	if s.cache != nil {
		cachedNote, err := s.cache.GetNote(ctx, id)
		if err == nil && cachedNote != nil {
			// Check access permissions for cached data
			if cachedNote.OwnerID == userID {
				return cachedNote, nil
			}
			// If not owner, check if user has access via shares
			share, err := s.shareRepo.GetNoteShare(id, userID)
			if err == nil && share != nil {
				return cachedNote, nil
			}
		}
	}

	// Cache miss or access check failed, fallback to DB
	note, err := s.noteRepo.GetByIDWithAccess(id, userID)
	if err != nil {
		return nil, err
	}

	// Update cache with fresh data
	if s.cache != nil {
		if err := s.cache.SetNote(ctx, note); err != nil {
			// Log cache error but don't fail the operation
			_ = err
		}
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

	// Write-through: Update cache after successful DB write
	if s.cache != nil {
		ctx := context.Background()
		if err := s.cache.SetNote(ctx, note); err != nil {
			// Log cache error but don't fail the operation
			_ = err
		}
	}

	// Send asset event for note update
	if s.assetProducer != nil {
		event := kafka.AssetEvent{
			EventType: kafka.NoteUpdated,
			AssetType: "note",
			AssetId:   fmt.Sprintf("%d", note.ID),
			OwnerId:   note.OwnerID,
			ActionBy:  userID,
			Timestamp: time.Now(),
		}
		s.assetProducer.ProduceAssetEvent(event)
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
	err = s.db.Transaction(func(tx *gorm.DB) error {
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

	if err != nil {
		return err
	}

	// Write-through: Invalidate cache after successful DB deletion
	if s.cache != nil {
		ctx := context.Background()
		if err := s.cache.DeleteNote(ctx, id); err != nil {
			// Log cache error but don't fail the operation
			_ = err
		}
	}

	// Send asset event for note deletion
	if s.assetProducer != nil {
		event := kafka.AssetEvent{
			EventType: kafka.NoteDeleted,
			AssetType: "note",
			AssetId:   fmt.Sprintf("%d", id),
			OwnerId:   note.OwnerID,
			ActionBy:  userID,
			Timestamp: time.Now(),
		}
		s.assetProducer.ProduceAssetEvent(event)
	}

	return nil
}
