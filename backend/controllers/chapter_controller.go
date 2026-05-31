package controllers

import (
	"net/http"

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
	result, err := cc.chapterService.GetByMangaSlugAndKey(c.Param("slug"), c.Param("chapter"))
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "chapter detail fetched", result)
}
