package worker

import (
	"context"
	"fmt"
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
	scanRepo repository.FirmwareScanRepository
	vulnRepo repository.VulnerabilityRepository
	
	// For testing
	sleepFunc func(time.Duration)
	randFunc  func(int) int
}

func NewFirmwareAnalysisWorker(scanRepo repository.FirmwareScanRepository, vulnRepo repository.VulnerabilityRepository) *FirmwareAnalysisWorker {
	return &FirmwareAnalysisWorker{
		scanRepo: scanRepo,
		vulnRepo: vulnRepo,
		sleepFunc: time.Sleep,
		randFunc:  rand.Intn,
	}
}

func (w *FirmwareAnalysisWorker) Work(ctx context.Context, job *river.Job[FirmwareAnalysisArgs]) error {
	// Random delay 1-60s
	delay := time.Duration(w.randFunc(59)+1) * time.Second
	
	// We use a select with a timer that can be mocked or bypassed
	if w.sleepFunc != nil {
		w.sleepFunc(delay)
	}

	outcome := w.randFunc(100)

	// 30% Fail (only if SIMULATE_FAILURE is true)
	if os.Getenv("SIMULATE_FAILURE") == "true" && outcome < 30 {
		return fmt.Errorf("simulated temporary failure for scan %d", job.Args.ID)
	}

	// 30% Found, 40% Not Found
	status := "completed"
	var vulns []string

	if outcome < 60 { // 30-59 is 30%
		cve, err := w.vulnRepo.GetRandom(ctx)
		if err == nil && cve != "" {
			vulns = append(vulns, cve)
		}
	}

	return w.scanRepo.UpdateResult(ctx, job.Args.ID, status, vulns)
}
