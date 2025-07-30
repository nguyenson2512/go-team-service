package controllers

import (
	"net/http"
	"team-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateNotePayload struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

type UpdateNotePayload struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

// CreateNote creates a new note in a folder
func CreateNote(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userId := c.GetString("userId")
	folderIdParam := c.Param("folderId")

	var payload CreateNotePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var folder models.Folder
	if err := db.First(&folder, "id = ? AND owner_id = ?", folderIdParam, userId).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Folder not found or access denied"})
		return
	}

	note := models.Note{
		Title:    payload.Title,
		Body:     payload.Body,
		FolderID: folder.ID,
		OwnerID:  userId,
	}

	if err := db.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// GetNote retrieves a note by ID
func GetNote(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userId := c.GetString("userId")
	noteId := c.Param("noteId")

	var note models.Note

	err := db.
		Where(`
			id = ? AND (
				owner_id = ? OR 
				id IN (SELECT note_id FROM note_shares WHERE user_id = ?)
			)
		`, noteId, userId, userId).
		First(&note).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// UpdateNote updates a note's content
func UpdateNote(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userId := c.GetString("userId")
	noteId := c.Param("noteId")

	var payload UpdateNotePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var note models.Note
	if err := db.First(&note, "id = ?", noteId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Check ownership or write access
	if note.OwnerID != userId {
		var share models.NoteShare
		if err := db.First(&share, "note_id = ? AND user_id = ?", note.ID, userId).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if share.Access != "write" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Write permission required"})
			return
		}
	}

	note.Title = payload.Title
	note.Body = payload.Body

	if err := db.Save(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote deletes a note
func DeleteNote(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userId := c.GetString("userId")
	noteId := c.Param("noteId")

	var note models.Note
	if err := db.First(&note, "id = ?", noteId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if note.OwnerID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owner can delete the note"})
		return
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("note_id = ?", note.ID).Delete(&models.NoteShare{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&note).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}
