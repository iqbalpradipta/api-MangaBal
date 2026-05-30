package config

import (
	"fmt"
	"log"
	"time"

	"scrapingmanga/backend/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		utils.GetEnv("DB_HOST", "127.0.0.1"),
		utils.GetEnv("DB_PORT", "5432"),
		utils.GetEnv("DB_USER", "postgres"),
		utils.GetEnv("DB_PASS", ""),
		utils.GetEnv("DB_NAME", "manga_api"),
		utils.GetEnv("DB_SSLMODE", "disable"),
		utils.GetEnv("DB_TIMEZONE", "Asia/Jakarta"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get database handle: %v", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(2 * time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	log.Println("database connected successfully")
	return db
}
