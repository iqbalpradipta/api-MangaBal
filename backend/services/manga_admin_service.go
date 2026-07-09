package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"scrapingmanga/backend/config"
	"scrapingmanga/backend/model"
	"scrapingmanga/backend/repository"
	"scrapingmanga/backend/utils"
)

// CreateMangaInput holds validated fields for manual manga creation.
type CreateMangaInput struct {
	Title       string
	NativeTitle string
	Author      string
	Status      string
	Type        string
	Synopsis    string
	Genres      []string // genre slugs
	CoverFile   *multipart.FileHeader // optional
}

// UpdateMangaInput holds fields that can be updated.
type UpdateMangaInput struct {
	Title       *string
	NativeTitle *string
	Author      *string
	Status      *string
	Type        *string
	Synopsis    *string
	Genres      []string // genre slugs; nil = no change, empty slice = clear
}

// CreateChapterInput holds validated fields for manual chapter creation.
type CreateChapterInput struct {
	ChapterIndex int    // used as upstream_index
	Title        string
}

// UploadPagesInput holds a set of page files for a chapter.
type UploadPagesInput struct {
	Files []*multipart.FileHeader
}

// MangaAdminService handles manual create/update/delete of manga and chapters.
type MangaAdminService interface {
	CreateManga(input CreateMangaInput) (*model.Manga, error)
	UpdateManga(slug string, input UpdateMangaInput) (*model.Manga, error)
	DeleteManga(slug string) error
	CreateChapter(mangaSlug string, input CreateChapterInput) (*model.Chapter, error)
	DeleteChapter(mangaSlug string, chapterIndex int) error
	UploadPages(mangaSlug string, chapterIndex int, input UploadPagesInput) ([]model.MangaPage, error)
}

type mangaAdminService struct {
	mangaRepo   repository.MangaRepository
	chapterRepo repository.ChapterRepository
	pageRepo    repository.MangaPageRepository
	genreRepo   repository.GenreRepository
	upload      UploadService
	balCfg      config.BalStorageConfig
}

func NewMangaAdminService(
	mangaRepo repository.MangaRepository,
	chapterRepo repository.ChapterRepository,
	pageRepo repository.MangaPageRepository,
	genreRepo repository.GenreRepository,
	balCfg config.BalStorageConfig,
) MangaAdminService {
	return &mangaAdminService{
		mangaRepo:   mangaRepo,
		chapterRepo: chapterRepo,
		pageRepo:    pageRepo,
		genreRepo:   genreRepo,
		upload:      NewUploadService(balCfg),
		balCfg:      balCfg,
	}
}

// CreateManga creates a new manga record, optionally uploading a cover image.
func (s *mangaAdminService) CreateManga(input CreateMangaInput) (*model.Manga, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, utils.ErrBadRequest
	}

	slug := utils.Slugify(input.Title)
	if !utils.ValidSlug(slug) {
		return nil, utils.ErrInvalidSlug
	}

	// resolve genres
	genres, err := s.resolveGenres(input.Genres)
	if err != nil {
		return nil, err
	}

	manga := &model.Manga{
		Slug:        slug,
		Title:       input.Title,
		NativeTitle: input.NativeTitle,
		Author:      input.Author,
		Status:      input.Status,
		Type:        input.Type,
		Synopsis:    input.Synopsis,
		Source:      "manual",
		Genres:      genres,
	}

	// upload cover if provided
	if input.CoverFile != nil {
		uploaded, folderID, err := s.uploadCover(slug, input.CoverFile)
		if err != nil {
			return nil, fmt.Errorf("cover upload: %w", err)
		}
		manga.CoverFileID = uploaded.FileID
		manga.CoverPreviewURL = uploaded.PreviewURL
		manga.CoverThumbnailURL = uploaded.ThumbnailURL
		manga.BalStorageFolderID = folderID
	}

	if err := s.mangaRepo.Upsert(manga); err != nil {
		return nil, err
	}
	if len(genres) > 0 {
		if err := s.genreRepo.ReplaceMangaGenres(manga.ID, genres); err != nil {
			return nil, err
		}
	}
	return s.mangaRepo.FindBySlug(slug)
}

// UpdateManga patches mutable fields on an existing manga.
func (s *mangaAdminService) UpdateManga(slug string, input UpdateMangaInput) (*model.Manga, error) {
	if !utils.ValidSlug(slug) {
		return nil, utils.ErrInvalidSlug
	}
	manga, err := s.mangaRepo.FindBySlug(slug)
	if err != nil {
		return nil, utils.ErrNotFound
	}

	if input.Title != nil {
		manga.Title = *input.Title
	}
	if input.NativeTitle != nil {
		manga.NativeTitle = *input.NativeTitle
	}
	if input.Author != nil {
		manga.Author = *input.Author
	}
	if input.Status != nil {
		manga.Status = *input.Status
	}
	if input.Type != nil {
		manga.Type = *input.Type
	}
	if input.Synopsis != nil {
		manga.Synopsis = *input.Synopsis
	}

	if err := s.mangaRepo.Upsert(manga); err != nil {
		return nil, err
	}

	// nil = caller didn't send genres field at all → skip
	if input.Genres != nil {
		genres, err := s.resolveGenres(input.Genres)
		if err != nil {
			return nil, err
		}
		if err := s.genreRepo.ReplaceMangaGenres(manga.ID, genres); err != nil {
			return nil, err
		}
	}

	return s.mangaRepo.FindBySlug(slug)
}

// DeleteManga soft-deletes a manga by slug.
func (s *mangaAdminService) DeleteManga(slug string) error {
	if !utils.ValidSlug(slug) {
		return utils.ErrInvalidSlug
	}
	manga, err := s.mangaRepo.FindBySlug(slug)
	if err != nil {
		return utils.ErrNotFound
	}
	return s.mangaRepo.Delete(manga.ID)
}

// CreateChapter adds a new chapter to an existing manga.
func (s *mangaAdminService) CreateChapter(mangaSlug string, input CreateChapterInput) (*model.Chapter, error) {
	if !utils.ValidSlug(mangaSlug) {
		return nil, utils.ErrInvalidSlug
	}
	manga, err := s.mangaRepo.FindBySlug(mangaSlug)
	if err != nil {
		return nil, utils.ErrNotFound
	}

	chapterKey := fmt.Sprintf("%d", input.ChapterIndex)
	chapter := &model.Chapter{
		MangaID:       manga.ID,
		UpstreamIndex: input.ChapterIndex,
		ChapterKey:    chapterKey,
		Slug:          fmt.Sprintf("%s-chapter-%d", mangaSlug, input.ChapterIndex),
		Title:         input.Title,
	}

	if err := s.chapterRepo.Upsert(chapter); err != nil {
		return nil, err
	}
	return s.chapterRepo.FindByMangaIDAndKey(manga.ID, chapterKey, input.ChapterIndex)
}

// DeleteChapter soft-deletes a chapter by manga slug + chapter index.
func (s *mangaAdminService) DeleteChapter(mangaSlug string, chapterIndex int) error {
	if !utils.ValidSlug(mangaSlug) {
		return utils.ErrInvalidSlug
	}
	manga, err := s.mangaRepo.FindBySlug(mangaSlug)
	if err != nil {
		return utils.ErrNotFound
	}
	chapterKey := fmt.Sprintf("%d", chapterIndex)
	chapter, err := s.chapterRepo.FindByMangaIDAndKey(manga.ID, chapterKey, chapterIndex)
	if err != nil {
		return utils.ErrNotFound
	}
	return s.chapterRepo.Delete(chapter.ID)
}

// UploadPages uploads multipart images and stores them as MangaPage records.
func (s *mangaAdminService) UploadPages(mangaSlug string, chapterIndex int, input UploadPagesInput) ([]model.MangaPage, error) {
	if !utils.ValidSlug(mangaSlug) {
		return nil, utils.ErrInvalidSlug
	}
	if len(input.Files) == 0 {
		return nil, utils.ErrBadRequest
	}

	manga, err := s.mangaRepo.FindBySlug(mangaSlug)
	if err != nil {
		return nil, utils.ErrNotFound
	}

	chapterKey := fmt.Sprintf("%d", chapterIndex)
	chapter, err := s.chapterRepo.FindByMangaIDAndKey(manga.ID, chapterKey, chapterIndex)
	if err != nil {
		return nil, utils.ErrNotFound
	}

	// ensure BalStorage folder: Manga/{Title}/Chapter {N}/
	rootID, err := s.upload.EnsureFolder(s.balCfg.RootFolderName, nil)
	if err != nil {
		return nil, fmt.Errorf("balstorage root folder: %w", err)
	}
	mangaFolderID, err := s.upload.EnsureFolder(manga.Title, &rootID)
	if err != nil {
		return nil, fmt.Errorf("balstorage manga folder: %w", err)
	}
	chapterFolderName := fmt.Sprintf("Chapter %d", chapterIndex)
	chapterFolderID, err := s.upload.EnsureFolder(chapterFolderName, &mangaFolderID)
	if err != nil {
		return nil, fmt.Errorf("balstorage chapter folder: %w", err)
	}

	pages := make([]model.MangaPage, 0, len(input.Files))
	for i, fh := range input.Files {
		f, err := fh.Open()
		if err != nil {
			return nil, fmt.Errorf("open file %s: %w", fh.Filename, err)
		}
		uploaded, uploadErr := s.uploadFile(chapterFolderID, fh.Filename, fh.Header.Get("Content-Type"), f)
		f.Close()
		if uploadErr != nil {
			return nil, fmt.Errorf("upload page %d: %w", i+1, uploadErr)
		}

		pages = append(pages, model.MangaPage{
			ChapterID:          chapter.ID,
			PageNumber:         i + 1,
			BalStorageFileID:   uploaded.FileID,
			BalStorageFolderID: chapterFolderID,
			PreviewURL:         uploaded.PreviewURL,
			DownloadURL:        uploaded.DownloadURL,
			ThumbnailURL:       uploaded.ThumbnailURL,
			MimeType:           uploaded.MimeType,
			Size:               uploaded.Size,
		})
	}

	if err := s.pageRepo.UpsertMany(pages); err != nil {
		return nil, err
	}

	// update chapter total_pages
	chapter.TotalPages = len(pages)
	s.chapterRepo.Upsert(chapter) // best-effort

	return pages, nil
}

// --- helpers ---

func (s *mangaAdminService) resolveGenres(slugs []string) ([]model.Genre, error) {
	genres := make([]model.Genre, 0, len(slugs))
	for _, slug := range slugs {
		g, err := s.genreRepo.FindBySlug(slug)
		if err != nil {
			return nil, fmt.Errorf("genre not found: %s", slug)
		}
		genres = append(genres, *g)
	}
	return genres, nil
}

func (s *mangaAdminService) uploadCover(mangaSlug string, fh *multipart.FileHeader) (*UploadedFile, string, error) {
	rootID, err := s.upload.EnsureFolder(s.balCfg.RootFolderName, nil)
	if err != nil {
		return nil, "", err
	}
	folderID, err := s.upload.EnsureFolder(mangaSlug, &rootID)
	if err != nil {
		return nil, "", err
	}
	f, err := fh.Open()
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	uploaded, err := s.uploadFile(folderID, fh.Filename, fh.Header.Get("Content-Type"), f)
	return uploaded, folderID, err
}

func (s *mangaAdminService) uploadFile(folderID, filename, mime string, r io.Reader) (*UploadedFile, error) {
	if mime == "" {
		mime = mimeFromFilename(filename)
	}
	return s.upload.UploadFile(folderID, filename, mime, r)
}

func mimeFromFilename(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}
