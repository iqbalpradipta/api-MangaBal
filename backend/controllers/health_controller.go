package controllers

import (
	"net/http"

	"scrapingmanga/backend/helpers"
	"scrapingmanga/backend/services"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type HealthController struct {
	db    *gorm.DB
	cache services.CacheService
}

func NewHealthController(db *gorm.DB, cache services.CacheService) *HealthController {
	return &HealthController{db: db, cache: cache}
}

func (h *HealthController) Check(c echo.Context) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return helpers.JSON(c, http.StatusServiceUnavailable, false, "database unavailable", nil)
	}
	if err := sqlDB.Ping(); err != nil {
		return helpers.JSON(c, http.StatusServiceUnavailable, false, "database unreachable", nil)
	}

	data := echo.Map{
		"database": "ok",
	}

	return helpers.JSON(c, http.StatusOK, true, "service healthy", data)
}
