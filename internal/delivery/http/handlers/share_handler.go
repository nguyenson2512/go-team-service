package handlers

import (
	"net/http"
	"strconv"
	"team-service/internal/usecases"
	"team-service/pkg/response"

	"github.com/gin-gonic/gin"
)

type ShareHandler struct {
	shareService usecases.ShareService
}

func NewShareHandler(shareService usecases.ShareService) *ShareHandler {
	return &ShareHandler{
		shareService: shareService,
	}
}

type ShareFolderRequest struct {
	UserID string `json:"userId" binding:"required"`
	Access string `json:"access" binding:"required,oneof=read write"`
}

type ShareNoteRequest struct {
	UserID string `json:"userId" binding:"required"`
	Access string `json:"access" binding:"required,oneof=read write"`
}

func (h *ShareHandler) ShareFolder(c *gin.Context) {
	userID := c.GetString("userId")
	folderIDStr := c.Param("folderId")

	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid folder ID")
		return
	}

	var req ShareFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.shareService.ShareFolder(uint(folderID), req.UserID, req.Access, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Folder shared successfully"})
}

func (h *ShareHandler) RevokeFolderShare(c *gin.Context) {
	userID := c.GetString("userId")
	folderIDStr := c.Param("folderId")
	targetUserID := c.Param("userId")

	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid folder ID")
		return
	}

	err = h.shareService.RevokeFolderShare(uint(folderID), targetUserID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Folder access revoked"})
}

func (h *ShareHandler) ShareNote(c *gin.Context) {
	userID := c.GetString("userId")
	noteIDStr := c.Param("noteId")

	noteID, err := strconv.ParseUint(noteIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid note ID")
		return
	}

	var req ShareNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.shareService.ShareNote(uint(noteID), req.UserID, req.Access, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Note shared successfully"})
}

func (h *ShareHandler) RevokeNoteShare(c *gin.Context) {
	userID := c.GetString("userId")
	noteIDStr := c.Param("noteId")
	targetUserID := c.Param("userId")

	noteID, err := strconv.ParseUint(noteIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid note ID")
		return
	}

	err = h.shareService.RevokeNoteShare(uint(noteID), targetUserID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Access revoked"})
}

func (h *ShareHandler) GetTeamAssets(c *gin.Context) {
	teamIDStr := c.Param("teamId")

	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	assets, err := h.shareService.GetTeamAssets(uint(teamID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, assets)
}

func (h *ShareHandler) GetUserAssets(c *gin.Context) {
	targetUserID := c.Param("userId")

	assets, err := h.shareService.GetUserAssets(targetUserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, assets)
}
