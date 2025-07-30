package main

import (
	"log"
	"os"

	"team-service/internal/delivery/http"
	"team-service/internal/delivery/http/handlers"
	"team-service/internal/repository"
	"team-service/internal/usecases"
	"team-service/pkg/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env file")
	}

	// Setup database
	dsn := os.Getenv("DATABASE_DSN")
	database, err := db.SetupDatabase(dsn)
	if err != nil {
		log.Fatal("Failed to setup database: ", err)
	}

	// Initialize repositories
	folderRepo := repository.NewFolderRepository(database)
	noteRepo := repository.NewNoteRepository(database)
	shareRepo := repository.NewShareRepository(database)
	teamRepo := repository.NewTeamRepository(database)

	// Initialize use cases/services
	folderService := usecases.NewFolderService(folderRepo, noteRepo, shareRepo, database)
	noteService := usecases.NewNoteService(noteRepo, folderRepo, shareRepo, database)
	shareService := usecases.NewShareService(shareRepo, folderRepo, noteRepo, teamRepo, database)
	teamService := usecases.NewTeamService(teamRepo)

	// Initialize handlers
	folderHandler := handlers.NewFolderHandler(folderService)
	noteHandler := handlers.NewNoteHandler(noteService)
	shareHandler := handlers.NewShareHandler(shareService)
	teamHandler := handlers.NewTeamHandler(teamService)

	// Initialize router
	router := http.NewRouter(folderHandler, noteHandler, shareHandler, teamHandler)

	// Setup Gin engine
	r := gin.Default()

	// Add database to context
	r.Use(func(c *gin.Context) {
		c.Set("db", database)
		c.Next()
	})

	// Setup routes
	router.SetupRoutes(r)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	log.Printf("Server starting on port %s", port)
	r.Run(port)
}
