package repository

import (
	"strings"

	"scrapingmanga/backend/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MangaRepository interface {
	List(page, limit int) ([]model.Manga, int64, error)
	Search(query string, page, limit int) ([]model.Manga, int64, error)
	FindBySlug(slug string) (*model.Manga, error)
	Upsert(manga *model.Manga) error
}

type mangaRepository struct {
	db *gorm.DB
}

func NewMangaRepository(db *gorm.DB) MangaRepository {
	return &mangaRepository{db: db}
}

func (r *mangaRepository) List(page, limit int) ([]model.Manga, int64, error) {
	var manga []model.Manga
	var total int64

	q := r.db.Model(&model.Manga{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := q.Preload("Genres").Offset(offset).Limit(limit).Order("title ASC").Find(&manga).Error
	return manga, total, err
}

func (r *mangaRepository) Search(query string, page, limit int) ([]model.Manga, int64, error) {
	var manga []model.Manga
	var total int64

	term := "%" + strings.ToLower(query) + "%"
	q := r.db.Model(&model.Manga{}).
		Where("LOWER(title) LIKE ? OR LOWER(slug) LIKE ? OR LOWER(native_title) LIKE ?", term, term, term)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := q.Preload("Genres").Offset(offset).Limit(limit).Order("title ASC").Find(&manga).Error
	return manga, total, err
}

func (r *mangaRepository) FindBySlug(slug string) (*model.Manga, error) {
	var manga model.Manga
	err := r.db.Preload("Genres").First(&manga, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &manga, nil
}

func (r *mangaRepository) Upsert(manga *model.Manga) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"upstream_id",
			"title",
			"native_title",
			"author",
			"status",
			"type",
			"format",
			"rating",
			"total_chapters",
			"synopsis",
			"cover_file_id",
			"cover_preview_url",
			"cover_thumbnail_url",
			"bal_storage_folder_id",
			"source",
			"last_synced_at",
			"updated_at",
		}),
	}).Create(manga).Error
}
