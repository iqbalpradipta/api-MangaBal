package config

import (
	"strconv"
	"time"

	"scrapingmanga/backend/utils"
)

type IngestConfig struct {
	PythonBin          string
	AllScript          string
	SeriesScript       string
	InternalToken      string
	APIBaseURL         string
	MaxParallelJobs    int
	PollInterval       time.Duration
	MaxSeries          string
	MaxChapters        string
	BalStorageBaseURL  string
	BalStorageEmail    string
	BalStoragePassword string
	BalStorageRoot     string
}

func LoadIngestConfig() IngestConfig {
	maxParallel, _ := strconv.Atoi(utils.GetEnv("INGEST_MAX_PARALLEL_JOBS", "1"))
	if maxParallel < 1 {
		maxParallel = 1
	}

	pollSeconds, _ := strconv.Atoi(utils.GetEnv("INGEST_POLL_INTERVAL_SECONDS", "5"))
	if pollSeconds < 1 {
		pollSeconds = 5
	}

	return IngestConfig{
		PythonBin:          utils.GetEnv("PYTHON_BIN", "py"),
		AllScript:          utils.GetEnv("INGEST_ALL_SCRIPT", "../scripts/ingest_all.py"),
		SeriesScript:       utils.GetEnv("INGEST_SERIES_SCRIPT", "../scripts/ingest_series.py"),
		InternalToken:      utils.GetEnv("INGEST_INTERNAL_TOKEN", ""),
		APIBaseURL:         utils.GetEnv("INGEST_API_BASE_URL", "http://localhost:8001/api/v1"),
		MaxParallelJobs:    maxParallel,
		PollInterval:       time.Duration(pollSeconds) * time.Second,
		MaxSeries:          utils.GetEnv("INGEST_MAX_SERIES", ""),
		MaxChapters:        utils.GetEnv("INGEST_MAX_CHAPTERS", ""),
		BalStorageBaseURL:  utils.GetEnv("BALSTORAGE_BASE_URL", "http://localhost:8000/api/v1"),
		BalStorageEmail:    utils.GetEnv("BALSTORAGE_EMAIL", ""),
		BalStoragePassword: utils.GetEnv("BALSTORAGE_PASSWORD", ""),
		BalStorageRoot:     utils.GetEnv("BALSTORAGE_ROOT_FOLDER_NAME", "Manga"),
	}
}
