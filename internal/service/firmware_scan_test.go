package service

import (
	"context"
	"database/sql"
	"errors"
	"firmguard/internal/model"
	"testing"

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

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewFirmwareScanService(repo)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		repo.On("GetByDeviceAndHash", ctx, "d1", "h1").Return(nil, sql.ErrNoRows)
		repo.On("Create", ctx, scan).Return(nil)

		result, err := svc.CreateScan(ctx, scan)
		assert.NoError(t, err)
		assert.Equal(t, "pending", result.Status)
		repo.AssertExpectations(t)
	})

	t.Run("already exists", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewFirmwareScanService(repo)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}
		existing := &model.FirmwareScan{ID: 1, DeviceID: "d1", BinaryHash: "h1"}

		repo.On("GetByDeviceAndHash", ctx, "d1", "h1").Return(existing, nil)

		result, err := svc.CreateScan(ctx, scan)
		assert.ErrorIs(t, err, ErrScanAlreadyExists)
		assert.Equal(t, existing, result)
		repo.AssertExpectations(t)
	})

	t.Run("repo error on get", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewFirmwareScanService(repo)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

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

		repo.On("GetByDeviceAndHash", ctx, "d1", "h1").Return(nil, sql.ErrNoRows)
		repo.On("Create", ctx, scan).Return(errors.New("db error"))

		result, err := svc.CreateScan(ctx, scan)
		assert.Error(t, err)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})
}
