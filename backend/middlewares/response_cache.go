package middlewares

import (
	"bytes"
	"net/http"
	"strings"

	"scrapingmanga/backend/services"

	"github.com/labstack/echo/v4"
)

type cacheResponseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *cacheResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func ResponseCache(cache services.CacheService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			if cache == nil || !cache.Enabled() || req.Method != http.MethodGet {
				return next(c)
			}

			key := cache.PublicKey(req.URL.RequestURI())
			if cached, ok := cache.GetBytes(req.Context(), key); ok {
				c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
				c.Response().Header().Set("X-Cache", "HIT")
				return c.Blob(http.StatusOK, echo.MIMEApplicationJSONCharsetUTF8, cached)
			}

			recorder := &cacheResponseWriter{
				ResponseWriter: c.Response().Writer,
				body:           bytes.NewBuffer(nil),
			}
			c.Response().Writer = recorder
			c.Response().Header().Set("X-Cache", "MISS")

			if err := next(c); err != nil {
				return err
			}

			contentType := c.Response().Header().Get(echo.HeaderContentType)
			if c.Response().Status == http.StatusOK &&
				strings.Contains(contentType, echo.MIMEApplicationJSON) &&
				recorder.body.Len() > 0 {
				cache.SetBytes(req.Context(), key, recorder.body.Bytes())
			}

			return nil
		}
	}
}
