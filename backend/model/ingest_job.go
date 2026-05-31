package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	IngestTypeAll     = "all"
	IngestTypeSeries  = "series"
	IngestTypeChapter = "chapter"
	IngestTypeGenres  = "genres"

	IngestStatusQueued    = "queued"
	IngestStatusRunning   = "running"
	IngestStatusDone      = "done"
	IngestStatusFailed    = "failed"
	IngestStatusCancelled = "cancelled"
)

type IngestJob struct {
	ID                string     `json:"id" gorm:"type:uuid;primaryKey"`
	Type              string     `json:"type" gorm:"not null;index"`
	Status            string     `json:"status" gorm:"not null;index"`
	TargetSlug        string     `json:"target_slug"`
	TargetChapter     int        `json:"target_chapter"`
	Force             bool       `json:"force"`
	MissingOnly       bool       `json:"missing_only"`
	TotalManga        int        `json:"total_manga"`
	ProcessedManga    int        `json:"processed_manga"`
	TotalChapters     int        `json:"total_chapters"`
	ProcessedChapters int        `json:"processed_chapters"`
	TotalPages        int        `json:"total_pages"`
	ProcessedPages    int        `json:"processed_pages"`
	FailedItems       int        `json:"failed_items"`
	Message           string     `json:"message" gorm:"type:text"`
	ErrorMessage      string     `json:"error_message" gorm:"type:text"`
	StartedAt         *time.Time `json:"started_at"`
	FinishedAt        *time.Time `json:"finished_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (j *IngestJob) BeforeCreate(tx *gorm.DB) error {
	if j.ID == "" {
		j.ID = uuid.New().String()
	}
	if j.Status == "" {
		j.Status = IngestStatusQueued
	}
	return nil
}
