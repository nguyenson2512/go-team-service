package utils

import (
	"errors"
	"team-service/models"

	"gorm.io/gorm"
)

func IsUserManagerOfTeam(db *gorm.DB, userId string, teamId string) (bool, error) {
	var roster models.Roster
	err := db.Where(`"userId" = ? AND "teamId" = ? AND "isLeader" = TRUE`, userId, teamId).First(&roster).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func IsUserMemberOfTeam(db *gorm.DB, userId string, teamId string) (bool, error) {
	var roster models.Roster
	err := db.Where(`"userId" = ? AND "teamId" = ?`, userId, teamId).First(&roster).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
