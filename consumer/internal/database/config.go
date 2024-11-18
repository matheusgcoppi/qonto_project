package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"qonto_project/consumer/internal/models"
)

func ConnectDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		return nil, err
	}
	return db, nil
}

func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Balance{},
	)
}
