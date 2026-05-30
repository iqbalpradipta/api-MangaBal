package controllers

import (
	"net/http"

	"scrapingmanga/backend/helpers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type HealthController struct {
	db *gorm.DB
}

func NewHealthController(db *gorm.DB) *HealthController {
	return &HealthController{db: db}
}

func (h *HealthController) Check(c echo.Context) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return helpers.JSON(c, http.StatusServiceUnavailable, false, "database unavailable", nil)
	}
	if err := sqlDB.Ping(); err != nil {
		return helpers.JSON(c, http.StatusServiceUnavailable, false, "database unreachable", nil)
	}

	return helpers.JSON(c, http.StatusOK, true, "service healthy", echo.Map{
		"database": "ok",
	})
}
