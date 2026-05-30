package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"scrapingmanga/backend/config"
	"scrapingmanga/backend/migration"
	"scrapingmanga/backend/repository"
	"scrapingmanga/backend/routes"
	"scrapingmanga/backend/services"
	"scrapingmanga/backend/utils"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env vars")
	}

	db := config.InitDB()
	migration.AutoMigrate(db)

	ingestCfg := config.LoadIngestConfig()
	ingestJobRepo := repository.NewIngestJobRepository(db)
	go services.NewIngestWorkerService(ingestJobRepo, ingestCfg).Start(context.Background())

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{utils.GetEnv("FRONTEND_ORIGIN", "*")},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Admin-Token",
			"X-Internal-Token",
		},
	}))

	routes.Register(e, db)
	e.File("/swagger", "docs/index.html")
	e.Static("/swagger", "docs")

	port := utils.GetEnv("APP_PORT", "8001")
	addr := fmt.Sprintf(":%s", port)
	log.Printf("server starting on %s", addr)

	if err := e.Start(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
