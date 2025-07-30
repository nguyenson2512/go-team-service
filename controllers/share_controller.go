package controllers

import (
	"errors"
	"net/http"
	"team-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ShareFolderPayload struct {
	UserID string `json:"userId" binding:"required"`
	Access string `json:"access" binding:"required,oneof=read write"`
}

type ShareNotePayload struct {
	UserID string `json:"userId" binding:"required"`
	Access string `json:"access" binding:"required,oneof=read write"`
}

// ShareFolder shares a folder with another user
func ShareFolder(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userId := c.GetString("userId")
	folderId := c.Param("folderId")

	var payload ShareFolderPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if payload.UserID == userId {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot share folder with yourself"})
		return
	}

	var folder models.Folder
	if err := db.First(&folder, "id = ?", folderId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}
	if folder.OwnerID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can share this folder"})
		return
	}

	// Transaction: upsert folder share + upsert note shares
	err := db.Transaction(func(tx *gorm.DB) error {
		var folderShare models.FolderShare
		if err := tx.Where("folder_id = ? AND user_id = ?", folder.ID, payload.UserID).
			First(&folderShare).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&models.FolderShare{
					FolderID: folder.ID,
					UserID:   payload.UserID,
					Access:   payload.Access,
				}).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			folderShare.Access = payload.Access
			if err := tx.Save(&folderShare).Error; err != nil {
				return err
			}
		}

		var notes []models.Note
		if err := tx.Where("folder_id = ?", folder.ID).Find(&notes).Error; err != nil {
			return err
		}

		for _, note := range notes {
			var noteShare models.NoteShare
			err := tx.Where("note_id = ? AND user_id = ?", note.ID, payload.UserID).
				First(&noteShare).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					noteShare = models.NoteShare{
						NoteID: note.ID,
						UserID: payload.UserID,
						Access: payload.Access,
					}
					if err := tx.Create(&noteShare).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				noteShare.Access = payload.Access
				if err := tx.Save(&noteShare).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share folder: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder shared successfully"})
}

// RevokeFolderShare revokes folder access from a user
func RevokeFolderShare(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	requesterId := c.GetString("userId")

	folderId := c.Param("folderId")
	targetUserId := c.Param("userId")

	var folder models.Folder
	if err := db.First(&folder, "id = ?", folderId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}
	if folder.OwnerID != requesterId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the folder owner can revoke access"})
		return
	}

	if err := db.Where("folder_id = ? AND user_id = ?", folder.ID, targetUserId).
		Delete(&models.FolderShare{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke folder share"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder access revoked"})
}

// ShareNote shares a note with another user
func ShareNote(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userId := c.GetString("userId")
	noteId := c.Param("noteId")

	var payload ShareNotePayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var note models.Note
	if err := db.First(&note, "id = ?", noteId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if note.OwnerID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owner can share the note"})
		return
	}

	var existing models.NoteShare
	err := db.Where("note_id = ? AND user_id = ?", note.ID, payload.UserID).First(&existing).Error
	if err == nil {
		existing.Access = payload.Access
		if err := db.Save(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update share"})
			return
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		newShare := models.NoteShare{
			NoteID: note.ID,
			UserID: payload.UserID,
			Access: payload.Access,
		}
		if err := db.Create(&newShare).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note share"})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note shared successfully"})
}

// RevokeNoteShare revokes note access from a user
func RevokeNoteShare(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userId := c.GetString("userId")

	noteId := c.Param("noteId")
	targetUserId := c.Param("userId")

	var note models.Note
	if err := db.First(&note, "id = ?", noteId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}
	if note.OwnerID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owner can revoke access"})
		return
	}

	if err := db.Where("note_id = ? AND user_id = ?", note.ID, targetUserId).
		Delete(&models.NoteShare{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Access revoked"})
}

// GetTeamAssets retrieves all assets for a team (manager API)
func GetTeamAssets(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	teamId := c.Param("teamId")

	var userIds []string
	if err := db.
		Table("rosters").
		Where("team_id = ?", teamId).
		Pluck("user_id", &userIds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch team members"})
		return
	}

	if len(userIds) == 0 {
		c.JSON(http.StatusOK, gin.H{"ownedFolders": []models.Folder{}, "sharedFolders": []models.Folder{}, "ownedNotes": []models.Note{}, "sharedNotes": []models.Note{}})
		return
	}

	var ownedFolders []models.Folder
	var ownedNotes []models.Note

	db.Where("owner_id IN ?", userIds).Find(&ownedFolders)
	db.Where("owner_id IN ?", userIds).Find(&ownedNotes)

	var sharedFolders []models.Folder
	var sharedNotes []models.Note

	db.
		Model(&models.Folder{}).
		Joins("JOIN folder_shares fs ON fs.folder_id = folders.id").
		Where("fs.user_id IN ?", userIds).
		Select("folders.*, fs.access").
		Find(&sharedFolders)

	db.
		Model(&models.Note{}).
		Joins("JOIN note_shares ns ON ns.note_id = notes.id").
		Where("ns.user_id IN ?", userIds).
		Select("notes.*, ns.access").
		Find(&sharedNotes)

	c.JSON(http.StatusOK, gin.H{
		"ownedFolders":  ownedFolders,
		"sharedFolders": sharedFolders,
		"ownedNotes":    ownedNotes,
		"sharedNotes":   sharedNotes,
	})
}

// GetUserAssets retrieves all assets for a specific user (manager API)
func GetUserAssets(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	targetUserId := c.Param("userId")

	var ownedFolders []models.Folder
	var sharedFolders []models.Folder
	var ownedNotes []models.Note
	var sharedNotes []models.Note

	// Folders owned by user
	db.Where("owner_id = ?", targetUserId).Find(&ownedFolders)

	// Notes owned by user
	db.Where("owner_id = ?", targetUserId).Find(&ownedNotes)

	// Folders shared to user
	db.
		Model(&models.Folder{}).
		Joins("JOIN folder_shares ON folders.id = folder_shares.folder_id").
		Where("folder_shares.user_id = ?", targetUserId).
		Select("folders.*, folder_shares.access").
		Find(&sharedFolders)

	// Notes shared to user
	db.
		Model(&models.Note{}).
		Joins("JOIN note_shares ON notes.id = note_shares.note_id").
		Where("note_shares.user_id = ?", targetUserId).
		Select("notes.*, note_shares.access").
		Find(&sharedNotes)

	// Response
	c.JSON(http.StatusOK, gin.H{
		"ownedFolders":  ownedFolders,
		"sharedFolders": sharedFolders,
		"ownedNotes":    ownedNotes,
		"sharedNotes":   sharedNotes,
	})
}
