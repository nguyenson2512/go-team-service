package entities

import "time"

// Team represents a team entity
type Team struct {
	TeamId    uint      `json:"teamId" gorm:"primaryKey;autoIncrement;column:teamId"`
	TeamName  string    `json:"teamName" gorm:"column:teamName"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:createdAt;autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updatedAt;autoUpdateTime"`
}

func (Team) TableName() string {
	return "Teams"
}

// Manager represents a manager entity
type Manager struct {
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	ManagerId   string `json:"managerId"`
	ManagerName string `json:"managerName"`
	TeamId      uint   `json:"teamId"`
}

// Member represents a member entity
type Member struct {
	ID         uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	MemberId   string `json:"memberId"`
	MemberName string `json:"memberName"`
	TeamId     uint   `json:"teamId"`
}

// Roster represents team membership
type Roster struct {
	RosterId uint   `json:"rosterId" gorm:"primaryKey;autoIncrement;column:rosterId"`
	TeamId   uint   `json:"teamId" gorm:"column:teamId"`
	UserId   string `json:"userId" gorm:"column:userId"`
	IsLeader bool   `json:"isLeader" gorm:"column:isLeader"`
}

func (Roster) TableName() string {
	return "Rosters"
}
