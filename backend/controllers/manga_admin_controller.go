package controllers

import (
	"net/http"
	"strconv"

	"scrapingmanga/backend/helpers"
	"scrapingmanga/backend/services"
	"scrapingmanga/backend/utils"

	"github.com/labstack/echo/v4"
)

type MangaAdminController struct {
	svc services.MangaAdminService
}

func NewMangaAdminController(svc services.MangaAdminService) *MangaAdminController {
	return &MangaAdminController{svc: svc}
}

// CreateManga handles POST /api/v1/admin/manga
// multipart/form-data: title*, native_title, author, status, type, synopsis, genres[] (slugs), cover (file)
func (ctl *MangaAdminController) CreateManga(c echo.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		return helpers.JSON(c, http.StatusBadRequest, false, "invalid multipart form", nil)
	}

	input := services.CreateMangaInput{
		Title:       formVal(form.Value, "title"),
		NativeTitle: formVal(form.Value, "native_title"),
		Author:      formVal(form.Value, "author"),
		Status:      formVal(form.Value, "status"),
		Type:        formVal(form.Value, "type"),
		Synopsis:    formVal(form.Value, "synopsis"),
		Genres:      form.Value["genres"],
	}

	if files := form.File["cover"]; len(files) > 0 {
		input.CoverFile = files[0]
	}

	manga, err := ctl.svc.CreateManga(input)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusCreated, true, "manga created", manga)
}

// UpdateManga handles PUT /api/v1/admin/manga/:slug
// application/json body (all fields optional)
func (ctl *MangaAdminController) UpdateManga(c echo.Context) error {
	slug := c.Param("slug")

	var body struct {
		Title       *string  `json:"title"`
		NativeTitle *string  `json:"native_title"`
		Author      *string  `json:"author"`
		Status      *string  `json:"status"`
		Type        *string  `json:"type"`
		Synopsis    *string  `json:"synopsis"`
		Genres      []string `json:"genres"` // nil = omitted, [] = clear
	}
	if err := c.Bind(&body); err != nil {
		return helpers.JSON(c, http.StatusBadRequest, false, "invalid request body", nil)
	}

	input := services.UpdateMangaInput{
		Title:       body.Title,
		NativeTitle: body.NativeTitle,
		Author:      body.Author,
		Status:      body.Status,
		Type:        body.Type,
		Synopsis:    body.Synopsis,
		Genres:      body.Genres,
	}

	manga, err := ctl.svc.UpdateManga(slug, input)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "manga updated", manga)
}

// DeleteManga handles DELETE /api/v1/admin/manga/:slug
func (ctl *MangaAdminController) DeleteManga(c echo.Context) error {
	if err := ctl.svc.DeleteManga(c.Param("slug")); err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "manga deleted", nil)
}

// CreateChapter handles POST /api/v1/admin/manga/:slug/chapters
// application/json: { "chapter_index": 1, "title": "..." }
func (ctl *MangaAdminController) CreateChapter(c echo.Context) error {
	var body struct {
		ChapterIndex int    `json:"chapter_index"`
		Title        string `json:"title"`
	}
	if err := c.Bind(&body); err != nil || body.ChapterIndex <= 0 {
		return helpers.JSON(c, http.StatusBadRequest, false, "chapter_index required and must be > 0", nil)
	}

	chapter, err := ctl.svc.CreateChapter(c.Param("slug"), services.CreateChapterInput{
		ChapterIndex: body.ChapterIndex,
		Title:        body.Title,
	})
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusCreated, true, "chapter created", chapter)
}

// DeleteChapter handles DELETE /api/v1/admin/manga/:slug/chapters/:chapter
func (ctl *MangaAdminController) DeleteChapter(c echo.Context) error {
	idx, err := strconv.Atoi(c.Param("chapter"))
	if err != nil || idx <= 0 {
		return helpers.JSON(c, http.StatusBadRequest, false, utils.ErrInvalidChapter.Error(), nil)
	}
	if err := ctl.svc.DeleteChapter(c.Param("slug"), idx); err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "chapter deleted", nil)
}

// UploadPages handles POST /api/v1/admin/manga/:slug/chapters/:chapter/pages
// multipart/form-data: files[] (images, ordered)
func (ctl *MangaAdminController) UploadPages(c echo.Context) error {
	idx, err := strconv.Atoi(c.Param("chapter"))
	if err != nil || idx <= 0 {
		return helpers.JSON(c, http.StatusBadRequest, false, utils.ErrInvalidChapter.Error(), nil)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return helpers.JSON(c, http.StatusBadRequest, false, "invalid multipart form", nil)
	}

	files := form.File["files"]
	if len(files) == 0 {
		return helpers.JSON(c, http.StatusBadRequest, false, "no files provided", nil)
	}

	pages, err := ctl.svc.UploadPages(c.Param("slug"), idx, services.UploadPagesInput{Files: files})
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusCreated, true, "pages uploaded", pages)
}

// formVal safely gets first value from multipart form values map.
func formVal(values map[string][]string, key string) string {
	if v, ok := values[key]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}
