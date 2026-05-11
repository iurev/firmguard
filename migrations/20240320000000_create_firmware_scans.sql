-- +goose Up
CREATE TABLE firmware_scans (
    id SERIAL PRIMARY KEY,
    device_id TEXT NOT NULL,
    firmware_version TEXT NOT NULL,
    binary_hash TEXT NOT NULL,
    metadata JSONB,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, binary_hash)
);

CREATE INDEX idx_firmware_scans_device_id ON firmware_scans(device_id);
CREATE INDEX idx_firmware_scans_binary_hash ON firmware_scans(binary_hash);

-- +goose Down
DROP TABLE firmware_scans;
