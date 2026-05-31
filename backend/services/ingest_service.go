package services

import (
	"time"

	"scrapingmanga/backend/model"
	"scrapingmanga/backend/repository"
	"scrapingmanga/backend/utils"
)

type StartAllIngestInput struct {
	Force       bool `json:"force"`
	MissingOnly bool `json:"missing_only"`
}

type StartSeriesIngestInput struct {
	Slug        string `json:"slug"`
	Force       bool   `json:"force"`
	MissingOnly bool   `json:"missing_only"`
}

type StartChapterIngestInput struct {
	Slug        string `json:"slug"`
	Chapter     int    `json:"chapter"`
	Force       bool   `json:"force"`
	MissingOnly bool   `json:"missing_only"`
}

type IngestProgressInput struct {
	TotalManga        int    `json:"total_manga"`
	ProcessedManga    int    `json:"processed_manga"`
	TotalChapters     int    `json:"total_chapters"`
	ProcessedChapters int    `json:"processed_chapters"`
	TotalPages        int    `json:"total_pages"`
	ProcessedPages    int    `json:"processed_pages"`
	FailedItems       int    `json:"failed_items"`
	Message           string `json:"message"`
}

type IngestMangaInput struct {
	JobID              string   `json:"job_id"`
	UpstreamID         int      `json:"upstream_id"`
	Slug               string   `json:"slug"`
	Title              string   `json:"title"`
	NativeTitle        string   `json:"native_title"`
	Author             string   `json:"author"`
	Status             string   `json:"status"`
	Type               string   `json:"type"`
	Format             string   `json:"format"`
	Rating             string   `json:"rating"`
	TotalChapters      int      `json:"total_chapters"`
	Synopsis           string   `json:"synopsis"`
	CoverFileID        string   `json:"cover_file_id"`
	CoverPreviewURL    string   `json:"cover_preview_url"`
	CoverThumbnailURL  string   `json:"cover_thumbnail_url"`
	BalStorageFolderID string   `json:"balstorage_folder_id"`
	Genres             []string `json:"genres"`
}

type IngestChapterItem struct {
	Index              int    `json:"index"`
	Slug               string `json:"slug"`
	Title              string `json:"title"`
	Views              int    `json:"views"`
	TotalPages         int    `json:"total_pages"`
	BalStorageFolderID string `json:"balstorage_folder_id"`
}

type IngestChaptersInput struct {
	JobID     string              `json:"job_id"`
	MangaSlug string              `json:"manga_slug"`
	Chapters  []IngestChapterItem `json:"chapters"`
}

type IngestPageItem struct {
	PageNumber         int    `json:"page_number"`
	SourceImageURL     string `json:"source_image_url"`
	BalStorageFileID   string `json:"balstorage_file_id"`
	BalStorageFolderID string `json:"balstorage_folder_id"`
	PreviewURL         string `json:"preview_url"`
	DownloadURL        string `json:"download_url"`
	ThumbnailURL       string `json:"thumbnail_url"`
	MimeType           string `json:"mime_type"`
	Size               int64  `json:"size"`
}

type IngestPagesInput struct {
	JobID        string           `json:"job_id"`
	MangaSlug    string           `json:"manga_slug"`
	ChapterIndex int              `json:"chapter_index"`
	Pages        []IngestPageItem `json:"pages"`
}

type IngestFinishInput struct {
	Message string `json:"message"`
}

type IngestFailInput struct {
	ErrorMessage string `json:"error_message"`
}

type IngestJobListResult struct {
	Data       []model.IngestJob `json:"data"`
	Pagination utils.Pagination  `json:"pagination"`
}

type IngestService interface {
	StartAll(input StartAllIngestInput) (*model.IngestJob, error)
	StartSeries(input StartSeriesIngestInput) (*model.IngestJob, error)
	StartChapter(input StartChapterIngestInput) (*model.IngestJob, error)
	ListJobs(page, limit int) (*IngestJobListResult, error)
	GetJob(id string) (*model.IngestJob, error)
	CancelJob(id string) (*model.IngestJob, error)
	UpdateProgress(id string, input IngestProgressInput) (*model.IngestJob, error)
	FinishJob(id string, input IngestFinishInput) (*model.IngestJob, error)
	FailJob(id string, input IngestFailInput) (*model.IngestJob, error)
	UpsertManga(input IngestMangaInput) (*model.Manga, error)
	UpsertChapters(input IngestChaptersInput) error
	UpsertPages(input IngestPagesInput) error
}

type ingestService struct {
	mangaRepo   repository.MangaRepository
	chapterRepo repository.ChapterRepository
	pageRepo    repository.MangaPageRepository
	genreRepo   repository.GenreRepository
	jobRepo     repository.IngestJobRepository
}

func NewIngestService(
	mangaRepo repository.MangaRepository,
	chapterRepo repository.ChapterRepository,
	pageRepo repository.MangaPageRepository,
	genreRepo repository.GenreRepository,
	jobRepo repository.IngestJobRepository,
) IngestService {
	return &ingestService{
		mangaRepo:   mangaRepo,
		chapterRepo: chapterRepo,
		pageRepo:    pageRepo,
		genreRepo:   genreRepo,
		jobRepo:     jobRepo,
	}
}

func (s *ingestService) StartAll(input StartAllIngestInput) (*model.IngestJob, error) {
	active, err := s.jobRepo.FindActiveByType(model.IngestTypeAll)
	if err != nil {
		return nil, err
	}
	if len(active) > 0 {
		return nil, utils.ErrIngestAlreadyRunning
	}

	job := &model.IngestJob{
		Type:        model.IngestTypeAll,
		Status:      model.IngestStatusQueued,
		Force:       input.Force,
		MissingOnly: input.MissingOnly,
		Message:     "waiting for worker",
	}
	return job, s.jobRepo.Create(job)
}

func (s *ingestService) StartSeries(input StartSeriesIngestInput) (*model.IngestJob, error) {
	if !utils.ValidSlug(input.Slug) {
		return nil, utils.ErrInvalidSlug
	}

	job := &model.IngestJob{
		Type:        model.IngestTypeSeries,
		Status:      model.IngestStatusQueued,
		TargetSlug:  input.Slug,
		Force:       input.Force,
		MissingOnly: input.MissingOnly,
		Message:     "waiting for worker",
	}
	return job, s.jobRepo.Create(job)
}

func (s *ingestService) StartChapter(input StartChapterIngestInput) (*model.IngestJob, error) {
	if !utils.ValidSlug(input.Slug) {
		return nil, utils.ErrInvalidSlug
	}
	if input.Chapter < 1 {
		return nil, utils.ErrInvalidChapter
	}

	job := &model.IngestJob{
		Type:          model.IngestTypeChapter,
		Status:        model.IngestStatusQueued,
		TargetSlug:    input.Slug,
		TargetChapter: input.Chapter,
		Force:         input.Force,
		MissingOnly:   input.MissingOnly,
		Message:       "waiting for worker",
	}
	return job, s.jobRepo.Create(job)
}

func (s *ingestService) ListJobs(page, limit int) (*IngestJobListResult, error) {
	jobs, total, err := s.jobRepo.List(page, limit)
	if err != nil {
		return nil, err
	}

	return &IngestJobListResult{
		Data: jobs,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}

func (s *ingestService) GetJob(id string) (*model.IngestJob, error) {
	return s.jobRepo.FindByID(id)
}

func (s *ingestService) CancelJob(id string) (*model.IngestJob, error) {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if job.Status != model.IngestStatusQueued {
		return nil, utils.ErrIngestJobNotCancellable
	}

	now := time.Now()
	job.Status = model.IngestStatusCancelled
	job.FinishedAt = &now
	job.Message = "job cancelled"
	return job, s.jobRepo.Update(job)
}

func (s *ingestService) UpdateProgress(id string, input IngestProgressInput) (*model.IngestJob, error) {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	job.TotalManga = input.TotalManga
	job.ProcessedManga = input.ProcessedManga
	job.TotalChapters = input.TotalChapters
	job.ProcessedChapters = input.ProcessedChapters
	job.TotalPages = input.TotalPages
	job.ProcessedPages = input.ProcessedPages
	job.FailedItems = input.FailedItems
	if input.Message != "" {
		job.Message = input.Message
	}

	return job, s.jobRepo.Update(job)
}

func (s *ingestService) FinishJob(id string, input IngestFinishInput) (*model.IngestJob, error) {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	job.Status = model.IngestStatusDone
	job.FinishedAt = &now
	if input.Message != "" {
		job.Message = input.Message
	} else {
		job.Message = "ingest finished"
	}

	return job, s.jobRepo.Update(job)
}

func (s *ingestService) FailJob(id string, input IngestFailInput) (*model.IngestJob, error) {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	job.Status = model.IngestStatusFailed
	job.FinishedAt = &now
	job.ErrorMessage = input.ErrorMessage
	job.Message = "ingest failed"

	return job, s.jobRepo.Update(job)
}

func (s *ingestService) UpsertManga(input IngestMangaInput) (*model.Manga, error) {
	if !utils.ValidSlug(input.Slug) {
		return nil, utils.ErrInvalidSlug
	}

	now := time.Now()
	manga := &model.Manga{
		UpstreamID:         input.UpstreamID,
		Slug:               input.Slug,
		Title:              input.Title,
		NativeTitle:        input.NativeTitle,
		Author:             input.Author,
		Status:             input.Status,
		Type:               input.Type,
		Format:             input.Format,
		Rating:             input.Rating,
		TotalChapters:      input.TotalChapters,
		Synopsis:           input.Synopsis,
		CoverFileID:        input.CoverFileID,
		CoverPreviewURL:    input.CoverPreviewURL,
		CoverThumbnailURL:  input.CoverThumbnailURL,
		BalStorageFolderID: input.BalStorageFolderID,
		Source:             "primary",
		LastSyncedAt:       &now,
	}

	if err := s.mangaRepo.Upsert(manga); err != nil {
		return nil, err
	}

	stored, err := s.mangaRepo.FindBySlug(input.Slug)
	if err != nil {
		return nil, err
	}

	if len(input.Genres) > 0 {
		genres := make([]model.Genre, 0, len(input.Genres))
		for _, genreName := range input.Genres {
			slug := utils.Slugify(genreName)
			genre := &model.Genre{Name: genreName, Slug: slug}
			if err := s.genreRepo.Upsert(genre); err != nil {
				return nil, err
			}
			storedGenre, err := s.genreRepo.FindBySlug(slug)
			if err != nil {
				return nil, err
			}
			genres = append(genres, *storedGenre)
		}
		if err := s.genreRepo.ReplaceMangaGenres(stored.ID, genres); err != nil {
			return nil, err
		}
	}

	return s.mangaRepo.FindBySlug(input.Slug)
}

func (s *ingestService) UpsertChapters(input IngestChaptersInput) error {
	if !utils.ValidSlug(input.MangaSlug) {
		return utils.ErrInvalidSlug
	}

	manga, err := s.mangaRepo.FindBySlug(input.MangaSlug)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, item := range input.Chapters {
		if item.Index < 1 {
			return utils.ErrInvalidChapter
		}
		chapter := &model.Chapter{
			MangaID:            manga.ID,
			UpstreamIndex:      item.Index,
			Slug:               item.Slug,
			Title:              item.Title,
			Views:              item.Views,
			TotalPages:         item.TotalPages,
			BalStorageFolderID: item.BalStorageFolderID,
			LastSyncedAt:       &now,
		}
		if err := s.chapterRepo.Upsert(chapter); err != nil {
			return err
		}
	}

	return nil
}

func (s *ingestService) UpsertPages(input IngestPagesInput) error {
	if !utils.ValidSlug(input.MangaSlug) {
		return utils.ErrInvalidSlug
	}
	if input.ChapterIndex < 1 {
		return utils.ErrInvalidChapter
	}

	manga, err := s.mangaRepo.FindBySlug(input.MangaSlug)
	if err != nil {
		return err
	}
	chapter, err := s.chapterRepo.FindByMangaIDAndIndex(manga.ID, input.ChapterIndex)
	if err != nil {
		return err
	}

	pages := make([]model.MangaPage, 0, len(input.Pages))
	for _, item := range input.Pages {
		if item.PageNumber < 1 {
			return utils.ErrBadRequest
		}
		pages = append(pages, model.MangaPage{
			ChapterID:          chapter.ID,
			PageNumber:         item.PageNumber,
			SourceImageURL:     item.SourceImageURL,
			BalStorageFileID:   item.BalStorageFileID,
			BalStorageFolderID: item.BalStorageFolderID,
			PreviewURL:         item.PreviewURL,
			DownloadURL:        item.DownloadURL,
			ThumbnailURL:       item.ThumbnailURL,
			MimeType:           item.MimeType,
			Size:               item.Size,
		})
	}

	return s.pageRepo.UpsertMany(pages)
}
