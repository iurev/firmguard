package service

import (
	"context"
	"errors"
	"firmguard/internal/model"
	"firmguard/internal/repository"
	"log"
	"math/rand"
	"os"
	"time"
)

var ErrScanAlreadyExists = errors.New("scan already exists")

var scanWorkerDelay = 2 * time.Second
var randIntn = rand.Intn

type FirmwareScanService interface {
	CreateScan(ctx context.Context, scan *model.FirmwareScan) (*model.FirmwareScan, error)
}

type firmwareScanService struct {
	repo repository.FirmwareScanRepository
}

func NewFirmwareScanService(repo repository.FirmwareScanRepository) FirmwareScanService {
	return &firmwareScanService{repo: repo}
}

func (s *firmwareScanService) CreateScan(ctx context.Context, scan *model.FirmwareScan) (*model.FirmwareScan, error) {
	scan.Status = "pending"
	if err := s.repo.Create(ctx, scan); err != nil {
		if repository.IsUniqueViolation(err) {
			existing, getErr := s.repo.GetByDeviceAndHash(ctx, scan.DeviceID, scan.BinaryHash)
			if getErr == nil {
				return existing, ErrScanAlreadyExists
			}
			return nil, err
		}
		return nil, err
	}

	// Trigger background worker
	go func(id int) {
		time.Sleep(scanWorkerDelay)
		status := "completed"
		if os.Getenv("SIMULATE_FAILURE") == "true" {
			if randIntn(100) < 50 {
				status = "failed"
			}
		}
		if err := s.repo.UpdateStatus(context.Background(), id, status); err != nil {
			log.Printf("failed to update scan status for id %d: %v", id, err)
		}
	}(scan.ID)

	return scan, nil
}
