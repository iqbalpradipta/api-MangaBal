package services

import (
	"fmt"
	"strings"

	"scrapingmanga/backend/config"
)

type BalStorageService interface {
	PreviewURL(fileID string) string
	DownloadURL(fileID string) string
	ThumbnailURL(fileID string) string
}

type balStorageService struct {
	baseURL string
}

func NewBalStorageService(cfg config.BalStorageConfig) BalStorageService {
	return &balStorageService{baseURL: strings.TrimRight(cfg.BaseURL, "/")}
}

func (s *balStorageService) PreviewURL(fileID string) string {
	return fmt.Sprintf("%s/files/%s/preview", s.baseURL, fileID)
}

func (s *balStorageService) DownloadURL(fileID string) string {
	return fmt.Sprintf("%s/files/%s/download", s.baseURL, fileID)
}

func (s *balStorageService) ThumbnailURL(fileID string) string {
	return fmt.Sprintf("%s/files/%s/thumbnail", s.baseURL, fileID)
}
