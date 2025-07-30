package repository

import (
	"team-service/internal/entities"

	"gorm.io/gorm"
)

type TeamRepository interface {
	Create(team *entities.Team) error
	GetByID(id uint) (*entities.Team, error)
	Update(team *entities.Team) error
	Delete(id uint) error

	// Roster operations
	CreateRoster(roster *entities.Roster) error
	DeleteRoster(teamID uint, userID string, isLeader bool) error
	GetRosterByTeamAndUser(teamID uint, userID string) (*entities.Roster, error)
	GetTeamMembers(teamID uint) ([]entities.Roster, error)
	IsUserManagerOfTeam(userID string, teamID uint) (bool, error)
	IsUserMemberOfTeam(userID string, teamID uint) (bool, error)
	GetUsersByTeamID(teamID uint) ([]string, error)
}

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(team *entities.Team) error {
	return r.db.Create(team).Error
}

func (r *teamRepository) GetByID(id uint) (*entities.Team, error) {
	var team entities.Team
	err := r.db.First(&team, id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) Update(team *entities.Team) error {
	return r.db.Save(team).Error
}

func (r *teamRepository) Delete(id uint) error {
	return r.db.Delete(&entities.Team{}, id).Error
}

// Roster operations
func (r *teamRepository) CreateRoster(roster *entities.Roster) error {
	return r.db.Create(roster).Error
}

func (r *teamRepository) DeleteRoster(teamID uint, userID string, isLeader bool) error {
	return r.db.Where(`"teamId" = ? AND "userId" = ? AND "isLeader" = ?`, teamID, userID, isLeader).Delete(&entities.Roster{}).Error
}

func (r *teamRepository) GetRosterByTeamAndUser(teamID uint, userID string) (*entities.Roster, error) {
	var roster entities.Roster
	err := r.db.Where(`"teamId" = ? AND "userId" = ?`, teamID, userID).First(&roster).Error
	if err != nil {
		return nil, err
	}
	return &roster, nil
}

func (r *teamRepository) GetTeamMembers(teamID uint) ([]entities.Roster, error) {
	var rosters []entities.Roster
	err := r.db.Where(`"teamId" = ?`, teamID).Find(&rosters).Error
	return rosters, err
}

func (r *teamRepository) IsUserManagerOfTeam(userID string, teamID uint) (bool, error) {
	var count int64
	err := r.db.Model(&entities.Roster{}).Where(`"userId" = ? AND "teamId" = ? AND "isLeader" = TRUE`, userID, teamID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *teamRepository) IsUserMemberOfTeam(userID string, teamID uint) (bool, error) {
	var count int64
	err := r.db.Model(&entities.Roster{}).Where(`"userId" = ? AND "teamId" = ?`, userID, teamID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *teamRepository) GetUsersByTeamID(teamID uint) ([]string, error) {
	var userIds []string
	err := r.db.Table("rosters").Where("team_id = ?", teamID).Pluck("user_id", &userIds).Error
	return userIds, err
}
