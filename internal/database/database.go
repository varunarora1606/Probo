package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(db_url string) {
	dsn := db_url

	var err error // Declare err separately

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Database error:", err.Error())
	}

	fmt.Println("🚀 Connected to PostgreSQL successfully!")
}
