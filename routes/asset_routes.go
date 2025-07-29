package routes

import (
	"team-service/controllers"
	"team-service/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterAssetRoutes(r *gin.Engine) {
	routes := r.Group("/")
	routes.Use(middlewares.AuthMiddleware())
	// Folder Management
	routes.POST("/folders", controllers.CreateFolder)
	routes.GET("/folders/:folderId", controllers.GetFolder)
	routes.PUT("/folders/:folderId", controllers.UpdateFolder)
	routes.DELETE("/folders/:folderId", controllers.DeleteFolder)

	// Note Management
	routes.POST("/folders/:folderId/notes", controllers.CreateNote)
	routes.GET("/notes/:noteId", controllers.GetNote)
	routes.PUT("/notes/:noteId", controllers.UpdateNote)
	routes.DELETE("/notes/:noteId", controllers.DeleteNote)

	// Sharing API
	routes.POST("/folders/:folderId/share", controllers.ShareFolder)
	routes.DELETE("/folders/:folderId/share/:userId", controllers.RevokeFolderShare)
	routes.POST("/notes/:noteId/share", controllers.ShareNote)
	routes.DELETE("/notes/:noteId/share/:userId", controllers.RevokeNoteShare)

	// Manager-only APIs
	routes.GET("/teams/:teamId/assets", controllers.GetTeamAssets)
	routes.GET("/users/:userId/assets", controllers.GetUserAssets)
}
