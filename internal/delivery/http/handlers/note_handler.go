package handlers

import (
	"net/http"
	"strconv"
	"team-service/internal/usecases"
	"team-service/pkg/response"

	"github.com/gin-gonic/gin"
)

type NoteHandler struct {
	noteService usecases.NoteService
}

func NewNoteHandler(noteService usecases.NoteService) *NoteHandler {
	return &NoteHandler{
		noteService: noteService,
	}
}

type CreateNoteRequest struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

type UpdateNoteRequest struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

func (h *NoteHandler) CreateNote(c *gin.Context) {
	userID := c.GetString("userId")
	folderIDStr := c.Param("folderId")

	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid folder ID")
		return
	}

	var req CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	note, err := h.noteService.CreateNote(req.Title, req.Body, uint(folderID), userID)
	if err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, note)
}

func (h *NoteHandler) GetNote(c *gin.Context) {
	userID := c.GetString("userId")
	noteIDStr := c.Param("noteId")

	noteID, err := strconv.ParseUint(noteIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid note ID")
		return
	}

	note, err := h.noteService.GetNote(uint(noteID), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Note not found or access denied")
		return
	}

	response.Success(c, http.StatusOK, note)
}

func (h *NoteHandler) UpdateNote(c *gin.Context) {
	userID := c.GetString("userId")
	noteIDStr := c.Param("noteId")

	noteID, err := strconv.ParseUint(noteIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid note ID")
		return
	}

	var req UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	note, err := h.noteService.UpdateNote(uint(noteID), req.Title, req.Body, userID)
	if err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	response.Success(c, http.StatusOK, note)
}

func (h *NoteHandler) DeleteNote(c *gin.Context) {
	userID := c.GetString("userId")
	noteIDStr := c.Param("noteId")

	noteID, err := strconv.ParseUint(noteIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid note ID")
		return
	}

	err = h.noteService.DeleteNote(uint(noteID), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Note deleted successfully"})
}
