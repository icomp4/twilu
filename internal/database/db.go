package database

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"twilu/internal/model"
)

func New() (*gorm.DB, error) {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	dbURL := os.Getenv("DB_URL")

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// AutoMigrate your models here
	if err := db.AutoMigrate(&model.User{}, &model.Folder{}, &model.Item{}); err != nil {
		return nil, err
	}

	return db, nil
}
