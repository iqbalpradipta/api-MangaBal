package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestAppAdsRoute(t *testing.T) {
	e := echo.New()
	e.GET("/app-ads.txt", func(c echo.Context) error {
		return c.String(http.StatusOK, appAdsTxt)
	})

	req := httptest.NewRequest(http.MethodGet, "/app-ads.txt", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "pub-8675873135912570") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
