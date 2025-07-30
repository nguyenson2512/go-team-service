package usecases

import (
	"errors"
	"team-service/internal/entities"
	"team-service/internal/repository"

	"gorm.io/gorm"
)

type FolderService interface {
	CreateFolder(name, ownerID string) (*entities.Folder, error)
	GetFolder(id uint, userID string) (*entities.Folder, error)
	UpdateFolder(id uint, name, userID string) (*entities.Folder, error)
	DeleteFolder(id uint, userID string) error
}

type folderService struct {
	folderRepo repository.FolderRepository
	noteRepo   repository.NoteRepository
	shareRepo  repository.ShareRepository
	db         *gorm.DB
}

func NewFolderService(folderRepo repository.FolderRepository, noteRepo repository.NoteRepository, shareRepo repository.ShareRepository, db *gorm.DB) FolderService {
	return &folderService{
		folderRepo: folderRepo,
		noteRepo:   noteRepo,
		shareRepo:  shareRepo,
		db:         db,
	}
}

func (s *folderService) CreateFolder(name, ownerID string) (*entities.Folder, error) {
	folder := &entities.Folder{
		Name:    name,
		OwnerID: ownerID,
	}

	err := s.folderRepo.Create(folder)
	if err != nil {
		return nil, err
	}

	return folder, nil
}

func (s *folderService) GetFolder(id uint, userID string) (*entities.Folder, error) {
	folder, err := s.folderRepo.GetByIDWithAccess(id, userID)
	if err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *folderService) UpdateFolder(id uint, name, userID string) (*entities.Folder, error) {
	folder, err := s.folderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if folder.OwnerID != userID {
		return nil, errors.New("not authorized or folder not found")
	}

	folder.Name = name
	err = s.folderRepo.Update(folder)
	if err != nil {
		return nil, err
	}

	return folder, nil
}

func (s *folderService) DeleteFolder(id uint, userID string) error {
	folder, err := s.folderRepo.GetByID(id)
	if err != nil {
		return err
	}

	if folder.OwnerID != userID {
		return errors.New("not authorized or folder not found")
	}

	// Use transaction to delete folder and all related data
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete note shares for all notes in this folder
		if err := tx.Exec(`
			DELETE FROM note_shares 
			WHERE note_id IN (SELECT id FROM notes WHERE folder_id = ?)
		`, folder.ID).Error; err != nil {
			return err
		}

		// Delete all notes in the folder
		if err := tx.Where("folder_id = ?", folder.ID).Delete(&entities.Note{}).Error; err != nil {
			return err
		}

		// Delete folder shares
		if err := tx.Where("folder_id = ?", folder.ID).Delete(&entities.FolderShare{}).Error; err != nil {
			return err
		}

		// Delete the folder
		if err := tx.Delete(&entities.Folder{}, folder.ID).Error; err != nil {
			return err
		}

		return nil
	})
}
