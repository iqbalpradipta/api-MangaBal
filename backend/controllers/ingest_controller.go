package controllers

import (
	"net/http"

	"scrapingmanga/backend/helpers"
	"scrapingmanga/backend/services"
	"scrapingmanga/backend/utils"

	"github.com/labstack/echo/v4"
)

type IngestController struct {
	ingestService services.IngestService
}

func NewIngestController(ingestService services.IngestService) *IngestController {
	return &IngestController{ingestService: ingestService}
}

func (ic *IngestController) StartAll(c echo.Context) error {
	job, err := ic.ingestService.StartAll()
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusAccepted, true, "ingest job queued", job)
}

func (ic *IngestController) StartSeries(c echo.Context) error {
	var input services.StartSeriesIngestInput
	if err := c.Bind(&input); err != nil {
		return helpers.HandleError(c, utils.ErrBadRequest)
	}

	job, err := ic.ingestService.StartSeries(input)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusAccepted, true, "ingest job queued", job)
}

func (ic *IngestController) StartChapter(c echo.Context) error {
	var input services.StartChapterIngestInput
	if err := c.Bind(&input); err != nil {
		return helpers.HandleError(c, utils.ErrBadRequest)
	}

	job, err := ic.ingestService.StartChapter(input)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusAccepted, true, "ingest job queued", job)
}

func (ic *IngestController) ListJobs(c echo.Context) error {
	page, limit := utils.ParsePagination(c)
	result, err := ic.ingestService.ListJobs(page, limit)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "ingest jobs fetched", result)
}

func (ic *IngestController) GetJob(c echo.Context) error {
	job, err := ic.ingestService.GetJob(c.Param("id"))
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "ingest job fetched", job)
}

func (ic *IngestController) CancelJob(c echo.Context) error {
	job, err := ic.ingestService.CancelJob(c.Param("id"))
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "ingest job cancelled", job)
}

func (ic *IngestController) UpdateProgress(c echo.Context) error {
	var input services.IngestProgressInput
	if err := c.Bind(&input); err != nil {
		return helpers.HandleError(c, utils.ErrBadRequest)
	}

	job, err := ic.ingestService.UpdateProgress(c.Param("id"), input)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "ingest progress updated", job)
}

func (ic *IngestController) FinishJob(c echo.Context) error {
	var input services.IngestFinishInput
	_ = c.Bind(&input)

	job, err := ic.ingestService.FinishJob(c.Param("id"), input)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "ingest job finished", job)
}

func (ic *IngestController) FailJob(c echo.Context) error {
	var input services.IngestFailInput
	_ = c.Bind(&input)

	job, err := ic.ingestService.FailJob(c.Param("id"), input)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "ingest job failed", job)
}

func (ic *IngestController) UpsertManga(c echo.Context) error {
	var input services.IngestMangaInput
	if err := c.Bind(&input); err != nil {
		return helpers.HandleError(c, utils.ErrBadRequest)
	}

	manga, err := ic.ingestService.UpsertManga(input)
	if err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "manga upserted", manga)
}

func (ic *IngestController) UpsertChapters(c echo.Context) error {
	var input services.IngestChaptersInput
	if err := c.Bind(&input); err != nil {
		return helpers.HandleError(c, utils.ErrBadRequest)
	}

	if err := ic.ingestService.UpsertChapters(input); err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "chapters upserted", nil)
}

func (ic *IngestController) UpsertPages(c echo.Context) error {
	var input services.IngestPagesInput
	if err := c.Bind(&input); err != nil {
		return helpers.HandleError(c, utils.ErrBadRequest)
	}

	if err := ic.ingestService.UpsertPages(input); err != nil {
		return helpers.HandleError(c, err)
	}
	return helpers.JSON(c, http.StatusOK, true, "pages upserted", nil)
}
