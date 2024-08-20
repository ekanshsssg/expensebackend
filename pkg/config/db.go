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
	// dsn := "root:pass@123@tcp(localhost:3306)/expensedb?parseTime=true"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// if err := database.AutoMigrate(&Group{}); err != nil {
    //     fmt.Println("Failed to migrate database:", err)
    //     return
    // }
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
