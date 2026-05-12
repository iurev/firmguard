package worker

import (
	"context"
	"firmguard/internal/model"
	"firmguard/internal/repository"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
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

func (m *MockScanRepository) GetByID(ctx context.Context, id int) (*model.FirmwareScan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FirmwareScan), args.Error(1)
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
	w := NewFirmwareAnalysisWorker(scanRepo, vulnRepo)
	assert.NotNil(t, w)
	assert.Equal(t, scanRepo, w.scanRepo)
	assert.Equal(t, vulnRepo, w.vulnRepo)
}

func TestFirmwareAnalysisWorker_Work(t *testing.T) {
	ctx := context.Background()

	newJob := func() *river.Job[FirmwareAnalysisArgs] {
		return &river.Job[FirmwareAnalysisArgs]{
			JobRow: &rivertype.JobRow{Attempt: 1, MaxAttempts: 3},
			Args:   FirmwareAnalysisArgs{ID: 1},
		}
	}

	setupWorker := func() (*FirmwareAnalysisWorker, *MockScanRepository, *MockVulnRepository) {
		scanRepo := new(MockScanRepository)
		vulnRepo := new(MockVulnRepository)
		w := NewFirmwareAnalysisWorker(scanRepo, vulnRepo)
		w.sleepFunc = func(time.Duration) {}
		return w, scanRepo, vulnRepo
	}

	t.Run("outcome fail - non-final attempt", func(t *testing.T) {
		w, scanRepo, vulnRepo := setupWorker()
		w.randFunc = func(n int) int {
			if n == 100 {
				return 20 // < 30
			}
			return 0
		}

		err := w.Work(ctx, newJob())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated temporary failure")
		scanRepo.AssertExpectations(t)
		vulnRepo.AssertExpectations(t)
	})

	t.Run("outcome fail - final attempt marks scan as failed", func(t *testing.T) {
		w, scanRepo, vulnRepo := setupWorker()
		w.randFunc = func(n int) int {
			if n == 100 {
				return 20 // < 30
			}
			return 0
		}

		finalJob := &river.Job[FirmwareAnalysisArgs]{
			JobRow: &rivertype.JobRow{Attempt: 3, MaxAttempts: 3},
			Args:   FirmwareAnalysisArgs{ID: 1},
		}
		scanRepo.On("UpdateResult", mock.Anything, 1, "failed", mock.Anything).Return(nil)

		err := w.Work(ctx, finalJob)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated temporary failure")
		scanRepo.AssertExpectations(t)
		vulnRepo.AssertExpectations(t)
	})

	t.Run("outcome found", func(t *testing.T) {
		w, scanRepo, vulnRepo := setupWorker()
		w.randFunc = func(n int) int {
			if n == 100 {
				return 45 // 30-59
			}
			return 0
		}

		vulnRepo.On("GetRandom", mock.Anything).Return("CVE-123", nil)
		scanRepo.On("UpdateResult", mock.Anything, 1, "completed", []string{"CVE-123"}).Return(nil)

		err := w.Work(ctx, newJob())
		assert.NoError(t, err)
		scanRepo.AssertExpectations(t)
		vulnRepo.AssertExpectations(t)
	})

	t.Run("outcome not found", func(t *testing.T) {
		w, scanRepo, vulnRepo := setupWorker()
		w.randFunc = func(n int) int {
			if n == 100 {
				return 80 // >= 60
			}
			return 0
		}

		scanRepo.On("UpdateResult", mock.Anything, 1, "completed", []string(nil)).Return(nil)

		err := w.Work(ctx, newJob())
		assert.NoError(t, err)
		scanRepo.AssertExpectations(t)
		vulnRepo.AssertExpectations(t)
	})

	t.Run("outcome found but no CVE in db", func(t *testing.T) {
		w, scanRepo, vulnRepo := setupWorker()
		w.randFunc = func(n int) int {
			if n == 100 {
				return 45
			}
			return 0
		}

		vulnRepo.On("GetRandom", mock.Anything).Return("", nil)
		scanRepo.On("UpdateResult", mock.Anything, 1, "completed", []string(nil)).Return(nil)

		err := w.Work(ctx, newJob())
		assert.NoError(t, err)
		scanRepo.AssertExpectations(t)
		vulnRepo.AssertExpectations(t)
	})
}
