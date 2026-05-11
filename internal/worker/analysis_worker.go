package worker

import (
	"context"
	"firmguard/internal/repository"
	"math/rand"
	"os"
	"time"

	"github.com/riverqueue/river"
)

type FirmwareAnalysisArgs struct {
	ID int `json:"id"`
}

func (FirmwareAnalysisArgs) Kind() string { return "firmware_analysis" }

type FirmwareAnalysisWorker struct {
	river.WorkerDefaults[FirmwareAnalysisArgs]
	repo  repository.FirmwareScanRepository
	delay time.Duration
}

func NewFirmwareAnalysisWorker(repo repository.FirmwareScanRepository) *FirmwareAnalysisWorker {
	return &FirmwareAnalysisWorker{
		repo:  repo,
		delay: 2 * time.Second,
	}
}

func (w *FirmwareAnalysisWorker) Work(ctx context.Context, job *river.Job[FirmwareAnalysisArgs]) error {
	// Respect delay
	if w.delay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(w.delay):
		}
	}

	status := "completed"
	if os.Getenv("SIMULATE_FAILURE") == "true" {
		if rand.Intn(100) < 50 {
			status = "failed"
		}
	}

	return w.repo.UpdateStatus(ctx, job.Args.ID, status)
}
