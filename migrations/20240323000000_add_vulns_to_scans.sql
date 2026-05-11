-- +goose Up
ALTER TABLE firmware_scans ADD COLUMN IF NOT EXISTS vulns JSONB NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE firmware_scans DROP COLUMN IF EXISTS vulns;
