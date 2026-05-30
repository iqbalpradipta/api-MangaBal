package controllers

import (
	"net/http"

	"scrapingmanga/backend/helpers"
	"scrapingmanga/backend/services"
	"scrapingmanga/backend/utils"

	"github.com/labstack/echo/v4"
)

type MangaController struct {
	mangaService services.MangaService
}

func NewMangaController(mangaService services.MangaService) *MangaController {
	return &MangaController{mangaService: mangaService}
}

func (m *MangaController) List(c echo.Context) error {
	page, limit := utils.ParsePagination(c)
	result, err := m.mangaService.List(page, limit)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "manga list fetched", result)
}

func (m *MangaController) Search(c echo.Context) error {
	page, limit := utils.ParsePagination(c)
	result, err := m.mangaService.Search(c.QueryParam("q"), page, limit)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "manga search fetched", result)
}

func (m *MangaController) GetBySlug(c echo.Context) error {
	result, err := m.mangaService.GetBySlug(c.Param("slug"))
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "manga detail fetched", result)
}
