package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Manga struct {
	ID                 string         `json:"id" gorm:"type:uuid;primaryKey"`
	UpstreamID         int            `json:"upstream_id" gorm:"index"`
	Slug               string         `json:"slug" gorm:"uniqueIndex;not null"`
	Title              string         `json:"title" gorm:"not null;index"`
	NativeTitle        string         `json:"native_title"`
	Author             string         `json:"author"`
	Status             string         `json:"status"`
	Type               string         `json:"type"`
	Format             string         `json:"format"`
	Rating             string         `json:"rating"`
	TotalChapters      int            `json:"total_chapters"`
	Synopsis           string         `json:"synopsis" gorm:"type:text"`
	CoverFileID        string         `json:"cover_file_id"`
	CoverPreviewURL    string         `json:"cover_preview_url" gorm:"type:text"`
	CoverThumbnailURL  string         `json:"cover_thumbnail_url" gorm:"type:text"`
	BalStorageFolderID string         `json:"balstorage_folder_id"`
	Source             string         `json:"source" gorm:"default:primary"`
	LastSyncedAt       *time.Time     `json:"last_synced_at"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
	Genres             []Genre        `json:"genres,omitempty" gorm:"many2many:manga_genres;"`
}

func (m *Manga) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
