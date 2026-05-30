package repository

import (
	"scrapingmanga/backend/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChapterRepository interface {
	ListByMangaID(mangaID string, page, limit int) ([]model.Chapter, int64, error)
	FindByMangaIDAndIndex(mangaID string, index int) (*model.Chapter, error)
	Upsert(chapter *model.Chapter) error
}

type chapterRepository struct {
	db *gorm.DB
}

func NewChapterRepository(db *gorm.DB) ChapterRepository {
	return &chapterRepository{db: db}
}

func (r *chapterRepository) ListByMangaID(mangaID string, page, limit int) ([]model.Chapter, int64, error) {
	var chapters []model.Chapter
	var total int64

	q := r.db.Model(&model.Chapter{}).Where("manga_id = ?", mangaID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := q.Offset(offset).Limit(limit).Order("upstream_index ASC").Find(&chapters).Error
	return chapters, total, err
}

func (r *chapterRepository) FindByMangaIDAndIndex(mangaID string, index int) (*model.Chapter, error) {
	var chapter model.Chapter
	err := r.db.Preload("Pages", func(db *gorm.DB) *gorm.DB {
		return db.Order("page_number ASC")
	}).First(&chapter, "manga_id = ? AND upstream_index = ?", mangaID, index).Error
	if err != nil {
		return nil, err
	}
	return &chapter, nil
}

func (r *chapterRepository) Upsert(chapter *model.Chapter) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "manga_id"}, {Name: "upstream_index"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"slug",
			"title",
			"views",
			"total_pages",
			"bal_storage_folder_id",
			"last_synced_at",
			"updated_at",
		}),
	}).Create(chapter).Error
}
