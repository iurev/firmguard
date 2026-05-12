package service

import (
	"context"
	"firmguard/internal/database"
	"firmguard/internal/model"
	"firmguard/internal/repository"
	"firmguard/internal/worker"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type FirmwareScanService interface {
	CreateScan(ctx context.Context, scan *model.FirmwareScan) (*model.FirmwareScan, bool, error)
	GetScan(ctx context.Context, id int) (*model.FirmwareScan, error)
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

func (s *firmwareScanService) CreateScan(ctx context.Context, scan *model.FirmwareScan) (*model.FirmwareScan, bool, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback(ctx)

	scan.Status = "pending"
	res, err := s.repo.Create(ctx, tx, scan)
	if err != nil {
		return nil, false, err
	}

	// Only enqueue background analysis if it's a new scan.
	if res.IsInserted {
		_, err = s.river.InsertTx(ctx, tx, worker.FirmwareAnalysisArgs{ID: res.Scan.ID}, nil)
		if err != nil {
			return nil, false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}

	return res.Scan, res.IsInserted, nil
}

func (s *firmwareScanService) GetScan(ctx context.Context, id int) (*model.FirmwareScan, error) {
	return s.repo.GetByID(ctx, id)
}
