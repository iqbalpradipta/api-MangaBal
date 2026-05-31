package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Chapter struct {
	ID                 string         `json:"id" gorm:"type:uuid;primaryKey"`
	MangaID            string         `json:"manga_id" gorm:"type:uuid;not null;index;uniqueIndex:idx_manga_chapter"`
	UpstreamIndex      int            `json:"upstream_index" gorm:"not null;uniqueIndex:idx_manga_chapter"`
	ChapterKey         string         `json:"chapter_key" gorm:"index"`
	Slug               string         `json:"slug" gorm:"index"`
	Title              string         `json:"title"`
	Views              int            `json:"views"`
	TotalPages         int            `json:"total_pages"`
	BalStorageFolderID string         `json:"balstorage_folder_id"`
	LastSyncedAt       *time.Time     `json:"last_synced_at"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
	Pages              []MangaPage    `json:"pages,omitempty" gorm:"foreignKey:ChapterID"`
}

func (c *Chapter) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
