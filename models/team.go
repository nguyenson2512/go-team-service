package models

import "time"

// Manager struct
type Manager struct {
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	ManagerId   string `json:"managerId"`
	ManagerName string `json:"managerName"`
	TeamId      uint   `json:"teamId"`
}

// Member struct
type Member struct {
	ID         uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	MemberId   string `json:"memberId"`
	MemberName string `json:"memberName"`
	TeamId     uint   `json:"teamId"`
}

// Team struct
type Team struct {
	TeamId   uint      `json:"teamId" gorm:"primaryKey;autoIncrement;column:teamId"`
	TeamName string    `json:"teamName" gorm:"column:teamName"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:createdAt;autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updatedAt;autoUpdateTime"`
}

func (Team) TableName() string {
	return "Teams"
}