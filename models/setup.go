package models

import (
	"github.com/lucblassel/training-tracker/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() error {
	path := config.GetDBPath()

	database, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return err
	}

	database.AutoMigrate(&Run{}, &Tag{})

	DB = database

	return nil
}
