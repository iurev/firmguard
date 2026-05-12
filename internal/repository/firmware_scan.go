package repository

import (
	"context"
	"firmguard/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateResult struct {
	Scan       *model.FirmwareScan
	IsInserted bool
}

type FirmwareScanRepository interface {
	Create(ctx context.Context, tx pgx.Tx, scan *model.FirmwareScan) (CreateResult, error)
	GetByID(ctx context.Context, id int) (*model.FirmwareScan, error)
	UpdateResult(ctx context.Context, id int, status string, vulns []string) error
}

type firmwareScanRepository struct {
	pool *pgxpool.Pool
}

func NewFirmwareScanRepository(pool *pgxpool.Pool) FirmwareScanRepository {
	return &firmwareScanRepository{pool: pool}
}

func (r *firmwareScanRepository) Create(ctx context.Context, tx pgx.Tx, scan *model.FirmwareScan) (CreateResult, error) {
	query := `INSERT INTO firmware_scans (device_id, firmware_version, binary_hash, metadata, status)
			  VALUES ($1, $2, $3, $4, $5)
			  ON CONFLICT (device_id, binary_hash)
			  DO UPDATE SET updated_at = NOW()
			  RETURNING id, status, created_at, updated_at, (xmax = 0) AS is_inserted`

	var isInserted bool
	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, scan.DeviceID, scan.FirmwareVersion, scan.BinaryHash, scan.Metadata, scan.Status).Scan(&scan.ID, &scan.Status, &scan.CreatedAt, &scan.UpdatedAt, &isInserted)
	} else {
		err = r.pool.QueryRow(ctx, query, scan.DeviceID, scan.FirmwareVersion, scan.BinaryHash, scan.Metadata, scan.Status).Scan(&scan.ID, &scan.Status, &scan.CreatedAt, &scan.UpdatedAt, &isInserted)
	}
	return CreateResult{Scan: scan, IsInserted: isInserted}, err
}

func (r *firmwareScanRepository) GetByID(ctx context.Context, id int) (*model.FirmwareScan, error) {
	query := `SELECT id, device_id, firmware_version, binary_hash, metadata, status, vulns, created_at, updated_at
	          FROM firmware_scans WHERE id = $1`
	var scan model.FirmwareScan
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&scan.ID, &scan.DeviceID, &scan.FirmwareVersion, &scan.BinaryHash,
		&scan.Metadata, &scan.Status, &scan.Vulns, &scan.CreatedAt, &scan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &scan, nil
}

func (r *firmwareScanRepository) UpdateResult(ctx context.Context, id int, status string, vulns []string) error {
	if vulns == nil {
		vulns = []string{}
	}

	res, err := r.pool.Exec(ctx, "UPDATE firmware_scans SET status = $1, vulns = $2, updated_at = NOW() WHERE id = $3", status, vulns, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
