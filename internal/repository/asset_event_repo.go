package repository

import (
	"team-service/internal/entities"

	"gorm.io/gorm"
)

type AssetEventRepository interface {
	Create(record *entities.AssetEventRecord) error
}

type assetEventRepository struct {
	db *gorm.DB
}

func NewAssetEventRepository(db *gorm.DB) AssetEventRepository {
	return &assetEventRepository{db: db}
}

func (r *assetEventRepository) Create(record *entities.AssetEventRecord) error {
	return r.db.Create(record).Error
} 