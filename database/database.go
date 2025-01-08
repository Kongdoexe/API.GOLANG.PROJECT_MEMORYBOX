package database

import (
	"fmt"
	"log"
	"os"

	"API.GOLANG.PROJECT_MEMORYBOX/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DBconn *gorm.DB

func InitDB() {

	user := os.Getenv("db_user")
	pass := os.Getenv("db_password")
	host := os.Getenv("db_host")
	dbname := os.Getenv("db_dbname")

	if user == "" || pass == "" || host == "" || dbname == "" {
		log.Fatal("Missing database environment variables")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, dbname)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})

	if err != nil {
		panic("Database connect failed.")
	}
	// Auto Migrate the User and Chat models
	err = db.AutoMigrate(&models.User{}, &models.Event{}, &models.Join{}, &models.Favorite{}, &models.Media{}, &models.Notification{}, &models.ResetPassToken{})
	if err != nil {
		log.Fatal("Failed to migrate database schema:", err)
	}

	log.Println("Connect Success.")
	DBconn = db
}
