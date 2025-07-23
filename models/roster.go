package models

type Roster struct {
	RosterId uint   `json:"rosterId" gorm:"primaryKey;autoIncrement; column:rosterId"`
	TeamId   uint   `json:"teamId" gorm:"column:teamId"`
	UserId   string `json:"userId" gorm:"column:userId"`
	IsLeader bool   `json:"isLeader" gorm:"column:isLeader"`
}

func (Roster) TableName() string {
	return "Rosters"
}