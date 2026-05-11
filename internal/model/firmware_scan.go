package model

import (
	"encoding/json"
	"time"
)

type FirmwareScan struct {
	ID              int             `db:"id" json:"id"`
	DeviceID        string          `db:"device_id" json:"device_id"`
	FirmwareVersion string          `db:"firmware_version" json:"firmware_version"`
	BinaryHash      string          `db:"binary_hash" json:"binary_hash"`
	Metadata        json.RawMessage `db:"metadata" json:"metadata"`
	Status          string          `db:"status" json:"status"`
	Vulns           json.RawMessage `db:"vulns" json:"vulns"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}
