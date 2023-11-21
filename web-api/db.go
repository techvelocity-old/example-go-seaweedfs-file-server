package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	postgres_user     = os.Getenv("POSTGRES_USER")
	postgres_password = os.Getenv("POSTGRES_PASSWORD")
	postgres_db       = os.Getenv("POSTGRES_DB")
	postgres_host     = os.Getenv("POSTGRES_HOST")
	postgres_port     = os.Getenv("POSTGRES_PORT")
	postgres_uri      = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", postgres_user, postgres_password, postgres_host, postgres_port, postgres_db)
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(postgres_uri), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	FileRecord := FileRecord{}
	db.AutoMigrate(&FileRecord)
	return db
}
