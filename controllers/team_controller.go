package controllers

import (
	"fmt"
	"net/http"
	"team-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateTeamPayload struct {
	TeamName string           `json:"teamName" binding:"required"`
	Managers []models.Manager `json:"managers"`
	Members  []models.Member  `json:"members"`
}

func CreateTeam(c *gin.Context) {
	var payload CreateTeamPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	team := models.Team{
		TeamName: payload.TeamName,
	}

	if err := db.Create(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
		return
	}

	for _, m := range payload.Managers {
		roster := models.Roster{
			TeamId:   team.TeamId,
			UserId:   m.ManagerId,
			IsLeader: true,
		}
		db.Create(&roster)
	}

	for _, m := range payload.Members {
		roster := models.Roster{
			TeamId:   team.TeamId,
			UserId:   m.MemberId,
			IsLeader: false,
		}
		db.Create(&roster)
	}

	c.JSON(http.StatusCreated, gin.H{
		"teamId":   team.TeamId,
		"teamName": team.TeamName,
		"managers": payload.Managers,
		"members":  payload.Members,
	})
}

func DeleteMember(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	teamId := c.Param("teamId")
	memberId := c.Param("memberId")

	if err := db.Where(`"teamId" = ? AND "userId" = ? AND "isLeader" = ?`, teamId, memberId, false).Delete(&models.Roster{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

func DeleteManager(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	teamId := c.Param("teamId")
	managerId := c.Param("managerId")

	if err := db.Where(`"teamId" = ? AND "userId" = ? AND "isLeader" = ?`, teamId, managerId, true).Delete(&models.Roster{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove manager", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Manager removed successfully"})
}

type AddMemberPayload struct {
	MemberId   string `json:"memberId" binding:"required"`
	MemberName string `json:"memberName"`
}

func AddMember(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	teamId := c.Param("teamId")

	var payload AddMemberPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roster := models.Roster{
		TeamId:   parseUint(teamId),
		UserId:   payload.MemberId,
		IsLeader: false,
	}
	if err := db.Create(&roster).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}

type AddManagerPayload struct {
	ManagerId   string `json:"managerId" binding:"required"`
	ManagerName string `json:"managerName"`
}

func AddManager(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	teamId := c.Param("teamId")

	var payload AddManagerPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roster := models.Roster{
		TeamId:   parseUint(teamId),
		UserId:   payload.ManagerId,
		IsLeader: true,
	}
	if err := db.Create(&roster).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add manager", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Manager added successfully"})
}

// Helper chuyển string sang uint
func parseUint(s string) uint {
	var v uint
	fmt.Sscanf(s, "%d", &v)
	return v
}