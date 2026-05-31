package services

import (
	"scrapingmanga/backend/model"
	"scrapingmanga/backend/repository"
	"scrapingmanga/backend/utils"
)

type ChapterListResult struct {
	Data       []model.Chapter  `json:"data"`
	Pagination utils.Pagination `json:"pagination"`
}

type ChapterService interface {
	ListByMangaSlug(slug string, page, limit int) (*ChapterListResult, error)
	GetByMangaSlugAndKey(slug string, chapterKey string) (*model.Chapter, error)
}

type chapterService struct {
	mangaRepo   repository.MangaRepository
	chapterRepo repository.ChapterRepository
}

func NewChapterService(
	mangaRepo repository.MangaRepository,
	chapterRepo repository.ChapterRepository,
) ChapterService {
	return &chapterService{
		mangaRepo:   mangaRepo,
		chapterRepo: chapterRepo,
	}
}

func (s *chapterService) ListByMangaSlug(slug string, page, limit int) (*ChapterListResult, error) {
	if !utils.ValidSlug(slug) {
		return nil, utils.ErrInvalidSlug
	}

	manga, err := s.mangaRepo.FindBySlug(slug)
	if err != nil {
		return nil, err
	}

	chapters, total, err := s.chapterRepo.ListByMangaID(manga.ID, page, limit)
	if err != nil {
		return nil, err
	}

	return &ChapterListResult{
		Data: chapters,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}

func (s *chapterService) GetByMangaSlugAndKey(slug string, chapterKey string) (*model.Chapter, error) {
	if !utils.ValidSlug(slug) {
		return nil, utils.ErrInvalidSlug
	}
	chapterKey = utils.NormalizeChapterKey(chapterKey)
	if !utils.ValidChapterKey(chapterKey) {
		return nil, utils.ErrInvalidChapter
	}

	manga, err := s.mangaRepo.FindBySlug(slug)
	if err != nil {
		return nil, err
	}

	return s.chapterRepo.FindByMangaIDAndKey(manga.ID, chapterKey, utils.ChapterStorageIndex(chapterKey))
}
