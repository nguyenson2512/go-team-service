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

type FolderService interface {
	CreateFolder(name, ownerID string) (*entities.Folder, error)
	GetFolder(id uint, userID string) (*entities.Folder, error)
	UpdateFolder(id uint, name, userID string) (*entities.Folder, error)
	DeleteFolder(id uint, userID string) error
}

type folderService struct {
	folderRepo    repository.FolderRepository
	noteRepo      repository.NoteRepository
	shareRepo     repository.ShareRepository
	cache         cache.AssetCache
	accessControl cache.AccessControlCache
	assetProducer kafka.AssetEventProducer
	db            *gorm.DB
}

func NewFolderService(folderRepo repository.FolderRepository, noteRepo repository.NoteRepository, shareRepo repository.ShareRepository, cache cache.AssetCache, accessControl cache.AccessControlCache, assetProducer kafka.AssetEventProducer, db *gorm.DB) FolderService {
	return &folderService{
		folderRepo:    folderRepo,
		noteRepo:      noteRepo,
		shareRepo:     shareRepo,
		cache:         cache,
		accessControl: accessControl,
		assetProducer: assetProducer,
		db:            db,
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

	// Write-through: Update cache after successful DB write
	if s.cache != nil {
		ctx := context.Background()
		if err := s.cache.SetFolder(ctx, folder); err != nil {
			// Log cache error but don't fail the operation
			_ = err
		}
	}

	// Send asset event for folder creation
	if s.assetProducer != nil {
		event := kafka.AssetEvent{
			EventType: kafka.FolderCreated,
			AssetType: "folder",
			AssetId:   fmt.Sprintf("%d", folder.ID),
			OwnerId:   folder.OwnerID,
			ActionBy:  ownerID,
			Timestamp: time.Now(),
		}
		s.assetProducer.ProduceAssetEvent(event)
	}

	return folder, nil
}

func (s *folderService) GetFolder(id uint, userID string) (*entities.Folder, error) {
	ctx := context.Background()

	// Try cache first
	if s.cache != nil {
		cachedFolder, err := s.cache.GetFolder(ctx, id)
		if err == nil && cachedFolder != nil {
			// Check access permissions for cached data
			if cachedFolder.OwnerID == userID {
				return cachedFolder, nil
			}
			// If not owner, check if user has access via ACL in Redis first
			if s.accessControl != nil {
				accessType, err := s.accessControl.GetAssetAccess(ctx, fmt.Sprintf("%d", id), userID)
				if err == nil && accessType != "" {
					// User has access via ACL
					return cachedFolder, nil
				}
			}
			// Fallback to DB check if not in ACL
			share, err := s.shareRepo.GetFolderShare(id, userID)
			if err == nil && share != nil {
				return cachedFolder, nil
			}
		}
	}

	// Cache miss or access check failed, fallback to DB
	folder, err := s.folderRepo.GetByIDWithAccess(id, userID)
	if err != nil {
		return nil, err
	}

	// Update cache with fresh data
	if s.cache != nil {
		if err := s.cache.SetFolder(ctx, folder); err != nil {
			// Log cache error but don't fail the operation
			_ = err
		}
	}

	return folder, nil
}

func (s *folderService) UpdateFolder(id uint, name, userID string) (*entities.Folder, error) {
	folder, err := s.folderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if folder.OwnerID != userID {
		// Check if user has access via ACL in Redis first
		if s.accessControl != nil {
			ctx := context.Background()
			accessType, err := s.accessControl.GetAssetAccess(ctx, fmt.Sprintf("%d", id), userID)
			if err == nil && accessType != "" {
				// User has access via ACL
				if accessType != "write" {
					return nil, errors.New("write permission required")
				}
				// User has write access, continue with update
			} else {
				// Fallback to DB check if not in ACL
				share, err := s.shareRepo.GetFolderShare(id, userID)
				if err != nil || share == nil || share.Access != "write" {
					return nil, errors.New("not authorized or folder not found")
				}
			}
		} else {
			// Fallback to DB check if no access control cache
			share, err := s.shareRepo.GetFolderShare(id, userID)
			if err != nil || share == nil || share.Access != "write" {
				return nil, errors.New("not authorized or folder not found")
			}
		}
	}

	folder.Name = name
	err = s.folderRepo.Update(folder)
	if err != nil {
		return nil, err
	}

	// Write-through: Update cache after successful DB write
	if s.cache != nil {
		ctx := context.Background()
		if err := s.cache.SetFolder(ctx, folder); err != nil {
			// Log cache error but don't fail the operation
			_ = err
		}
	}

	// Send asset event for folder update
	if s.assetProducer != nil {
		event := kafka.AssetEvent{
			EventType: kafka.FolderUpdated,
			AssetType: "folder",
			AssetId:   fmt.Sprintf("%d", folder.ID),
			OwnerId:   folder.OwnerID,
			ActionBy:  userID,
			Timestamp: time.Now(),
		}
		s.assetProducer.ProduceAssetEvent(event)
	}

	return folder, nil
}

func (s *folderService) DeleteFolder(id uint, userID string) error {
	folder, err := s.folderRepo.GetByID(id)
	if err != nil {
		return err
	}

	if folder.OwnerID != userID {
		// Check if user has access via ACL in Redis first
		if s.accessControl != nil {
			ctx := context.Background()
			accessType, err := s.accessControl.GetAssetAccess(ctx, fmt.Sprintf("%d", id), userID)
			if err == nil && accessType != "" {
				// User has access via ACL
				if accessType != "write" {
					return errors.New("write permission required")
				}
				// User has write access, continue with delete
			} else {
				// Fallback to DB check if not in ACL
				share, err := s.shareRepo.GetFolderShare(id, userID)
				if err != nil || share == nil {
					return errors.New("not authorized or folder not found")
				}
				if share.Access != "write" {
					return errors.New("write permission required")
				}
			}
		} else {
			// Fallback to DB check if no access control cache
			share, err := s.shareRepo.GetFolderShare(id, userID)
			if err != nil || share == nil {
				return errors.New("not authorized or folder not found")
			}
			if share.Access != "write" {
				return errors.New("write permission required")
			}
		}
	}

	// Use transaction to delete folder and all related data
	err = s.db.Transaction(func(tx *gorm.DB) error {
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

	if err != nil {
		return err
	}

	// Write-through: Invalidate cache after successful DB deletion
	if s.cache != nil {
		ctx := context.Background()
		if err := s.cache.DeleteFolder(ctx, id); err != nil {
			// Log cache error but don't fail the operation
			_ = err
		}
	}

	// Send asset event for folder deletion
	if s.assetProducer != nil {
		event := kafka.AssetEvent{
			EventType: kafka.FolderDeleted,
			AssetType: "folder",
			AssetId:   fmt.Sprintf("%d", id),
			OwnerId:   folder.OwnerID,
			ActionBy:  userID,
			Timestamp: time.Now(),
		}
		s.assetProducer.ProduceAssetEvent(event)
	}

	return nil
}
