package controllers

import (
	"context"
	"net/http"
	"time"

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
		"redis":    "disabled",
	}
	if h.cache != nil && h.cache.Configured() {
		if !h.cache.Enabled() {
			data["redis"] = "unavailable"
			return helpers.JSON(c, http.StatusOK, true, "service healthy", data)
		}

		pingCtx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := h.cache.Ping(pingCtx); err != nil {
			data["redis"] = "unavailable"
		} else {
			data["redis"] = "ok"
		}
	}

	return helpers.JSON(c, http.StatusOK, true, "service healthy", data)
}
