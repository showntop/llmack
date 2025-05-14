package database

import (
	"fmt"
	"log"

	"github.com/bookkeeper-ai/bookkeeper/config"
	"github.com/bookkeeper-ai/bookkeeper/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	DB = db

	// 自动迁移数据库表
	err = DB.AutoMigrate(&models.Transaction{}, &models.Category{}, &models.Budget{}, &models.User{})
	if err != nil {
		return err
	}

	log.Println("Database connected successfully")
	return nil
}
