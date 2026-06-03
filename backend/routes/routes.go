package routes

import (
	"scrapingmanga/backend/config"
	"scrapingmanga/backend/controllers"
	"scrapingmanga/backend/middlewares"
	"scrapingmanga/backend/repository"
	"scrapingmanga/backend/services"
	"scrapingmanga/backend/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func Register(e *echo.Echo, db *gorm.DB, cache services.CacheService) {
	api := e.Group("/api/v1")

	healthController := controllers.NewHealthController(db, cache)
	api.GET("/health", healthController.Check)

	mangaRepo := repository.NewMangaRepository(db)
	chapterRepo := repository.NewChapterRepository(db)
	pageRepo := repository.NewMangaPageRepository(db)
	genreRepo := repository.NewGenreRepository(db)
	ingestJobRepo := repository.NewIngestJobRepository(db)

	mangaSvc := services.NewMangaService(mangaRepo)
	chapterSvc := services.NewChapterService(mangaRepo, chapterRepo)
	genreSvc := services.NewGenreService(genreRepo)
	ingestSvc := services.NewIngestService(mangaRepo, chapterRepo, pageRepo, genreRepo, ingestJobRepo)

	mangaController := controllers.NewMangaController(mangaSvc)
	chapterController := controllers.NewChapterController(chapterSvc)
	genreController := controllers.NewGenreController(genreSvc)
	ingestController := controllers.NewIngestController(ingestSvc, cache)

	publicCache := middlewares.ResponseCache(cache)
	api.GET("/manga", mangaController.List, publicCache)
	api.GET("/manga/search", mangaController.Search, publicCache)
	api.GET("/manga/:slug", mangaController.GetBySlug, publicCache)
	api.GET("/manga/:slug/chapters", chapterController.ListByMangaSlug, publicCache)
	api.GET("/manga/:slug/chapters/:chapter", chapterController.GetByMangaSlugAndIndex, publicCache)
	api.GET("/genres", genreController.List, publicCache)

	admin := api.Group("/admin")
	admin.Use(middlewares.AdminToken(utils.GetEnv("ADMIN_TOKEN", "")))
	admin.POST("/ingest/all", ingestController.StartAll, middlewares.RateLimitByIP(3, 1))
	admin.POST("/ingest/series", ingestController.StartSeries, middlewares.RateLimitByIP(10, 3))
	admin.POST("/ingest/chapter", ingestController.StartChapter, middlewares.RateLimitByIP(20, 5))
	admin.GET("/ingest/jobs", ingestController.ListJobs)
	admin.GET("/ingest/jobs/:id", ingestController.GetJob)
	admin.POST("/ingest/jobs/:id/cancel", ingestController.CancelJob)

	internal := api.Group("/internal")
	internal.Use(middlewares.InternalToken(config.LoadIngestConfig().InternalToken))
	internal.POST("/ingest/jobs/:id/progress", ingestController.UpdateProgress)
	internal.POST("/ingest/manga", ingestController.UpsertManga)
	internal.POST("/ingest/chapters", ingestController.UpsertChapters)
	internal.POST("/ingest/pages", ingestController.UpsertPages)
	internal.POST("/ingest/jobs/:id/finish", ingestController.FinishJob)
	internal.POST("/ingest/jobs/:id/fail", ingestController.FailJob)
}
