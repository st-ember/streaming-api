package jobapp

import (
	"context"
	"errors"
	"testing"

	repomocks "github.com/st-ember/streaming-api/internal/application/ports/repo/mocks"
	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFailTranscodeJob_SuccessCase(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockVideoRepo := repomocks.NewMockVideoRepo(t)
	mockJobRepo := repomocks.NewMockJobRepo(t)
	mockUow := repomocks.NewMockUnitOfWork(t)
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)

	// Create valid domain objects for the test.
	// Job must be 'Running' to be failed.
	startJob, err := job.NewJob("job-id", "video-id", job.TypeTranscode)
	require.NoError(t, err)
	startJob.Status = job.StatusRunning

	// Video must be 'Processing' to be failed.
	relatedVideo, err := video.NewVideo("video-id", "title", "desc", "file.mp4", "resource-id")
	require.NoError(t, err)
	relatedVideo.Status = video.StatusProcessing

	errMsg := "transcode failed: invalid codec"

	// Define mock expectations for the success path.
	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo).Once()
	mockUow.EXPECT().JobRepo().Return(mockJobRepo).Once()
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	mockUow.EXPECT().Commit(mock.Anything).Return(nil).Once()

	mockVideoRepo.EXPECT().FindByID(mock.Anything, "video-id").Return(relatedVideo, nil).Once()
	mockVideoRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*video.Video")).Return(nil).Once()
	mockJobRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*job.Job")).Return(nil).Once()

	// --- ACT ---
	usecase := NewFailTranscodeJobUsecase(mockUowFactory)
	err = usecase.Execute(context.Background(), startJob, errMsg)

	// --- ASSERT ---
	require.NoError(t, err)
	// Assert that the domain objects were updated to their final state.
	require.Equal(t, job.StatusFailed, startJob.Status)
	require.Equal(t, errMsg, startJob.ErrorMsg)
	require.Equal(t, video.StatusFailed, relatedVideo.Status)
}

func TestFailTranscodeJob_FailsIfJobCannotBeFailed(t *testing.T) {
	t.Parallel()
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)

	// Create a job that is already completed.
	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	startJob.Status = job.StatusCompleted

	usecase := NewFailTranscodeJobUsecase(mockUowFactory)
	err := usecase.Execute(context.Background(), startJob, "some error")

	// We expect a domain error here, before any mocks are called.
	require.Error(t, err)
	require.ErrorIs(t, err, job.ErrCannotBeMarkedAsFailed)
}

func TestFailTranscodeJob_FailsOnFindVideoByID(t *testing.T) {
	t.Parallel()
	mockVideoRepo := repomocks.NewMockVideoRepo(t)
	mockJobRepo := repomocks.NewMockJobRepo(t)
	mockUow := repomocks.NewMockUnitOfWork(t)
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)

	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	startJob.Status = job.StatusRunning

	expectedErr := errors.New("video not found")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo).Once()
	mockUow.EXPECT().JobRepo().Return(mockJobRepo).Once()
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	mockVideoRepo.EXPECT().FindByID(mock.Anything, "video-id").Return(nil, expectedErr).Once()

	usecase := NewFailTranscodeJobUsecase(mockUowFactory)
	err := usecase.Execute(context.Background(), startJob, "some error")

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestFailTranscodeJob_FailsOnVideoMarkAsFailed(t *testing.T) {
	t.Parallel()
	mockVideoRepo := repomocks.NewMockVideoRepo(t)
	mockJobRepo := repomocks.NewMockJobRepo(t)
	mockUow := repomocks.NewMockUnitOfWork(t)
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)
	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	startJob.Status = job.StatusRunning

	// Create a video that is NOT in 'Processing' state, so MarkAsFailed() will fail.
	relatedVideo, _ := video.NewVideo("video-id", "title", "desc", "file.mp4", "resource-id")
	relatedVideo.Status = video.StatusPublished

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo).Once()
	mockUow.EXPECT().JobRepo().Return(mockJobRepo).Once()
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	mockVideoRepo.EXPECT().FindByID(mock.Anything, "video-id").Return(relatedVideo, nil).Once()

	usecase := NewFailTranscodeJobUsecase(mockUowFactory)
	err := usecase.Execute(context.Background(), startJob, "some error")

	require.Error(t, err)
	require.ErrorIs(t, err, video.ErrCannotBeMarkedAsFailed)
}

func TestFailTranscodeJob_FailsOnCommit(t *testing.T) {
	t.Parallel()
	mockVideoRepo := repomocks.NewMockVideoRepo(t)
	mockJobRepo := repomocks.NewMockJobRepo(t)
	mockUow := repomocks.NewMockUnitOfWork(t)
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)
	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	startJob.Status = job.StatusRunning
	relatedVideo, _ := video.NewVideo("video-id", "title", "desc", "file.mp4", "resource-id")
	relatedVideo.Status = video.StatusProcessing
	expectedErr := errors.New("commit failed")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo).Once()
	mockUow.EXPECT().JobRepo().Return(mockJobRepo).Once()
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	mockUow.EXPECT().Commit(mock.Anything).Return(expectedErr).Once()

	mockVideoRepo.EXPECT().FindByID(mock.Anything, "video-id").Return(relatedVideo, nil).Once()
	mockVideoRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*video.Video")).Return(nil).Once()
	mockJobRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*job.Job")).Return(nil).Once()

	usecase := NewFailTranscodeJobUsecase(mockUowFactory)
	err := usecase.Execute(context.Background(), startJob, "some error")

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}
