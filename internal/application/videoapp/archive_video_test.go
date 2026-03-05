package videoapp

import (
	"errors"
	"testing"

	repoMocks "github.com/st-ember/streaming-api/internal/application/ports/repo/mocks"
	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestArchiveVideo_SuccessCase(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	videoID := "video-123"
	resourceID := "resource-123"
	testVideo, _ := video.NewVideo(videoID, "Test", "Test", "test.mp4", resourceID)
	testVideo.Status = video.StatusPublished

	// Unit of Work Factory expectations
	mockUowFactory.EXPECT().
		NewUnitOfWork(mock.Anything).
		Return(mockUow, nil).
		Once()

	// Unit of Work expectations
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().Commit(mock.Anything).Return(nil).Once()
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Maybe()

	// Repo expectations
	mockVideoRepo.EXPECT().FindByID(mock.Anything, videoID).Return(testVideo, nil).Once()
	mockVideoRepo.EXPECT().Save(mock.Anything, testVideo).Return(nil).Once()

	// Create usecase
	usecase := NewArchiveVideoUsecase(mockUowFactory)

	// Execute
	err := usecase.Execute(t.Context(), videoID)

	// Assert
	require.NoError(t, err)
	require.Equal(t, video.StatusArchived, testVideo.Status)
}

func TestArchiveVideo_VideoNotFound(t *testing.T) {
	t.Parallel()

	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	videoID := "non-existent"

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	mockVideoRepo.EXPECT().FindByID(mock.Anything, videoID).Return(nil, errors.New("not found")).Once()

	usecase := NewArchiveVideoUsecase(mockUowFactory)
	err := usecase.Execute(t.Context(), videoID)

	require.Error(t, err)
	require.Contains(t, err.Error(), "find video")
}

func TestArchiveVideo_InvalidStateTransition(t *testing.T) {
	t.Parallel()

	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	videoID := "video-123"
	// Video is in 'Pending' state (cannot be archived yet)
	testVideo, _ := video.NewVideo(videoID, "Test", "Test", "test.mp4", "res-123")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	mockVideoRepo.EXPECT().FindByID(mock.Anything, videoID).Return(testVideo, nil).Once()

	usecase := NewArchiveVideoUsecase(mockUowFactory)
	err := usecase.Execute(t.Context(), videoID)

	require.Error(t, err)
	require.ErrorIs(t, err, video.ErrCannotBeArchived)
}
