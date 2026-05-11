package service

import (
	"context"
	"errors"
	"firmguard/internal/model"
	"firmguard/internal/worker"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
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

type MockRiverClient struct {
	mock.Mock
}

func (m *MockRiverClient) Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	callArgs := m.Called(ctx, args, opts)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(*rivertype.JobInsertResult), callArgs.Error(1)
}

func TestCreateScan(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(repo, riverClient)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		repo.On("Create", ctx, scan).Return(nil)
		riverClient.On("Insert", ctx, worker.FirmwareAnalysisArgs{ID: 1}, mock.Anything).Return(&rivertype.JobInsertResult{}, nil)

		result, err := svc.CreateScan(ctx, scan)
		assert.NoError(t, err)
		assert.Equal(t, "pending", result.Status)

		repo.AssertExpectations(t)
		riverClient.AssertExpectations(t)
	})

	t.Run("already exists", func(t *testing.T) {
		repo := new(MockRepository)
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(repo, riverClient)
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
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(repo, riverClient)
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
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(repo, riverClient)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		repo.On("Create", ctx, scan).Return(errors.New("db error"))

		result, err := svc.CreateScan(ctx, scan)
		assert.Error(t, err)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("river error", func(t *testing.T) {
		repo := new(MockRepository)
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(repo, riverClient)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		repo.On("Create", ctx, scan).Return(nil)
		riverClient.On("Insert", ctx, worker.FirmwareAnalysisArgs{ID: 1}, mock.Anything).Return(nil, errors.New("river error"))

		result, err := svc.CreateScan(ctx, scan)
		assert.Error(t, err)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})
}
