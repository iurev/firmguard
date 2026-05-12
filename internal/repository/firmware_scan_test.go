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

		res, err := repo.Create(ctx, nil, scan)
		require.NoError(t, err)
		assert.True(t, res.IsInserted)
		assert.NotZero(t, res.Scan.ID)
		assert.NotEmpty(t, res.Scan.CreatedAt)
		assert.NotEmpty(t, res.Scan.UpdatedAt)

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

			res, err = repo.Create(ctx, tx, scan2)
			require.NoError(t, err)
			assert.True(t, res.IsInserted)
			assert.NotZero(t, res.Scan.ID)

			err = tx.Commit(ctx)
			require.NoError(t, err)
		})

		t.Run("DuplicateEntry", func(t *testing.T) {
			scanDup := &model.FirmwareScan{
				DeviceID:        "dev1",
				FirmwareVersion: "1.0.0",
				BinaryHash:      "hash1",
				Status:          "pending",
			}
			res, err := repo.Create(ctx, nil, scanDup)
			require.NoError(t, err)
			assert.False(t, res.IsInserted)
		})
	})

	t.Run("GetByID", func(t *testing.T) {
		scan := &model.FirmwareScan{
			DeviceID:        "dev-getbyid",
			FirmwareVersion: "1.0.0",
			BinaryHash:      "hash-getbyid",
			Status:          "pending",
		}
		res, err := repo.Create(ctx, nil, scan)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, res.Scan.ID)
		require.NoError(t, err)
		assert.Equal(t, res.Scan.ID, found.ID)
		assert.Equal(t, "dev-getbyid", found.DeviceID)

		_, err = repo.GetByID(ctx, 999999)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("GetByDeviceAndHash", func(t *testing.T) {
		scan, err := repo.GetByDeviceAndHash(ctx, "dev1", "hash1")
		require.NoError(t, err)
		assert.Equal(t, "dev1", scan.DeviceID)
		assert.Equal(t, "hash1", scan.BinaryHash)

		_, err = repo.GetByDeviceAndHash(ctx, "nonexistent", "hash")
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("UpdateResult", func(t *testing.T) {
		scan := &model.FirmwareScan{
			DeviceID:        "dev4",
			FirmwareVersion: "4.0.0",
			BinaryHash:      "hash4",
			Status:          "pending",
		}
		_, err := repo.Create(ctx, nil, scan)
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

		err = repo.UpdateResult(ctx, scan.ID, "completed", nil)
		require.NoError(t, err)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		_, err := repo.Create(cancelCtx, nil, &model.FirmwareScan{})
		assert.Error(t, err)

		_, err = repo.GetByID(cancelCtx, 1)
		assert.Error(t, err)

		_, err = repo.GetByDeviceAndHash(cancelCtx, "d", "h")
		assert.Error(t, err)

		err = repo.UpdateResult(cancelCtx, 1, "completed", nil)
		assert.Error(t, err)
	})
}
