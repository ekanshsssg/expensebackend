package config

import (
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := os.Getenv("DB_URL")
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = database
	// PingDatabase()
}

func GetDB() *gorm.DB {
    return DB
}

// func PingDatabase() {
// 	sqlDB, err := DB.DB()
// 	if err != nil {
// 		log.Fatal("Failed to get database instance:", err)
// 	}

// 	if err := sqlDB.Ping(); err != nil {
// 		log.Fatal("Failed to ping database:", err)
// 	}

// 	log.Println("Database connection successfully established")

// }
