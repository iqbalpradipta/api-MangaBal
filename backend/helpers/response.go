package helpers

import (
	"errors"
	"log"
	"net/http"

	"scrapingmanga/backend/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func JSON(c echo.Context, status int, success bool, message string, data interface{}) error {
	return c.JSON(status, Response{
		Success: success,
		Message: message,
		Data:    data,
	})
}

func HandleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, utils.ErrBadRequest),
		errors.Is(err, utils.ErrInvalidSlug),
		errors.Is(err, utils.ErrInvalidChapter):
		return JSON(c, http.StatusBadRequest, false, err.Error(), nil)
	case errors.Is(err, utils.ErrUnauthorized):
		return JSON(c, http.StatusUnauthorized, false, err.Error(), nil)
	case errors.Is(err, utils.ErrForbidden):
		return JSON(c, http.StatusForbidden, false, err.Error(), nil)
	case errors.Is(err, utils.ErrNotFound),
		errors.Is(err, utils.ErrIngestJobNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		return JSON(c, http.StatusNotFound, false, "resource not found", nil)
	case errors.Is(err, utils.ErrConflict),
		errors.Is(err, utils.ErrIngestAlreadyRunning),
		errors.Is(err, utils.ErrIngestJobNotCancellable):
		return JSON(c, http.StatusConflict, false, err.Error(), nil)
	case errors.Is(err, utils.ErrBalStorageFailed),
		errors.Is(err, utils.ErrPythonProcessFailed):
		log.Printf("upstream error: %v", err)
		return JSON(c, http.StatusBadGateway, false, err.Error(), nil)
	case errors.Is(err, echo.ErrUnauthorized):
		return JSON(c, http.StatusUnauthorized, false, "unauthorized", nil)
	default:
		log.Printf("internal error: %v", err)
		return JSON(c, http.StatusInternalServerError, false, "internal server error", nil)
	}
}
