package videoapp

import (
	"errors"
	"testing"

	repoMocks "github.com/st-ember/streaming-api/internal/application/ports/repo/mocks"
	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetVideoInfo_SuccessCase(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockJobRepo := repoMocks.NewMockJobRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	videoID := "video-123"
	resourceID := "resource-123"
	testVideo, _ := video.NewVideo(videoID, "Test Title", "Test Desc", "test.mp4", resourceID)
	testJob := &job.Job{
		ID:       "job-123",
		VideoID:  videoID,
		Status:   job.StatusCompleted,
		Result:   "storage/video-123/manifest.m3u8",
		ErrorMsg: "",
	}

	// Unit of Work Factory expectations
	mockUowFactory.EXPECT().
		NewUnitOfWork(mock.Anything).
		Return(mockUow, nil).
		Once()

	// Unit of Work expectations
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	// Repo expectations
	mockVideoRepo.EXPECT().FindByID(mock.Anything, videoID).Return(testVideo, nil).Once()
	mockJobRepo.EXPECT().FindByVideoID(mock.Anything, videoID).Return(testJob, nil).Once()

	// Create usecase
	usecase := NewGetVideoInfoUsecase(mockUowFactory)

	// Execute
	result, err := usecase.Execute(t.Context(), videoID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, testVideo.ID, result.Video.ID)
	require.Equal(t, testJob.Result, result.ManifestPath)
	require.Equal(t, "", result.ErrorMsg)
}

func TestGetVideoInfo_VideoNotFound(t *testing.T) {
	t.Parallel()

	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockJobRepo := repoMocks.NewMockJobRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	videoID := "non-existent"
	expectedErr := errors.New("not found")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	mockVideoRepo.EXPECT().FindByID(mock.Anything, videoID).Return(nil, expectedErr).Once()

	usecase := NewGetVideoInfoUsecase(mockUowFactory)
	result, err := usecase.Execute(t.Context(), videoID)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), expectedErr.Error())
}

func TestGetVideoInfo_JobNotFound(t *testing.T) {
	t.Parallel()

	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockJobRepo := repoMocks.NewMockJobRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	videoID := "video-123"
	resourceID := "resource-123"
	testVideo, _ := video.NewVideo(videoID, "Test", "Test", "test.mp4", resourceID)
	expectedErr := errors.New("job not found")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	mockVideoRepo.EXPECT().FindByID(mock.Anything, videoID).Return(testVideo, nil).Once()
	mockJobRepo.EXPECT().FindByVideoID(mock.Anything, videoID).Return(nil, expectedErr).Once()

	usecase := NewGetVideoInfoUsecase(mockUowFactory)
	result, err := usecase.Execute(t.Context(), videoID)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), expectedErr.Error())
}
