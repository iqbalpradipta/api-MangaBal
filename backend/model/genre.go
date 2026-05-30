package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Genre struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey"`
	UpstreamID int       `json:"upstream_id" gorm:"index"`
	Name       string    `json:"name" gorm:"uniqueIndex;not null"`
	Slug       string    `json:"slug" gorm:"uniqueIndex;not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (g *Genre) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	return nil
}

type MangaGenre struct {
	MangaID string `json:"manga_id" gorm:"type:uuid;primaryKey"`
	GenreID string `json:"genre_id" gorm:"type:uuid;primaryKey"`
}
