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
	// Fix 1: recover jobs yang stuck running saat server restart
	s.recoverStuckJobs()

	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()

	// Fix 2: watchdog ticker — cek job stuck setiap 1 menit
	watchdog := time.NewTicker(1 * time.Minute)
	defer watchdog.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOnce(ctx)
		case <-watchdog.C:
			s.killStuckJobs()
		}
	}
}

// recoverStuckJobs dipanggil saat startup — mark semua job 'running' jadi 'failed'
// karena Python process pasti sudah mati saat server restart.
func (s *IngestWorkerService) recoverStuckJobs() {
	// pakai threshold 0 = semua job running tanpa batas waktu
	jobs, err := s.jobRepo.FindStuckRunning(time.Now())
	if err != nil {
		log.Printf("recovery: failed to find stuck jobs: %v", err)
		return
	}
	for i := range jobs {
		job := &jobs[i]
		now := time.Now()
		job.Status = model.IngestStatusFailed
		job.FinishedAt = &now
		job.ErrorMessage = "job recovered after server restart — process was killed"
		job.Message = "recovered: marked failed on startup"
		if err := s.jobRepo.Update(job); err != nil {
			log.Printf("recovery: failed to update job %s: %v", job.ID, err)
		} else {
			log.Printf("recovery: job %s marked failed (was stuck running)", job.ID)
		}
	}
}

// killStuckJobs dipanggil periodic — kill job running melebihi JobTimeout.
func (s *IngestWorkerService) killStuckJobs() {
	threshold := time.Now().Add(-s.cfg.JobTimeout)
	jobs, err := s.jobRepo.FindStuckRunning(threshold)
	if err != nil {
		log.Printf("watchdog: failed to find stuck jobs: %v", err)
		return
	}
	for i := range jobs {
		job := &jobs[i]
		now := time.Now()
		job.Status = model.IngestStatusFailed
		job.FinishedAt = &now
		job.ErrorMessage = "job timed out — exceeded maximum allowed duration"
		job.Message = "watchdog: marked failed due to timeout"
		if err := s.jobRepo.Update(job); err != nil {
			log.Printf("watchdog: failed to update job %s: %v", job.ID, err)
		} else {
			log.Printf("watchdog: job %s timed out after %v, marked failed", job.ID, s.cfg.JobTimeout)
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

	// Fix 3: pakai job-scoped context dengan timeout, bukan parent ctx langsung
	// supaya shutdown server tidak langsung kill job tanpa mark failed
	jobCtx, cancel := context.WithTimeout(context.Background(), s.cfg.JobTimeout)
	defer cancel()

	args := s.commandArgs(job)
	cmd := exec.CommandContext(jobCtx, s.cfg.PythonBin, args...)
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
		// hanya update jika Python belum callback finish/fail sendiri
		if stored.Status == model.IngestStatusRunning {
			finished := time.Now()
			stored.Status = model.IngestStatusFailed
			stored.FinishedAt = &finished
			if stored.ErrorMessage == "" {
				stored.ErrorMessage = err.Error()
			}
			// bedakan timeout vs error lain
			if jobCtx.Err() == context.DeadlineExceeded {
				stored.ErrorMessage = "job timed out — python process killed after " + s.cfg.JobTimeout.String()
			}
			stored.Message = "python ingest process failed"
			if updateErr := s.jobRepo.Update(stored); updateErr != nil {
				log.Printf("failed to mark ingest job failed: %v", updateErr)
			}
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
