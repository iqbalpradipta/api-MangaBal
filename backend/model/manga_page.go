package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MangaPage struct {
	ID                 string         `json:"id" gorm:"type:uuid;primaryKey"`
	ChapterID          string         `json:"chapter_id" gorm:"type:uuid;not null;index;uniqueIndex:idx_chapter_page"`
	PageNumber         int            `json:"page_number" gorm:"not null;uniqueIndex:idx_chapter_page"`
	SourceImageURL     string         `json:"source_image_url" gorm:"type:text"`
	BalStorageFileID   string         `json:"balstorage_file_id" gorm:"index"`
	BalStorageFolderID string         `json:"balstorage_folder_id"`
	PreviewURL         string         `json:"preview_url" gorm:"type:text"`
	DownloadURL        string         `json:"download_url" gorm:"type:text"`
	ThumbnailURL       string         `json:"thumbnail_url" gorm:"type:text"`
	MimeType           string         `json:"mime_type"`
	Size               int64          `json:"size"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
}

func (p *MangaPage) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
