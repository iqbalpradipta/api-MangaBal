package utils

import "errors"

var (
	ErrBadRequest              = errors.New("bad request")
	ErrNotFound                = errors.New("resource not found")
	ErrConflict                = errors.New("resource conflict")
	ErrForbidden               = errors.New("forbidden")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrInvalidSlug             = errors.New("invalid slug")
	ErrInvalidChapter          = errors.New("invalid chapter")
	ErrIngestAlreadyRunning    = errors.New("ingest job already running")
	ErrIngestJobNotFound       = errors.New("ingest job not found")
	ErrIngestJobNotCancellable = errors.New("ingest job cannot be cancelled")
	ErrBalStorageFailed        = errors.New("balstorage request failed")
	ErrPythonProcessFailed     = errors.New("python process failed")
)
