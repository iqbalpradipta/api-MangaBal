package controllers

import (
	"net/http"

	"scrapingmanga/backend/helpers"
	"scrapingmanga/backend/services"

	"github.com/labstack/echo/v4"
)

type GenreController struct {
	genreService services.GenreService
}

func NewGenreController(genreService services.GenreService) *GenreController {
	return &GenreController{genreService: genreService}
}

func (gc *GenreController) List(c echo.Context) error {
	genres, err := gc.genreService.List()
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "genre list fetched", genres)
}
