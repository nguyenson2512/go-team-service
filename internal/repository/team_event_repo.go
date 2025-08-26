package repository

import (
	"team-service/internal/entities"

	"gorm.io/gorm"
)

type TeamEventRepository interface {
	Create(record *entities.TeamEventRecord) error
}

type teamEventRepository struct {
	db *gorm.DB
}

func NewTeamEventRepository(db *gorm.DB) TeamEventRepository {
	return &teamEventRepository{db: db}
}

func (r *teamEventRepository) Create(record *entities.TeamEventRecord) error {
	return r.db.Create(record).Error
} 