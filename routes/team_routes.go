package routes

import (
	"team-service/controllers"
	"team-service/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterTeamRoutes(r *gin.Engine) {
	teamGroup := r.Group("/teams")
	teamGroup.Use(middlewares.AuthMiddleware())

	teamGroup.POST("", middlewares.RequireNotMember(), controllers.CreateTeam)

	protected := teamGroup.Group("/:teamId")
	protected.Use(middlewares.RequireManagerOfTeam())

	protected.POST("/members", controllers.AddMember)
	protected.DELETE("/members/:memberId", controllers.DeleteMember)
	protected.POST("/managers", controllers.AddManager)
	protected.DELETE("/managers/:managerId", controllers.DeleteManager)
}
