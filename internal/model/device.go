package model

import (
	"encoding/json"
	"time"
)

type Device struct {
	ID              string          `db:"id" json:"id"`
	FirmwareVersion string          `db:"firmware_version" json:"firmware_version"`
	BinaryHash      string          `db:"binary_hash" json:"binary_hash"`
	Vulns           json.RawMessage `db:"vulns" json:"vulns"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}
