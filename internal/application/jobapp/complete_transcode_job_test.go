package jobapp

import (
	"context"
	"errors"
	"testing"
	"time"

	repomocks "github.com/st-ember/streaming-api/internal/application/ports/repo/mocks"
	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCompleteTranscodeJob_SuccessCase(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockVideoRepo := repomocks.NewMockVideoRepo(t)
	mockJobRepo := repomocks.NewMockJobRepo(t)
	mockUow := repomocks.NewMockUnitOfWork(t)
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)

	// Create valid domain objects for the test.
	// Job must be 'Running' to be completed.
	startJob, err := job.NewJob("job-id", "video-id", job.TypeTranscode)
	require.NoError(t, err)
	startJob.Status = job.StatusRunning

	// Video must be 'Processing' to be published.
	relatedVideo, err := video.NewVideo("video-id", "title", "desc", "file.mp4", "resource-id")
	require.NoError(t, err)
	relatedVideo.Status = video.StatusProcessing

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
	usecase := NewCompleteTranscodeJobUsecase(mockUowFactory)
	err = usecase.Execute(context.Background(), startJob, "success", 120*time.Second)

	// --- ASSERT ---
	require.NoError(t, err)
	// Assert that the domain objects were updated to their final state.
	require.Equal(t, job.StatusCompleted, startJob.Status)
	require.Equal(t, video.StatusPublished, relatedVideo.Status)
	require.Equal(t, 120*time.Second, relatedVideo.Duration)
}

func TestCompleteTranscodeJob_FailsIfJobCannotBeCompleted(t *testing.T) {
	t.Parallel()
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)

	// Create a job that is already completed.
	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	startJob.Status = job.StatusCompleted

	usecase := NewCompleteTranscodeJobUsecase(mockUowFactory)
	err := usecase.Execute(context.Background(), startJob, "success", 120*time.Second)

	// We expect a domain error here, before any mocks are called.
	require.Error(t, err)
	require.ErrorIs(t, err, job.ErrCannotBeCompleted)
}

func TestCompleteTranscodeJob_FailsOnFindVideoByID(t *testing.T) {
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

	usecase := NewCompleteTranscodeJobUsecase(mockUowFactory)
	err := usecase.Execute(context.Background(), startJob, "success", 120*time.Second)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestCompleteTranscodeJob_FailsOnVideoPublish(t *testing.T) {
	t.Parallel()
	mockVideoRepo := repomocks.NewMockVideoRepo(t)
	mockJobRepo := repomocks.NewMockJobRepo(t)
	mockUow := repomocks.NewMockUnitOfWork(t)
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)
	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	startJob.Status = job.StatusRunning

	// Create a video that is NOT in 'Processing' state, so Publish() will fail.
	relatedVideo, _ := video.NewVideo("video-id", "title", "desc", "file.mp4", "resource-id")
	relatedVideo.Status = video.StatusPending

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo).Once()
	mockUow.EXPECT().JobRepo().Return(mockJobRepo).Once()
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	mockVideoRepo.EXPECT().FindByID(mock.Anything, "video-id").Return(relatedVideo, nil).Once()

	usecase := NewCompleteTranscodeJobUsecase(mockUowFactory)
	err := usecase.Execute(context.Background(), startJob, "success", 120*time.Second)

	require.Error(t, err)
	require.ErrorIs(t, err, video.ErrCannotBePublished)
}

func TestCompleteTranscodeJob_FailsOnCommit(t *testing.T) {
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

	usecase := NewCompleteTranscodeJobUsecase(mockUowFactory)
	err := usecase.Execute(context.Background(), startJob, "success", 120*time.Second)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}
