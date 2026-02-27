package worker_test

import (
	"errors"
	"os"
	"path/filepath"
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
	"github.com/stretchr/testify/require"
)

func TestTranscodeWorker_Start(t *testing.T) {
	t.Run("successful transcode workflow", func(t *testing.T) {
		startUC := mockjob.NewMockStartTranscodeJobUsecase(t)
		completeUC := mockjob.NewMockCompleteTranscodeJobUsecase(t)
		failUC := mockjob.NewMockFailTranscodeJobUsecase(t)
		storer := mockstorage.NewMockAssetStorer(t)
		logger := mocklog.NewMockLogger(t)
		transcoder := mocktranscode.NewMockTranscoder(t)
		jobCh := make(chan *job.Job, 1)

		w := worker.NewTranscodeWorker(startUC, completeUC, failUC, storer, logger, transcoder, jobCh)

		testJob, _ := job.NewJob("job-1", "video-1", job.TypeTranscode)
		resourceID := "res-1"
		sourceFile := "input.mp4"

		tempDir := t.TempDir()
		manifestName := "manifest.m3u8"
		segmentName := "seg1.ts"

		err := os.WriteFile(filepath.Join(tempDir, manifestName), []byte("manifest content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, segmentName), []byte("segment content"), 0644)
		require.NoError(t, err)

		startUC.EXPECT().Execute(mock.Anything, testJob).Return(&jobapp.StartTranscodeJobResponse{
			ResourceID:     resourceID,
			SourceFilename: sourceFile,
		}, nil)

		transcoder.EXPECT().Transcode(mock.Anything, resourceID, sourceFile).Return(&transcode.TranscodeOutput{
			Duration:     10 * time.Second,
			ManifestPath: filepath.Join(tempDir, manifestName),
			OutputFiles:  []string{manifestName, segmentName},
		}, nil)

		storer.EXPECT().Save(mock.Anything, resourceID, manifestName, mock.Anything).Return(nil)
		storer.EXPECT().Save(mock.Anything, resourceID, segmentName, mock.Anything).Return(nil)

		completeUC.EXPECT().Execute(mock.Anything, testJob, filepath.Join(tempDir, manifestName), 10*time.Second).Return(nil)
		logger.EXPECT().Infof(mock.Anything, mock.Anything).Maybe()

		go w.Start(t.Context())
		jobCh <- testJob
		close(jobCh)

		time.Sleep(100 * time.Millisecond)
		_, err = os.Stat(tempDir)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("should continue if starting job fails", func(t *testing.T) {
		startUC := mockjob.NewMockStartTranscodeJobUsecase(t)
		completeUC := mockjob.NewMockCompleteTranscodeJobUsecase(t)
		failUC := mockjob.NewMockFailTranscodeJobUsecase(t)
		storer := mockstorage.NewMockAssetStorer(t)
		logger := mocklog.NewMockLogger(t)
		transcoder := mocktranscode.NewMockTranscoder(t)
		jobCh := make(chan *job.Job, 1)

		w := worker.NewTranscodeWorker(startUC, completeUC, failUC, storer, logger, transcoder, jobCh)

		testJob, _ := job.NewJob("job-1", "video-1", job.TypeTranscode)

		startUC.EXPECT().Execute(mock.Anything, testJob).Return(nil, errors.New("start failed"))
		logger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything).Once()
		logger.EXPECT().Infof(mock.Anything, mock.Anything).Maybe()

		go w.Start(t.Context())
		jobCh <- testJob
		close(jobCh)

		time.Sleep(50 * time.Millisecond)
	})

	t.Run("should mark as failed if transcoding fails", func(t *testing.T) {
		startUC := mockjob.NewMockStartTranscodeJobUsecase(t)
		completeUC := mockjob.NewMockCompleteTranscodeJobUsecase(t)
		failUC := mockjob.NewMockFailTranscodeJobUsecase(t)
		storer := mockstorage.NewMockAssetStorer(t)
		logger := mocklog.NewMockLogger(t)
		transcoder := mocktranscode.NewMockTranscoder(t)
		jobCh := make(chan *job.Job, 1)

		w := worker.NewTranscodeWorker(startUC, completeUC, failUC, storer, logger, transcoder, jobCh)

		testJob, _ := job.NewJob("job-1", "video-1", job.TypeTranscode)
		resourceID := "res-1"
		sourceFile := "input.mp4"

		startUC.EXPECT().Execute(mock.Anything, testJob).Return(&jobapp.StartTranscodeJobResponse{
			ResourceID:     resourceID,
			SourceFilename: sourceFile,
		}, nil)

		transcoder.EXPECT().Transcode(mock.Anything, resourceID, sourceFile).Return(nil, errors.New("transcode failed"))
		failUC.EXPECT().Execute(mock.Anything, testJob, "transcode failed").Return(nil)
		logger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything).Once()
		logger.EXPECT().Infof(mock.Anything, mock.Anything).Maybe()

		go w.Start(t.Context())
		jobCh <- testJob
		close(jobCh)

		time.Sleep(50 * time.Millisecond)
	})

	t.Run("should mark as failed if saving to storage fails", func(t *testing.T) {
		startUC := mockjob.NewMockStartTranscodeJobUsecase(t)
		completeUC := mockjob.NewMockCompleteTranscodeJobUsecase(t)
		failUC := mockjob.NewMockFailTranscodeJobUsecase(t)
		storer := mockstorage.NewMockAssetStorer(t)
		logger := mocklog.NewMockLogger(t)
		transcoder := mocktranscode.NewMockTranscoder(t)
		jobCh := make(chan *job.Job, 1)

		w := worker.NewTranscodeWorker(startUC, completeUC, failUC, storer, logger, transcoder, jobCh)

		testJob, _ := job.NewJob("job-1", "video-1", job.TypeTranscode)
		resourceID := "res-1"
		sourceFile := "input.mp4"

		tempDir := t.TempDir()
		manifestName := "manifest.m3u8"
		os.WriteFile(filepath.Join(tempDir, manifestName), []byte("content"), 0644)

		startUC.EXPECT().Execute(mock.Anything, testJob).Return(&jobapp.StartTranscodeJobResponse{
			ResourceID:     resourceID,
			SourceFilename: sourceFile,
		}, nil)

		transcoder.EXPECT().Transcode(mock.Anything, resourceID, sourceFile).Return(&transcode.TranscodeOutput{
			Duration:     10 * time.Second,
			ManifestPath: filepath.Join(tempDir, manifestName),
			OutputFiles:  []string{manifestName},
		}, nil)

		storer.EXPECT().Save(mock.Anything, resourceID, manifestName, mock.Anything).Return(errors.New("save failed"))
		failUC.EXPECT().Execute(mock.Anything, testJob, "failed to save transcoded output").Return(nil)
		logger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		logger.EXPECT().Infof(mock.Anything, mock.Anything).Maybe()

		go w.Start(t.Context())
		jobCh <- testJob
		close(jobCh)

		time.Sleep(50 * time.Millisecond)
		_, err := os.Stat(tempDir)
		require.True(t, os.IsNotExist(err))
	})
}
