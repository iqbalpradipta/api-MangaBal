package repository

import (
	"scrapingmanga/backend/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GenreRepository interface {
	List() ([]model.Genre, error)
	Upsert(genre *model.Genre) error
	FindBySlug(slug string) (*model.Genre, error)
	ReplaceMangaGenres(mangaID string, genres []model.Genre) error
}

type genreRepository struct {
	db *gorm.DB
}

func NewGenreRepository(db *gorm.DB) GenreRepository {
	return &genreRepository{db: db}
}

func (r *genreRepository) List() ([]model.Genre, error) {
	var genres []model.Genre
	err := r.db.Order("name ASC").Find(&genres).Error
	return genres, err
}

func (r *genreRepository) Upsert(genre *model.Genre) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"upstream_id",
			"name",
			"updated_at",
		}),
	}).Create(genre).Error
}

func (r *genreRepository) FindBySlug(slug string) (*model.Genre, error) {
	var genre model.Genre
	err := r.db.First(&genre, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &genre, nil
}

func (r *genreRepository) ReplaceMangaGenres(mangaID string, genres []model.Genre) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("manga_id = ?", mangaID).Delete(&model.MangaGenre{}).Error; err != nil {
			return err
		}

		for _, genre := range genres {
			join := model.MangaGenre{MangaID: mangaID, GenreID: genre.ID}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&join).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
