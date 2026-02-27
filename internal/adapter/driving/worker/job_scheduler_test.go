package worker_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/st-ember/streaming-api/internal/adapter/driving/worker"
	mockjob "github.com/st-ember/streaming-api/internal/application/jobapp/mocks"
	mocklog "github.com/st-ember/streaming-api/internal/application/ports/log/mocks"
	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestJobScheduler_Run(t *testing.T) {
	t.Run("should shut down gracefully on context cancellation", func(t *testing.T) {
		findNextUC := mockjob.NewMockFindNextPendingTranscodeJobUsecase(t)
		logger := mocklog.NewMockLogger(t)
		jobCh := make(chan *job.Job, 1)

		ctx, cancel := context.WithCancel(t.Context())
		s := worker.NewJobScheduler(findNextUC, logger, jobCh, 10*time.Millisecond)

		logger.EXPECT().Infof("starting worker pool").Once()
		logger.EXPECT().Infof("shutting down worker pool").Once()

		findNextUC.EXPECT().Execute(mock.Anything).Return(nil, nil).Maybe()

		done := make(chan struct{})
		go func() {
			s.Run(ctx)
			close(done)
		}()

		time.Sleep(20 * time.Millisecond)
		cancel()

		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Fatal("JobScheduler did not shut down in time")
		}
	})

	t.Run("should find and queue job", func(t *testing.T) {
		findNextUC := mockjob.NewMockFindNextPendingTranscodeJobUsecase(t)
		logger := mocklog.NewMockLogger(t)
		jobCh := make(chan *job.Job, 1)

		s := worker.NewJobScheduler(findNextUC, logger, jobCh, 10*time.Millisecond)

		testJob, _ := job.NewJob("job-1", "video-1", job.TypeTranscode)

		logger.EXPECT().Infof("starting worker pool").Once()
		findNextUC.EXPECT().Execute(mock.Anything).Return(testJob, nil).Once()
		logger.EXPECT().Infof("job %s is added to queue", mock.Anything).Once()

		// Setup expectations for subsequent iterations to avoid noise or allow shutdown
		findNextUC.EXPECT().Execute(mock.Anything).Return(nil, nil).Maybe()
		logger.EXPECT().Infof("shutting down worker pool").Maybe()

		go s.Run(t.Context())

		select {
		case queuedJob := <-jobCh:
			require.Equal(t, testJob.ID, queuedJob.ID)
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Job was not queued in time")
		}
	})

	t.Run("should continue when no jobs are found", func(t *testing.T) {
		findNextUC := mockjob.NewMockFindNextPendingTranscodeJobUsecase(t)
		logger := mocklog.NewMockLogger(t)
		jobCh := make(chan *job.Job, 1)

		s := worker.NewJobScheduler(findNextUC, logger, jobCh, 10*time.Millisecond)

		logger.EXPECT().Infof("starting worker pool").Once()
		findNextUC.EXPECT().Execute(mock.Anything).Return(nil, sql.ErrNoRows).Once()
		findNextUC.EXPECT().Execute(mock.Anything).Return(nil, nil).Maybe()
		logger.EXPECT().Infof("shutting down worker pool").Maybe()

		go s.Run(t.Context())

		time.Sleep(50 * time.Millisecond)
		require.Equal(t, 0, len(jobCh))
	})

	t.Run("should log error and continue when finding job fails", func(t *testing.T) {
		findNextUC := mockjob.NewMockFindNextPendingTranscodeJobUsecase(t)
		logger := mocklog.NewMockLogger(t)
		jobCh := make(chan *job.Job, 1)

		s := worker.NewJobScheduler(findNextUC, logger, jobCh, 10*time.Millisecond)

		logger.EXPECT().Infof("starting worker pool").Once()
		findNextUC.EXPECT().Execute(mock.Anything).Return(nil, errors.New("db error")).Once()
		logger.EXPECT().Errorf(mock.Anything, mock.Anything).Once()

		findNextUC.EXPECT().Execute(mock.Anything).Return(nil, nil).Maybe()
		logger.EXPECT().Infof("shutting down worker pool").Maybe()

		go s.Run(t.Context())

		time.Sleep(50 * time.Millisecond)
	})

	t.Run("should log when queue is full", func(t *testing.T) {
		findNextUC := mockjob.NewMockFindNextPendingTranscodeJobUsecase(t)
		logger := mocklog.NewMockLogger(t)
		jobCh := make(chan *job.Job, 1)

		s := worker.NewJobScheduler(findNextUC, logger, jobCh, 10*time.Millisecond)

		testJob1, _ := job.NewJob("job-1", "video-1", job.TypeTranscode)
		testJob2, _ := job.NewJob("job-2", "video-1", job.TypeTranscode)

		// Fill the channel
		jobCh <- testJob1

		logger.EXPECT().Infof("starting worker pool").Once()
		findNextUC.EXPECT().Execute(mock.Anything).Return(testJob2, nil).Once()
		logger.EXPECT().Infof("job queue full now, will try again in %v seconds", mock.Anything).Once()

		findNextUC.EXPECT().Execute(mock.Anything).Return(nil, nil).Maybe()
		logger.EXPECT().Infof("shutting down worker pool").Maybe()

		go s.Run(t.Context())

		time.Sleep(50 * time.Millisecond)
	})
}
