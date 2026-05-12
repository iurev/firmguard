package worker

import (
	"context"
	"log"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type SentryMock struct{}

func (s *SentryMock) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	if job.Attempt >= job.MaxAttempts {
		log.Printf("[SentryMock] final attempt %d/%d failed for job %q: %v", job.Attempt, job.MaxAttempts, job.Kind, err)
	}
	return &river.ErrorHandlerResult{}
}

func (s *SentryMock) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, trace string) *river.ErrorHandlerResult {
	log.Printf("[SentryMock] panic in job %q (attempt %d): %v\n%s", job.Kind, job.Attempt, panicVal, trace)
	return &river.ErrorHandlerResult{}
}
