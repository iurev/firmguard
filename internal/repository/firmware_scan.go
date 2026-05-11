package repository

import (
	"context"
	"database/sql"
	"errors"
	"firmguard/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
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
	Create(ctx context.Context, scan *model.FirmwareScan) error
	GetByDeviceAndHash(ctx context.Context, deviceID, hash string) (*model.FirmwareScan, error)
	UpdateStatus(ctx context.Context, id int, status string) error
}

type firmwareScanRepository struct {
	db *sqlx.DB
}

func NewFirmwareScanRepository(db *sqlx.DB) FirmwareScanRepository {
	return &firmwareScanRepository{db: db}
}

func (r *firmwareScanRepository) Create(ctx context.Context, scan *model.FirmwareScan) error {
	query := `INSERT INTO firmware_scans (device_id, firmware_version, binary_hash, metadata, status) 
			  VALUES (:device_id, :firmware_version, :binary_hash, :metadata, :status) RETURNING id, created_at, updated_at`
	rows, err := r.db.NamedQueryContext(ctx, query, scan)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&scan.ID, &scan.CreatedAt, &scan.UpdatedAt)
	}
	return nil
}

func (r *firmwareScanRepository) GetByDeviceAndHash(ctx context.Context, deviceID, hash string) (*model.FirmwareScan, error) {
	var scan model.FirmwareScan
	err := r.db.GetContext(ctx, &scan, "SELECT * FROM firmware_scans WHERE device_id = $1 AND binary_hash = $2", deviceID, hash)
	if err != nil {
		return nil, err
	}
	return &scan, nil
}

func (r *firmwareScanRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	res, err := r.db.ExecContext(ctx, "UPDATE firmware_scans SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
