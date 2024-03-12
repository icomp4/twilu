package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"twilu/internal/model"
)

func New() (*gorm.DB, error) {
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
