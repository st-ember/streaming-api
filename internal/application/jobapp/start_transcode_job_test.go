package jobapp

import (
	"errors"
	"testing"

	repomocks "github.com/st-ember/streaming-api/internal/application/ports/repo/mocks"
	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStartTranscodeJob_SuccessCase(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockVideoRepo := repomocks.NewMockVideoRepo(t)
	mockJobRepo := repomocks.NewMockJobRepo(t)
	mockUow := repomocks.NewMockUnitOfWork(t)
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)

	// Create valid domain objects for the test
	startJob, err := job.NewJob("job-id", "video-id", job.TypeTranscode)
	require.NoError(t, err)
	relatedVideo, err := video.NewVideo("video-id", "title", "desc", "file.mp4", "resource-id")
	require.NoError(t, err)

	// Define mock expectations for the success path
	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	mockUow.EXPECT().Commit(mock.Anything).Return(nil).Once()

	mockVideoRepo.EXPECT().FindByID(mock.Anything, "video-id").Return(relatedVideo, nil).Once()
	mockVideoRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*video.Video")).Return(nil).Once()
	mockJobRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*job.Job")).Return(nil).Once()

	// --- ACT ---
	usecase := NewStartTranscodeJobUsecase(mockUowFactory)
	resp, err := usecase.Execute(t.Context(), startJob)

	// --- ASSERT ---
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, relatedVideo.ResourceID, resp.ResourceID)
	require.Equal(t, relatedVideo.Filename, resp.SourceFilename)
	// Assert that the domain objects were updated
	require.Equal(t, video.StatusProcessing, relatedVideo.Status)
	require.Equal(t, job.StatusRunning, startJob.Status)
}

func TestStartTranscodeJob_FailsOnUOWCreation(t *testing.T) {
	t.Parallel()
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)
	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	expectedErr := errors.New("db down")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(nil, expectedErr).Once()

	usecase := NewStartTranscodeJobUsecase(mockUowFactory)
	_, err := usecase.Execute(t.Context(), startJob)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestStartTranscodeJob_FailsOnFindVideoByID(t *testing.T) {
	t.Parallel()
	mockVideoRepo := repomocks.NewMockVideoRepo(t)
	mockJobRepo := repomocks.NewMockJobRepo(t)
	mockUow := repomocks.NewMockUnitOfWork(t)
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)
	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	expectedErr := errors.New("video not found")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	mockVideoRepo.EXPECT().FindByID(mock.Anything, "video-id").Return(nil, expectedErr).Once()

	usecase := NewStartTranscodeJobUsecase(mockUowFactory)
	_, err := usecase.Execute(t.Context(), startJob)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestStartTranscodeJob_FailsOnJobSave(t *testing.T) {
	t.Parallel()
	mockVideoRepo := repomocks.NewMockVideoRepo(t)
	mockJobRepo := repomocks.NewMockJobRepo(t)
	mockUow := repomocks.NewMockUnitOfWork(t)
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)
	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	relatedVideo, _ := video.NewVideo("video-id", "title", "desc", "file.mp4", "resource-id")
	expectedErr := errors.New("job save failed")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	mockVideoRepo.EXPECT().FindByID(mock.Anything, "video-id").Return(relatedVideo, nil).Once()
	mockJobRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*job.Job")).Return(expectedErr).Once()

	usecase := NewStartTranscodeJobUsecase(mockUowFactory)
	_, err := usecase.Execute(t.Context(), startJob)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestStartTranscodeJob_FailsOnVideoSave(t *testing.T) {
	t.Parallel()
	mockVideoRepo := repomocks.NewMockVideoRepo(t)
	mockJobRepo := repomocks.NewMockJobRepo(t)
	mockUow := repomocks.NewMockUnitOfWork(t)
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)
	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	relatedVideo, _ := video.NewVideo("video-id", "title", "desc", "file.mp4", "resource-id")
	expectedErr := errors.New("video save failed")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	mockVideoRepo.EXPECT().FindByID(mock.Anything, "video-id").Return(relatedVideo, nil).Once()
	mockJobRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*job.Job")).Return(nil).Once()
	mockVideoRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*video.Video")).Return(expectedErr).Once()

	usecase := NewStartTranscodeJobUsecase(mockUowFactory)
	_, err := usecase.Execute(t.Context(), startJob)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestStartTranscodeJob_FailsOnCommit(t *testing.T) {
	t.Parallel()
	mockVideoRepo := repomocks.NewMockVideoRepo(t)
	mockJobRepo := repomocks.NewMockJobRepo(t)
	mockUow := repomocks.NewMockUnitOfWork(t)
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)
	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	relatedVideo, _ := video.NewVideo("video-id", "title", "desc", "file.mp4", "resource-id")
	expectedErr := errors.New("commit failed")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	mockVideoRepo.EXPECT().FindByID(mock.Anything, "video-id").Return(relatedVideo, nil).Once()
	mockJobRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*job.Job")).Return(nil).Once()
	mockVideoRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*video.Video")).Return(nil).Once()
	mockUow.EXPECT().Commit(mock.Anything).Return(expectedErr).Once()

	usecase := NewStartTranscodeJobUsecase(mockUowFactory)
	_, err := usecase.Execute(t.Context(), startJob)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestStartTranscodeJob_FailsIfJobCannotBeStarted(t *testing.T) {
	t.Parallel()
	mockUowFactory := repomocks.NewMockUnitOfWorkFactory(t)

	// Create a job that is already in a running state
	startJob, _ := job.NewJob("job-id", "video-id", job.TypeTranscode)
	startJob.Status = job.StatusRunning

	usecase := NewStartTranscodeJobUsecase(mockUowFactory)
	_, err := usecase.Execute(t.Context(), startJob)

	// We expect a domain error here, before any mocks are called.
	require.Error(t, err)
	require.ErrorIs(t, err, job.ErrCannotBeStarted)
}
