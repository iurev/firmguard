package worker

import (
	"context"
	"firmguard/internal/model"
	"firmguard/internal/repository"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockScanRepository struct {
	mock.Mock
}

func (m *MockScanRepository) Create(ctx context.Context, tx pgx.Tx, scan *model.FirmwareScan) (repository.CreateResult, error) {
	args := m.Called(ctx, tx, scan)
	return args.Get(0).(repository.CreateResult), args.Error(1)
}

func (m *MockScanRepository) GetByDeviceAndHash(ctx context.Context, deviceID, hash string) (*model.FirmwareScan, error) {
	args := m.Called(ctx, deviceID, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FirmwareScan), args.Error(1)
}

func (m *MockScanRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockScanRepository) UpdateResult(ctx context.Context, id int, status string, vulns []string) error {
	args := m.Called(ctx, id, status, vulns)
	return args.Error(0)
}

type MockVulnRepository struct {
	mock.Mock
}

func (m *MockVulnRepository) Upsert(ctx context.Context, cveIDs []string) error {
	args := m.Called(ctx, cveIDs)
	return args.Error(0)
}

func (m *MockVulnRepository) List(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockVulnRepository) GetRandom(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func TestFirmwareAnalysisArgs_Kind(t *testing.T) {
	args := FirmwareAnalysisArgs{}
	assert.Equal(t, "firmware_analysis", args.Kind())
}

func TestNewFirmwareAnalysisWorker(t *testing.T) {
	scanRepo := new(MockScanRepository)
	vulnRepo := new(MockVulnRepository)
	worker := NewFirmwareAnalysisWorker(scanRepo, vulnRepo)
	assert.NotNil(t, worker)
	assert.Equal(t, scanRepo, worker.scanRepo)
	assert.Equal(t, vulnRepo, worker.vulnRepo)
}

func TestFirmwareAnalysisWorker_Work(t *testing.T) {
	ctx := context.Background()

	setupWorker := func() (*FirmwareAnalysisWorker, *MockScanRepository, *MockVulnRepository) {
		scanRepo := new(MockScanRepository)
		vulnRepo := new(MockVulnRepository)
		worker := NewFirmwareAnalysisWorker(scanRepo, vulnRepo)
		worker.sleepFunc = func(time.Duration) {} // No sleep in tests
		return worker, scanRepo, vulnRepo
	}

	t.Run("outcome fail", func(t *testing.T) {
		os.Setenv("SIMULATE_FAILURE", "true")
		defer os.Unsetenv("SIMULATE_FAILURE")
		
		worker, scanRepo, vulnRepo := setupWorker()
		worker.randFunc = func(n int) int {
			if n == 100 {
				return 20 // < 30
			}
			return 0
		}
		job := &river.Job[FirmwareAnalysisArgs]{Args: FirmwareAnalysisArgs{ID: 1}}

		err := worker.Work(ctx, job)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated temporary failure")
		scanRepo.AssertExpectations(t)
		vulnRepo.AssertExpectations(t)
	})

	t.Run("outcome found", func(t *testing.T) {
		worker, scanRepo, vulnRepo := setupWorker()
		worker.randFunc = func(n int) int {
			if n == 100 {
				return 45 // 30-59
			}
			return 0
		}
		job := &river.Job[FirmwareAnalysisArgs]{Args: FirmwareAnalysisArgs{ID: 1}}

		vulnRepo.On("GetRandom", mock.Anything).Return("CVE-123", nil)
		scanRepo.On("UpdateResult", mock.Anything, 1, "completed", []string{"CVE-123"}).Return(nil)

		err := worker.Work(ctx, job)
		assert.NoError(t, err)
		scanRepo.AssertExpectations(t)
		vulnRepo.AssertExpectations(t)
	})

	t.Run("outcome not found", func(t *testing.T) {
		worker, scanRepo, vulnRepo := setupWorker()
		worker.randFunc = func(n int) int {
			if n == 100 {
				return 80 // >= 60
			}
			return 0
		}
		job := &river.Job[FirmwareAnalysisArgs]{Args: FirmwareAnalysisArgs{ID: 1}}

		scanRepo.On("UpdateResult", mock.Anything, 1, "completed", []string(nil)).Return(nil)

		err := worker.Work(ctx, job)
		assert.NoError(t, err)
		scanRepo.AssertExpectations(t)
		vulnRepo.AssertExpectations(t)
	})

	t.Run("outcome found but no CVE in db", func(t *testing.T) {
		worker, scanRepo, vulnRepo := setupWorker()
		worker.randFunc = func(n int) int {
			if n == 100 {
				return 45
			}
			return 0
		}
		job := &river.Job[FirmwareAnalysisArgs]{Args: FirmwareAnalysisArgs{ID: 1}}

		vulnRepo.On("GetRandom", mock.Anything).Return("", nil)
		scanRepo.On("UpdateResult", mock.Anything, 1, "completed", []string(nil)).Return(nil)

		err := worker.Work(ctx, job)
		assert.NoError(t, err)
		scanRepo.AssertExpectations(t)
		vulnRepo.AssertExpectations(t)
	})
}
