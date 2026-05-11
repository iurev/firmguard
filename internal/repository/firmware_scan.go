package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"firmguard/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

var ErrDuplicate = errors.New("duplicate entry")

type FirmwareScanRepository interface {
	Create(ctx context.Context, tx pgx.Tx, scan *model.FirmwareScan) error
	GetByDeviceAndHash(ctx context.Context, deviceID, hash string) (*model.FirmwareScan, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	UpdateResult(ctx context.Context, id int, status string, vulns []string) error
}

type firmwareScanRepository struct {
	pool *pgxpool.Pool
}

func NewFirmwareScanRepository(pool *pgxpool.Pool) FirmwareScanRepository {
	return &firmwareScanRepository{pool: pool}
}

func (r *firmwareScanRepository) Create(ctx context.Context, tx pgx.Tx, scan *model.FirmwareScan) error {
	query := `INSERT INTO firmware_scans (device_id, firmware_version, binary_hash, metadata, status) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	
	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, scan.DeviceID, scan.FirmwareVersion, scan.BinaryHash, scan.Metadata, scan.Status).Scan(&scan.ID, &scan.CreatedAt, &scan.UpdatedAt)
	} else {
		err = r.pool.QueryRow(ctx, query, scan.DeviceID, scan.FirmwareVersion, scan.BinaryHash, scan.Metadata, scan.Status).Scan(&scan.ID, &scan.CreatedAt, &scan.UpdatedAt)
	}
	return err
}

func (r *firmwareScanRepository) GetByDeviceAndHash(ctx context.Context, deviceID, hash string) (*model.FirmwareScan, error) {
	var scan model.FirmwareScan
	query := `SELECT id, device_id, firmware_version, binary_hash, metadata, status, vulns, created_at, updated_at 
			  FROM firmware_scans WHERE device_id = $1 AND binary_hash = $2`
	
	err := r.pool.QueryRow(ctx, query, deviceID, hash).Scan(
		&scan.ID, &scan.DeviceID, &scan.FirmwareVersion, &scan.BinaryHash, &scan.Metadata, &scan.Status, &scan.Vulns, &scan.CreatedAt, &scan.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &scan, nil
}

func (r *firmwareScanRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	res, err := r.pool.Exec(ctx, "UPDATE firmware_scans SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
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
		return sql.ErrNoRows
	}
	return nil
}
