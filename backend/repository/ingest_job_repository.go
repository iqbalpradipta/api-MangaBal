package repository

import (
	"scrapingmanga/backend/model"

	"gorm.io/gorm"
)

type IngestJobRepository interface {
	Create(job *model.IngestJob) error
	FindByID(id string) (*model.IngestJob, error)
	FindQueued(limit int) ([]model.IngestJob, error)
	FindActiveByType(jobType string) ([]model.IngestJob, error)
	List(page, limit int) ([]model.IngestJob, int64, error)
	Update(job *model.IngestJob) error
}

type ingestJobRepository struct {
	db *gorm.DB
}

func NewIngestJobRepository(db *gorm.DB) IngestJobRepository {
	return &ingestJobRepository{db: db}
}

func (r *ingestJobRepository) Create(job *model.IngestJob) error {
	return r.db.Create(job).Error
}

func (r *ingestJobRepository) FindByID(id string) (*model.IngestJob, error) {
	var job model.IngestJob
	err := r.db.First(&job, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *ingestJobRepository) FindQueued(limit int) ([]model.IngestJob, error) {
	var jobs []model.IngestJob
	err := r.db.Where("status = ?", model.IngestStatusQueued).
		Order("created_at ASC").
		Limit(limit).
		Find(&jobs).Error
	return jobs, err
}

func (r *ingestJobRepository) FindActiveByType(jobType string) ([]model.IngestJob, error) {
	var jobs []model.IngestJob
	err := r.db.Where("type = ? AND status IN ?", jobType, []string{
		model.IngestStatusQueued,
		model.IngestStatusRunning,
	}).Find(&jobs).Error
	return jobs, err
}

func (r *ingestJobRepository) List(page, limit int) ([]model.IngestJob, int64, error) {
	var jobs []model.IngestJob
	var total int64

	q := r.db.Model(&model.IngestJob{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := q.Offset(offset).Limit(limit).Order("created_at DESC").Find(&jobs).Error
	return jobs, total, err
}

func (r *ingestJobRepository) Update(job *model.IngestJob) error {
	return r.db.Save(job).Error
}
