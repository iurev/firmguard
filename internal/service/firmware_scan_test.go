package service

import (
	"context"
	"errors"
	"firmguard/internal/model"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, scan *model.FirmwareScan) error {
	args := m.Called(ctx, scan)
	err := args.Error(0)
	if err == nil {
		scan.ID = 1
	}
	return err
}

func (m *MockRepository) GetByDeviceAndHash(ctx context.Context, deviceID, hash string) (*model.FirmwareScan, error) {
	args := m.Called(ctx, deviceID, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FirmwareScan), args.Error(1)
}

func (m *MockRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func TestCreateScan(t *testing.T) {
	ctx := context.Background()
	// Speed up background worker for tests
	originalDelay := scanWorkerDelay
	scanWorkerDelay = 0
	defer func() { scanWorkerDelay = originalDelay }()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewFirmwareScanService(repo)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		repo.On("Create", ctx, scan).Return(nil)
		repo.On("UpdateStatus", mock.Anything, 1, "completed").Return(nil)

		result, err := svc.CreateScan(ctx, scan)
		assert.NoError(t, err)
		assert.Equal(t, "pending", result.Status)
		
		// Wait for goroutine
		time.Sleep(10 * time.Millisecond)
		repo.AssertExpectations(t)
	})

	t.Run("already exists", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewFirmwareScanService(repo)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}
		existing := &model.FirmwareScan{ID: 1, DeviceID: "d1", BinaryHash: "h1"}

		uniqueErr := &pgconn.PgError{Code: "23505"}
		repo.On("Create", ctx, scan).Return(uniqueErr)
		repo.On("GetByDeviceAndHash", ctx, "d1", "h1").Return(existing, nil)

		result, err := svc.CreateScan(ctx, scan)
		assert.ErrorIs(t, err, ErrScanAlreadyExists)
		assert.Equal(t, existing, result)
		repo.AssertExpectations(t)
	})

	t.Run("repo error on get after unique violation", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewFirmwareScanService(repo)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		uniqueErr := &pgconn.PgError{Code: "23505"}
		repo.On("Create", ctx, scan).Return(uniqueErr)
		repo.On("GetByDeviceAndHash", ctx, "d1", "h1").Return(nil, errors.New("db error"))

		result, err := svc.CreateScan(ctx, scan)
		assert.Error(t, err)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("repo error on create", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewFirmwareScanService(repo)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		repo.On("Create", ctx, scan).Return(errors.New("db error"))

		result, err := svc.CreateScan(ctx, scan)
		assert.Error(t, err)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("simulate failure branch - failed", func(t *testing.T) {
		os.Setenv("SIMULATE_FAILURE", "true")
		defer os.Unsetenv("SIMULATE_FAILURE")
		
		originalRand := randIntn
		randIntn = func(n int) int { return 40 } // < 50
		defer func() { randIntn = originalRand }()

		repo := new(MockRepository)
		svc := NewFirmwareScanService(repo)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		repo.On("Create", ctx, scan).Return(nil)
		repo.On("UpdateStatus", mock.Anything, 1, "failed").Return(nil)

		_, err := svc.CreateScan(ctx, scan)
		assert.NoError(t, err)
		
		time.Sleep(10 * time.Millisecond)
		repo.AssertExpectations(t)
	})

	t.Run("simulate failure branch - completed", func(t *testing.T) {
		os.Setenv("SIMULATE_FAILURE", "true")
		defer os.Unsetenv("SIMULATE_FAILURE")
		
		originalRand := randIntn
		randIntn = func(n int) int { return 60 } // >= 50
		defer func() { randIntn = originalRand }()

		repo := new(MockRepository)
		svc := NewFirmwareScanService(repo)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		repo.On("Create", ctx, scan).Return(nil)
		repo.On("UpdateStatus", mock.Anything, 1, "completed").Return(nil)

		_, err := svc.CreateScan(ctx, scan)
		assert.NoError(t, err)
		
		time.Sleep(10 * time.Millisecond)
		repo.AssertExpectations(t)
	})
}
