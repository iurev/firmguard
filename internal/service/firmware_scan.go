package service

import (
	"context"
	"errors"
	"firmguard/internal/database"
	"firmguard/internal/model"
	"firmguard/internal/repository"
	"firmguard/internal/worker"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

var ErrScanAlreadyExists = errors.New("scan already exists")

type FirmwareScanService interface {
	CreateScan(ctx context.Context, scan *model.FirmwareScan) (*model.FirmwareScan, error)
}

type RiverClient interface {
	Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
	InsertTx(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

type firmwareScanService struct {
	db    database.Service
	repo  repository.FirmwareScanRepository
	river RiverClient
}

func NewFirmwareScanService(db database.Service, repo repository.FirmwareScanRepository, riverClient RiverClient) FirmwareScanService {
	return &firmwareScanService{
		db:    db,
		repo:  repo,
		river: riverClient,
	}
}

func (s *firmwareScanService) CreateScan(ctx context.Context, scan *model.FirmwareScan) (*model.FirmwareScan, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	scan.Status = "pending"
	if err := s.repo.Create(ctx, tx, scan); err != nil {
		return nil, err
	}

	// Only enqueue background analysis if it's a new scan.
	// We detect this by checking if CreatedAt and UpdatedAt are equal,
	// which happens on the initial INSERT but not on ON CONFLICT UPDATE.
	if scan.CreatedAt.Equal(scan.UpdatedAt) {
		_, err = s.river.InsertTx(ctx, tx, worker.FirmwareAnalysisArgs{ID: scan.ID}, nil)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return scan, nil
}
