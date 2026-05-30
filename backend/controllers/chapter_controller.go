package controllers

import (
	"net/http"
	"strconv"

	"scrapingmanga/backend/helpers"
	"scrapingmanga/backend/services"
	"scrapingmanga/backend/utils"

	"github.com/labstack/echo/v4"
)

type ChapterController struct {
	chapterService services.ChapterService
}

func NewChapterController(chapterService services.ChapterService) *ChapterController {
	return &ChapterController{chapterService: chapterService}
}

func (cc *ChapterController) ListByMangaSlug(c echo.Context) error {
	page, limit := utils.ParsePagination(c)
	result, err := cc.chapterService.ListByMangaSlug(c.Param("slug"), page, limit)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "chapter list fetched", result)
}

func (cc *ChapterController) GetByMangaSlugAndIndex(c echo.Context) error {
	chapterIndex, err := strconv.Atoi(c.Param("chapter"))
	if err != nil {
		return helpers.HandleError(c, utils.ErrInvalidChapter)
	}

	result, err := cc.chapterService.GetByMangaSlugAndIndex(c.Param("slug"), chapterIndex)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "chapter detail fetched", result)
}
