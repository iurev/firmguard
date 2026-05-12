package service

import (
	"context"
	"errors"
	"firmguard/internal/model"
	"firmguard/internal/repository"
	"firmguard/internal/worker"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Health() (map[string]string, error) { return nil, nil }
func (m *MockDB) Close() error                       { return nil }
func (m *MockDB) GetPool() *pgxpool.Pool             { return nil }
func (m *MockDB) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pgx.Tx), args.Error(1)
}

type MockTx struct {
	mock.Mock
	pgx.Tx
}

func (m *MockTx) Commit(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, tx pgx.Tx, scan *model.FirmwareScan) (repository.CreateResult, error) {
	args := m.Called(ctx, tx, scan)
	err := args.Error(1)
	if err == nil {
		scan.ID = 1
	}
	return args.Get(0).(repository.CreateResult), err
}

func (m *MockRepository) GetByID(ctx context.Context, id int) (*model.FirmwareScan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FirmwareScan), args.Error(1)
}

func (m *MockRepository) UpdateResult(ctx context.Context, id int, status string, vulns []string) error {
	args := m.Called(ctx, id, status, vulns)
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

func (m *MockRiverClient) InsertTx(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	callArgs := m.Called(ctx, tx, args, opts)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(*rivertype.JobInsertResult), callArgs.Error(1)
}

func TestCreateScan(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		db := new(MockDB)
		tx := new(MockTx)
		repo := new(MockRepository)
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(db, repo, riverClient)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		db.On("Begin", ctx).Return(tx, nil)
		repo.On("Create", ctx, tx, scan).Return(repository.CreateResult{Scan: scan, IsInserted: true}, nil)
		riverClient.On("InsertTx", ctx, tx, worker.FirmwareAnalysisArgs{ID: 1}, mock.Anything).Return(&rivertype.JobInsertResult{}, nil)
		tx.On("Commit", ctx).Return(nil)
		tx.On("Rollback", ctx).Return(nil)

		result, _, err := svc.CreateScan(ctx, scan)
		assert.NoError(t, err)
		assert.Equal(t, "pending", result.Status)

		repo.AssertExpectations(t)
		riverClient.AssertExpectations(t)
		db.AssertExpectations(t)
	})

	t.Run("already exists", func(t *testing.T) {
		db := new(MockDB)
		tx := new(MockTx)
		repo := new(MockRepository)
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(db, repo, riverClient)

		now := time.Now()
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}
		existing := &model.FirmwareScan{
			ID:        1,
			DeviceID:  "d1",
			BinaryHash: "h1",
			CreatedAt: now,
			UpdatedAt: now.Add(time.Second),
		}

		db.On("Begin", ctx).Return(tx, nil)
		repo.On("Create", ctx, tx, scan).Run(func(args mock.Arguments) {
			s := args.Get(2).(*model.FirmwareScan)
			*s = *existing
		}).Return(repository.CreateResult{Scan: existing, IsInserted: false}, nil)
		tx.On("Commit", ctx).Return(nil)
		tx.On("Rollback", ctx).Return(nil)

		result, _, err := svc.CreateScan(ctx, scan)
		assert.NoError(t, err)
		assert.Equal(t, existing, result)
	})

	t.Run("repo error on create", func(t *testing.T) {
		db := new(MockDB)
		tx := new(MockTx)
		repo := new(MockRepository)
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(db, repo, riverClient)
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		db.On("Begin", ctx).Return(tx, nil)
		repo.On("Create", ctx, tx, scan).Return(repository.CreateResult{}, errors.New("db error"))
		tx.On("Rollback", ctx).Return(nil)

		result, _, err := svc.CreateScan(ctx, scan)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("db begin error", func(t *testing.T) {
		db := new(MockDB)
		repo := new(MockRepository)
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(db, repo, riverClient)

		db.On("Begin", ctx).Return(nil, errors.New("begin error"))

		result, _, err := svc.CreateScan(ctx, &model.FirmwareScan{})
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("river insert error", func(t *testing.T) {
		db := new(MockDB)
		tx := new(MockTx)
		repo := new(MockRepository)
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(db, repo, riverClient)

		now := time.Now()
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		db.On("Begin", ctx).Return(tx, nil)
		repo.On("Create", ctx, tx, scan).Run(func(args mock.Arguments) {
			s := args.Get(2).(*model.FirmwareScan)
			s.ID = 1
			s.CreatedAt = now
			s.UpdatedAt = now
		}).Return(repository.CreateResult{Scan: scan, IsInserted: true}, nil)
		riverClient.On("InsertTx", ctx, tx, worker.FirmwareAnalysisArgs{ID: 1}, mock.Anything).Return(nil, errors.New("river error"))
		tx.On("Rollback", ctx).Return(nil)

		result, _, err := svc.CreateScan(ctx, scan)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("tx commit error", func(t *testing.T) {
		db := new(MockDB)
		tx := new(MockTx)
		repo := new(MockRepository)
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(db, repo, riverClient)

		now := time.Now()
		scan := &model.FirmwareScan{DeviceID: "d1", BinaryHash: "h1"}

		db.On("Begin", ctx).Return(tx, nil)
		repo.On("Create", ctx, tx, scan).Run(func(args mock.Arguments) {
			s := args.Get(2).(*model.FirmwareScan)
			s.ID = 1
			s.CreatedAt = now
			s.UpdatedAt = now
		}).Return(repository.CreateResult{Scan: scan, IsInserted: true}, nil)
		riverClient.On("InsertTx", ctx, tx, worker.FirmwareAnalysisArgs{ID: 1}, mock.Anything).Return(&rivertype.JobInsertResult{}, nil)
		tx.On("Commit", ctx).Return(errors.New("commit error"))
		tx.On("Rollback", ctx).Return(nil)

		result, _, err := svc.CreateScan(ctx, scan)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetScan(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		db := new(MockDB)
		repo := new(MockRepository)
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(db, repo, riverClient)

		expected := &model.FirmwareScan{ID: 1, DeviceID: "d1"}
		repo.On("GetByID", ctx, 1).Return(expected, nil)

		result, err := svc.GetScan(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		repo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		db := new(MockDB)
		repo := new(MockRepository)
		riverClient := new(MockRiverClient)
		svc := NewFirmwareScanService(db, repo, riverClient)

		repo.On("GetByID", ctx, 999).Return(nil, pgx.ErrNoRows)

		result, err := svc.GetScan(ctx, 999)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})
}
