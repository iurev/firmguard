package worker

import (
	"context"
	"firmguard/internal/model"
	"os"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, scan *model.FirmwareScan) error {
	args := m.Called(ctx, scan)
	return args.Error(0)
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

func TestFirmwareAnalysisArgs_Kind(t *testing.T) {
	args := FirmwareAnalysisArgs{}
	assert.Equal(t, "firmware_analysis", args.Kind())
}

func TestNewFirmwareAnalysisWorker(t *testing.T) {
	repo := new(MockRepository)
	worker := NewFirmwareAnalysisWorker(repo)
	assert.NotNil(t, worker)
	assert.Equal(t, repo, worker.repo)
	assert.Equal(t, 2*time.Second, worker.delay)
}

func TestFirmwareAnalysisWorker_Work(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		worker := NewFirmwareAnalysisWorker(repo)
		worker.delay = 0 // Fast test
		job := &river.Job[FirmwareAnalysisArgs]{
			Args: FirmwareAnalysisArgs{ID: 1},
		}

		repo.On("UpdateStatus", mock.Anything, 1, "completed").Return(nil)

		err := worker.Work(ctx, job)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("context cancelled", func(t *testing.T) {
		repo := new(MockRepository)
		worker := NewFirmwareAnalysisWorker(repo)
		worker.delay = 100 * time.Millisecond
		job := &river.Job[FirmwareAnalysisArgs]{
			Args: FirmwareAnalysisArgs{ID: 1},
		}

		cancelCtx, cancel := context.WithCancel(ctx)
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		err := worker.Work(cancelCtx, job)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("simulate failure", func(t *testing.T) {
		os.Setenv("SIMULATE_FAILURE", "true")
		defer os.Unsetenv("SIMULATE_FAILURE")

		hitCompleted := false
		hitFailed := false

		for i := 0; i < 100 && (!hitCompleted || !hitFailed); i++ {
			repo := new(MockRepository)
			worker := NewFirmwareAnalysisWorker(repo)
			worker.delay = 0
			job := &river.Job[FirmwareAnalysisArgs]{
				Args: FirmwareAnalysisArgs{ID: 1},
			}

			var capturedStatus string
			repo.On("UpdateStatus", mock.Anything, 1, mock.Anything).Run(func(args mock.Arguments) {
				capturedStatus = args.String(2)
				if capturedStatus == "completed" {
					hitCompleted = true
				} else if capturedStatus == "failed" {
					hitFailed = true
				}
			}).Return(nil)

			err := worker.Work(ctx, job)
			assert.NoError(t, err)
		}

		assert.True(t, hitCompleted, "should have hit completed status")
		assert.True(t, hitFailed, "should have hit failed status")
	})
}
