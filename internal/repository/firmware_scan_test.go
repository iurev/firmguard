package repository

import (
	"context"
	"encoding/json"
	"firmguard/internal/model"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirmwareScanRepository(t *testing.T) {
	repo := NewFirmwareScanRepository(testPool)
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		scan := &model.FirmwareScan{
			DeviceID:        "dev1",
			FirmwareVersion: "1.0.0",
			BinaryHash:      "hash1",
			Metadata:        json.RawMessage(`{"key": "value"}`),
			Status:          "pending",
		}

		err := repo.Create(ctx, nil, scan)
		require.NoError(t, err)
		assert.NotZero(t, scan.ID)
		assert.NotEmpty(t, scan.CreatedAt)
		assert.NotEmpty(t, scan.UpdatedAt)

		t.Run("WithTransaction", func(t *testing.T) {
			tx, err := testPool.Begin(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			scan2 := &model.FirmwareScan{
				DeviceID:        "dev2",
				FirmwareVersion: "2.0.0",
				BinaryHash:      "hash2",
				Status:          "pending",
			}

			err = repo.Create(ctx, tx, scan2)
			require.NoError(t, err)
			assert.NotZero(t, scan2.ID)

			err = tx.Commit(ctx)
			require.NoError(t, err)
		})

		t.Run("DuplicateEntry", func(t *testing.T) {
			// Create is designed with ON CONFLICT DO UPDATE SET updated_at = EXCLUDED.updated_at
			// So it won't return an error for duplicates unless we specifically cause one.
			// Let's check the behavior.
			scanDup := &model.FirmwareScan{
				DeviceID:        "dev1",
				FirmwareVersion: "1.0.0",
				BinaryHash:      "hash1",
				Status:          "pending",
			}
			err := repo.Create(ctx, nil, scanDup)
			require.NoError(t, err) // Should succeed due to ON CONFLICT
		})
	})

	t.Run("GetByDeviceAndHash", func(t *testing.T) {
		scan, err := repo.GetByDeviceAndHash(ctx, "dev1", "hash1")
		require.NoError(t, err)
		assert.Equal(t, "dev1", scan.DeviceID)
		assert.Equal(t, "hash1", scan.BinaryHash)

		_, err = repo.GetByDeviceAndHash(ctx, "nonexistent", "hash")
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		// Create a scan to update
		scan := &model.FirmwareScan{
			DeviceID:        "dev3",
			FirmwareVersion: "3.0.0",
			BinaryHash:      "hash3",
			Status:          "pending",
		}
		err := repo.Create(ctx, nil, scan)
		require.NoError(t, err)

		err = repo.UpdateStatus(ctx, scan.ID, "completed")
		require.NoError(t, err)

		updated, err := repo.GetByDeviceAndHash(ctx, "dev3", "hash3")
		require.NoError(t, err)
		assert.Equal(t, "completed", updated.Status)

		err = repo.UpdateStatus(ctx, 999999, "completed")
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("UpdateResult", func(t *testing.T) {
		scan := &model.FirmwareScan{
			DeviceID:        "dev4",
			FirmwareVersion: "4.0.0",
			BinaryHash:      "hash4",
			Status:          "pending",
		}
		err := repo.Create(ctx, nil, scan)
		require.NoError(t, err)

		vulns := []string{"CVE-2021-1234", "CVE-2021-5678"}
		err = repo.UpdateResult(ctx, scan.ID, "completed", vulns)
		require.NoError(t, err)

		updated, err := repo.GetByDeviceAndHash(ctx, "dev4", "hash4")
		require.NoError(t, err)
		assert.Equal(t, "completed", updated.Status)
		
		var resultVulns []string
		err = json.Unmarshal(updated.Vulns, &resultVulns)
		require.NoError(t, err)
		assert.ElementsMatch(t, vulns, resultVulns)

		err = repo.UpdateResult(ctx, 999999, "completed", vulns)
		assert.ErrorIs(t, err, pgx.ErrNoRows)

		// Test with nil vulns
		err = repo.UpdateResult(ctx, scan.ID, "completed", nil)
		require.NoError(t, err)
	})

	t.Run("IsUniqueViolation", func(t *testing.T) {
		// We need to trigger a real unique violation.
		// firmware_scans has UNIQUE(device_id, binary_hash) but Create uses ON CONFLICT.
		// Let's try to insert directly with an error.
		_, err := testPool.Exec(ctx, "INSERT INTO firmware_scans (device_id, firmware_version, binary_hash, status) VALUES ('dev1', '1.0.0', 'hash1', 'pending')")
		assert.True(t, IsUniqueViolation(err))
		assert.False(t, IsUniqueViolation(nil))
		assert.False(t, IsUniqueViolation(assert.AnError))
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		err := repo.Create(cancelCtx, nil, &model.FirmwareScan{})
		assert.Error(t, err)

		_, err = repo.GetByDeviceAndHash(cancelCtx, "d", "h")
		assert.Error(t, err)

		err = repo.UpdateStatus(cancelCtx, 1, "s")
		assert.Error(t, err)

		err = repo.UpdateResult(cancelCtx, 1, "s", nil)
		assert.Error(t, err)
	})
}
