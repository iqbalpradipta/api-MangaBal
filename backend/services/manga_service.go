package services

import (
	"strings"

	"scrapingmanga/backend/model"
	"scrapingmanga/backend/repository"
	"scrapingmanga/backend/utils"
)

type MangaListResult struct {
	Data       []model.Manga    `json:"data"`
	Pagination utils.Pagination `json:"pagination"`
}

type MangaService interface {
	List(page, limit int) (*MangaListResult, error)
	Search(query string, page, limit int) (*MangaListResult, error)
	GetBySlug(slug string) (*model.Manga, error)
}

type mangaService struct {
	mangaRepo repository.MangaRepository
}

func NewMangaService(mangaRepo repository.MangaRepository) MangaService {
	return &mangaService{mangaRepo: mangaRepo}
}

func (s *mangaService) List(page, limit int) (*MangaListResult, error) {
	items, total, err := s.mangaRepo.List(page, limit)
	if err != nil {
		return nil, err
	}

	return &MangaListResult{
		Data: items,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}

func (s *mangaService) Search(query string, page, limit int) (*MangaListResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, utils.ErrBadRequest
	}

	items, total, err := s.mangaRepo.Search(query, page, limit)
	if err != nil {
		return nil, err
	}

	return &MangaListResult{
		Data: items,
		Pagination: utils.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}

func (s *mangaService) GetBySlug(slug string) (*model.Manga, error) {
	if !utils.ValidSlug(slug) {
		return nil, utils.ErrInvalidSlug
	}
	return s.mangaRepo.FindBySlug(slug)
}
