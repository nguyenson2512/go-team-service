package handlers

import (
	"net/http"
	"strconv"
	"team-service/internal/usecases"
	"team-service/pkg/response"
	"team-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

type FolderHandler struct {
	folderService usecases.FolderService
}

func NewFolderHandler(folderService usecases.FolderService) *FolderHandler {
	return &FolderHandler{
		folderService: folderService,
	}
}

type CreateFolderRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateFolderRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *FolderHandler) CreateFolder(c *gin.Context) {
	userID := c.GetString("userId")
	var req CreateFolderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	folder, err := h.folderService.CreateFolder(req.Name, userID)
	if err != nil {
		logger.Logger.Error().Msg("Failed to create folder")
		response.Error(c, http.StatusInternalServerError, "Failed to create folder")
		return
	}

	response.Success(c, http.StatusCreated, folder)
}

func (h *FolderHandler) GetFolder(c *gin.Context) {
	userID := c.GetString("userId")
	folderIDStr := c.Param("folderId")

	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid folder ID")
		return
	}

	folder, err := h.folderService.GetFolder(uint(folderID), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Folder not found or access denied")
		return
	}

	response.Success(c, http.StatusOK, folder)
}

func (h *FolderHandler) UpdateFolder(c *gin.Context) {
	userID := c.GetString("userId")
	folderIDStr := c.Param("folderId")

	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid folder ID")
		return
	}

	var req UpdateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	folder, err := h.folderService.UpdateFolder(uint(folderID), req.Name, userID)
	if err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	response.Success(c, http.StatusOK, folder)
}

func (h *FolderHandler) DeleteFolder(c *gin.Context) {
	userID := c.GetString("userId")
	folderIDStr := c.Param("folderId")

	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid folder ID")
		return
	}

	err = h.folderService.DeleteFolder(uint(folderID), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Folder and its notes deleted successfully"})
}
