package migration

import (
	"log"

	"scrapingmanga/backend/model"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	if err := db.AutoMigrate(
		&model.Manga{},
		&model.Chapter{},
		&model.MangaPage{},
		&model.Genre{},
		&model.MangaGenre{},
		&model.IngestJob{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
}
