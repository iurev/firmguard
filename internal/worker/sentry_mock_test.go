package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/assert"
)

func TestSentryMock_HandleError(t *testing.T) {
	s := &SentryMock{}
	ctx := context.Background()
	theErr := errors.New("something failed")

	t.Run("non-final attempt returns empty result", func(t *testing.T) {
		job := &rivertype.JobRow{Attempt: 1, MaxAttempts: 3, Kind: "firmware_analysis"}
		result := s.HandleError(ctx, job, theErr)
		assert.Equal(t, &river.ErrorHandlerResult{}, result)
	})

	t.Run("final attempt returns empty result", func(t *testing.T) {
		job := &rivertype.JobRow{Attempt: 3, MaxAttempts: 3, Kind: "firmware_analysis"}
		result := s.HandleError(ctx, job, theErr)
		assert.Equal(t, &river.ErrorHandlerResult{}, result)
	})
}

func TestSentryMock_HandlePanic(t *testing.T) {
	s := &SentryMock{}
	ctx := context.Background()
	job := &rivertype.JobRow{Attempt: 1, MaxAttempts: 3, Kind: "firmware_analysis"}

	result := s.HandlePanic(ctx, job, "index out of range", "goroutine 1 [running]:\n...")
	assert.Equal(t, &river.ErrorHandlerResult{}, result)
}
