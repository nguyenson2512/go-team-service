package handlers

import (
	"net/http"
	"strconv"
	"team-service/internal/entities"
	"team-service/internal/usecases"
	"team-service/pkg/response"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamService usecases.TeamService
}

func NewTeamHandler(teamService usecases.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

type CreateTeamRequest struct {
	TeamName string             `json:"teamName" binding:"required"`
	Managers []entities.Manager `json:"managers"`
	Members  []entities.Member  `json:"members"`
}

type AddMemberRequest struct {
	MemberId   string `json:"memberId" binding:"required"`
	MemberName string `json:"memberName"`
}

type AddManagerRequest struct {
	ManagerId   string `json:"managerId" binding:"required"`
	ManagerName string `json:"managerName"`
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.teamService.CreateTeam(req.TeamName, req.Managers, req.Members)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create team")
		return
	}

	response.Success(c, http.StatusCreated, result)
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.teamService.AddMember(uint(teamID), req.MemberId)
	if err != nil {
		response.ErrorWithDetails(c, http.StatusInternalServerError, "Failed to add member", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"message": "Member added successfully"})
}

func (h *TeamHandler) DeleteMember(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	memberID := c.Param("memberId")

	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	err = h.teamService.DeleteMember(uint(teamID), memberID)
	if err != nil {
		response.ErrorWithDetails(c, http.StatusInternalServerError, "Failed to remove member", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Member removed successfully"})
}

func (h *TeamHandler) AddManager(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req AddManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.teamService.AddManager(uint(teamID), req.ManagerId)
	if err != nil {
		response.ErrorWithDetails(c, http.StatusInternalServerError, "Failed to add manager", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"message": "Manager added successfully"})
}

func (h *TeamHandler) DeleteManager(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	managerID := c.Param("managerId")

	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	err = h.teamService.DeleteManager(uint(teamID), managerID)
	if err != nil {
		response.ErrorWithDetails(c, http.StatusInternalServerError, "Failed to remove manager", err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Manager removed successfully"})
}
