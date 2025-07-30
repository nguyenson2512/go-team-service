package controllers

import (
	"net/http"
	"team-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateFolderPayload struct {
	Name string `json:"name" binding:"required"`
}

type UpdateFolderPayload struct {
	Name string `json:"name" binding:"required"`
}

// CreateFolder creates a new folder
func CreateFolder(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userId := c.GetString("userId")
	var payload CreateFolderPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder := models.Folder{
		Name:    payload.Name,
		OwnerID: userId,
	}

	if err := db.Create(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}
	c.JSON(http.StatusCreated, folder)
}

// GetFolder retrieves a folder by ID
func GetFolder(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userId := c.GetString("userId")
	folderId := c.Param("folderId")

	var folder models.Folder

	err := db.Preload("Notes").
		Where(`
			id = ? AND (
				owner_id = ? OR 
				id IN (SELECT folder_id FROM folder_shares WHERE user_id = ?)
			)
		`, folderId, userId, userId).
		First(&folder).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, folder)
}

// UpdateFolder updates a folder's information
func UpdateFolder(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userId := c.GetString("userId")
	folderId := c.Param("folderId")

	var payload UpdateFolderPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var folder models.Folder
	if err := db.First(&folder, "id = ? AND owner_id = ?", folderId, userId).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized or folder not found"})
		return
	}

	folder.Name = payload.Name
	if err := db.Save(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update folder"})
		return
	}

	c.JSON(http.StatusOK, folder)
}

// DeleteFolder deletes a folder and all its contents
func DeleteFolder(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userId := c.GetString("userId")
	folderId := c.Param("folderId")

	var folder models.Folder
	if err := db.First(&folder, "id = ? AND owner_id = ?", folderId, userId).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized or folder not found"})
		return
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			DELETE FROM note_shares 
			WHERE note_id IN (SELECT id FROM notes WHERE folder_id = ?)
		`, folder.ID).Error; err != nil {
			return err
		}

		if err := tx.Where("folder_id = ?", folder.ID).Delete(&models.Note{}).Error; err != nil {
			return err
		}

		if err := tx.Where("folder_id = ?", folder.ID).Delete(&models.FolderShare{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&folder).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder and notes: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder and its notes deleted successfully"})
}
