package services

import (
	"scrapingmanga/backend/model"
	"scrapingmanga/backend/repository"
)

type GenreService interface {
	List() ([]model.Genre, error)
}

type genreService struct {
	genreRepo repository.GenreRepository
}

func NewGenreService(genreRepo repository.GenreRepository) GenreService {
	return &genreService{genreRepo: genreRepo}
}

func (s *genreService) List() ([]model.Genre, error) {
	return s.genreRepo.List()
}
