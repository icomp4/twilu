package database

import (
	"log"
	"os"
	"twilu/model"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var err error

func StartDB() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	dbURL := os.Getenv("DB_URL")
	
	DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database")
	}
	DB.AutoMigrate(&model.User{}, &model.Folder{}, &model.Picture{})
}
