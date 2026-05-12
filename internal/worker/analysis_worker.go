package worker

import (
	"context"
	"firmguard/internal/repository"
	"fmt"
	"math/rand"
	"time"

	"github.com/riverqueue/river"
)

type FirmwareAnalysisArgs struct {
	ID int `json:"id"`
}

func (FirmwareAnalysisArgs) Kind() string { return "firmware_analysis" }

type FirmwareAnalysisWorker struct {
	river.WorkerDefaults[FirmwareAnalysisArgs]
	scanRepo repository.FirmwareScanRepository
	vulnRepo repository.VulnerabilityRepository

	// For testing
	sleepFunc func(time.Duration)
	randFunc  func(int) int
}

func NewFirmwareAnalysisWorker(scanRepo repository.FirmwareScanRepository, vulnRepo repository.VulnerabilityRepository) *FirmwareAnalysisWorker {
	return &FirmwareAnalysisWorker{
		scanRepo:  scanRepo,
		vulnRepo:  vulnRepo,
		sleepFunc: time.Sleep,
		randFunc:  rand.Intn,
	}
}

func (w *FirmwareAnalysisWorker) MaxAttempts() int { return 3 }

func (w *FirmwareAnalysisWorker) Work(ctx context.Context, job *river.Job[FirmwareAnalysisArgs]) error {
	delay := time.Duration(w.randFunc(59)+1) * time.Second
	if w.sleepFunc != nil {
		w.sleepFunc(delay)
	}

	outcome := w.randFunc(100)

	// 30% fail, 30% found (30-59), 40% not found (60-99)
	if outcome < 30 {
		if job.Attempt >= w.MaxAttempts() {
			_ = w.scanRepo.UpdateResult(ctx, job.Args.ID, "failed", nil)
		}
		return fmt.Errorf("simulated temporary failure for scan %d", job.Args.ID)
	}

	status := "completed"
	var vulns []string

	if outcome < 60 {
		cve, err := w.vulnRepo.GetRandom(ctx)
		if err == nil && cve != "" {
			vulns = append(vulns, cve)
		}
	}

	return w.scanRepo.UpdateResult(ctx, job.Args.ID, status, vulns)
}
