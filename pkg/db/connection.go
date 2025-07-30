package db

import (
	"log"
	"team-service/internal/entities"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect establishes a database connection
func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate all entities
	err = db.AutoMigrate(
		&entities.Folder{},
		&entities.Note{},
		&entities.FolderShare{},
		&entities.NoteShare{},
		// &entities.Team{},
		// &entities.Roster{},
		// &entities.User{},
	)
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return nil, err
	}

	return db, nil
}

// SetupDatabase initializes the database connection and runs migrations
func SetupDatabase(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		log.Fatal("DATABASE_DSN is not set in environment variables")
	}

	db, err := Connect(dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, err
	}

	log.Println("Database connected and migrated successfully")
	return db, nil
}
