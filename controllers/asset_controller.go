package controllers

import (
	"log"
	"net/http"
	"team-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateFolderPayload struct {
	Name string `json:"name" binding:"required"`
}

// Folder CRUD
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
		OwnerID: userId, // set server-side
	}

	if err := db.Create(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}
	c.JSON(http.StatusCreated, folder)


}

func GetFolder(c *gin.Context) {
	// ...implementation...
}

func UpdateFolder(c *gin.Context) {
	// ...implementation...
}

func DeleteFolder(c *gin.Context) {
	// ...implementation...
}

// Note CRUD
func CreateNote(c *gin.Context) {
	// ...implementation...
}

func GetNote(c *gin.Context) {
	// ...implementation...
}

func UpdateNote(c *gin.Context) {
	// ...implementation...
}

func DeleteNote(c *gin.Context) {
	// ...implementation...
}

// Sharing APIs
func ShareFolder(c *gin.Context) {
	// ...implementation...
}

func RevokeFolderShare(c *gin.Context) {
	// ...implementation...
}

func ShareNote(c *gin.Context) {
	// ...implementation...
}

func RevokeNoteShare(c *gin.Context) {
	// ...implementation...
}

// Manager APIs
func GetTeamAssets(c *gin.Context) {
	// ...implementation...
}

func GetUserAssets(c *gin.Context) {
	// ...implementation...
}
