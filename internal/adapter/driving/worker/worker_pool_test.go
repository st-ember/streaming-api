package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/st-ember/streaming-api/internal/adapter/driving/worker"
	"github.com/st-ember/streaming-api/internal/application/jobapp"
	mockjob "github.com/st-ember/streaming-api/internal/application/jobapp/mocks"
	mocklog "github.com/st-ember/streaming-api/internal/application/ports/log/mocks"
	mockstorage "github.com/st-ember/streaming-api/internal/application/ports/storage/mocks"
	"github.com/st-ember/streaming-api/internal/application/ports/transcode"
	mocktranscode "github.com/st-ember/streaming-api/internal/application/ports/transcode/mocks"
	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/stretchr/testify/mock"
)

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	// Setup mocks
	findNextUC := mockjob.NewMockFindNextPendingTranscodeJobUsecase(t)
	startUC := mockjob.NewMockStartTranscodeJobUsecase(t)
	completeUC := mockjob.NewMockCompleteTranscodeJobUsecase(t)
	failUC := mockjob.NewMockFailTranscodeJobUsecase(t)
	storer := mockstorage.NewMockAssetStorer(t)
	logger := mocklog.NewMockLogger(t)
	transcoder := mocktranscode.NewMockTranscoder(t)

	// Create pool with 1 worker
	p := worker.NewWorkerPool(
		findNextUC, startUC, completeUC, failUC,
		storer, logger, transcoder, 2, 1,
	)

	ctx, cancel := context.WithCancel(context.Background())

	// Test Data
	testJob, _ := job.NewJob("job-1", "video-1", job.TypeTranscode)

	// Expectations
	logger.EXPECT().Infof("job scheduler started").Once()

	// Scheduler: returns one job, then we'll cancel context during the next poll
	findNextUC.EXPECT().Execute(mock.Anything).Return(testJob, nil).Once()

	// Signal when job processing starts
	jobProcessingStarted := make(chan struct{})

	startUC.EXPECT().Execute(mock.Anything, testJob).Run(func(ctx context.Context, j *job.Job) {
		close(jobProcessingStarted)
		// Simulate work that takes time. The pool MUST wait for this to finish.
		time.Sleep(100 * time.Millisecond)
	}).Return(&jobapp.StartTranscodeJobResult{ResourceID: "res-1", SourceFilename: "in.mp4"}, nil).Once()

	// Minimal transcode success to reach completion
	transcoder.EXPECT().Transcode(mock.Anything, "res-1", "in.mp4").Return(&transcode.TranscodeOutput{
		Duration:     10 * time.Second,
		ManifestPath: "/tmp/fake/manifest.m3u8",
		OutputFiles:  []string{},
	}, nil).Once()

	completeUC.EXPECT().Execute(mock.Anything, testJob, "/tmp/fake/manifest.m3u8", 10*time.Second).Return(nil).Once()

	// Subsequent scheduler poll triggers the context cancellation
	findNextUC.EXPECT().Execute(mock.Anything).Run(func(ctx context.Context) {
		cancel()
	}).Return(nil, nil).Maybe()

	logger.EXPECT().Infof("job scheduler shutting down").Once()
	logger.EXPECT().Infof(mock.Anything, mock.Anything).Maybe()
	logger.EXPECT().Infof(mock.Anything).Maybe() // Catch "worker finished..." or other info logs

	// Start Pool
	p.Start(ctx)

	// Wait until we know the worker has picked up the job
	<-jobProcessingStarted

	// Shutdown sequence starts
	done := make(chan struct{})
	go func() {
		p.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success: Wait() returned only after the 100ms job finished
	case <-time.After(3 * time.Second):
		t.Fatal("WorkerPool.Wait() timed out; did not shut down gracefully")
	}
}
