package repository

import (
	"scrapingmanga/backend/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MangaPageRepository interface {
	ListByChapterID(chapterID string) ([]model.MangaPage, error)
	UpsertMany(pages []model.MangaPage) error
}

type mangaPageRepository struct {
	db *gorm.DB
}

func NewMangaPageRepository(db *gorm.DB) MangaPageRepository {
	return &mangaPageRepository{db: db}
}

func (r *mangaPageRepository) ListByChapterID(chapterID string) ([]model.MangaPage, error) {
	var pages []model.MangaPage
	err := r.db.Where("chapter_id = ?", chapterID).Order("page_number ASC").Find(&pages).Error
	return pages, err
}

func (r *mangaPageRepository) UpsertMany(pages []model.MangaPage) error {
	if len(pages) == 0 {
		return nil
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "chapter_id"}, {Name: "page_number"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"source_image_url",
			"bal_storage_file_id",
			"bal_storage_folder_id",
			"preview_url",
			"download_url",
			"thumbnail_url",
			"mime_type",
			"size",
			"updated_at",
		}),
	}).Create(&pages).Error
}
