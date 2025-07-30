package http

import (
	"team-service/internal/delivery/http/handlers"
	"team-service/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type Router struct {
	folderHandler *handlers.FolderHandler
	noteHandler   *handlers.NoteHandler
	shareHandler  *handlers.ShareHandler
	teamHandler   *handlers.TeamHandler
}

func NewRouter(
	folderHandler *handlers.FolderHandler,
	noteHandler *handlers.NoteHandler,
	shareHandler *handlers.ShareHandler,
	teamHandler *handlers.TeamHandler,
) *Router {
	return &Router{
		folderHandler: folderHandler,
		noteHandler:   noteHandler,
		shareHandler:  shareHandler,
		teamHandler:   teamHandler,
	}
}

func (r *Router) SetupRoutes(engine *gin.Engine) {
	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Asset routes (folders, notes, sharing)
	assetRoutes := engine.Group("/")
	assetRoutes.Use(middleware.AuthMiddleware())
	{
		// Folder Management
		assetRoutes.POST("/folders", r.folderHandler.CreateFolder)
		assetRoutes.GET("/folders/:folderId", r.folderHandler.GetFolder)
		assetRoutes.PUT("/folders/:folderId", r.folderHandler.UpdateFolder)
		assetRoutes.DELETE("/folders/:folderId", r.folderHandler.DeleteFolder)

		// Note Management
		assetRoutes.POST("/folders/:folderId/notes", r.noteHandler.CreateNote)
		assetRoutes.GET("/notes/:noteId", r.noteHandler.GetNote)
		assetRoutes.PUT("/notes/:noteId", r.noteHandler.UpdateNote)
		assetRoutes.DELETE("/notes/:noteId", r.noteHandler.DeleteNote)

		// Sharing API
		assetRoutes.POST("/folders/:folderId/share", r.shareHandler.ShareFolder)
		assetRoutes.DELETE("/folders/:folderId/share/:userId", r.shareHandler.RevokeFolderShare)
		assetRoutes.POST("/notes/:noteId/share", r.shareHandler.ShareNote)
		assetRoutes.DELETE("/notes/:noteId/share/:userId", r.shareHandler.RevokeNoteShare)

		// Manager-only APIs
		assetRoutes.GET("/teams/:teamId/assets", r.shareHandler.GetTeamAssets)
		assetRoutes.GET("/users/:userId/assets", r.shareHandler.GetUserAssets)
	}

	// Team routes
	teamRoutes := engine.Group("/teams")
	teamRoutes.Use(middleware.AuthMiddleware())
	{
		teamRoutes.POST("", middleware.RequireNotMember(), r.teamHandler.CreateTeam)

		protected := teamRoutes.Group("/:teamId")
		protected.Use(middleware.RequireManagerOfTeam())
		{
			protected.POST("/members", r.teamHandler.AddMember)
			protected.DELETE("/members/:memberId", r.teamHandler.DeleteMember)
			protected.POST("/managers", r.teamHandler.AddManager)
			protected.DELETE("/managers/:managerId", r.teamHandler.DeleteManager)
		}
	}
}
