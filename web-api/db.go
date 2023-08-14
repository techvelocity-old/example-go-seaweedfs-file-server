package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	postgres_user     = "postgres"
	postgres_password = "postgres"
	postgres_db       = "files_db"
	postgres_host     = "postgres"
	postgres_port     = "5432"
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
