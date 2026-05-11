package service

import (
	"context"
	"database/sql"
	"errors"
	"firmguard/internal/model"
	"firmguard/internal/repository"
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
	existing, err := s.repo.GetByDeviceAndHash(ctx, scan.DeviceID, scan.BinaryHash)
	if err == nil && existing != nil {
		return existing, ErrScanAlreadyExists
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	scan.Status = "pending"
	if err := s.repo.Create(ctx, scan); err != nil {
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
		_ = s.repo.UpdateStatus(context.Background(), id, status)
	}(scan.ID)

	return scan, nil
}
