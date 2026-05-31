package services

import (
	"context"
	"log"
	"os/exec"
	"strconv"
	"time"

	"scrapingmanga/backend/config"
	"scrapingmanga/backend/model"
	"scrapingmanga/backend/repository"
)

type IngestWorkerService struct {
	jobRepo repository.IngestJobRepository
	cfg     config.IngestConfig
}

func NewIngestWorkerService(jobRepo repository.IngestJobRepository, cfg config.IngestConfig) *IngestWorkerService {
	return &IngestWorkerService{
		jobRepo: jobRepo,
		cfg:     cfg,
	}
}

func (s *IngestWorkerService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOnce(ctx)
		}
	}
}

func (s *IngestWorkerService) runOnce(ctx context.Context) {
	jobs, err := s.jobRepo.FindQueued(s.cfg.MaxParallelJobs)
	if err != nil {
		log.Printf("failed to fetch ingest jobs: %v", err)
		return
	}

	for i := range jobs {
		job := jobs[i]
		go s.runJob(ctx, &job)
	}
}

func (s *IngestWorkerService) runJob(ctx context.Context, job *model.IngestJob) {
	now := time.Now()
	job.Status = model.IngestStatusRunning
	job.StartedAt = &now
	job.Message = "python ingest process started"
	if err := s.jobRepo.Update(job); err != nil {
		log.Printf("failed to mark ingest job running: %v", err)
		return
	}

	args := s.commandArgs(job)
	cmd := exec.CommandContext(ctx, s.cfg.PythonBin, args...)
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		log.Printf("ingest job %s output: %s", job.ID, string(output))
	}

	stored, findErr := s.jobRepo.FindByID(job.ID)
	if findErr != nil {
		log.Printf("failed to reload ingest job %s: %v", job.ID, findErr)
		return
	}

	if err != nil {
		finished := time.Now()
		stored.Status = model.IngestStatusFailed
		stored.FinishedAt = &finished
		if stored.ErrorMessage == "" {
			stored.ErrorMessage = err.Error()
		}
		stored.Message = "python ingest process failed"
		if updateErr := s.jobRepo.Update(stored); updateErr != nil {
			log.Printf("failed to mark ingest job failed: %v", updateErr)
		}
		return
	}

	if stored.Status == model.IngestStatusRunning {
		finished := time.Now()
		stored.Status = model.IngestStatusDone
		stored.FinishedAt = &finished
		stored.Message = "python ingest process finished"
		if updateErr := s.jobRepo.Update(stored); updateErr != nil {
			log.Printf("failed to mark ingest job done: %v", updateErr)
		}
	}
}

func (s *IngestWorkerService) commandArgs(job *model.IngestJob) []string {
	script := s.cfg.AllScript
	if job.Type == model.IngestTypeSeries || job.Type == model.IngestTypeChapter {
		script = s.cfg.SeriesScript
	}

	args := []string{
		script,
		"--job-id", job.ID,
		"--api-base", s.cfg.APIBaseURL,
		"--internal-token", s.cfg.InternalToken,
		"--balstorage-base", s.cfg.BalStorageBaseURL,
		"--balstorage-email", s.cfg.BalStorageEmail,
		"--balstorage-password", s.cfg.BalStoragePassword,
		"--balstorage-root", s.cfg.BalStorageRoot,
	}

	if job.Type == model.IngestTypeSeries || job.Type == model.IngestTypeChapter {
		args = append(args, "--slug", job.TargetSlug)
	}
	if job.Type == model.IngestTypeChapter {
		chapter := job.TargetChapterKey
		if chapter == "" {
			chapter = strconv.Itoa(job.TargetChapter)
		}
		args = append(args, "--chapter", chapter)
	}
	if job.Force {
		args = append(args, "--force")
	}
	if job.MissingOnly {
		args = append(args, "--missing-only")
	}
	if s.cfg.MaxSeries != "" {
		args = append(args, "--max-series", s.cfg.MaxSeries)
	}
	if s.cfg.MaxChapters != "" {
		args = append(args, "--max-chapters", s.cfg.MaxChapters)
	}

	return args
}
