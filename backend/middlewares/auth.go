package middlewares

import (
	"scrapingmanga/backend/helpers"
	"scrapingmanga/backend/utils"

	"github.com/labstack/echo/v4"
)

func AdminToken(expected string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if expected == "" || c.Request().Header.Get("X-Admin-Token") != expected {
				return helpers.HandleError(c, utils.ErrUnauthorized)
			}
			return next(c)
		}
	}
}
