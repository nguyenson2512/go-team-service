package repository

import (
	"team-service/internal/entities"

	"gorm.io/gorm"
)

type FolderRepository interface {
	Create(folder *entities.Folder) error
	GetByID(id uint) (*entities.Folder, error)
	GetByIDWithAccess(id uint, userID string) (*entities.Folder, error)
	Update(folder *entities.Folder) error
	Delete(id uint) error
	GetByOwnerID(ownerID string) ([]entities.Folder, error)
}

type folderRepository struct {
	db *gorm.DB
}

func NewFolderRepository(db *gorm.DB) FolderRepository {
	return &folderRepository{db: db}
}

func (r *folderRepository) Create(folder *entities.Folder) error {
	return r.db.Create(folder).Error
}

func (r *folderRepository) GetByID(id uint) (*entities.Folder, error) {
	var folder entities.Folder
	err := r.db.Preload("Notes").First(&folder, id).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *folderRepository) GetByIDWithAccess(id uint, userID string) (*entities.Folder, error) {
	var folder entities.Folder
	err := r.db.Preload("Notes").
		Where(`
			id = ? AND (
				owner_id = ? OR 
				id IN (SELECT folder_id FROM folder_shares WHERE user_id = ?)
			)
		`, id, userID, userID).
		First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *folderRepository) Update(folder *entities.Folder) error {
	return r.db.Save(folder).Error
}

func (r *folderRepository) Delete(id uint) error {
	return r.db.Delete(&entities.Folder{}, id).Error
}

func (r *folderRepository) GetByOwnerID(ownerID string) ([]entities.Folder, error) {
	var folders []entities.Folder
	err := r.db.Where("owner_id = ?", ownerID).Find(&folders).Error
	return folders, err
}
