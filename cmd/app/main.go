package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"team-service/internal/delivery/http"
	"team-service/internal/delivery/http/handlers"
	kafka "team-service/internal/kafka"
	"team-service/internal/repository"
	"team-service/internal/usecases"
	"team-service/pkg/cache"
	"team-service/pkg/db"
	"team-service/pkg/logger"

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

	logger.SetupLogger()

	// Initialize Kafka producer
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092" // Default Kafka broker
	}
	kafkaBrokersList := strings.Split(kafkaBrokers, ",")

	kafkaTopic := os.Getenv("KAFKA_TEAM_ACTIVITY_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "team.activity" // Default topic
	}

	assetTopic := os.Getenv("KAFKA_ASSET_ACTIVITY_TOPIC")
	if assetTopic == "" {
		assetTopic = "asset.changes" // Default topic
	}

	kafkaProducer := kafka.NewTeamEventProducer(kafkaBrokersList, kafkaTopic)
	defer kafkaProducer.Close()

	assetProducer := kafka.NewAssetEventProducer(kafkaBrokersList, assetTopic)
	defer assetProducer.Close()

	// Initialize Redis cache
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	teamCache := cache.NewRedisTeamCache(redisAddr, "", 0)
	assetCache := cache.NewRedisAssetCache(redisAddr, "", 0)

	// Initialize repositories
	folderRepo := repository.NewFolderRepository(database)
	noteRepo := repository.NewNoteRepository(database)
	shareRepo := repository.NewShareRepository(database)
	teamRepo := repository.NewTeamRepository(database)
	eventRepo := repository.NewTeamEventRepository(database)
	assetEventRepo := repository.NewAssetEventRepository(database)

	// Initialize Kafka consumer
	kafkaConsumer := kafka.NewTeamEventConsumer(kafkaBrokersList, kafkaTopic, "team-service-consumer", eventRepo, teamCache)
	defer kafkaConsumer.Close()

	assetConsumer := kafka.NewAssetEventConsumer(kafkaBrokersList, assetTopic, "asset-service-consumer", assetCache, assetEventRepo)
	defer assetConsumer.Close()

	// Start consumers in goroutines
	go kafkaConsumer.Consume(context.Background())
	go assetConsumer.Consume(context.Background())

	// Initialize use cases/services
	folderService := usecases.NewFolderService(folderRepo, noteRepo, shareRepo, assetCache, assetProducer, database)
	noteService := usecases.NewNoteService(noteRepo, folderRepo, shareRepo, assetCache, assetProducer, database)
	shareService := usecases.NewShareService(shareRepo, folderRepo, noteRepo, teamRepo, teamCache, assetProducer, database)
	teamService := usecases.NewTeamService(teamRepo, kafkaProducer)

	// Initialize handlers
	folderHandler := handlers.NewFolderHandler(folderService)
	noteHandler := handlers.NewNoteHandler(noteService)
	shareHandler := handlers.NewShareHandler(shareService)
	teamHandler := handlers.NewTeamHandler(teamService)

	// Initialize router
	router := http.NewRouter(folderHandler, noteHandler, shareHandler, teamHandler)

	// Setup Gin engine
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		logger.Logger.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Msg("Incoming request")
	})

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
